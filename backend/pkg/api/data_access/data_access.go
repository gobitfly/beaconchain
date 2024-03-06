package dataaccess

import (
	"context"
	"fmt"
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
)

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.UserDashboardsData, error)
	GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error)

	CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error)
	RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error

	GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (t.DashboardInfo, error)
	GetValidatorDashboardInfoByPublicId(dashboardId t.VDBIdPublic) (t.DashboardInfo, error)

	GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (t.VDBOverviewData, error)
	GetValidatorDashboardOverviewByPublicId(publicDashboardId t.VDBIdPublic) (t.VDBOverviewData, error)
	GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (t.VDBOverviewData, error)

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (t.VDBOverviewGroup, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error

	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error)
	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error
	GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)
	GetValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)
	GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error)

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error

	GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error)
	GetValidatorDashboardSlotVizByPublicId(dashboardId t.VDBIdPublic) ([]t.SlotVizEpoch, error)
	GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardSummaryByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBIdPrimary, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardGroupSummaryByPublicId(dashboardId t.VDBIdPublic, groupId uint64) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet) (t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error)
	GetValidatorDashboardSummaryChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error)
	GetValidatorDashboardSummaryChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error)

	GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardRewardsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardGroupRewardsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBIdPrimary) (t.ChartData[int], error)
	GetValidatorDashboardRewardsChartByPublicId(dashboardId t.VDBIdPublic) (t.ChartData[int], error)
	GetValidatorDashboardRewardsChartByValidators(dashboardId t.VDBIdValidatorSet) (t.ChartData[int], error)

	GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)
	GetValidatorDashboardDutiesByPublicId(dashboardId t.VDBIdPublic, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)
	GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)
	GetValidatorDashboardBlocksByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)
	GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId t.VDBIdPrimary) (t.VDBHeatmap, error)
	GetValidatorDashboardHeatmapByPublicId(dashboardId t.VDBIdPublic) (t.VDBHeatmap, error)
	GetValidatorDashboardHeatmapByValidators(dashboardId t.VDBIdValidatorSet) (t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupHeatmapByPublicId(dashboardId t.VDBIdPublic, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardElDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardElDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDepositsByPublicId(dashboardId t.VDBIdPublic, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardClDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawalsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)
	GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error)

	CloseDataAccessService()
}

type DataAccessService struct {
	dummy DummyService

	ReaderDb                *sqlx.DB
	WriterDb                *sqlx.DB
	Bigtable                *db.Bigtable
	PersistentRedisDbClient *redis.Client
}

// ensure DataAccessService implements DataAccessInterface
var _ DataAccessInterface = DataAccessService{}

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

func (d DataAccessService) GetUserDashboards(userId uint64) (t.UserDashboardsData, error) {
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

// param validators: slice of validator public keys or indices, a index should resolve to the newest index version
func (d DataAccessService) GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error) {
	// TODO @recy21
	return d.dummy.GetValidatorsFromStrings(validators)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error) {
	// WORKING spletka
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (t.DashboardInfo, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardInfo(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardInfoByPublicId(dashboardId t.VDBIdPublic) (t.DashboardInfo, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardInfoByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (t.VDBOverviewData, error) {
	// WORKING Rami

	// TODO: Get the validators from the dashboardId
	validators := make([]uint64, 100)
	for i := uint64(0); i < uint64(len(validators)); i++ {
		validators[i] = i
	}

	data := t.VDBOverviewData{}
	// get data from DB
	// data.Groups =
	qry := `SELECT
		status AS statename, COUNT(*) AS statecount
	FROM
		validators
	WHERE
		validatorindex = ANY($1)
	GROUP BY
		status`
	var currentStateCounts []*struct {
		Name  string `db:"statename"`
		Count uint64 `db:"statecount"`
	}
	err := db.ReaderDb.Select(&currentStateCounts, qry, validators)
	if err != nil {
		// utils.Log(err, "error retrieving validators data", 0)
		return t.VDBOverviewData{}, err
	}
	for _, state := range currentStateCounts {
		// count exited, pending, slashed?
		data.Validators.Total += state.Count
		switch state.Name {
		// exiting_online, exiting_offline?
		case "active_online":
			fallthrough
		case "active_offline":
			data.Validators.Active += state.Count
		case "pending":
			data.Validators.Pending += state.Count
		case "slashed":
			fallthrough
		case "exited":
			data.Validators.Exited += state.Count
		case "slashing_online":
			fallthrough
		case "slashing_offline":
			data.Validators.Slashed += state.Count
		}
	}
	income := types.ValidatorIncomePerformance{}
	err = db.GetValidatorIncomePerformance(validators, &income)
	if err != nil {
		return t.VDBOverviewData{}, err
	}
	data.Rewards.Day.El = income.ElIncomeWei1d
	data.Rewards.Day.Cl = income.ClIncomeWei1d
	data.Rewards.Week.El = income.ElIncomeWei7d
	data.Rewards.Week.Cl = income.ClIncomeWei7d
	data.Rewards.Month.El = income.ElIncomeWei31d
	data.Rewards.Month.Cl = income.ClIncomeWei31d
	data.Rewards.Year.El = income.ElIncomeWei365d
	data.Rewards.Year.Cl = income.ClIncomeWei365d
	data.Rewards.Total.El = income.ElIncomeWeiTotal
	data.Rewards.Total.Cl = income.ClIncomeWeiTotal

	//data.Luck =
	//data.Apr =

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
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboard(dashboardId)
}

func (d DataAccessService) RemoveValidatorDashboardByPublicId(dashboardId t.VDBIdPublic) error {
	// TODO @recy21
	return d.dummy.RemoveValidatorDashboardByPublicId(dashboardId)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (t.VDBOverviewGroup, error) {
	// WORKING spletka
	return d.dummy.CreateValidatorDashboardGroup(dashboardId, name)
}

func (d DataAccessService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboardGroup(dashboardId, groupId)
}

func (d DataAccessService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	// WORKING spletka
	return d.dummy.AddValidatorDashboardValidators(dashboardId, groupId, validators)
}

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByPublicId(dashboardId t.VDBIdPublic, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByPublicId(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboardValidators(dashboardId, validators)
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// WORKING spletka
	return d.dummy.CreateValidatorDashboardPublicId(dashboardId, name, showGroupNames)
}

func (d DataAccessService) UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	// WORKING spletka
	return d.dummy.UpdateValidatorDashboardPublicId(dashboardId, publicDashboardId, name, showGroupNames)
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboardPublicId(dashboardId, publicDashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error) {
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

func (d DataAccessService) GetValidatorDashboardSlotVizByPublicId(dashboardId t.VDBIdPublic) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotVizByPublicId(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotVizByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardSummaryByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
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

func (d DataAccessService) GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummaryByValidators(dashboardId)
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

func (d DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardRewardsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
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

func (d DataAccessService) GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewardsByValidators(dashboardId, epoch)
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

func (d DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardDutiesByPublicId(dashboardId t.VDBIdPublic, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDutiesByPublicId(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDutiesByValidators(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocksByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocksByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
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

func (d DataAccessService) GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmapByValidators(dashboardId, epoch)
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

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawalsByPublicId(dashboardId t.VDBIdPublic, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawalsByPublicId(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawalsByValidators(dashboardId, cursor, sort, search, limit)
}
