package blobindexer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/blobstore"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"go.uber.org/atomic"

	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/errgroup"
)

var enableCheckingBeforePutting = false
var waitForOtherBlobIndexerDuration = time.Second * 60

type BlobIndexer struct {
	BlobStore         blobstore.BlobStore
	running           bool
	runningMu         *sync.Mutex
	clEndpoint        string
	cl                consapi.ClientInt
	id                string
	networkID         string
	writtenBlobsCache *lru.Cache[string, bool]
}

type BlobIndexerStatus struct {
	LastIndexedFinalizedSlot     uint64    `json:"last_indexed_finalized_slot"`      // last finalized slot that was indexed
	LastIndexedFinalizedBlobSlot uint64    `json:"last_indexed_finalized_blob_slot"` // last finalized slot that included a blob
	CurrentBlobIndexerId         string    `json:"current_blob_indexer_id"`
	LastUpdate                   time.Time `json:"last_update"`
	BlobIndexerVersion           string    `json:"blob_indexer_version"`
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

	blobStore := blobstore.NewS3BlobStore(s3Client)
	writtenBlobsCache, err := lru.New[string, bool](1000)
	if err != nil {
		return nil, err
	}

	id := utils.GetUUID()
	clientCreator := &consapi.DefaultClientCreator{}

	bi := &BlobIndexer{
		BlobStore:         blobStore,
		runningMu:         &sync.Mutex{},
		clEndpoint:        "http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port,
		cl:                clientCreator.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port),
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
	db.WriterDb, db.ReaderDb = db.MustInitDB(&utils.Config.WriterDatabase, &utils.Config.ReaderDatabase, "pgx", "postgres")
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

	headHeader, finalizedHeader, spec, err := bi.fetchNodeData(ctx)
	if err != nil {
		return err
	}

	if err := validateSpec(spec); err != nil {
		return err
	}

	bi.networkID = fmt.Sprintf("%d", uint64(spec.Data.DepositNetworkID))

	status, err := bi.getIndexerStatusFromS3()
	if err != nil {
		return fmt.Errorf("error getting indexer status from S3: %w", err)
	}

	// skip if another blobIndexer is already indexing - it is ok if multiple blobIndexers are
	// indexing the same finalized slot, this is just best effort to avoid duplicate work
	if bi.shouldSkipBlobIndexing(status) {
		log.InfoWithFields(log.Fields{
			"lastIndexedFinalizedSlot": status.LastIndexedFinalizedSlot,
			"currentBlobIndexerId":     status.CurrentBlobIndexerId,
			"finalizedSlot":            finalizedHeader.Data.Header.Message.Slot,
			"lastUpdate":               status.LastUpdate},
			"found other blobIndexer indexing, skipping")
		return nil
	}

	// check if node still has last indexed blobs (if its outside the range defined by MAX_REQUEST_BLOCKS_DENEB),
	// otherwise assume that the node has pruned too far and we would miss blobs
	minBlobSlot := calculateMinBlobSlot(spec, headHeader)

	// check if node has pruned too far
	if err := bi.checkNodePruning(minBlobSlot, status); err != nil {
		return err
	}

	startSlot, denebForkSlot := calculateStartSlot(status, spec)
	if headHeader.Data.Header.Message.Slot <= startSlot {
		return fmt.Errorf("headHeader.Data.Header.Message.Slot <= startSlot: %v < %v (denebForkEpoch: %v, denebForkSlot: %v, slotsPerEpoch: %v)", headHeader.Data.Header.Message.Slot, startSlot, utils.Config.Chain.ClConfig.DenebForkEpoch, denebForkSlot, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}

	if err := bi.indexBlobsInBatches(status, headHeader, finalizedHeader, startSlot); err != nil {
		return err
	}
	return nil
}

func (bi *BlobIndexer) getIndexerStatusFromS3() (*BlobIndexerStatus, error) {
	key := fmt.Sprintf("%s/blob-indexer-status.json", bi.networkID)
	label := "blobindexer_get_indexer_status"
	obj, err := bi.getObjectFromS3(key, label)
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3 with key %s: %w", key, err)
	}

	status := &BlobIndexerStatus{}
	err = json.NewDecoder(obj.Body).Decode(status)
	return status, err
}

func (bi *BlobIndexer) storeIndexerStatusInS3(status BlobIndexerStatus) error {
	key := fmt.Sprintf("%s/blob-indexer-status.json", bi.networkID)
	contentType := "application/json"
	metadata := map[string]string{
		"last_indexed_finalized_slot":      fmt.Sprintf("%d", status.LastIndexedFinalizedSlot),
		"last_indexed_finalized_blob_slot": fmt.Sprintf("%d", status.LastIndexedFinalizedBlobSlot),
		"current_blob_indexer_id":          status.CurrentBlobIndexerId,
		"last_update":                      status.LastUpdate.Format(time.RFC3339),
		"blob_indexer_version":             status.BlobIndexerVersion,
	}

	body, err := json.Marshal(&status)
	if err != nil {
		return err
	}

	err = bi.putObjectInS3(key, "blobindexer_put_indexer_status", &contentType, body, metadata)
	if err != nil {
		return fmt.Errorf("error putting object in S3 with key %s: %w", key, err)
	}
	return nil
}

