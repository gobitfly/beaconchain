package main

import (
	"bytes"
	"context"

	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/cmd/misc/commands"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/gobitfly/beaconchain/pkg/exporter/services"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	go_ens "github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"flag"

	"github.com/Gurpartap/storekit-go"
)

var opts = struct {
	Command             string
	User                uint64
	Addresses           string
	TargetVersion       int64
	TargetDatabase      string
	StartEpoch          uint64
	EndEpoch            uint64
	StartDay            uint64
	EndDay              uint64
	Validator           uint64
	StartBlock          uint64
	EndBlock            uint64
	BatchSize           uint64
	DataConcurrency     uint64
	Transformers        string
	Table               string
	Columns             string
	Family              string
	Key                 string
	ValidatorNameRanges string
	DryRun              bool
}{}

func main() {
	statsPartitionCommand := commands.StatsMigratorCommand{}

	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, initBigtableSchema, epoch-export, debug-rewards, debug-blocks, clear-bigtable, index-old-eth1-blocks, update-aggregation-bits, historic-prices-export, index-missing-blocks, export-epoch-missed-slots, migrate-last-attestation-slot-bigtable, export-genesis-validators, update-block-finalization-sequentially, nameValidatorsByRanges, export-stats-totals, export-sync-committee-periods, export-sync-committee-validator-stats, partition-validator-stats, migrate-app-purchases")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.Int64Var(&opts.TargetVersion, "target-version", 0, "Db migration target version. `-3` downgrades the database by one version, `-2` upgrades to the latest version, `-1` upgrades by one version, other negative numbers downgrade to their absolute value, and positive numbers upgrade to their specified version.")
	flag.StringVar(&opts.TargetDatabase, "target-database", "", "Database to apply the schema to")
	flag.StringVar(&opts.Table, "table", "", "bigtable table")
	flag.StringVar(&opts.Family, "family", "", "big table family")
	flag.StringVar(&opts.Key, "key", "", "big table key")
	flag.Uint64Var(&opts.StartBlock, "blocks.start", 0, "Block to start indexing")
	flag.Uint64Var(&opts.EndBlock, "blocks.end", 0, "Block to finish indexing")
	flag.Uint64Var(&opts.DataConcurrency, "data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	flag.Uint64Var(&opts.BatchSize, "data.batchSize", 1000, "Batch size")
	flag.StringVar(&opts.Transformers, "transformers", "", "Comma separated list of transformers used by the eth1 indexer")
	flag.StringVar(&opts.ValidatorNameRanges, "validator-name-ranges", "https://config.dencun-devnet-8.ethpandaops.io/api/v1/nodes/validator-ranges", "url to or json of validator-ranges (format must be: {'ranges':{'X-Y':'name'}})")
	flag.StringVar(&opts.Addresses, "addresses", "", "Comma separated list of addresses that should be processed by the command")
	flag.StringVar(&opts.Columns, "columns", "", "Comma separated list of columns that should be affected by the command")
	dryRun := flag.String("dry-run", "true", "if 'false' it deletes all rows starting with the key, per default it only logs the rows that would be deleted, but does not really delete them")
	versionFlag := flag.Bool("version", false, "Show version and exit")

	statsPartitionCommand.ParseCommandOptions()
	flag.Parse()

	if *versionFlag {
		log.Infof(version.Version)
		return
	}

	opts.DryRun = *dryRun != "false"

	log.InfoWithFields(map[string]interface{}{"config": *configPath, "version": version.Version}, "starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	chainIdString := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Fatal(err, "error initializing bigtable", 0)
	}

	cl := consapi.NewClient("http://" + cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
	nodeImpl, ok := cl.ClientInt.(*consapi.NodeClient)
	if !ok {
		log.Fatal(nil, "lighthouse client can only be used with real node impl", 0)
	}
	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient(nodeImpl, chainIDBig)
	if err != nil {
		log.Fatal(err, "lighthouse client error", 0)
	}

	erigonClient, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		log.Fatal(err, "error initializing erigon client", 0)
	}

	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
	}, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	db.FrontendWriterDB, db.FrontendReaderDB = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.Frontend.WriterDatabase.Username,
		Password:     cfg.Frontend.WriterDatabase.Password,
		Name:         cfg.Frontend.WriterDatabase.Name,
		Host:         cfg.Frontend.WriterDatabase.Host,
		Port:         cfg.Frontend.WriterDatabase.Port,
		MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.Frontend.ReaderDatabase.Username,
		Password:     cfg.Frontend.ReaderDatabase.Password,
		Name:         cfg.Frontend.ReaderDatabase.Name,
		Host:         cfg.Frontend.ReaderDatabase.Host,
		Port:         cfg.Frontend.ReaderDatabase.Port,
		MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
	}, "pgx", "postgres")
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	// clickhouse
	//nolint:forbidigo
	db.ClickHouseWriter, db.ClickHouseReader = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.ClickHouse.WriterDatabase.Username,
		Password:     cfg.ClickHouse.WriterDatabase.Password,
		Name:         cfg.ClickHouse.WriterDatabase.Name,
		Host:         cfg.ClickHouse.WriterDatabase.Host,
		Port:         cfg.ClickHouse.WriterDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.WriterDatabase.MaxOpenConns,
		SSL:          true,
		MaxIdleConns: cfg.ClickHouse.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ClickHouse.ReaderDatabase.Username,
		Password:     cfg.ClickHouse.ReaderDatabase.Password,
		Name:         cfg.ClickHouse.ReaderDatabase.Name,
		Host:         cfg.ClickHouse.ReaderDatabase.Host,
		Port:         cfg.ClickHouse.ReaderDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.ReaderDatabase.MaxOpenConns,
		SSL:          true,
		MaxIdleConns: cfg.ClickHouse.ReaderDatabase.MaxIdleConns,
	}, "clickhouse", "clickhouse")
	defer db.ClickHouseReader.Close()
	defer db.ClickHouseWriter.Close() //nolint:forbidigo

	// Initialize the persistent redis client
	rdc := redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisSessionStoreEndpoint,
		ReadTimeout: time.Second * 20,
	})

	if err := rdc.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err, "error connecting to persistent redis store", 0)
	}

	db.PersistentRedisDbClient = rdc
	defer db.PersistentRedisDbClient.Close()

	switch opts.Command {
	case "nameValidatorsByRanges":
		err := nameValidatorsByRanges(opts.ValidatorNameRanges)
		if err != nil {
			log.Fatal(err, "error naming validators by ranges", 0)
		}
	case "updateAPIKey":
		err := updateAPIKey(opts.User)
		if err != nil {
			log.Fatal(err, "error updating API key", 0)
		}
	case "applyDbSchema":
		log.Infof("applying db schema")
		// require that version is set. require that database name is set
		if opts.TargetVersion == 0 {
			log.Fatal(nil, "target version must be set", 0)
		}
		if opts.TargetDatabase == "" {
			log.Fatal(nil, "target database must be set", 0)
		}
		err := db.ApplyEmbeddedDbSchema(opts.TargetVersion, opts.TargetDatabase)
		if err != nil {
			log.Fatal(err, "error applying db schema", 0)
		}
		log.Infof("db schema applied successfully")
	case "initBigtableSchema":
		log.Infof("initializing bigtable schema")
		err := db.InitBigtableSchema()
		if err != nil {
			log.Fatal(err, "error initializing bigtable schema", 0)
		}
		log.Infof("bigtable schema initialization completed")
	case "epoch-export":
		log.Infof("exporting epochs %v - %v", opts.StartEpoch, opts.EndEpoch)
		for epoch := opts.StartEpoch; epoch <= opts.EndEpoch; epoch++ {
			tx, err := db.WriterDb.Beginx()
			if err != nil {
				log.Fatal(err, "error starting tx", 0)
			}
			for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
				err = modules.ExportSlot(rpcClient, slot, false, tx)

				if err != nil {
					_ = tx.Rollback()
					log.Fatal(err, "error exporting slot", 0, map[string]interface{}{"slot": slot})
				}
				log.Infof("finished export for slot %v", slot)
			}
			err = tx.Commit()
			if err != nil {
				log.Fatal(err, "error committing tx", 0)
			}
		}
	case "export-epoch-missed-slots":
		log.Infof("exporting epochs with missed slots")
		latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			log.Error(err, "error getting latest finalized epoch from db", 0)
		}
		epochs := []uint64{}
		err = db.ReaderDb.Select(&epochs, `
			WITH last_exported_epoch AS (
				SELECT (MAX(epoch)*$1) AS slot
				FROM epochs
				WHERE epoch <= $2
				AND rewards_exported
			)
			SELECT epoch
			FROM blocks
			WHERE status = '0'
				AND slot < (SELECT slot FROM last_exported_epoch)
			GROUP BY epoch
			ORDER BY epoch;
		`, utils.Config.Chain.ClConfig.SlotsPerEpoch, latestFinalizedEpoch)
		if err != nil {
			log.Error(err, "Error getting epochs with missing slot status from db", 0)
			return
		} else if len(epochs) == 0 {
			log.Infof("No epochs with missing slot status found")
			return
		}

		log.Infof("Found %v epochs with missing slot status", len(epochs))
		for _, epoch := range epochs {
			tx, err := db.WriterDb.Beginx()
			if err != nil {
				log.Fatal(err, "error starting tx", 0)
			}
			for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
				err = modules.ExportSlot(rpcClient, slot, false, tx)

				if err != nil {
					_ = tx.Rollback()
					log.Fatal(err, "error exporting slot", 0, map[string]interface{}{"slot": slot})
				}
				log.Infof("finished export for slot %v", slot)
			}
			err = tx.Commit()
			if err != nil {
				log.Fatal(err, "error committing tx", 0)
			}
		}
	case "debug-rewards":
		compareRewards(opts.StartDay, opts.EndDay, opts.Validator, bt)
	case "debug-blocks":
		err = debugBlocks(rpcClient)
	case "clear-bigtable":
		clearBigtable(opts.Table, opts.Family, opts.Columns, opts.Key, opts.DryRun, bt)
	case "index-old-eth1-blocks":
		indexOldEth1Blocks(opts.StartBlock, opts.EndBlock, opts.BatchSize, opts.DataConcurrency, opts.Transformers, bt, erigonClient)
	case "update-aggregation-bits":
		updateAggreationBits(rpcClient, opts.StartEpoch, opts.EndEpoch, opts.DataConcurrency)
	case "update-block-finalization-sequentially":
		err = updateBlockFinalizationSequentially()
	case "historic-prices-export":
		exportHistoricPrices(opts.StartDay, opts.EndDay)
	case "index-missing-blocks":
		indexMissingBlocks(opts.StartBlock, opts.EndBlock, bt, erigonClient)
	case "migrate-last-attestation-slot-bigtable":
		migrateLastAttestationSlotToBigtable()
	case "migrate-app-purchases":
		err = migrateAppPurchases(opts.Key)
	case "export-genesis-validators":
		log.Infof("retrieving genesis validator state")
		validators, err := rpcClient.GetValidatorState(0)
		if err != nil {
			log.Fatal(fmt.Errorf("error retrieving genesis validator state"), "", 0)
		}

		validatorsArr := make([]*types.Validator, 0, len(validators.Data))

		for _, validator := range validators.Data {
			validatorsArr = append(validatorsArr, &types.Validator{
				Index:                      validator.Index,
				PublicKey:                  validator.Validator.Pubkey,
				WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
				Balance:                    validator.Balance,
				EffectiveBalance:           validator.Validator.EffectiveBalance,
				Slashed:                    validator.Validator.Slashed,
				ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
				ActivationEpoch:            validator.Validator.ActivationEpoch,
				ExitEpoch:                  validator.Validator.ExitEpoch,
				WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
				Status:                     "active_online",
			})
		}

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			log.Fatal(err, "error starting tx", 0)
		}
		defer func() {
			err := tx.Rollback()
			if err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		batchSize := 10000
		for i := 0; i < len(validatorsArr); i += batchSize {
			data := &types.EpochData{
				SyncDuties:        make(map[types.Slot]map[types.ValidatorIndex]bool),
				AttestationDuties: make(map[types.Slot]map[types.ValidatorIndex][]types.Slot),
				ValidatorAssignmentes: &types.EpochAssignments{
					ProposerAssignments: map[uint64]uint64{},
					AttestorAssignments: map[string]uint64{},
					SyncAssignments:     make([]uint64, 0),
				},
				Blocks:                  make(map[uint64]map[string]*types.Block),
				FutureBlocks:            make(map[uint64]map[string]*types.Block),
				EpochParticipationStats: &types.ValidatorParticipation{},
				Finalized:               false,
			}

			data.Validators = make([]*types.Validator, 0, batchSize)

			start := i
			end := i + batchSize
			if end >= len(validatorsArr) {
				end = len(validatorsArr) - 1
			}
			data.Validators = append(data.Validators, validatorsArr[start:end]...)

			log.Infof("saving validators %v-%v", data.Validators[0].Index, data.Validators[len(data.Validators)-1].Index)

			err = edb.SaveValidators(0, data.Validators, rpcClient, len(data.Validators), tx)
			if err != nil {
				log.Fatal(err, "error saving validators", 0)
			}
		}

		log.Infof("exporting deposit data for genesis %v validators", len(validators.Data))
		for i, validator := range validators.Data {
			if i%1000 == 0 {
				log.Infof("exporting deposit data for genesis validator %v (of %v/%v)", validator.Index, i, len(validators.Data))
			}
			_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
			VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
				validator.Index, validator.Validator.Pubkey, validator.Validator.WithdrawalCredentials, validator.Balance, []byte{0x0},
			)
			if err != nil {
				log.Error(err, "error exporting genesis-deposits", 0)
				time.Sleep(time.Minute)
				continue
			}
		}

		_, err = tx.Exec(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, syncaggregate_participation, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, withdrawalcount, voluntaryexitscount, proposer, status, exec_transactions_count, eth1data_depositcount)
		VALUES (0, 0, '\x'::bytea, '\x'::bytea, '\x'::bytea, '\x'::bytea, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
		if err != nil {
			log.Fatal(err, "error saving block to db", 0)
		}

		err = db.BigtableClient.SaveValidatorBalances(0, validatorsArr)
		if err != nil {
			log.Fatal(err, "error saving validator balances", 0)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err, "error committing tx", 0)
		}

	case "export-stats-totals":
		exportStatsTotals(opts.Columns, opts.StartDay, opts.EndDay, opts.DataConcurrency)
	case "export-sync-committee-periods":
		exportSyncCommitteePeriods(rpcClient, opts.StartDay, opts.EndDay, opts.DryRun)
	case "export-sync-committee-validator-stats":
		exportSyncCommitteeValidatorStats(rpcClient, opts.StartDay, opts.EndDay, opts.DryRun, true)
	case "fix-exec-transactions-count":
		err = fixExecTransactionsCount()
	case "partition-validator-stats":
		statsPartitionCommand.Config.DryRun = opts.DryRun
		err = statsPartitionCommand.StartStatsPartitionCommand()
	case "fix-ens":
		err = fixEns(erigonClient)
	case "fix-ens-addresses":
		err = fixEnsAddresses(erigonClient)
	default:
		log.Fatal(nil, fmt.Sprintf("unknown command %s", opts.Command), 0)
	}

	if err != nil {
		log.Fatal(err, "command returned error", 0)
	} else {
		log.Infof("command executed successfully")
	}
}

