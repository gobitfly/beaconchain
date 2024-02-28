package dataaccess

import (
	"context"
	"fmt"
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
)

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.DashboardData, error)

	CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error)
	GetValidatorDashboardOverview(dashboardId uint64) (t.VDBOverviewData, error)
	RemoveValidatorDashboardOverview(dashboardId uint64) error

	CreateValidatorDashboardGroup(dashboardId uint64, name string) (t.VDBOverviewGroup, error)
	RemoveValidatorDashboardGroup(dashboardId uint64, groupId uint64) error

	AddValidatorDashboardValidators(dashboardId uint64, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error)
	GetValidatorDashboardValidators(dashboardId uint64, groupId uint64, cursor string, sort []t.Sort[t.VDBValidatorsColumn], search string, limit uint64) ([]t.VDBGetValidatorsData, error)
	RemoveValidatorDashboardValidators(dashboardId uint64, validators []string) error

	CreateValidatorDashboardPublicId(dashboardId uint64, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string) error

	GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId uint64) (t.ChartData[int], error)

	GetValidatorDashboardRewards(dashboardId uint64, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId uint64) (t.ChartData[int], error)

	GetValidatorDashboardDuties(dashboardId uint64, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId uint64) (t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId uint64, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)

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

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId uint64) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) RemoveValidatorDashboardOverview(dashboardId uint64) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId uint64, name string) (t.VDBOverviewGroup, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardGroup(dashboardId, name)
}

func (d DataAccessService) RemoveValidatorDashboardGroup(dashboardId uint64, groupId uint64) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardGroup(dashboardId, groupId)
}

func (d DataAccessService) AddValidatorDashboardValidators(dashboardId uint64, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error) {
	// TODO @recy21
	return d.dummy.AddValidatorDashboardValidators(dashboardId, groupId, validators)
}

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId uint64, groupId uint64, cursor string, sort []t.Sort[t.VDBValidatorsColumn], search string, limit uint64) ([]t.VDBGetValidatorsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) RemoveValidatorDashboardValidators(dashboardId uint64, validators []string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardValidators(dashboardId, validators)
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId uint64, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboardPublicId(dashboardId, name, showGroupNames)
}

func (d DataAccessService) UpdateValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// TODO @recy21
	return d.dummy.UpdateValidatorDashboardPublicId(dashboardId, publicDashboardId, name, showGroupNames)
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardPublicId(dashboardId, publicDashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotViz(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardSummaryChart(dashboardId uint64) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewards(dashboardId uint64, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupRewards(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardRewardsChart(dashboardId uint64) (t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardDuties(dashboardId uint64, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardHeatmap(dashboardId uint64) (t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardElDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId uint64, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}
