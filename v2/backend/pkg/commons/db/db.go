package db

import (
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"

	"github.com/gobitfly/beaconchain/pkg/commons/types"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

var DBPGX *pgxpool.Conn

// DB is a pointer to the explorer-database
var WriterDb *sqlx.DB
var ReaderDb *sqlx.DB

var AlloyReader *sqlx.DB
var AlloyWriter *sqlx.DB

var PersistentRedisDbClient *redis.Client

var FarFutureEpoch = uint64(18446744073709551615)
var MaxSqlNumber = uint64(9223372036854775807)

const WithdrawalsQueryLimit = 10000
const BlsChangeQueryLimit = 10000
const MaxSqlInteger = 2147483647

const DefaultInfScrollRows = 25

var ErrNoStats = errors.New("no stats available")

func dbTestConnection(dbConn *sqlx.DB, dataBaseName string) {
	// The golang sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout := time.NewTimer(15 * time.Second)

	go func() {
		<-dbConnectionTimeout.C
		log.Fatal(fmt.Errorf("timeout while connecting to %s", dataBaseName), "", 0)
	}()

	err := dbConn.Ping()
	if err != nil {
		log.Fatal(fmt.Errorf("unable to ping %s", dataBaseName), "", 0)
	}

	dbConnectionTimeout.Stop()
}

func MustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) (*sqlx.DB, *sqlx.DB) {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}

	if reader.MaxOpenConns == 0 {
		reader.MaxOpenConns = 50
	}
	if reader.MaxIdleConns == 0 {
		reader.MaxIdleConns = 10
	}
	if reader.MaxOpenConns < reader.MaxIdleConns {
		reader.MaxIdleConns = reader.MaxOpenConns
	}

	sslMode := "disable"
	if writer.SSL {
		sslMode = "require"
	}

	log.Infof("initializing writer db connection to %v with %v/%v conn limit", writer.Host, writer.MaxIdleConns, writer.MaxOpenConns)
	dbConnWriter, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", writer.Username, writer.Password, net.JoinHostPort(writer.Host, writer.Port), writer.Name, sslMode))
	if err != nil {
		log.Fatal(err, "error getting Connection Writer database", 0)
	}

	dbTestConnection(dbConnWriter, "database")
	dbConnWriter.SetConnMaxIdleTime(time.Second * 30)
	dbConnWriter.SetConnMaxLifetime(time.Minute)
	dbConnWriter.SetMaxOpenConns(writer.MaxOpenConns)
	dbConnWriter.SetMaxIdleConns(writer.MaxIdleConns)

	if reader == nil {
		return dbConnWriter, dbConnWriter
	}

	sslMode = "disable"
	if reader.SSL {
		sslMode = "require"
	}

	log.Infof("initializing reader db connection to %v with %v/%v conn limit", writer.Host, reader.MaxIdleConns, reader.MaxOpenConns)
	dbConnReader, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", reader.Username, reader.Password, net.JoinHostPort(reader.Host, reader.Port), reader.Name, sslMode))
	if err != nil {
		log.Fatal(err, "error getting Connection Reader database", 0)
	}

	dbTestConnection(dbConnReader, "read replica database")
	dbConnReader.SetConnMaxIdleTime(time.Second * 30)
	dbConnReader.SetConnMaxLifetime(time.Minute)
	dbConnReader.SetMaxOpenConns(reader.MaxOpenConns)
	dbConnReader.SetMaxIdleConns(reader.MaxIdleConns)
	return dbConnWriter, dbConnReader
}

func ApplyEmbeddedDbSchema(version int64) error {
	goose.SetBaseFS(EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if version == -2 {
		if err := goose.Up(WriterDb.DB, "migrations"); err != nil {
			return err
		}
	} else if version == -1 {
		if err := goose.UpByOne(WriterDb.DB, "migrations"); err != nil {
			return err
		}
	} else {
		if err := goose.UpTo(WriterDb.DB, "migrations", version); err != nil {
			return err
		}
	}

	return nil
}

func GetEth1DepositsJoinEth2Deposits(query string, length, start uint64, orderBy, orderDir string, latestEpoch, validatorOnlineThresholdSlot uint64) ([]*types.EthOneDepositsData, uint64, error) {
	// Initialize the return values
	deposits := []*types.EthOneDepositsData{}
	totalCount := uint64(0)

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"tx_hash", "tx_input", "tx_index", "block_number", "block_ts", "from_address", "publickey", "withdrawal_credentials", "amount", "signature", "merkletree_index", "state", "valid_signature"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_ts"
	}

	var param interface{}
	var searchQuery string
	var err error

	// Define the base queries
	deposistsCountQuery := `
		SELECT COUNT(*) FROM eth1_deposits as eth1
		%s`

	deposistsQuery := `
		SELECT
			eth1.tx_hash as tx_hash,
			eth1.tx_input as tx_input,
			eth1.tx_index as tx_index,
			eth1.block_number as block_number,
			eth1.block_ts as block_ts,
			eth1.from_address as from_address,
			eth1.publickey as publickey,
			eth1.withdrawal_credentials as withdrawal_credentials,
			eth1.amount as amount,
			eth1.signature as signature,
			eth1.merkletree_index as merkletree_index,
			eth1.valid_signature as valid_signature,
			COALESCE(v.state, 'deposited') as state
		FROM
			eth1_deposits as eth1
		LEFT JOIN
			(
				SELECT pubkey, status AS state
				FROM validators
			) as v
		ON
			v.pubkey = eth1.publickey
		%s
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`

	// Get the search query and parameter for it
	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var hash []byte
	if len(trimmedQuery)%2 == 0 && utils.HashLikeRegex.MatchString(trimmedQuery) {
		hash, err = hex.DecodeString(trimmedQuery)
		if err != nil {
			return nil, 0, err
		}
	}
	if trimmedQuery == "" {
		err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, ""))
		if err != nil {
			return nil, 0, err
		}

		err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, "", orderBy, orderDir), length, start)
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}

		return deposits, totalCount, nil
	}

	param = hash
	if utils.IsHash(trimmedQuery) {
		searchQuery = `WHERE eth1.publickey = $3`
	} else if utils.IsEth1Tx(trimmedQuery) {
		// Withdrawal credentials have the same length as a tx hash
		if utils.IsValidWithdrawalCredentials(trimmedQuery) {
			searchQuery = `
				WHERE
					eth1.tx_hash = $3
					OR eth1.withdrawal_credentials = $3`
		} else {
			searchQuery = `WHERE eth1.tx_hash = $3`
		}
	} else if utils.IsEth1Address(trimmedQuery) {
		searchQuery = `WHERE eth1.from_address = $3`
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 31); parseErr == nil { // Limit to 31 bits to stay within math.MaxInt32
		param = uiQuery
		searchQuery = `WHERE eth1.block_number = $3`
	} else {
		// The query does not fulfill any of the requirements for a search
		return deposits, totalCount, nil
	}

	// The deposits count query only has one parameter for the search
	countSearchQuery := strings.ReplaceAll(searchQuery, "$3", "$1")

	err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, countSearchQuery), param)
	if err != nil {
		return nil, 0, err
	}

	err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, searchQuery, orderBy, orderDir), length, start, param)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return deposits, totalCount, nil
}

func GetEth1DepositsLeaderboard(query string, length, start uint64, orderBy, orderDir string) ([]*types.EthOneDepositLeaderboardData, uint64, error) {
	deposits := []*types.EthOneDepositLeaderboardData{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{
		"from_address",
		"amount",
		"validcount",
		"invalidcount",
		"slashedcount",
		"totalcount",
		"activecount",
		"pendingcount",
		"voluntary_exit_count",
	}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "amount"
	}

	var err error
	var totalCount uint64
	if query != "" {
		err = ReaderDb.Get(&totalCount, `
		SELECT COUNT(*) FROM eth1_deposits_aggregated WHERE ENCODE(from_address, 'hex') LIKE LOWER($1)`, query+"%")
	} else {
		err = ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM eth1_deposits_aggregated AS count")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	if query != "" {
		err = ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count
			FROM eth1_deposits_aggregated
			WHERE ENCODE(from_address, 'hex') LIKE LOWER($3)
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, query+"%")
	} else {
		err = ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count
			FROM eth1_deposits_aggregated
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return deposits, totalCount, nil
}