func fixEns(erigonClient *rpc.ErigonClient) error {
	log.Infof("command: fix-ens")
	addrs := []struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}{}
	err := db.WriterDb.Select(&addrs, `select address, ens_name from ens where is_primary_name = true`)
	if err != nil {
		return err
	}

	log.Infof("found %v ens entries", len(addrs))

	g := new(errgroup.Group)
	g.SetLimit(10) // limit load on the node

	batchSize := 100
	total := len(addrs)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := addrs[i:to]

		log.Infof("processing batch %v-%v / %v", i, to, total)
		for _, addr := range batch {
			addr := addr
			g.Go(func() error {
				ensAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), addr.EnsName)
				if err != nil {
					if err.Error() == "unregistered name" ||
						err.Error() == "no address" ||
						err.Error() == "no resolver" ||
						err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
						strings.Contains(err.Error(), "execution reverted") ||
						err.Error() == "invalid jump destination" {
						log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed resolve: %v", err.Error())}, "deleting ens entry")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				dbAddr := common.BytesToAddress(addr.Address)
				if dbAddr.Cmp(ensAddr) != 0 {
					log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("dbAddr != resolved ensAddr: %#x != %#x", addr.Address, ensAddr.Bytes())}, "deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				reverseName, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), dbAddr)
				if err != nil {
					if err.Error() == "not a resolver" || err.Error() == "no resolution" {
						log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed reverse-resolve: %v", err.Error())}, "updating ens entry: is_primary_name = false")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				if reverseName != addr.EnsName {
					log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("resolved != reverseResolved: %v != %v", addr.EnsName, reverseName)}, "updating ens entry: is_primary_name = false")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				return nil
			})
		}

		err = g.Wait()
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func fixEnsAddresses(erigonClient *rpc.ErigonClient) error {
	log.InfoWithFields(log.Fields{"dry": opts.DryRun}, "command: fix-ens-addresses")
	if opts.Addresses == "" {
		return errors.New("no addresses specified")
	}

	type DbEntry struct {
		NameHash      []byte    `db:"name_hash"`
		EnsName       string    `db:"ens_name"`
		Address       []byte    `db:"address"`
		IsPrimaryName bool      `db:"is_primary_name"`
		ValidTo       time.Time `db:"valid_to"`
	}

	for _, addrHex := range strings.Split(opts.Addresses, ",") {
		if !common.IsHexAddress(addrHex) {
			return fmt.Errorf("invalid address: %v", addrHex)
		}

		addr := common.HexToAddress(addrHex)

		dbEntry := &DbEntry{}
		err := db.WriterDb.Get(dbEntry, `select name_hash, ens_name, address, is_primary_name, valid_to from ens where address = $1`, addr.Bytes())
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting ens entry for addr [%v]: %w", addr.Hex(), err)
		}
		if err == sql.ErrNoRows {
			dbEntry = nil
		}

		name, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), addr)
		if err != nil {
			if err.Error() == "not a resolver" ||
				err.Error() == "no resolution" {
				log.WarnWithFields(log.Fields{"addr": addr.Hex(), "err": err}, "error reverse-resolving name")
				if dbEntry != nil {
					log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("error reverse-resolving name: %v", err)}, "deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
						if err != nil {
							return fmt.Errorf("error deleting ens entry: %w", err)
						}
					}
				}
				continue
			} else {
				return fmt.Errorf("error go_ens.ReverseResolve for addr %v: %w", addr.Hex(), err)
			}
		}

		if !strings.HasSuffix(name, ".eth") {
			log.Infof("need to add .eth to %v for addr %v", name, addr.Hex())
			name = name + ".eth"
		}

		resolvedAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
		if err != nil {
			if err.Error() == "unregistered name" ||
				err.Error() == "no address" ||
				err.Error() == "no resolver" ||
				err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
				strings.Contains(err.Error(), "execution reverted") ||
				err.Error() == "invalid jump destination" {
				if dbEntry != nil {
					log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("error resolving name: %v", err)}, "deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
						if err != nil {
							return fmt.Errorf("error deleting ens entry: %w", err)
						}
					}
				}
			} else {
				return fmt.Errorf("error go_ens.Resolve(%v) for addr %v: %w", name, addr.Hex(), err)
			}
		}

		if !bytes.Equal(resolvedAddr.Bytes(), addr.Bytes()) {
			log.WarnWithFields(log.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("addr != resolvedAddr: %v != %v", addr.Hex(), resolvedAddr.Hex())}, "deleting ens entry")
			if !opts.DryRun {
				_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
				if err != nil {
					return fmt.Errorf("error deleting ens entry: %w", err)
				}
			}
		}

		nameHash, err := go_ens.NameHash(name)
		if err != nil {
			return fmt.Errorf("error go_ens.NameHash(%v) for addr %v: %w", name, addr.Hex(), err)
		}
		parts := strings.Split(name, ".")
		mainName := strings.Join(parts[len(parts)-2:], ".")
		ensName, err := go_ens.NewName(erigonClient.GetNativeClient(), mainName)
		if err != nil {
			return fmt.Errorf("error could not create name via go_ens.NewName for [%v]: %w", name, err)
		}
		expires, err := ensName.Expires()
		if err != nil {
			return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
		}

		if dbEntry == nil || dbEntry.EnsName != name || !bytes.Equal(dbEntry.NameHash, nameHash[:]) || !bytes.Equal(dbEntry.Address, resolvedAddr.Bytes()) || dbEntry.ValidTo != expires {
			logFields := log.Fields{"resolvedAddr": resolvedAddr, "addr": addr.Hex(), "name": name, "nameHash": fmt.Sprintf("%#x", nameHash), "expires": expires}
			if dbEntry == nil {
				logFields["db"] = "nil"
				log.WarnWithFields(logFields, "adding ens entry")
			} else {
				logFields["db.name"] = dbEntry.EnsName
				logFields["db.nameHash"] = fmt.Sprintf("%#x", dbEntry.NameHash)
				logFields["db.addr"] = fmt.Sprintf("%#x", dbEntry.Address)
				logFields["db.expire"] = dbEntry.ValidTo
				log.WarnWithFields(logFields, "updating ens entry")
			}

			if !opts.DryRun {
				_, err = db.WriterDb.Exec(`
					INSERT INTO ens (
						name_hash,
						ens_name,
						address,
						is_primary_name,
						valid_to)
					VALUES ($1, $2, $3, $4, $5)
					ON CONFLICT
						(name_hash)
					DO UPDATE SET
						ens_name = excluded.ens_name,
						address = excluded.address,
						is_primary_name = excluded.is_primary_name,
						valid_to = excluded.valid_to`,
					nameHash[:], name, addr.Bytes(), true, expires)
				if err != nil {
					return fmt.Errorf("error writing ens data for addr [%v]: %w", addr.Hex(), err)
				}
			}
		}
	}
	return nil
}

