package dataaccess

import (
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	// WORKING Rami
	ret := &t.VDBHeatmap{}
	queryResult := []struct {
		GroupId uint64  `db:"group_id"`
		Epoch   uint64  `db:"epoch"`
		Rewards float64 `db:"attestations_eff"`
	}{}
	// WIP
	// epoch based data doesn't seem possible, we only store data of the last ~14 epochs
	// -> wait for redesign of the heatmap (prob daily data)
	if dashboardId.Validators == nil {
		wg := errgroup.Group{}

		wg.Go(func() error {
			query := `
			SELECT
				id
			FROM
				users_val_dashboards_groups
			WHERE
				dashboard_id = $1
			ORDER BY
				id`
			return d.alloyReader.Select(&ret.GroupIds, query, dashboardId.Id)
		})

		wg.Go(func() error {
			query := `
			SELECT
				group_id,
				epoch,
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward), 0) AS attestations_eff
			FROM
				validator_dashboard_data_epoch epochs
			LEFT JOIN users_val_dashboards_validators validators ON validators.validator_index = epochs.validator_index
			WHERE
				dashboard_id = $1
			GROUP BY
				epoch, group_id
			ORDER BY
				epoch`

			return d.alloyReader.Select(&queryResult, query, dashboardId.Id)
		})
		err := wg.Wait()
		if err != nil {
			return nil, err
		}
	} else {
		validators, err := d.getDashboardValidators(dashboardId)
		if err != nil || len(validators) == 0 {
			return ret, err
		}

		query := `
		SELECT
			$2::int AS group_id,
			epoch,
			SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward), 0) AS attestations_eff
		FROM
			validator_dashboard_data_epoch
		WHERE
			validator_index = ANY($1)
		GROUP BY
			epoch
		ORDER BY
			epoch`

		err = d.alloyReader.Select(&queryResult, query, validators, t.DefaultGroupId)
		if err != nil {
			return ret, err
		}
		ret.GroupIds = append(ret.GroupIds, t.DefaultGroupId)
	}

	groupIdxMap := make(map[uint64]uint64)
	for i, group := range ret.GroupIds {
		groupIdxMap[group] = uint64(i)
	}
	for _, res := range queryResult {
		if len(ret.Epochs) == 0 || ret.Epochs[len(ret.Epochs)-1] < res.Epoch {
			ret.Epochs = append(ret.Epochs, res.Epoch)
		}
		cell := t.VDBHeatmapCell{
			X: uint64(len(ret.Epochs) - 1),
			Y: groupIdxMap[res.GroupId],
		}
		if res.Rewards > 0 {
			cell.Value = res.Rewards * 100
		}
		ret.Data = append(ret.Data, cell)
	}
	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	ret := &t.VDBHeatmapTooltipData{Epoch: epoch}

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

	for _, slashed := range slashings {
		ret.Slashings = append(ret.Slashings, t.VDBHeatmapTooltipDuty{
			Validator: slashed,
			Status:    "success",
		})
	}

	var totalAttestationReward int64
	var totalAttestationIdealReward int64
	var totalAttestationsScheduled uint64
	var totalAttestationsHeadExecuted uint64
	var totalAttestationsSourceExecuted uint64
	var totalAttestationsTargetExecuted uint64
	for _, res := range queryResult {
		if res.Slashed {
			ret.Slashings = append(ret.Slashings, t.VDBHeatmapTooltipDuty{
				Validator: res.Validator,
				Status:    "failed",
			})
		}
		if res.SyncScheduled > 0 /* && epoch % 256 == 0 */ { // move to circle logic
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
		}

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
