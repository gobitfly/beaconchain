package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/types/api"

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.DashboardData, error)

	CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostData, error)
	GetValidatorDashboardOverview(userId uint64, dashboardId uint64) (t.VDBOverviewData, error)

	GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummary, error)

	GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)
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

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardOverview(userId uint64, dashboardId uint64) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverview(userId, dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummary, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}
