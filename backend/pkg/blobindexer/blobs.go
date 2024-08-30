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

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"

	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/coocood/freecache"
	"golang.org/x/sync/errgroup"
)

type BlobIndexer struct {
	S3Client   *s3.Client
	running    bool
	runningMu  *sync.Mutex
	clEndpoint string
	cache      *freecache.Cache
	cl         consapi.Client
}

func NewBlobIndexer() (*BlobIndexer, error) {
	s3Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               utils.Config.BlobIndexer.S3.Endpoint,
			SigningRegion:     "us-east-2",
			HostnameImmutable: true,
		}, nil
	})
	s3Client := s3.NewFromConfig(aws.Config{
		Region: "us-east-2",
		Credentials: credentials.NewStaticCredentialsProvider(
			utils.Config.BlobIndexer.S3.AccessKeyId,
			utils.Config.BlobIndexer.S3.AccessKeySecret,
			"",
		),
		EndpointResolverWithOptions: s3Resolver,
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	bi := &BlobIndexer{
		S3Client:   s3Client,
		runningMu:  &sync.Mutex{},
		clEndpoint: "http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port,
		cache:      freecache.NewCache(1024 * 1024),
		cl:         consapi.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port),
	}
	return bi, nil
}

func (bi *BlobIndexer) Start() {
	bi.runningMu.Lock()
	if bi.running {
		bi.runningMu.Unlock()
		return
	}
	bi.running = true
	bi.runningMu.Unlock()

	log.InfoWithFields(log.Fields{"version": version.Version, "clEndpoint": bi.clEndpoint, "s3Endpoint": utils.Config.BlobIndexer.S3.Endpoint}, "starting blobindexer")
	for {
		err := bi.Index()
		if err != nil {
			log.Error(err, "failed indexing blobs", 0)
		}
		time.Sleep(time.Second * 10)
	}
}

func (bi *BlobIndexer) Index() error {
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
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		headHeader, err = bi.cl.GetBlockHeader("head")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		finalizedHeader, err = bi.cl.GetBlockHeader("finalized")
		if err != nil {
			return err
		}
		return nil
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	nodeDepositNetworkId := uint64(spec.Data.DepositNetworkID)
	if utils.Config.Chain.ClConfig.DepositNetworkID != nodeDepositNetworkId {
		return fmt.Errorf("config.DepositNetworkId != node.DepositNetworkId: %v != %v", utils.Config.Chain.ClConfig.DepositNetworkID, nodeDepositNetworkId)
	}

	status, err := bi.GetIndexerStatus()
	if err != nil {
		return err
	}

	denebForkSlot := utils.Config.Chain.ClConfig.DenebForkEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	startSlot := status.LastIndexedFinalizedSlot + 1
	if status.LastIndexedFinalizedSlot <= denebForkSlot {
		startSlot = denebForkSlot
	}

	if headHeader.Data.Header.Message.Slot <= startSlot {
		return fmt.Errorf("headHeader.Data.Header.Message.Slot <= startSlot: %v < %v", headHeader.Data.Header.Message.Slot, startSlot)
	}

	start := time.Now()
	log.InfoWithFields(log.Fields{"lastIndexedFinalizedSlot": status.LastIndexedFinalizedSlot, "headSlot": headHeader.Data.Header.Message.Slot}, "indexing blobs")
	defer func() {
		log.InfoWithFields(log.Fields{
			"startSlot": startSlot,
			"endSlot":   headHeader.Data.Header.Message.Slot,
			"duration":  time.Since(start),
		}, "finished indexing blobs")
	}()

	batchSize := uint64(100)
	for batchStart := startSlot; batchStart <= headHeader.Data.Header.Message.Slot; batchStart += batchSize {
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
				err := bi.IndexBlobsAtSlot(slot)
				if err != nil {
					return err
				}
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			return err
		}
		if batchEnd <= finalizedHeader.Data.Header.Message.Slot {
			err := bi.PutIndexerStatus(BlobIndexerStatus{
				LastIndexedFinalizedSlot: batchEnd,
			})
			if err != nil {
				return fmt.Errorf("error updating indexer status at slot %v: %w", batchEnd, err)
			}
			log.InfoWithFields(log.Fields{"lastIndexedFinalizedSlot": batchEnd}, "updated indexer status")
		}
	}
	return nil
}

func (bi *BlobIndexer) IndexBlobsAtSlot(slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tGetBlobSidcar := time.Now()

	blobSidecar, err := bi.cl.GetBlobSidecars(slot)
	if err != nil {
		httpErr := network.SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			// no sidecar for this slot
			return nil
		}
		return err
	}
	metrics.TaskDuration.WithLabelValues("blobindexer_get_blob_sidecars").Observe(time.Since(tGetBlobSidcar).Seconds())

	if len(blobSidecar.Data) <= 0 {
		return nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for _, d := range blobSidecar.Data {
		d := d
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			versionedBlobHash := fmt.Sprintf("%#x", utils.VersionedBlobHash(d.KzgCommitment).Bytes())
			key := fmt.Sprintf("blobs/%s", versionedBlobHash)

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
					tS3PutObj := time.Now()
					_, putErr := bi.S3Client.PutObject(gCtx, &s3.PutObjectInput{
						Bucket: &utils.Config.BlobIndexer.S3.Bucket,
						Key:    &key,
						Body:   bytes.NewReader(d.Blob),
						Metadata: map[string]string{
							"slot":              fmt.Sprintf("%d", d.Slot),
							"index":             fmt.Sprintf("%d", d.Index),
							"block_root":        d.BlockRoot.String(),
							"block_parent_root": d.BlockParentRoot.String(),
							"proposer_index":    fmt.Sprintf("%d", d.ProposerIndex),
							"kzg_commitment":    d.KzgCommitment.String(),
							"kzg_proof":         d.KzgProof.String(),
						},
					})
					metrics.TaskDuration.WithLabelValues("blobindexer_put_blob").Observe(time.Since(tS3PutObj).Seconds())
					if putErr != nil {
						return fmt.Errorf("error putting object: %s (%v/%v): %w", key, d.Slot, d.Index, putErr)
					}
					return nil
				}
				return fmt.Errorf("error getting headObject: %s (%v/%v): %w", key, d.Slot, d.Index, err)
			}
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error indexing blobs at slot %v: %w", slot, err)
	}

	return nil
}

func (bi *BlobIndexer) GetIndexerStatus() (*BlobIndexerStatus, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_get_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := "blob-indexer-status.json"
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

func (bi *BlobIndexer) PutIndexerStatus(status BlobIndexerStatus) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("blobindexer_put_indexer_status").Observe(time.Since(start).Seconds())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	key := "blob-indexer-status.json"
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
			"last_indexed_finalized_slot": fmt.Sprintf("%d", status.LastIndexedFinalizedSlot),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

type BlobIndexerStatus struct {
	LastIndexedFinalizedSlot uint64 `json:"last_indexed_finalized_slot"`
	// LastIndexedFinalizedRoot string `json:"last_indexed_finalized_root"`
	// IndexedUnfinalized       map[string]uint64 `json:"indexed_unfinalized"`
}
