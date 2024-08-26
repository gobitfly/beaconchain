package dataaccess

import (
	"context"
	"slices"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
)

type HealthzRepository interface {
	GetHealthz(ctx context.Context, showAll bool) types.HealthzData
}

func (d *DataAccessService) GetHealthz(ctx context.Context, showAll bool) types.HealthzData {
	var results []types.HealthzResult
	var response types.HealthzData
	query := `
		WITH sub AS
			(
				SELECT
					emitter,
					event_id,
					max(inserted_at) AS inserted_at,
					max(expires_at) AS expires_at,
					any(status) AS status,
					any(mapSort(metadata)) AS metadata
				FROM status_reports AS s
				WHERE s.expires_at > now()
				GROUP BY
					1,
					2,
					s.status
				ORDER BY
					inserted_at DESC,
					1 ASC,
					2 ASC
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
					'metadata',
					CAST(metadata, 'String')
				)
			) AS result
		FROM sub
		GROUP BY
			event_id,
			status
		ORDER BY event_id, max(inserted_at) DESC
	`

	response.Reports = make(map[string][]types.HealthzResult)
	err := db.ClickHouseReader.SelectContext(ctx, &results, query)
	if err != nil {
		response.Reports["response_error"] = []types.HealthzResult{
			{
				EventId: "response_error",
				Status:  "failure",
				Result:  []map[string]string{{"error": "failed to fetch status reports"}},
			},
		}

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
	}
	for _, result := range results {
		response.Reports[result.EventId] = append(response.Reports[result.EventId], result)
	}
	for _, id := range mustExist {
		if _, ok := response.Reports[id]; !ok {
			response.Reports[id] = []types.HealthzResult{
				{
					EventId: id,
					Status:  "failure",
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
			if report.Status == "failure" {
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
