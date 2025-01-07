package dataaccess

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/lib/pq"
)

type ArchiverRepository interface {
	GetValidatorDashboardsCountInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error)
	UpdateValidatorDashboardsArchiving(ctx context.Context, dashboards []t.ArchiverDashboardArchiveReason) error
	RemoveValidatorDashboards(ctx context.Context, dashboardIds []uint64) error
}

func (d *DataAccessService) GetValidatorDashboardsCountInfo(ctx context.Context) (map[uint64][]t.ArchiverDashboard, error) {
	result := make(map[uint64][]t.ArchiverDashboard)

	type DashboardInfo struct {
		Id             uint64         `db:"id"`
		UserId         uint64         `db:"user_id"`
		IsArchived     sql.NullString `db:"is_archived"`
		GroupCount     uint64         `db:"group_count"`
		ValidatorCount uint64         `db:"validator_count"`
	}

	var dbReturn []DashboardInfo
	err := d.readerDb.SelectContext(ctx, &dbReturn, `
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
		ORDER BY
			uvd.created_at ASC
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

func (d *DataAccessService) UpdateValidatorDashboardsArchiving(ctx context.Context, dashboards []t.ArchiverDashboardArchiveReason) error {
	ds := goqu.Dialect("postgres").Update("users_val_dashboards")

	cases := goqu.Case()
	for _, dashboard := range dashboards {
		cases = cases.When(goqu.I("id").Eq(dashboard.DashboardId), dashboard.ArchivedReason.ToString())
	}

	ds = ds.Set(goqu.Record{"is_archived": cases})

	// Restrict the query to the ids we want to set
	ids := make([]interface{}, len(dashboards))
	for i, dashboard := range dashboards {
		ids[i] = dashboard.DashboardId
	}
	ds = ds.Where(goqu.I("id").In(ids...))

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}

	_, err = d.writerDb.ExecContext(ctx, query, args...)
	return err
}

func (d *DataAccessService) RemoveValidatorDashboards(ctx context.Context, dashboardIds []uint64) error {
	// Delete the dashboards
	_, err := d.writerDb.ExecContext(ctx, `
		DELETE FROM users_val_dashboards WHERE id = ANY($1)
	`, dashboardIds)
	if err != nil {
		return err
	}

	var prefixes []string
	for _, dashboardId := range dashboardIds {
		prefixes = append(prefixes, fmt.Sprintf("%s:%d:%%", ValidatorDashboardEventPrefix, dashboardId))
	}

	// Remove all events related to the dashboards
	_, err = d.userWriter.ExecContext(ctx, `
		DELETE FROM users_subscriptions WHERE event_filter LIKE ANY($1)
	`, pq.Array(prefixes))

	return err
}
