package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

type DataAccessor interface {
	GetUserInfo(email string) (*t.User, error)
	GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error)
	GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error)

	GetValidatorsFromSlices(indices []uint64, publicKeys []string) ([]t.VDBValidator, error)

	GetUserDashboards(userId uint64) (*t.UserDashboardsData, error)

	// TODO move dashboard functions to a new interface+file
	CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error)
	RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error

	GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error)

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error)
	UpdateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error

	GetValidatorDashboardGroupExists(dashboardId t.VDBIdPrimary, groupId uint64) (bool, error)
	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error)
	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error
	GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error)

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) error

	GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error)
	GetValidatorDashboardValidatorIndices(dashboardId t.VDBId, groupId int64, duty enums.ValidatorDuty, period enums.TimePeriod) ([]uint64, error)

	GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error)

	GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error)

	CloseDataAccessService()
}

type DataAccessService struct {
	dummy *DummyService

	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	alloyReader             *sqlx.DB
	alloyWriter             *sqlx.DB
	userReader              *sqlx.DB
	userWriter              *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client

	services *services.Services
}

// ensure DataAccessService pointer implements DataAccessor
var _ DataAccessor = (*DataAccessService)(nil)

func NewDataAccessService(cfg *types.Config) *DataAccessService {
	// Create the data access service
	das := createDataAccessService(cfg)

	// TODO: We set the global db connections here to have access to the functions in the db package
	// which use them without having to rewrite every single one.
	// This should be removed and the db functions should become methods of a struct that contains the db pointers.
	db.ReaderDb = das.readerDb
	db.WriterDb = das.writerDb
	db.AlloyReader = das.alloyReader
	db.AlloyWriter = das.alloyWriter
	db.BigtableClient = das.bigtable
	db.PersistentRedisDbClient = das.persistentRedisDbClient

	// Create the services
	das.services = services.NewServices(das.readerDb, das.writerDb, das.alloyReader, das.alloyWriter, das.bigtable, das.persistentRedisDbClient)

	// Initialize the services
	das.services.InitServices()

	return das
}

func createDataAccessService(cfg *types.Config) *DataAccessService {
	dataAccessService := DataAccessService{
		dummy: NewDummyService()}

	// Initialize the database
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.writerDb, dataAccessService.readerDb = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.WriterDatabase.Username,
				Password:     cfg.WriterDatabase.Password,
				Name:         cfg.WriterDatabase.Name,
				Host:         cfg.WriterDatabase.Host,
				Port:         cfg.WriterDatabase.Port,
				MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
				MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			},
			&types.DatabaseConfig{
				Username:     cfg.ReaderDatabase.Username,
				Password:     cfg.ReaderDatabase.Password,
				Name:         cfg.ReaderDatabase.Name,
				Host:         cfg.ReaderDatabase.Host,
				Port:         cfg.ReaderDatabase.Port,
				MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
				MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			},
		)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.alloyWriter, dataAccessService.alloyReader = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.AlloyWriter.Username,
				Password:     cfg.AlloyWriter.Password,
				Name:         cfg.AlloyWriter.Name,
				Host:         cfg.AlloyWriter.Host,
				Port:         cfg.AlloyWriter.Port,
				MaxOpenConns: cfg.AlloyWriter.MaxOpenConns,
				MaxIdleConns: cfg.AlloyWriter.MaxIdleConns,
				SSL:          cfg.AlloyWriter.SSL,
			},
			&types.DatabaseConfig{
				Username:     cfg.AlloyReader.Username,
				Password:     cfg.AlloyReader.Password,
				Name:         cfg.AlloyReader.Name,
				Host:         cfg.AlloyReader.Host,
				Port:         cfg.AlloyReader.Port,
				MaxOpenConns: cfg.AlloyReader.MaxOpenConns,
				MaxIdleConns: cfg.AlloyReader.MaxIdleConns,
				SSL:          cfg.AlloyReader.SSL,
			},
		)
	}()

	// Initialize the user database
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.userWriter, dataAccessService.userReader = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.Frontend.WriterDatabase.Username,
				Password:     cfg.Frontend.WriterDatabase.Password,
				Name:         cfg.Frontend.WriterDatabase.Name,
				Host:         cfg.Frontend.WriterDatabase.Host,
				Port:         cfg.Frontend.WriterDatabase.Port,
				MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
				MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
			},
			&types.DatabaseConfig{
				Username:     cfg.Frontend.ReaderDatabase.Username,
				Password:     cfg.Frontend.ReaderDatabase.Password,
				Name:         cfg.Frontend.ReaderDatabase.Name,
				Host:         cfg.Frontend.ReaderDatabase.Host,
				Port:         cfg.Frontend.ReaderDatabase.Port,
				MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
				MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
			},
		)
	}()

	// Initialize the bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		dataAccessService.bigtable = bt
	}()

	// Initialize the tiered cache (redis)
	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	// Initialize the persistent redis client
	wg.Add(1)
	go func() {
		defer wg.Done()
		rdc := redis.NewClient(&redis.Options{
			Addr:        utils.Config.RedisSessionStoreEndpoint,
			ReadTimeout: time.Second * 60,
		})

		if err := rdc.Ping(context.Background()).Err(); err != nil {
			log.Fatal(err, "error connecting to persistent redis store", 0)
		}
		dataAccessService.persistentRedisDbClient = rdc
	}()

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
	}

	// Return the result
	return &dataAccessService
}

func (d *DataAccessService) CloseDataAccessService() {
	if d.readerDb != nil {
		d.readerDb.Close()
	}
	if d.writerDb != nil {
		d.writerDb.Close()
	}
	if d.alloyReader != nil {
		d.alloyReader.Close()
	}
	if d.alloyWriter != nil {
		d.alloyWriter.Close()
	}
	if d.bigtable != nil {
		d.bigtable.Close()
	}
}

var ErrNotFound = errors.New("not found")

//////////////////// 		Helper functions

