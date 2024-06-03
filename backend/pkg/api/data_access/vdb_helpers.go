package dataaccess

import (
	"database/sql"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/shopspring/decimal"
)

type ValidatorDashboardRepository interface {
	GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error)
	GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error)
	GetValidatorDashboardName(dashboardId t.VDBIdPrimary) (string, error)
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

	CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, showGroupNames bool) (*t.VDBPublicId, error)
	GetValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error)
	UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, showGroupNames bool) (*t.VDBPublicId, error)
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
