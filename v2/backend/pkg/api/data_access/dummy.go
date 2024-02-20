package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/types/api"

type DummyService struct {
}

func NewDummyService() DummyService {
	return DummyService{}
}

func (d DummyService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	return t.DashboardData{}, nil
}
