package constants

import (
	"time"
)

// status enum
type StatusType string
type Event string

const (
	Running                                 StatusType    = "running"
	Success                                 StatusType    = "success"
	Failure                                 StatusType    = "failure"
	Default                                 time.Duration = -1 * time.Second
	Event_ApiServiceAvgEfficiency           Event         = "api_service_avg_efficiency"
	Event_ApiServiceSlotViz                 Event         = "api_service_slot_viz"
	Event_ApiServiceValidatorMapping        Event         = "api_service_validator_mapping"
	Event_ExporterLegacyNetworkLiveness     Event         = "exporter_legacy_network_liveness"
	Event_ExporterLegacySyncCommittees      Event         = "exporter_legacy_sync_committees"
	Event_ExporterLegacySyncCommitteesCount Event         = "exporter_legacy_sync_committees_count"
	Event_ExporterLegacyRocketPool          Event         = "exporter_legacy_rocket_pool"
	Event_ExporterLegacyPubkeyTags          Event         = "exporter_legacy_pubkey_tags"
	Event_ExporterModuleELRewardsFinalizer  Event         = "exporter_module_el_rewards_finalizer"
	Event_ExporterModuleELPayloadExporter   Event         = "exporter_module_el_payload_exporter"
	Event_ExporterModuleELDepositsExporter  Event         = "exporter_module_el_deposits_exporter"
	Event_ExporterModuleSlotExporter        Event         = "exporter_module_slot_exporter"
	Event_ExporterModuleDashboardData       Event         = "exporter_module_dashboard_data"
	Event_MonitoringCleanShutdown           Event         = "clean_shutdown"
	Event_MonitoringCleanShutdownSpam       Event         = "monitoring_clean_shutdown_spam"
	Event_MonitoringTimeouts                Event         = "monitoring_timeouts"
	Event_ClickhouseDashboardEpoch          Event         = "ch_dashboard_epoch"
	Event_ClickhouseRolling_1h              Event         = "ch_rolling_1h"
	Event_ClickhouseRolling_24h             Event         = "ch_rolling_24h"
	Event_ClickhouseRolling_7d              Event         = "ch_rolling_7d"
	Event_ClickhouseRolling_30d             Event         = "ch_rolling_30d"
	Event_ClickhouseRolling_90d             Event         = "ch_rolling_90d"
	Event_ClickhouseRolling_total           Event         = "ch_rolling_total"
	Event_DBConnReaderDB                    Event         = "db_conn_reader_db"
	Event_DBConnWriterDB                    Event         = "db_conn_writer_db"
	Event_DBConnUserReader                  Event         = "db_conn_user_reader"
	Event_DBConnUserWriter                  Event         = "db_conn_user_writer"
	Event_DBConnAlloyReader                 Event         = "db_conn_alloy_reader"
	Event_DBConnAlloyWriter                 Event         = "db_conn_alloy_writer"
	Event_DBConnFrontendReaderDB            Event         = "db_conn_frontend_reader_db"
	Event_DBConnFrontendWriterDB            Event         = "db_conn_frontend_writer_db"
	Event_DBConnClickhouseReader            Event         = "db_conn_clickhouse_reader"
	Event_DBConnClickhouseWriter            Event         = "db_conn_clickhouse_writer"
	Event_DBConnClickhouseNativeWriter      Event         = "db_conn_clickhouse_native_writer"
	Event_DBConnPersistentRedisDbClient     Event         = "db_conn_persistent_redis_db_client"
	Event_DBConnTieredCache                 Event         = "db_conn_tiered_cache"
)

// events that if not present in the monitoring system, should cause an alert
var RequiredEvents = []Event{
	Event_ApiServiceAvgEfficiency,
	Event_ApiServiceSlotViz,
	Event_ApiServiceValidatorMapping,
	Event_MonitoringTimeouts,
	Event_ClickhouseDashboardEpoch,
	Event_ClickhouseRolling_1h,
	Event_ClickhouseRolling_24h,
	Event_ClickhouseRolling_7d,
	Event_ClickhouseRolling_30d,
	Event_ClickhouseRolling_90d,
	Event_ClickhouseRolling_total,
}

// events that only need to be present in production
var ProductionRequiredEvents = []Event{
	Event_ExporterLegacyNetworkLiveness,
	Event_ExporterLegacySyncCommittees,
	Event_ExporterLegacySyncCommitteesCount,
	Event_ExporterModuleELRewardsFinalizer,
	Event_ExporterModuleELPayloadExporter,
	Event_ExporterModuleELDepositsExporter,
	Event_ExporterModuleSlotExporter,
}