func migrateAppPurchases(appStoreSecret string) error {
	// This code runs once so please don't judge code style too harshly

	if appStoreSecret == "" {
		return fmt.Errorf("appStoreSecret is empty")
	}

	client := storekit.NewVerificationClient().OnProductionEnv()

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete marked as duplicate, though the duplicate reject reason is not always set - mainly missing on historical data
	_, err = tx.Exec("DELETE FROM users_app_subscriptions WHERE store = 'ios-appstore' AND reject_reason = 'duplicate';")
	if err != nil {
		return errors.Wrap(err, "error deleting duplicate receipt")
	}

	// Backup legacy receipts into custom column
	_, err = tx.Exec("UPDATE users_app_subscriptions set legacy_receipt = receipt where legacy_receipt is null;")
	if err != nil {
		return errors.Wrap(err, "error backing up legacy receipts")
	}

	receipts := []*types.PremiumData{}
	err = tx.Select(&receipts,
		"SELECT id, receipt, store, active, validate_remotely, expires_at, product_id, user_id from users_app_subscriptions order by id desc",
	)
	if err != nil {
		return errors.Wrap(err, "error getting app subscriptions")
	}

	for _, receipt := range receipts {
		if receipt.Store != "ios-appstore" { // only interested in migrating iOS
			continue
		}
		if len(receipt.Receipt) < 100 { // dont migrate data that has already been migrated (new receipt is a number of a hand full of digits while old one is insanely large)
			continue
		}

		receiptData, err := base64.StdEncoding.DecodeString(receipt.Receipt)
		if err != nil {
			return errors.Wrap(err, "error decoding receipt")
		}

		// Call old deprecated endpoint to get the origin transaction id (new receipt info for new endpoints)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, resp, err := client.Verify(ctx, &storekit.ReceiptRequest{
			ReceiptData:            receiptData,
			Password:               appStoreSecret,
			ExcludeOldTransactions: true,
		})

		if err != nil {
			return errors.Wrap(err, "error verifying receipt")
		}

		if resp.LatestReceiptInfo == nil || len(resp.LatestReceiptInfo) == 0 {
			log.Infof("no receipt info for purchase id %v", receipt.ID)
			if receipt.Active && receipt.ValidateRemotely { // sanity, if there is an active subscription without receipt info we cam't delete it.
				return fmt.Errorf("no receipt info for active purchase id %v", receipt.ID)
			}
			// since it is not active any more and we don't get any new receipt info from apple, just drop the receipt info
			// hash can stay the same since a collision is unlikely (new and old receipt info)
			_, err = tx.Exec("UPDATE users_app_subscriptions SET receipt = '' WHERE id = $1", receipt.ID)
			if err != nil {
				return errors.Wrap(err, "error deleting duplicate receipt")
			}
			continue
		}

		latestReceiptInfo := resp.LatestReceiptInfo[0]
		log.Infof("Update purchase id %v with new receipt %v", receipt.ID, latestReceiptInfo.OriginalTransactionId)

		_, err = tx.Exec("UPDATE users_app_subscriptions SET receipt = $1, receipt_hash = $2 WHERE id = $3", latestReceiptInfo.OriginalTransactionId, utils.HashAndEncode(latestReceiptInfo.OriginalTransactionId), receipt.ID)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") { // handle historic duplicates
				// get the duplicate receipt
				duplicateReceipt := types.PremiumData{}
				err = tx.Get(&duplicateReceipt, "SELECT id, user_id, active FROM users_app_subscriptions WHERE receipt_hash = $1", utils.HashAndEncode(latestReceiptInfo.OriginalTransactionId))
				if err != nil {
					return errors.Wrap(err, "error getting duplicate receipt")
				}

				// Keep the active receipt and delete the other one. In case both are inactive keep the newest
				var deleteReceiptID uint64
				if !duplicateReceipt.Active && receipt.Active {
					deleteReceiptID = duplicateReceipt.ID
				} else if duplicateReceipt.Active && !receipt.Active {
					deleteReceiptID = receipt.ID
				} else if !duplicateReceipt.Active && !receipt.Active {
					if duplicateReceipt.ID > receipt.ID { // keep the newer one
						deleteReceiptID = duplicateReceipt.ID
					} else {
						deleteReceiptID = receipt.ID
					}
				} else {
					return fmt.Errorf("duplicate receipt has same active status: %v != %v for id: %v != %v", duplicateReceipt.Active, receipt.Active, duplicateReceipt.ID, receipt.ID)
				}

				// new ios handler will automatically update the product id if the user switched the package, so we will just drop this receipt
				_, err = tx.Exec("DELETE FROM users_app_subscriptions WHERE id = $1", deleteReceiptID)
				if err != nil {
					return errors.Wrap(err, "error deleting duplicate receipt")
				}
				log.Infof("deleted duplicate receipt id %v", receipt.ID)

				// the one we keep and update is opposite of the one we deleted
				var updateReceiptID uint64
				if deleteReceiptID == duplicateReceipt.ID {
					updateReceiptID = receipt.ID
				} else {
					updateReceiptID = duplicateReceipt.ID
				}

				_, err = tx.Exec("UPDATE users_app_subscriptions SET receipt = $1, receipt_hash = $2 WHERE id = $3", latestReceiptInfo.OriginalTransactionId, utils.HashAndEncode(latestReceiptInfo.OriginalTransactionId), updateReceiptID)
				if err != nil {
					return errors.Wrap(err, "error updating receipt")
				}
			} else {
				return errors.Wrap(err, "error updating purchase id")
			}
		}

		log.Infof("Migrated purchase id %v\n", receipt.ID)
		time.Sleep(100 * time.Millisecond)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "error committing tx")
	}

	log.Infof("done migrating data")
	return nil
}

