package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// we wrap all our errors in this codebase

func (d *dashboardData) backfillTask() {
	jobs := []edb.BackfillType{
		edb.BackfillTypeRoi,
		edb.BackfillTypeEBLookup,
		edb.BackfillTypeElectraForkEpochEvents,
		edb.BackfillTypeMissingMissedRewards,
	}
	for _, backfillType := range jobs {
		// fork for every backfill we have to do
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
func (d *dashboardData) handleIncompleteBackfills(t edb.BackfillType) error {
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

func (d *dashboardData) backfillEpochs(t edb.BackfillType, epochs []edb.BackfillMetadata) error {
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
		eg.Go(func() (err error) {
			//d.log.InfoWithFields("doing backfill batch %s", id)
			d.log.InfoWithFields(log.Fields{"backfillType": t, "backfillBatchId": id, "epochRange": []uint64{epochs[0].Epoch, epochs[len(epochs)-1].Epoch}}, "doing backfill batch")

			switch t {
			case edb.BackfillTypeRoi:
				err = edb.BackfillRoi(epochs)
			case edb.BackfillTypeEBLookup:
				err = edb.BackfillEBLookup(epochs)
			case edb.BackfillTypeElectraForkEpochEvents:
				err = d.BackfillElectraForkEpochEvents(epochs)
			case edb.BackfillTypeMissingMissedRewards:
				err = edb.BackfillMissingMissedRewards(epochs)
			default:
				return fmt.Errorf("unknown backfill type %s", t)
			}

			if err != nil {
				d.log.Error(err, "failed to backfill epochs", 0, log.Fields{"epochs": epochs, "type": t})
				return errors.Wrap(err, "failed to transfer epochs")
			}

			now := time.Now()
			for i := range epochs {
				if epochs[i].BackfillName != t {
					return fmt.Errorf("backfill name %s does not match backfill type %s", epochs[i].BackfillName, t)
				}
				if epochs[i].BackfillBatchId == nil {
					return fmt.Errorf("backfill batch id is nil for epoch %v", epochs[i])
				}
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

func (d *dashboardData) handlePendingBackfills(t edb.BackfillType) error {
	pending, err := edb.GetPendingBackfillEpochs(t, utils.Config.DashboardExporter.BackfillAtOnce*utils.Config.DashboardExporter.BackfillInParallel)
	if err != nil {
		return errors.Wrap(err, "failed to get pending backfill epochs")
	}
	if len(pending) == 0 {
		d.log.Debugf("handlePendingBackfills, no pending backfill epochs") // this log messages, and basically all others, should include the backfill type
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

func (d *dashboardData) BackfillElectraForkEpochEvents(epochs []edb.BackfillMetadata) (err error) {
	metricPrefix := string("dashboard_data_exporter_backfill_" + edb.BackfillTypeElectraForkEpochEvents)
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(metricPrefix + "_batch").Observe(time.Since(start).Seconds())
	}()
	// filter out epoch that arent the fork epoch
	forkEpoch := utils.Config.Chain.ClConfig.ElectraForkEpoch
	var forkEpochs []edb.BackfillMetadata
	for _, e := range epochs {
		if e.Epoch == forkEpoch {
			forkEpochs = append(forkEpochs, e)
		}
	}
	if len(forkEpochs) == 0 {
		log.Debugf("no epochs to backfill for %s", edb.BackfillTypeElectraForkEpochEvents)
		return nil
	}
	if len(forkEpochs) > 1 {
		return fmt.Errorf("more than one epoch to backfill for %s", edb.BackfillTypeElectraForkEpochEvents)
	}
	backfillBatchID := epochs[0].BackfillBatchId
	epoch := forkEpochs[0].Epoch

	// reuse the existing stuff to do this, tho obv could be leaner
	nodeData := NewMultiEpochData(1)
	// prefill epochBasedData.epochs
	nodeData.epochBasedData.epochs = []uint64{forkEpochs[0].Epoch}
	eg := &errgroup.Group{}
	eg.Go(func() error {
		return d.fetchEpochValidatorStates(epoch, epoch, &nodeData)
	})
	eg.Go(func() error {
		return d.fetchElectraRemovedExcessBalances(epoch, epoch, &nodeData)
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to fetch validator states for epoch %d: %w", epoch, err)
	}

	// prepare the target data. i hate this as much as you do, but i dont have the time to refactor this rn
	// was never really intended to be used like this anyway
	processedData := make([]types.VDBDataEpochColumns, 1)
	nodeData.epochBasedData.tarIndices = []int{0}
	nodeData.epochBasedData.tarOffsets = []int{0}
	valiCount := len(nodeData.epochBasedData.validatorStates[int64(epoch)].Data)
	processedData[0], err = types.NewVDBDataEpochColumns(valiCount)
	if err != nil {
		return fmt.Errorf("failed to create new VDBDataEpochColumns: %w", err)
	}

	processedData[0].EpochsContained = []uint64{epoch}
	eg = &errgroup.Group{} // docs say we should not reuse a group for different tasks
	eg.Go(func() error {
		// sets epochTimestamp and validatorIndex
		return d.processValidatorStates(&nodeData, &processedData)
	})
	eg.Go(func() error {
		// sets withdrawalAmount and withdrawalCount
		return d.processElectraRemovedExcessBalanceEvents(&nodeData, &processedData)
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to process electra removed excess balance events: %w", err)
	}

	// now write it to the insert sink
	err = db.UltraFastDumpToClickhouse(&processedData[0], "_insert_sink_backfill_electra_fork_epoch_events", backfillBatchID.String())
	if err != nil {
		d.log.Error(err, "failed to insert epochs", 0, log.Fields{"epochs": forkEpochs})
		return errors.Wrap(err, "failed to insert epochs")
	}

	return nil
}