func GetEth2Deposits(query string, length, start uint64, orderBy, orderDir string) ([]*types.EthTwoDepositData, uint64, error) {
	// Initialize the return values
	deposits := []*types.EthTwoDepositData{}
	totalCount := uint64(0)

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "publickey", "amount", "withdrawalcredentials", "signature"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	var param interface{}
	var searchQuery string
	var err error

	// Define the base queries
	deposistsCountQuery := `
		SELECT COUNT(*)
		FROM blocks_deposits
		INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
		%s`

	deposistsQuery := `
			SELECT
				blocks_deposits.block_slot,
				blocks_deposits.block_index,
				blocks_deposits.proof,
				blocks_deposits.publickey,
				blocks_deposits.withdrawalcredentials,
				blocks_deposits.amount,
				blocks_deposits.signature
			FROM blocks_deposits
			INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
			%s
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`

	// Get the search query and parameter for it
	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var hash []byte
	if len(trimmedQuery)%2 == 0 && utils.HashLikeRegex.MatchString(trimmedQuery) {
		hash, err = hex.DecodeString(trimmedQuery)
		if err != nil {
			return nil, 0, err
		}
	}
	if trimmedQuery == "" {
		err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, ""))
		if err != nil {
			return nil, 0, err
		}

		err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, "", orderBy, orderDir), length, start)
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}

		return deposits, totalCount, nil
	}

	if utils.IsHash(trimmedQuery) {
		param = hash
		searchQuery = `WHERE blocks_deposits.publickey = $3`
	} else if utils.IsValidWithdrawalCredentials(trimmedQuery) {
		param = hash
		searchQuery = `WHERE blocks_deposits.withdrawalcredentials = $3`
	} else if utils.IsEth1Address(trimmedQuery) {
		param = hash
		searchQuery = `
				LEFT JOIN eth1_deposits ON blocks_deposits.publickey = eth1_deposits.publickey
				WHERE eth1_deposits.from_address = $3`
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 31); parseErr == nil { // Limit to 31 bits to stay within math.MaxInt32
		param = uiQuery
		searchQuery = `WHERE blocks_deposits.block_slot = $3`
	} else {
		// The query does not fulfill any of the requirements for a search
		return deposits, totalCount, nil
	}

	// The deposits count query only has one parameter for the search
	countSearchQuery := strings.ReplaceAll(searchQuery, "$3", "$1")

	err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, countSearchQuery), param)
	if err != nil {
		return nil, 0, err
	}

	err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, searchQuery, orderBy, orderDir), length, start, param)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return deposits, totalCount, nil
}

func GetSlashingCount() (uint64, error) {
	slashings := uint64(0)

	err := ReaderDb.Get(&slashings, `
		SELECT SUM(count)
		FROM
		(
			SELECT COUNT(*)
			FROM
				blocks_attesterslashings
				INNER JOIN blocks on blocks.slot = blocks_attesterslashings.block_slot and blocks.status = '1'
			UNION
			SELECT COUNT(*)
			FROM
				blocks_proposerslashings
				INNER JOIN blocks on blocks.slot = blocks_proposerslashings.block_slot and blocks.status = '1'
		) as tbl`)
	if err != nil {
		return 0, err
	}

	return slashings, nil
}

// GetLatestEpoch will return the latest epoch from the database
func GetLatestEpoch() (uint64, error) {
	var epoch uint64
	err := WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %w", err)
	}

	return epoch, nil
}

func GetAllSlots(tx *sqlx.Tx) ([]uint64, error) {
	var slots []uint64
	err := tx.Select(&slots, "SELECT slot FROM blocks ORDER BY slot")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all slots from the DB: %w", err)
	}

	return slots, nil
}

func SetSlotFinalizationAndStatus(slot uint64, finalized bool, status string, tx *sqlx.Tx) error {
	_, err := tx.Exec("UPDATE blocks SET finalized = $1, status = $2 WHERE slot = $3", finalized, status, slot)

	if err != nil {
		return fmt.Errorf("error setting slot finalization and status: %w", err)
	}

	return nil
}

type GetAllNonFinalizedSlotsRow struct {
	Slot      uint64 `db:"slot"`
	BlockRoot []byte `db:"blockroot"`
	Finalized bool   `db:"finalized"`
	Status    string `db:"status"`
}

func GetAllNonFinalizedSlots() ([]*GetAllNonFinalizedSlotsRow, error) {
	var slots []*GetAllNonFinalizedSlotsRow
	err := WriterDb.Select(&slots, "SELECT slot, blockroot, finalized, status FROM blocks WHERE NOT finalized ORDER BY slot")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all non finalized slots from the DB: %w", err)
	}

	return slots, nil
}

// Get latest finalized epoch
func GetLatestFinalizedEpoch() (uint64, error) {
	var latestFinalized uint64
	err := WriterDb.Get(&latestFinalized, "SELECT epoch FROM epochs WHERE finalized ORDER BY epoch DESC LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Error(err, "error retrieving latest exported finalized epoch from the database", 0)
		return 0, err
	}

	return latestFinalized, nil
}

// GetValidatorPublicKeys will return the public key for a list of validator indices and or public keys
func GetValidatorPublicKeys(indices []uint64, keys [][]byte) ([][]byte, error) {
	var publicKeys [][]byte
	err := ReaderDb.Select(&publicKeys, "SELECT pubkey FROM validators WHERE validatorindex = ANY($1) OR pubkey = ANY($2)", indices, keys)

	return publicKeys, err
}

// GetValidatorIndex will return the validator-index for a public key from the database
func GetValidatorIndex(publicKey []byte) (uint64, error) {
	var index uint64
	err := ReaderDb.Get(&index, "SELECT validatorindex FROM validators WHERE pubkey = $1", publicKey)

	return index, err
}