func fixExecTransactionsCount() error {
	startBlockNumber := opts.StartBlock
	endBlockNumber := opts.EndBlock

	log.InfoWithFields(log.Fields{"startBlockNumber": startBlockNumber, "endBlockNumber": endBlockNumber}, "fixExecTransactionsCount")

	batchSize := int64(1000)

	dbUpdates := []struct {
		BlockNumber  uint64
		ExecTxsCount uint64
	}{}

	for i := startBlockNumber; i <= endBlockNumber; i += uint64(batchSize) {
		firstBlock := int64(i)
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > int64(endBlockNumber) {
			lastBlock = int64(endBlockNumber)
		}
		blocksChan := make(chan *types.Eth1Block, batchSize)
		go func(stream chan *types.Eth1Block) {
			high := lastBlock
			low := lastBlock - batchSize + 1
			if firstBlock > low {
				low = firstBlock
			}

			err := db.BigtableClient.GetFullBlocksDescending(stream, uint64(high), uint64(low))
			if err != nil {
				log.Error(err, "error getting blocks descending high: %v low: %v err: %v", 0, map[string]interface{}{"high": high, "low": low})
			}
			close(stream)
		}(blocksChan)
		totalTxsCount := 0
		for b := range blocksChan {
			if len(b.Transactions) > 0 {
				totalTxsCount += len(b.Transactions)
				dbUpdates = append(dbUpdates, struct {
					BlockNumber  uint64
					ExecTxsCount uint64
				}{b.Number, uint64(len(b.Transactions))})
			}
		}
		log.Infof("%v-%v: totalTxsCount: %v", firstBlock, lastBlock, totalTxsCount)
	}

	log.Infof("dbUpdates: %v", len(dbUpdates))

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer utils.Rollback(tx)

	for b := 0; b < len(dbUpdates); b += int(batchSize) {
		start := b
		end := b + int(batchSize)
		if len(dbUpdates) < end {
			end = len(dbUpdates)
		}

		valueStrings := []string{}
		for _, v := range dbUpdates[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("(%v,%v)", v.BlockNumber, v.ExecTxsCount))
		}

		stmt := fmt.Sprintf(`
			update blocks as a set exec_transactions_count = b.exec_transactions_count
			from (values %s) as b(exec_block_number, exec_transactions_count)
			where a.exec_block_number = b.exec_block_number`, strings.Join(valueStrings, ","))

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

		log.Infof("updated %v-%v / %v", start, end, len(dbUpdates))
	}

	return tx.Commit()
}

func updateBlockFinalizationSequentially() error {
	var err error

	var maxSlot uint64
	err = db.WriterDb.Get(&maxSlot, `select max(slot) from blocks`)
	if err != nil {
		return err
	}

	lookback := uint64(0)
	if maxSlot > 1e4 {
		lookback = maxSlot - 1e4
	}
	var minNonFinalizedSlot uint64
	for {
		err = db.WriterDb.Get(&minNonFinalizedSlot, `select coalesce(min(slot),0) from blocks where finalized = false and slot >= $1 and slot <= $1+1e4`, lookback)
		if err != nil {
			return err
		}
		if minNonFinalizedSlot == 0 {
			break
		}
		if minNonFinalizedSlot == lookback && lookback > 1e4 {
			lookback -= 1e4
			continue
		}
		break
	}

	log.InfoWithFields(log.Fields{"minNonFinalizedSlot": minNonFinalizedSlot}, "updateBlockFinalizationSequentially")
	nextStartEpoch := minNonFinalizedSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	stepSize := uint64(100)
	for ; ; time.Sleep(time.Millisecond * 50) {
		t0 := time.Now()
		var finalizedEpoch uint64
		err = db.WriterDb.Get(&finalizedEpoch, `SELECT COALESCE(MAX(epoch) - 3, 0) FROM epochs WHERE finalized`)
		if err != nil {
			return err
		}
		lastEpoch := nextStartEpoch + stepSize - 1
		if lastEpoch > finalizedEpoch {
			lastEpoch = finalizedEpoch
		}
		_, err = db.WriterDb.Exec(`UPDATE blocks SET finalized = true WHERE epoch >= $1 AND epoch <= $2 AND NOT finalized;`, nextStartEpoch, lastEpoch)
		if err != nil {
			return err
		}
		err = cache.LatestFinalizedEpoch.Set(lastEpoch)
		if err != nil {
			return err
		}

		secondsPerEpoch := time.Since(t0).Seconds() / float64(stepSize)
		timeLeft := time.Second * time.Duration(float64(finalizedEpoch-lastEpoch)*time.Since(t0).Seconds()/float64(stepSize))
		log.InfoWithFields(log.Fields{"finalizedEpoch": finalizedEpoch, "epochs": fmt.Sprintf("%v-%v", nextStartEpoch, lastEpoch), "timeLeft": timeLeft, "secondsPerEpoch": secondsPerEpoch}, "did set blocks to finalized")
		if finalizedEpoch <= lastEpoch {
			log.Infof("all relevant blocks have been set to finalized (up to epoch %v)", finalizedEpoch)
			return nil
		}
		nextStartEpoch = nextStartEpoch + stepSize
	}
}