func (d DataAccessService) getDashboardValidators(dashboardId t.VDBId) ([]uint32, error) {
	var validatorsArray []uint32
	if len(dashboardId.Validators) == 0 {
		err := d.alloyReader.Select(&validatorsArray, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY validator_index`, dashboardId.Id)
		if err != nil {
			return nil, err
		}
	} else {
		validatorsArray = make([]uint32, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			validatorsArray = append(validatorsArray, uint32(validator.Index))
		}
	}
	return validatorsArray, nil
}

func (d DataAccessService) calculateTotalEfficiency(attestationEff, proposalEff, syncEff sql.NullFloat64) float64 {
	efficiency := float64(0)

	if !attestationEff.Valid && !proposalEff.Valid && !syncEff.Valid {
		efficiency = 0
	} else if attestationEff.Valid && !proposalEff.Valid && !syncEff.Valid {
		efficiency = attestationEff.Float64 * 100.0
	} else if attestationEff.Valid && proposalEff.Valid && !syncEff.Valid {
		efficiency = ((56.0 / 64.0 * attestationEff.Float64) + (8.0 / 64.0 * proposalEff.Float64)) * 100.0
	} else if attestationEff.Valid && !proposalEff.Valid && syncEff.Valid {
		efficiency = ((62.0 / 64.0 * attestationEff.Float64) + (2.0 / 64.0 * syncEff.Float64)) * 100.0
	} else {
		efficiency = (((54.0 / 64.0) * attestationEff.Float64) + ((8.0 / 64.0) * proposalEff.Float64) + ((2.0 / 64.0) * syncEff.Float64)) * 100.0
	}

	if efficiency < 0 {
		efficiency = 0
	}

	return efficiency
}

//////////////////// 		Data Access

func (d *DataAccessService) GetUserInfo(email string) (*t.User, error) {
	// TODO @recy21
	result := &t.User{}
	err := d.userReader.Get(result, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id FROM users_app_subscriptions 
				left join users on users.id = user_id 
				WHERE users.email = $1 AND active = true
				ORDER BY CASE product_id
					WHEN 'whale' THEN 1
					WHEN 'goldfish' THEN 2
					WHEN 'plankton' THEN 3
					ELSE 4  -- For any other product_id values
				END, users_app_subscriptions.created_at DESC LIMIT 1
			)
		SELECT users.id as id, password, COALESCE(product_id, '') as product_id, COALESCE(user_group, '') AS user_group 
		FROM users
		left join latest_and_greatest_sub on latest_and_greatest_sub.user_id = users.id  
		WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: user with email %s not found", ErrNotFound, email)
	}
	return result, err
}

func (d *DataAccessService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.alloyReader.Get(result, `
		SELECT 
			id, 
			user_id
		FROM users_val_dashboards
		WHERE id = $1
	`, dashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: dashboard with id %v not found", ErrNotFound, dashboardId)
	}
	return result, err
}

func (d *DataAccessService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.alloyReader.Get(result, `
		SELECT 
			uvd.id,
			uvd.user_id
		FROM users_val_dashboards_sharing uvds
		LEFT JOIN users_val_dashboards uvd ON uvd.id = uvds.dashboard_id
		WHERE uvds.public_id = $1
	`, publicDashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: dashboard with public id %v not found", ErrNotFound, publicDashboardId)
	}
	return result, err
}

// param validators: slice of validator public keys or indices
func (d *DataAccessService) GetValidatorsFromSlices(indices []uint64, publicKeys []string) ([]t.VDBValidator, error) {
	if len(indices) == 0 && len(publicKeys) == 0 {
		return nil, nil
	}

	_, err := d.services.GetPubkeysOfValidatorIndexSlice(indices)
	if err != nil {
		return nil, err
	}

	extraIndices, err := d.services.GetValidatorIndexOfPubkeySlice(publicKeys)
	if err != nil {
		return nil, err
	}

	// convert to t.VDBValidator slice
	validators := make([]t.VDBValidator, len(indices)+len(publicKeys))
	for i, index := range append(indices, extraIndices...) {
		validators[i] = t.VDBValidator{Index: index}
	}

	// Create a map to remove potential duplicates
	validatorResultMap := utils.SliceToMap(validators)
	result := maps.Keys(validatorResultMap)

	return result, nil
}

func (d *DataAccessService) GetUserDashboards(userId uint64) (*t.UserDashboardsData, error) {
	result := &t.UserDashboardsData{}

	// Get the validator dashboards
	err := d.alloyReader.Select(&result.ValidatorDashboards, `
		SELECT 
			id,
			name
		FROM users_val_dashboards
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}

	// Get the account dashboards
	err = d.alloyReader.Select(&result.AccountDashboards, `
		SELECT 
			id,
			name
		FROM users_acc_dashboards
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to create a validator dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	// Create validator dashboard for user
	err = tx.Get(result, `
		INSERT INTO users_val_dashboards (user_id, network, name)
			VALUES ($1, $2, $3)
		RETURNING id, user_id, name, network, created_at
	`, userId, network, name)
	if err != nil {
		return nil, err
	}

	// Create a default group for the new dashboard
	_, err = tx.Exec(`
		INSERT INTO users_val_dashboards_groups (dashboard_id, name)
			VALUES ($1, $2)
	`, result.Id, t.DefaultGroupName)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to create a validator dashboard: %w", err)
	}

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions to remove a validator dashboard: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards WHERE id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all groups for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all validators for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_validators WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	// Delete all shared dashboards for the dashboard
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_sharing WHERE dashboard_id = $1
	`, dashboardId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to remove a validator dashboard: %w", err)
	}
	return nil
}

func (d *DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error) {
	validators, err := d.getDashboardValidators(dashboardId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving validators from dashboard id: %v", err)
	}
	wg := errgroup.Group{}
	data := t.VDBOverviewData{}

	// Groups
	if len(dashboardId.Validators) == 0 {
		// should have valid primary id
		wg.Go(func() error {
			var queryResult []struct {
				Id    uint32 `db:"id"`
				Name  string `db:"name"`
				Count uint64 `db:"count"`
			}
			query := `SELECT id, name, COUNT(validator_index)
			FROM
				users_val_dashboards_groups groups
			LEFT JOIN users_val_dashboards_validators validators
					ON groups.dashboard_id = validators.dashboard_id AND groups.id = validators.group_id
			WHERE
				groups.dashboard_id = $1
			GROUP BY
				groups.id, groups.name`
			if err := d.alloyReader.Select(&queryResult, query, dashboardId.Id); err != nil {
				return err
			}
			for _, res := range queryResult {
				data.Groups = append(data.Groups, t.VDBOverviewGroup{Id: uint64(res.Id), Name: res.Name, Count: res.Count})
			}
			return nil
		})
	}

	// Validator Status
	wg.Go(func() error {
		query := `SELECT status AS statename, COUNT(*) AS statecount
		FROM validators
		WHERE validatorindex = ANY($1)
		GROUP BY status`
		var queryResult []struct {
			Name  string `db:"statename"`
			Count uint64 `db:"statecount"`
		}
		err = d.readerDb.Select(&queryResult, query, validators)
		if err != nil {
			return fmt.Errorf("error retrieving validators data: %v", err)
		}
		for _, state := range queryResult {
			switch state.Name {
			case "exiting_online":
				fallthrough
			case "slashing_online":
				fallthrough
			case "active_online":
				data.Validators.Online += state.Count
			case "exiting_offline":
				fallthrough
			case "slashing_offline":
				fallthrough
			case "active_offline":
				data.Validators.Offline += state.Count
			case "deposited":
				fallthrough
			case "pending":
				data.Validators.Pending += state.Count
			case "slashed":
				data.Validators.Slashed += state.Count
			case "exited":
				data.Validators.Exited += state.Count
			}
		}
		return nil
	})

	// Rewards + Efficiency
	retrieveData := func(tableName string) {
		wg.Go(func() error {
			query := `select
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency,

				SUM(balance_start) AS balance_start,
				SUM(balance_end) AS balance_end,
				SUM(deposits_amount) AS deposits_amount,
				SUM(withdrawals_amount) AS withdrawals_amount,
				SUM(blocks_el_reward) AS blocks_el_reward
			from %[1]s
			where validator_index = ANY($1)
			`
			var queryResult struct {
				AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
				ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
				SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`

				BalanceStart      sql.NullInt64 `db:"balance_start"`
				BalanceEnd        sql.NullInt64 `db:"balance_end"`
				DepositsAmount    sql.NullInt64 `db:"deposits_amount"`
				WithdrawalsAmount sql.NullInt64 `db:"withdrawals_amount"`
				BlocksElReward    sql.NullInt64 `db:"blocks_el_reward"`
			}
			err = d.alloyReader.Get(&queryResult, fmt.Sprintf(query, tableName), validators)
			if err != nil {
				return fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
			var rewardsField *t.ClElValue[decimal.Decimal]
			var efficiencyField *float64
			switch tableName {
			case "validator_dashboard_data_rolling_daily":
				rewardsField = &data.Rewards.Last24h
				efficiencyField = &data.Efficiency.Last24h
			case "validator_dashboard_data_rolling_weekly":
				rewardsField = &data.Rewards.Last7d
				efficiencyField = &data.Efficiency.Last7d
			case "validator_dashboard_data_rolling_monthly":
				rewardsField = &data.Rewards.Last30d
				efficiencyField = &data.Efficiency.Last30d
			case "validator_dashboard_data_rolling_total":
				rewardsField = &data.Rewards.AllTime
				efficiencyField = &data.Efficiency.AllTime
			}
			(*rewardsField).El = decimal.NewFromInt(queryResult.BlocksElReward.Int64)
			(*rewardsField).Cl = decimal.NewFromInt(queryResult.BalanceEnd.Int64 + queryResult.WithdrawalsAmount.Int64 - queryResult.BalanceStart.Int64 - queryResult.DepositsAmount.Int64)
			*efficiencyField = d.calculateTotalEfficiency(queryResult.AttestationEfficiency, queryResult.ProposerEfficiency, queryResult.SyncEfficiency)
			return nil
		})
	}

	retrieveData("validator_dashboard_data_rolling_daily")
	retrieveData("validator_dashboard_data_rolling_weekly")
	retrieveData("validator_dashboard_data_rolling_monthly")
	retrieveData("validator_dashboard_data_rolling_total")

	// Apr
	// TODO APR is WIP; imo we need activation time per validator, calculate its respective apr and accumulate the average per timeframe
	// But waiting for Peter implementation of apr calc

	err = wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard overview data: %v", err)
	}

	return &data, nil
}

func (d *DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	result := &t.VDBPostCreateGroupData{}

	// Create a new group that has the smallest unique id possible
	err := d.alloyWriter.Get(result, `
		WITH NextAvailableId AS (
		    SELECT COALESCE(MIN(uvdg1.id) + 1, 0) AS next_id
		    FROM users_val_dashboards_groups uvdg1
		    LEFT JOIN users_val_dashboards_groups uvdg2 ON uvdg1.id + 1 = uvdg2.id AND uvdg1.dashboard_id = uvdg2.dashboard_id
		    WHERE uvdg1.dashboard_id = $1 AND uvdg2.id IS NULL
		)
		INSERT INTO users_val_dashboards_groups (id, dashboard_id, name)
			SELECT next_id, $1, $2
		FROM NextAvailableId
		RETURNING id, name
	`, dashboardId, name)

	return result, err
}

// updates the group name
func (d *DataAccessService) UpdateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Update the group name
	_, err = tx.Exec(`
		UPDATE users_val_dashboards_groups SET name = $1 WHERE dashboard_id = $2 AND id = $3
	`, name, dashboardId, groupId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to update a validator dashboard group: %w", err)
	}

	ret := &t.VDBPostCreateGroupData{
		Id:   groupId,
		Name: name,
	}
	return ret, nil
}

func (d *DataAccessService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	tx, err := d.alloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete the group
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1 AND id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	// Delete all validators for the group
	_, err = tx.Exec(`
		DELETE FROM users_val_dashboards_validators WHERE dashboard_id = $1 AND group_id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing tx to remove a validator dashboard group: %w", err)
	}
	return nil
}

func (d *DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// Initialize the cursor
	var currentCursor t.ValidatorsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ValidatorsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ValidatorsCursor: %w", err)
		}
	}

	type ValidatorGroupInfo struct {
		GroupId   uint64
		GroupName string
	}
	validatorGroupMap := make(map[uint64]ValidatorGroupInfo)
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			GroupId        uint64 `db:"group_id"`
			GroupName      string `db:"group_name"`
		}{}

		validatorsQuery := `
		SELECT 
			v.validator_index,
			v.group_id,
			g.name AS group_name
		FROM users_val_dashboards_validators v
		LEFT JOIN users_val_dashboards_groups g ON v.group_id = g.id AND v.dashboard_id = g.dashboard_id
		WHERE v.dashboard_id = $1
		`
		validatorsParams := []interface{}{dashboardId.Id}

		if groupId != t.AllGroups {
			validatorsQuery += " AND group_id = $2"
			validatorsParams = append(validatorsParams, groupId)
		}
		err := d.alloyReader.Select(&queryResult, validatorsQuery, validatorsParams...)
		if err != nil {
			return nil, nil, err
		}

		for _, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = ValidatorGroupInfo{
				GroupId:   res.GroupId,
				GroupName: res.GroupName,
			}
			validators = append(validators, res.ValidatorIndex)
		}
	} else {
		// In case a list of validators is provided, set the group to the default
		for _, validator := range dashboardId.Validators {
			validatorGroupMap[validator.Index] = ValidatorGroupInfo{
				GroupId:   t.DefaultGroupId,
				GroupName: t.DefaultGroupName,
			}
			validators = append(validators, validator.Index)
		}
	}
	var paging t.Paging

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, &paging, nil
	}

	// Get the current validator state
	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, nil, err
	}

	// Get the validator duties to check the last fulfilled attestation
	dutiesInfo, releaseValDutiesLock, err := d.services.GetCurrentDutiesInfo()
	defer releaseValDutiesLock()
	if err != nil {
		return nil, nil, err
	}

	// Set the threshold for "online" => "offline" to 2 epochs without attestation
	attestationThresholdSlot := uint64(0)
	twoEpochs := 2 * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if dutiesInfo.LatestSlot >= twoEpochs {
		attestationThresholdSlot = dutiesInfo.LatestSlot - twoEpochs
	}

	// Fill the data
	data := []t.VDBManageValidatorsTableRow{}
	for _, validator := range validators {
		metadata := validatorMapping.ValidatorMetadata[validator]

		row := t.VDBManageValidatorsTableRow{
			Index:                validator,
			PublicKey:            t.PubKey(hexutil.Encode(metadata.PublicKey)),
			GroupId:              validatorGroupMap[validator].GroupId,
			Balance:              utils.GWeiToWei(big.NewInt(int64(metadata.Balance))),
			WithdrawalCredential: t.Hash(hexutil.Encode(metadata.WithdrawalCredentials)),
		}

		status := ""
		switch constypes.ValidatorStatus(metadata.Status) {
		case constypes.PendingInitialized:
			status = "deposited"
		case constypes.PendingQueued:
			status = "pending"
			if metadata.Queues.ActivationIndex.Valid {
				row.QueuePosition = uint64(metadata.Queues.ActivationIndex.Int64)
			}
		case constypes.ActiveOngoing, constypes.ActiveExiting, constypes.ActiveSlashed:
			var lastAttestionSlot uint32
			for slot, attested := range dutiesInfo.EpochAttestationDuties[uint32(validator)] {
				if attested && slot > lastAttestionSlot {
					lastAttestionSlot = slot
				}
			}
			if lastAttestionSlot < uint32(attestationThresholdSlot) {
				status = "offline"
			} else {
				status = "online"
			}
		case constypes.ExitedUnslashed, constypes.ExitedSlashed, constypes.WithdrawalPossible, constypes.WithdrawalDone:
			if metadata.Slashed {
				status = "slashed"
			} else {
				status = "exited"
			}
		}
		row.Status = status

		if search == "" {
			data = append(data, row)
		} else {
			index, err := strconv.ParseUint(search, 10, 64)
			indexSearch := err == nil && index == row.Index

			pubKey := strings.ToLower(strings.TrimPrefix(search, "0x"))
			pubkeySearch := pubKey == strings.TrimPrefix(string(row.PublicKey), "0x")

			groupNameSearch := search == validatorGroupMap[validator].GroupName

			if indexSearch || pubkeySearch || groupNameSearch {
				data = append(data, row)
			}
		}
	}

	// no data found (searched for something that does not exist)
	if len(data) == 0 {
		return nil, &paging, nil
	}

	// Sort the result
	sort.Slice(data, func(i, j int) bool {
		switch colSort.Column {
		case enums.VDBManageValidatorsIndex:
			if data[i].Index != data[j].Index {
				return (data[i].Index < data[j].Index) != colSort.Desc
			}
		case enums.VDBManageValidatorsPublicKey:
			if data[i].PublicKey != data[j].PublicKey {
				return (data[i].PublicKey < data[j].PublicKey) != colSort.Desc
			}
		case enums.VDBManageValidatorsBalance:
			if data[i].Balance.Cmp(data[j].Balance) != 0 {
				return (data[i].Balance.Cmp(data[j].Balance) < 0) != colSort.Desc
			}
		case enums.VDBManageValidatorsStatus:
			if data[i].Status != data[j].Status {
				return (data[i].Status < data[j].Status) != colSort.Desc
			}
		case enums.VDBManageValidatorsWithdrawalCredential:
			if data[i].WithdrawalCredential != data[j].WithdrawalCredential {
				return (data[i].WithdrawalCredential < data[j].WithdrawalCredential) != colSort.Desc
			}
		}
		return false
	})

	// Find the index for the cursor and limit the data
	var cursorIndex uint64
	if currentCursor.IsValid() {
		for idx, row := range data {
			if row.Index == currentCursor.Index {
				cursorIndex = uint64(idx)
				break
			}
		}
	}

	var result []t.VDBManageValidatorsTableRow
	if currentCursor.IsReverse() {
		// opposite direction
		var limitCutoff uint64
		if cursorIndex > limit+1 {
			limitCutoff = cursorIndex - limit - 1
		}
		result = data[limitCutoff:cursorIndex]
	} else {
		if currentCursor.IsValid() {
			cursorIndex++
		}
		limitCutoff := utilMath.MinU64(cursorIndex+limit+1, uint64(len(data)))
		result = data[cursorIndex:limitCutoff]
	}

	// flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// no paging required
		return result, &paging, nil
	}

	// remove the last entry from data as it is only required for the check
	if moreDataFlag {
		if currentCursor.IsReverse() {
			result = result[1:]
		} else {
			result = result[:len(result)-1]
		}
	}

	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupExists(dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	groupExists := false
	err := d.alloyReader.Get(&groupExists, `
		SELECT EXISTS(
			SELECT
				dashboard_id,
				id
			FROM users_val_dashboards_groups
			WHERE dashboard_id = $1 AND id = $2
		)
	`, dashboardId, groupId)
	return groupExists, err
}

