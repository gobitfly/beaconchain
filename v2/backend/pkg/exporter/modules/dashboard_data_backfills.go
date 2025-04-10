package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/exporter/db"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// we wrap all our errors in this codebase

func (d *dashboardData) roiBackfillTask() {
	jobs := []edb.BackfillType{
		edb.BackfillTypeRoi,
		edb.BackfillTypeEBLookup,
	}
	for _, backfillType := range jobs {
		// fork for every backfill we have to do
		backfillType := backfillType
		go func() {
			log.Tracef("starting backfill for %s", backfillType)
			for {
				// loop to complete incomplete epochs
				err := d.handleIncompleteBackfills(backfillType)
				if err != nil {
					d.log.Error(err, "failed to handle incomplete backfills", 0, log.Fields{"backfillType": backfillType})
					time.Sleep(10 * time.Second)
					continue
				}
				err = d.handlePendingBackfills(backfillType)
				if err != nil {
					d.log.Error(err, "failed to handle pending backfills", 0, log.Fields{"backfillType": backfillType})
					time.Sleep(10 * time.Second)
					continue
				}
				time.Sleep(5 * time.Second)
				epochsBackfilled, epochsToBackfill, err := edb.GetBackfillProgress(backfillType)
				if err != nil {
					d.log.Error(err, "failed to get backfill progress", 0)
					continue
				}
				metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_backfill_%s_backfilled", backfillType)).Set(float64(epochsBackfilled))
				metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_backfill_%s_to_backfill", backfillType)).Set(float64(epochsToBackfill))
			}
		}()
	}
}
func (d *dashboardData) handleIncompleteBackfills(t db.BackfillType) error {
	incomplete, err := edb.GetIncompleteBackfillEpochs(t)
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete backfill epochs")
	}
	if len(incomplete) == 0 {
		d.log.Debugf("handleIncompleteBackfills, no incomplete backfill epochs")
		return nil
	}
	d.log.InfoWithFields(log.Fields{"backfillType": t, "incompleteCount": len(incomplete)}, "handleIncompleteBackfills, found incomplete backfill epochs")

	err = d.backfillEpochs(t, incomplete)
	if err != nil {
		return errors.Wrap(err, "failed to backfill incomplete epochs")
	}
	return nil
}

func (d *dashboardData) backfillEpochs(t db.BackfillType, epochs []edb.BackfillMetadata) error {
	now := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_backfill_%s_overall", t)).Observe(time.Since(now).Seconds())
	}()
	backfillBatchEpochs := make(map[uuid.UUID][]edb.BackfillMetadata)
	for _, e := range epochs {
		if e.BackfillBatchId == nil {
			return fmt.Errorf("backfill batch id is nil for epoch %v", e)
		}
		id := *e.BackfillBatchId
		backfillBatchEpochs[id] = append(backfillBatchEpochs[id], e)
	}
	eg := &errgroup.Group{}
	eg.SetLimit(int(utils.Config.DashboardExporter.BackfillInParallel))
	for id, epochs := range backfillBatchEpochs {
		epochs := epochs
		id := id
		eg.Go(func() (err error) {
			//d.log.InfoWithFields("doing backfill batch %s", id)
			d.log.InfoWithFields(log.Fields{"backfillType": t, "backfillBatchId": id, "epochRange": []uint64{epochs[0].Epoch, epochs[len(epochs)-1].Epoch}}, "doing backfill batch")

			switch t {
			case edb.BackfillTypeRoi:
				err = edb.BackfillRoi(epochs)
			case edb.BackfillTypeEBLookup:
				err = edb.BackfillEBLookup(epochs)
			default:
				return fmt.Errorf("unknown backfill type %s", t)
			}

			if err != nil {
				d.log.Error(err, "failed to backfill epochs", 0, log.Fields{"epochs": epochs, "type": t})
				return errors.Wrap(err, "failed to transfer epochs")
			}

			now := time.Now()
			for i := range epochs {
				epochs[i].SuccessfulBackfill = &now
			}
			err = edb.PushBackfillMetadata(epochs)
			if err != nil {
				d.log.Error(err, "failed to push backfill metadata", 0, log.Fields{"epochs": epochs, "type": t})
				return errors.Wrap(err, "failed to push backfill metadata")
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to complete transfer batches")
	}
	return nil
}

func (d *dashboardData) handlePendingBackfills(t db.BackfillType) error {
	pending, err := edb.GetPendingBackfillEpochs(t, utils.Config.DashboardExporter.BackfillAtOnce*utils.Config.DashboardExporter.BackfillInParallel)
	if err != nil {
		return errors.Wrap(err, "failed to get pending backfill epochs")
	}
	if len(pending) == 0 {
		d.log.Debugf("handlePendingBackfills, no pending backfill epochs")
		return nil
	}
	d.log.InfoWithFields(log.Fields{"backfillType": t, "pendingCount": len(pending)}, "handlePendingBackfills, found pending backfill epochs")

	// allocate transfer batch ids
	batchIds := make([]uuid.UUID, 0)
	for i := range pending {
		if i%int(utils.Config.DashboardExporter.BackfillAtOnce) == 0 {
			id := uuid.New()
			batchIds = append(batchIds, id)
		}
		pending[i].BackfillBatchId = &batchIds[len(batchIds)-1]
	}

	err = edb.PushBackfillMetadata(pending)
	if err != nil {
		return errors.Wrap(err, "failed to push backfill metadata")
	}

	err = d.backfillEpochs(t, pending)
	if err != nil {
		return errors.Wrap(err, "failed to transfer pending epochs")
	}
	return nil
}
