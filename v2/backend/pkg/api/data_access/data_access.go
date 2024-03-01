package dataaccess

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.DashboardData, error)

	CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error)
	GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (t.VDBOverviewData, error)
	GetValidatorDashboardOverviewByPublicId(publicDashboardId t.VDBIdPublic) (t.VDBOverviewData, error)
	GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (t.VDBOverviewData, error)
	RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error
	RemoveValidatorDashboardByPublicId(dashboardId t.VDBIdPublic) error

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (t.VDBOverviewGroup, error)
	CreateValidatorDashboardGroupByPublicId(dashboardId t.VDBIdPublic, name string) (t.VDBOverviewGroup, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error
	RemoveValidatorDashboardGroupByPublicId(dashboardId t.VDBIdPublic, groupId uint64) error

	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error)
	AddValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error)
	GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)
	GetValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)
	GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)
	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []string) error
	RemoveValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, validators []string) error

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	CreateValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error
	RemoveValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, publicDashboardId string) error

	GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error)
	GetValidatorDashboardSlotVizByPublicId(dashboardId t.VDBIdPublic) ([]t.SlotVizEpoch, error)
	GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardSummaryByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBIdPrimary, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardGroupSummaryByPublicId(dashboardId t.VDBIdPublic, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error)
	GetValidatorDashboardSummaryChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error)
	GetValidatorDashboardSummaryChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error)

	GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardRewardsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardGroupRewardsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error)
	GetValidatorDashboardRewardsChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error)
	GetValidatorDashboardRewardsChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error)

	GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)
	GetValidatorDashboardDutiesByPublicId(dashboardId t.VDBIdPublic, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)
	GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)
	GetValidatorDashboardBlocksByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)
	GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId t.VDBIdPrimary) (t.VDBHeatmap, error)
	GetValidatorDashboardHeatmapByPublicId(dashboardId t.VDBIdPublic) (t.VDBHeatmap, error)
	GetValidatorDashboardHeatmapByValidators(dashboardId t.VDBIdValidatorSet) (t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupHeatmapByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardElDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardElDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawalsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)

	CloseDataAccessService()
}

type DataAccessService struct {
	dummy DummyService

	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client
}

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

		dataAccessService.readerDb = db.ReaderDb
		dataAccessService.writerDb = db.WriterDb
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
			ReadTimeout: time.Second * 20,
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
	return dataAccessService
}

func (d DataAccessService) CloseDataAccessService() {
	if d.readerDb != nil {
		d.readerDb.Close()
	}
	if d.writerDb != nil {
		d.writerDb.Close()
	}
	if d.bigtable != nil {
		d.bigtable.Close()
	}
}

func (d DataAccessService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardOverviewByPublicId(publicDashboardId t.VDBIdPublic) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverviewByPublicId(publicDashboardId)
}

func (d DataAccessService) GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverviewByValidators(validators)
}

func (d DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboard(dashboardId)
}

func (d DataAccessService) RemoveValidatorDashboardByPublicId(dashboardId t.VDBIdPublic) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardByPublicId(dashboardId)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (t.VDBOverviewGroup, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardGroup(dashboardId, name)
}

func (d DataAccessService) CreateValidatorDashboardGroupByPublicId(dashboardId t.VDBIdPublic, name string) (t.VDBOverviewGroup, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardGroupByPublicId(dashboardId, name)
}

func (d DataAccessService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardGroup(dashboardId, groupId)
}

func (d DataAccessService) RemoveValidatorDashboardGroupByPublicId(dashboardId t.VDBIdPublic, groupId uint64) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardGroupByPublicId(dashboardId, groupId)
}

func (d DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error) {
	// TODO @recy21
	return d.dummy.AddValidatorDashboardValidators(dashboardId, groupId, validators)
}

func (d DataAccessService) AddValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error) {
	// TODO @recy21
	return d.dummy.AddValidatorDashboardValidatorsByPublicId(dashboardId, groupId, validators)
}

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByPublicId(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, cursor string, sort []t.Sort[t.VDBManageValidatorsTableColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardValidators(dashboardId, validators)
}

func (d DataAccessService) RemoveValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, validators []string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardValidatorsByPublicId(dashboardId, validators)
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardPublicId(dashboardId, name, showGroupNames)
}

func (d DataAccessService) CreateValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardPublicIdByPublicId(dashboardId, name, showGroupNames)
}