func (d *DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	if len(validators) == 0 {
		// No validators to add
		return nil, nil
	}

	validatorIndices := make([]uint64, 0, len(validators))
	for _, v := range validators {
		validatorIndices = append(validatorIndices, v.Index)
	}

	pubkeys := []struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Pubkey         []byte `db:"pubkey"`
	}{}

	addedValidators := []struct {
		ValidatorIndex uint64 `db:"validator_index"`
		GroupId        uint64 `db:"group_id"`
	}{}

	// Query to find the pubkey for each validator index
	pubkeysQuery := `
		SELECT
			validatorindex,
			pubkey
		FROM validators
		WHERE validatorindex = ANY($1)
	`

	// Query to add the validators to the dashboard and group
	addValidatorsQuery := `
		INSERT INTO users_val_dashboards_validators (dashboard_id, group_id, validator_index)
			VALUES 
	`

	for idx := range validatorIndices {
		addValidatorsQuery += fmt.Sprintf("($1, $2, $%d), ", idx+3)
	}
	addValidatorsQuery = addValidatorsQuery[:len(addValidatorsQuery)-2] // remove trailing comma

	// If a validator is already in the dashboard, update the group
	// If the validator is already in that group nothing changes but we will include it in the result anyway
	addValidatorsQuery += `
		ON CONFLICT (dashboard_id, validator_index) DO UPDATE SET 
			dashboard_id = EXCLUDED.dashboard_id,
			group_id = EXCLUDED.group_id,
			validator_index = EXCLUDED.validator_index
		RETURNING validator_index, group_id
	`

	// Find all the pubkeys
	err := d.alloyReader.Select(&pubkeys, pubkeysQuery, pq.Array(validatorIndices))
	if err != nil {
		return nil, err
	}

	// Add all the validators to the dashboard and group
	addValidatorsArgsIntf := []interface{}{dashboardId, groupId}
	for _, validatorIndex := range validatorIndices {
		addValidatorsArgsIntf = append(addValidatorsArgsIntf, validatorIndex)
	}
	err = d.alloyWriter.Select(&addedValidators, addValidatorsQuery, addValidatorsArgsIntf...)
	if err != nil {
		return nil, err
	}

	// Combine the pubkeys and group ids for the result
	pubkeysMap := make(map[uint64]string, len(pubkeys))
	for _, pubKeyInfo := range pubkeys {
		pubkeysMap[pubKeyInfo.ValidatorIndex] = fmt.Sprintf("%#x", pubKeyInfo.Pubkey)
	}

	addedValidatorsMap := make(map[uint64]uint64, len(addedValidators))
	for _, addedValidatorInfo := range addedValidators {
		addedValidatorsMap[addedValidatorInfo.ValidatorIndex] = addedValidatorInfo.GroupId
	}

	result := []t.VDBPostValidatorsData{}
	for _, validator := range validatorIndices {
		result = append(result, t.VDBPostValidatorsData{
			PublicKey: pubkeysMap[validator],
			GroupId:   addedValidatorsMap[validator],
		})
	}

	return result, nil
}
func (d *DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	if len(validators) == 0 {
		// Remove all validators for the dashboard
		_, err := d.alloyWriter.Exec(`
			DELETE FROM users_val_dashboards_validators 
			WHERE dashboard_id = $1
		`, dashboardId)
		return err
	}

	validatorIndices := make([]uint64, 0, len(validators))
	for _, v := range validators {
		validatorIndices = append(validatorIndices, v.Index)
	}

	//Create the query to delete validators
	deleteValidatorsQuery := `
		DELETE FROM users_val_dashboards_validators
		WHERE dashboard_id = $1 AND validator_index = ANY($2)
	`

	// Delete the validators
	_, err := d.alloyWriter.Exec(deleteValidatorsQuery, dashboardId, pq.Array(validatorIndices))

	return err
}

