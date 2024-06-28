package modules

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
)

type epochToHourAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

// How long aggregated hours will remain in the database is defined in getHourRetentionDurationEpochs.
// For ETH mainnet this will be 225 epochs, as 225 epochs is exactly the range we need in the day table (equals 1 day).
// This buffer can be used to increase or decrease from that 225 epoch target. A value of 1 will keep exactly those 25 (225 / 9) needed hour aggregations in the database.
const hourRetentionBuffer = 1.1 // do not go below 1

func getHourAggregateWidth() uint64 {
	return utils.EpochsPerDay() / 24
}

func newEpochToHourAggregator(d *dashboardData) *epochToHourAggregator {
	return &epochToHourAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

func (d *epochToHourAggregator) clearOldHourAggregations(removeBelowEpoch int64) error {
	partitions, err := edb.GetPartitionNamesOfTable(edb.HourWriterTableName)
	if err != nil {
		return errors.Wrap(err, "failed to get partitions")
	}

	for _, partition := range partitions {
		epochFrom, epochTo, err := parseEpochRange(fmt.Sprintf(`%s_(\d+)_(\d+)`, edb.HourWriterTableName), partition)
		if err != nil {
			return errors.Wrap(err, "failed to parse epoch range")
		}

		if int64(epochTo) < removeBelowEpoch {
			d.mutex.Lock()
			err := d.deleteHourlyPartition(epochFrom, epochTo)
			d.log.Infof("Deleted old hourly partition %d-%d", epochFrom, epochTo)
			d.mutex.Unlock()
			if err != nil {
				return errors.Wrap(err, "failed to delete hourly partition")
			}
		}
	}

	return nil
}

// Assumes no gaps in epochs
func (d *epochToHourAggregator) aggregate1h(currentExportedEpoch uint64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	d.log.Info("aggregating 1h")
	defer func() {
		d.log.Infof("aggregate 1h took %v", time.Since(startTime))
		metrics.TaskDuration.WithLabelValues("exporter_v2dash_agg_1h").Observe(time.Since(startTime).Seconds())
	}()

	lastHourExported, err := edb.GetLastExportedHour()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest dashboard hourly epoch")
	}

	differenceToCurrentEpoch := currentExportedEpoch + 1 - lastHourExported.EpochEnd // epochEnd is excl hence the +1

	if differenceToCurrentEpoch > d.getHourRetentionDurationEpochs() {
		d.log.Warnf("difference to current epoch is larger than retention duration, skipping for now: %v", differenceToCurrentEpoch)
		return nil
	}

	gaps, err := edb.GetMissingEpochsBetween(int64(lastHourExported.EpochEnd), int64(currentExportedEpoch+1))
	if err != nil {
		return errors.Wrap(err, "failed to get dashboard epoch gaps")
	}

	if len(gaps) > 0 {
		return fmt.Errorf("gaps in dashboard epoch, skipping for now: %v", gaps) // sanity, this should never happen
	}

	_, currentEndBound := getHourAggregateBounds(currentExportedEpoch)

	for epoch := lastHourExported.EpochStart; epoch <= currentEndBound; epoch += getHourAggregateWidth() {
		boundsStart, boundsEnd := getHourAggregateBounds(epoch)
		if lastHourExported.EpochEnd == boundsEnd { // no need to update last hour entry if it is complete
			d.log.Infof("skipping updating last hour entry since it is complete")
			continue
		}

		// no need to aggregate epoch data that hasn't been exported yet
		if boundsEnd > currentEndBound {
			continue
		}

		// define start bounds as lastHourExported.EpochEnd for first iteration
		if epoch == lastHourExported.EpochStart {
			boundsStart = lastHourExported.EpochEnd
		}

		// scope down to max currentExportedEpoch (since epoch data is inclusive, add 1)
		if currentExportedEpoch+1 >= boundsStart && currentExportedEpoch+1 < boundsEnd {
			boundsEnd = currentExportedEpoch + 1
		}

		err = d.aggregate1hWithBounds(boundsStart, boundsEnd)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate 1h")
		}
	}

	return nil
}

