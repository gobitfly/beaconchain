package dataaccess

import (
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	// WORKING spletka
	return d.dummy.GetValidatorDashboardRewards(dashboardId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	// TODO @peter_bitfly

	ret := &t.VDBGroupRewardsData{}

	query := `select
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestations_executed, 0) as attestations_executed,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			COALESCE(sync_rewards, 0) as sync_rewards,
		from users_val_dashboards_validators
		join validator_dashboard_data_epoch on validator_dashboard_data_epoch.validator_index = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and group_id = $2 and epoch = $3)
		`

	type queryResult struct {
		ValidatorIndex               uint32          `db:"validator_index"`
		AttestationSourceReward      decimal.Decimal `db:"attestations_source_reward"`
		AttestationTargetReward      decimal.Decimal `db:"attestations_target_reward"`
		AttestationHeadReward        decimal.Decimal `db:"attestations_head_reward"`
		AttestationInactivitytReward decimal.Decimal `db:"attestations_inactivity_reward"`
		AttestationInclusionReward   decimal.Decimal `db:"attestations_inclusion_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationsExecuted      int64 `db:"attestations_executed"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  decimal.Decimal `db:"blocks_cl_reward"`
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32          `db:"sync_scheduled"`
		SyncExecuted  uint32          `db:"sync_executed"`
		SyncRewards   decimal.Decimal `db:"sync_rewards"`
	}

	var rows []*queryResult

	err := d.alloyReader.Select(&rows, query, dashboardId, groupId, epoch)
	if err != nil {
		log.Error(err, "Error while getting validator dashboard group rewards", 0)
		return nil, err
	}

	gWei := decimal.NewFromInt(1e9)

	for _, row := range rows {
		ret.AttestationsHead.Income = ret.AttestationsHead.Income.Add(row.AttestationHeadReward.Mul(gWei))
		ret.AttestationsHead.StatusCount.Success = uint64(row.AttestationHeadExecuted)
		ret.AttestationsHead.StatusCount.Failed = uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

		ret.AttestationsSource.Income = ret.AttestationsSource.Income.Add(row.AttestationSourceReward.Mul(gWei))
		ret.AttestationsSource.StatusCount.Success = uint64(row.AttestationSourceExecuted)
		ret.AttestationsSource.StatusCount.Failed = uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

		ret.AttestationsTarget.Income = ret.AttestationsTarget.Income.Add(row.AttestationTargetReward.Mul(gWei))
		ret.AttestationsTarget.StatusCount.Success = uint64(row.AttestationTargetExecuted)
		ret.AttestationsTarget.StatusCount.Failed = uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

		ret.Inactivity.Income = ret.Inactivity.Income.Add(row.AttestationInactivitytReward.Mul(gWei))

		ret.Proposal.Income = ret.Proposal.Income.Add(row.BlocksClReward.Mul(gWei))
		ret.Proposal.StatusCount.Success += uint64(row.BlocksProposed)
		ret.Proposal.StatusCount.Failed += uint64(row.BlocksScheduled) - uint64(row.BlocksProposed)

		ret.Sync.Income.Add(row.SyncRewards.Mul(gWei))
		ret.Sync.StatusCount.Success += uint64(row.SyncExecuted)
		ret.Sync.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)
	}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// TODO @recy21
	// bar chart for the CL and EL rewards for each group for each epoch. NO series for all groups combined
	// series id is group id, series property is 'cl' or 'el'
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, groupId, cursor, colSort, search, limit)
}