func (bi *BlobIndexer) getBlobSidecarsAtSlot(slot uint64) (*constypes.StandardBlobSidecarsResponse, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_get_blob_sidecars").Observe(time.Since(start).Seconds())
	}()

	blobSidecar, err := bi.cl.GetBlobSidecars(slot)
	if err != nil {
		httpErr := network.SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			// no sidecar for this slot
			return nil, nil
		}
		return nil, err
	}

	if len(blobSidecar.Data) == 0 {
		return nil, nil
	}

	return blobSidecar, nil
}

func (bi *BlobIndexer) putObjectInS3(key, logLabel string, contentType *string, data []byte, metadata map[string]string) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(logLabel).Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	blob := blobstore.Blob{
		Body:        bytes.NewReader(data),
		ContentType: contentType,
		Metadata:    metadata,
	}

	err := bi.BlobStore.Put(ctx,
		utils.Config.BlobIndexer.S3.Bucket,
		key,
		blob)

	if err != nil {
		return err
	}

	return nil
}

func (bi *BlobIndexer) getObjectFromS3(key, logLabel string) (blobstore.Blob, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(logLabel).Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	blob, err := bi.BlobStore.Get(ctx,
		utils.Config.BlobIndexer.S3.Bucket,
		key)
	if err != nil {
		return blobstore.Blob{}, err
	}

	return blob, nil
}

func (bi *BlobIndexer) checkIfObjectExistsInS3(key, logLabel string) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(logLabel).Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	exists, err := bi.BlobStore.Exist(ctx,
		utils.Config.BlobIndexer.S3.Bucket,
		key)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("object with key %s does not exist", key)
	}

	return nil
}

