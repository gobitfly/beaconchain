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

func (d DummyService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	r := t.DashboardData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostData, error) {
	r := t.VDBPostData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardOverview(userId uint64, dashboardId uint64) (t.VDBOverviewData, error) {
	r := t.VDBOverviewData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.VDBSlotVizEpoch, error) {
	r := []t.VDBSlotVizEpoch{}
	err := commonFakeData(&r)
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

func (d DummyService) GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	r := []t.VDBBlocksTableRow{}
	p := t.Paging{}
	_ = commonFakeData(&r)
	err := commonFakeData(&p)
	return r, p, err
}