// UpdateCanonicalBlocks will update the blocks for an epoch range in the database
func UpdateCanonicalBlocks(startEpoch, endEpoch uint64, blocks []*types.MinimalBlock) error {
	if len(blocks) == 0 {
		return nil
	}
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_canonical_blocks").Observe(time.Since(start).Seconds())
	}()

	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)

	lastSlotNumber := uint64(0)
	for _, block := range blocks {
		if block.Slot > lastSlotNumber {
			lastSlotNumber = block.Slot
		}
	}

	_, err = tx.Exec("UPDATE blocks SET status = 3 WHERE epoch >= $1 AND epoch <= $2 AND (status = '1' OR status = '3') AND slot <= $3", startEpoch, endEpoch, lastSlotNumber)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if block.Canonical {
			log.Infof("marking block %x at slot %v as canonical", block.BlockRoot, block.Slot)
			_, err = tx.Exec("UPDATE blocks SET status = '1' WHERE blockroot = $1", block.BlockRoot)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func SetBlockStatus(blocks []*types.CanonBlock) error {
	if len(blocks) == 0 {
		return nil
	}

	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)

	canonBlocks := make(pq.ByteaArray, 0)
	orphanedBlocks := make(pq.ByteaArray, 0)
	for _, block := range blocks {
		if !block.Canonical {
			log.Infof("marking block %x at slot %v as orphaned", block.BlockRoot, block.Slot)
			orphanedBlocks = append(orphanedBlocks, block.BlockRoot)
		} else {
			log.Infof("marking block %x at slot %v as canonical", block.BlockRoot, block.Slot)
			canonBlocks = append(canonBlocks, block.BlockRoot)
		}
	}

	_, err = tx.Exec("UPDATE blocks SET status = '1' WHERE blockroot = ANY($1)", canonBlocks)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE blocks SET status = '3' WHERE blockroot = ANY($1)", orphanedBlocks)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveValidatorQueue will save the validator queue into the database
func SaveValidatorQueue(validators *types.ValidatorQueue, tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO queue (ts, entering_validators_count, exiting_validators_count)
		VALUES (date_trunc('hour', now()), $1, $2)
		ON CONFLICT (ts) DO UPDATE SET
			entering_validators_count = excluded.entering_validators_count,
			exiting_validators_count = excluded.exiting_validators_count`,
		validators.Activating, validators.Exiting)
	return err
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_epochs_status").Observe(time.Since(start).Seconds())
	}()

	_, err := tx.Exec(`
		UPDATE epochs SET
			eligibleether = $1,
			globalparticipationrate = $2,
			votedether = $3,
			finalized = $4,
			blockscount = (SELECT COUNT(*) FROM blocks WHERE epoch = $5 AND status = '1'),
			proposerslashingscount = (SELECT COALESCE(SUM(proposerslashingscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			attesterslashingscount = (SELECT COALESCE(SUM(attesterslashingscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			attestationscount = (SELECT COALESCE(SUM(attestationscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			depositscount = (SELECT COALESCE(SUM(depositscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			withdrawalcount = (SELECT COALESCE(SUM(withdrawalcount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			voluntaryexitscount = (SELECT COALESCE(SUM(voluntaryexitscount),0) FROM blocks WHERE epoch = $5 AND status = '1')
		WHERE epoch = $5`,
		stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Finalized, stats.Epoch)

	return err
}

func GetRelayDataForIndexedBlocks(blocks []*types.Eth1BlockIndexed) (map[common.Hash]types.RelaysData, error) {
	var execBlockHashes [][]byte
	var relaysData []types.RelaysData

	for _, block := range blocks {
		execBlockHashes = append(execBlockHashes, block.Hash)
	}
	// try to get mev rewards from relys_blocks table
	err := ReaderDb.Select(&relaysData,
		`SELECT proposer_fee_recipient, value, exec_block_hash, tag_id, builder_pubkey FROM relays_blocks WHERE relays_blocks.exec_block_hash = ANY($1)`,
		pq.ByteaArray(execBlockHashes),
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var relaysDataMap = make(map[common.Hash]types.RelaysData)
	for _, relayData := range relaysData {
		relaysDataMap[common.BytesToHash(relayData.ExecBlockHash)] = relayData
	}

	return relaysDataMap, nil
}

// GetValidatorIndices will return the total-validator-indices
func GetValidatorIndices() ([]uint64, error) {
	indices := []uint64{}
	err := ReaderDb.Select(&indices, "select validatorindex from validators order by validatorindex;")
	return indices, err
}

// GetTotalValidatorsCount will return the total-validator-count
func GetTotalValidatorsCount() (uint64, error) {
	var totalCount uint64
	err := ReaderDb.Get(&totalCount, "select coalesce(max(validatorindex) + 1, 0) from validators;")
	return totalCount, err
}

// GetActiveValidatorCount will return the total-validator-count
func GetActiveValidatorCount() (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, "select count(*) from validators where status in ('active_offline', 'active_online');")
	return count, err
}

func UpdateQueueDeposits(tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		log.Infof("took %v seconds to update queue deposits", time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("update_queue_deposits").Observe(time.Since(start).Seconds())
	}()

	// first we remove any validator that isn't queued anymore
	_, err := tx.Exec(`
		DELETE FROM validator_queue_deposits
		WHERE validator_queue_deposits.validatorindex NOT IN (
			SELECT validatorindex
			FROM validators
			WHERE activationepoch=9223372036854775807 and status='pending')`)
	if err != nil {
		log.Error(err, "error removing queued publickeys from validator_queue_deposits", 0)
		return err
	}

	// then we add any new ones that are queued
	_, err = tx.Exec(`
		INSERT INTO validator_queue_deposits
		SELECT validatorindex FROM validators WHERE activationepoch=$1 and status='pending' ON CONFLICT DO NOTHING
	`, MaxSqlNumber)
	if err != nil {
		log.Error(err, "error adding queued publickeys to validator_queue_deposits", 0)
		return err
	}

	// now we add the activationeligibilityepoch where it is missing
	_, err = tx.Exec(`
		UPDATE validator_queue_deposits
		SET
			activationeligibilityepoch=validators.activationeligibilityepoch
		FROM validators
		WHERE
			validator_queue_deposits.activationeligibilityepoch IS NULL AND
			validator_queue_deposits.validatorindex = validators.validatorindex
	`)
	if err != nil {
		log.Error(err, "error updating activationeligibilityepoch on validator_queue_deposits", 0)
		return err
	}

	// efficiently collect the tnx that pushed each validator over 32 ETH.
	_, err = tx.Exec(`
		UPDATE validator_queue_deposits
		SET
			block_slot=data.block_slot,
			block_index=data.block_index
		FROM (
			WITH CumSum AS
			(
				SELECT publickey, block_slot, block_index,
					/* generate partion per publickey ordered by newest to oldest. store cum sum of deposits */
					SUM(amount) OVER (partition BY publickey ORDER BY (block_slot, block_index) ASC) AS cumTotal
				FROM blocks_deposits
				WHERE publickey IN (
					/* get the pubkeys of the indexes */
					select pubkey from validators where validators.validatorindex in (
						/* get the indexes we need to update */
						select validatorindex from validator_queue_deposits where block_slot is null or block_index is null
					)
				)
				ORDER BY block_slot, block_index ASC
			)
			/* we only care about one deposit per vali */
			SELECT DISTINCT ON(publickey) validators.validatorindex, block_slot, block_index
			FROM CumSum
			/* join so we can retrieve the validator index again */
			left join validators on validators.pubkey = CumSum.publickey
			/* we want the deposit that pushed the cum sum over 32 ETH */
			WHERE cumTotal>=32000000000
			ORDER BY publickey, cumTotal asc
		) AS data
		WHERE validator_queue_deposits.validatorindex=data.validatorindex`)
	if err != nil {
		log.Error(err, "error updating validator_queue_deposits: %v", 0)
		return err
	}
	return nil
}

func GetQueueAheadOfValidator(validatorIndex uint64) (uint64, error) {
	var res uint64
	var selected struct {
		BlockSlot                  uint64 `db:"block_slot"`
		BlockIndex                 uint64 `db:"block_index"`
		ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	}
	err := ReaderDb.Get(&selected, `
		SELECT
			COALESCE(block_index, 0) as block_index,
			COALESCE(block_slot, 0) as block_slot,
			COALESCE(activationeligibilityepoch, $2) as activationeligibilityepoch
		FROM validator_queue_deposits
		WHERE
			validatorindex = $1
		`, validatorIndex, MaxSqlNumber)
	if err == sql.ErrNoRows {
		// If we did not find our validator in the queue it is most likly that he has not yet been added so we put him as last
		err = ReaderDb.Get(&res, `
			SELECT count(*)
			FROM validator_queue_deposits
		`)
		if err == nil {
			return res, nil
		}
	}
	if err != nil {
		return res, err
	}
	err = ReaderDb.Get(&res, `
	SELECT count(*)
	FROM validator_queue_deposits
	WHERE
		COALESCE(activationeligibilityepoch, 0) < $1 OR
		block_slot < $2 OR
		block_slot = $2 AND block_index < $3`, selected.ActivationEligibilityEpoch, selected.BlockSlot, selected.BlockIndex)
	return res, err
}

func GetValidatorNames() (map[uint64]string, error) {
	rows, err := ReaderDb.Query(`
		SELECT validatorindex, validator_names.name
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE validator_names.name IS NOT NULL`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	validatorIndexToNameMap := make(map[uint64]string, 30000)

	for rows.Next() {
		var index uint64
		var name string

		err := rows.Scan(&index, &name)

		if err != nil {
			return nil, err
		}
		validatorIndexToNameMap[index] = name
	}

	return validatorIndexToNameMap, nil
}

// GetPendingValidatorCount queries the pending validators currently in the queue
func GetPendingValidatorCount() (uint64, error) {
	count := uint64(0)
	err := ReaderDb.Get(&count, "SELECT entering_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error retrieving validator queue count: %w", err)
	}
	return count, nil
}

func GetTotalEligibleEther() (uint64, error) {
	var total uint64

	err := ReaderDb.Get(&total, `
		SELECT eligibleether FROM epochs ORDER BY epoch DESC LIMIT 1
	`)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return total / 1e9, nil
}

// GetValidatorsGotSlashed returns the validators that got slashed after `epoch` either by an attestation violation or a proposer violation
func GetValidatorsGotSlashed(epoch uint64) ([]struct {
	Epoch                  uint64 `db:"epoch"`
	SlasherIndex           uint64 `db:"slasher"`
	SlasherPubkey          string `db:"slasher_pubkey"`
	SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
	SlashedValidatorPubkey []byte `db:"slashedvalidator_pubkey"`
	Reason                 string `db:"reason"`
}, error) {
	var dbResult []struct {
		Epoch                  uint64 `db:"epoch"`
		SlasherIndex           uint64 `db:"slasher"`
		SlasherPubkey          string `db:"slasher_pubkey"`
		SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
		SlashedValidatorPubkey []byte `db:"slashedvalidator_pubkey"`
		Reason                 string `db:"reason"`
	}
	err := ReaderDb.Select(&dbResult, `
		WITH
			slashings AS (
				SELECT DISTINCT ON (slashedvalidator)
					slot,
					epoch,
					slasher,
					slashedvalidator,
					reason
				FROM (
					SELECT
						blocks.slot,
						blocks.epoch,
						blocks.proposer AS slasher,
						UNNEST(ARRAY(
							SELECT UNNEST(attestation1_indices)
								INTERSECT
							SELECT UNNEST(attestation2_indices)
						)) AS slashedvalidator,
						'Attestation Violation' AS reason
					FROM blocks_attesterslashings
					LEFT JOIN blocks ON blocks_attesterslashings.block_slot = blocks.slot
					WHERE blocks.status = '1' AND blocks.epoch > $1
					UNION ALL
						SELECT
							blocks.slot,
							blocks.epoch,
							blocks.proposer AS slasher,
							blocks_proposerslashings.proposerindex AS slashedvalidator,
							'Proposer Violation' AS reason
						FROM blocks_proposerslashings
						LEFT JOIN blocks ON blocks_proposerslashings.block_slot = blocks.slot
						WHERE blocks.status = '1' AND blocks.epoch > $1
				) a
				ORDER BY slashedvalidator, slot
			)
		SELECT slasher, vk.pubkey as slasher_pubkey, slashedvalidator, vv.pubkey as slashedvalidator_pubkey, epoch, reason
		FROM slashings s
	    INNER JOIN validators vk ON s.slasher = vk.validatorindex
		INNER JOIN validators vv ON s.slashedvalidator = vv.validatorindex`, epoch)
	if err != nil {
		return nil, err
	}
	return dbResult, nil
}

func GetSlotVizData(latestEpoch uint64) ([]*types.SlotVizEpochs, error) {
	type sqlBlocks struct {
		Slot                    uint64
		BlockRoot               []byte
		Epoch                   uint64
		Status                  string
		Globalparticipationrate float64
		Finalized               bool
		Justified               bool
		Previousjustified       bool
	}

	var blks []sqlBlocks = []sqlBlocks{}
	if latestEpoch > 4 {
		latestEpoch = latestEpoch - 4
	} else {
		latestEpoch = 0
	}

	latestFinalizedEpoch, err := GetLatestFinalizedEpoch()
	if err != nil {
		return nil, err
	}
	err = ReaderDb.Select(&blks, `
	SELECT
		b.slot,
		b.blockroot,
		CASE
			WHEN b.status = '0' THEN 'scheduled'
			WHEN b.status = '1' THEN 'proposed'
			WHEN b.status = '2' THEN 'missed'
			WHEN b.status = '3' THEN 'orphaned'
			ELSE 'unknown'
		END AS status,
		b.epoch,
		COALESCE(e.globalparticipationrate, 0) AS globalparticipationrate,
		(b.epoch <= $2) AS finalized
	FROM blocks b
		LEFT JOIN epochs e ON e.epoch = b.epoch
	WHERE b.epoch >= $1
	ORDER BY slot DESC;
`, latestEpoch, latestFinalizedEpoch)
	if err != nil {
		return nil, err
	}

	currentSlot := utils.TimeToSlot(uint64(time.Now().Unix()))

	epochMap := map[uint64]*types.SlotVizEpochs{}

	res := []*types.SlotVizEpochs{}

	for _, b := range blks {
		_, exists := epochMap[b.Epoch]
		if !exists {
			r := types.SlotVizEpochs{
				Epoch:          b.Epoch,
				Finalized:      b.Finalized,
				Particicpation: b.Globalparticipationrate,
				Slots:          []*types.SlotVizSlots{},
			}
			r.Slots = make([]*types.SlotVizSlots, utils.Config.Chain.ClConfig.SlotsPerEpoch)
			epochMap[b.Epoch] = &r
		}

		slotIndex := b.Slot - (b.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch)

		// if epochMap[b.Epoch].Slots[slotIndex] != nil && len(b.BlockRoot) > len(epochMap[b.Epoch].Slots[slotIndex].BlockRoot) {
		// 	log.LogInfo("CONFLICTING block found for slotindex %v", slotIndex)
		// }

		if epochMap[b.Epoch].Slots[slotIndex] == nil || len(b.BlockRoot) > len(epochMap[b.Epoch].Slots[slotIndex].BlockRoot) {
			epochMap[b.Epoch].Slots[slotIndex] = &types.SlotVizSlots{
				Epoch:     b.Epoch,
				Slot:      b.Slot,
				Status:    b.Status,
				Active:    b.Slot == currentSlot,
				BlockRoot: b.BlockRoot,
			}
		}
	}

	for _, epoch := range epochMap {
		for i := uint64(0); i < utils.Config.Chain.ClConfig.SlotsPerEpoch; i++ {
			if epoch.Slots[i] == nil {
				status := "scheduled"
				slot := (epoch.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch) + i
				if slot < currentSlot {
					status = "scheduled-missed"
				}
				// log.LogInfo("FILLING MISSING SLOT: %v", slot)
				epoch.Slots[i] = &types.SlotVizSlots{
					Epoch:  epoch.Epoch,
					Slot:   slot,
					Status: status,
					Active: slot == currentSlot,
				}
			}
		}
	}

	for _, epoch := range epochMap {
		for _, slot := range epoch.Slots {
			slot.Active = slot.Slot == currentSlot
			if slot.Status != "proposed" && slot.Status != "missed" {
				if slot.Status == "scheduled" && slot.Slot < currentSlot {
					slot.Status = "scheduled-missed"
				}

				if slot.Slot >= currentSlot {
					slot.Status = "scheduled"
				}
			}
		}
		if epoch.Finalized {
			epoch.Justified = true
		}
		res = append(res, epoch)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Epoch > res[j].Epoch
	})

	for i := 0; i < len(res); i++ {
		if !res[i].Finalized && i != 0 {
			res[i-1].Justifying = true
		}
		if res[i].Finalized && i != 0 {
			res[i-1].Justified = true
			break
		}
	}

	return res, nil
}

func GetBlockNumber(slot uint64) (block uint64, err error) {
	err = ReaderDb.Get(&block, `SELECT exec_block_number FROM blocks where slot >= $1 AND exec_block_number > 0 ORDER BY slot LIMIT 1`, slot)
	return
}

func SaveChartSeriesPoint(date time.Time, indicator string, value any) error {
	_, err := WriterDb.Exec(`INSERT INTO chart_series (time, indicator, value) VALUES($1, $2, $3) ON CONFLICT (time, indicator) DO UPDATE SET value = EXCLUDED.value`, date, indicator, value)
	if err != nil {
		return fmt.Errorf("error saving chart_series: %v: %w", indicator, err)
	}
	return err
}

func GetSlotWithdrawals(slot uint64) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals

	err := ReaderDb.Select(&withdrawals, `
		SELECT
			w.withdrawalindex as index,
			w.validatorindex,
			w.address,
			w.amount
		FROM
			blocks_withdrawals w
		LEFT JOIN blocks b ON b.blockroot = w.block_root
		WHERE w.block_slot = $1 AND b.status = '1'
		ORDER BY w.withdrawalindex
	`, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for slot: %d: %w", slot, err)
	}

	return withdrawals, nil
}

func GetTotalWithdrawals() (total uint64, err error) {
	err = ReaderDb.Get(&total, `
	SELECT
		COALESCE(MAX(withdrawalindex), 0)
	FROM
		blocks_withdrawals`)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return
}

func GetWithdrawalsCountForQuery(query string) (uint64, error) {
	t0 := time.Now()
	defer func() {
		log.InfoWithFields(log.Fields{"duration": time.Since(t0)}, "finished GetWithdrawalsCountForQuery")
	}()
	count := uint64(0)

	withdrawalsQuery := `
		SELECT COUNT(*) FROM (
			SELECT b.slot
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
			%s
			LIMIT %d
		) a`

	var err error = nil

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	if utils.IsEth1Address(query) {
		searchQuery := `WHERE w.address = $1`
		addr, decErr := hex.DecodeString(trimmedQuery)
		if decErr != nil {
			return 0, decErr
		}
		err = ReaderDb.Get(&count, fmt.Sprintf(withdrawalsQuery, searchQuery, WithdrawalsQueryLimit),
			addr)
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
		// Check whether the query can be used for a validator, slot or epoch search
		searchQuery := `
			WHERE w.validatorindex = $1
				OR w.block_slot = $1
				OR w.block_slot BETWEEN $1*$2 AND ($1+1)*$2-1`
		err = ReaderDb.Get(&count, fmt.Sprintf(withdrawalsQuery, searchQuery, WithdrawalsQueryLimit),
			uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetWithdrawals(query string, length, start uint64, orderBy, orderDir string) ([]*types.Withdrawals, error) {
	t0 := time.Now()
	defer func() {
		log.InfoWithFields(log.Fields{"duration": time.Since(t0)}, "finished GetWithdrawals")
	}()
	withdrawals := []*types.Withdrawals{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "withdrawalindex", "validatorindex", "address", "amount"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	withdrawalsQuery := `
		SELECT
			w.block_slot as slot,
			w.withdrawalindex as index,
			w.validatorindex,
			w.address,
			w.amount
		FROM blocks_withdrawals w
		INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
		%s
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`

	var err error = nil

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	if trimmedQuery != "" {
		if utils.IsEth1Address(query) {
			searchQuery := `WHERE w.address = $3`
			addr, decErr := hex.DecodeString(trimmedQuery)
			if decErr != nil {
				return nil, decErr
			}
			err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, searchQuery, orderBy, orderDir),
				length, start, addr)
		} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
			// Check whether the query can be used for a validator, slot or epoch search
			searchQuery := `
				WHERE w.validatorindex = $3
					OR w.block_slot = $3
					OR w.block_slot BETWEEN $3*$4 AND ($3+1)*$4-1`
			err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, searchQuery, orderBy, orderDir),
				length, start, uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
	} else {
		err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, "", orderBy, orderDir), length, start)
	}

	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func GetTotalAmountWithdrawn() (sum uint64, count uint64, err error) {
	var res = struct {
		Sum   uint64 `db:"sum"`
		Count uint64 `db:"count"`
	}{}
	lastExportedDay, err := GetLastExportedStatisticDay()
	if err != nil {
		return 0, 0, fmt.Errorf("error getting latest exported statistic day for withdrawals count: %w", err)
	}
	_, lastEpochOfDay := utils.GetFirstAndLastEpochForDay(lastExportedDay)
	cutoffSlot := (lastEpochOfDay * utils.Config.Chain.ClConfig.SlotsPerEpoch) + 1

	err = ReaderDb.Get(&res, `
		WITH today AS (
			SELECT
				COALESCE(SUM(w.amount), 0) as sum,
				COUNT(*) as count
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
			WHERE w.block_slot >= $1
		),
		stats AS (
			SELECT
				COALESCE(SUM(withdrawals_amount_total), 0) as sum,
				COALESCE(SUM(withdrawals_total), 0) as count
			FROM validator_stats
			WHERE day = $2
		)
		SELECT
			today.sum + stats.sum as sum,
			today.count + stats.count as count
		FROM today, stats`, cutoffSlot, lastExportedDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("error fetching total withdrawal count and amount: %w", err)
	}

	return res.Sum, res.Count, err
}

func GetTotalAmountDeposited() (uint64, error) {
	var total uint64
	err := ReaderDb.Get(&total, `
	SELECT
		COALESCE(sum(d.amount), 0) as sum
	FROM blocks_deposits d
	INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'`)
	return total, err
}

func GetBLSChangeCount() (uint64, error) {
	var total uint64
	err := ReaderDb.Get(&total, `
	SELECT
		COALESCE(count(*), 0) as count
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'`)
	return total, err
}

func GetEpochWithdrawalsTotal(epoch uint64) (total uint64, err error) {
	err = ReaderDb.Get(&total, `
	SELECT
		COALESCE(sum(w.amount), 0) as sum
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.block_slot >= $1 AND w.block_slot < $2`, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	return
}

func GetEpochWithdrawals(epoch uint64) ([]*types.WithdrawalsNotification, error) {
	var withdrawals []*types.WithdrawalsNotification

	err := ReaderDb.Select(&withdrawals, `
	SELECT
		w.block_slot as slot,
		w.withdrawalindex as index,
		w.validatorindex,
		w.address,
		w.amount,
		v.pubkey as pubkey
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	LEFT JOIN validators v on v.validatorindex = w.validatorindex
	WHERE w.block_slot >= $1 AND w.block_slot < $2 ORDER BY w.withdrawalindex`, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for epoch: %d: %w", epoch, err)
	}

	return withdrawals, nil
}

func GetValidatorWithdrawals(validator uint64, limit uint64, offset uint64, orderBy string, orderDir string) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}

	err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
	SELECT
		w.block_slot as slot,
		w.withdrawalindex as index,
		w.validatorindex,
		w.address,
		w.amount
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE validatorindex = $1
	ORDER BY  w.%s %s
	LIMIT $2 OFFSET $3`, orderBy, orderDir), validator, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validator: %d: %w", validator, err)
	}

	return withdrawals, nil
}

func GetValidatorsWithdrawals(validators []uint64, fromEpoch uint64, toEpoch uint64) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals

	err := ReaderDb.Select(&withdrawals, `
	SELECT
		w.block_slot as slot,
		w.withdrawalindex as index,
		w.block_root as blockroot,
		w.validatorindex,
		w.address,
		w.amount
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE validatorindex = ANY($1)
	AND (w.block_slot / $4) >= $2 AND (w.block_slot / $4) <= $3
	ORDER BY w.withdrawalindex`, pq.Array(validators), fromEpoch, toEpoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validators: %+v: %w", validators, err)
	}

	return withdrawals, nil
}

func GetValidatorsWithdrawalsByEpoch(validator []uint64, startEpoch uint64, endEpoch uint64) ([]*types.WithdrawalsByEpoch, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	var withdrawals []*types.WithdrawalsByEpoch

	err := ReaderDb.Select(&withdrawals, `
	SELECT
		w.validatorindex,
		w.block_slot / $4 as epoch,
		sum(w.amount) as amount
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1' AND b.slot >= $2 AND b.slot <= $3
	WHERE validatorindex = ANY($1)
	GROUP BY w.validatorindex, w.block_slot / $4
	ORDER BY w.block_slot / $4 DESC LIMIT 100`, pq.Array(validator), startEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, endEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch+utils.Config.Chain.ClConfig.SlotsPerEpoch-1, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validator: %d: %w", validator, err)
	}
	return withdrawals, nil
}

// GetAddressWithdrawalsTotal returns the total withdrawals for an address
func GetAddressWithdrawalsTotal(address []byte) (uint64, error) {
	var total uint64

	err := ReaderDb.Get(&total, `
	/*+
	BitmapScan(w)
	NestLoop(b w)
	*/
	SELECT
		COALESCE(sum(w.amount), 0) as total
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.address = $1`, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting blocks_withdrawals for address: %x: %w", address, err)
	}

	return total, nil
}

func GetDashboardWithdrawals(validators []uint64, limit uint64, offset uint64, orderBy string, orderDir string) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}
	validatorFilter := pq.Array(validators)
	err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
		/*+
		BitmapScan(w)
		NestLoop(b w)
		*/
		SELECT
			w.block_slot as slot,
			w.withdrawalindex as index,
			w.validatorindex,
			w.address,
			w.amount
		FROM blocks_withdrawals w
		INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
		WHERE validatorindex = ANY($1)
		ORDER BY  w.%s %s
		LIMIT $2 OFFSET $3`, orderBy, orderDir), validatorFilter, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting dashboard blocks_withdrawals for validators: %d: %w", validators, err)
	}

	return withdrawals, nil
}

func GetTotalWithdrawalsCount(validators []uint64) (uint64, error) {
	var count uint64
	validatorFilter := pq.Array(validators)
	lastExportedDay, err := GetLastExportedStatisticDay()
	if err != nil && err != ErrNoStats {
		return 0, fmt.Errorf("error getting latest exported statistic day for withdrawals count: %w", err)
	}

	cutoffSlot := uint64(0)
	if err == nil {
		_, lastEpochOfDay := utils.GetFirstAndLastEpochForDay(lastExportedDay)
		cutoffSlot = (lastEpochOfDay * utils.Config.Chain.ClConfig.SlotsPerEpoch) + 1
	}

	err = ReaderDb.Get(&count, `
		WITH today AS (
			SELECT COUNT(*) as count
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
			WHERE w.validatorindex = ANY($1) AND w.block_slot >= $2
		),
		stats AS (
			SELECT COALESCE(SUM(withdrawals_total), 0) as count
			FROM validator_stats
			WHERE validatorindex = ANY($1) AND day = $3
		)
		SELECT today.count + stats.count
		FROM today, stats`, validatorFilter, cutoffSlot, lastExportedDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting dashboard validator blocks_withdrawals count for validators: %d: %w", validators, err)
	}

	return count, nil
}

func GetLastWithdrawalEpoch(validators []uint64) (map[uint64]uint64, error) {
	var dbResponse []struct {
		ValidatorIndex     uint64 `db:"validatorindex"`
		LastWithdrawalSlot uint64 `db:"last_withdawal_slot"`
	}

	res := make(map[uint64]uint64)
	err := ReaderDb.Select(&dbResponse, `
		SELECT w.validatorindex as validatorindex, COALESCE(max(block_slot), 0) as last_withdawal_slot
		FROM blocks_withdrawals w
		INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
		WHERE w.validatorindex = ANY($1)
		GROUP BY w.validatorindex`, validators)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, nil
		}
		return nil, fmt.Errorf("error getting validator blocks_withdrawals count for validators: %d: %w", validators, err)
	}

	for _, row := range dbResponse {
		res[row.ValidatorIndex] = row.LastWithdrawalSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	}

	return res, nil
}

func GetMostRecentWithdrawalValidator() (uint64, error) {
	var validatorindex uint64

	err := ReaderDb.Get(&validatorindex, `
	SELECT
		w.validatorindex
	FROM
		blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	ORDER BY
		withdrawalindex DESC LIMIT 1;`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting most recent blocks_withdrawals validatorindex: %w", err)
	}

	return validatorindex, nil
}

// get all ad configurations
func GetAdConfigurations() ([]*types.AdConfig, error) {
	var adConfigs []*types.AdConfig

	err := ReaderDb.Select(&adConfigs, `
	SELECT
		id,
		template_id,
		jquery_selector,
		insert_mode,
		refresh_interval,
		enabled,
		for_all_users,
		banner_id,
		html_content
	FROM
		ad_configurations`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.AdConfig{}, nil
		}
		return nil, fmt.Errorf("error getting ad configurations: %w", err)
	}

	return adConfigs, nil
}

// get the ad configuration for a specific template that are active
func GetAdConfigurationsForTemplate(ids []string, noAds bool) ([]*types.AdConfig, error) {
	var adConfigs []*types.AdConfig
	forAllUsers := ""
	if noAds {
		forAllUsers = " AND for_all_users = true"
	}
	err := ReaderDb.Select(&adConfigs, fmt.Sprintf(`
	SELECT
		id,
		template_id,
		jquery_selector,
		insert_mode,
		refresh_interval,
		enabled,
		for_all_users,
		banner_id,
		html_content
	FROM
		ad_configurations
	WHERE
		template_id = ANY($1) AND
		enabled = true %v`, forAllUsers), pq.Array(ids))
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.AdConfig{}, nil
		}
		return nil, fmt.Errorf("error getting ad configurations for template: %v %s", err, ids)
	}

	return adConfigs, nil
}

// insert new ad configuration
func InsertAdConfigurations(adConfig types.AdConfig) error {
	_, err := WriterDb.Exec(`
		INSERT INTO ad_configurations (
			id,
			template_id,
			jquery_selector,
			insert_mode,
			refresh_interval,
			enabled,
			for_all_users,
			banner_id,
			html_content)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING`,
		adConfig.Id,
		adConfig.TemplateId,
		adConfig.JQuerySelector,
		adConfig.InsertMode,
		adConfig.RefreshInterval,
		adConfig.Enabled,
		adConfig.ForAllUsers,
		adConfig.BannerId,
		adConfig.HtmlContent)
	if err != nil {
		return fmt.Errorf("error inserting ad configuration: %w", err)
	}
	return nil
}

// update exisiting ad configuration
func UpdateAdConfiguration(adConfig types.AdConfig) error {
	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)
	_, err = tx.Exec(`
		UPDATE ad_configurations SET
			template_id = $2,
			jquery_selector = $3,
			insert_mode = $4,
			refresh_interval = $5,
			enabled = $6,
			for_all_users = $7,
			banner_id = $8,
			html_content = $9
		WHERE id = $1;`,
		adConfig.Id,
		adConfig.TemplateId,
		adConfig.JQuerySelector,
		adConfig.InsertMode,
		adConfig.RefreshInterval,
		adConfig.Enabled,
		adConfig.ForAllUsers,
		adConfig.BannerId,
		adConfig.HtmlContent)
	if err != nil {
		return fmt.Errorf("error updating ad configuration: %w", err)
	}
	return tx.Commit()
}

// delete ad configuration
func DeleteAdConfiguration(id string) error {
	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)

	// delete ad configuration
	_, err = WriterDb.Exec(`
		DELETE FROM ad_configurations
		WHERE
			id = $1;`,
		id)
	return err
}

// get all explorer configurations
func GetExplorerConfigurations() ([]*types.ExplorerConfig, error) {
	var configs []*types.ExplorerConfig

	err := ReaderDb.Select(&configs, `
	SELECT
		category,
		key,
		value,
		data_type
	FROM
		explorer_configurations`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.ExplorerConfig{}, nil
		}
		return nil, fmt.Errorf("error getting explorer configurations: %w", err)
	}

	return configs, nil
}

// save current configurations
func SaveExplorerConfiguration(configs []types.ExplorerConfig) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	for i, config := range configs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%v, $%v, $%v, $%v)", i*4+1, i*4+2, i*4+3, i*4+4))

		valueArgs = append(valueArgs, config.Category)
		valueArgs = append(valueArgs, config.Key)
		valueArgs = append(valueArgs, config.Value)
		valueArgs = append(valueArgs, config.DataType)
	}
	query := fmt.Sprintf(`
		INSERT INTO explorer_configurations (
			category,
			key,
			value,
			data_type)
    	VALUES %s
		ON CONFLICT
			(category, key)
		DO UPDATE SET
			value = excluded.value,
			data_type = excluded.data_type
			`, strings.Join(valueStrings, ","))

	_, err := WriterDb.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error inserting/updating explorer configurations: %w", err)
	}
	return nil
}

func GetTotalBLSChanges() (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, `
		SELECT count(*) FROM blocks_bls_change`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting total blocks_bls_change: %w", err)
	}

	return count, nil
}

