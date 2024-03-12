package dataaccess

import (
	"context"
	"encoding/hex"
	"fmt"
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
	if d.Bigtable != nil {
		d.Bigtable.Close()
	}
}

func (d DataAccessService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.ReaderDb.Get(result, `
		SELECT 
			id, 
			user_id
		FROM users_val_dashboards
		WHERE id = $1
	`, dashboardId)
	return result, err
}

func (d DataAccessService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	result := &t.DashboardInfo{}

	err := d.ReaderDb.Get(result, `
		SELECT 
			uvd.id,
			uvd.user_id
		FROM users_val_dashboards_sharing uvds
		LEFT JOIN users_val_dashboards uvd ON uvd.id = uvds.dashboard_id
		WHERE uvds.public_id = $1
	`, publicDashboardId)
	return result, err
}

// param validators: slice of validator public keys or indices, a index should resolve to the newest index version
func (d DataAccessService) GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error) {
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

	// Query the database for the validators, those that are not found are ignored
	validatorsFromIdxPubkey := []t.VDBValidator{}
	err := d.ReaderDb.Select(&validatorsFromIdxPubkey, `
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
		return nil, fmt.Errorf("not all validators were found")
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
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	result := &t.VDBPostReturnData{}

	const nameCharLimit = 50
	const defaultGrpName = "default"

	if len(name) > nameCharLimit {
		return nil, fmt.Errorf("validator dashboard name too long, max %d characters, given %d characters", nameCharLimit, len(name))
	}

	tx, err := d.WriterDb.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

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
		return nil, errors.Wrap(err, "error committing tx")
	}

	return result, nil
}

func (d DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	tx, err := d.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

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
		return errors.Wrap(err, "error committing tx")
	}
	return nil
}

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (*t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverviewByValidators(validators)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBOverviewGroup, error) {
	result := &t.VDBOverviewGroup{}

	nameCharLimit := 50
	if len(name) > nameCharLimit {
		return nil, fmt.Errorf("validator dashboard group name too long, max %d characters, given %d characters", nameCharLimit, len(name))
	}

	// Create a new group that has the smallest unique id possible
	err := d.WriterDb.Get(result, `
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
	tx, err := d.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	// Delete the group
	result, err := tx.Exec(`
		DELETE FROM users_val_dashboards_groups WHERE dashboard_id = $1 AND id = $2
	`, dashboardId, groupId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("error group %v does not exist", groupId)
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
		return errors.Wrap(err, "error committing tx")
	}
	return nil
}

func (d DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	if len(validators) == 0 {
		// No validators to add
		return nil, nil
	}

	result := []t.VDBPostValidatorsData{}

	// Check that the group exists in the dashboard
	groupExists := false
	err := d.ReaderDb.Get(&groupExists, `
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
		return nil, fmt.Errorf("error group %v does not exist", groupId)
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
	err = d.ReaderDb.Select(&pubkeys, pubkeysQuery, flattenedValidators...)
	if err != nil {
		return []t.VDBPostValidatorsData{}, err
	}

	// Add all the validators to the dashboard and group
	addValidatorsArgsIntf := append([]interface{}{dashboardId, groupId}, flattenedValidators...)
	err = d.WriterDb.Select(&addedValidators, addValidatorsQuery, addValidatorsArgsIntf...)
	if err != nil {
		return []t.VDBPostValidatorsData{}, err
	}

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
		_, err := d.WriterDb.Exec(`
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
	_, err := d.WriterDb.Exec(deleteValidatorsQuery, deleteValidatorsArgsIntf...)

	return err
}

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	const nameCharLimit = 50
	if len(name) > nameCharLimit {
		return nil, fmt.Errorf("public validator dashboard name too long, max %d characters, given %d characters", nameCharLimit, len(name))
	}

	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}
	err := d.WriterDb.Get(&dbReturn, `
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

func (d DataAccessService) UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	const nameCharLimit = 50
	if len(name) > nameCharLimit {
		return nil, fmt.Errorf("public validator dashboard name too long, max %d characters, given %d characters", nameCharLimit, len(name))
	}

	dbReturn := struct {
		PublicId     string `db:"public_id"`
		Name         string `db:"name"`
		SharedGroups bool   `db:"shared_groups"`
	}{}
	err := d.WriterDb.Get(&dbReturn, `
		UPDATE users_val_dashboards_sharing SET
			name = $1,
			shared_groups = $2
		WHERE public_id = $3
		RETURNING public_id, name, shared_groups
	`, name, showGroupNames, publicDashboardId)
	if err != nil {
		return nil, err
	}

	result := &t.VDBPostPublicIdData{}
	result.PublicId = dbReturn.PublicId
	result.Name = dbReturn.Name
	result.ShareSettings.GroupNames = dbReturn.SharedGroups

	return result, nil
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error {
	_, err := d.WriterDb.Exec(`
		DELETE FROM users_val_dashboards_sharing WHERE public_id = $1
	`, publicDashboardId)

	return err
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error) {
	log.Infof("retrieving data for dashboard with id %d", dashboardId)

	// TODO: Get the validators from the dashboardId
	setSize := uint32(1000)

	validatorsMap := make(map[uint32]bool, setSize)

	validatorsArray := make([]uint32, 0, setSize)
	for i := uint32(0); i < setSize; i++ {
		validatorsMap[i] = true
		validatorsArray = append(validatorsArray, i)
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
	// TODO @peter_bitfly
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error) {
	// TODO @peter_bitfly
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
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
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}
