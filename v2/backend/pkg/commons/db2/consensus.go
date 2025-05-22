package db2

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ConsensusRepository interface {
	GetLatestEpoch() (uint64, error)
	GetDepositsCountForBlockSlot() (uint64, error)
	SaveBlockDeposits(validatorIndex uint64, pubkey []byte, withdrawalCredentials []byte, balance uint64) error
	UpdateBlockDepositsSignature() error
	UpdateBlockDepositCount(count int) error
	SaveNetworkLivenessData(head *types.ChainHead) error
	GetNetworkLivenessPreviousHeadEpoch() (uint64, error)
	GetRelays() ([]types.Relay, error)
	UpdateRelay(tagID, endpoint string) error
	UpdateRelayLastExportTry(tagID, endpoint string) error
	UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error
	GetFirstRelayBlock(tagID string) (types.RelayBlock, error)
	GetLastRelayBlock(tagID string) (types.RelayBlock, error)
	SaveBlockTagsAndRelays(tagID string, payload types.BidTrace) error
	GetSyncCommitteesCountPerValidator() (uint64, error)
	GetTotalPeriodSyncCommitteesCountPerValidator() (uint64, error)
	GetCountSoFarSyncCommitteesCountPerValidator(period uint64) (float64, error)
	SaveSyncCommitteesCount(period uint64, count float64) error
	GetEpochValidatorsCount(epoch uint64) (uint64, error)
	GetLatestFinalizedEpoch() (uint64, error)
	SaveSyncCommitteeData(data []types.SyncCommittee) error
	GetSyncCommitteesPeriods() ([]uint64, error)
	UpdatePubkeyTags() error
	SavePendingDepositsQueue(pendingDeposits []types.PendingDeposit) error
	GetLastIndexedConsensusLayerEventsEpoch() (int64, error)
	GetConsensusLayerEventsOfType(slot uint64, eventName types.ConsensusLayerEventName) ([]types.ConsensusLayerEvent, error)
	GetWriterTx() (*sqlx.Tx, error)
	GetConsensusLayerEventsForFilters(filters []types.ConsensusLayerEventFilter) ([]types.ConsensusLayerEvent, error)
	UpdateConsensusLayerExporterMetadata(*sqlx.Tx, []types.EpochBlockRoots) error
}

type ConsensusDB struct {
	ReaderDb *sqlx.DB
	WriterDb *sqlx.DB
}

func NewConsensusRepository(reader, writer *sqlx.DB) *ConsensusDB {
	return &ConsensusDB{
		ReaderDb: reader,
		WriterDb: writer,
	}
}

// GetLatestEpoch will return the latest epoch from the database
func (c *ConsensusDB) GetLatestEpoch() (uint64, error) {
	var epoch uint64
	err := c.WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %w", err)
	}

	return epoch, nil
}

func (c *ConsensusDB) GetDepositsCountForBlockSlot() (uint64, error) {
	var count uint64
	err := c.WriterDb.Get(&count, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *ConsensusDB) SaveBlockDeposits(vIndex uint64, vPubkey, vWithdrawalCredentials []byte, vBalance uint64) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
	VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
		vIndex, vPubkey, vWithdrawalCredentials, vBalance, []byte{0x0},
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) UpdateBlockDepositsSignature() error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	// hydrate the eth1 deposit signature for all genesis validators that have a corresponding eth1 deposit
	_, err = tx.Exec(`
		UPDATE blocks_deposits
		SET signature = a.signature
		FROM (
			SELECT DISTINCT ON(publickey) publickey, signature
			FROM eth1_deposits
			WHERE valid_signature = true) AS a
		WHERE block_slot = 0 AND blocks_deposits.publickey = a.publickey AND blocks_deposits.signature = '\x'`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) UpdateBlockDepositCount(count int) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", count)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) SaveNetworkLivenessData(head *types.ChainHead) error {
	_, err := c.WriterDb.Exec(`
        INSERT INTO network_liveness (ts, headepoch, finalizedepoch, justifiedepoch, previousjustifiedepoch)
        VALUES (NOW(), $1, $2, $3, $4)`,
		head.HeadEpoch, head.FinalizedEpoch, head.JustifiedEpoch, head.PreviousJustifiedEpoch)

	return err
}