func (d *DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Create the public validator dashboard, multiple entries for the same dashboard are possible
	err := d.alloyWriter.Get(&dbReturn, `
		INSERT INTO users_val_dashboards_sharing (dashboard_id, name, shared_groups)
			VALUES ($1, $2, $3)
		RETURNING public_id, name, shared_groups
	`, dashboardId, name, showGroupNames)
	if err != nil {
		return nil, err
	}

	result := &t.VDBPostPublicIdData{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Update the name and settings of the public validator dashboard
	err := d.alloyWriter.Get(&dbReturn, `
		UPDATE users_val_dashboards_sharing SET
			name = $1,
			shared_groups = $2
		WHERE public_id = $3
		RETURNING public_id, name, shared_groups
	`, name, showGroupNames, publicDashboardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: public dashboard id %v not found", ErrNotFound, publicDashboardId)
		}
		return nil, err
	}

	result := &t.VDBPostPublicIdData{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d *DataAccessService) RemoveValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) error {
	// Delete the public validator dashboard
	result, err := d.alloyWriter.Exec(`
		DELETE FROM users_val_dashboards_sharing WHERE public_id = $1
	`, publicDashboardId)
	if err != nil {
		return err
	}

	// Check if the public validator dashboard was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("error public dashboard id %v does not exist, cannot remove it", publicDashboardId)
	}

	return err
}

