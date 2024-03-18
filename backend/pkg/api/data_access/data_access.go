package dataaccess

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	isort "sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type DataAccessor interface {
	GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error)
	GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error)

	GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error)

	GetUserDashboards(userId uint64) (*t.UserDashboardsData, error)

	// TODO move dashboard functions to a new interface+file
	CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error)
	RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error

	GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error)

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBOverviewGroup, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error

	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error)
	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error
	GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error)

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(publicDashboardId string, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(publicDashboardId string) error

	GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int], error)

	GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int], error)

	GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error)

	CloseDataAccessService()
}

type DataAccessService struct {
	dummy DummyService

	ReaderDb                *sqlx.DB
	WriterDb                *sqlx.DB
	AlloyReader             *sqlx.DB
	AlloyWriter             *sqlx.DB
	Bigtable                *db.Bigtable
	PersistentRedisDbClient *redis.Client
}

// ensure DataAccessService implements DataAccessor
var _ DataAccessor = DataAccessService{}

func NewDataAccessService(cfg *types.Config) DataAccessService {
	// Create the data access service
	dataAccessService := DataAccessService{
		dummy: NewDummyService()}

	// Initialize the database
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		//TODO adjust db functions to be able to set local reader/writer without setting the global ones
		db.MustInitDB(&types.DatabaseConfig{
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
		})

		dataAccessService.ReaderDb = db.ReaderDb
		dataAccessService.WriterDb = db.WriterDb
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitAlloyDb(&types.DatabaseConfig{
			Username:     cfg.AlloyWriter.Username,
			Password:     cfg.AlloyWriter.Password,
			Name:         cfg.AlloyWriter.Name,
			Host:         cfg.AlloyWriter.Host,
			Port:         cfg.AlloyWriter.Port,
			MaxOpenConns: cfg.AlloyWriter.MaxOpenConns,
			MaxIdleConns: cfg.AlloyWriter.MaxIdleConns,
			SSL:          cfg.AlloyWriter.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.AlloyReader.Username,
			Password:     cfg.AlloyReader.Password,
			Name:         cfg.AlloyReader.Name,
			Host:         cfg.AlloyReader.Host,
			Port:         cfg.AlloyReader.Port,
			MaxOpenConns: cfg.AlloyReader.MaxOpenConns,
			MaxIdleConns: cfg.AlloyReader.MaxIdleConns,
			SSL:          cfg.AlloyReader.SSL,
		})

		dataAccessService.AlloyReader = db.AlloyReader
		dataAccessService.AlloyWriter = db.AlloyWriter
	}()

	// Initialize the bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		dataAccessService.Bigtable = bt
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
		dataAccessService.PersistentRedisDbClient = rdc
	}()

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
	}

	// Return the result
	return dataAccessService
}

func (d DataAccessService) CloseDataAccessService() {
	if d.ReaderDb != nil {
		d.ReaderDb.Close()
	}
	if d.WriterDb != nil {
		d.WriterDb.Close()
	}
	if d.AlloyReader != nil {
		d.AlloyReader.Close()
	}
	if d.AlloyWriter != nil {
		d.AlloyWriter.Close()
	}
	if d.Bigtable != nil {
		d.Bigtable.Close()
	}
}

func (d DataAccessService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.AlloyReader.Get(result, `
		SELECT 
			id, 
			user_id
		FROM users_val_dashboards
		WHERE id = $1
	`, dashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("dashboard with id %d not found, err: %w", dashboardId, err)
	}
	return result, err
}

func (d DataAccessService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.AlloyReader.Get(result, `
		SELECT 
			uvd.id,
			uvd.user_id
		FROM users_val_dashboards_sharing uvds
		LEFT JOIN users_val_dashboards uvd ON uvd.id = uvds.dashboard_id
		WHERE uvds.public_id = $1
	`, publicDashboardId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("dashboard with public id %s not found, err: %w", publicDashboardId, err)
	}
	return result, err
}