func (c *ConsensusDB) GetNetworkLivenessPreviousHeadEpoch() (uint64, error) {
	var headEpoch uint64
	err := c.WriterDb.Get(&headEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	return headEpoch, err
}

func (c *ConsensusDB) GetRelays() ([]types.Relay, error) {
	var relays []types.Relay
	err := c.ReaderDb.Select(&relays, `
		SELECT tag_id, endpoint, public_link, is_censoring, is_ethical, export_failure_count, last_export_try_ts, last_export_success_ts
		FROM relays`)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return relays, nil
}

func (c *ConsensusDB) UpdateRelay(tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
		UPDATE relays SET
			export_failure_count = 0,
			last_export_try_ts = NOW() AT TIME ZONE 'utc',
			last_export_success_ts = NOW() AT TIME ZONE 'utc'
		WHERE tag_id = $1 AND endpoint = $2`, tagID, endpoint)

	return err
}

func (c *ConsensusDB) UpdateRelayLastExportTry(tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
			UPDATE relays SET
				last_export_try_ts = (NOW() AT TIME ZONE 'utc')
			WHERE tag_id = $1 AND endpoint = $2`, tagID, endpoint)

	return err
}

func (c *ConsensusDB) UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
	UPDATE relays SET
		export_failure_count = $1,
		last_export_try_ts = (NOW() AT TIME ZONE 'utc')
	WHERE tag_id = $2 AND endpoint = $3`, exportFailureCount+1, tagID, endpoint)

	return err
}

func (c *ConsensusDB) GetFirstRelayBlock(tagID string) (types.RelayBlock, error) {
	var block types.RelayBlock
	err := c.ReaderDb.Get(&block, `
		SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient
		FROM relays_blocks
		WHERE tag_id=$1
		ORDER BY block_slot ASC
		LIMIT 1`, tagID)

	return block, err
}

func (c *ConsensusDB) GetLastRelayBlock(tagID string) (types.RelayBlock, error) {
	var block types.RelayBlock
	err := c.ReaderDb.Get(&block, `
		SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient
		FROM relays_blocks
		WHERE tag_id=$1
		ORDER BY block_slot DESC
		LIMIT 1`, tagID)

	return block, err
}

func (c *ConsensusDB) SaveBlockTagsAndRelays(tagID string, payload types.BidTrace) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	// first insert the tag into the blocks_tags table
	_, err = tx.Exec(`
	INSERT INTO blocks_tags
	SELECT blocks.slot, blocks.blockroot, $1
	FROM blocks
	WHERE blocks.slot = $2 AND blocks.exec_block_hash = $3
	ON CONFLICT DO NOTHING`, tagID, payload.Slot,
		utils.MustParseHex(payload.BlockHash))

	if err != nil {
		log.Error(fmt.Errorf("failed to insert payload into blocks_tags table"), "", 0, map[string]interface{}{"relay": tagID})
		return err
	}

	// save relays
	_, err = tx.Exec(`
		INSERT INTO relays_blocks
		(
			tag_id,
			block_slot,
			block_root,
			exec_block_hash,
			value,
			builder_pubkey,
			proposer_pubkey,
			proposer_fee_recipient
		)
		SELECT
			$1,	blocks.slot, blocks.blockroot, blocks.exec_block_hash, $4, $5, $6, $7
		FROM blocks
		WHERE
			blocks.slot = $2 and
			blocks.exec_block_hash = $3
		ON CONFLICT (block_slot, block_root, tag_id) DO NOTHING`,
		tagID,
		payload.Slot,
		utils.MustParseHex(payload.BlockHash),
		payload.Value,
		utils.MustParseHex(payload.BuilderPubkey),
		utils.MustParseHex(payload.ProposerPubkey),
		utils.MustParseHex(payload.ProposerFeeRecipient),
	)

	if err != nil {
		log.Error(fmt.Errorf("failed to insert payload into relays_blocks table"), "", 0, map[string]interface{}{"relay": tagID})
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) GetLatestFinalizedEpoch() (uint64, error) {
	var latestFinalized uint64
	err := c.WriterDb.Get(&latestFinalized, "SELECT epoch FROM epochs WHERE finalized ORDER BY epoch DESC LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Error(err, "error retrieving latest exported finalized epoch from the database", 0)
		return 0, err
	}

	return latestFinalized, nil
}

func (c *ConsensusDB) GetSyncCommitteesCountPerValidator() (uint64, error) {
	var rowCount uint64
	err := c.WriterDb.Get(&rowCount, `SELECT COUNT(*) FROM sync_committees_count_per_validator`)
	return rowCount, err
}

func (c *ConsensusDB) GetTotalPeriodSyncCommitteesCountPerValidator() (uint64, error) {
	var dbPeriod uint64
	err := c.WriterDb.Get(&dbPeriod, `SELECT MAX(period) FROM sync_committees_count_per_validator`)
	return dbPeriod, err
}

func (c *ConsensusDB) GetCountSoFarSyncCommitteesCountPerValidator(period uint64) (float64, error) {
	var countSoFar float64
	err := c.WriterDb.Get(&countSoFar, `SELECT count_so_far FROM sync_committees_count_per_validator WHERE period = $1`, period)
	return countSoFar, err
}

func (c *ConsensusDB) SaveSyncCommitteesCount(period uint64, count float64) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees_count_per_validator (period, count_so_far)
			VALUES (%d, %f)
			ON CONFLICT (period) DO UPDATE SET
				period = excluded.period,
				count_so_far = excluded.count_so_far`,
			period, count))

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) GetEpochValidatorsCount(epoch uint64) (uint64, error) {
	var totalCount uint64
	err := c.WriterDb.Get(&totalCount, "SELECT validatorscount FROM epochs WHERE epoch = $1", epoch)
	return totalCount, err
}