func (d DataAccessService) UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.UpdateValidatorDashboardPublicId(dashboardId, publicDashboardId, name, showGroupNames)
}

func (d DataAccessService) UpdateValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.UpdateValidatorDashboardPublicIdByPublicId(dashboardId, publicDashboardId, name, showGroupNames)
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardPublicId(dashboardId, publicDashboardId)
}

var getValidatorDashboardSlotVizMux = &sync.Mutex{}

func (d DataAccessService) RemoveValidatorDashboardPublicIdByPublicId(dashboardId t.VDBIdPublic, publicDashboardId string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardPublicIdByPublicId(dashboardId, publicDashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error) {
	log.Infof("retrieving data for dashboard with id %d", dashboardId)
	// make sure that the function is only executed once during development not to go oom
	getValidatorDashboardSlotVizMux.Lock()
	defer getValidatorDashboardSlotVizMux.Unlock()
	// TODO: Get the validators from the dashboardId

	dummyValidators := []uint64{900005, 900006, 900007, 900008, 900009}
	validatorsMap := make(map[uint64]bool)
	for _, v := range dummyValidators {
		validatorsMap[v] = true
	}

	// Get min/max slot/epoch
	headEpoch := cache.LatestEpoch.Get() // Reminder: Currently it is possible to get the head epoch from the cache but nothing sets it in v2
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2
	maxEpoch := headEpoch + 1

	// create waiting group for concurrency
	gOuter := &errgroup.Group{}

	// Get the fulfilled duties
	var validatorDutiesInfo []types.ValidatorDutyInfo
	gOuter.Go(func() error {
		var err error
		validatorDutiesInfo, err = db.GetValidatorDutiesInfo(d.readerDb, minEpoch*slotsPerEpoch)
		return err
	})

	// Gather the assignments data
	propAssignmentsForSlot := make(map[uint64]uint64)
	attAssignmentsForSlot := make(map[uint64]map[uint64]bool)
	totalSyncAssignmentsForEpoch := make(map[uint64][]uint64)
	syncAssignmentsForEpoch := make(map[uint64]map[uint64]bool)
	{
		muxPropAssignmentsForSlot := &sync.Mutex{}
		muxAttAssignmentsForSlot := &sync.Mutex{}
		muxTotalSyncAssignmentsForEpoch := &sync.Mutex{}
		muxSyncAssignmentsForEpoch := &sync.Mutex{}

		for e := minEpoch; e <= maxEpoch; e++ {
			epoch := e
			gOuter.Go(func() error {
				// Get the epoch assignments data
				key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "ea", epoch)
				encodedRedisCachedEpochAssignments, err := d.persistentRedisDbClient.Get(context.Background(), key).Result()
				if err != nil {
					return err
				}

				var serializedAssignmentsData bytes.Buffer
				_, err = serializedAssignmentsData.Write([]byte(encodedRedisCachedEpochAssignments))
				if err != nil {
					return err
				}
				var decodedRedisCachedEpochAssignments types.RedisCachedEpochAssignments

				dec := gob.NewDecoder(&serializedAssignmentsData)
				err = dec.Decode(&decodedRedisCachedEpochAssignments)
				if err != nil {
					return err
				}

				// Save the assignments data in maps

				// Proposals
				for slot, propValidator := range decodedRedisCachedEpochAssignments.Assignments.ProposerAssignments {
					// Only add results for validators we care about
					if _, isValid := validatorsMap[propValidator]; isValid {
						muxPropAssignmentsForSlot.Lock()
						propAssignmentsForSlot[slot] = propValidator
						muxPropAssignmentsForSlot.Unlock()
					}
				}

				// Attestations
				for key, attValidator := range decodedRedisCachedEpochAssignments.Assignments.AttestorAssignments {
					keyParts := strings.Split(key, "-")
					slot, err := strconv.ParseUint(keyParts[0], 10, 64)
					if err != nil {
						return err
					}

					muxAttAssignmentsForSlot.Lock()
					if attAssignmentsForSlot[slot] == nil {
						attAssignmentsForSlot[slot] = make(map[uint64]bool, 0)
					}
					// Only add results for validators we care about
					if _, isValid := validatorsMap[attValidator]; isValid {
						attAssignmentsForSlot[slot][attValidator] = true
					}
					muxAttAssignmentsForSlot.Unlock()
				}

				// Syncs
				muxTotalSyncAssignmentsForEpoch.Lock()
				totalSyncAssignmentsForEpoch[epoch] = decodedRedisCachedEpochAssignments.Assignments.SyncAssignments
				muxTotalSyncAssignmentsForEpoch.Unlock()
				muxSyncAssignmentsForEpoch.Lock()
				if syncAssignmentsForEpoch[epoch] == nil {
					syncAssignmentsForEpoch[epoch] = make(map[uint64]bool, 0)
				}
				for _, validator := range decodedRedisCachedEpochAssignments.Assignments.SyncAssignments {
					// Only add results for validators we care about
					if _, isValid := validatorsMap[validator]; isValid {
						syncAssignmentsForEpoch[epoch][validator] = true
					}
				}
				muxSyncAssignmentsForEpoch.Unlock()

				return nil
			})
		}
	}

	// wait for routines to complete
	if err := gOuter.Wait(); err != nil {
		return nil, err
	}

	// Restructure proposal status, attestations, sync duties and slashings
	latestSlot := uint64(0)
	slotStatus := make(map[uint64]uint64)
	slotBlock := make(map[uint64]uint64)
	slotAttested := make(map[uint64]map[uint64]bool)
	slotSyncParticipated := make(map[uint64]map[uint64]bool)
	slotValiPropSlashed := make(map[uint64]bool)
	slotValiAttSlashed := make(map[uint64]bool)
	for _, duty := range validatorDutiesInfo {
		if duty.Slot > latestSlot {
			latestSlot = duty.Slot
		}
		slotStatus[duty.Slot] = duty.Status
		slotBlock[duty.Slot] = duty.Block
		if duty.Status == 1 { // 1: Proposed
			// Attestations
			if duty.AttestedSlot.Valid {
				attestedSlot := uint64(duty.AttestedSlot.Int64)
				if slotAttested[attestedSlot] == nil {
					slotAttested[attestedSlot] = make(map[uint64]bool, 0)
				}
				for _, validator := range duty.Validators {
					slotAttested[attestedSlot][uint64(validator)] = true
				}
			}
			// Syncs
			if slotSyncParticipated[duty.Slot] == nil {
				slotSyncParticipated[duty.Slot] = make(map[uint64]bool, 0)

				partValidators := utils.GetParticipatingSyncCommitteeValidators(duty.SyncAggregateBits, totalSyncAssignmentsForEpoch[utils.EpochOfSlot(duty.Slot)])
				for _, validator := range partValidators {
					slotSyncParticipated[duty.Slot][validator] = true
				}
			}
			// Slashings
			if duty.ProposerSlashingsCount > 0 {
				slotValiPropSlashed[duty.Slot] = true
			}
			if duty.AttesterSlashingsCount > 0 {
				slotValiAttSlashed[duty.Slot] = true
			}
		}
	}

	slotVizEpochs := make([]t.SlotVizEpoch, maxEpoch-minEpoch+1)
	for epochIdx := uint64(0); epochIdx <= maxEpoch-minEpoch; epochIdx++ {
		epoch := maxEpoch - epochIdx

		// Set the epoch number
		slotVizEpochs[epochIdx].Epoch = epoch

		// Set the slots
		slotVizEpochs[epochIdx].Slots = make([]t.VDBSlotVizSlot, slotsPerEpoch)
		for slotIdx := uint64(0); slotIdx < slotsPerEpoch; slotIdx++ {
			// Set the slot number
			slot := epoch*slotsPerEpoch + slotIdx
			slotVizEpochs[epochIdx].Slots[slotIdx].Slot = slot

			// Set the slot status
			status := "scheduled"
			if _, ok := slotStatus[slot]; ok {
				switch slotStatus[slot] {
				case 0, 2:
					status = "missed"
				case 1:
					status = "proposed"
				case 3:
					status = "orphaned"
				}
			}
			slotVizEpochs[epochIdx].Slots[slotIdx].Status = status

			// Get the proposals for the slot
			if _, ok := propAssignmentsForSlot[slot]; ok {
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposal = &t.VDBSlotVizActiveDuty{}

				slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.Validator = propAssignmentsForSlot[slot]

				status := "scheduled"
				dutyObject := slot
				if _, ok := slotStatus[slot]; ok {
					switch slotStatus[slot] {
					case 0, 2:
						status = "failed"
					case 1, 3:
						status = "success"
						dutyObject = slotBlock[slot]
					}
				}
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.Status = status
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.DutyObject = dutyObject
			}

			// Get the attestation summary for the slot
			if len(attAssignmentsForSlot[slot]) > 0 {
				slotVizEpochs[epochIdx].Slots[slotIdx].Attestations = &t.VDBSlotVizPassiveDuty{}
				for validator := range attAssignmentsForSlot[slot] {
					if slot >= latestSlot {
						// If the latest slot is the one that must be attested we still show it as pending
						// as the attestation cannot yet have been included in a block
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.PendingCount++
					} else if _, ok := slotAttested[slot][validator]; ok {
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.SuccessCount++
					} else {
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.FailedCount++
					}
				}
			}

			// Get the sync summary for the slot
			if len(syncAssignmentsForEpoch[epoch]) > 0 {
				slotVizEpochs[epochIdx].Slots[slotIdx].Sync = &t.VDBSlotVizPassiveDuty{}
				for validator := range syncAssignmentsForEpoch[epoch] {
					if slot > latestSlot {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.PendingCount++
					} else if _, ok := slotSyncParticipated[slot][validator]; ok {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.SuccessCount++
					} else {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.FailedCount++
					}
				}
			}

			// Get the slashings for the slot
			slashedValidators := []uint64{}
			if _, ok := slotValiPropSlashed[slot]; ok {
				slashedPropValidators := []uint64{}
				err := d.readerDb.Select(&slashedPropValidators, `
					SELECT
						proposerindex
					FROM blocks_proposerslashings
					WHERE block_slot = $1`, slot)
				if err != nil {
					return nil, err
				}

				slashedValidators = append(slashedValidators, slashedPropValidators...)
			}
			if _, ok := slotValiAttSlashed[slot]; ok {
				slashedAttValidators := []pq.Int64Array{}

				err := d.readerDb.Select(&slashedAttValidators, `
					SELECT
						attestation2_indices
					FROM blocks_attesterslashings
					WHERE block_slot = $1`, slot)
				if err != nil {
					return nil, err
				}

				for _, slashedAttValidator := range slashedAttValidators {
					for _, validator := range slashedAttValidator {
						slashedValidators = append(slashedValidators, uint64(validator))
					}
				}
			}
			if _, ok := propAssignmentsForSlot[slot]; ok {
				// One of the dashboard validators slashed
				for _, validator := range slashedValidators {
					slotVizEpochs[epochIdx].Slots[slotIdx].Slashing = append(slotVizEpochs[epochIdx].Slots[slotIdx].Slashing,
						t.VDBSlotVizActiveDuty{
							Status:     "success",
							Validator:  propAssignmentsForSlot[slot], // Dashboard validator
							DutyObject: validator,                    // Validator that got slashed
						})
				}
			}
			for _, validator := range slashedValidators {
				if _, ok := validatorsMap[validator]; !ok {
					continue
				}
				// One of the dashboard validators got slashed
				slotVizEpochs[epochIdx].Slots[slotIdx].Slashing = append(slotVizEpochs[epochIdx].Slots[slotIdx].Slashing,
					t.VDBSlotVizActiveDuty{
						Status:     "failed",
						Validator:  validator, // Dashboard validator
						DutyObject: validator, // Validator that got slashed
					})
			}
		}
	}

	return slotVizEpochs, nil
}

func (d DataAccessService) GetValidatorDashboardSlotVizByPublicId(dashboardId t.VDBIdPublic) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotVizByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotVizByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardSummaryByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBIdPrimary, groupId uint64) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardGroupSummaryByPublicId(dashboardId t.VDBIdPublic, groupId uint64) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummaryByPublicId(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummaryByValidators(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummaryChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChartByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummaryChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChartByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardRewardsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupRewardsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewardsByPublicId(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewardsByValidators(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewardsChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChartByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewardsChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChartByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardDutiesByPublicId(dashboardId t.VDBIdPublic, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDutiesByPublicId(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDutiesByValidators(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocksByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocksByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocksByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBIdPrimary) (t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardHeatmapByPublicId(dashboardId t.VDBIdPublic) (t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmapByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardHeatmapByValidators(dashboardId t.VDBIdValidatorSet) (t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmapByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmapByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmapByPublicId(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmapByValidators(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardElDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardElDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDepositsByPublicId(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardElDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDepositsByValidators(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDepositsByPublicId(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDepositsByValidators(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawalsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawalsByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawalsByValidators(dashboardId, cursor, sort, search, limit)
}
