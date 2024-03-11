package dataaccess

import (
	"context"
	"fmt"
	isort "sort"
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

	GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (*t.VDBOverviewData, error)
	GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (*t.VDBOverviewData, error)

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBOverviewGroup, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error

	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error)
	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error
	GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error)
	GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error)

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error)
	RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error

	GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error)
	GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error)
	GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBIdPrimary, groupId uint64) (*t.VDBGroupSummaryData, error)
	GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet) (*t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBIdPrimary) (*t.ChartData[int], error)
	GetValidatorDashboardSummaryChartByValidators(dashboardId t.VDBIdValidatorSet) (*t.ChartData[int], error)

	GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error)
	GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.VDBGroupRewardsData, error)
	GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (*t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBIdPrimary) (*t.ChartData[int], error)
	GetValidatorDashboardRewardsChartByValidators(dashboardId t.VDBIdValidatorSet) (*t.ChartData[int], error)

	GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error)
	GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error)
	GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error)

	GetValidatorDashboardHeatmap(dashboardId t.VDBIdPrimary) (*t.VDBHeatmap, error)
	GetValidatorDashboardHeatmapByValidators(dashboardId t.VDBIdValidatorSet) (*t.VDBHeatmap, error)
	GetValidatorDashboardGroupHeatmap(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (*t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardElDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardClDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error)
	GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error)

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
	// WORKING spletka
	return d.dummy.GetValidatorDashboardInfo(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	// WORKING spletka
	return d.dummy.GetValidatorDashboardInfoByPublicId(publicDashboardId)
}

// param validators: slice of validator public keys or indices, a index should resolve to the newest index version
func (d DataAccessService) GetValidatorsFromStrings(validators []string) ([]t.VDBValidator, error) {
	// TODO @recy21
	return d.dummy.GetValidatorsFromStrings(validators)
}

