package dataaccess

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	t "github.com/gobitfly/beaconchain/pkg/types/api"
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

func (d DummyService) GetUserDashboards(userId uint64) ([]t.DashboardData, error) {
	r := []t.DashboardData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) CreateValidatorDashboard(userId uint64, name string, network t.Network) (t.VDBPostData, error) {
	r := t.VDBPostData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardOverview(userId uint64, dashboardId string) (t.VDBOverviewData, error) {
	r := t.VDBOverviewData{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardSummary(dashboardId string, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, error) {
	r := []t.VDBSummaryTableRow{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardGroupSummary(dashboardId string, groupId uint64) (t.VDBGroupSummary, error) {
	r := t.VDBGroupSummary{}
	err := commonFakeData(&r)
	return r, err
}

func (d DummyService) GetValidatorDashboardBlocks(dashboardId string, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, error) {
	r := []t.VDBBlocksTableRow{}
	err := commonFakeData(&r)
	return r, err
}
