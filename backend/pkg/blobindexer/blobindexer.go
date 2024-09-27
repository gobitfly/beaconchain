package blobindexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"go.uber.org/atomic"

	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/errgroup"
)

var enableCheckingBeforePutting = false
var waitForOtherBlobIndexerDuration = time.Second * 60

type BlobIndexer struct {
	S3Client          *s3.Client
	running           bool
	runningMu         *sync.Mutex
	clEndpoint        string
	cl                consapi.Client
	id                string
	networkID         string
	writtenBlobsCache *lru.Cache[string, bool]
}

func NewBlobIndexer() (*BlobIndexer, error) {
	initDB()
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			utils.Config.BlobIndexer.S3.AccessKeyId,
			utils.Config.BlobIndexer.S3.AccessKeySecret,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(utils.Config.BlobIndexer.S3.Endpoint)
	})

	writtenBlobsCache, err := lru.New[string, bool](1000)
	if err != nil {
		return nil, err
	}

	id := utils.GetUUID()
	bi := &BlobIndexer{
		S3Client:          s3Client,
		runningMu:         &sync.Mutex{},
		clEndpoint:        "http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port,
		cl:                consapi.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port),
		id:                id,
		writtenBlobsCache: writtenBlobsCache,
	}
	return bi, nil
}

func initDB() {
	if utils.Config.BlobIndexer.DisableStatusReports {
		return
	}
	if db.WriterDb != nil && db.ReaderDb != nil {
		return
	}
	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
		Username:     utils.Config.WriterDatabase.Username,
		Password:     utils.Config.WriterDatabase.Password,
		Name:         utils.Config.WriterDatabase.Name,
		Host:         utils.Config.WriterDatabase.Host,
		Port:         utils.Config.WriterDatabase.Port,
		MaxOpenConns: utils.Config.WriterDatabase.MaxOpenConns,
		MaxIdleConns: utils.Config.WriterDatabase.MaxIdleConns,
		SSL:          utils.Config.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     utils.Config.ReaderDatabase.Username,
		Password:     utils.Config.ReaderDatabase.Password,
		Name:         utils.Config.ReaderDatabase.Name,
		Host:         utils.Config.ReaderDatabase.Host,
		Port:         utils.Config.ReaderDatabase.Port,
		MaxOpenConns: utils.Config.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: utils.Config.ReaderDatabase.MaxIdleConns,
		SSL:          utils.Config.ReaderDatabase.SSL,
	}, "pgx", "postgres")
}

func (bi *BlobIndexer) Start() {
	bi.runningMu.Lock()
	if bi.running {
		bi.runningMu.Unlock()
		return
	}
	bi.running = true
	bi.runningMu.Unlock()

	log.InfoWithFields(log.Fields{"version": version.Version, "clEndpoint": bi.clEndpoint, "s3Endpoint": utils.Config.BlobIndexer.S3.Endpoint, "id": bi.id}, "starting blobindexer")
	for {
		err := bi.index()
		if err != nil {
			log.Error(err, "failed indexing blobs", 0)
		}
		time.Sleep(time.Second * 10)
	}
}

