package dataaccess

import (
	"context"
	"slices"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type HealthzRepository interface {
	GetHealthz(ctx context.Context, showAll bool) types.HealthzData
}

func (d *DataAccessService) GetHealthz(ctx context.Context, showAll bool) types.HealthzData {
	var results []types.HealthzResult
	var response types.HealthzData
	query := `
		with clean_shutdown_events as (
			SELECT
				emitter,
				toNullable(inserted_at) as inserted_at
			FROM
				status_reports
			WHERE
				deployment_type = {deployment_type:String}
				AND inserted_at >= now() - interval 1 days
				AND event_id = {clean_shutdown_event_id:String}
		), active_reports as (
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
			LEFT JOIN clean_shutdown_events cse ON status_reports.emitter = clean_shutdown_events.emitter
			WHERE expires_at > now() and deployment_type = {deployment_type:String} and (status_reports.inserted_at < cse.inserted_at or cse.inserted_at is null)
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
		), latest_report_per_status as (
			select 
				event_id,
				emitter,
				status,
				any(inserted_at) as inserted_at, 
				any(expires_at) as expires_at,
				any(timeouts_at) as timeouts_at,
				any(metadata) AS metadata
			from latest_report_per_run
			group by event_id, emitter, status
		)
		SELECT
			event_id,
			status,
			groupArray(
						map(
							'emitter',
							CAST(emitter, 'String'),
							'inserted_at',
							CAST(inserted_at, 'String'),
							'expires_at',
							CAST(expires_at, 'String'),
							'timeouts_at',
							CAST(timeouts_at, 'String'),
							'metadata',
							CAST(mapSort(metadata), 'String')
						)
					) as result
		FROM
			latest_report_per_status
		GROUP BY
			event_id, 
			status
		ORDER BY event_id ASC, max(inserted_at) DESC
		SETTINGS
			use_query_cache = true,
			query_cache_compress_entries = false,
			query_cache_nondeterministic_function_handling='save',
			query_cache_ttl=10
	`

	response.Reports = make(map[string][]types.HealthzResult)
	response.ReportingUUID = utils.GetUUID()
	response.DeploymentType = utils.Config.DeploymentType
	err := db.ClickHouseReader.SelectContext(ctx, &results, query, ch.Named("deployment_type", utils.Config.DeploymentType), ch.Named("clean_shutdown_event_id", constants.CleanShutdownEvent))
	if err != nil {
		response.Reports["response_error"] = []types.HealthzResult{
			{
				EventId: "response_error",
				Status:  "failure",
				Result:  []map[string]string{{"error": "failed to fetch status reports"}},
			},
		}
		log.Error(err, "failed to fetch status reports", 0)

		return response
	}

	mustExist := []string{
		"ch_rolling_1h",
		"ch_rolling_24h",
		"ch_rolling_7d",
		"ch_rolling_30d",
		"ch_rolling_90d",
		"ch_rolling_total",
		"ch_dashboard_epoch",
		"api_service_avg_efficiency",
		"api_service_validator_mapping",
		"api_service_slot_viz",
		"monitoring_timeouts",
	}
	for _, result := range results {
		response.Reports[result.EventId] = append(response.Reports[result.EventId], result)
	}
	for _, id := range mustExist {
		if _, ok := response.Reports[id]; !ok {
			response.Reports[id] = []types.HealthzResult{
				{
					Status: constants.Failure,
					Result: []map[string]string{
						{"error": "no status report found"},
					},
				},
			}
		}
	}

	failures := 0
	for _, r := range response.Reports {
		for _, report := range r {
			if report.Status == constants.Failure {
				failures++
			}
		}
	}
	if len(results) > 0 {
		response.TotalOkPercentage = 1 - float64(failures)/float64(len(results))
	}

	if !showAll {
		// we will filter out all reports that arent failure
		for id, result := range response.Reports {
			response.Reports[id] = slices.DeleteFunc(result, func(r types.HealthzResult) bool {
				return r.Status != "failure"
			})
			if len(response.Reports[id]) == 0 {
				delete(response.Reports, id)
			}
		}
	}

	return response
}
