package modules

import (
	"context"
	"fmt"
	"sync"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

type epochWriter struct {
	*dashboardData
	mutex *sync.Mutex
}

func newEpochWriter(d *dashboardData) *epochWriter {
	return &epochWriter{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

// How wide each table partition is in epochs
const PartitionEpochWidth = 3

// How long epochs will remain in the database is defined in getRetentionEpochDuration.
// For ETH mainnet this will be 9 epochs, as 9 epochs is exactly the range we need in the hour table (roughly one hour).
// This buffer can be used to increase or decrease from that 9 epoch target. A value of 1 will keep exactly those 9 needed epochs in the database.
const retentionBuffer = 1.1 // do not go below 1

func (d *epochWriter) getRetentionEpochDuration() uint64 {
	return uint64(float64(utils.EpochsPerDay()) / 24 * retentionBuffer)
}

func (d *epochWriter) getPartitionRange(epoch uint64) (uint64, uint64) {
	startOfPartition := epoch / PartitionEpochWidth * PartitionEpochWidth // inclusive
	endOfPartition := startOfPartition + PartitionEpochWidth              // exclusive
	return startOfPartition, endOfPartition
}

func (d *epochWriter) clearOldEpochs(removeBelowEpoch int64) error {
	if debugSkipOldEpochClear {
		return nil
	}

	partitions, err := edb.GetPartitionNamesOfTable("validator_dashboard_data_epoch")
	if err != nil {
		return errors.Wrap(err, "failed to get partitions")
	}

	for _, partition := range partitions {
		epochFrom, epochTo, err := parseEpochRange(`validator_dashboard_data_epoch_(\d+)_(\d+)`, partition)
		if err != nil {
			return errors.Wrap(err, "failed to parse epoch range")
		}

		if int64(epochTo) < removeBelowEpoch {
			d.mutex.Lock()
			err := d.deleteEpochPartition(epochFrom, epochTo)
			d.log.Infof("Deleted old epoch partition %d-%d", epochFrom, epochTo)
			d.mutex.Unlock()
			if err != nil {
				return errors.Wrap(err, "failed to delete epoch partition")
			}
		}
	}

	return nil
}

func (d *epochWriter) WriteEpochData(epoch uint64, data []*validatorDashboardDataRow) error {
	// Create table if needed
	startOfPartition, endOfPartition := d.getPartitionRange(epoch)

	d.mutex.Lock()
	err := d.createEpochPartition(startOfPartition, endOfPartition)
	if epoch == startOfPartition && debugAddToColumnEngine {
		err = edb.AddToColumnEngineAllColumns(fmt.Sprintf("validator_dashboard_data_epoch_%d_%d", startOfPartition, endOfPartition))
		if err != nil {
			d.log.Warnf("Failed to add epoch to column engine: %v", err)
		}
	}
	d.mutex.Unlock()
	if err != nil {
		return errors.Wrap(err, "failed to create epoch partition")
	}

	conn, err := db.AlloyWriter.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())

		if err != nil {
			return errors.Wrap(err, "error starting transaction")
		}

		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				d.log.Error(err, "error rolling back transaction", 0)
			}
		}()

		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"validator_dashboard_data_epoch"}, []string{
			"validator_index",
			"epoch",
			"attestations_source_reward",
			"attestations_target_reward",
			"attestations_head_reward",
			"attestations_inactivity_reward",
			"attestations_inclusion_reward",
			"attestations_reward",
			"attestations_ideal_source_reward",
			"attestations_ideal_target_reward",
			"attestations_ideal_head_reward",
			"attestations_ideal_inactivity_reward",
			"attestations_ideal_inclusion_reward",
			"attestations_ideal_reward",
			"blocks_scheduled",
			"blocks_proposed",
			"blocks_cl_reward",
			"sync_scheduled",
			"sync_executed",
			"sync_rewards",
			"slashed",
			"balance_start",
			"balance_end",
			"deposits_count",
			"deposits_amount",
			"withdrawals_count",
			"withdrawals_amount",
			"inclusion_delay_sum",
			"blocks_expected",
			"attestations_scheduled",
			"attestations_executed",
			"attestation_head_executed",
			"attestation_source_executed",
			"attestation_target_executed",
			"optimal_inclusion_delay_sum",
			"slashed_by",
			"slashed_violation",
			"slasher_reward",
			"last_executed_duty_epoch",
			"blocks_cl_attestations_reward",
			"blocks_cl_sync_aggregate_reward",
			"sync_committees_expected",
		}, pgx.CopyFromSlice(len(data), func(i int) ([]interface{}, error) {
			return []interface{}{
				i,
				epoch,
				data[i].AttestationsSourceReward,
				data[i].AttestationsTargetReward,
				data[i].AttestationsHeadReward,
				data[i].AttestationsInactivityPenalty,
				data[i].AttestationsInclusionsReward,
				data[i].AttestationReward,
				data[i].AttestationsIdealSourceReward,
				data[i].AttestationsIdealTargetReward,
				data[i].AttestationsIdealHeadReward,
				data[i].AttestationsIdealInactivityPenalty,
				data[i].AttestationsIdealInclusionsReward,
				data[i].AttestationIdealReward,
				data[i].BlockScheduled,
				data[i].BlocksProposed,
				data[i].BlocksClReward,
				data[i].SyncScheduled,
				data[i].SyncExecuted,
				data[i].SyncReward,
				data[i].Slashed,
				data[i].BalanceStart,
				data[i].BalanceEnd,
				data[i].DepositsCount,
				data[i].DepositsAmount,
				data[i].WithdrawalsCount,
				data[i].WithdrawalsAmount,
				data[i].InclusionDelaySum,
				data[i].BlocksExpected,
				data[i].AttestationsScheduled,
				data[i].AttestationsExecuted,
				data[i].AttestationHeadExecuted,
				data[i].AttestationSourceExecuted,
				data[i].AttestationTargetExecuted,
				data[i].OptimalInclusionDelay,
				data[i].SlashedBy,
				data[i].SlashedViolation,
				data[i].SlasherRewards,
				data[i].LastSubmittedDutyEpoch,
				data[i].BlocksClAttestestationsReward,
				data[i].BlocksClSyncAggregateReward,
				data[i].SyncCommitteesExpected,
			}, nil
		}))

		if err != nil {
			return errors.Wrap(err, "error copying data")
		}

		err = tx.Commit(context.Background())
		if err != nil {
			if !utils.IsDuplicatedKeyError(err) {
				return errors.Wrap(err, "error committing transaction")
			}
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "error writing data")
	}

	return nil
}

func (d *epochWriter) createEpochPartition(epochFrom, epochTo uint64) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS validator_dashboard_data_epoch_%d_%d 
		PARTITION OF validator_dashboard_data_epoch
			FOR VALUES FROM (%[1]d) TO (%[2]d)
		`,
		epochFrom, epochTo,
	))
	return err
}

func (d *epochWriter) deleteEpochPartition(epochFrom, epochTo uint64) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		DROP TABLE IF EXISTS validator_dashboard_data_epoch_%d_%d
		`,
		epochFrom, epochTo,
	))

	return err
}
