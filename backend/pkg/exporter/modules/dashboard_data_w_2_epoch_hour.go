package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
)

type epochToHourAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

const HourAggregateWidth = 9    // todo gnosis
const hourRetentionBuffer = 2.0 // change to 1.6

func newEpochToHourAggregator(d *dashboardData) *epochToHourAggregator {
	return &epochToHourAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

// Assumes no gaps in epochs
func (d *epochToHourAggregator) aggregate1h() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		d.log.Infof("aggregate1h took %v", time.Since(startTime))
	}()

	lastHourExported, err := edb.GetLastExportedHour()
	if err != nil && err != sql.ErrNoRows {
		d.log.Error(err, "failed to get latest dashboard hourly epoch", 0)
		return
	}

	currentEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		d.log.Error(err, "failed to get latest dashboard epoch", 0)
		return
	}

	if lastHourExported.EpochStart == 0 {
		lastHourExported.EpochStart = currentEpoch
	}

	_, currentEndBound := d.getHourAggregateBounds(currentEpoch)

	for epoch := lastHourExported.EpochStart; epoch <= currentEndBound; epoch += HourAggregateWidth {
		boundsStart, boundsEnd := d.getHourAggregateBounds(epoch)
		if lastHourExported.EpochEnd == boundsEnd { // no need to update last hour entry if it is complete
			d.log.Infof("skipping updating last hour entry since it is complete")
			continue
		}

		err = d.aggregate1hSpecific(boundsStart, boundsEnd)
		if err != nil {
			d.log.Error(err, "failed to aggregate 1h", 0)
			return
		}
	}
}

func (d *epochToHourAggregator) getHourAggregateBounds(epoch uint64) (uint64, uint64) {
	offset := utils.GetEpochOffsetGenesis()
	epoch += offset                                                     // offset to utc
	startOfPartition := epoch / HourAggregateWidth * HourAggregateWidth // inclusive
	endOfPartition := startOfPartition + HourAggregateWidth             // exclusive
	return startOfPartition - offset, endOfPartition - offset
}

func (d *epochToHourAggregator) GetHourPartitionRange(epoch uint64) (uint64, uint64) {
	startOfPartition := epoch / (PartitionEpochWidth * HourAggregateWidth) * PartitionEpochWidth * HourAggregateWidth // inclusive
	endOfPartition := startOfPartition + PartitionEpochWidth*HourAggregateWidth                                       // exclusive
	return startOfPartition, endOfPartition
}

func (d *epochToHourAggregator) getHourRetentionDurationEpochs() uint64 {
	return utils.EpochsPerDay() * hourRetentionBuffer
}

func (d *epochToHourAggregator) createHourlyPartition(epochStartFrom, epochStartTo uint64) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS validator_dashboard_data_hourly_%d_%d 
		PARTITION OF validator_dashboard_data_hourly
			FOR VALUES FROM (%[1]d) TO (%[2]d)
		`,
		epochStartFrom, epochStartTo,
	))
	return err
}

func (d *epochToHourAggregator) deleteHourlyPartition(epochStartFrom, epochStartTo uint64) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		DROP TABLE IF EXISTS validator_dashboard_data_hourly_%d_%d
		`,
		epochStartFrom, epochStartTo,
	))

	return err
}

