package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/types/api"

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.DashboardData, error)

	CreateValidatorDashboard(userId uint64, name string, network t.Network) (t.VDBPostData, error)
	GetValidatorDashboardOverview(userId uint64, dashboardId string) (t.VDBOverviewData, error)

	GetValidatorDashboardSummary(dashboardId string, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, error)
	GetValidatorDashboardGroupSummary(dashboardId string, groupId uint64) (t.VDBGroupSummary, error)

	GetValidatorDashboardBlocks(dashboardId string, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, error)
}

type DataAccessService struct {
	dummy DummyService
	// TODO @recy21 add persistance, e.g. DB, cache, bigtable, etc.
}

// TODO @recy21 add persistance params, e.g. DB host, port, user, password, etc.
func NewDataAccessService() DataAccessService {
	// TODO @recy21 probably init persistance here
	return DataAccessService{dummy: NewDummyService()}
}

func (d DataAccessService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network t.Network) (t.VDBPostData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardOverview(userId uint64, dashboardId string) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverview(userId, dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId string, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId string, groupId uint64) (t.VDBGroupSummary, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId string, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}