func (c *ConsensusDB) GetSyncCommitteesPeriods() ([]uint64, error) {
	var periods []uint64
	err := c.WriterDb.Select(&periods, `SELECT period FROM sync_committees GROUP BY period`)
	return periods, err
}

func (c *ConsensusDB) SaveSyncCommitteeData(data []types.SyncCommittee) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	nArgs := 3
	ids := make([]string, len(data))
	queryArgs := make([]interface{}, len(data)*nArgs)
	for i, entry := range data {
		ids[i] = fmt.Sprintf("($%d,$%d,$%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3)
		queryArgs[i*nArgs] = entry.Period
		queryArgs[i*nArgs+1] = entry.ValidatorIndex
		queryArgs[i*nArgs+2] = entry.CommitteeIndex
	}

	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO sync_committees (period, validatorindex, committeeindex)
			VALUES %s ON CONFLICT (period, validatorindex, committeeindex) DO NOTHING`,
			strings.Join(ids, ",")),
		queryArgs...)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) UpdatePubkeyTags() error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`INSERT INTO validator_tags (publickey, tag)
		SELECT publickey, FORMAT('pool:%s', sps.name) tag
		FROM eth1_deposits
		inner join stake_pools_stats as sps on ENCODE(from_address::bytea, 'hex')=sps.address
		WHERE sps.name NOT LIKE '%Rocketpool -%'
		ON CONFLICT (publickey, tag) DO NOTHING`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) SavePendingDepositsQueue(pendingDeposits []types.PendingDeposit) error {
	dat := make([][]interface{}, len(pendingDeposits))
	for i, r := range pendingDeposits {
		dat[i] = []interface{}{r.ID, r.ValidatorIndex, utils.DBEncodeToHex(r.Pubkey), utils.DBEncodeToHex(r.WithdrawalCredentials), r.Amount, utils.DBEncodeToHex(r.Signature), r.Slot, r.QueuedBalanceAhead, r.EstClearEpoch}
	}

	conn, err := c.WriterDb.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()
	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()
		tx, err := conn.Begin(context.Background())
		if err != nil {
			return err
		}

		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		err = db.ClearAndCopyToTable(tx, "pending_deposits_queue", []string{"id", "validator_index", "pubkey", "withdrawal_credentials", "amount", "signature", "slot", "queued_balance_ahead", "est_clear_epoch"}, dat)
		if err != nil {
			return errors.Wrap(err, "failed to save pending deposits queue")
		}

		err = matchDepositRequests(tx)
		if err != nil {
			return fmt.Errorf("error matching data with blocks_deposit_requests_v2 table: %w", err)
		}

		return tx.Commit(context.Background())
	})

	return err
}

func matchDepositRequests(tx pgx.Tx) error {
	// matching will be wrong for postponed system-deposits
	// but likelihood to ever occur for one pubkey, amount, slot combo is effectively 0
	query := `
	WITH pdq_ranked AS (
		SELECT *, ROW_NUMBER() OVER (
			PARTITION BY pubkey, amount, slot ORDER BY id
		) AS rn
		FROM pending_deposits_queue
	),
	bdr_ranked AS (
		SELECT *, ROW_NUMBER() OVER (
			PARTITION BY pubkey, amount, slot_queued ORDER BY index_queued ASC
		) AS rn
		FROM blocks_deposit_requests_v2
		WHERE status = 'queued' OR status = 'postponed'
	),
	matches AS (
		SELECT pdq.id AS pdq_id, bdr.id AS bdr_id
		FROM pdq_ranked pdq
		JOIN bdr_ranked bdr
			ON pdq.pubkey = bdr.pubkey
			AND pdq.amount = bdr.amount
			AND (
				pdq.slot = bdr.slot_queued AND pdq.rn = bdr.rn OR
				(pdq.slot = 0 AND bdr.index_queued < 0)
			)
	)
	UPDATE pending_deposits_queue
	SET request_id = matches.bdr_id
	FROM matches
	WHERE pending_deposits_queue.id = matches.pdq_id;`

	_, err := tx.Exec(context.Background(), query)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (c *ConsensusDB) GetLastIndexedConsensusLayerEventsEpoch() (count int64, err error) {
	err = c.WriterDb.Get(&count, "SELECT COALESCE(MAX(epoch), -1) FROM consensus_layer_events_indexer_metadata")
	return count, err
}

func (c *ConsensusDB) UpdateConsensusLayerExporterMetadata(tx *sqlx.Tx, data []types.EpochBlockRoots) error {
	rows := make([]struct {
		Epoch     uint64 `db:"epoch"`
		BlockRoot []byte `db:"transition_block_hash"`
	}, len(data))
	for i, d := range data {
		rows[i].Epoch = d.Epoch
		rows[i].BlockRoot = d.EpochBlockRoot
	}
	_, err := tx.NamedExec(
		`INSERT INTO consensus_layer_events_indexer_metadata (epoch, transition_block_hash)
		VALUES (:epoch, :transition_block_hash)`,
		rows,
	)
	if err != nil {
		return errors.Wrap(err, "error inserting into consensus_layer_events_indexer_metadata")
	}
	return nil
}

func (c *ConsensusDB) GetConsensusLayerEventsOfType(slot uint64, eventName types.ConsensusLayerEventName) ([]types.ConsensusLayerEvent, error) {
	var events []types.ConsensusLayerEvent
	err := c.WriterDb.Select(&events, "SELECT event_name, id, block_root, slot, event_index, data FROM consensus_layer_events WHERE slot = $1 AND event_name = $2 and version = $3", slot, eventName, types.ConsensusLayerEventVersion)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (c *ConsensusDB) GetWriterTx() (*sqlx.Tx, error) {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *ConsensusDB) GetConsensusLayerEventsForFilters(filters []types.ConsensusLayerEventFilter) ([]types.ConsensusLayerEvent, error) {
	var events []types.ConsensusLayerEvent
	if len(filters) == 0 {
		return events, nil
	}

	var orConditions []goqu.Expression
	for _, filter := range filters {
		orConditions = append(orConditions, goqu.Ex{
			"slot":       filter.Slot,
			"block_root": filter.BlockRoot,
		})
	}

	ds := goqu.Dialect("postgres").From("consensus_layer_events").
		Select("event_name", "id", "block_root", "slot", "event_index", "data").
		Where(goqu.And(
			goqu.Ex{"version": types.ConsensusLayerEventVersion},
			goqu.Or(orConditions...),
		))

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = c.WriterDb.Select(&events, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return events, nil
}