func debugBlocks(clClient *rpc.LighthouseClient) error {
	elClient, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		return err
	}

	for i := opts.StartBlock; i <= opts.EndBlock; i++ {
		btBlock, err := db.BigtableClient.GetBlockFromBlocksTable(i)
		if err != nil {
			return err
		}

		elBlock, _, err := elClient.GetBlock(int64(i), "parity/geth")
		if err != nil {
			return err
		}

		slot := utils.TimeToSlot(uint64(elBlock.Time.Seconds))
		clBlock, err := clClient.GetBlockBySlot(slot)
		if err != nil {
			return err
		}
		logFields := log.Fields{
			"block":            i,
			"bt.hash":          fmt.Sprintf("%#x", btBlock.Hash),
			"bt.BlobGasUsed":   btBlock.BlobGasUsed,
			"bt.ExcessBlobGas": btBlock.ExcessBlobGas,
			"bt.txs":           len(btBlock.Transactions),
			"el.BlobGasUsed":   elBlock.BlobGasUsed,
			"el.hash":          fmt.Sprintf("%#x", elBlock.Hash),
			"el.ExcessBlobGas": elBlock.ExcessBlobGas,
			"el.txs":           len(elBlock.Transactions),
		}
		if !bytes.Equal(clBlock.ExecutionPayload.BlockHash, elBlock.Hash) {
			log.Warnf("clBlock.ExecutionPayload.BlockHash != i: %x != %x", clBlock.ExecutionPayload.BlockHash, elBlock.Hash)
		} else if clBlock.ExecutionPayload.BlockNumber != i {
			log.Warnf("clBlock.ExecutionPayload.BlockNumber != i: %v != %v", clBlock.ExecutionPayload.BlockNumber, i)
		} else {
			logFields["cl.txs"] = len(clBlock.ExecutionPayload.Transactions)
		}

		log.InfoWithFields(logFields, "debug block")

		for i := range elBlock.Transactions {
			btx := elBlock.Transactions[i]
			ctx := elBlock.Transactions[i]
			btxH := []string{}
			ctxH := []string{}
			for _, h := range btx.BlobVersionedHashes {
				btxH = append(btxH, fmt.Sprintf("%#x", h))
			}
			for _, h := range ctx.BlobVersionedHashes {
				ctxH = append(ctxH, fmt.Sprintf("%#x", h))
			}

			log.InfoWithFields(log.Fields{
				"b.hash":                 fmt.Sprintf("%#x", btx.Hash),
				"el.hash":                fmt.Sprintf("%#x", ctx.Hash),
				"b.BlobVersionedHashes":  fmt.Sprintf("%+v", btxH),
				"el.BlobVersionedHashes": fmt.Sprintf("%+v", ctxH),
				"b.maxFeePerBlobGas":     btx.MaxFeePerBlobGas,
				"el.maxFeePerBlobGas":    ctx.MaxFeePerBlobGas,
				"b.BlobGasPrice":         btx.BlobGasPrice,
				"el.BlobGasPrice":        ctx.BlobGasPrice,
				"b.BlobGasUsed":          btx.BlobGasUsed,
				"el.BlobGasUsed":         ctx.BlobGasUsed,
			}, "debug tx")
		}
	}
	return nil
}

func nameValidatorsByRanges(rangesUrl string) error {
	ranges := struct {
		Ranges map[string]string `json:"ranges"`
	}{}

	if strings.HasPrefix(rangesUrl, "http") {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := utils.HttpReq(ctx, http.MethodGet, rangesUrl, nil, &ranges)
		if err != nil {
			return err
		}
	} else {
		err := json.Unmarshal([]byte(rangesUrl), &ranges)
		if err != nil {
			return err
		}
	}

	for r, n := range ranges.Ranges {
		rs := strings.Split(r, "-")
		if len(rs) != 2 {
			return fmt.Errorf("invalid format, range must be X-Y")
		}
		rFrom, err := strconv.ParseUint(rs[0], 10, 64)
		if err != nil {
			return err
		}
		rTo, err := strconv.ParseUint(rs[1], 10, 64)
		if err != nil {
			return err
		}
		if rTo < rFrom {
			return fmt.Errorf("invalid format, range must be X-Y where X <= Y")
		}

		_, err = db.WriterDb.Exec("insert into validator_names(publickey, name) select pubkey as publickey, $1 as name from validators where validatorindex >= $2 and validatorindex <= $3 on conflict(publickey) do update set name = excluded.name;", n, rFrom, rTo)
		if err != nil {
			return err
		}
	}

	return nil
}

// one time migration of the last attestation slot values from postgres to bigtable
// will write the last attestation slot that is currently in postgres to bigtable
// this can safely be done for active validators as bigtable will only keep the most recent
// last attestation slot
func migrateLastAttestationSlotToBigtable() {
	validators := []types.Validator{}

	err := db.WriterDb.Select(&validators, "SELECT validatorindex, lastattestationslot FROM validators WHERE lastattestationslot IS NOT NULL ORDER BY validatorindex")

	if err != nil {
		log.Fatal(err, "error retrieving last attestation slot", 0)
	}

	for _, validator := range validators {
		log.Infof("setting last attestation slot %v for validator %v", validator.LastAttestationSlot, validator.Index)

		err := db.BigtableClient.SetLastAttestationSlot(validator.Index, uint64(validator.LastAttestationSlot.Int64))
		if err != nil {
			log.Fatal(err, "error setting last attestation slot", 0)
		}
	}
}

func updateAggreationBits(rpcClient *rpc.LighthouseClient, startEpoch uint64, endEpoch uint64, concurency uint64) {
	log.Infof("update-aggregation-bits epochs %v - %v", startEpoch, endEpoch)
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		log.Infof("Getting data from the node for epoch %v", epoch)
		data, err := rpcClient.GetEpochData(epoch, false)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error getting epoch[%v] data from the client", epoch), 0)
			return
		}

		ctx := context.Background()
		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(int(concurency))

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			log.Fatal(err, "error starting tx", 0)
		}
		defer func() {
			err := tx.Rollback()
			if err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		for _, bm := range data.Blocks {
			for _, b := range bm {
				block := b
				log.Infof("Updating data for slot %v", block.Slot)

				if len(block.Attestations) == 0 {
					log.Infof("No Attestations for slot %v", block.Slot)

					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return gCtx.Err()
						default:
						}

						// if we have some obsolete attestations we clean them from the db
						rows, err := tx.Exec(`
								DELETE FROM blocks_attestations
								WHERE
									block_slot=$1
							`, block.Slot)
						if err != nil {
							return fmt.Errorf("error deleting obsolete attestations for Slot [%v]:  %v", block.Slot, err)
						}
						if rowsAffected, _ := rows.RowsAffected(); rowsAffected > 0 {
							log.Infof("%v obsolete attestations removed for Slot[%v]", rowsAffected, block.Slot)
						} else {
							log.Infof("No obsolete attestations found for Slot[%v] so we move on", block.Slot)
						}

						return nil
					})
					continue
				}

				status := uint64(0)
				err := tx.Get(&status, `
				SELECT status
				FROM blocks WHERE
					slot=$1`, block.Slot)
				if err != nil {
					log.Error(err, fmt.Errorf("error getting Slot [%v] status", block.Slot), 0)
					return
				}
				importWholeBlock := false

				if status != block.Status {
					log.Infof("Slot[%v] has the wrong status [%v], but should be [%v]", block.Slot, status, block.Status)
					if block.Status == 1 {
						importWholeBlock = true
					} else {
						log.Error(err, fmt.Errorf("error on Slot [%v] - no update process for status [%v]", block.Slot, block.Status), 0)
						return
					}
				} else if len(block.Attestations) > 0 {
					count := 0
					err := tx.Get(&count, `
						SELECT COUNT(*)
						FROM
							blocks_attestations
						WHERE
							block_slot=$1`, block.Slot)
					if err != nil {
						log.Error(err, fmt.Errorf("error getting Slot [%v] status", block.Slot), 0)
						return
					}
					// We only know about cases where we have no attestations in the db but the node has one.
					// So we don't handle cases (for now) where there are attestations with different sizes - that would require a different handling
					if count == 0 {
						importWholeBlock = true
					}
				}

				if importWholeBlock {
					err := edb.SaveBlock(block, true, tx)
					if err != nil {
						log.Error(err, fmt.Errorf("error saving Slot [%v]", block.Slot), 0)
						return
					}
					continue
				}

				for i, a := range block.Attestations {
					att := a
					index := i
					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return gCtx.Err()
						default:
						}
						var aggregationbits *[]byte

						// block_slot and block_index are already unique, but to be sure we use the correct index we also check the signature
						err := tx.Get(&aggregationbits, `
							SELECT aggregationbits
							FROM blocks_attestations WHERE
								block_slot=$1 AND
								block_index=$2
						`, block.Slot, index)
						if err != nil {
							return fmt.Errorf("error getting aggregationbits on Slot [%v] Index [%v] with Sig [%v]: %v", block.Slot, index, att.Signature, err)
						}

						if !bytes.Equal(*aggregationbits, att.AggregationBits) {
							_, err = tx.Exec(`
								UPDATE blocks_attestations
								SET
									aggregationbits=$1
								WHERE
									block_slot=$2 AND
									block_index=$3
							`, att.AggregationBits, block.Slot, index)
							if err != nil {
								return fmt.Errorf("error updating aggregationbits on Slot [%v] Index [%v] :  %v", block.Slot, index, err)
							}
							log.Infof("Update of Slot[%v] Index[%v] complete", block.Slot, index)
						} else {
							log.Infof("Slot[%v] Index[%v] was already up to date", block.Slot, index)
						}

						return nil
					})
				}
			}
		}

		err = g.Wait()

		if err != nil {
			log.Error(err, fmt.Sprintf("error updating aggregationbits for epoch [%v]", epoch), 0)
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Error(err, fmt.Sprintf("error committing tx for epoch [%v]", epoch), 0)
			return
		}
		log.Infof("Update of Epoch[%v] complete", epoch)
	}
}