func (d DataAccessService) GetUserDashboards(userId uint64) (*t.UserDashboardsData, error) {
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	// WORKING spletka
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardOverview(dashboardId t.VDBIdPrimary) (*t.VDBOverviewData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardOverview(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardOverviewByValidators(validators t.VDBIdValidatorSet) (*t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverviewByValidators(validators)
}

func (d DataAccessService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboard(dashboardId)
}

func (d DataAccessService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBOverviewGroup, error) {
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

func (d DataAccessService) GetValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidators(dashboardId, groupId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardValidatorsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardValidatorsByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboardValidators(dashboardId, validators)
}

func (d DataAccessService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	// WORKING spletka
	return d.dummy.CreateValidatorDashboardPublicId(dashboardId, name, showGroupNames)
}

func (d DataAccessService) UpdateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string, name string, showGroupNames bool) (*t.VDBPostPublicIdData, error) {
	// WORKING spletka
	return d.dummy.UpdateValidatorDashboardPublicId(dashboardId, publicDashboardId, name, showGroupNames)
}

func (d DataAccessService) RemoveValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, publicDashboardId string) error {
	// WORKING spletka
	return d.dummy.RemoveValidatorDashboardPublicId(dashboardId, publicDashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId t.VDBIdPrimary) ([]t.SlotVizEpoch, error) {
	log.Infof("retrieving data for dashboard with id %d", dashboardId)

	var validatorsArray []uint32
	err := db.AlloyReader.Select(&validatorsArray, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY validator_index`, dashboardId)
	if err != nil {
		return nil, err
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

func (d DataAccessService) GetValidatorDashboardSlotVizByValidators(dashboardId t.VDBIdValidatorSet) ([]t.SlotVizEpoch, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSlotVizByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	// TODO: implement sorting, filtering & paging
	ret := make(map[uint64]*t.VDBSummaryTableRow) // map of group id to result row
	retMux := &sync.Mutex{}

	// retrieve efficiency data for each time period, we cannot do sorting & filtering here as we need access to the whole set
	wg := errgroup.Group{}

	log.Infof("GetValidatorDashboardSummary called for dashboard %d", dashboardId)
	query := `select group_id, ((0.84375 * attestation_efficiency) + (0.125 * proposer_efficiency) + (0.03125 * sync_efficiency)) * 100.0 AS efficiency FROM (
		select 
			group_id,
			sum(attestations_reward)::decimal / sum(attestations_ideal_reward)::decimal AS attestation_efficiency,
			COALESCE(SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0), 1) AS proposer_efficiency,
			COALESCE(SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0), 1) AS sync_efficiency
			from users_val_dashboards_validators 
		left join %[1]s on %[1]s.validatorindex = users_val_dashboards_validators.validator_index
		where dashboard_id = $1
		group by 1
	) as a;`

	type queryResult struct {
		GroupId    uint64  `db:"group_id"`
		Efficiency float64 `db:"efficiency"`
	}

	wg.Go(func() error {
		type validatorsPerGroup struct {
			GroupId        uint64 `db:"group_id"`
			ValidatorIndex uint64 `db:"validator_index"`
		}

		var queryResult []validatorsPerGroup

		err := db.AlloyReader.Select(&queryResult, `SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY group_id, validator_index`, dashboardId)
		if err != nil {
			return fmt.Errorf("error retrieving validator groups for dashboard: %v", err)
		}

		retMux.Lock()
		for _, result := range queryResult {
			if ret[result.GroupId] == nil {
				ret[result.GroupId] = &t.VDBSummaryTableRow{}
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

	wg.Go(func() error {
		var queryResult []queryResult
		err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, "validator_dashboard_data_rolling_daily"), dashboardId)
		if err != nil {
			return fmt.Errorf("error retrieving data for last 24h: %v", err)
		}

		retMux.Lock()
		for _, result := range queryResult {
			if result.Efficiency < 0 {
				result.Efficiency = 0
			}

			if ret[result.GroupId] == nil {
				ret[result.GroupId] = &t.VDBSummaryTableRow{}
			}
			ret[result.GroupId].GroupId = result.GroupId
			ret[result.GroupId].EfficiencyDay = result.Efficiency
		}
		retMux.Unlock()
		return nil
	})
	wg.Go(func() error {
		var queryResult []queryResult
		err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, "validator_dashboard_data_rolling_weekly"), dashboardId)
		if err != nil {
			return fmt.Errorf("error retrieving data for last 24h: %v", err)
		}

		retMux.Lock()
		for _, result := range queryResult {
			if result.Efficiency < 0 {
				result.Efficiency = 0
			}
			if ret[result.GroupId] == nil {
				ret[result.GroupId] = &t.VDBSummaryTableRow{}
			}
			ret[result.GroupId].GroupId = result.GroupId
			ret[result.GroupId].EfficiencyWeek = result.Efficiency
		}
		retMux.Unlock()
		return nil
	})
	wg.Go(func() error {
		var queryResult []queryResult
		err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, "validator_dashboard_data_rolling_monthly"), dashboardId)
		if err != nil {
			return fmt.Errorf("error retrieving data for last 24h: %v", err)
		}

		retMux.Lock()
		for _, result := range queryResult {
			if result.Efficiency < 0 {
				result.Efficiency = 0
			}
			if ret[result.GroupId] == nil {
				ret[result.GroupId] = &t.VDBSummaryTableRow{}
			}
			ret[result.GroupId].GroupId = result.GroupId
			ret[result.GroupId].EfficiencyMonth = result.Efficiency
		}
		retMux.Unlock()
		return nil
	})
	wg.Go(func() error {
		var queryResult []queryResult
		err := db.AlloyReader.Select(&queryResult, fmt.Sprintf(query, "validator_dashboard_data_rolling_total"), dashboardId)
		if err != nil {
			return fmt.Errorf("error retrieving data for last 24h: %v", err)
		}

		retMux.Lock()
		for _, result := range queryResult {
			if result.Efficiency < 0 {
				result.Efficiency = 0
			}
			if ret[result.GroupId] == nil {
				ret[result.GroupId] = &t.VDBSummaryTableRow{}
			}
			ret[result.GroupId].GroupId = result.GroupId
			ret[result.GroupId].EfficiencyTotal = result.Efficiency
		}
		retMux.Unlock()
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

func (d DataAccessService) GetValidatorDashboardSummaryByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBIdPrimary, groupId uint64) (*t.VDBGroupSummaryData, error) {

	ret := t.VDBGroupSummaryData{}
	wg := errgroup.Group{}

	log.Infof("GetValidatorDashboardGroupSummary called for dashboard %d with group id %v", dashboardId, groupId)
	query := `select
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
			withdrawals_amount
			from users_val_dashboards_validators
		left join %[1]s on %[1]s.validatorindex = users_val_dashboards_validators.validator_index
		where dashboard_id = $1 and group_id = $2
	`

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

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  uint64          `db:"blocks_cl_reward"`
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`
		SyncRewards   uint64 `db:"sync_rewards"`

		Slashed bool `db:"slashed"`

		BalanceStart uint64 `db:"balance_start"`
		BalanceEnd   uint64 `db:"balance_end"`

		DepositsCount  uint32 `db:"deposits_count"`
		DepositsAmount uint64 `db:"deposits_amount"`

		WithdrawalsCount  uint32 `db:"withdrawals_count"`
		WithdrawalsAmount uint64 `db:"withdrawals_amount"`
	}

	wg.Go(func() error {
		var rows []*queryResult
		err := db.AlloyReader.Select(&rows, fmt.Sprintf(query, "validator_dashboard_data_rolling_daily"), dashboardId, groupId)

		if err != nil {
			return err
		}

		totalAttestationRewards := int64(0)
		totalIdealAttestationRewards := int64(0)
		for _, row := range rows {
			totalAttestationRewards += row.AttestationReward
			totalIdealAttestationRewards += row.AttestationIdealReward

			ret.DetailsDay.Proposals.StatusCount.Success += uint64(row.BlocksProposed)
			ret.DetailsDay.Proposals.StatusCount.Failed += uint64(row.BlocksScheduled) - uint64(row.BlocksProposed)

			ret.DetailsDay.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
			ret.DetailsDay.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

			if row.Slashed {
				ret.DetailsDay.Slashed.StatusCount.Failed++
			} else {
				ret.DetailsDay.Slashed.StatusCount.Success++
			}
		}

		ret.DetailsDay.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
		if ret.DetailsDay.AttestationEfficiency < 0 {
			ret.DetailsDay.AttestationEfficiency = 0
		}
		return nil
	})
	err := wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}

	return &ret, nil
}

