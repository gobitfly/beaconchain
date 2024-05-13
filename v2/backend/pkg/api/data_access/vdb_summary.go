package dataaccess

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	// TODO: implement sorting, filtering & paging
	ret := make(map[uint64]*t.VDBSummaryTableRow) // map of group id to result row
	retMux := &sync.Mutex{}

	// retrieve efficiency data for each time period, we cannot do sorting & filtering here as we need access to the whole set
	wg := errgroup.Group{}

	validators := make([]uint64, 0)
	if dashboardId.Validators != nil {
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}

		ret[0] = &t.VDBSummaryTableRow{
			Validators: append([]uint64{}, validators...),
		}
	}

	type queryResult struct {
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	retrieveAndProcessData := func(dashboardId t.VDBIdPrimary, validatorList []uint64, tableName string) (map[uint64]float64, error) {
		var queryResult []queryResult

		if len(validatorList) > 0 {
			query := `select 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
				select 
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
					SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
					SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
					from  %[1]s
				where validator_index = ANY($1)
			) as a;`
			err := d.alloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), validatorList)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		} else {
			query := `select group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
				select 
					group_id,
					SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
					SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
					SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
					from users_val_dashboards_validators 
				join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
				where dashboard_id = $1
				group by 1
			) as a;`
			err := d.alloyReader.Select(&queryResult, fmt.Sprintf(query, tableName), dashboardId)
			if err != nil {
				return nil, fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
			}
		}

		data := make(map[uint64]float64)
		for _, result := range queryResult {
			data[result.GroupId] = d.calculateTotalEfficiency(result.AttestationEfficiency, result.ProposerEfficiency, result.SyncEfficiency)
		}
		return data, nil
	}

	if len(validators) == 0 { // retrieve the validators & groups from the dashboard table
		wg.Go(func() error {
			type validatorsPerGroup struct {
				GroupId        uint64 `db:"group_id"`
				ValidatorIndex uint64 `db:"validator_index"`
			}

			var queryResult []validatorsPerGroup

			err := d.alloyReader.Select(&queryResult, `SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY group_id, validator_index`, dashboardId.Id)
			if err != nil {
				return fmt.Errorf("error retrieving validator groups for dashboard: %v", err)
			}

			retMux.Lock()
			for _, result := range queryResult {
				if ret[result.GroupId] == nil {
					ret[result.GroupId] = &t.VDBSummaryTableRow{
						GroupId: result.GroupId,
					}
				}

				if ret[result.GroupId].Validators == nil {
					ret[result.GroupId].Validators = make([]uint64, 0, 10)
				}

				if len(ret[result.GroupId].Validators) < 10 {
					ret[result.GroupId].Validators = append(ret[result.GroupId].Validators, result.ValidatorIndex)
				}
			}
			retMux.Unlock()
			return nil
		})
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_daily")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last24h = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_weekly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last7d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_monthly")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.Last30d = efficiency
		}
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(dashboardId.Id, validators, "validator_dashboard_data_rolling_total")
		if err != nil {
			return err
		}

		retMux.Lock()
		defer retMux.Unlock()
		for groupId, efficiency := range data {
			if ret[groupId] == nil {
				ret[groupId] = &t.VDBSummaryTableRow{GroupId: groupId}
			}

			ret[groupId].Efficiency.AllTime = efficiency
		}
		return nil
	})
	err := wg.Wait()

	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard summary data: %v", err)
	}

	retArr := make([]t.VDBSummaryTableRow, 0, len(ret))

	for _, row := range ret {
		retArr = append(retArr, *row)
	}

	sort.Slice(retArr, func(i, j int) bool {
		return retArr[i].GroupId < retArr[j].GroupId
	})

	paging := &t.Paging{
		TotalCount: uint64(len(retArr)),
	}

	return retArr, paging, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64) (*t.VDBGroupSummaryData, error) {
	ret := &t.VDBGroupSummaryData{}
	wg := errgroup.Group{}

	query := `select
			users_val_dashboards_validators.validator_index,
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
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
			COALESCE(slashed, false) as slashed,
			COALESCE(balance_start, 0) as balance_start,
			COALESCE(balance_end, 0) as balance_end,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(deposits_amount, 0) as deposits_amount,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
			COALESCE(withdrawals_amount, 0) as withdrawals_amount,
			COALESCE(block_chance, 0) as block_chance,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum
		from users_val_dashboards_validators
		join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and group_id = $2)
		`

	if dashboardId.Validators != nil {
		query = `select
			validator_index,
			COALESCE(attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
			COALESCE(attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
			COALESCE(attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
			COALESCE(attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
			COALESCE(attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
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
			COALESCE(slashed, false) as slashed,
			COALESCE(balance_start, 0) as balance_start,
			COALESCE(balance_end, 0) as balance_end,
			COALESCE(deposits_count, 0) as deposits_count,
			COALESCE(deposits_amount, 0) as deposits_amount,
			COALESCE(withdrawals_count, 0) as withdrawals_count,
			COALESCE(withdrawals_amount, 0) as withdrawals_amount,
			COALESCE(block_chance, 0) as block_chance,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum
		from %[1]s
		where %[1]s.validator_index = ANY($1)
	`
	}

	validators := make([]uint64, 0)
	if dashboardId.Validators != nil {
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}
	}

	type queryResult struct {
		ValidatorIndex                    uint32 `db:"validator_index"`
		AttestationSourceReward           int64  `db:"attestations_source_reward"`
		AttestationTargetReward           int64  `db:"attestations_target_reward"`
		AttestationHeadReward             int64  `db:"attestations_head_reward"`
		AttestationInactivitytReward      int64  `db:"attestations_inactivity_reward"`
		AttestationInclusionReward        int64  `db:"attestations_inclusion_reward"`
		AttestationReward                 int64  `db:"attestations_reward"`
		AttestationIdealSourceReward      int64  `db:"attestations_ideal_source_reward"`
		AttestationIdealTargetReward      int64  `db:"attestations_ideal_target_reward"`
		AttestationIdealHeadReward        int64  `db:"attestations_ideal_head_reward"`
		AttestationIdealInactivitytReward int64  `db:"attestations_ideal_inactivity_reward"`
		AttestationIdealInclusionReward   int64  `db:"attestations_ideal_inclusion_reward"`
		AttestationIdealReward            int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationsExecuted      int64 `db:"attestations_executed"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  uint64          `db:"blocks_cl_reward"`
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`
		SyncRewards   int64  `db:"sync_rewards"`

		Slashed bool `db:"slashed"`

		BalanceStart int64 `db:"balance_start"`
		BalanceEnd   int64 `db:"balance_end"`

		DepositsCount  uint32 `db:"deposits_count"`
		DepositsAmount int64  `db:"deposits_amount"`

		WithdrawalsCount  uint32 `db:"withdrawals_count"`
		WithdrawalsAmount int64  `db:"withdrawals_amount"`

		BlockChance float64 `db:"block_chance"`

		InclusionDelaySum int64 `db:"inclusion_delay_sum"`
	}

	retrieveAndProcessData := func(query, table string, days int, dashboardId t.VDBIdPrimary, groupId int64, validators []uint64) (*t.VDBGroupSummaryColumn, error) {
		data := t.VDBGroupSummaryColumn{}
		var rows []*queryResult
		var err error

		if len(validators) > 0 {
			err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table), validators)
		} else {
			err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table), dashboardId, groupId)
		}

		if err != nil {
			return nil, err
		}

		totalAttestationRewards := int64(0)
		totalIdealAttestationRewards := int64(0)
		totalStartBalance := int64(0)
		totalEndBalance := int64(0)
		totalDeposits := int64(0)
		totalWithdrawals := int64(0)
		totalBlockChance := float64(0)
		totalInclusionDelaySum := int64(0)
		totalInclusionDelayDivisor := int64(0)

		syncValidators := make([]uint64, 0)
		for _, row := range rows {
			syncValidators = append(syncValidators, uint64(row.ValidatorIndex))
			totalAttestationRewards += row.AttestationReward
			totalIdealAttestationRewards += row.AttestationIdealReward

			data.AttestationCount.Success += uint64(row.AttestationsExecuted)
			data.AttestationCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationsExecuted)

			data.AttestationsHead.StatusCount.Success += uint64(row.AttestationHeadExecuted)
			data.AttestationsHead.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

			data.AttestationsSource.StatusCount.Success += uint64(row.AttestationSourceExecuted)
			data.AttestationsSource.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

			data.AttestationsTarget.StatusCount.Success += uint64(row.AttestationTargetExecuted)
			data.AttestationsTarget.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

			if row.ValidatorIndex == 0 && row.BlocksProposed > 0 && row.BlocksProposed != row.BlocksScheduled {
				row.BlocksProposed-- // subtract the genesis block from validator 0 (TODO: remove when fixed in the dashoard data exporter)
			}
			data.Proposals.StatusCount.Success += uint64(row.BlocksProposed)
			data.Proposals.StatusCount.Failed += uint64(row.BlocksScheduled) - uint64(row.BlocksProposed)

			if row.BlocksScheduled > 0 {
				if data.Proposals.Validators == nil {
					data.Proposals.Validators = make([]uint64, 0, 10)
				}
				data.Proposals.Validators = append(data.Proposals.Validators, uint64(row.ValidatorIndex))
			}

			data.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
			data.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

			if row.SyncScheduled > 0 {
				if data.SyncCommittee.Validators == nil {
					data.SyncCommittee.Validators = make([]uint64, 0, 10)
				}
				data.SyncCommittee.Validators = append(data.SyncCommittee.Validators, uint64(row.ValidatorIndex))
			}

			if row.Slashed {
				data.Slashed.StatusCount.Failed++
				if data.Slashed.Validators == nil {
					data.Slashed.Validators = make([]uint64, 0, 10)
					data.Slashed.Validators = append(data.Slashed.Validators, uint64(row.ValidatorIndex))
				}
			} else {
				data.Slashed.StatusCount.Success++
			}

			totalStartBalance += row.BalanceStart
			totalEndBalance += row.BalanceEnd
			totalDeposits += row.DepositsAmount
			totalWithdrawals += row.WithdrawalsAmount
			totalBlockChance += row.BlockChance
			totalInclusionDelaySum += row.InclusionDelaySum

			if row.InclusionDelaySum > 0 {
				totalInclusionDelayDivisor += row.AttestationsScheduled
			}
		}

		reward := totalEndBalance + totalWithdrawals - totalStartBalance - totalDeposits
		//log.Infof("rows: %d, totalEndBalance: %d, totalWithdrawals: %d, totalStartBalance: %d, totalDeposits: %d", len(rows), totalEndBalance, totalWithdrawals, totalStartBalance, totalDeposits)
		aprDivisor := days
		if days == -1 { // for all time APR
			aprDivisor = 1
		}
		apr := ((float64(reward) / float64(aprDivisor)) / (float64(32e9) * float64(len(rows)))) * 365.0 * 100.0
		if math.IsNaN(apr) {
			apr = 0
		}

		data.Apr.Cl = apr
		data.Income.Cl = decimal.NewFromInt(reward).Mul(decimal.NewFromInt(1e9))

		data.Apr.El = 0

		data.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
		if data.AttestationEfficiency < 0 || math.IsNaN(data.AttestationEfficiency) {
			data.AttestationEfficiency = 0
		}

		if totalBlockChance > 0 {
			data.Luck.Proposal.Percent = (float64(data.Proposals.StatusCount.Failed) + float64(data.Proposals.StatusCount.Success)) / totalBlockChance * 100
		} else {
			data.Luck.Proposal.Percent = 0
		}

		endSyncLuckEpoch := cache.LatestFinalizedEpoch.Get()
		startSyncLuckEpoch := endSyncLuckEpoch - utils.EpochsPerDay()*uint64(days)
		if days == -1 {
			startSyncLuckEpoch = utils.Config.Chain.ClConfig.AltairForkEpoch
		}
		if startSyncLuckEpoch > endSyncLuckEpoch {
			startSyncLuckEpoch = 0
		}

		expectedSync, err := d.internal_getExpectedSyncCommitteeSlots(syncValidators, startSyncLuckEpoch, endSyncLuckEpoch)
		if err != nil {
			return nil, err
		}
		if expectedSync == 0 {
			data.Luck.Sync.Percent = 0
		} else {
			data.Luck.Sync.Percent = (float64(data.SyncCommittee.StatusCount.Failed) + float64(data.SyncCommittee.StatusCount.Success)) / float64(expectedSync) * 100
		}

		if totalInclusionDelayDivisor > 0 {
			data.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
		} else {
			data.AttestationAvgInclDist = 0
		}

		return &data, nil
	}

	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_daily", 1, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last24h = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_weekly", 7, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last7d = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_monthly", 31, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.Last30d = *data
		return nil
	})
	wg.Go(func() error {
		data, err := retrieveAndProcessData(query, "validator_dashboard_data_rolling_total", -1, dashboardId.Id, groupId, validators)
		if err != nil {
			return err
		}
		ret.AllTime = *data
		return nil
	})
	err := wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}
	return ret, nil
}

