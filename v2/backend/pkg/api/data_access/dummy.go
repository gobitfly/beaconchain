package dataaccess

import (
	"context"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/shopspring/decimal"
)

type DummyService struct {
}

// ensure DummyService pointer implements DataAccessor
var _ DataAccessor = (*DummyService)(nil)

func NewDummyService() *DummyService {
	return &DummyService{}
}

// must pass a pointer to the data
func commonFakeData(a interface{}) error {
	// TODO fake decimal.Decimal
	return faker.FakeData(a, options.WithRandomMapAndSliceMaxSize(5))
}

func (d *DummyService) Close() {
	// nothing to close
}

func (d *DummyService) GetLatestSlot() (uint64, error) {
	r := uint64(0)
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	r := []t.EthConversionRate{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetUserInfo(userId uint64) (*t.UserInfo, error) {
	r := t.UserInfo{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetUser(email string) (*t.User, error) {
	r := t.User{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetProductSummary() (*t.ProductSummary, error) {
	r := t.ProductSummary{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardInfo(dashboardId t.VDBIdPrimary) (*t.DashboardInfo, error) {
	r := t.DashboardInfo{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardInfoByPublicId(publicDashboardId t.VDBIdPublic) (*t.DashboardInfo, error) {
	r := t.DashboardInfo{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardName(dashboardId t.VDBIdPrimary) (string, error) {
	r := ""
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetValidatorsFromSlices(indices []uint64, publicKeys []string) ([]t.VDBValidator, error) {
	r := []t.VDBValidator{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetUserDashboards(userId uint64) (*t.UserDashboardsData, error) {
	r := t.UserDashboardsData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) CreateValidatorDashboard(userId uint64, name string, network uint64) (*t.VDBPostReturnData, error) {
	r := t.VDBPostReturnData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardOverview(dashboardId t.VDBId) (*t.VDBOverviewData, error) {
	r := t.VDBOverviewData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) RemoveValidatorDashboard(dashboardId t.VDBIdPrimary) error {
	return nil
}

func (d *DummyService) UpdateValidatorDashboardName(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostReturnData, error) {
	r := t.VDBPostReturnData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) CreateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, name string) (*t.VDBPostCreateGroupData, error) {
	r := t.VDBPostCreateGroupData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) UpdateValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64, name string) (*t.VDBPostCreateGroupData, error) {
	r := t.VDBPostCreateGroupData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) RemoveValidatorDashboardGroup(dashboardId t.VDBIdPrimary, groupId uint64) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardGroupExists(dashboardId t.VDBIdPrimary, groupId uint64) (bool, error) {
	return true, nil
}

func (d *DummyService) AddValidatorDashboardValidators(dashboardId t.VDBIdPrimary, groupId int64, validators []t.VDBValidator) ([]t.VDBPostValidatorsData, error) {
	r := []t.VDBPostValidatorsData{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetValidatorDashboardValidators(dashboardId t.VDBId, groupId int64, cursor string, colSort t.Sort[enums.VDBManageValidatorsColumn], search string, limit uint64) ([]t.VDBManageValidatorsTableRow, *t.Paging, error) {
	r := []t.VDBManageValidatorsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) RemoveValidatorDashboardValidators(dashboardId t.VDBIdPrimary, validators []t.VDBValidator) error {
	return nil
}

func (d *DummyService) CreateValidatorDashboardPublicId(dashboardId t.VDBIdPrimary, name string, shareGroups bool) (*t.VDBPublicId, error) {
	r := t.VDBPublicId{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) (*t.VDBPublicId, error) {
	r := t.VDBPublicId{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) UpdateValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic, name string, shareGroups bool) (*t.VDBPublicId, error) {
	r := t.VDBPublicId{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) RemoveValidatorDashboardPublicId(publicDashboardId t.VDBIdPublic) error {
	return nil
}

func (d *DummyService) GetValidatorDashboardSlotViz(dashboardId t.VDBId) ([]t.SlotVizEpoch, error) {
	r := struct {
		Epochs []t.SlotVizEpoch `faker:"slice_len=4"`
	}{}
	err := commonFakeData(&r)
	return r.Epochs, err
}

func (d *DummyService) GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	r := []t.VDBSummaryTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}
func (d *DummyService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error) {
	r := t.VDBGroupSummaryData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error) {
	r := t.ChartData[int, float64]{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardValidatorIndices(dashboardId t.VDBId, groupId int64, duty enums.ValidatorDuty, period enums.TimePeriod) ([]uint64, error) {
	r := []uint64{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	r := []t.VDBRewardsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	r := t.VDBGroupRewardsData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	r := t.ChartData[int, decimal.Decimal]{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	r := []t.VDBEpochDutiesTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	r := []t.VDBBlocksTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardEpochHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	r := t.VDBHeatmap{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardDailyHeatmap(dashboardId t.VDBId, period enums.TimePeriod) (*t.VDBHeatmap, error) {
	r := t.VDBHeatmap{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardGroupEpochHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	r := t.VDBHeatmapTooltipData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardGroupDailyHeatmap(dashboardId t.VDBId, groupId uint64, day time.Time) (*t.VDBHeatmapTooltipData, error) {
	r := t.VDBHeatmapTooltipData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	r := []t.VDBExecutionDepositsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	r := []t.VDBConsensusDepositsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardTotalElDeposits(dashboardId t.VDBId) (*t.VDBTotalExecutionDepositsData, error) {
	r := t.VDBTotalExecutionDepositsData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardTotalClDeposits(dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error) {
	r := t.VDBTotalConsensusDepositsData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	r := []t.VDBWithdrawalsTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, &p, err
}

func (d *DummyService) GetValidatorDashboardTotalWithdrawals(dashboardId t.VDBId, search string) (*t.VDBTotalWithdrawalsData, error) {
	r := t.VDBTotalWithdrawalsData{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetAllNetworks() ([]t.NetworkInfo, error) {
	r := []t.NetworkInfo{}
	err := commonFakeData(&r)
	return r, err
}

func (d *DummyService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	r := t.SearchValidator{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	r := t.SearchValidator{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	r := t.SearchValidatorsByDepositAddress{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error) {
	r := t.SearchValidatorsByDepositEnsName{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error) {
	r := t.SearchValidatorsByWithdrwalCredential{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error) {
	r := t.SearchValidatorsByWithrawalEnsName{}
	err := commonFakeData(&r)
	return &r, err
}

func (d *DummyService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	r := t.SearchValidatorsByGraffiti{}
	err := commonFakeData(&r)
	return &r, err
}