func (bi *BlobIndexer) fetchNodeData(ctx context.Context) (*constypes.StandardBeaconHeaderResponse, *constypes.StandardBeaconHeaderResponse, *constypes.StandardSpecResponse, error) {
	headHeader := &constypes.StandardBeaconHeaderResponse{}
	finalizedHeader := &constypes.StandardBeaconHeaderResponse{}
	spec := &constypes.StandardSpecResponse{}
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(3)

	g.Go(func() error {
		return bi.fetchSpec(spec)
	})
	g.Go(func() error {
		return bi.fetchBlockHeader("head", headHeader)
	})
	g.Go(func() error {
		return bi.fetchBlockHeader("finalized", finalizedHeader)
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}
	return headHeader, finalizedHeader, spec, nil
}

func (bi *BlobIndexer) fetchSpec(spec *constypes.StandardSpecResponse) error {
	s, err := bi.cl.GetSpec()
	if err != nil {
		return fmt.Errorf("error fetching node spec: %w", err)
	}
	*spec = *s
	return nil
}

func (bi *BlobIndexer) fetchBlockHeader(blockID string, header *constypes.StandardBeaconHeaderResponse) error {
	h, err := bi.cl.GetBlockHeader(blockID)
	if err != nil {
		return fmt.Errorf("error fetching node block header (%s): %w", blockID, err)
	}
	*header = *h
	return nil
}

func validateSpec(spec *constypes.StandardSpecResponse) error {
	if spec.Data.DenebForkEpoch == nil {
		return fmt.Errorf("DENEB_FORK_EPOCH not set in spec")
	}
	if spec.Data.MinEpochsForBlobSidecarsRequests == nil {
		return fmt.Errorf("MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS not set in spec")
	}
	nodeDepositNetworkId := uint64(spec.Data.DepositNetworkID)
	if utils.Config.Chain.ClConfig.DepositNetworkID != nodeDepositNetworkId {
		return fmt.Errorf("ClConfig.DepositNetworkID mismatch: %v != %v", utils.Config.Chain.ClConfig.DepositNetworkID, nodeDepositNetworkId)
	}

	return nil
}

func (bi *BlobIndexer) shouldSkipBlobIndexing(status *BlobIndexerStatus) bool {
	return status.CurrentBlobIndexerId != bi.id &&
		status.LastUpdate.After(time.Now().Add(-waitForOtherBlobIndexerDuration))
}

func (bi *BlobIndexer) checkNodePruning(minBlobSlot uint64, status *BlobIndexerStatus) error {
	if status.LastIndexedFinalizedSlot < minBlobSlot && status.LastIndexedFinalizedBlobSlot > 0 {
		_, err := bi.getBlobSidecarsAtSlot(status.LastIndexedFinalizedBlobSlot)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bi *BlobIndexer) indexBlobsInBatches(status *BlobIndexerStatus, headHeader, finalizedHeader *constypes.StandardBeaconHeaderResponse, startSlot uint64) error {
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

	lastIndexedFinalizedBlobSlot := atomic.NewUint64(status.LastIndexedFinalizedBlobSlot)
	batchSize := uint64(100)
	for batchStart := startSlot; batchStart <= headHeader.Data.Header.Message.Slot; batchStart += batchSize {
		batchStartTs := time.Now()

		batchEnd := calculateBatchEnd(batchStart, batchSize, headHeader.Data.Header.Message.Slot)

		blobsIndexed, err := bi.processBlobs(batchStart, batchEnd, finalizedHeader, lastIndexedFinalizedBlobSlot)
		if err != nil {
			return err
		}

		newBlobIndexerStatus := updateIndexerStatus(batchEnd, finalizedHeader, lastIndexedFinalizedBlobSlot, status, bi)
		err = bi.storeIndexerStatusInS3(newBlobIndexerStatus)
		if err != nil {
			return fmt.Errorf("error updating indexer status at slot %v: %w", batchEnd, err)
		}

		slotsPerSecond := calculateSlotsPerSecond(batchEnd, batchStart, batchStartTs)
		blobsPerSecond := calculateBlobsPerSecond(blobsIndexed, batchStartTs)
		estimatedTimeToHead := calculateEstimatedTimeToHead(headHeader, batchStart, slotsPerSecond)
		estimatedTimeToHeadDuration := calculateEstimatedTimeToHeadDuration(estimatedTimeToHead)

		log.InfoWithFields(log.Fields{
			"lastIdxFinSlot":      newBlobIndexerStatus.LastIndexedFinalizedSlot,
			"lastIdxFinBlobSlot":  newBlobIndexerStatus.LastIndexedFinalizedBlobSlot,
			"batch":               fmt.Sprintf("%d-%d", batchStart, batchEnd),
			"duration":            time.Since(batchStartTs),
			"slotsPerSecond":      fmt.Sprintf("%.3f", slotsPerSecond),
			"blobsPerSecond":      fmt.Sprintf("%.3f", blobsPerSecond),
			"estimatedTimeToHead": estimatedTimeToHeadDuration,
			"blobsIndexed":        blobsIndexed.Load(),
		}, "updated indexer status")

		reportBlobIndexerStatus()
	}
	return nil
}

func (bi *BlobIndexer) processBlobs(batchStart, batchEnd uint64, finalizedHeader *constypes.StandardBeaconHeaderResponse, lastIndexedFinalizedBlobSlot *atomic.Uint64) (*atomic.Int64, error) {
	batchBlobsIndexed := atomic.NewInt64(0)

	g, gCtx := errgroup.WithContext(context.Background())
	g.SetLimit(4)

	for slot := batchStart; slot <= batchEnd; slot++ {
		slot := slot
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			blobSidecar, err := bi.getBlobSidecarsAtSlot(slot)
			if err != nil {
				return fmt.Errorf("error getting blob sidecars for slot %v: %w", slot, err)
			}

			err = bi.storeBlobsInS3(blobSidecar.Data)
			if err != nil {
				return fmt.Errorf("error indexing blobs at slot %v: %w", slot, err)
			}

			if len(blobSidecar.Data) > 0 &&
				slot <= finalizedHeader.Data.Header.Message.Slot &&
				slot > lastIndexedFinalizedBlobSlot.Load() {
				lastIndexedFinalizedBlobSlot.Store(slot)
			}

			batchBlobsIndexed.Add(int64(len(blobSidecar.Data)))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return batchBlobsIndexed, nil
}

func (bi *BlobIndexer) storeBlobsInS3(blobs []constypes.BlobSidecarsData) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(4)

	for _, d := range blobs {
		d := d
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				return bi.storeBlobInS3(d)
			}
		})
	}

	return g.Wait()
}