func (d *DataAccessService) internal_getExpectedSyncCommitteeSlots(validators []uint64, startEpoch, endEpoch uint64) (expectedSlots uint64, err error) {
	if endEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		// no sync committee duties before altair fork
		return 0, nil
	}

	// retrieve activation and exit epochs from database per validator
	type ValidatorInfo struct {
		Id                         int64  `db:"validatorindex"`
		ActivationEpoch            uint64 `db:"activationepoch"`
		ExitEpoch                  uint64 `db:"exitepoch"`
		FirstPossibleSyncCommittee uint64 // calculated
		LastPossibleSyncCommittee  uint64 // calculated
	}

	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return 0, err
	}

	// only check validators that are/have been active and that did not exit before altair
	noEpoch := uint64(math.MaxUint64)
	var validatorsInfo = make([]ValidatorInfo, 0, len(validators))
	for _, v := range validators {
		activationEpoch := noEpoch
		exitEpoch := noEpoch

		if validatorMapping.ValidatorMetadata[v].ActivationEpoch.Valid {
			activationEpoch = uint64(validatorMapping.ValidatorMetadata[v].ActivationEpoch.Int64)
		}
		if validatorMapping.ValidatorMetadata[v].ExitEpoch.Valid {
			activationEpoch = uint64(validatorMapping.ValidatorMetadata[v].ExitEpoch.Int64)
		}

		if activationEpoch != noEpoch && activationEpoch < endEpoch && (exitEpoch == noEpoch || exitEpoch >= utils.Config.Chain.ClConfig.AltairForkEpoch) {
			validatorsInfo = append(validatorsInfo, ValidatorInfo{
				Id:              int64(v),
				ActivationEpoch: activationEpoch,
				ExitEpoch:       exitEpoch,
			})
		}
	}

	if len(validatorsInfo) == 0 {
		// no validators relevant for sync duties
		return 0, nil
	}

	// we need all related and unique timeframes (activation and exit sync period) for all validators
	uniquePeriods := make(map[uint64]bool)
	for i := range validatorsInfo {
		// first epoch (activation epoch or Altair if Altair was later as there were no sync committees pre Altair)
		firstSyncEpoch := validatorsInfo[i].ActivationEpoch
		if validatorsInfo[i].ActivationEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			firstSyncEpoch = utils.Config.Chain.ClConfig.AltairForkEpoch
		}
		if firstSyncEpoch < startEpoch {
			firstSyncEpoch = startEpoch
		}

		validatorsInfo[i].FirstPossibleSyncCommittee = utils.SyncPeriodOfEpoch(firstSyncEpoch)
		uniquePeriods[validatorsInfo[i].FirstPossibleSyncCommittee] = true

		// last epoch (exit epoch or current epoch if not exited yet)
		lastSyncEpoch := endEpoch
		if validatorsInfo[i].ExitEpoch != noEpoch && validatorsInfo[i].ExitEpoch <= endEpoch {
			lastSyncEpoch = validatorsInfo[i].ExitEpoch
		}
		validatorsInfo[i].LastPossibleSyncCommittee = utils.SyncPeriodOfEpoch(lastSyncEpoch)
		uniquePeriods[validatorsInfo[i].LastPossibleSyncCommittee] = true
	}

	// transform map to slice; this will be used to query sync_committees_count_per_validator
	periodSlice := make([]uint64, 0, len(uniquePeriods))
	for period := range uniquePeriods {
		periodSlice = append(periodSlice, period)
	}

	// get aggregated count for all relevant committees from sync_committees_count_per_validator
	var countStatistics []struct {
		Period     uint64  `db:"period"`
		CountSoFar float64 `db:"count_so_far"`
	}

	query, args, errs := sqlx.In(`SELECT period, count_so_far FROM sync_committees_count_per_validator WHERE period IN (?) ORDER BY period ASC`, periodSlice)
	if errs != nil {
		return 0, errs
	}
	err = db.ReaderDb.Select(&countStatistics, db.ReaderDb.Rebind(query), args...)
	if err != nil {
		return 0, err
	}
	if len(countStatistics) != len(periodSlice) {
		return 0, fmt.Errorf("unable to retrieve all sync committee count statistics, required %v entries but got %v entries (startEpoch: %v, endEpoch: %v)", len(periodSlice), len(countStatistics), startEpoch, endEpoch)
	}

	// transform query result to map for easy access
	periodInfoMap := make(map[uint64]float64)
	for _, pl := range countStatistics {
		periodInfoMap[pl.Period] = pl.CountSoFar
	}

	// calculate expected committies for every single validator and aggregate them
	expectedCommitties := 0.0
	for _, vi := range validatorsInfo {
		expectedCommitties += periodInfoMap[vi.LastPossibleSyncCommittee] - periodInfoMap[vi.FirstPossibleSyncCommittee]
	}

	// transform committees to slots
	expectedSlots = uint64(expectedCommitties * float64(utils.SlotsPerSyncCommittee()))

	return expectedSlots, nil
}

