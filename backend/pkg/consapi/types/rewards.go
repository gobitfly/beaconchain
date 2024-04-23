package types

import "github.com/shopspring/decimal"

// /eth/v1/beacon/rewards/attestations/{epoch}
type StandardAttestationRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		IdealRewards []struct {
			EffectiveBalance int64 `json:"effective_balance,string"`
			Head             int32 `json:"head,string"`
			Target           int32 `json:"target,string"`
			Source           int32 `json:"source,string"`
			InclusionDelay   int32 `json:"inclusion_delay,string"`
			Inactivity       int32 `json:"inactivity,string"`
		} `json:"ideal_rewards"`
		TotalRewards []struct {
			ValidatorIndex uint64 `json:"validator_index,string"`
			Head           int32  `json:"head,string"`
			Target         int32  `json:"target,string"`
			Source         int32  `json:"source,string"`
			InclusionDelay int32  `json:"inclusion_delay,string"`
			Inactivity     int32  `json:"inactivity,string"`
		} `json:"total_rewards"`
	} `json:"data"`
}

// /eth/v1/beacon/rewards/sync_committee/{block_id}
type StandardSyncCommitteeRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		ValidatorIndex uint64 `json:"validator_index,string"`
		Reward         int64  `json:"reward,string"`
	} `json:"data"`
}

// /eth/v1/beacon/rewards/blocks/{block_id}
type StandardBlockRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		ProposerIndex     uint64          `json:"proposer_index,string"`
		Total             decimal.Decimal `json:"total"`
		Attestations      int64           `json:"attestations,string"`
		SyncAggregate     int64           `json:"sync_aggregate,string"`
		ProposerSlashings int64           `json:"proposer_slashings,string"`
		AttesterSlashings int64           `json:"attester_slashings,string"`
	} `json:"data"`
}