// Updates a users API key
func updateAPIKey(user uint64) error {
	type User struct {
		PHash  string `db:"password"`
		Email  string `db:"email"`
		OldKey string `db:"api_key"`
	}

	var u User
	err := db.FrontendWriterDB.Get(&u, `SELECT password, email, api_key from users where id = $1`, user)
	if err != nil {
		return fmt.Errorf("error getting current user, err: %w", err)
	}

	apiKey, err := utils.GenerateRandomAPIKey()
	if err != nil {
		return err
	}

	log.Infof("updating api key for user %v from old key: %v to new key: %v", user, u.OldKey, apiKey)

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`UPDATE api_statistics set apikey = $1 where apikey = $2`, apiKey, u.OldKey)
	if err != nil {
		return err
	}

	rows, err := tx.Exec(`UPDATE users SET api_key = $1 WHERE id = $2`, apiKey, user)
	if err != nil {
		return err
	}

	amount, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if amount > 1 {
		return fmt.Errorf("error too many rows affected expected 1 but got: %v", amount)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Debugging function to compare Rewards from the Statistic Table with the onces from the Big Table
func compareRewards(dayStart uint64, dayEnd uint64, validator uint64, bt *db.Bigtable) {
	for day := dayStart; day <= dayEnd; day++ {
		startEpoch := day * utils.EpochsPerDay()
		endEpoch := startEpoch + utils.EpochsPerDay() - 1
		hist, err := bt.GetValidatorIncomeDetailsHistory([]uint64{validator}, startEpoch, endEpoch)
		if err != nil {
			log.Fatal(err, "error retrieving validator income details history", 0, map[string]interface{}{"startEpoch": startEpoch, "endEpoch": endEpoch})
		}
		var tot int64
		for _, rew := range hist[validator] {
			tot += rew.TotalClRewards()
		}
		log.Infof("Total CL Rewards for day [%v]: %v", day, tot)
		var dbRewards *int64
		err = db.ReaderDb.Get(&dbRewards, `
		SELECT
		COALESCE(cl_rewards_gwei, 0) AS cl_rewards_gwei
		FROM validator_stats WHERE validatorindex = $2 AND day = $1`, day, validator)
		if err != nil {
			log.Fatal(err, "error getting cl_rewards_gwei from db", 0)
			return
		}
		if tot != *dbRewards {
			log.Error(fmt.Errorf("rewards are not the same on day %v-> big: %v, db: %v", day, tot, *dbRewards), "", 0)
		}
	}
}

func clearBigtable(table string, family string, columns string, key string, dryRun bool, bt *db.Bigtable) {
	if !dryRun {
		confirmation := utils.CmdPrompt(fmt.Sprintf("Are you sure you want to delete all big table entries starting with [%v] for family [%v] and columns [%v]?", key, family, columns))
		if confirmation != "yes" {
			log.Infof("Abort!")
			return
		}
	}

	if !strings.Contains(key, ":") {
		log.Fatal(fmt.Errorf("provided invalid prefix: %s", key), "", 0)
	}

	err := bt.ClearByPrefix(table, family, columns, key, dryRun)

	if err != nil {
		log.Fatal(err, "error deleting from bigtable", 0)
	}
	log.Infof("delete completed")
}

// Goes through the tableData table and checks what blocks in the given range from [start] to [end] are missing and exports/indexes the missing ones
//
//	Both [start] and [end] are inclusive
//	Pass math.MaxInt64 as [end] to export from [start] to the last block in the blocks table
func indexMissingBlocks(start uint64, end uint64, bt *db.Bigtable, client *rpc.ErigonClient) {
	if end == math.MaxInt64 {
		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			log.Error(err, "error retrieving last blocks from blocks table", 0)
			return
		}
		end = uint64(lastBlockFromBlocksTable)
	}

	errFields := map[string]interface{}{
		"start": start,
		"end":   end}

	batchSize := uint64(10000)
	for from := start; from <= end; from += batchSize {
		targetCount := batchSize
		if from+targetCount >= end {
			targetCount = end - from + 1
		}
		to := from + targetCount - 1

		errFields["from"] = from
		errFields["to"] = to
		errFields["targetCount"] = targetCount

		list, err := bt.GetBlocksDescending(to, targetCount)
		if err != nil {
			log.Error(err, "error retrieving blocks from tableData", 0, errFields)
			return
		}

		receivedLen := uint64(len(list))
		if receivedLen == targetCount {
			log.Infof("found all blocks [%v]->[%v], skipping batch", from, to)
			continue
		}

		log.Infof("%v blocks are missing from [%v]->[%v]", targetCount-receivedLen, from, to)

		blocksMap := make(map[uint64]bool)
		for _, item := range list {
			blocksMap[item.Number] = true
		}

		for block := from; block <= to; block++ {
			if blocksMap[block] {
				// block already saved, skip
				continue
			}

			log.Infof("block [%v] not found, will index it", block)
			if _, err := db.BigtableClient.GetBlockFromBlocksTable(block); err != nil {
				log.Infof("could not load [%v] from blocks table, will try to fetch it from the node and save it", block)

				bc, _, err := client.GetBlock(int64(block), "parity/geth")
				if err != nil {
					log.Error(err, fmt.Sprintf("error getting block %v from the node", block), 0)
					return
				}

				err = bt.SaveBlock(bc)
				if err != nil {
					log.Error(err, fmt.Sprintf("error saving block: %v ", block), 0)
					return
				}
			}

			indexOldEth1Blocks(block, block, 1, 1, "all", bt, client)
		}
	}
}

