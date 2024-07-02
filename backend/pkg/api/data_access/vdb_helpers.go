package dataaccess

import (
	"database/sql"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type ValidatorDashboardRepository interface {
	GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error)
	GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error)
	GetValidatorDashboardName(dashboardId t.VDBIdPrimary) (string, error)
	CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error)
	RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error

	UpdateValidatorDashboardName(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error)

	GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error)

	CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error)
	UpdateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error)
	RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error
	GetValidatorDashboardGroupCount(dashboardId t.VDBIdPrimary) (uint64, error)
	GetValidatorDashboardGroupExists(dashboardId t.VDBIdPrimary, groupId uint64) (bool, error)

	GetValidatorDashboardExistingValidatorCount(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) (uint64, error)
	AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId uint64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error)
	AddValidatorDashboardValidatorsByDepositAddress(dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error)
	AddValidatorDashboardValidatorsByWithdrawalAddress(dashboardId t.VDBIdPrimary, groupId uint64, address string, limit uint64) ([]t.VDBPostValidatorsData, error)
	AddValidatorDashboardValidatorsByGraffiti(dashboardId t.VDBIdPrimary, groupId uint64, graffiti string, limit uint64) ([]t.VDBPostValidatorsData, error)

	RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error
	GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error)
	GetValidatorDashboardValidatorsCount(dashboardId t.VDBIdPrimary) (uint64, error)

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error)
	GetValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error)
	UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error)
	RemoveValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) error
	GetValidatorDashboardPublicIdCount(dashboardId t.VDBIdPrimary) (uint64, error)

	GetValidatorDashboardSlotViz(dashboardId t.VDBId, groupIds []uint64) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBGroupSummaryData, error)
	GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error)
	GetValidatorDashboardSummaryValidators(dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error)
	GetValidatorDashboardSyncSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error)
	GetValidatorDashboardSlashingsSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error)
	GetValidatorDashboardProposalSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error)

	GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error)
	GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error)
	GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error)

	GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error)

	GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error)

	GetValidatorDashboardEpochHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error)
	GetValidatorDashboardDailyHeatmap(dashboardId t.VDBId, period enums.TimePeriod) (*t.VDBHeatmap, error)
	GetValidatorDashboardGroupEpochHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error)
	GetValidatorDashboardGroupDailyHeatmap(dashboardId t.VDBId, groupId uint64, date time.Time) (*t.VDBHeatmapTooltipData, error)

	GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error)
	GetValidatorDashboardTotalElDeposits(dashboardId t.VDBId) (*t.VDBTotalExecutionDepositsData, error)
	GetValidatorDashboardTotalClDeposits(dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error)

	GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error)
	GetValidatorDashboardTotalWithdrawals(dashboardId t.VDBId, search string) (*t.VDBTotalWithdrawalsData, error)
}

//////////////////// 		Helper functions (must be used by more than one VDB endpoint!)

func (d DataAccessService) getDashboardValidators(dashboardId t.VDBId) ([]t.VDBValidator, error) {
	if len(dashboardId.Validators) == 0 {
		var validatorsArray []t.VDBValidator
		err := d.alloyReader.Select(&validatorsArray, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY validator_index`, dashboardId.Id)
		return validatorsArray, err
	}
	return dashboardId.Validators, nil
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

func (d DataAccessService) getValidatorStatuses(validators []uint64) (map[uint64]enums.ValidatorStatus, error) {
	validatorStatuses := make(map[uint64]enums.ValidatorStatus, len(validators))

	// Get the current validator state
	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, err
	}

	// Get the validator duties to check the last fulfilled attestation
	dutiesInfo, releaseValDutiesLock, err := d.services.GetCurrentDutiesInfo()
	defer releaseValDutiesLock()
	if err != nil {
		return nil, err
	}

	// Set the threshold for "online" => "offline" to 2 epochs without attestation
	attestationThresholdSlot := uint64(0)
	twoEpochs := 2 * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if dutiesInfo.LatestSlot >= twoEpochs {
		attestationThresholdSlot = dutiesInfo.LatestSlot - twoEpochs
	}

	// Fill the data
	for _, validator := range validators {
		metadata := validatorMapping.ValidatorMetadata[validator]

		switch constypes.ValidatorStatus(metadata.Status) {
		case constypes.PendingInitialized:
			validatorStatuses[validator] = enums.ValidatorStatuses.Deposited
		case constypes.PendingQueued:
			validatorStatuses[validator] = enums.ValidatorStatuses.Pending
		case constypes.ActiveOngoing, constypes.ActiveExiting, constypes.ActiveSlashed:
			var lastAttestionSlot uint32
			for slot, attested := range dutiesInfo.EpochAttestationDuties[validator] {
				if attested && slot > lastAttestionSlot {
					lastAttestionSlot = slot
				}
			}
			if lastAttestionSlot < uint32(attestationThresholdSlot) {
				validatorStatuses[validator] = enums.ValidatorStatuses.Offline
			} else {
				validatorStatuses[validator] = enums.ValidatorStatuses.Online
			}
		case constypes.ExitedUnslashed, constypes.ExitedSlashed, constypes.WithdrawalPossible, constypes.WithdrawalDone:
			if metadata.Slashed {
				validatorStatuses[validator] = enums.ValidatorStatuses.Slashed
			} else {
				validatorStatuses[validator] = enums.ValidatorStatuses.Exited
			}
		}
	}

	return validatorStatuses, nil
}

func (d *DataAccessService) getWithdrawableCountFromCursor(validatorindex t.VDBValidator, cursor uint64) (uint64, error) {
	// the validators' balance will not be checked here as this is only a rough estimation
	// checking the balance for hundreds of thousands of validators is too expensive

	stats := cache.LatestStats.Get()
	if stats == nil || stats.ActiveValidatorCount == nil || stats.TotalValidatorCount == nil {
		return 0, errors.New("stats not available")
	}

	var maxValidatorIndex t.VDBValidator
	if *stats.TotalValidatorCount > 0 {
		maxValidatorIndex = *stats.TotalValidatorCount - 1
	}
	if maxValidatorIndex == 0 {
		return 0, nil
	}

	activeValidators := *stats.ActiveValidatorCount
	if activeValidators == 0 {
		activeValidators = maxValidatorIndex
	}

	if validatorindex > cursor {
		// if the validatorindex is after the cursor, simply return the number of validators between the cursor and the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (validatorindex - cursor) * activeValidators / maxValidatorIndex, nil
	} else if validatorindex < cursor {
		// if the validatorindex is before the cursor (wraparound case) return the number of validators between the cursor and the most recent validator plus the amount of validators from the validator 0 to the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (maxValidatorIndex - cursor + validatorindex) * activeValidators / maxValidatorIndex, nil
	} else {
		return 0, nil
	}
}

// GetTimeToNextWithdrawal calculates the time it takes for the validators next withdrawal to be processed.
func (d *DataAccessService) getTimeToNextWithdrawal(distance uint64) time.Time {
	minTimeToWithdrawal := time.Now().Add(time.Second * time.Duration((distance/utils.Config.Chain.ClConfig.MaxValidatorsPerWithdrawalSweep)*utils.Config.Chain.ClConfig.SecondsPerSlot))
	timeToWithdrawal := time.Now().Add(time.Second * time.Duration((float64(distance)/float64(utils.Config.Chain.ClConfig.MaxWithdrawalsPerPayload))*float64(utils.Config.Chain.ClConfig.SecondsPerSlot)))

	if timeToWithdrawal.Before(minTimeToWithdrawal) {
		return minTimeToWithdrawal
	}

	return timeToWithdrawal
}