func (bi *BlobIndexer) index() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	headHeader := &constypes.StandardBeaconHeaderResponse{}
	finalizedHeader := &constypes.StandardBeaconHeaderResponse{}
	spec := &constypes.StandardSpecResponse{}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(3)
	g.Go(func() error {
		var err error
		spec, err = bi.cl.GetSpec()
		if err != nil {
			return fmt.Errorf("error bi.cl.GetSpec: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		headHeader, err = bi.cl.GetBlockHeader("head")
		if err != nil {
			return fmt.Errorf("error bi.cl.GetBlockHeader(head): %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		finalizedHeader, err = bi.cl.GetBlockHeader("finalized")
		if err != nil {
			return fmt.Errorf("error bi.cl.GetBlockHeader(finalized): %w", err)
		}
		return nil
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	if spec.Data.DenebForkEpoch == nil {
		return fmt.Errorf("DENEB_FORK_EPOCH not set in spec")
	}
	if spec.Data.MinEpochsForBlobSidecarsRequests == nil {
		return fmt.Errorf("MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS not set in spec")
	}

	nodeDepositNetworkId := uint64(spec.Data.DepositNetworkID)
	if utils.Config.Chain.ClConfig.DepositNetworkID != nodeDepositNetworkId {
		return fmt.Errorf("config.DepositNetworkId != node.DepositNetworkId: %v != %v", utils.Config.Chain.ClConfig.DepositNetworkID, nodeDepositNetworkId)
	}
	bi.networkID = fmt.Sprintf("%d", nodeDepositNetworkId)

	status, err := bi.GetIndexerStatus()
	if err != nil {
		return fmt.Errorf("error bi.GetIndexerStatus: %w", err)
	}

	// skip if another blobIndexer is already indexing - it is ok if multiple blobIndexers are indexing the same finalized slot, this is just best effort to avoid duplicate work
	if status.CurrentBlobIndexerId != bi.id && status.LastUpdate.After(time.Now().Add(-waitForOtherBlobIndexerDuration)) {
		log.InfoWithFields(log.Fields{"lastIndexedFinalizedSlot": status.LastIndexedFinalizedSlot, "currentBlobIndexerId": status.CurrentBlobIndexerId, "finalizedSlot": finalizedHeader.Data.Header.Message.Slot, "lastUpdate": status.LastUpdate}, "found other blobIndexer indexing, skipping")
		return nil
	}

	// check if node still has last indexed blobs (if its outside the range defined by MAX_REQUEST_BLOCKS_DENEB), otherwise assume that the node has pruned too far and we would miss blobs
	minBlobSlotRange := *spec.Data.MinEpochsForBlobSidecarsRequests * uint64(spec.Data.SlotsPerEpoch)
	minBlobSlot := uint64(0)
	if headHeader.Data.Header.Message.Slot > minBlobSlotRange {
		minBlobSlot = headHeader.Data.Header.Message.Slot - minBlobSlotRange
	}
	pruneMarginSlotRange := utils.Config.BlobIndexer.PruneMarginEpochs * uint64(spec.Data.SlotsPerEpoch)
	if minBlobSlot > pruneMarginSlotRange {
		minBlobSlot = minBlobSlot - pruneMarginSlotRange
	}
	if status.LastIndexedFinalizedSlot < minBlobSlot && status.LastIndexedFinalizedBlobSlot > 0 {
		bs, err := bi.cl.GetBlobSidecars(status.LastIndexedFinalizedBlobSlot)
		if err != nil {
			return err
		}
		if len(bs.Data) == 0 {
			return fmt.Errorf("no blobs found at lastIndexedFinalizedBlobSlot: %v, node has pruned too far?", status.LastIndexedFinalizedBlobSlot)
		}
	}

	lastIndexedFinalizedBlobSlot := atomic.NewUint64(status.LastIndexedFinalizedBlobSlot)

	denebForkSlot := *spec.Data.DenebForkEpoch * uint64(spec.Data.SlotsPerEpoch)
	startSlot := status.LastIndexedFinalizedSlot + 1
	if status.LastIndexedFinalizedSlot <= denebForkSlot {
		startSlot = denebForkSlot
	}

	if headHeader.Data.Header.Message.Slot <= startSlot {
		return fmt.Errorf("headHeader.Data.Header.Message.Slot <= startSlot: %v < %v (denebForkEpoch: %v, denebForkSlot: %v, slotsPerEpoch: %v)", headHeader.Data.Header.Message.Slot, startSlot, utils.Config.Chain.ClConfig.DenebForkEpoch, denebForkSlot, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}

	start := time.Now()
	log.InfoWithFields(log.Fields{
		"lastIndexedFinalizedSlot": status.LastIndexedFinalizedSlot,
		"headSlot":                 headHeader.Data.Header.Message.Slot,
		"finalizedSlot":            finalizedHeader.Data.Header.Message.Slot,
		"startSlot":                startSlot,
		"networkID":                bi.networkID,
	}, "indexing blobs")
	defer func() {
		log.InfoWithFields(log.Fields{
			"startSlot":     startSlot,
			"headSlot":      headHeader.Data.Header.Message.Slot,
			"finalizedSlot": finalizedHeader.Data.Header.Message.Slot,
			"duration":      time.Since(start),
			"networkID":     bi.networkID,
		}, "finished indexing blobs")
	}()

	batchSize := uint64(100)
	for batchStart := startSlot; batchStart <= headHeader.Data.Header.Message.Slot; batchStart += batchSize {
		batchStartTs := time.Now()
		batchBlobsIndexed := atomic.NewInt64(0)
		batchEnd := batchStart + batchSize
		if batchEnd > headHeader.Data.Header.Message.Slot {
			batchEnd = headHeader.Data.Header.Message.Slot
		}
		g, gCtx = errgroup.WithContext(context.Background())
		g.SetLimit(4)
		for slot := batchStart; slot <= batchEnd; slot++ {
			slot := slot
			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}
				numBlobs, err := bi.indexBlobsAtSlot(slot)
				if err != nil {
					return fmt.Errorf("error bi.IndexBlobsAtSlot(%v): %w", slot, err)
				}
				if numBlobs > 0 && slot <= finalizedHeader.Data.Header.Message.Slot && slot > lastIndexedFinalizedBlobSlot.Load() {
					lastIndexedFinalizedBlobSlot.Store(slot)
				}
				batchBlobsIndexed.Add(int64(numBlobs))
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			return err
		}
		lastIndexedFinalizedSlot := uint64(0)
		if batchEnd <= finalizedHeader.Data.Header.Message.Slot {
			lastIndexedFinalizedSlot = batchEnd
		} else {
			lastIndexedFinalizedSlot = finalizedHeader.Data.Header.Message.Slot
		}
		newBlobIndexerStatus := BlobIndexerStatus{
			LastIndexedFinalizedSlot:     lastIndexedFinalizedSlot,
			LastIndexedFinalizedBlobSlot: lastIndexedFinalizedBlobSlot.Load(),
			CurrentBlobIndexerId:         bi.id,
			LastUpdate:                   time.Now(),
			BlobIndexerVersion:           version.Version,
		}
		if status.LastIndexedFinalizedBlobSlot > newBlobIndexerStatus.LastIndexedFinalizedBlobSlot {
			newBlobIndexerStatus.LastIndexedFinalizedBlobSlot = status.LastIndexedFinalizedBlobSlot
		}
		err := bi.putIndexerStatus(newBlobIndexerStatus)
		if err != nil {
			return fmt.Errorf("error updating indexer status at slot %v: %w", batchEnd, err)
		}
		slotsPerSecond := float64(batchEnd-batchStart) / time.Since(batchStartTs).Seconds()
		blobsPerSecond := float64(batchBlobsIndexed.Load()) / time.Since(batchStartTs).Seconds()
		estimatedTimeToHead := float64(headHeader.Data.Header.Message.Slot-batchStart) / slotsPerSecond
		estimatedTimeToHeadDuration := time.Duration(estimatedTimeToHead) * time.Second
		log.InfoWithFields(log.Fields{
			"lastIdxFinSlot":      newBlobIndexerStatus.LastIndexedFinalizedSlot,
			"lastIdxFinBlobSlot":  newBlobIndexerStatus.LastIndexedFinalizedBlobSlot,
			"batch":               fmt.Sprintf("%d-%d", batchStart, batchEnd),
			"duration":            time.Since(batchStartTs),
			"slotsPerSecond":      fmt.Sprintf("%.3f", slotsPerSecond),
			"blobsPerSecond":      fmt.Sprintf("%.3f", blobsPerSecond),
			"estimatedTimeToHead": estimatedTimeToHeadDuration,
			"blobsIndexed":        batchBlobsIndexed.Load(),
		}, "updated indexer status")
		if !utils.Config.BlobIndexer.DisableStatusReports {
			services.ReportStatus("blobindexer", "Running", nil)
		}
	}
	return nil
}

func (bi *BlobIndexer) indexBlobsAtSlot(slot uint64) (int, error) {
	tGetBlobSidcar := time.Now()

	blobSidecar, err := bi.cl.GetBlobSidecars(slot)
	if err != nil {
		httpErr := network.SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			// no sidecar for this slot
			return 0, nil
		}
		return 0, err
	}
	metrics.TaskDuration.WithLabelValues("blobindexer_get_blob_sidecars").Observe(time.Since(tGetBlobSidcar).Seconds())

	if len(blobSidecar.Data) <= 0 {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for _, d := range blobSidecar.Data {
		d := d
		versionedBlobHash := fmt.Sprintf("%#x", utils.VersionedBlobHash(d.KzgCommitment).Bytes())
		key := fmt.Sprintf("%s/blobs/%s", bi.networkID, versionedBlobHash)

		if bi.writtenBlobsCache.Contains(key) {
			continue
		}

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			if enableCheckingBeforePutting {
				tS3HeadObj := time.Now()
				_, err = bi.S3Client.HeadObject(gCtx, &s3.HeadObjectInput{
					Bucket: &utils.Config.BlobIndexer.S3.Bucket,
					Key:    &key,
				})
				metrics.TaskDuration.WithLabelValues("blobindexer_check_blob").Observe(time.Since(tS3HeadObj).Seconds())
				if err != nil {
					// Only put the object if it does not exist yet
					var httpResponseErr *awshttp.ResponseError
					if errors.As(err, &httpResponseErr) && (httpResponseErr.HTTPStatusCode() == http.StatusNotFound || httpResponseErr.HTTPStatusCode() == 403) {
						return nil
					}
					return fmt.Errorf("error getting headObject: %s (%v/%v): %w", key, d.SignedBlockHeader.Message.Slot, d.Index, err)
				}
			}

			tS3PutObj := time.Now()
			_, putErr := bi.S3Client.PutObject(gCtx, &s3.PutObjectInput{
				Bucket: &utils.Config.BlobIndexer.S3.Bucket,
				Key:    &key,
				Body:   bytes.NewReader(d.Blob),
				Metadata: map[string]string{
					"blob_index":        fmt.Sprintf("%d", d.Index),
					"block_slot":        fmt.Sprintf("%d", d.SignedBlockHeader.Message.Slot),
					"block_proposer":    fmt.Sprintf("%d", d.SignedBlockHeader.Message.ProposerIndex),
					"block_state_root":  d.SignedBlockHeader.Message.StateRoot.String(),
					"block_parent_root": d.SignedBlockHeader.Message.ParentRoot.String(),
					"block_body_root":   d.SignedBlockHeader.Message.BodyRoot.String(),
					"kzg_commitment":    d.KzgCommitment.String(),
					"kzg_proof":         d.KzgProof.String(),
				},
			})
			metrics.TaskDuration.WithLabelValues("blobindexer_put_blob").Observe(time.Since(tS3PutObj).Seconds())
			if putErr != nil {
				return fmt.Errorf("error putting object: %s (%v/%v): %w", key, d.SignedBlockHeader.Message.Slot, d.Index, putErr)
			}
			bi.writtenBlobsCache.Add(key, true)

			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return len(blobSidecar.Data), fmt.Errorf("error indexing blobs at slot %v: %w", slot, err)
	}

	return len(blobSidecar.Data), nil
}

func (bi *BlobIndexer) GetIndexerStatus() (*BlobIndexerStatus, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_get_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := fmt.Sprintf("%s/blob-indexer-status.json", bi.networkID)
	obj, err := bi.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &utils.Config.BlobIndexer.S3.Bucket,
		Key:    &key,
	})
	if err != nil {
		// If the object that you request doesn’t exist, the error that Amazon S3 returns depends on whether you also have the s3:ListBucket permission. If you have the s3:ListBucket permission on the bucket, Amazon S3 returns an HTTP status code 404 (Not Found) error. If you don’t have the s3:ListBucket permission, Amazon S3 returns an HTTP status code 403 ("access denied") error.
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) && (httpResponseErr.HTTPStatusCode() == 404 || httpResponseErr.HTTPStatusCode() == 403) {
			return &BlobIndexerStatus{}, nil
		}
		return nil, err
	}
	status := &BlobIndexerStatus{}
	err = json.NewDecoder(obj.Body).Decode(status)
	return status, err
}

func (bi *BlobIndexer) putIndexerStatus(status BlobIndexerStatus) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_put_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := fmt.Sprintf("%s/blob-indexer-status.json", bi.networkID)
	contentType := "application/json"
	body, err := json.Marshal(&status)
	if err != nil {
		return err
	}
	_, err = bi.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &utils.Config.BlobIndexer.S3.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(body),
		ContentType: &contentType,
		Metadata: map[string]string{
			"last_indexed_finalized_slot":      fmt.Sprintf("%d", status.LastIndexedFinalizedSlot),
			"last_indexed_finalized_blob_slot": fmt.Sprintf("%d", status.LastIndexedFinalizedBlobSlot),
			"current_blob_indexer_id":          status.CurrentBlobIndexerId,
			"last_update":                      status.LastUpdate.Format(time.RFC3339),
			"blob_indexer_version":             status.BlobIndexerVersion,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

type BlobIndexerStatus struct {
	LastIndexedFinalizedSlot     uint64    `json:"last_indexed_finalized_slot"`      // last finalized slot that was indexed
	LastIndexedFinalizedBlobSlot uint64    `json:"last_indexed_finalized_blob_slot"` // last finalized slot that included a blob
	CurrentBlobIndexerId         string    `json:"current_blob_indexer_id"`
	LastUpdate                   time.Time `json:"last_update"`
	BlobIndexerVersion           string    `json:"blob_indexer_version"`
}
