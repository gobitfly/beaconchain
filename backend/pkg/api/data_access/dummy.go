package dataaccess

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type DummyService struct {
}

func NewDummyService() DummyService {
	return DummyService{}
}

// must pass a pointer to the data
func commonFakeData(a interface{}) error {
	// TODO fake decimal.Decimal
	return faker.FakeData(a, options.WithRandomMapAndSliceMaxSize(5))
}

func (d DummyService) CloseDataAccessService() {
	// nothing to close
}

func (d DummyService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	r := t.DashboardData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error) {
	r := t.VDBPostReturnData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardOverview(dashboardId uint64) (t.VDBOverviewData, error) {
	r := t.VDBOverviewData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) RemoveValidatorDashboardOverview(dashboardId uint64) error {
	return nil
}

func (d DummyService) CreateValidatorDashboardGroup(dashboardId uint64, name string) (t.VDBOverviewGroup, error) {
	r := t.VDBOverviewGroup{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) RemoveValidatorDashboardGroup(dashboardId uint64, groupId uint64) error {
	return nil
}

func (d DummyService) AddValidatorDashboardValidators(dashboardId uint64, groupId uint64, validators []string) ([]t.VDBPostValidatorsData, error) {
	r := []t.VDBPostValidatorsData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardValidators(dashboardId uint64, groupId uint64, cursor string, sort []t.Sort[t.VDBValidatorsColumn], search string, limit uint64) ([]t.VDBGetValidatorsData, error) {
	r := []t.VDBGetValidatorsData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) RemoveValidatorDashboardValidators(dashboardId uint64, validators []string) error {
	return nil
}

func (d DummyService) CreateValidatorDashboardPublicId(dashboardId uint64, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	r := t.VDBPostPublicIdData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) UpdateValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string, name string, showGroupNames bool) (t.VDBPostPublicIdData, error) {
	r := t.VDBPostPublicIdData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) RemoveValidatorDashboardPublicId(dashboardId uint64, publicDashboardId string) error {
	return nil
}

func (d DummyService) GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.SlotVizEpoch, error) {
	r := []t.SlotVizEpoch{}
	var err error
	for i := 0; i < 4; i++ {
		epoch := t.SlotVizEpoch{}
		err = commonFakeData(&epoch)
		r = append(r, epoch)
	}
	return r, err
}

func (d DummyService) GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	r := []t.VDBSummaryTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummaryData, error) {
	r := t.VDBGroupSummaryData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardSummaryChart(dashboardId uint64) (t.ChartData[int], error) {
	r := t.ChartData[int]{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardRewards(dashboardId uint64, cursor string, sort []t.Sort[t.VDBRewardsTableColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, t.Paging, error) {
	r := []t.VDBRewardsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardGroupRewards(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBGroupRewardsData, error) {
	r := t.VDBGroupRewardsData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardRewardsChart(dashboardId uint64) (t.ChartData[int], error) {
	r := t.ChartData[int]{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardDuties(dashboardId uint64, epoch uint64, cursor string, sort []t.Sort[t.VDBDutiesTableColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, t.Paging, error) {
	r := []t.VDBEpochDutiesTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	r := []t.VDBBlocksTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardHeatmap(dashboardId uint64) (t.VDBHeatmap, error) {
	r := t.VDBHeatmap{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardGroupHeatmap(dashboardId uint64, groupId uint64, epoch uint64) (t.VDBHeatmapTooltipData, error) {
	r := t.VDBHeatmapTooltipData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardElDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, t.Paging, error) {
	r := []t.VDBExecutionDepositsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardClDeposits(dashboardId uint64, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, t.Paging, error) {
	r := []t.VDBConsensusDepositsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}

func (d DummyService) GetValidatorDashboardWithdrawals(dashboardId uint64, cursor string, sort []t.Sort[t.VDBWithdrawalsTableColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, t.Paging, error) {
	r := []t.VDBWithdrawalsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}