func GetBLSChangesCountForQuery(query string) (uint64, error) {
	count := uint64(0)

	blsQuery := `
		SELECT COUNT(*) FROM (
			SELECT b.slot
			FROM blocks_bls_change bls
			INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
			%s
			LIMIT %d
		) a
		`

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var err error = nil

	if utils.IsHash(query) {
		searchQuery := `WHERE bls.pubkey = $1`
		pubkey, decErr := hex.DecodeString(trimmedQuery)
		if decErr != nil {
			return 0, decErr
		}
		err = ReaderDb.Get(&count, fmt.Sprintf(blsQuery, searchQuery, BlsChangeQueryLimit),
			pubkey)
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
		// Check whether the query can be used for a validator, slot or epoch search
		searchQuery := `
			WHERE bls.validatorindex = $1
				OR bls.block_slot = $1
				OR bls.block_slot BETWEEN $1*$2 AND ($1+1)*$2-1`
		err = ReaderDb.Get(&count, fmt.Sprintf(blsQuery, searchQuery, BlsChangeQueryLimit),
			uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetBLSChanges(query string, length, start uint64, orderBy, orderDir string) ([]*types.BLSChange, error) {
	blsChange := []*types.BLSChange{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "validatorindex"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	blsQuery := `
		SELECT
			bls.block_slot as slot,
			bls.validatorindex,
			bls.signature,
			bls.pubkey,
			bls.address
		FROM blocks_bls_change bls
		INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
		%s
		ORDER BY bls.%s %s
		LIMIT $1
		OFFSET $2`

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var err error = nil

	if trimmedQuery != "" {
		if utils.IsHash(query) {
			searchQuery := `WHERE bls.pubkey = $3`
			pubkey, decErr := hex.DecodeString(trimmedQuery)
			if decErr != nil {
				return nil, decErr
			}
			err = ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, searchQuery, orderBy, orderDir),
				length, start, pubkey)
		} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
			// Check whether the query can be used for a validator, slot or epoch search
			searchQuery := `
				WHERE bls.validatorindex = $3
					OR bls.block_slot = $3
					OR bls.block_slot BETWEEN $3*$4 AND ($3+1)*$4-1`
			err = ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, searchQuery, orderBy, orderDir),
				length, start, uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
		if err != nil {
			return nil, err
		}
	} else {
		err := ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, "", orderBy, orderDir), length, start)
		if err != nil {
			return nil, err
		}
	}

	return blsChange, nil
}

