package db2

import (
	"context"
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
