package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/types/api"

type DataAccessInterface interface {
	//GetSummaryTablePage(cursor string, limit uint64, sorts []t.Sort[t.VDBSummaryTableColumn]) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetUserDashboards(userId uint64) (t.DashboardData, error)
}

type DataAccessService struct {
	Dummy DummyService
	// TODO add real data access, e.g. DB, cache, bigtable, etc.
}

// TODO add data access params, e.g. DB host, port, user, password, etc.
func NewDataAccessService() DataAccessService {
	return DataAccessService{Dummy: DummyService{}}
}

func (d DataAccessService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	return d.Dummy.GetUserDashboards(userId)
}
