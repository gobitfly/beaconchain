package dataaccess

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
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

	ReaderDb                *sqlx.DB
	WriterDb                *sqlx.DB
	Bigtable                *db.Bigtable
	PersistentRedisDbClient *redis.Client
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
	log.Infof("retrieving data for dashboard with id %d", dashboardId)

	// TODO: Get the validators from the dashboardId
	validatorsMap := make(map[uint64]bool)
	for i := 900000; i < 900000+100; i++ {
		validatorsMap[uint64(i)] = true
	}
	validatorsMap[381428] = true

	// Get min/max slot/epoch
	headEpoch := cache.LatestEpoch.Get() // Reminder: Currently it is possible to get the head epoch from the cache but nothing sets it in v2
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2
	maxEpoch := headEpoch + 1

	dutiesInfo, releaseLock, err := services.GetCurrentDutiesInfo()
	defer releaseLock() // important to unlock once done, otherwise data updater cant update the data
	if err != nil {
		return nil, err
	}

	// Restructure proposal status, attestations, sync duties and slashings
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
				if _, ok := validatorsMap[proposerIndex]; ok {
					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal = &t.VDBSlotVizActiveDuty{}

					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.Validator = dutiesInfo.PropAssignmentsForSlot[slot]

					status := "scheduled"
					dutyObject := slot
					if _, ok := dutiesInfo.SlotStatus[slot]; ok {
						switch dutiesInfo.SlotStatus[slot] {
						case 0, 2:
							status = "failed"
						case 1, 3:
							status = "success"
							dutyObject = dutiesInfo.SlotBlock[slot]
						}
					}
					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.Status = status
					slotVizEpochs[epochIdx].Slots[slotIdx].Proposal.DutyObject = dutyObject
				}
			}

			// Get the attestation summary for the slot
			if len(dutiesInfo.AttAssignmentsForSlot[slot]) > 0 {
				slotVizEpochs[epochIdx].Slots[slotIdx].Attestations = &t.VDBSlotVizPassiveDuty{}
				for validator := range dutiesInfo.AttAssignmentsForSlot[slot] {
					// only validators we care about
					if _, ok := validatorsMap[validator]; !ok {
						continue
					}
					if slot >= dutiesInfo.LatestSlot {
						// If the latest slot is the one that must be attested we still show it as pending
						// as the attestation cannot yet have been included in a block
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.PendingCount++
					} else if _, ok := dutiesInfo.SlotAttested[slot][validator]; ok {
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.SuccessCount++
					} else {
						slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.FailedCount++
					}
				}
			}

			// Get the sync summary for the slot
			if len(dutiesInfo.SyncAssignmentsForEpoch[epoch]) > 0 {
				slotVizEpochs[epochIdx].Slots[slotIdx].Sync = &t.VDBSlotVizPassiveDuty{}
				for validator := range dutiesInfo.SyncAssignmentsForEpoch[epoch] {
					// only validators we care about
					if _, ok := validatorsMap[validator]; !ok {
						continue
					}
					if slot > dutiesInfo.LatestSlot {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.PendingCount++
					} else if _, ok := dutiesInfo.SlotSyncParticipated[slot][validator]; ok {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.SuccessCount++
					} else {
						slotVizEpochs[epochIdx].Slots[slotIdx].Sync.FailedCount++
					}
				}
			}

			// Get the slashings for the slot
			slashedValidators := dutiesInfo.SlotValiPropSlashed[slot]
			slashedValidators = append(slashedValidators, dutiesInfo.SlotValiAttSlashed[slot]...)

			if proposerIndex, ok := dutiesInfo.PropAssignmentsForSlot[slot]; ok {
				// only add if we care about this validator
				if _, ok := validatorsMap[proposerIndex]; ok {
					// One of the dashboard validators slashed
					for _, validator := range slashedValidators {
						slotVizEpochs[epochIdx].Slots[slotIdx].Slashing = append(slotVizEpochs[epochIdx].Slots[slotIdx].Slashing,
							t.VDBSlotVizActiveDuty{
								Status:     "success",
								Validator:  dutiesInfo.PropAssignmentsForSlot[slot], // Dashboard validator
								DutyObject: validator,                               // Validator that got slashed
							})
					}
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
