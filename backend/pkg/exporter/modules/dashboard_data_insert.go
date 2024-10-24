package modules

import (
	"fmt"
	"slices"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

// | handleIncompleteInserts
// | - GetIncompleteInsertEpochs
// | - FetchEpochs A
// | - PushEpochs B
// | - PushEpochMetadata (successful_insert) C
// | handlePendingInserts
// | - GetPendingInsertEpochs
// | - allocate insert batch ids
// | - PushEpochMetadata (insert_batch_id)
// | - FetchEpochs A
// | - PushEpochs B
// | - PushEpochMetadata (successful_insert) C

func (d *dashboardData) insertTask() {
	for {
		// loop to complete incomplete epochs
		err := d.handleIncompleteInserts()
		if err != nil {
			d.log.Error(err, "failed to handle incomplete inserts", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		err = d.handlePendingInserts()
		if err != nil {
			d.log.Error(err, "failed to handle pending inserts", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

func (d *dashboardData) handleIncompleteInserts() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_handle_incomplete_inserts").Observe(time.Since(start).Seconds())
	}()
	// get latest unsafe epoch
	latestUnsafeEpoch, err := edb.GetLatestUnsafeEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get latest unsafe epoch")
	}
	metrics.State.WithLabelValues("dashboard_data_exporter_latest_unsafe_epoch").Set(float64(latestUnsafeEpoch))
	defer func() {
		latestUnsafeEpoch, err := edb.GetLatestUnsafeEpoch()
		if err != nil {
			d.log.Error(err, "failed to get latest unsafe epoch", 0)
			return
		}
		metrics.State.WithLabelValues("dashboard_data_exporter_latest_unsafe_epoch").Set(float64(latestUnsafeEpoch))
	}()

	incomplete, err := edb.GetIncompleteInsertEpochs()
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete insert epochs")
	}
	if len(incomplete) == 0 {
		d.log.Debugf("handleIncompleteInserts, no incomplete insert epochs")
		return nil
	}
	d.log.Infof("handleIncompleteInserts, found %d incomplete insert epochs", len(incomplete))
	// we do grouping here even tho fetchAndInsertEpochs does grouping again
	// this is because we can in theory have incomplete epochs that aren't next to each other
	// and fetchAndInsert wants to do only one fetch of all the epochs
	// also means incomplete uses lower memory than pending, which seems like a sane option
	insertBatchEpochs := make(map[uuid.UUID][]edb.EpochMetadata)
	for _, e := range incomplete {
		if e.InsertBatchID == nil {
			return fmt.Errorf("insert batch id is nil for epoch %v", e)
		}
		id := *e.InsertBatchID
		insertBatchEpochs[id] = append(insertBatchEpochs[id], e)
	}
	for _, epochs := range insertBatchEpochs {
		err = d.fetchAndInsertEpochs(epochs)
		// fail as soon as we run into an error. outer loop will call us again and
		// since the insert id (likely) isnt marked as successful, we will automagically try again
		if err != nil {
			return fmt.Errorf("failed to fetch and insert epochs: %v", err)
		}
	}

	return nil
}

var FetchAtOnceLimit int64 = 2
var InsertAtOnceLimit int64 = 2
var InsertInParallel int64 = 2 // up to 3 parallel inserts

func (d *dashboardData) handlePendingInserts() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_handle_pending_inserts").Observe(time.Since(start).Seconds())
	}()

	safeEpoch := d.latestSafeEpoch.Load()
	pending, err := edb.GetPendingInsertEpochs(safeEpoch, FetchAtOnceLimit)
	if err != nil {
		return errors.Wrap(err, "failed to get pending insert epochs")
	}
	if len(pending) == 0 {
		d.log.Debugf("handlePendingInserts, no pending insert epochs")
		return nil
	}
	d.log.Infof("handlePendingInserts, found %d pending insert epochs", len(pending))

	// allocate insert batch ids
	batchIDS := make([]uuid.UUID, 0)
	for i := range pending {
		if int64(i)%InsertAtOnceLimit == 0 {
			id := uuid.New()
			batchIDS = append(batchIDS, id)
		}
		pending[i].InsertBatchID = &batchIDS[len(batchIDS)-1]
	}

	err = edb.PushEpochMetadata(pending)
	if err != nil {
		return errors.Wrap(err, "failed to push epoch metadata")
	}

	// fetch and insert in parallel. do a single fetch for all epochs
	err = d.fetchAndInsertEpochs(pending)
	if err != nil {
		return errors.Wrap(err, "failed to fetch and insert pending epochs")
	}

	return nil
}

// a-c func
func (d *dashboardData) fetchAndInsertEpochs(epochs []edb.EpochMetadata) error {
	// sanity check that epochs are in order, have no gaps
	slices.SortFunc(epochs, func(a, b edb.EpochMetadata) int {
		return int(a.Epoch) - int(b.Epoch)
	})
	// check for gaps
	for i := 1; i < len(epochs); i++ {
		if epochs[i].Epoch-epochs[i-1].Epoch != 1 {
			return fmt.Errorf("epoch gap between %d and %d", epochs[i-1].Epoch, epochs[i].Epoch)
		}
	}
	insertBatchEpochs := make(map[uuid.UUID][]edb.EpochMetadata)
	for _, e := range epochs {
		if e.InsertBatchID == nil {
			return fmt.Errorf("insert batch id is nil for epoch %v", e)
		}
		id := *e.InsertBatchID
		insertBatchEpochs[id] = append(insertBatchEpochs[id], e)
	}

	// fetch
	rawData := NewMultiEpochData(len(epochs))
	err := d.getDataForEpochRange(epochs[0].Epoch, epochs[len(epochs)-1].Epoch, &rawData)
	if err != nil {
		return errors.Wrap(err, "failed to get data for epochs")
	}

	// process
	processedData := make([]types.VDBDataEpochColumns, len(maps.Keys(insertBatchEpochs))) // grouped by insert batch id
	err = d.processRunner(&rawData, &processedData, epochs)
	if err != nil {
		return errors.Wrap(err, "failed to process data")
	}
	// insert step wooho
	eg := &errgroup.Group{}
	eg.SetLimit(int(InsertInParallel))
	// loop processed data, its grouped by insert batch id already and ready to be inserted
	for i := range processedData {
		data := processedData[i]
		insertId := data.InsertBatchID[0]
		epochs, ok := insertBatchEpochs[insertId]
		if !ok {
			return fmt.Errorf("insert batch id %s not found in insert batch epochs", insertId)
		}
		eg.Go(func() error {
			d.log.Infof("doing insert batch %s", insertId)
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_insert_batch").Observe(time.Since(now).Seconds())
			}()
			err := db.UltraFastDumpToClickhouse(&data, edb.EpochWriterSink, insertId.String())
			if err != nil {
				d.log.Error(err, "failed to insert epochs", 0, log.Fields{"epochs": data})
				return errors.Wrap(err, "failed to insert epochs")
			}
			// mark as successful
			sfts := time.Now()
			for i := range epochs {
				epochs[i].SuccessfulInsert = &sfts
			}

			err = edb.PushEpochMetadata(epochs)
			if err != nil {
				d.log.Error(err, "failed to push epoch metadata", 0, log.Fields{"epochs": insertBatchEpochs[insertId]})
				return errors.Wrap(err, "failed to push epoch metadata")
			}
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to insert all epochs")
	}
	return nil
}
