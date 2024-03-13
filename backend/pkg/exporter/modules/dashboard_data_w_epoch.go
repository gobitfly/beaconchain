package modules

import (
	"context"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

type epochWriter struct {
	*dashboardData
}

func newEpochWriter(d *dashboardData) *epochWriter {
	return &epochWriter{
		dashboardData: d,
	}
}

const PartitionEpochWidth = 3
const retentionBuffer = 8 // todo set to 1.6 buffer

func (d *epochWriter) getRetentionEpochDuration() uint64 {
	return uint64(float64(utils.EpochsPerDay()) / 24 * retentionBuffer)
}

func (d *epochWriter) getPartitionRange(epoch uint64) (uint64, uint64) {
	startOfPartition := epoch / PartitionEpochWidth * PartitionEpochWidth // inclusive
	endOfPartition := startOfPartition + PartitionEpochWidth              // exclusive
	return startOfPartition, endOfPartition
}

func (d *epochWriter) writeEpochData(epoch uint64, data []*validatorDashboardDataRow) error {
	// Create table if needed
	startOfPartition, endOfPartition := d.getPartitionRange(epoch)

	err := d.createEpochPartition(startOfPartition, endOfPartition)
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

		/*
			AttestationsScheduled     int8 //done
			AttestationsExecuted      int8 //done
			AttestationHeadExecuted   int8 //done
			AttestationSourceExecuted int8 //done
			AttestationTargetExecuted int8 //done
		*/

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
			"blocks_el_reward",
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
			"sync_chance",
			"block_chance",
			"attestations_scheduled",
			"attestations_executed",
			"attestation_head_executed",
			"attestation_source_executed",
			"attestation_target_executed",
		}, pgx.CopyFromSlice(len(data), func(i int) ([]interface{}, error) {
			return []interface{}{
				i,
				epoch,
				data[i].AttestationsSourceReward,
				data[i].AttestationsTargetReward,
				data[i].AttestationsHeadReward,
				data[i].AttestationsInactivityReward,
				data[i].AttestationsInclusionsReward,
				data[i].AttestationReward,
				data[i].AttestationsIdealSourceReward,
				data[i].AttestationsIdealTargetReward,
				data[i].AttestationsIdealHeadReward,
				data[i].AttestationsIdealInactivityReward,
				data[i].AttestationsIdealInclusionsReward,
				data[i].AttestationIdealReward,
				data[i].BlockScheduled,
				data[i].BlocksProposed,
				data[i].BlocksClReward,
				data[i].BlocksElReward,
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
				data[i].SyncChance,
				data[i].BlockChance,
				data[i].AttestationsScheduled,
				data[i].AttestationsExecuted,
				data[i].AttestationHeadExecuted,
				data[i].AttestationSourceExecuted,
				data[i].AttestationTargetExecuted,
			}, nil
		}))

		if err != nil {
			return errors.Wrap(err, "error copying data")
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return errors.Wrap(err, "error committing transaction")
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