func (d *DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error) {
	validatorsArray, err := d.getDashboardValidators(dashboardId)
	if err != nil {
		return nil, err
	}

	validatorsMap := utils.SliceToMap(validatorsArray)

	// Get min/max slot/epoch
	headEpoch := cache.LatestEpoch.Get() // Reminder: Currently it is possible to get the head epoch from the cache but nothing sets it in v2
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2
	maxEpoch := headEpoch + 1

	maxValidatorsInResponse := 6

	dutiesInfo, releaseLock, err := d.services.GetCurrentDutiesInfo()
	defer releaseLock() // important to unlock once done, otherwise data updater cant update the data
	if err != nil {
		return nil, err
	}

	epochToIndexMap := make(map[uint64]uint64)
	slotToIndexMap := make(map[uint64]uint64)

	// attProcessing := time.Duration(0)
	// Restructure proposal status, attestations, sync duties and slashings
	slotVizEpochs := make([]t.SlotVizEpoch, maxEpoch-minEpoch+1)
	for epochIdx := uint64(0); epochIdx <= maxEpoch-minEpoch; epochIdx++ {
		epoch := maxEpoch - epochIdx
		epochToIndexMap[epoch] = epochIdx

		// Set the epoch number
		slotVizEpochs[epochIdx].Epoch = epoch

		// every validator can only attest once per epoch
		// attestedValidators := make(map[uint32]bool, len(validatorsArray))

		// Set the slots
		slotVizEpochs[epochIdx].Slots = make([]t.VDBSlotVizSlot, slotsPerEpoch)
		for slotIdx := uint64(0); slotIdx < slotsPerEpoch; slotIdx++ {
			// Set the slot number
			slot := epoch*slotsPerEpoch + slotIdx
			slotVizEpochs[epochIdx].Slots[slotIdx].Slot = slot
			slotToIndexMap[slot] = slotIdx
			// Set the slot status
			status := "scheduled"
			if _, ok := dutiesInfo.SlotStatus[slot]; ok {
				switch dutiesInfo.SlotStatus[slot] {
				case 0, 2, 3:
					status = "missed"
				case 1:
					status = "proposed"
					// case 3:
					// 	status = "orphaned"
				}
			}
			slotVizEpochs[epochIdx].Slots[slotIdx].Status = status

			// Get the proposals for the slot
			if proposerIndex, ok := dutiesInfo.PropAssignmentsForSlot[slot]; ok {
				// Only add results for validators we care about
				if _, ok := validatorsMap[uint32(proposerIndex)]; ok {
					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal = &t.VDBSlotVizTuple{}

					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.Validator = dutiesInfo.PropAssignmentsForSlot[slot]
					dutyObject := slot
					if _, ok := dutiesInfo.SlotStatus[slot]; ok {
						if dutiesInfo.SlotStatus[slot] == 1 || dutiesInfo.SlotStatus[slot] == 3 {
							dutyObject = dutiesInfo.SlotBlock[slot]
						}
					}
					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.DutyObject = dutyObject
				}
			}

			// Get the sync summary for the slot
			if len(dutiesInfo.SyncAssignmentsForEpoch[epoch]) > 0 {
				for validator := range dutiesInfo.SyncAssignmentsForEpoch[epoch] {
					// only validators we care about
					if _, ok := validatorsMap[uint32(validator)]; !ok {
						continue
					}

					if slotVizEpochs[epochIdx].Slots[slotIdx].Syncs == nil {
						slotVizEpochs[epochIdx].Slots[slotIdx].Syncs = &t.VDBSlotVizStatus[t.VDBSlotVizDuty]{}
					}
					syncsRef := slotVizEpochs[epochIdx].Slots[slotIdx].Syncs

					if slot > dutiesInfo.LatestSlot {
						if syncsRef.Scheduled == nil {
							syncsRef.Scheduled = &t.VDBSlotVizDuty{}
						}
						syncsRef.Scheduled.TotalCount++
						if len(syncsRef.Scheduled.Validators) < maxValidatorsInResponse {
							syncsRef.Scheduled.Validators = append(syncsRef.Scheduled.Validators, validator)
						}
					} else if _, ok := dutiesInfo.SlotSyncParticipated[slot][validator]; ok {
						if syncsRef.Success == nil {
							syncsRef.Success = &t.VDBSlotVizDuty{}
						}
						syncsRef.Success.TotalCount++
					} else {
						if syncsRef.Failed == nil {
							syncsRef.Failed = &t.VDBSlotVizDuty{}
						}
						syncsRef.Failed.TotalCount++
						if len(syncsRef.Failed.Validators) < maxValidatorsInResponse {
							syncsRef.Failed.Validators = append(syncsRef.Failed.Validators, validator)
						}
					}
				}
			}

			// Get the slashings for the slot
			slashedValidators := dutiesInfo.SlotValiPropSlashed[slot]
			slashedValidators = append(slashedValidators, dutiesInfo.SlotValiAttSlashed[slot]...)

			if proposerIndex, ok := dutiesInfo.PropAssignmentsForSlot[slot]; ok {
				// only add if we care about this validator
				if _, ok := validatorsMap[uint32(proposerIndex)]; ok {
					// One of the dashboard validators slashed
					for _, validator := range slashedValidators {
						if slotVizEpochs[epochIdx].Slots[slotIdx].Slashings == nil {
							slotVizEpochs[epochIdx].Slots[slotIdx].Slashings = &t.VDBSlotVizStatus[t.VDBSlotVizSlashing]{}
						}
						slashingsRef := slotVizEpochs[epochIdx].Slots[slotIdx].Slashings

						if slashingsRef.Success == nil {
							slashingsRef.Success = &t.VDBSlotVizSlashing{}
						}

						slashingsRef.Success.TotalCount++

						if len(slashingsRef.Success.Slashings) < maxValidatorsInResponse {
							slashing := t.VDBSlotVizTuple{
								Validator:  dutiesInfo.PropAssignmentsForSlot[slot], // Slashing validator
								DutyObject: validator,                               // Slashed validator
							}
							slashingsRef.Success.Slashings = append(slashingsRef.Success.Slashings, slashing)
						}
					}
				}
			}
			for _, validator := range slashedValidators {
				if _, ok := validatorsMap[uint32(validator)]; !ok {
					continue
				}
				// One of the dashboard validators got slashed
				if slotVizEpochs[epochIdx].Slots[slotIdx].Slashings == nil {
					slotVizEpochs[epochIdx].Slots[slotIdx].Slashings = &t.VDBSlotVizStatus[t.VDBSlotVizSlashing]{}
				}
				slashingsRef := slotVizEpochs[epochIdx].Slots[slotIdx].Slashings

				if slashingsRef.Failed == nil {
					slashingsRef.Failed = &t.VDBSlotVizSlashing{}
				}

				slashingsRef.Failed.TotalCount++

				if len(slashingsRef.Failed.Slashings) < maxValidatorsInResponse {
					slashing := t.VDBSlotVizTuple{
						Validator:  dutiesInfo.PropAssignmentsForSlot[slot], // Slashing validator
						DutyObject: validator,                               // Slashed validator
					}
					slashingsRef.Failed.Slashings = append(slashingsRef.Failed.Slashings, slashing)
				}
			}
		}
	}

	// Hydrate the attestation data
	for _, validator := range validatorsArray {
		for slot, duty := range dutiesInfo.EpochAttestationDuties[validator] {
			epoch := utils.EpochOfSlot(uint64(slot))
			epochIdx, ok := epochToIndexMap[epoch]
			if !ok {
				continue
			}
			slotIdx, ok := slotToIndexMap[uint64(slot)]
			if !ok {
				continue
			}

			if slotVizEpochs[epochIdx].Slots[slotIdx].Attestations == nil {
				slotVizEpochs[epochIdx].Slots[slotIdx].Attestations = &t.VDBSlotVizStatus[t.VDBSlotVizDuty]{}
			}
			attestationsRef := slotVizEpochs[epochIdx].Slots[slotIdx].Attestations

			if uint64(slot) >= dutiesInfo.LatestSlot {
				if attestationsRef.Scheduled == nil {
					attestationsRef.Scheduled = &t.VDBSlotVizDuty{}
				}
				attestationsRef.Scheduled.TotalCount++
				if len(attestationsRef.Scheduled.Validators) < maxValidatorsInResponse {
					attestationsRef.Scheduled.Validators = append(attestationsRef.Scheduled.Validators, uint64(validator))
				}
			} else if duty {
				if attestationsRef.Success == nil {
					attestationsRef.Success = &t.VDBSlotVizDuty{}
				}
				attestationsRef.Success.TotalCount++
			} else {
				if attestationsRef.Failed == nil {
					attestationsRef.Failed = &t.VDBSlotVizDuty{}
				}
				attestationsRef.Failed.TotalCount++
				if len(attestationsRef.Failed.Validators) < maxValidatorsInResponse {
					attestationsRef.Failed.Validators = append(attestationsRef.Failed.Validators, uint64(validator))
				}
			}
		}
	}

	return slotVizEpochs, nil
}