func GetSlotBLSChange(slot uint64) ([]*types.BLSChange, error) {
	var change []*types.BLSChange

	err := ReaderDb.Select(&change, `
	SELECT
		bls.validatorindex,
		bls.signature,
		bls.pubkey,
		bls.address
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	WHERE block_slot = $1
	ORDER BY bls.validatorindex`, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return change, nil
		}
		return nil, fmt.Errorf("error getting slot blocks_bls_change: %w", err)
	}

	return change, nil
}

func GetValidatorBLSChange(validatorindex uint64) (*types.BLSChange, error) {
	change := &types.BLSChange{}

	err := ReaderDb.Get(change, `
	SELECT
		bls.block_slot as slot,
		bls.signature,
		bls.pubkey,
		bls.address
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	WHERE validatorindex = $1
	ORDER BY bls.block_slot`, validatorindex)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting validator blocks_bls_change: %w", err)
	}

	return change, nil
}

// GetValidatorsBLSChange returns the BLS change for a list of validators
func GetValidatorsBLSChange(validators []uint64) ([]*types.ValidatorsBLSChange, error) {
	change := make([]*types.ValidatorsBLSChange, 0, len(validators))

	err := ReaderDb.Select(&change, `
	SELECT
		bls.block_slot AS slot,
		bls.block_root,
		bls.signature,
		bls.pubkey,
		bls.validatorindex,
		bls.address,
		d.withdrawalcredentials
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	LEFT JOIN validators v ON v.validatorindex = bls.validatorindex
	LEFT JOIN (
		SELECT ROW_NUMBER() OVER (PARTITION BY publickey ORDER BY block_slot) AS rn, withdrawalcredentials, publickey, block_root FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'
	) AS d ON d.publickey = v.pubkey AND rn = 1
	WHERE bls.validatorindex = ANY($1)
	ORDER BY bls.block_slot DESC
	`, pq.Array(validators))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting validators blocks_bls_change: %w", err)
	}

	return change, nil
}

