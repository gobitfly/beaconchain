package db2

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/jmoiron/sqlx"
)

type MonitoringDB struct {
	ClickHouseReader       *sqlx.DB
	ClickHouseNativeWriter driver.Conn
}

func NewMonitoringDB() *MonitoringDB {
	return &MonitoringDB{
		ClickHouseReader:       db.ClickHouseReader,
		ClickHouseNativeWriter: db.ClickHouseNativeWriter,
	}
}

type StatusReport struct {
	ID         constants.Event
	Flake      int64
	ExpiresAt  time.Time
	TimeoutsAt time.Time
	Metadata   map[string]string
}

func (m *MonitoringDB) SaveNewStatusReport(status StatusReport) error {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// wrap in clickhouse context so we can set the setting throw_if_deduplication_in_dependent_materialized_views_enabled_with_async_insert to 0
	// we have no materialized views on the status_reports table, but it triggers as we need the deduplication setting for the other tables
	ctx := clickhouse.Context(timeoutContext, clickhouse.WithSettings(
		clickhouse.Settings{
			"throw_if_deduplication_in_dependent_materialized_views_enabled_with_async_insert": 0,
		},
	))

	stmt := `INSERT INTO status_reports (
		emitter, 
		event_id, 
		deployment_type, 
		insert_id, 
		expires_at, 
		timeouts_at, 
		metadata
	) 
	VALUES 
		(?, ?, ?, ?, ?, ?, ?)`

	err := m.ClickHouseNativeWriter.AsyncInsert(
		ctx,
		stmt,
		false, // true means wait for settlement, but we want to shoot and forget. false does mean we cant log any errors that occur during settlement
		utils.GetUUID(),
		status.ID,
		utils.Config.DeploymentType,
		status.Flake,
		status.ExpiresAt,
		status.TimeoutsAt,
		status.Metadata,
	)

	return err
}

func (m *MonitoringDB) GetEmitters() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := `
		SELECT
			emitter
		FROM
			status_reports
		WHERE
			deployment_type = ?
			AND inserted_at >= now() - interval 5 minutes
			AND event_id = ?
		`

	var emitters []string
	err := m.ClickHouseReader.SelectContext(ctx, &emitters, query, utils.Config.DeploymentType, constants.Event_MonitoringCleanShutdown)
	if err != nil {
		return nil, err
	}

	return emitters, nil
}

func (m *MonitoringDB) GetVDLatestEpochTs() (time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var ts time.Time
	err := m.ClickHouseReader.GetContext(
		ctx,
		&ts,
		"SELECT MAX(t) FROM view_validator_dashboard_data_epoch_max_ts",
	)
	if err != nil {
		return time.Time{}, err
	}

	return ts, err
}

func (m *MonitoringDB) GetVDRollingEpochEnd(rolling string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var epochEnd uint64
	err := m.ClickHouseReader.GetContext(ctx, &epochEnd, fmt.Sprintf(`
			SELECT
				max(epoch_end)
			FROM validator_dashboard_data_rolling_%s`,
		rolling,
	),
	)
	if err != nil {
		return 0, err
	}

	return epochEnd, err
}