func (bi *BlobIndexer) storeBlobInS3(blob constypes.BlobSidecarsData) error {
	versionedBlobHash := fmt.Sprintf("%#x", utils.VersionedBlobHash(blob.KzgCommitment).Bytes())
	key := fmt.Sprintf("%s/blobs/%s", bi.networkID, versionedBlobHash)
	label := "blobindexer_check_blob"
	metadata := map[string]string{
		"blob_index":        fmt.Sprintf("%d", blob.Index),
		"block_slot":        fmt.Sprintf("%d", blob.SignedBlockHeader.Message.Slot),
		"block_proposer":    fmt.Sprintf("%d", blob.SignedBlockHeader.Message.ProposerIndex),
		"block_state_root":  blob.SignedBlockHeader.Message.StateRoot.String(),
		"block_parent_root": blob.SignedBlockHeader.Message.ParentRoot.String(),
		"block_body_root":   blob.SignedBlockHeader.Message.BodyRoot.String(),
		"kzg_commitment":    blob.KzgCommitment.String(),
		"kzg_proof":         blob.KzgProof.String(),
	}

	if bi.writtenBlobsCache.Contains(key) {
		return nil
	}

	// check if the blob exists in S3
	if enableCheckingBeforePutting {
		if err := bi.checkIfObjectExistsInS3(key, label); err != nil {
			return fmt.Errorf("error checking object metadata in S3 with key %s: %w", key, err)
		}
	}

	err := bi.putObjectInS3(key, "blobindexer_put_blob", nil, blob.Blob, metadata)
	if err != nil {
		return fmt.Errorf("error putting object in S3: %s (%v/%v): %w", key, blob.SignedBlockHeader.Message.Slot, blob.Index, err)
	}

	bi.writtenBlobsCache.Add(key, true)

	return nil
}

func updateIndexerStatus(batchEnd uint64, finalizedHeader *constypes.StandardBeaconHeaderResponse, lastIndexedFinalizedBlobSlot *atomic.Uint64, status *BlobIndexerStatus, bi *BlobIndexer) BlobIndexerStatus {
	lastIndexedFinalizedSlot := getLastIndexedFinalizedSlot(batchEnd, finalizedHeader)
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
	return newBlobIndexerStatus
}

func calculateBatchEnd(batchStart, batchSize, headSlot uint64) uint64 {
	batchEnd := batchStart + batchSize
	if batchEnd > headSlot {
		batchEnd = headSlot
	}
	return batchEnd
}

func calculateMinBlobSlot(spec *constypes.StandardSpecResponse, headHeader *constypes.StandardBeaconHeaderResponse) uint64 {
	minBlobSlotRange := *spec.Data.MinEpochsForBlobSidecarsRequests * uint64(spec.Data.SlotsPerEpoch)
	minBlobSlot := uint64(0)
	if headHeader.Data.Header.Message.Slot > minBlobSlotRange {
		minBlobSlot = headHeader.Data.Header.Message.Slot - minBlobSlotRange
	}
	pruneMarginSlotRange := utils.Config.BlobIndexer.PruneMarginEpochs * uint64(spec.Data.SlotsPerEpoch)
	if minBlobSlot > pruneMarginSlotRange {
		minBlobSlot = minBlobSlot - pruneMarginSlotRange
	}
	return minBlobSlot
}

func calculateStartSlot(status *BlobIndexerStatus, spec *constypes.StandardSpecResponse) (uint64, uint64) {
	denebForkSlot := *spec.Data.DenebForkEpoch * uint64(spec.Data.SlotsPerEpoch)
	startSlot := status.LastIndexedFinalizedSlot + 1
	if status.LastIndexedFinalizedSlot <= denebForkSlot {
		startSlot = denebForkSlot
	}
	return startSlot, denebForkSlot
}

func reportBlobIndexerStatus() {
	if !utils.Config.BlobIndexer.DisableStatusReports {
		services.ReportStatus("blobindexer", "Running", nil)
	}
}

func calculateEstimatedTimeToHeadDuration(estimatedTimeToHead float64) time.Duration {
	estimatedTimeToHeadDuration := time.Duration(estimatedTimeToHead) * time.Second
	return estimatedTimeToHeadDuration
}

func calculateEstimatedTimeToHead(headHeader *constypes.StandardBeaconHeaderResponse, batchStart uint64, slotsPerSecond float64) float64 {
	estimatedTimeToHead := (float64(headHeader.Data.Header.Message.Slot) - float64(batchStart)) / slotsPerSecond
	return estimatedTimeToHead
}

func calculateBlobsPerSecond(batchBlobsIndexed *atomic.Int64, batchStartTs time.Time) float64 {
	blobsPerSecond := float64(batchBlobsIndexed.Load()) / time.Since(batchStartTs).Seconds()
	return blobsPerSecond
}

func calculateSlotsPerSecond(batchEnd, batchStart uint64, batchStartTs time.Time) float64 {
	slotsPerSecond := float64(batchEnd-batchStart) / time.Since(batchStartTs).Seconds()
	return slotsPerSecond
}

func getLastIndexedFinalizedSlot(batchEnd uint64, finalizedHeader *constypes.StandardBeaconHeaderResponse) uint64 {
	lastIndexedFinalizedSlot := uint64(0)
	if batchEnd <= finalizedHeader.Data.Header.Message.Slot {
		lastIndexedFinalizedSlot = batchEnd
	} else {
		lastIndexedFinalizedSlot = finalizedHeader.Data.Header.Message.Slot
	}
	return lastIndexedFinalizedSlot
}