func GetWithdrawableValidatorCount(epoch uint64) (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, `
	SELECT
		count(*)
	FROM
		validators
	INNER JOIN (
		SELECT validatorindex,
                end_effective_balance,
                end_balance,
                DAY
        FROM
                validator_stats
        WHERE DAY = (SELECT COALESCE(MAX(day), 0) FROM validator_stats_status)) as stats
	ON stats.validatorindex = validators.validatorindex
	WHERE
		validators.withdrawalcredentials LIKE '\x01' || '%'::bytea AND ((stats.end_effective_balance = $1 AND stats.end_balance > $1) OR (validators.withdrawableepoch <= $2 AND stats.end_balance > 0));`, utils.Config.Chain.ClConfig.MaxEffectiveBalance, epoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting withdrawable validator count: %w", err)
	}

	return count, nil
}

func GetPendingBLSChangeValidatorCount() (uint64, error) {
	var count uint64

	err := ReaderDb.Get(&count, `
	SELECT
		count(*)
	FROM
		validators
	WHERE
		withdrawalcredentials LIKE '\x00' || '%'::bytea`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting withdrawable validator count: %w", err)
	}

	return count, nil
}

func GetLastExportedStatisticDay() (uint64, error) {
	var lastStatsDay sql.NullInt64
	err := ReaderDb.Get(&lastStatsDay, "SELECT MAX(day) FROM validator_stats_status WHERE status")

	if err != nil {
		return 0, fmt.Errorf("error getting lastStatsDay %v", err)
	}

	if !lastStatsDay.Valid {
		return 0, ErrNoStats
	}
	return uint64(lastStatsDay.Int64), nil
}

