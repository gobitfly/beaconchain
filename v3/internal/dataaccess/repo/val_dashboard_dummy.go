package dataaccess

import (
	"context"
)

type DummyValidatorDashboardRepository struct{}

func (r *DummyValidatorDashboardRepository) Ping() error {
	return nil
}

var dummyDashboardResponse = ValidatorDashboard{DashboardId: 1, UserId: 2, IsPublic: false}

func (r *DummyValidatorDashboardRepository) CreateValidatorDashboard(ctx context.Context, userId uint64) (*ValidatorDashboard, error) {
	return &dummyDashboardResponse, nil
}

func (r *DummyValidatorDashboardRepository) GetValidatorDashboardsByUserId(ctx context.Context, userId uint64) (*[]ValidatorDashboard, error) {
	return &[]ValidatorDashboard{dummyDashboardResponse}, nil
}

func (r *DummyValidatorDashboardRepository) GetValidatorDashboardByDashboardId(ctx context.Context, dashboardId uint64) (*ValidatorDashboard, error) {
	return &dummyDashboardResponse, nil
}

func (r *DummyValidatorDashboardRepository) ModifyValidatorDashboard(ctx context.Context, dashboardId uint64, isPublic bool) (*ValidatorDashboard, error) {
	return &dummyDashboardResponse, nil
}

func (r *DummyValidatorDashboardRepository) DeleteValidatorDashboard(ctx context.Context, dashboardId uint64) error {
	return nil
}