func indexOldEth1Blocks(startBlock uint64, endBlock uint64, batchSize uint64, concurrency uint64, transformerFlag string, bt *db.Bigtable, client *rpc.ErigonClient) {
	if endBlock > 0 && endBlock < startBlock {
		log.Error(nil, fmt.Sprintf("endBlock [%v] < startBlock [%v]", endBlock, startBlock), 0)
		return
	}
	if concurrency == 0 {
		log.Error(nil, "concurrency must be greater than 0", 0)
		return
	}
	if bt == nil {
		log.Error(nil, "no bigtable provided", 0)
		return
	}

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)

	log.Infof("transformerFlag: %v", transformerFlag)
	transformerList := strings.Split(transformerFlag, ",")
	if transformerFlag == "all" {
		transformerList = []string{"TransformBlock", "TransformTx", "TransformBlobTx", "TransformItx", "TransformERC20", "TransformERC721", "TransformERC1155", "TransformWithdrawals", "TransformUncle", "TransformEnsNameRegistered", "TransformContract"}
	} else if len(transformerList) == 0 {
		log.Error(nil, "no transformer functions provided", 0)
		return
	}
	log.Infof("transformers: %v", transformerList)
	importENSChanges := false
	/**
	* Add additional transformers you want to sync to this switch case
	**/
	for _, t := range transformerList {
		switch t {
		case "TransformBlock":
			transforms = append(transforms, bt.TransformBlock)
		case "TransformTx":
			transforms = append(transforms, bt.TransformTx)
		case "TransformBlobTx":
			transforms = append(transforms, bt.TransformBlobTx)
		case "TransformItx":
			transforms = append(transforms, bt.TransformItx)
		case "TransformERC20":
			transforms = append(transforms, bt.TransformERC20)
		case "TransformERC721":
			transforms = append(transforms, bt.TransformERC721)
		case "TransformERC1155":
			transforms = append(transforms, bt.TransformERC1155)
		case "TransformWithdrawals":
			transforms = append(transforms, bt.TransformWithdrawals)
		case "TransformUncle":
			transforms = append(transforms, bt.TransformUncle)
		case "TransformEnsNameRegistered":
			transforms = append(transforms, bt.TransformEnsNameRegistered)
			importENSChanges = true
		case "TransformContract":
			transforms = append(transforms, bt.TransformContract)
		default:
			log.Error(nil, "Invalid transformer flag %v", 0)
			return
		}
	}

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit

	to := endBlock
	if endBlock == math.MaxInt64 {
		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			log.Error(err, "error retrieving last blocks from blocks table", 0)
			return
		}

		to = uint64(lastBlockFromBlocksTable)
	}
	blockCount := utilMath.MaxU64(1, batchSize)

	log.Infof("Starting to index all blocks ranging from %d to %d", startBlock, to)
	for from := startBlock; from <= to; from = from + blockCount {
		toBlock := utilMath.MinU64(to, from+blockCount-1)

		log.Infof("indexing blocks %v to %v in data table ...", from, toBlock)
		err := bt.IndexEventsWithTransformers(int64(from), int64(toBlock), transforms, int64(concurrency), cache)
		if err != nil {
			log.Error(err, "error indexing from bigtable", 0)
		}
		cache.Clear()
	}

	if importENSChanges {
		if err := bt.ImportEnsUpdates(client.GetNativeClient(), math.MaxInt64); err != nil {
			log.Error(err, "error importing ens from events", 0)
			return
		}
	}

	log.Infof("index run completed")
}

func exportHistoricPrices(dayStart uint64, dayEnd uint64) {
	log.Infof("exporting historic prices for days %v - %v", dayStart, dayEnd)
	for day := dayStart; day <= dayEnd; day++ {
		timeStart := time.Now()
		ts := utils.DayToTime(int64(day)).UTC().Truncate(utils.Day)
		err := services.WriteHistoricPricesForDay(ts)
		if err != nil {
			errMsg := fmt.Sprintf("error exporting historic prices for day %v", day)
			log.Error(err, errMsg, 0)
			return
		}
		log.Infof("finished export for day %v, took %v", day, time.Since(timeStart))

		if day < dayEnd {
			// Wait to not overload the API
			time.Sleep(5 * time.Second)
		}
	}

	log.Infof("historic price update run completed")
}