// for summary charts: series id is group id, no stack

func (d *DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error) {
	ret := &t.ChartData[int, float64]{}

	type queryResult struct {
		StartEpoch            uint64          `db:"epoch_start"`
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	var queryResults []queryResult

	cutOffDate := time.Date(2023, 9, 27, 23, 59, 59, 0, time.UTC).Add(time.Hour*24*30).AddDate(0, 0, -30)

	if dashboardId.Validators != nil {
		validatorList := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			validatorList = append(validatorList, validator.Index)
		}

		query := `select epoch_start, 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
			epoch_start,
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
				from  validator_dashboard_data_daily
			WHERE day > $1 AND validator_index = ANY($2)
		) as a ORDER BY epoch_start, group_id;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, validatorList)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	} else {
		query := `select epoch_start, group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
			epoch_start, 
				group_id,
				SUM(attestations_reward)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				SUM(blocks_proposed)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				SUM(sync_executed)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
				from users_val_dashboards_validators 
			join validator_dashboard_data_daily on validator_dashboard_data_daily.validator_index = users_val_dashboards_validators.validator_index
			where day > $1 AND dashboard_id = $2
			group by 1, 2
		) as a ORDER BY epoch_start, group_id;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, dashboardId.Id)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	epochsMap := make(map[uint64]bool)
	groups := make(map[uint64]bool)
	data := make(map[uint64]map[uint64]float64)
	for _, row := range queryResults {
		epochsMap[row.StartEpoch] = true
		groups[row.GroupId] = true

		if data[row.StartEpoch] == nil {
			data[row.StartEpoch] = make(map[uint64]float64)
		}
		data[row.StartEpoch][row.GroupId] = d.calculateTotalEfficiency(row.AttestationEfficiency, row.ProposerEfficiency, row.SyncEfficiency)
	}

	epochsArray := make([]uint64, 0, len(epochsMap))
	for epoch := range epochsMap {
		epochsArray = append(epochsArray, epoch)
	}
	sort.Slice(epochsArray, func(i, j int) bool {
		return epochsArray[i] < epochsArray[j]
	})

	groupsArray := make([]uint64, 0, len(groups))
	for group := range groups {
		groupsArray = append(groupsArray, group)
	}
	sort.Slice(groupsArray, func(i, j int) bool {
		return groupsArray[i] < groupsArray[j]
	})

	ret.Categories = epochsArray
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[uint64]*t.ChartSeries[int, float64])
	for group := range groups {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(epochsMap)),
		}
		seriesMap[group] = &series
	}

	for _, epoch := range epochsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[epoch][group])
		}
	}

	for _, series := range seriesMap {
		ret.Series = append(ret.Series, *series)
	}

	sort.Slice(ret.Series, func(i, j int) bool {
		return ret.Series[i].Id < ret.Series[j].Id
	})

	return ret, nil
}