func GetTotalValidatorDeposits(validators []uint64, totalDeposits *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(totalDeposits, `
		SELECT
			COALESCE(SUM(amount), 0)
		FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'
		WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))
	`, validatorsPQArray)
}

func GetFirstActivationEpoch(validators []uint64, firstActivationEpoch *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(firstActivationEpoch, `
		SELECT
			activationepoch
		FROM validators
		WHERE validatorindex = ANY($1)
		ORDER BY activationepoch LIMIT 1
	`, validatorsPQArray)
}

func GetValidatorDepositsForSlots(validators []uint64, fromSlot uint64, toSlot uint64, deposits *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(deposits, `
		SELECT
			COALESCE(SUM(amount), 0)
		FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' and b.slot >= $2 and b.slot <= $3
		WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))
	`, validatorsPQArray, fromSlot, toSlot)
}

func GetValidatorWithdrawalsForSlots(validators []uint64, fromSlot uint64, toSlot uint64, withdrawals *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(withdrawals, `
		SELECT
			COALESCE(SUM(amount), 0)
		FROM blocks_withdrawals d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' and b.slot >= $2 and b.slot <= $3
		WHERE validatorindex = ANY($1)
	`, validatorsPQArray, fromSlot, toSlot)
}

func GetValidatorBalanceForDay(validators []uint64, day uint64, balance *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(balance, `
		SELECT
			COALESCE(SUM(end_balance), 0)
		FROM validator_stats
		WHERE validatorindex = ANY($1) AND day = $2
	`, validatorsPQArray, day)
}

func GetValidatorActivationBalance(validators []uint64, balance *uint64) error {
	if len(validators) == 0 {
		return fmt.Errorf("passing empty validator array is unsupported")
	}

	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(balance, `
		SELECT
			SUM(balanceactivation)
		FROM validators
		WHERE validatorindex = ANY($1)
	`, validatorsPQArray)
}

func GetValidatorProposals(validators []uint64, proposals *[]types.ValidatorProposalInfo) error {
	validatorsPQArray := pq.Array(validators)

	return ReaderDb.Select(proposals, `
		SELECT
			slot,
			status,
			COALESCE(exec_block_number, 0) as exec_block_number
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot ASC
		`, validatorsPQArray)
}

func GetValidatorDutiesInfo(startSlot uint64) ([]types.ValidatorDutyInfo, error) {
	validatorDutyInfo := []types.ValidatorDutyInfo{}

	err := ReaderDb.Select(&validatorDutyInfo, `
		SELECT
			blocks.slot,
			blocks.status,
			COALESCE(blocks.exec_block_number, 0) AS exec_block_number,
			blocks.syncaggregate_bits,
			blocks_attestations.validators,
			blocks_attestations.slot AS attested_slot,
			blocks.proposerslashingscount,
			blocks.attesterslashingscount
		FROM blocks
		LEFT JOIN blocks_attestations ON blocks.slot = blocks_attestations.block_slot
		WHERE blocks.slot >= $1
		`, startSlot)

	return validatorDutyInfo, err
}

func GetMissedSlots(slots []uint64) ([]uint64, error) {
	slotsPQArray := pq.Array(slots)
	missed := []uint64{}

	err := ReaderDb.Select(&missed, `
		SELECT
			slot
		FROM blocks
		WHERE slot = ANY($1) AND status = '2'
		`, slotsPQArray)

	return missed, err
}

func GetMissedSlotsMap(slots []uint64) (map[uint64]bool, error) {
	missedSlots, err := GetMissedSlots(slots)
	if err != nil {
		return nil, err
	}
	missedSlotsMap := make(map[uint64]bool, len(missedSlots))
	for _, slot := range missedSlots {
		missedSlotsMap[slot] = true
	}
	return missedSlotsMap, nil
}

func GetOrphanedSlots(slots []uint64) ([]uint64, error) {
	slotsPQArray := pq.Array(slots)
	orphaned := []uint64{}

	err := ReaderDb.Select(&orphaned, `
		SELECT
			slot
		FROM blocks
		WHERE slot = ANY($1) AND status = '3'
		`, slotsPQArray)

	return orphaned, err
}