func exportStatsTotals(columns string, dayStart, dayEnd, concurrency uint64) {
	start := time.Now()
	exportToToday := false
	if dayEnd <= 0 {
		exportToToday = true
		dayEnd = math.MaxInt
	}

	log.Infof("exporting stats totals for columns '%v'", columns)

	// validate columns input
	columnsSlice := strings.Split(columns, ",")
	validColumns := []string{
		"cl_rewards_gwei_total",
		"el_rewards_wei_total",
		"mev_rewards_wei_total",
		"missed_attestations_total",
		"participated_sync_total",
		"missed_sync_total",
		"orphaned_sync_total",
		"withdrawals_total",
		"withdrawals_amount_total",
		"deposits_total",
		"deposits_amount_total",
	}

OUTER:
	for _, c := range columnsSlice {
		for _, vc := range validColumns {
			if c == vc {
				// valid column found, continue to next column from input
				continue OUTER
			}
		}
		// no valid column matched, exit with error
		log.Fatal(nil, "invalid column provided, please use a valid one", 0, map[string]interface{}{
			"usedColumn":   c,
			"validColumns": validColumns,
		})
	}

	// build insert query from input columns
	var totalClauses []string
	var conflictClauses []string

	for _, col := range columnsSlice {
		totalClause := fmt.Sprintf("COALESCE(vs1.%s, 0) + COALESCE(vs2.%s, 0)", strings.TrimSuffix(col, "_total"), col)
		totalClauses = append(totalClauses, totalClause)

		conflictClause := fmt.Sprintf("%s = excluded.%s", col, col)
		conflictClauses = append(conflictClauses, conflictClause)
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO validator_stats (validatorindex, day, %s)
		SELECT
			vs1.validatorindex,
			vs1.day,
			%s
		FROM validator_stats vs1
		LEFT JOIN validator_stats vs2
		ON vs2.day = vs1.day - 1 AND vs2.validatorindex = vs1.validatorindex
		WHERE vs1.day = $1 AND vs1.validatorindex >= $2 AND vs1.validatorindex <= $3
		ON CONFLICT (validatorindex, day) DO UPDATE SET %s;`,
		strings.Join(columnsSlice, ",\n\t"),
		strings.Join(totalClauses, ",\n\t\t"),
		strings.Join(conflictClauses, ",\n\t"))

	for day := dayStart; day <= dayEnd; day++ {
		timeDay := time.Now()
		log.Infof("exporting total sync and for columns %v for day %v", columns, day)

		// get max validator index for day
		firstEpoch, _ := utils.GetFirstAndLastEpochForDay(day + 1)
		var maxValidatorIndex uint64
		err := db.ReaderDb.Get(&maxValidatorIndex, `SELECT MAX(validatorindex) FROM validator_stats WHERE day = $1`, day)
		if err != nil {
			log.Fatal(err, "error: could not get max validator index", 0, map[string]interface{}{
				"epoch": firstEpoch,
			})
		} else if maxValidatorIndex == uint64(0) {
			log.Fatal(err, "error: no validator found", 0, map[string]interface{}{
				"epoch": firstEpoch,
			})
		}

		ctx := context.Background()
		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(int(concurrency))

		batchSize := 1000

		// insert stats totals for each batch of validators
		for b := 0; b <= int(maxValidatorIndex); b += batchSize {
			start := b
			end := b + batchSize - 1
			if int(maxValidatorIndex) < end {
				end = int(maxValidatorIndex)
			}

			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				_, err = db.WriterDb.Exec(insertQuery, day, start, end)
				return err
			})
		}
		if err = g.Wait(); err != nil {
			log.Fatal(err, "error exporting stats totals", 0, map[string]interface{}{
				"day":     day,
				"columns": columns,
			})
		}

		if exportToToday {
			dayEnd, err = db.GetLastExportedStatisticDay()
			if err != nil {
				log.Error(err, "error getting last exported statistic day", 0)
				return
			}
		}
		log.Infof("finished exporting stats totals for columns '%v for day %v, took %v", columns, day, time.Since(timeDay))
	}

	log.Infof("finished all exporting stats totals for columns '%v' for days %v - %v, took %v", columns, dayStart, dayEnd, time.Since(start))
}

/*
Instead of deleting entries from the sync_committee table in a prod environment and wait for the exporter to sync back all entries,
this method will replace each sync committee period one by one with the new one. Which is much nicer for a prod environment.
*/
func exportSyncCommitteePeriods(rpcClient rpc.Client, startDay, endDay uint64, dryRun bool) {
	var lastEpoch uint64

	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)
	if startDay > 0 {
		firstEpoch, _ := utils.GetFirstAndLastEpochForDay(startDay)
		firstPeriod = utils.SyncPeriodOfEpoch(firstEpoch)
	}

	if endDay <= 0 {
		var err error
		lastEpoch, err = db.GetLatestFinalizedEpoch()
		if err != nil {
			log.Error(err, "error getting latest finalized epoch", 0)
			return
		}
		if lastEpoch > 0 { // guard against underflows
			lastEpoch = lastEpoch - 1
		}
	} else {
		_, lastEpoch = utils.GetFirstAndLastEpochForDay(endDay)
	}

	lastPeriod := utils.SyncPeriodOfEpoch(lastEpoch) + 1 // we can look into the future

	start := time.Now()
	for p := firstPeriod; p <= lastPeriod; p++ {
		t0 := time.Now()

		err := reExportSyncCommittee(rpcClient, p, dryRun)
		if err != nil {
			if strings.Contains(err.Error(), "not found 404") {
				log.InfoWithFields(log.Fields{"period": p}, "reached max period, stopping")
				break
			} else {
				log.Error(err, "error re-exporting sync_committee", 0, map[string]interface{}{
					"period": p,
				})
				return
			}
		}

		log.InfoWithFields(log.Fields{
			"period":   p,
			"epoch":    utils.FirstEpochOfSyncPeriod(p),
			"duration": time.Since(t0),
		}, "re-exported sync_committee")
	}

	log.Infof("finished all exporting sync_committee for periods %v - %v, took %v", firstPeriod, lastPeriod, time.Since(start))
}

//nolint:unparam
func exportSyncCommitteeValidatorStats(rpcClient rpc.Client, startDay, endDay uint64, dryRun, skipPhase1 bool) {
	if endDay <= 0 {
		lastEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			log.Error(err, "error getting latest finalized epoch", 0)
			return
		}
		if lastEpoch > 0 { // guard against underflows
			lastEpoch = lastEpoch - 1
		}

		_, err = db.GetLastExportedStatisticDay()
		if err != nil {
			log.Infof("skipping exporting stats, first day has not been indexed yet")
			return
		}

		epochsPerDay := utils.EpochsPerDay()
		currentDay := lastEpoch / epochsPerDay
		endDay = currentDay - 1 // current day will be picked up by exporter
	}

	start := time.Now()

	for day := startDay; day <= endDay; day++ {
		startDay := time.Now()
		err := UpdateValidatorStatisticsSyncData(day, dryRun)
		if err != nil {
			log.Error(err, fmt.Errorf("error exporting stats for day %v", day), 0)
			break
		}

		log.Infof("finished updating validators_stats for day %v, took %v", day, time.Since(startDay))
	}

	log.Infof("finished all exporting stats for days %v - %v, took %v", startDay, endDay, time.Since(start))
	log.Infof("REMEMBER: To execute export-stats-totals now to update the totals")
}

func UpdateValidatorStatisticsSyncData(day uint64, dryRun bool) error {
	exportStart := time.Now()
	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	log.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	if err := db.CheckIfDayIsFinalized(day); err != nil && !dryRun {
		return err
	}

	log.Infof("getting exported state for day %v", day)

	var err error
	var maxValidatorIndex uint64
	err = db.ReaderDb.Get(&maxValidatorIndex, `SELECT MAX(validatorindex) FROM validator_stats WHERE day = $1`, day)
	if err != nil {
		log.Fatal(err, "error: could not get max validator index", 0, map[string]interface{}{
			"epoch": firstEpoch,
		})
	} else if maxValidatorIndex == uint64(0) {
		log.Fatal(err, "error: no validator found", 0, map[string]interface{}{
			"epoch": firstEpoch,
		})
	}
	maxValidatorIndex += 10000 // add some buffer, exact number is not important. Should just be bigger than max validators that can join in a day

	validatorData := make([]*types.ValidatorStatsTableDbRow, 0, maxValidatorIndex)
	validatorDataMux := &sync.Mutex{}

	log.Infof("processing statistics for validators 0-%d", maxValidatorIndex)
	for i := uint64(0); i <= maxValidatorIndex; i++ {
		validatorData = append(validatorData, &types.ValidatorStatsTableDbRow{
			ValidatorIndex: i,
			Day:            int64(day),
		})
	}

	g := &errgroup.Group{}

	g.Go(func() error {
		if err := db.GatherValidatorSyncDutiesForDay(nil, day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorSyncDutiesForDay: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	onlySyncCommitteeValidatorData := make([]*types.ValidatorStatsTableDbRow, 0, len(validatorData))
	for index := range validatorData {
		if validatorData[index].ParticipatedSync > 0 || validatorData[index].MissedSync > 0 || validatorData[index].OrphanedSync > 0 {
			onlySyncCommitteeValidatorData = append(onlySyncCommitteeValidatorData, validatorData[index])
		}
	}

	if len(onlySyncCommitteeValidatorData) == 0 {
		return nil // no sync committee yet skip
	}

	log.Infof("statistics data collection for day %v completed", day)

	var statisticsDataToday []*types.ValidatorStatsTableDbRow
	if dryRun {
		var err error
		statisticsDataToday, err = db.GatherStatisticsForDay(int64(day)) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer utils.Rollback(tx)

	log.Infof("updating statistics data into the validator_stats table %v | %v", len(onlySyncCommitteeValidatorData), len(validatorData))

	for _, data := range onlySyncCommitteeValidatorData {
		if dryRun {
			log.Infof(
				"validator %v: participated sync: %v -> %v, missed sync: %v -> %v, orphaned sync: %v -> %v",
				data.ValidatorIndex, statisticsDataToday[data.ValidatorIndex].ParticipatedSync, data.ParticipatedSync, statisticsDataToday[data.ValidatorIndex].MissedSync, data.MissedSync, statisticsDataToday[data.ValidatorIndex].OrphanedSync,
				data.OrphanedSync,
			)
		} else {
			_, err := tx.Exec(`
				UPDATE validator_stats set
				participated_sync = $1,
				missed_sync = $2,
				orphaned_sync = $3
				WHERE day = $4 AND validatorindex = $5`,
				data.ParticipatedSync,
				data.MissedSync,
				data.OrphanedSync,
				data.Day, data.ValidatorIndex)
			if err != nil {
				log.Error(err, "error updating validator stats", 0)
			}
		}
	}

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error during statistics data insert: %w", err)
	}

	log.Infof("statistics sync re-export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func reExportSyncCommittee(rpcClient rpc.Client, p uint64, dryRun bool) error {
	if dryRun {
		var currentData []struct {
			ValidatorIndex uint64 `db:"validatorindex"`
			CommitteeIndex uint64 `db:"committeeindex"`
		}

		err := db.WriterDb.Select(&currentData, `SELECT validatorindex, committeeindex FROM sync_committees WHERE period = $1`, p)
		if err != nil {
			return errors.Wrap(err, "select old entries")
		}

		newData, err := modules.GetSyncCommitteAtPeriod(rpcClient, p)
		if err != nil {
			return errors.Wrap(err, "export")
		}

		// now we compare currentData with newData and print any difference in committeeindex
		for _, d := range currentData {
			for _, n := range newData {
				if d.ValidatorIndex == n.ValidatorIndex && d.CommitteeIndex != n.CommitteeIndex {
					log.Infof("validator %v has different committeeindex: %v -> %v", d.ValidatorIndex, d.CommitteeIndex, n.CommitteeIndex)
				}
			}
		}
		return nil
	} else {
		tx, err := db.WriterDb.Beginx()
		if err != nil {
			return errors.Wrap(err, "tx")
		}

		defer func() {
			err := tx.Rollback()
			if err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()
		_, err = tx.Exec(`DELETE FROM sync_committees WHERE period = $1`, p)
		if err != nil {
			return errors.Wrap(err, "delete old entries")
		}

		err = modules.ExportSyncCommitteeAtPeriod(rpcClient, p, tx)
		if err != nil {
			return errors.Wrap(err, "export")
		}

		return tx.Commit()
	}
}