// allowed periods are: all_time, last_24h, last_7d, last_30d
func (d *DataAccessService) GetValidatorDashboardValidatorIndices(dashboardId t.VDBId, groupId int64, duty enums.ValidatorDuty, period enums.TimePeriod) ([]uint64, error) {
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators in case a dashboard id is provided
		validatorsQuery := `
		SELECT 
			validator_index
		FROM users_val_dashboards_validators
		WHERE dashboard_id = $1
		`
		validatorsParams := []interface{}{dashboardId.Id}

		if groupId != t.AllGroups {
			validatorsQuery += " AND group_id = $2"
			validatorsParams = append(validatorsParams, groupId)
		}
		err := d.alloyReader.Select(&validators, validatorsQuery, validatorsParams...)
		if err != nil {
			return nil, err
		}
	} else {
		// In case a list of validators is provided use them
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return []uint64{}, nil
	}

	if duty == enums.ValidatorDuties.None {
		// If we don't need to filter by duty return all validators in the dashboard and group
		return validators, nil
	}

	// Get the table name based on the period
	tableName := ""
	switch period {
	case enums.TimePeriods.AllTime:
		tableName = "validator_dashboard_data_rolling_total"
	case enums.TimePeriods.Last24h:
		tableName = "validator_dashboard_data_rolling_daily"
	case enums.TimePeriods.Last7d:
		tableName = "validator_dashboard_data_rolling_weekly"
	case enums.TimePeriods.Last30d:
		tableName = "validator_dashboard_data_rolling_monthly"
	}

	// Get the column condition based on the duty
	columnCond := ""
	switch duty {
	case enums.ValidatorDuties.Sync:
		columnCond = "sync_scheduled > 0"
	case enums.ValidatorDuties.Proposal:
		columnCond = "blocks_scheduled > 0"
	case enums.ValidatorDuties.Slashed:
		// TODO: Wait for slashings to be available in the database
		// columnCond = "(slashed OR slashings_executed > 0)"
		columnCond = "slashed"
	}

	// Get ALL validator indices for the given filters
	query := fmt.Sprintf(`
		SELECT
			validator_index
		FROM %s
		WHERE validator_index = ANY($1) AND %s`, tableName, columnCond)

	var result []uint64
	err := d.alloyReader.Select(&result, query, pq.Array(validators))
	return result, err
}