// Returns the epoch_start and epoch_end (the epoch bounds of an hourly aggregation) for a given epoch.
// epoch_start is inclusive, epoch_end is exclusive.
func getHourAggregateBounds(epoch uint64) (uint64, uint64) {
	offset := utils.GetEpochOffsetGenesis()
	epoch += offset                                                               // offset to utc
	startOfPartition := epoch / getHourAggregateWidth() * getHourAggregateWidth() // inclusive
	endOfPartition := startOfPartition + getHourAggregateWidth()                  // exclusive
	if startOfPartition < offset {
		startOfPartition = offset
	}
	return startOfPartition - offset, endOfPartition - offset
}

func (d *epochToHourAggregator) GetHourPartitionRange(epoch uint64) (uint64, uint64) {
	startOfPartition := epoch / (PartitionEpochWidth * getHourAggregateWidth()) * PartitionEpochWidth * getHourAggregateWidth() // inclusive
	endOfPartition := startOfPartition + PartitionEpochWidth*getHourAggregateWidth()                                            // exclusive
	return startOfPartition, endOfPartition
}

func (d *epochToHourAggregator) getHourRetentionDurationEpochs() uint64 {
	return uint64(float64(utils.EpochsPerDay()) * hourRetentionBuffer)
}

func (d *epochToHourAggregator) createHourlyPartition(epochStartFrom, epochStartTo uint64) error {
	_, err := db.AlloyWriter.ExecContext(func() context.Context {
		a, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Minute))
		return a
	}(), fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[3]s_%[1]d_%[2]d 
		PARTITION OF %[3]s
			FOR VALUES FROM (%[1]d) TO (%[2]d)
		`,
		epochStartFrom, epochStartTo, edb.HourWriterTableName,
	))
	return err
}

func (d *epochToHourAggregator) deleteHourlyPartition(epochStartFrom, epochStartTo uint64) error {
	_, err := db.AlloyWriter.ExecContext(func() context.Context {
		a, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Minute))
		return a
	}(), fmt.Sprintf(`
		DROP TABLE IF EXISTS %s_%d_%d
		`,
		edb.HourWriterTableName, epochStartFrom, epochStartTo,
	))

	return err
}

// epochStart incl, epochEnd excl
func (d *epochToHourAggregator) aggregate1hWithBounds(epochStart, epochEnd uint64) error {
	tx, err := db.AlloyWriter.BeginTxx(func() context.Context {
		a, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Minute))
		return a
	}(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	partitionStartRange, partitionEndRange := d.GetHourPartitionRange(epochStart)

	err = d.createHourlyPartition(partitionStartRange, partitionEndRange)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create hourly partition, startRange: %d, endRange: %d", partitionStartRange, partitionEndRange))
	}

	boundsStart, _ := getHourAggregateBounds(epochStart)

	if epochStart == partitionStartRange && debugAddToColumnEngine {
		err = edb.AddToColumnEngineAllColumns(fmt.Sprintf("%s_%d_%d", edb.HourWriterTableName, partitionStartRange, partitionEndRange))
		if err != nil {
			d.log.Warnf("Failed to add epoch to column engine: %v", err)
		}
	}

	d.log.Infof("aggregating 1h with bounds, startEpoch: %d endEpoch: %d", epochStart, epochEnd)

	err = AddToRollingCustom(tx, CustomRolling{
		StartEpoch:           epochStart,
		EndEpoch:             epochEnd,
		StartBoundEpoch:      int64(boundsStart),
		TableFrom:            edb.EpochWriterTableName,
		TableTo:              edb.HourWriterTableName,
		TableFromEpochColumn: "epoch",
		Log:                  d.log,
		TailBalancesQuery: fmt.Sprintf(`balance_starts as (
				SELECT validator_index, balance_start FROM %s WHERE epoch = $3
		),`, edb.EpochWriterTableName),
		TailBalancesJoinQuery:         `LEFT JOIN balance_starts ON aggregate_head.validator_index = balance_starts.validator_index`,
		TailBalancesInsertColumnQuery: `balance_start,`,
		TableConflict:                 "(epoch_start, validator_index)",
	})

	if err != nil {
		return errors.Wrap(err, "failed to insert hourly data")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}
