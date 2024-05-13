package dataaccess

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) queryHeatmap(params []interface{}, data interface{}, timeframe string, days, validators bool) error {
	from := `validator_dashboard_data_epoch`
	timestampCol := `epoch`
	if days {
		timestampCol = `day`
		from = `validator_dashboard_data_daily`
	}
	where := `dashboard_id = $1`
	group_id := `group_id`
	group_by := timestampCol + `, group_id`
	// sort most efficient group to the top
	order_by := timestampCol + `, AVG(SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward), 0)) over (partition by group_id) DESC`
	if validators {
		where = `validator_index = ANY($1)`
		group_id = `$2::int AS ` + group_id
		group_by = timestampCol
		order_by = timestampCol
	}
	if days {
		where += ` AND day > now() - interval '` + timeframe + `'`
	}
	query := fmt.Sprintf(`
		SELECT
			%s,
			%s,
			SUM(attestations_reward)::decimal * 100 / NULLIF(SUM(attestations_ideal_reward), 0) AS attestations_eff,
			COALESCE(SUM(blocks_scheduled), 0)::int::bool AS blocks_scheduled,
			COALESCE(SUM(sync_scheduled), 0)::int::bool AS sync_scheduled,
			COALESCE(SUM((slashed_by is not NULL)::int) FILTER (where slashed = true), 0)::int::bool AS slashings
		FROM
			%s data
		LEFT JOIN
			users_val_dashboards_validators validators ON validators.validator_index = data.validator_index
		WHERE
			%s
		GROUP BY
			%s
		ORDER BY
			%s`,
		group_id, timestampCol, from, where, group_by, order_by)
	return d.alloyReader.Select(data, query, params...)
}

