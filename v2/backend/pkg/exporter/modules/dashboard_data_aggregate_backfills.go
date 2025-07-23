package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// we wrap all our errors in this codebase

func (d *dashboardData) aggregateBackfillTask() {
	jobs := []edb.AggregateBackfillType{
		edb.AggregateBackfillTypeBotchedEpochStart,
	}
	aggregates := []edb.AggregateType{
		edb.AggregateHourly,
		edb.AggregateDaily,
		edb.AggregateWeekly,
		edb.AggregateMonthly,
	}
	for _, backfillType := range jobs {
		exists, err := edb.CheckIfAggregateBackfillExists(backfillType)
		if err != nil {
			d.log.Error(err, "failed to check if aggregate backfill exists", 0, log.Fields{"backfillType": backfillType})
			continue
		}
		if !exists {
			d.log.Debugf("aggregate backfill %s does not exist, skipping", backfillType)
			continue
		}
		// fork for every backfill we have to do
		backfillType := backfillType
		for _, aggregateType := range aggregates {
			aggregateType := aggregateType
			fields := log.Fields{
				"backfillType":  backfillType,
				"aggregateType": aggregateType,
			}
			go func() {
				log.Tracef("starting aggregate backfill for %s", backfillType)
				for {
					start := time.Now()
					// loop to complete incomplete epochs
					err := d.handleIncompleteAggregateBackfills(backfillType, aggregateType)
					if err != nil {
						d.log.Error(err, "failed to handle incomplete aggregate backfills", 0, fields)
						time.Sleep(10 * time.Second)
						continue
					}
					err = d.handlePendingAggregateBackfills(backfillType, aggregateType)
					if err != nil {
						d.log.Error(err, "failed to handle pending aggregate backfills", 0, fields)
						time.Sleep(10 * time.Second)
						continue
					}
					time.Sleep(max(0, 5*time.Second-time.Since(start)))
					count_checked, count_backfilled, count_to_backfill, err := edb.GetAggregateBackfillProgress(backfillType, aggregateType)
					if err != nil {
						d.log.Error(err, "failed to get aggregate backfill progress", 0, fields)
						continue
					}
					metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_aggregate_backfill_%s_aggregate_%s_checked", backfillType, aggregateType)).Set(float64(count_checked))
					metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_aggregate_backfill_%s_aggregate_%s_backfilled", backfillType, aggregateType)).Set(float64(count_backfilled))
					metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_aggregate_backfill_%s_aggregate_%s_to_backfill", backfillType, aggregateType)).Set(float64(count_to_backfill))
				}
			}()
		}
	}
}

func (d *dashboardData) handleIncompleteAggregateBackfills(t edb.AggregateBackfillType, a edb.AggregateType) error {
	incomplete, err := edb.GetIncompleteAggregateBackfillTimestamps(t, a)
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete aggregate backfill timestamps")
	}
	if len(incomplete) == 0 {
		d.log.Debugf("handleIncompleteBackfills, no incomplete aggregate backfill timestamps")
		return nil
	}
	d.log.InfoWithFields(log.Fields{"backfillType": t, "incompleteCount": len(incomplete), "aggregateType": a}, "handleIncompleteBackfills, found incomplete aggregate backfill timestamps")

	err = d.backfillAggregates(t, a, incomplete)
	if err != nil {
		return errors.Wrap(err, "failed to backfill incomplete aggregate backfill timestamps")
	}
	return nil
}

