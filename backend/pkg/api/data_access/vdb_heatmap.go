package dataaccess

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
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
	ret := &t.VDBHeatmapTooltipData{}

	var validators []uint64
	err := d.alloyReader.Select(&validators, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 AND group_id = $2`, dashboardId.Id, groupId)
	if err != nil {
		return nil, err
	}

	queryResult := []struct {
		Validator                  uint64 `db:"validator_index"`
		BlocksScheduled            uint8  `db:"blocks_scheduled"`
		BlocksProposed             uint8  `db:"blocks_proposed"`
		SyncScheduled              uint8  `db:"sync_scheduled"`
		SyncExecuted               uint8  `db:"sync_executed"`
		Slashed                    bool   `db:"slashed"`
		AttestationReward          int64  `db:"attestations_reward"`
		AttestationIdealReward     int64  `db:"attestations_ideal_reward"`
		AttestationsScheduled      uint8  `db:"attestations_scheduled"`
		AttestationsHeadExecuted   uint8  `db:"attestation_head_executed"`
		AttestationsSourceExecuted uint8  `db:"attestation_source_executed"`
		AttestationsTargetExecuted uint8  `db:"attestation_target_executed"`
	}{}
	wg := errgroup.Group{}
	wg.Go(func() error {
		query := `
			SELECT
				validator_index,
				blocks_scheduled,
				blocks_proposed,
				sync_scheduled,
				sync_executed,
				slashed,
				attestations_reward,
				attestations_ideal_reward,
				attestations_scheduled,
				attestation_head_executed,
				attestation_source_executed,
				attestation_target_executed
			FROM
				validator_dashboard_data_epoch
			WHERE
				validator_index = ANY($1) AND epoch = $2`

		return d.alloyReader.Select(&queryResult, query, validators, epoch)
	})

	var slashings []uint64
	wg.Go(func() error {
		query := `
			SELECT
				validator_index
			FROM
				validator_dashboard_data_rolling_daily
			WHERE
				slashed_by = ANY($1)`

		return d.alloyReader.Select(&slashings, query, validators, epoch)
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	/*for _, slashed := range slashings {
		ret.Slashings = append(ret.Slashings, t.VDBHeatmapTooltipDuty{
			Validator: slashed,
			Status:    "success",
		})
	}*/

	var totalAttestationReward int64
	var totalAttestationIdealReward int64
	var totalAttestationsScheduled uint64
	var totalAttestationsHeadExecuted uint64
	var totalAttestationsSourceExecuted uint64
	var totalAttestationsTargetExecuted uint64
	for _, res := range queryResult {
		/*if res.Slashed {
			ret.Slashings = append(ret.Slashings, t.VDBHeatmapTooltipDuty{
				Validator: res.Validator,
				Status:    "failed",
			})
		}
		if res.SyncScheduled > 0  { // && epoch % 256 == 0 move to circle logic
			ret.Syncs = append(ret.Syncs, res.Validator)
		}
		for i := uint8(0); i < res.BlocksScheduled; i++ {
			status := "success"
			if i >= res.BlocksProposed {
				status = "failed"
			}
			ret.Proposers = append(ret.Proposers, t.VDBHeatmapTooltipDuty{
				Validator: res.Validator,
				Status:    status,
			})
		}*/

		totalAttestationReward += res.AttestationReward
		totalAttestationIdealReward += res.AttestationIdealReward
		totalAttestationsScheduled += uint64(res.AttestationsScheduled)
		totalAttestationsHeadExecuted += uint64(res.AttestationsHeadExecuted)
		totalAttestationsSourceExecuted += uint64(res.AttestationsSourceExecuted)
		totalAttestationsTargetExecuted += uint64(res.AttestationsTargetExecuted)
	}

	ret.AttestationIncome = decimal.NewFromInt(totalAttestationReward)
	if totalAttestationIdealReward != 0 {
		ret.AttestationEfficiency = float64(totalAttestationReward) / float64(totalAttestationIdealReward)
	}
	ret.AttestationsHead = t.StatusCount{Success: totalAttestationsHeadExecuted, Failed: totalAttestationsScheduled - totalAttestationsHeadExecuted}
	ret.AttestationsSource = t.StatusCount{Success: totalAttestationsSourceExecuted, Failed: totalAttestationsScheduled - totalAttestationsSourceExecuted}
	ret.AttestationsTarget = t.StatusCount{Success: totalAttestationsTargetExecuted, Failed: totalAttestationsScheduled - totalAttestationsTargetExecuted}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupDailyHeatmap(dashboardId t.VDBId, groupId uint64, day time.Time) (*t.VDBHeatmapTooltipData, error) {
	// TODO @remoterami
	return d.dummy.GetValidatorDashboardGroupDailyHeatmap(dashboardId, groupId, day)
}