func (d *DataAccessService) queryHeatmapDetails(data *t.VDBHeatmapTooltipData, params []interface{}, days bool) error {
	// TODO slashed_by subquery is slow, need an index; missing permissions to create
	queryResult := struct {
		BlocksScheduled uint64 `db:"blocks_scheduled"`
		BlocksProposed  uint64 `db:"blocks_proposed"`
		SyncScheduled   uint64 `db:"sync_scheduled"`
		// SyncExecuted               uint8  `db:"sync_executed"`
		OwnSlashed                 uint64  `db:"own_slashed"`
		OthersSlashed              uint64  `db:"others_slashed"`
		AttestationReward          int64   `db:"attestations_reward"`
		AttestationEfficiency      float64 `db:"attestations_eff"`
		AttestationsScheduled      uint64  `db:"attestations_scheduled"`
		AttestationsHeadExecuted   uint64  `db:"attestation_head_executed"`
		AttestationsSourceExecuted uint64  `db:"attestation_source_executed"`
		AttestationsTargetExecuted uint64  `db:"attestation_target_executed"`
	}{}
	timeframe := "epoch"
	table := "validator_dashboard_data_epoch"
	if days {
		timeframe = "day"
		table = "validator_dashboard_data_daily"
	}
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(blocks_scheduled), 0) AS blocks_scheduled,
			COALESCE(SUM(blocks_proposed), 0) AS blocks_proposed,
			COALESCE(SUM(sync_scheduled), 0) AS sync_scheduled,
			COALESCE(SUM(slashed::int) FILTER (WHERE slashed_by IS NOT NULL), 0) AS own_slashed,
			(SELECT COUNT(*) FROM %s WHERE slashed_by = ANY($1) AND %s = $2) AS others_slashed,
			COALESCE(SUM(attestations_reward), 0) AS attestations_reward,
			COALESCE(SUM(attestations_reward)::decimal / SUM(attestations_ideal_reward) * 100, 0) AS attestations_eff,
			COALESCE(SUM(attestations_scheduled), 0) AS attestations_scheduled,
			COALESCE(SUM(attestation_head_executed), 0) AS attestation_head_executed,
			COALESCE(SUM(attestation_source_executed), 0) AS attestation_source_executed,
			COALESCE(SUM(attestation_target_executed), 0) AS attestation_target_executed
		FROM
			%s
		WHERE
			validator_index = ANY($1) AND %s = $2`, table, timeframe, table, timeframe)

	err := d.alloyReader.Get(&queryResult, query, params...)
	if err != nil {
		return err
	}

	data.Proposers.Success = queryResult.BlocksProposed
	data.Proposers.Failed = queryResult.BlocksScheduled - queryResult.BlocksProposed
	data.Syncs = queryResult.SyncScheduled
	data.Slashings.Success = queryResult.OthersSlashed
	data.Slashings.Failed = queryResult.OwnSlashed

	data.AttestationsHead.Success = queryResult.AttestationsHeadExecuted
	data.AttestationsHead.Failed = queryResult.AttestationsScheduled - queryResult.AttestationsHeadExecuted
	data.AttestationsSource.Success = queryResult.AttestationsSourceExecuted
	data.AttestationsSource.Failed = queryResult.AttestationsScheduled - queryResult.AttestationsSourceExecuted
	data.AttestationsTarget.Success = queryResult.AttestationsTargetExecuted
	data.AttestationsTarget.Failed = queryResult.AttestationsScheduled - queryResult.AttestationsTargetExecuted
	data.AttestationIncome = decimal.NewFromInt(queryResult.AttestationReward)
	data.AttestationEfficiency = queryResult.AttestationEfficiency
	return nil
}

// retrieve data for last hour
func (d *DataAccessService) GetValidatorDashboardEpochHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	ret := &t.VDBHeatmap{Aggregation: "epoch"}
	queryResult := []struct {
		GroupId         uint64  `db:"group_id"`
		Epoch           uint64  `db:"epoch"`
		Rewards         float64 `db:"attestations_eff"`
		BlocksScheduled bool    `db:"blocks_scheduled"`
		SyncScheduled   bool    `db:"sync_scheduled"`
		Slashings       bool    `db:"slashings"` // only includes own slashings for now (red ones)
	}{}
	groupIdxMap := make(map[uint64]uint64)
	if dashboardId.Validators == nil {
		var params []interface{} = []interface{}{dashboardId.Id}
		err := d.queryHeatmap(params, &queryResult, "", false, false)
		if err != nil {
			return nil, err
		}
		// fill in groups
		for i, val := range queryResult {
			if len(ret.GroupIds) > 0 && ret.GroupIds[0] == val.GroupId {
				// results are secondary-sorted by group, so they'll loop for each epoch
				break
			}
			ret.GroupIds = append(ret.GroupIds, val.GroupId)
			groupIdxMap[val.GroupId] = uint64(i)
		}
	} else {
		validators, err := d.getDashboardValidators(dashboardId)
		if err != nil || len(validators) == 0 {
			return nil, err
		}
		var params []interface{} = []interface{}{validators, t.DefaultGroupId}
		err = d.queryHeatmap(params, &queryResult, "", false, true)
		if err != nil {
			return nil, err
		}
		ret.GroupIds = append(ret.GroupIds, t.DefaultGroupId)
		groupIdxMap[t.DefaultGroupId] = uint64(0)
	}
	for _, res := range queryResult {
		if len(ret.Timestamps) == 0 || ret.Timestamps[len(ret.Timestamps)-1] < utils.EpochToTime(res.Epoch).Unix() {
			ret.Timestamps = append(ret.Timestamps, utils.EpochToTime(res.Epoch).Unix())
		}
		cell := t.VDBHeatmapCell{
			X: int64(len(ret.Timestamps) - 1),
			Y: groupIdxMap[res.GroupId],
		}
		if res.Rewards > 0 {
			cell.Value = res.Rewards
		}
		ret.Data = append(ret.Data, cell)
	}
	return ret, nil
}

// allowed periods are: last_7d, last_30d, last_365d
func (d *DataAccessService) GetValidatorDashboardDailyHeatmap(dashboardId t.VDBId, period enums.TimePeriod) (*t.VDBHeatmap, error) {
	ret := &t.VDBHeatmap{Aggregation: "day"}
	queryResult := []struct {
		GroupId         uint64    `db:"group_id"`
		Day             time.Time `db:"day"`
		Rewards         float64   `db:"attestations_eff"`
		BlocksScheduled bool      `db:"blocks_scheduled"`
		SyncScheduled   bool      `db:"sync_scheduled"`
		Slashings       bool      `db:"slashings"` // only includes own slashings for now (red ones)
	}{}
	timeframe := ""
	groupIdxMap := make(map[uint64]uint64)
	switch period {
	case enums.Last7d:
		timeframe = "7 days"
	case enums.Last30d:
		timeframe = "30 days"
	case enums.Last365d:
		timeframe = "365 days"
	}
	if dashboardId.Validators == nil {
		var params []interface{} = []interface{}{dashboardId.Id}
		err := d.queryHeatmap(params, &queryResult, timeframe, true, false)
		if err != nil {
			return nil, err
		}
		// fill in groups
		for i, val := range queryResult {
			if len(ret.GroupIds) > 0 && ret.GroupIds[0] == val.GroupId {
				// results are secondary-sorted by group, so they'll loop for each day
				break
			}
			ret.GroupIds = append(ret.GroupIds, val.GroupId)
			groupIdxMap[val.GroupId] = uint64(i)
		}
	} else {
		validators, err := d.getDashboardValidators(dashboardId)
		if err != nil || len(validators) == 0 {
			return nil, err
		}
		var params []interface{} = []interface{}{validators, t.DefaultGroupId}
		err = d.queryHeatmap(params, &queryResult, timeframe, true, true)
		if err != nil {
			return nil, err
		}
		ret.GroupIds = append(ret.GroupIds, t.DefaultGroupId)
		groupIdxMap[t.DefaultGroupId] = uint64(0)
	}

	for _, res := range queryResult {
		if len(ret.Timestamps) == 0 || ret.Timestamps[len(ret.Timestamps)-1] < res.Day.Unix() {
			ret.Timestamps = append(ret.Timestamps, res.Day.Unix())
		}
		cell := t.VDBHeatmapCell{
			X: int64(len(ret.Timestamps) - 1),
			Y: groupIdxMap[res.GroupId],
		}
		// negative rewards are possible, but we're showing those as 0% for now
		if res.Rewards > 0 {
			cell.Value = res.Rewards
		}
		ret.Data = append(ret.Data, cell)
	}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupEpochHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	ret := &t.VDBHeatmapTooltipData{Timestamp: utils.EpochToTime(epoch).Unix()}

	validators, err := d.getDashboardValidators(dashboardId)
	if err != nil {
		return nil, err
	}

	if len(validators) == 0 {
		return ret, nil
	}

	err = d.queryHeatmapDetails(ret, []interface{}{validators, epoch}, false)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupDailyHeatmap(dashboardId t.VDBId, groupId uint64, day time.Time) (*t.VDBHeatmapTooltipData, error) {
	ret := &t.VDBHeatmapTooltipData{Timestamp: day.Unix()}

	validators, err := d.getDashboardValidators(dashboardId)
	if err != nil {
		return nil, err
	}

	if len(validators) == 0 {
		return ret, nil
	}

	err = d.queryHeatmapDetails(ret, []interface{}{validators, day.Format("2006-01-02")}, true)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
