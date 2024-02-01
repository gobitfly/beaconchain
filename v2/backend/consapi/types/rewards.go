package types

import "github.com/shopspring/decimal"

// /eth/v1/beacon/rewards/attestations/{epoch}
type StandardAttestationRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		IdealRewards []struct {
			EffectiveBalance decimal.Decimal `json:"effective_balance"`
			Head             int64           `json:"head,string"`
			Target           int64           `json:"target,string"`
			Source           int64           `json:"source,string"`
			InclusionDelay   int64           `json:"inclusion_delay,string"`
			Inactivity       int64           `json:"inactivity,string"`
		} `json:"ideal_rewards"`
		TotalRewards []struct {
			ValidatorIndex Index `json:"validator_index,string"`
			Head           int64 `json:"head,string"`
			Target         int64 `json:"target,string"`
			Source         int64 `json:"source,string"`
			InclusionDelay int64 `json:"inclusion_delay,string"`
			Inactivity     int64 `json:"inactivity,string"`
		} `json:"total_rewards"`
	} `json:"data"`
}

// /eth/v1/beacon/rewards/sync_committee/{block_id}
type StandardSyncCommitteeRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		ValidatorIndex Index `json:"validator_index,string"`
		Reward         int64 `json:"reward,string"`
	} `json:"data"`
}

// /eth/v1/beacon/rewards/blocks/{block_id}
type StandardBlockRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		ProposerIndex     Index           `json:"proposer_index,string"`
		Total             decimal.Decimal `json:"total"`
		Attestations      int64           `json:"attestations,string"`
		SyncAggregate     int64           `json:"sync_aggregate,string"`
		ProposerSlashings int64           `json:"proposer_slashings,string"`
		AttesterSlashings int64           `json:"attester_slashings,string"`
	} `json:"data"`
}