// param validators: slice of validator public keys or indices, a index should resolve to the newest index version
func (d DataAccessService) GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error) {
	if len(validators) == 0 {
		return nil, nil
	}

	// Create a map to remove potential duplicates
	validatorMap := make(map[string]bool)
	for _, v := range validators {
		v = strings.TrimPrefix(v, "0x")
		validatorMap[v] = true
	}

	// Split the validators into pubkey and index slices
	validatorIdxs := pq.Int64Array{}
	validatorPubkeys := pq.ByteaArray{}
	for validator := range validatorMap {
		if utils.IsHash(validator) {
			validatorPubkey, err := hex.DecodeString(validator)
			if err != nil {
				return nil, err
			}
			validatorPubkeys = append(validatorPubkeys, validatorPubkey)
		} else if validatorIdx, parseErr := strconv.ParseUint(validator, 10, 31); parseErr == nil { // Limit to 31 bits to stay within math.MaxInt32
			validatorIdxs = append(validatorIdxs, int64(validatorIdx))
		} else {
			return nil, fmt.Errorf("invalid validator index or pubkey: %s", validator)
		}
	}

	// Query the database for the validators
	validatorsFromIdxPubkey := []t.VDBValidator{}
	err := d.AlloyReader.Select(&validatorsFromIdxPubkey, `
		SELECT 
			validator_index,
			MAX(validator_index_version) as validator_index_version
		FROM validators
		WHERE validator_index = ANY($1)
		GROUP BY validator_index
		UNION ALL
		SELECT 
			validator_index,
			validator_index_version
		FROM validators
		WHERE pubkey = ANY($2)
	`, validatorIdxs, validatorPubkeys)
	if err != nil {
		return nil, err
	}

	// Return an error if not every validator was found
	if len(validatorsFromIdxPubkey) != len(validatorMap) {
		return nil, fmt.Errorf("not all validators from strings were found")
	}

	// Create a map to remove potential duplicates
	validatorResultMap := make(map[t.VDBValidator]bool)
	for _, v := range validatorsFromIdxPubkey {
		validatorResultMap[v] = true
	}
	result := make([]t.VDBValidator, 0, len(validatorResultMap))
	for validator := range validatorResultMap {
		result = append(result, validator)
	}

	return result, nil
}