func GetOrphanedSlotsMap(slots []uint64) (map[uint64]bool, error) {
	orphanedSlots, err := GetOrphanedSlots(slots)
	if err != nil {
		return nil, err
	}
	orphanedSlotsMap := make(map[uint64]bool, len(orphanedSlots))
	for _, slot := range orphanedSlots {
		orphanedSlotsMap[slot] = true
	}
	return orphanedSlotsMap, nil
}

func GetBlockStatus(block int64, latestFinalizedEpoch uint64, epochInfo *types.EpochInfo) error {
	return ReaderDb.Get(epochInfo, `
				SELECT (epochs.epoch <= $2) AS finalized, epochs.globalparticipationrate
				FROM blocks
				LEFT JOIN epochs ON blocks.epoch = epochs.epoch
				WHERE blocks.exec_block_number = $1
				AND blocks.status='1'`,
		block, latestFinalizedEpoch)
}

func GetSyncCommitteeValidators(readerDb *sqlx.DB, epoch uint64) ([]uint64, error) {
	validatoridxs := []uint64{}

	err := readerDb.Select(&validatoridxs, `
			SELECT
				validatorindex
			FROM sync_committees
			WHERE period = $1
			ORDER BY committeeindex`,
		utils.SyncPeriodOfEpoch(epoch))
	if err != nil {
		return nil, err
	}

	return validatoridxs, nil
}

// Returns the participation rate for every slot between startSlot and endSlot (both inclusive) as a map with the slot as key
//
// If a slot is missed, the map will not contain an entry for it
func GetSyncParticipationBySlotRange(startSlot, endSlot uint64) (map[uint64]uint64, error) {
	rows := []struct {
		Slot         uint64
		Participated uint64
	}{}

	err := ReaderDb.Select(&rows, `SELECT slot, syncaggregate_participation * $1 AS participated FROM blocks WHERE slot >= $2 AND slot <= $3 AND status = '1'`,
		utils.Config.Chain.ClConfig.SyncCommitteeSize,
		startSlot,
		endSlot)

	if err != nil {
		return nil, err
	}

	ret := make(map[uint64]uint64)

	for _, row := range rows {
		ret[row.Slot] = row.Participated
	}

	return ret, nil
}

// Should be used when retrieving data for a very large amount of validators (for the notifications process)
func GetValidatorAttestationHistoryForNotifications(startEpoch uint64, endEpoch uint64) (map[types.Epoch]map[types.ValidatorIndex]bool, error) {
	// first retrieve activation & exit epoch for all validators
	activityData := []struct {
		ValidatorIndex  types.ValidatorIndex
		ActivationEpoch types.Epoch
		ExitEpoch       types.Epoch
	}{}

	err := ReaderDb.Select(&activityData, "SELECT validatorindex, activationepoch, exitepoch FROM validators ORDER BY validatorindex;")
	if err != nil {
		return nil, fmt.Errorf("error retrieving activation & exit epoch for validators: %w", err)
	}

	log.Infof("retrieved activation & exit epochs")

	// next retrieve all attestation data from the db (need to retrieve data for the endEpoch+1 epoch as that could still contain attestations for the endEpoch)
	firstSlot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	lastSlot := ((endEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)
	lastQuerySlot := ((endEpoch+2)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)

	rows, err := ReaderDb.Query(`SELECT
	blocks_attestations.slot,
	validators
	FROM blocks_attestations
	LEFT JOIN blocks ON blocks_attestations.block_root = blocks.blockroot WHERE
	blocks_attestations.block_slot >= $1 AND blocks_attestations.block_slot <= $2 AND blocks.status = '1' ORDER BY block_slot`, firstSlot, lastQuerySlot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving attestation data from the db: %w", err)
	}
	defer rows.Close()

	log.Infof("retrieved attestation raw data")

	// next process the data and fill up the epoch participation
	// validators that participated in an epoch will have the flag set to true
	// validators that missed their participation will have it set to false
	epochParticipation := make(map[types.Epoch]map[types.ValidatorIndex]bool)
	for rows.Next() {
		var slot types.Slot
		var attestingValidators pq.Int64Array

		err := rows.Scan(&slot, &attestingValidators)
		if err != nil {
			return nil, fmt.Errorf("error scanning attestation data: %w", err)
		}

		if slot < types.Slot(firstSlot) || slot > types.Slot(lastSlot) {
			continue
		}

		epoch := types.Epoch(utils.EpochOfSlot(uint64(slot)))

		participation := epochParticipation[epoch]

		if participation == nil {
			epochParticipation[epoch] = make(map[types.ValidatorIndex]bool)

			// log.LogInfo("seeding validator duties for epoch %v", epoch)
			for _, data := range activityData {
				if data.ActivationEpoch <= epoch && epoch < data.ExitEpoch {
					epochParticipation[epoch][data.ValidatorIndex] = false
				}
			}

			participation = epochParticipation[epoch]
		}

		for _, validator := range attestingValidators {
			participation[types.ValidatorIndex(validator)] = true
		}
	}

	return epochParticipation, nil
}

func CacheQuery(query string, viewName string, indexes ...[]string) error {
	tmpViewName := "_tmp_" + viewName
	trashViewName := "_trash_" + viewName
	tx, err := AlloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting tx: %w", err)
	}
	defer utils.Rollback(tx)

	// pre-cleanup
	_, err = tx.Exec(fmt.Sprintf(`drop materialized view if exists %s`, tmpViewName))
	if err != nil {
		return fmt.Errorf("error dropping %s materialized view: %w", tmpViewName, err)
	}
	_, err = tx.Exec(fmt.Sprintf("drop materialized view if exists %s", trashViewName))
	if err != nil {
		return fmt.Errorf("error dropping %s materialized view: %w", trashViewName, err)
	}
	// create the new view
	_, err = tx.Exec(fmt.Sprintf(`CREATE MATERIALIZED VIEW %s AS %s`, tmpViewName, query))
	if err != nil {
		return fmt.Errorf("error creating %s materialized view: %w", tmpViewName, err)
	}
	tmpIndexNames := make([]string, len(indexes))
	for i, index := range indexes {
		tmpIndexNames[i] = fmt.Sprintf("%s_%d_idx", tmpViewName, i)
		_, err = tx.Exec(fmt.Sprintf("CREATE INDEX %s ON %s (%s)", tmpIndexNames[i], tmpViewName, strings.Join(index, ",")))
		if err != nil {
			return fmt.Errorf("error creating index %s over columns %v: %w", tmpIndexNames[i], index, err)
		}
	}
	// fix permissions
	_, err = tx.Exec(fmt.Sprintf("GRANT SELECT ON %s TO readaccess;", tmpViewName))
	if err != nil {
		return fmt.Errorf("error granting select on %s materialized view: %w", tmpViewName, err)
	}
	_, err = tx.Exec(fmt.Sprintf("GRANT ALL ON %s TO alloydbsuperuser;", tmpViewName))
	if err != nil {
		return fmt.Errorf("error granting all on %s materialized view: %w", tmpViewName, err)
	}
	// swap views
	_, err = tx.Exec(fmt.Sprintf(`ALTER MATERIALIZED VIEW if exists %s RENAME TO %s;`, viewName, trashViewName))
	if err != nil {
		return fmt.Errorf("error renaming existing %s materialized view: %w", viewName, err)
	}
	_, err = tx.Exec(fmt.Sprintf(`ALTER MATERIALIZED VIEW %s RENAME TO %s;`, tmpViewName, viewName))
	if err != nil {
		return fmt.Errorf("error renaming %s materialized view: %w", tmpViewName, err)
	}
	// drop old view
	_, err = tx.Exec(fmt.Sprintf("drop materialized view if exists %s", trashViewName))
	if err != nil {
		return fmt.Errorf("error dropping %s materialized view: %w", trashViewName, err)
	}
	// rename indexes
	for i := range indexes {
		indexName := fmt.Sprintf("%s_%d_idx", viewName, i)
		_, err = tx.Exec(fmt.Sprintf("ALTER INDEX %s RENAME TO %s;", tmpIndexNames[i], indexName))
		if err != nil {
			return fmt.Errorf("error renaming index %s: %w", tmpIndexNames[i], err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx: %w", err)
	}
	return nil
}