func (d *DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	// TODO: implement sorting, filtering & paging
	ret := make(map[uint64]*t.VDBSummaryTableRow) // map of group id to result row
	retMux := &sync.Mutex{}

	// retrieve efficiency data for each time period, we cannot do sorting & filtering here as we need access to the whole set
	wg := errgroup.Group{}

	validators := make([]uint64, 0)
	if dashboardId.Validators != nil {
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}

		ret[0] = &t.VDBSummaryTableRow{
			Validators: append([]uint64{}, validators...),
		}
	}

	type queryResult struct {
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	retrieveAndProcessData := func(dashboardId t.VDBIdPrimary, validatorList []uint64, tableName string) (map[uint64]float64, error) {
		var queryResult []queryResult

		if len(validatorList) > 0 {
			query := `select 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
				select 
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
					SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
					SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
					from  %[1]s
				where validator_index = ANY($1)
			) as a;`
			err := d.alloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), validatorList)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		} else {
			query := `select group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
				select 
					group_id,
					SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
					SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
					SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
					from users_val_dashboards_validators 
				join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
				where dashboard_id = $1
				group by 1
			) as a;`
			err := d.alloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), dashboardId)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		}

		data := make(map[uint64]float64)
		for _, result := range queryResult {
			data[result.GroupId] = d.calculateTotalEfficiency(result.AttestationEfficiency, result.ProposerEfficiency, result.SyncEfficiency)
		}
		return data, nil
	}

	if len(validators) == 0 { // retrieve the validators & groups from the dashboard table
		wg.Go(func() error {
			type validatorsPerGroup struct {
				GroupId        uint64 `db:"group_id"`
				ValidatorIndex uint64 `db:"validator_index"`
			}

			var queryResult []validatorsPerGroup

			err := d.alloyReader.Select(&queryResult, `SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY group_id, validator_index`, dashboardId.Id)
			if err != nil {
				return fmt.Errorf("error retrieving validator groups for dashboard: %v", err)
			}

			retMux.Lock()
			for _, result := range queryResult {
				if ret[result.GroupId] == nil {
					ret[result.GroupId] = &t.VDBSummaryTableRow{
						GroupId: result.GroupId,
					}
				}

				if ret[result.GroupId].Validators == nil {
					ret[result.GroupId].Validators = make([]uint64, 0, 10)
				}

				if len(ret[result.GroupId].Validators) < 10 {
					ret[result.GroupId].Validators = append(ret[result.GroupId].Validators, result.ValidatorIndex)
				}
			}
			retMux.Unlock()
			return nil
		})
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_daily")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last24h = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_weekly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last7d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_monthly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last30d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_total")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.AllTime = efficiency
		}
		return nil
	})
	err := wg.Wait()

	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard summary data: %v", err)
	}

	retArr := make([]t.VDBSummaryTableRow, 0, len(ret))

	for _, row := range ret {
		retArr = append(retArr, *row)
	}

	sort.Slice(retArr, func(i, j int) bool {
		return retArr[i].GroupId < retArr[j].GroupId
	})

	paging := &t.Paging{
		TotalCount: uint64(len(retArr)),
	}

	return retArr, paging, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error) {
	ret := &t.VDBGroupSummaryData{}
	wg := errgroup.Group{}

	query := `select
			users_val_dashboards_validators.validator_index,
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestations_executed, 0) as attestations_executed,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			COALESCE(sync_rewards, 0) as sync_rewards,
			COALESCE(slashed, false) as slashed,
			COALESCE(balance_start, 0) as balance_start,
			COALESCE(balance_end, 0) as balance_end,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(deposits_amount, 0) as deposits_amount,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
			COALESCE(withdrawals_amount, 0) as withdrawals_amount,
			COALESCE(sync_chance, 0) as sync_chance,
			COALESCE(block_chance, 0) as block_chance,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum
		from users_val_dashboards_validators
		join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and group_id = $2)
		`

	if dashboardId.Validators != nil {
		query = `select
			validator_index,
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestations_executed, 0) as attestations_executed,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			COALESCE(sync_rewards, 0) as sync_rewards,
			COALESCE(slashed, false) as slashed,
			COALESCE(balance_start, 0) as balance_start,
			COALESCE(balance_end, 0) as balance_end,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(deposits_amount, 0) as deposits_amount,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
			COALESCE(withdrawals_amount, 0) as withdrawals_amount,
			COALESCE(sync_chance, 0) as sync_chance,
			COALESCE(block_chance, 0) as block_chance,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum
		from %[1]s
		where %[1]s.validator_index = ANY($1)
	`
	}

	validators := make([]uint64, 0)
	if dashboardId.Validators != nil {
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}
	}

	type queryResult struct {
		ValidatorIndex                    uint32 `db:"validator_index"`
		AttestationSourceReward           int64  `db:"attestations_source_reward"`
		AttestationTargetReward           int64  `db:"attestations_target_reward"`
		AttestationHeadReward             int64  `db:"attestations_head_reward"`
		AttestationInactivitytReward      int64  `db:"attestations_inactivity_reward"`
		AttestationInclusionReward        int64  `db:"attestations_inclusion_reward"`
		AttestationReward                 int64  `db:"attestations_reward"`
		AttestationIdealSourceReward      int64  `db:"attestations_ideal_source_reward"`
		AttestationIdealTargetReward      int64  `db:"attestations_ideal_target_reward"`
		AttestationIdealHeadReward        int64  `db:"attestations_ideal_head_reward"`
		AttestationIdealInactivitytReward int64  `db:"attestations_ideal_inactivity_reward"`
		AttestationIdealInclusionReward   int64  `db:"attestations_ideal_inclusion_reward"`
		AttestationIdealReward            int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationsExecuted      int64 `db:"attestations_executed"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  uint64          `db:"blocks_cl_reward"`
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`
		SyncRewards   int64  `db:"sync_rewards"`

		Slashed bool `db:"slashed"`

		BalanceStart int64 `db:"balance_start"`
		BalanceEnd   int64 `db:"balance_end"`

		DepositsCount  uint32 `db:"deposits_count"`
		DepositsAmount int64  `db:"deposits_amount"`

		WithdrawalsCount  uint32 `db:"withdrawals_count"`
		WithdrawalsAmount int64  `db:"withdrawals_amount"`

		SyncChance  float64 `db:"sync_chance"`
		BlockChance float64 `db:"block_chance"`

		InclusionDelaySum int64 `db:"inclusion_delay_sum"`
	}

	retrieveAndProcessData := func(query, table string, aprDivisor int, dashboardId t.VDBIdPrimary, groupId int64, validators []uint64) (*t.VDBGroupSummaryColumn, error) {
		data := t.VDBGroupSummaryColumn{}
		var rows []*queryResult
		var err error

		if len(validators) > 0 {
			err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table), validators)
		} else {
			err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table), dashboardId, groupId)
		}

		if err != nil {
			return nil, err
		}

		totalAttestationRewards := int64(0)
		totalIdealAttestationRewards := int64(0)
		totalStartBalance := int64(0)
		totalEndBalance := int64(0)
		totalDeposits := int64(0)
		totalWithdrawals := int64(0)
		totalSyncChance := float64(0)
		totalBlockChance := float64(0)
		totalInclusionDelaySum := int64(0)
		totalInclusionDelayDivisor := int64(0)

		for _, row := range rows {
			totalAttestationRewards += row.AttestationReward
			totalIdealAttestationRewards += row.AttestationIdealReward

			data.AttestationCount.Success += uint64(row.AttestationsExecuted)
			data.AttestationCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationsExecuted)

			data.AttestationsHead.StatusCount.Success += uint64(row.AttestationHeadExecuted)
			data.AttestationsHead.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

			data.AttestationsSource.StatusCount.Success += uint64(row.AttestationSourceExecuted)
			data.AttestationsSource.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

			data.AttestationsTarget.StatusCount.Success += uint64(row.AttestationTargetExecuted)
			data.AttestationsTarget.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

			data.Proposals.StatusCount.Success += uint64(row.BlocksProposed)
			data.Proposals.StatusCount.Failed += uint64(row.BlocksScheduled) - uint64(row.BlocksProposed)

			if row.BlocksScheduled > 0 {
				if data.Proposals.Validators == nil {
					data.Proposals.Validators = make([]uint64, 0, 10)
				}
				data.Proposals.Validators = append(data.Proposals.Validators, uint64(row.ValidatorIndex))
			}

			data.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
			data.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

			if row.SyncScheduled > 0 {
				if data.SyncCommittee.Validators == nil {
					data.SyncCommittee.Validators = make([]uint64, 0, 10)
				}
				data.SyncCommittee.Validators = append(data.SyncCommittee.Validators, uint64(row.ValidatorIndex))
			}

			if row.Slashed {
				data.Slashed.StatusCount.Failed++
				if data.Slashed.Validators == nil {
					data.Slashed.Validators = make([]uint64, 0, 10)
					data.Slashed.Validators = append(data.Slashed.Validators, uint64(row.ValidatorIndex))
				}
			} else {
				data.Slashed.StatusCount.Success++
			}

			totalStartBalance += row.BalanceStart
			totalEndBalance += row.BalanceEnd
			totalDeposits += row.DepositsAmount
			totalWithdrawals += row.WithdrawalsAmount
			totalSyncChance += row.SyncChance
			totalBlockChance += row.BlockChance
			totalInclusionDelaySum += row.InclusionDelaySum

			if row.InclusionDelaySum > 0 {
				totalInclusionDelayDivisor += row.AttestationsScheduled
			}
		}

		reward := totalEndBalance + totalWithdrawals - totalStartBalance - totalDeposits
		log.Infof("rows: %d, totalEndBalance: %d, totalWithdrawals: %d, totalStartBalance: %d, totalDeposits: %d", len(rows), totalEndBalance, totalWithdrawals, totalStartBalance, totalDeposits)
		apr := ((float64(reward) / float64(aprDivisor)) / (float64(32e9) * float64(len(rows)))) * 365.0 * 100.0
		if math.IsNaN(apr) {
			apr = 0
		}

		data.Apr.Cl = apr
		data.Income.Cl = decimal.NewFromInt(reward).Mul(decimal.NewFromInt(1e9))

		data.Apr.El = 0

		data.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
		if data.AttestationEfficiency < 0 || math.IsNaN(data.AttestationEfficiency) {
			data.AttestationEfficiency = 0
		}

		if totalBlockChance > 0 {
			data.Luck.Proposal.Percent = (float64(data.Proposals.StatusCount.Failed) + float64(data.Proposals.StatusCount.Success)) / totalBlockChance * 100
		} else {
			data.Luck.Proposal.Percent = 0
		}
		if totalSyncChance > 0 {
			data.Luck.Sync.Percent = (float64(data.SyncCommittee.StatusCount.Failed) + float64(data.SyncCommittee.StatusCount.Success)) / totalSyncChance * 100
		} else {
			data.Luck.Sync.Percent = 0
		}
		if totalInclusionDelayDivisor > 0 {
			data.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
		} else {
			data.AttestationAvgInclDist = 0
		}

		return &data, nil
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_daily", 1, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last24h = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_weekly", 7, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last7d = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_monthly", 31, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last30d = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_total", 1, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.AllTime = *data
		return nil
	})
	err := wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}
	return ret, nil
}