func (d DataAccessService) GetUserDashboards(userId uint64) (*t.UserDashboardsData, error) {
	result := &t.UserDashboardsData{}

	// Get the validator dashboards
	err := db.AlloyReader.Select(&result.ValidatorDashboards, `
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
	err = db.AlloyReader.Select(&result.AccountDashboards, `
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

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	const defaultGrpName = "default"

	tx, err := d.AlloyWriter.Beginx()
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
	`, result.Id, defaultGrpName)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error committing tx to create a validator dashboard: %w", err)
	}

	return result, nil
}

func (d DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	tx, err := d.AlloyWriter.Beginx()
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

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBOverviewGroup, error) {
	result := &t.VDBOverviewGroup{}

	// Create a new group that has the smallest unique id possible
	err := d.AlloyWriter.Get(result, `
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

func (d DataAccessService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	tx, err := d.AlloyWriter.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions to remove a validator dashboard group: %w", err)
	}
	defer utils.Rollback(tx)

	// Delete the group
	result, err := tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1 AND id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	// Check if the group was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("error group %v does not exist, cannot remove it", groupId)
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

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// WORKING spletka
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	if len(validators) == 0 {
		// No validators to add
		return nil, nil
	}

	// Check that the group exists in the dashboard
	groupExists := false
	err := d.AlloyReader.Get(&groupExists, `
		SELECT EXISTS(
			SELECT
				dashboard_id,
				id
			FROM users_val_dashboards_groups
			WHERE dashboard_id = $1 AND id = $2
		)
	`, dashboardId, groupId)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, fmt.Errorf("error group %v does not exist, cannot add validators to it", groupId)
	}

	pubkeys := []struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		ValidatorIndexVersion uint64 `db:"validator_index_version"`
		Pubkey                []byte `db:"pubkey"`
	}{}

	addedValidators := []struct {
		ValidatorIndex        uint64 `db:"validator_index"`
		ValidatorIndexVersion uint64 `db:"validator_index_version"`
		GroupId               uint64 `db:"group_id"`
	}{}

	// Query to find the pubkey for each validator index and version pair
	pubkeysQuery := `
		SELECT
			validator_index,
			validator_index_version,
			pubkey
		FROM validators
		WHERE (validator_index, validator_index_version) IN (
	`

	// Query to add the validator and version pairs to the dashboard and group
	addValidatorsQuery := `
		INSERT INTO users_val_dashboards_validators (dashboard_id, group_id, validator_index, validator_index_version)
			VALUES 
	`

	flattenedValidators := make([]interface{}, 0, len(validators)*2)
	for idx, v := range validators {
		flattenedValidators = append(flattenedValidators, v.Index, v.Version)
		pubkeysQuery += fmt.Sprintf("($%d, $%d), ", idx*2+1, idx*2+2)
		addValidatorsQuery += fmt.Sprintf("($1, $2, $%d, $%d), ", idx*2+3, idx*2+4)
	}
	pubkeysQuery = pubkeysQuery[:len(pubkeysQuery)-2] + ")"             // remove trailing comma
	addValidatorsQuery = addValidatorsQuery[:len(addValidatorsQuery)-2] // remove trailing comma

	// If a validator is already in the dashboard, update the group
	// If the validator is already in that group nothing changes but we will include it in the result anyway
	addValidatorsQuery += `
		ON CONFLICT (dashboard_id, validator_index, validator_index_version) DO UPDATE SET 
			dashboard_id = EXCLUDED.dashboard_id,
			group_id = EXCLUDED.group_id,
			validator_index = EXCLUDED.validator_index,
			validator_index_version = EXCLUDED.validator_index_version
		RETURNING validator_index, validator_index_version, group_id
	`

	// Find all the pubkeys
	err = d.AlloyReader.Select(&pubkeys, pubkeysQuery, flattenedValidators...)
	if err != nil {
		return nil, err
	}

	// Add all the validators to the dashboard and group
	addValidatorsArgsIntf := append([]interface{}{dashboardId, groupId}, flattenedValidators...)
	err = d.AlloyWriter.Select(&addedValidators, addValidatorsQuery, addValidatorsArgsIntf...)
	if err != nil {
		return nil, err
	}

	// Combine the pubkeys and group ids for the result
	pubkeysMap := make(map[t.VDBValidator]string, len(pubkeys))
	for _, pubKeyInfo := range pubkeys {
		pubkeysMap[t.VDBValidator{
			Index:   pubKeyInfo.ValidatorIndex,
			Version: pubKeyInfo.ValidatorIndexVersion}] = fmt.Sprintf("%#x", pubKeyInfo.Pubkey)
	}

	addedValidatorsMap := make(map[t.VDBValidator]uint64, len(addedValidators))
	for _, addedValidatorInfo := range addedValidators {
		addedValidatorsMap[t.VDBValidator{
			Index:   addedValidatorInfo.ValidatorIndex,
			Version: addedValidatorInfo.ValidatorIndexVersion}] = addedValidatorInfo.GroupId
	}

	result := []t.VDBPostValidatorsData{}
	for _, validator := range validators {
		result = append(result, t.VDBPostValidatorsData{
			PublicKey: pubkeysMap[validator],
			GroupId:   addedValidatorsMap[validator],
		})
	}

	return result, nil
}
func (d DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	if len(validators) == 0 {
		// Remove all validators for the dashboard
		_, err := d.AlloyWriter.Exec(`
			DELETE FROM users_val_dashboards_validators 
			WHERE dashboard_id = $1
		`, dashboardId)
		return err
	}

	//Create the query to delete validators
	deleteValidatorsQuery := `
		DELETE FROM users_val_dashboards_validators
		WHERE dashboard_id = $1 AND (validator_index, validator_index_version) IN (
	`

	flattenedValidators := make([]interface{}, 0, len(validators)*2)
	for idx, v := range validators {
		flattenedValidators = append(flattenedValidators, v.Index, v.Version)
		deleteValidatorsQuery += fmt.Sprintf("($%d, $%d), ", idx*2+2, idx*2+3)
	}
	deleteValidatorsQuery = deleteValidatorsQuery[:len(deleteValidatorsQuery)-2] + ")" // remove trailing comma

	// Delete the validators
	deleteValidatorsArgsIntf := append([]interface{}{dashboardId}, flattenedValidators...)
	_, err := d.AlloyWriter.Exec(deleteValidatorsQuery, deleteValidatorsArgsIntf...)

	return err
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Create the public validator dashboard, multiple entries for the same dashboard are possible
	err := d.AlloyWriter.Get(&dbReturn, `
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

func (d DataAccessService) UpdateValidatorDashboardPublicId(publicDashboardId string, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}

	// Update the name and settings of the public validator dashboard
	err := d.AlloyWriter.Get(&dbReturn, `
		UPDATE users_val_dashboards_sharing SET
			name = $1,
			shared_groups = $2
		WHERE public_id = $3
		RETURNING public_id, name, shared_groups
	`, name, showGroupNames, publicDashboardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("public dashboard id %v does not exist, cannot update it", publicDashboardId)
		}
		return nil, err
	}

	result := &t.VDBPostPublicIdData{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(publicDashboardId string) error {
	// Delete the public validator dashboard
	result, err := d.AlloyWriter.Exec(`
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

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error) {
	var validatorsArray []uint32
	if dashboardId.Validators == nil {
		err := db.AlloyReader.Select(&validatorsArray, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY validator_index`, dashboardId.Id)
		if err != nil {
			return nil, err
		}
	} else {
		validatorsArray = make([]uint32, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			validatorsArray = append(validatorsArray, uint32(validator.Index))
		}
	}

	validatorsMap := make(map[uint32]bool, len(validatorsArray))
	for _, validatorIndex := range validatorsArray {
		validatorsMap[validatorIndex] = true
	}

	// Get min/max slot/epoch
	headEpoch := cache.LatestEpoch.Get() // Reminder: Currently it is possible to get the head epoch from the cache but nothing sets it in v2
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2
	maxEpoch := headEpoch + 1

	maxValidatorsInResponse := 6

	dutiesInfo, releaseLock, err := services.GetCurrentDutiesInfo()
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

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
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
			err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), validatorList)
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
			err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), dashboardId)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		}

		data := make(map[uint64]float64)
		for _, result := range queryResult {
			efficiency := float64(0)

			if !result.AttestationEfficiency.Valid && !result.ProposerEfficiency.Valid && !result.SyncEfficiency.Valid {
				efficiency = 0
			} else if result.AttestationEfficiency.Valid && !result.ProposerEfficiency.Valid && !result.SyncEfficiency.Valid {
				efficiency = result.AttestationEfficiency.Float64 * 100.0
			} else if result.AttestationEfficiency.Valid && result.ProposerEfficiency.Valid && !result.SyncEfficiency.Valid {
				efficiency = ((56.0 / 64.0 * result.AttestationEfficiency.Float64) + (8.0 / 64.0 * result.ProposerEfficiency.Float64)) * 100.0
			} else if result.AttestationEfficiency.Valid && !result.ProposerEfficiency.Valid && result.SyncEfficiency.Valid {
				efficiency = ((62.0 / 64.0 * result.AttestationEfficiency.Float64) + (2.0 / 64.0 * result.SyncEfficiency.Float64)) * 100.0
			} else {
				efficiency = (((54.0 / 64.0) * result.AttestationEfficiency.Float64) + ((8.0 / 64.0) * result.ProposerEfficiency.Float64) + ((2.0 / 64.0) * result.SyncEfficiency.Float64)) * 100.0
			}

			if efficiency < 0 {
				efficiency = 0
			}

			data[result.GroupId] = efficiency
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

			err := db.AlloyReader.Select(&queryResult, `SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY group_id, validator_index`, dashboardId.Id)
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

			ret[groupId].EfficiencyLast24h = efficiency
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

			ret[groupId].EfficiencyLast7d = efficiency
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

			ret[groupId].EfficiencyLast31d = efficiency
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

			ret[groupId].EfficiencyAllTime = efficiency
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

	isort.Slice(retArr, func(i, j int) bool {
		return retArr[i].GroupId < retArr[j].GroupId
	})

	paging := &t.Paging{
		TotalCount: uint64(len(retArr)),
	}

	return retArr, paging, nil
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error) {
	ret := t.VDBGroupSummaryData{}
	wg := errgroup.Group{}

	query := `select
			users_val_dashboards_validators.validator_index,
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
			attestations_scheduled,
			attestations_executed,
			attestation_head_executed,
			attestation_source_executed,
			attestation_target_executed,
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
			sync_chance,
			block_chance,
			inclusion_delay_sum
		from users_val_dashboards_validators
		join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and group_id = $2)
		`

	if dashboardId.Validators != nil {
		query = `select
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
			attestations_scheduled,
			attestations_executed,
			attestation_head_executed,
			attestation_source_executed,
			attestation_target_executed,
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
			sync_chance,
			block_chance,
			inclusion_delay_sum
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

	retrieveAndProcessData := func(query, table string, dashboardId t.VDBIdPrimary, groupId int64, validators []uint64) (*t.VDBGroupSummaryColumn, error) {
		data := t.VDBGroupSummaryColumn{}
		var rows []*queryResult
		var err error

		if len(validators) > 0 {
			err = db.AlloyReader.Select(&rows, fmt.Sprintf(query, table), validators)
		} else {
			err = db.AlloyReader.Select(&rows, fmt.Sprintf(query, table), dashboardId, groupId)
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
		apr := (float64(reward) / (float64(32e9) * float64(len(rows)))) * 365.0 * 100.0

		data.Apr.Cl = apr
		data.Income.Cl = decimal.NewFromInt(reward).Mul(decimal.NewFromInt(1e9))

		data.Apr.El = 0

		data.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
		if data.AttestationEfficiency < 0 {
			data.AttestationEfficiency = 0
		}

		data.Luck.Proposal.Percent = (float64(data.Proposals.StatusCount.Failed) + float64(data.Proposals.StatusCount.Success)) / totalBlockChance * 100
		data.Luck.Sync.Percent = (float64(data.SyncCommittee.StatusCount.Failed) + float64(data.SyncCommittee.StatusCount.Success)) / totalSyncChance * 100
		if totalInclusionDelayDivisor > 0 {
			data.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
		} else {
			data.AttestationAvgInclDist = 0
		}

		return &data, nil
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_daily", dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last24h = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_weekly", dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last7d = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_monthly", dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last31d = *data
		return nil
	})
	// wg.Go(func() error {
	// 	data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_total", dashboardId.Id, groupId, validators)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	ret.AllTime = *data
	// 	return nil
	// })
	err := wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}

	return &ret, nil
}

// for summary charts: series id is group id, no stack

func (d DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardClDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}
