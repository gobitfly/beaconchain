package dataaccess

import (
	"context"
	"database/sql"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type ArchiverRepository interface {
	GetValidatorDashboardsInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error)
}

func (d *DataAccessService) GetValidatorDashboardsInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error) {
	result := make(map[uint64][]t.ArchiverDashboard)

	type DashboardInfo struct {
		Id             uint64         `db:"id"`
		UserId         uint64         `db:"user_id"`
		IsArchived     sql.NullString `db:"is_archived"`
		GroupCount     uint64         `db:"group_count"`
		ValidatorCount uint64         `db:"validator_count"`
	}

	var dbReturn []DashboardInfo
	err := d.readerDb.Select(&dbReturn, `
		WITH dashboards_groups AS
			(SELECT
				dashboard_id,
				COUNT(id) AS group_count
			FROM users_val_dashboards_groups
			GROUP BY dashboard_id),
		dashboards_validators AS
			(SELECT
				dashboard_id,
				COUNT(validator_index) AS validator_count
			FROM users_val_dashboards_validators
			GROUP BY dashboard_id)
		SELECT
			uvd.id,
			uvd.user_id,
			uvd.is_archived,
		    COALESCE(dg.group_count, 0) AS group_count,
		    COALESCE(dv.validator_count, 0) AS validator_count
		FROM users_val_dashboards uvd
		LEFT JOIN dashboards_groups dg ON uvd.id = dg.dashboard_id
		LEFT JOIN dashboards_validators dv ON uvd.id = dv.dashboard_id
	`)
	if err != nil {
		return nil, err
	}

	for _, dashboardInfo := range dbReturn {
		if _, ok := result[dashboardInfo.UserId]; !ok {
			result[dashboardInfo.UserId] = make([]t.ArchiverDashboard, 0)
		}

		dashboard := t.ArchiverDashboard{
			DashboardId:    dashboardInfo.Id,
			IsArchived:     dashboardInfo.IsArchived.Valid,
			GroupCount:     dashboardInfo.GroupCount,
			ValidatorCount: dashboardInfo.ValidatorCount,
		}

		result[dashboardInfo.UserId] = append(result[dashboardInfo.UserId], dashboard)
	}

	return result, nil
}