func (d DataAccessService) GetValidatorDashboardGroupSummaryByValidators(dashboardId t.VDBIdValidatorSet) (*t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummaryByValidators(dashboardId)
}

// for summary charts: series id is group id, no stack

func (d DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBIdPrimary) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummaryChartByValidators(dashboardId t.VDBIdValidatorSet) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummaryChartByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardRewardsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupRewardsByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupRewardsByValidators(dashboardId, epoch)
}

func (d DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBIdPrimary) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardRewardsChartByValidators(dashboardId t.VDBIdValidatorSet) (*t.ChartData[int], error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardRewardsChartByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBIdPrimary, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardDutiesByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64, cursor string, sort []t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDutiesByValidators(dashboardId, epoch, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardBlocksByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocksByValidators(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBIdPrimary) (*t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardHeatmapByValidators(dashboardId t.VDBIdValidatorSet) (*t.VDBHeatmap, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardHeatmapByValidators(dashboardId)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBIdPrimary, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}

func (d DataAccessService) GetValidatorDashboardGroupHeatmapByValidators(dashboardId t.VDBIdValidatorSet, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupHeatmapByValidators(dashboardId, epoch)
}

func (d DataAccessService) GetValidatorDashboardElDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardElDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardElDepositsByValidators(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBIdPrimary, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardClDeposits(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardClDepositsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardClDepositsByValidators(dashboardId, cursor, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBIdPrimary, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawals(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardWithdrawalsByValidators(dashboardId t.VDBIdValidatorSet, cursor string, sort []t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardWithdrawalsByValidators(dashboardId, cursor, sort, search, limit)
}