func (d *dashboardData) backfillAggregates(t edb.AggregateBackfillType, a edb.AggregateType, timestamps []edb.AggregateBackfillMetadata) error {
	now := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_aggregate_backfill_%s_overall", t)).Observe(time.Since(now).Seconds())
	}()
	batches := make(map[uuid.UUID][]edb.AggregateBackfillMetadata)
	for _, t := range timestamps {
		if t.BackfillBatchId == nil {
			return fmt.Errorf("backfill batch id is nil for timestamp %s", t.Timestamp)
		}
		id := *t.BackfillBatchId
		batches[id] = append(batches[id], t)
	}
	eg := &errgroup.Group{}
	eg.SetLimit(int(utils.Config.DashboardExporter.AggregateBackfillsIncompleteInParallel))
	for id, batch := range batches {
		batch := batch
		id := id
		eg.Go(func() (err error) {
			//d.log.InfoWithFields("doing backfill batch %s", id)
			d.log.InfoWithFields(log.Fields{"backfillType": t, "backfillBatchId": id, "timestampRange": []time.Time{batch[0].Timestamp, batch[len(batch)-1].Timestamp}}, "doing aggregate backfill batch")

			switch t {
			case edb.AggregateBackfillTypeBotchedEpochStart:
				err = edb.AggregateBackfillBotchedEpochStart(a, batch)
			default:
				return fmt.Errorf("unknown backfill type %s", t)
			}

			if err != nil {
				d.log.Error(err, "failed to backfill aggregate timestamps", 0, log.Fields{"batch": batch, "type": t, "aggregateType": a})
				return errors.Wrap(err, "failed to backfill aggregate timestamps")
			}

			now := time.Now()
			for i := range batch {
				if b := batch[i]; b.BackfillName != t {
					return fmt.Errorf("backfill name %s does not match backfill type %s", b.BackfillName, t)
				}
				if b := batch[i]; b.Aggregation != a {
					return fmt.Errorf("aggregate type %s does not match aggregate type %s", b.Aggregation, a)
				}
				if b := batch[i]; b.DidCheck == nil {
					return fmt.Errorf("did check is nil for timestamp %s", b.Timestamp)
				}
				if t := batch[i]; t.BackfillBatchId == nil {
					return fmt.Errorf("backfill batch id is nil for timestamp %v", t.Timestamp)
				}
				batch[i].SuccessfulBackfill = &now
			}
			err = edb.PushAggregateBackfillMetadata(batch)
			if err != nil {
				d.log.Error(err, "failed to push backfill metadata", 0, log.Fields{"batch": batch, "type": t, "aggregateType": a})
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

func (d *dashboardData) handlePendingAggregateBackfills(t edb.AggregateBackfillType, a edb.AggregateType) error {
	pending, err := edb.GetPendingAggregateBackfillTimestamps(t, a, utils.Config.DashboardExporter.AggregateBackfillsPendingInParallel)
	if err != nil {
		return errors.Wrap(err, "failed to get pending aggregate backfill timestamps")
	}
	if len(pending) == 0 {
		d.log.Debugf("handlePendingBackfills, no pending aggregate backfill timestamps") // this log messages, and basically all others, should include the backfill type
		return nil
	}
	d.log.InfoWithFields(log.Fields{"backfillType": t, "pendingCount": len(pending), "aggregateType": a}, "handlePendingAggregateBackfills, found pending aggregate backfill timestamps")
	// check them
	for i := range pending {
		if b := pending[i]; b.BackfillName != t {
			return fmt.Errorf("backfill name %s does not match backfill type %s", b.BackfillName, t)
		}
		if b := pending[i]; b.Aggregation != a {
			return fmt.Errorf("aggregate type %s does not match aggregate type %s", b.Aggregation, a)
		}
	}
	needsBackfillMap, err := edb.CheckIfAggregateIsBotchedEpochStart(a, pending)
	if err != nil {
		return errors.Wrap(err, "failed to check if aggregate is botched epoch start")
	}
	now := time.Now()
	for i := range pending {
		pending[i].DidCheck = &now
		if needsBackfillMap[pending[i].Timestamp] {
			d.log.Infof("aggregate %s for timestamp %s needs backfill, setting backfill id", a, pending[i].Timestamp)
			backfillUuid := uuid.New()
			pending[i].BackfillBatchId = &backfillUuid
		} else {
			d.log.Infof("aggregate %s for timestamp %s does not need backfill", a, pending[i].Timestamp)
		}
	}

	err = edb.PushAggregateBackfillMetadata(pending)
	if err != nil {
		return errors.Wrap(err, "failed to push aggregate backfill metadata")
	}

	return nil
}
