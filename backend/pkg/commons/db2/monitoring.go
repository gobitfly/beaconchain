package db2

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/jmoiron/sqlx"
)

type MonitoringClickhouse struct {
	ClickHouseReader       *sqlx.DB
	ClickHouseNativeWriter driver.Conn
}

func NewMonitoringClickhouse(chReader *sqlx.DB, chNativeWriter driver.Conn) *MonitoringClickhouse {
	return &MonitoringClickhouse{
		ClickHouseReader:       chReader,
		ClickHouseNativeWriter: chNativeWriter,
	}
}

type StatusReport struct {
	ID         constants.Event
	Flake      int64
	ExpiresAt  time.Time
	TimeoutsAt time.Time
	Metadata   map[string]string
}

func (m *MonitoringClickhouse) SaveNewStatusReport(status StatusReport) error {
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

type Victims struct {
	EventID    string            `db:"event_id"`
	Emitter    string            `db:"emitter"`
	Status     string            `db:"status"`
	InsertedAt time.Time         `db:"inserted_at"`
	ExpiresAt  time.Time         `db:"expires_at"`
	TimeoutsAt time.Time         `db:"timeouts_at"`
	Metadata   map[string]string `db:"metadata"`
}

func (m *MonitoringClickhouse) GetLatestStatusReport() ([]Victims, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := `
		with active_reports as (
			SELECT
				event_id,
				emitter,
				run_id,
				inserted_at,
				insert_id,
				expires_at,
				timeouts_at,
				status,
				metadata
			FROM status_reports
			WHERE expires_at > now() and deployment_type = ? and emitter not in (select distinct emitter from status_reports where event_id = ? and inserted_at > now() - interval 1 days)
			ORDER BY
				event_id ASC,
				emitter ASC,
				run_id ASC,
				insert_id DESC
		), latest_report_per_run as (
			SELECT
				event_id,
				emitter,
				any(inserted_at) as inserted_at, 
				any(insert_id) as insert_id, 
				any(expires_at) as expires_at,
				any(timeouts_at) as timeouts_at,
				any(status) AS status,
				any(metadata) AS metadata
			FROM
				active_reports
			GROUP BY
				event_id,
				emitter,
				run_id
			order by insert_id desc
		)
		SELECT
			event_id,
			emitter,
			status,
			inserted_at,
			expires_at,
			timeouts_at,
			metadata
		FROM
			latest_report_per_run
		WHERE status = 'running' and timeouts_at < now()
		ORDER BY event_id ASC, inserted_at DESC
		`

	var victims []Victims
	err := m.ClickHouseReader.SelectContext(ctx, &victims, query, utils.Config.DeploymentType, constants.Event_MonitoringCleanShutdown)
	if err != nil {
		return nil, err
	}

	return victims, nil
}

func (m *MonitoringClickhouse) GetEmitters() ([]string, error) {
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

func (m *MonitoringClickhouse) GetLatestEpoch() (time.Time, error) {
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

	return ts, nil
}

func (m *MonitoringClickhouse) GetEpochEnd(rolling string) (uint64, error) {
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

	return epochEnd, nil
}