// for summary charts: series id is group id, no stack

func (d *DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error) {
	ret := &t.ChartData[int, float64]{}

	type queryResult struct {
		StartEpoch            uint64          `db:"epoch_start"`
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	var queryResults []queryResult

	cutOffDate := time.Date(2023, 9, 27, 23, 59, 59, 0, time.UTC).Add(time.Hour*24*30).AddDate(0, 0, -30)

	if dashboardId.Validators != nil {
		validatorList := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			validatorList = append(validatorList, validator.Index)
		}

		query := `select epoch_start, 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
			epoch_start,
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
				from  validator_dashboard_data_daily
			WHERE day > $1 AND validator_index = ANY($2)
		) as a ORDER BY epoch_start, group_id;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, validatorList)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	} else {
		query := `select epoch_start, group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
			epoch_start, 
				group_id,
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
				from users_val_dashboards_validators 
			join validator_dashboard_data_daily on validator_dashboard_data_daily.validator_index = users_val_dashboards_validators.validator_index
			where day > $1 AND dashboard_id = $2
			group by 1, 2
		) as a ORDER BY epoch_start, group_id;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, dashboardId.Id)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	epochsMap := make(map[uint64]bool)
	groups := make(map[uint64]bool)
	data := make(map[uint64]map[uint64]float64)
	for _, row := range queryResults {
		epochsMap[row.StartEpoch] = true
		groups[row.GroupId] = true

		if data[row.StartEpoch] == nil {
			data[row.StartEpoch] = make(map[uint64]float64)
		}
		data[row.StartEpoch][row.GroupId] = d.calculateTotalEfficiency(row.AttestationEfficiency, row.ProposerEfficiency, row.SyncEfficiency)
	}

	epochsArray := make([]uint64, 0, len(epochsMap))
	for epoch := range epochsMap {
		epochsArray = append(epochsArray, epoch)
	}
	sort.Slice(epochsArray, func(i, j int) bool {
		return epochsArray[i] < epochsArray[j]
	})

	groupsArray := make([]uint64, 0, len(groups))
	for group := range groups {
		groupsArray = append(groupsArray, group)
	}
	sort.Slice(groupsArray, func(i, j int) bool {
		return groupsArray[i] < groupsArray[j]
	})

	ret.Categories = epochsArray
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[uint64]*t.ChartSeries[int, float64])
	for group := range groups {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(epochsMap)),
		}
		seriesMap[group] = &series
	}

	for _, epoch := range epochsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[epoch][group])
		}
	}

	for _, series := range seriesMap {
		ret.Series = append(ret.Series, *series)
	}

	sort.Slice(ret.Series, func(i, j int) bool {
		return ret.Series[i].Id < ret.Series[j].Id
	})

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardValidatorIndices(dashboardId t.VDBId, groupId int64, duty enums.ValidatorDuty, period enums.TimePeriod) ([]uint64, error) {
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators in case a dashboard id is provided
		validatorsQuery := `
		SELECT 
			validator_index
		FROM users_val_dashboards_validators
		WHERE dashboard_id = $1
		`
		validatorsParams := []interface{}{dashboardId.Id}

		if groupId != t.AllGroups {
			validatorsQuery += " AND group_id = $2"
			validatorsParams = append(validatorsParams, groupId)
		}
		err := d.alloyReader.Select(&validators, validatorsQuery, validatorsParams...)
		if err != nil {
			return nil, err
		}
	} else {
		// In case a list of validators is provided use them
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, nil
	}

	if duty == enums.ValidatorDuties.None {
		// If we don't need to filter by duty return all validators in the dashboard and group
		return validators, nil
	}

	// Get the table name based on the period
	tableName := ""
	switch period {
	case enums.TimePeriods.AllTime:
		tableName = "validator_dashboard_data_rolling_total"
	case enums.TimePeriods.Last24h:
		tableName = "validator_dashboard_data_rolling_daily"
	case enums.TimePeriods.Last7d:
		tableName = "validator_dashboard_data_rolling_weekly"
	case enums.TimePeriods.Last30d:
		tableName = "validator_dashboard_data_rolling_monthly"
	}

	// Get the column condition based on the duty
	columnCond := ""
	switch duty {
	case enums.ValidatorDuties.Sync:
		columnCond = "sync_scheduled > 0"
	case enums.ValidatorDuties.Proposal:
		columnCond = "blocks_scheduled > 0"
	case enums.ValidatorDuties.Slashed:
		// TODO: Wait for slashings to be available in the database
		// columnCond = "(slashed OR slashings_executed > 0)"
		columnCond = "slashed"
	}

	// Get ALL validator indices for the given filters
	query := fmt.Sprintf(`
		SELECT
			validator_index
		FROM %s
		WHERE validator_index = ANY($1) AND %s`, tableName, columnCond)

	var result []uint64
	err := d.alloyReader.Select(&result, query, pq.Array(validators))
	return result, err
}

func (d *DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// TODO @recy21
	// bar chart for the CL and EL rewards for each group for each epoch. NO series for all groups combined
	// series id is group id, series property is 'cl' or 'el'
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, groupId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}