func (d *epochToHourAggregator) aggregate1hSpecific(epochStart, epochEnd uint64) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	partitionStartRange, partitionEndRange := d.GetHourPartitionRange(epochStart)

	err = d.createHourlyPartition(partitionStartRange, partitionEndRange)
	if err != nil {
		return err
	}

	d.log.Infof("aggregating 1h, startEpoch: %d endEpoch: %d", epochStart, epochEnd)

	_, err = tx.Exec(`
		WITH
			end_epoch as (
				SELECT max(epoch) as epoch FROM validator_dashboard_data_epoch where epoch < $2 AND epoch >= $1
			),
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_epoch WHERE epoch = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_epoch WHERE epoch = (SELECT epoch FROM end_epoch)
			),
			aggregate as (
				SELECT 
					validator_index,
					COALESCE(SUM(COALESCE(attestations_source_reward, 0)),0) as attestations_source_reward,
					COALESCE(SUM(COALESCE(attestations_target_reward, 0)),0) as attestations_target_reward,
					COALESCE(SUM(COALESCE(attestations_head_reward, 0)),0) as attestations_head_reward,
					COALESCE(SUM(COALESCE(attestations_inactivity_reward, 0)),0) as attestations_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_inclusion_reward, 0)),0) as attestations_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_reward, 0)),0) as attestations_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_source_reward, 0)),0) as attestations_ideal_source_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_target_reward, 0)),0) as attestations_ideal_target_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_head_reward, 0)),0) as attestations_ideal_head_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inactivity_reward, 0)),0) as attestations_ideal_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inclusion_reward, 0)),0) as attestations_ideal_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_reward, 0)),0) as attestations_ideal_reward,
					COALESCE(SUM(COALESCE(blocks_scheduled, 0)),0) as blocks_scheduled,
					COALESCE(SUM(COALESCE(blocks_proposed, 0)),0) as blocks_proposed,
					COALESCE(SUM(COALESCE(blocks_cl_reward, 0)),0) as blocks_cl_reward,
					COALESCE(SUM(COALESCE(blocks_el_reward, 0)),0) as blocks_el_reward,
					COALESCE(SUM(COALESCE(sync_scheduled, 0)),0) as sync_scheduled,
					COALESCE(SUM(COALESCE(sync_executed, 0)),0) as sync_executed,
					COALESCE(SUM(COALESCE(sync_rewards, 0)),0) as sync_rewards,
					bool_or(slashed) as slashed,
					COALESCE(SUM(COALESCE(deposits_count, 0)),0) as deposits_count,
					COALESCE(SUM(COALESCE(deposits_amount, 0)),0) as deposits_amount,
					COALESCE(SUM(COALESCE(withdrawals_count, 0)),0) as withdrawals_count,
					COALESCE(SUM(COALESCE(withdrawals_amount, 0)),0) as withdrawals_amount,
					COALESCE(SUM(COALESCE(inclusion_delay_sum, 0)),0) as inclusion_delay_sum,
					COALESCE(SUM(COALESCE(sync_chance, 0)),0) as sync_chance,
					COALESCE(SUM(COALESCE(block_chance, 0)),0) as block_chance,
					COALESCE(SUM(COALESCE(attestations_scheduled, 0)),0) as attestations_scheduled,
					COALESCE(SUM(COALESCE(attestations_executed, 0)),0) as attestations_executed,
					COALESCE(SUM(COALESCE(attestation_head_executed, 0)),0) as attestation_head_executed,
					COALESCE(SUM(COALESCE(attestation_source_executed, 0)),0) as attestation_source_executed,
					COALESCE(SUM(COALESCE(attestation_target_executed, 0)),0) as attestation_target_executed,
					COALESCE(SUM(COALESCE(optimal_inclusion_delay_sum, 0)),0) as optimal_inclusion_delay_sum
				FROM validator_dashboard_data_epoch
				WHERE epoch >= $1 AND epoch < $2
				GROUP BY validator_index
			)
			INSERT INTO validator_dashboard_data_hourly (
				epoch_start,
				epoch_end,
				validator_index,
				attestations_source_reward,
				attestations_target_reward,
				attestations_head_reward,
				attestations_inactivity_reward,
				attestations_inclusion_reward,
				attestations_reward,
				attestations_ideal_source_reward,
				attestations_ideal_target_reward,
				attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward,
				attestations_ideal_reward,
				blocks_scheduled,
				blocks_proposed,
				blocks_cl_reward,
				blocks_el_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				balance_start,
				balance_end,
				deposits_count,
				deposits_amount,
				withdrawals_count,
				withdrawals_amount,
				inclusion_delay_sum,
				sync_chance,
				block_chance,
				attestations_scheduled,
				attestations_executed,
				attestation_head_executed,
				attestation_source_executed,
				attestation_target_executed,
				optimal_inclusion_delay_sum
			)
			SELECT 
				$1,
				(SELECT epoch FROM end_epoch) as epoch,
				aggregate.validator_index,
				attestations_source_reward,
				attestations_target_reward,
				attestations_head_reward,
				attestations_inactivity_reward,
				attestations_inclusion_reward,
				attestations_reward,
				attestations_ideal_source_reward,
				attestations_ideal_target_reward,
				attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward,
				attestations_ideal_reward,
				blocks_scheduled,
				blocks_proposed,
				blocks_cl_reward,
				blocks_el_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				COALESCE(balance_start, 0),
				COALESCE(balance_end,0),
				deposits_count,
				deposits_amount,
				withdrawals_count,
				withdrawals_amount,
				inclusion_delay_sum,
				sync_chance,
				block_chance,
				attestations_scheduled,
				attestations_executed,
				attestation_head_executed,
				attestation_source_executed,
				attestation_target_executed,
				optimal_inclusion_delay_sum
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
			ON CONFLICT (epoch_start, validator_index) DO UPDATE SET
				attestations_source_reward = EXCLUDED.attestations_source_reward,
				attestations_target_reward = EXCLUDED.attestations_target_reward,
				attestations_head_reward = EXCLUDED.attestations_head_reward,
				attestations_inactivity_reward = EXCLUDED.attestations_inactivity_reward,
				attestations_inclusion_reward = EXCLUDED.attestations_inclusion_reward,
				attestations_reward = EXCLUDED.attestations_reward,
				attestations_ideal_source_reward = EXCLUDED.attestations_ideal_source_reward,
				attestations_ideal_target_reward = EXCLUDED.attestations_ideal_target_reward,
				attestations_ideal_head_reward = EXCLUDED.attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward = EXCLUDED.attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward = EXCLUDED.attestations_ideal_inclusion_reward,
				attestations_ideal_reward = EXCLUDED.attestations_ideal_reward,
				blocks_scheduled = EXCLUDED.blocks_scheduled,
				blocks_proposed = EXCLUDED.blocks_proposed,
				blocks_cl_reward = EXCLUDED.blocks_cl_reward,
				blocks_el_reward = EXCLUDED.blocks_el_reward,
				sync_scheduled = EXCLUDED.sync_scheduled,
				sync_executed = EXCLUDED.sync_executed,
				sync_rewards = EXCLUDED.sync_rewards,
				slashed = EXCLUDED.slashed,
				balance_start = EXCLUDED.balance_start,
				balance_end = EXCLUDED.balance_end,
				deposits_count = EXCLUDED.deposits_count,
				deposits_amount = EXCLUDED.deposits_amount,
				withdrawals_count = EXCLUDED.withdrawals_count,
				withdrawals_amount = EXCLUDED.withdrawals_amount,
				inclusion_delay_sum = EXCLUDED.inclusion_delay_sum,
				sync_chance = EXCLUDED.sync_chance,
				block_chance = EXCLUDED.block_chance,
				attestations_scheduled = EXCLUDED.attestations_scheduled,
				attestations_executed = EXCLUDED.attestations_executed,
				attestation_head_executed = EXCLUDED.attestation_head_executed,
				attestation_source_executed = EXCLUDED.attestation_source_executed,
				attestation_target_executed = EXCLUDED.attestation_target_executed,
				optimal_inclusion_delay_sum = EXCLUDED.optimal_inclusion_delay_sum,
				epoch_end = EXCLUDED.epoch_end
	`, epochStart, epochEnd)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	minDbEpoch, err := edb.GetOldestDashboardEpoch()
	if err != nil {
		return err
	}

	// Clear old epoch partitions
	var delEpoch uint64
	for i := uint64(0); ; i += PartitionEpochWidth {
		delEpoch = epochStart - d.epochWriter.getRetentionEpochDuration() - i

		startOfPartition, endOfPartition := d.epochWriter.getPartitionRange(delEpoch)
		err := d.epochWriter.deleteEpochPartition(startOfPartition, endOfPartition)
		if err != nil {
			return err
		}
		d.log.Infof("deleted old epoch partition %d_%d", startOfPartition, endOfPartition)

		if delEpoch < minDbEpoch {
			break
		}
	}
	return nil
}