func (d *DataAccessService) GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	var err error
	currentDirection := enums.DESC // TODO: expose over parameter
	var currentCursor t.CLDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.CLDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLDepositsCursor: %w", err)
		}
	}

	var byteaArray pq.ByteaArray

	// Resolve validator indices to pubkeys
	if dashboardId.Validators != nil {
		validatorsArray := make([]uint64, len(dashboardId.Validators))
		for i, v := range dashboardId.Validators {
			validatorsArray[i] = v.Index
		}
		validatorPubkeys, err := d.services.GetPubkeysOfValidatorIndexSlice(validatorsArray)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve validator indices to pubkeys: %w", err)
		}

		// Convert pubkeys to bytes for PostgreSQL
		byteaArray = make(pq.ByteaArray, len(validatorPubkeys))
		for i, p := range validatorPubkeys {
			byteaArray[i], _ = hexutil.Decode(p)
		}
	}

	// Custom type for block_index
	var data []struct {
		GroupId              sql.NullInt64   `db:"group_id"`
		PublicKey            []byte          `db:"publickey"`
		Slot                 int64           `db:"block_slot"`
		SlotIndex            int64           `db:"block_index"`
		WithdrawalCredential []byte          `db:"withdrawalcredentials"`
		Amount               decimal.Decimal `db:"amount"`
		Signature            []byte          `db:"signature"`
	}

	query := `
			SELECT
				bd.publickey,
				bd.block_slot,
				bd.block_index,
				bd.amount,
				bd.signature,
				bd.withdrawalcredentials
		`

	var filter interface{}
	if dashboardId.Validators != nil {
		query += `
			FROM
				blocks_deposits bd
			WHERE
				bd.publickey = ANY ($1)`
		filter = byteaArray
	} else {
		query += `
			, cbdl.group_id
			FROM
				cached_blocks_deposits_lookup cbdl
				LEFT JOIN blocks_deposits bd ON bd.block_slot = cbdl.block_slot
					AND bd.block_index = cbdl.block_index
			WHERE
				cbdl.dashboard_id = $1`
		filter = dashboardId.Id
	}

	params := []interface{}{filter}
	filterFragment := ` ORDER BY bd.block_slot DESC, bd.block_index DESC`
	if currentCursor.IsValid() {
		filterFragment = ` AND (bd.block_slot < $2 or (bd.block_slot = $2 and bd.block_index < $3)) ` + filterFragment
		params = append(params, currentCursor.Slot, currentCursor.SlotIndex)
	}

	if currentDirection == enums.ASC && !currentCursor.IsReverse() || currentDirection == enums.DESC && currentCursor.IsReverse() {
		filterFragment = strings.Replace(strings.Replace(filterFragment, "<", ">", -1), "DESC", "ASC", -1)
	}

	if dashboardId.Validators == nil {
		filterFragment = strings.Replace(filterFragment, "bd.", "cbdl.", -1)
	}

	params = append(params, limit+1)
	filterFragment += fmt.Sprintf(" LIMIT $%d", len(params))

	err = db.AlloyReader.Select(&data, query+filterFragment, params...)

	if err != nil {
		return nil, nil, err
	}

	pubkeys := make([]string, len(data))
	for i, row := range data {
		pubkeys[i] = hexutil.Encode(row.PublicKey)
	}
	indices, err := d.services.GetValidatorIndexOfPubkeySlice(pubkeys)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to recover indices after query: %w", err)
	}

	responseData := make([]t.VDBConsensusDepositsTableRow, len(data))
	for i, row := range data {
		responseData[i] = t.VDBConsensusDepositsTableRow{
			PublicKey:            t.PubKey(pubkeys[i]),
			Index:                indices[i],
			Epoch:                utils.EpochOfSlot(uint64(row.Slot)),
			Slot:                 uint64(row.Slot),
			WithdrawalCredential: t.Hash(hexutil.Encode(row.WithdrawalCredential)),
			Amount:               row.Amount,
			Signature:            t.Hash(hexutil.Encode(row.WithdrawalCredential)),
		}
		if row.GroupId.Valid {
			responseData[i].GroupId = uint64(row.GroupId.Int64)
		} else {
			responseData[i].GroupId = t.DefaultGroupId
		}
	}
	var paging t.Paging

	moreDataFlag := len(responseData) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return responseData, &paging, nil
	}
	if moreDataFlag {
		// Remove the last entry as it is only required for the more data flag
		responseData = responseData[:len(responseData)-1]
		data = data[:len(data)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
		slices.Reverse(data)
	}

	p, err := utils.GetPagingFromData(data, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func (d *DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	// TODO: implement sorting, filtering & paging

	validatorGroupMap := make(map[uint64]uint64)
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			GroupId        uint64 `db:"group_id"`
		}{}

		validatorsQuery := `
			SELECT
			    validator_index,
			    group_id
			FROM
			    users_val_dashboards_validators
			WHERE
			    dashboard_id = $1`

		err := d.alloyReader.Select(&queryResult, validatorsQuery, dashboardId.Id)
		if err != nil {
			return nil, nil, err
		}

		for _, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = res.GroupId
			validators = append(validators, res.ValidatorIndex)
		}
	} else {
		// In case a list of validators is provided, set the group to default 0
		for _, validator := range dashboardId.Validators {
			validatorGroupMap[validator.Index] = t.DefaultGroupId
			validators = append(validators, validator.Index)
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, nil, nil
	}

	// Get the withdrawals for the validators
	queryResult := []struct {
		ValidatorIndex uint64 `db:"validator_index"`
		Epoch          uint64 `db:"epoch"`
		Amount         int64  `db:"withdrawals_amount"`
	}{}
	err := d.alloyReader.Select(&queryResult, `
		SELECT
		    validator_index,
			epoch,
		    withdrawals_amount
		FROM
		    validator_dashboard_data_epoch
		WHERE
		    validator_index = ANY ($1) AND withdrawals_amount > 0`, pq.Array(validators))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("error getting withdrawals for validators: %+v: %w", validators, err)
	}

	// Get the validators with withdrawals
	validatorsWithWithdrawals := make(map[uint64]bool, 0)
	for _, withdrawal := range queryResult {
		validatorsWithWithdrawals[withdrawal.ValidatorIndex] = true
	}

	// Get the current validator state
	validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseLock()
	if err != nil {
		return nil, nil, err
	}

	// Get the withdrawal addresses for the validators
	withdrawalAddresses := make(map[uint64]string)
	addressEns := make(map[string]string)
	for validator := range validatorsWithWithdrawals {
		withdrawalCredentials := validatorMapping.ValidatorMetadata[validator].WithdrawalCredentials
		withdrawalAddress, err := utils.GetAddressOfWithdrawalCredentials(withdrawalCredentials)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid withdrawal credentials for validator %d: %s", validator, hexutil.Encode(withdrawalCredentials))
		}
		withdrawalAddresses[validator] = hexutil.Encode(withdrawalAddress.Bytes())
		addressEns[hexutil.Encode(withdrawalAddress.Bytes())] = ""
	}

	// Get the ENS names for the addresses
	if err := db.GetEnsNamesForAddresses(addressEns); err != nil {
		return nil, nil, err
	}

	// Create the result
	result := make([]t.VDBWithdrawalsTableRow, 0)
	for _, withdrawal := range queryResult {
		address := withdrawalAddresses[withdrawal.ValidatorIndex]
		result = append(result, t.VDBWithdrawalsTableRow{
			Epoch:   withdrawal.Epoch,
			Index:   withdrawal.ValidatorIndex,
			GroupId: validatorGroupMap[withdrawal.ValidatorIndex],
			Recipient: t.Address{
				Hash: t.Hash(address),
				Ens:  addressEns[address],
			},
			Amount: decimal.NewFromInt(withdrawal.Amount),
		})
	}

	paging := &t.Paging{
		TotalCount: uint64(len(result)),
	}

	return result, paging, nil
}
