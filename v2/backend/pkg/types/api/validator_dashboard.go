package apitypes

import (
	"github.com/shopspring/decimal"
)

// ------------------------------------------------------------
// Overview

// ------------------------------------------------------------
// Slot Viz

type VDBSlotVizResponse struct {
	Data []VDBSlotVizEpoch `json:"data"`
}

type VDBSlotVizEpoch struct {
	Epoch uint64           `json:"epoch"`
	State string           `json:"state" ts_type:"'head' | 'finalized' | 'scheduled'"`
	Slots []VDBSlotVizSlot `json:"slots"`
}

type VDBSlotVizSlot struct {
	Slot   uint64           `json:"slot"`
	Status string           `json:"status" ts_type:"'proposed' | 'missed' | 'scheduled' | 'orphaned'"`
	Duties []VDBSlotVizDuty `json:"duties"`
}

type VDBSlotVizDuty struct {
	Type            string          `json:"type" ts_type:"'proposal' | 'attestation' | 'sync' | 'slashing'"`
	PendingCount    uint64          `json:"pending_count"`
	SuccessCount    uint64          `json:"success_count"`
	SuccessEarnings decimal.Decimal `json:"success_earnings"`
	FailedCount     uint64          `json:"failed_count"`
	FailedEarnings  decimal.Decimal `json:"failed_earnings"`
	Block           uint64          `json:"block,omitempty"`
	Validator       uint64          `json:"validator,omitempty"`
}

// ------------------------------------------------------------
// Summary Tab

type VDBSummaryTableResponse struct {
	Paging Paging               `json:"paging"`
	Data   []VDBSummaryTableRow `json:"data"`
}

type VDBSummaryTableRow struct {
	GroupId uint64 `json:"group_id"`

	Efficiency24h float64 `json:"efficiency_24h"`
	Efficiency7d  float64 `json:"efficiency_7d"`
	Efficiency31d float64 `json:"efficiency_31d"`
	EfficiencyAll float64 `json:"efficiency_all"`

	Validators []uint64 `json:"validators"`
}

type VDBGroupSummaryResponse struct {
	Details24h VDBGroupSummaryColumn `json:"details_24h"`
	Details7d  VDBGroupSummaryColumn `json:"details_7d"`
	Details31d VDBGroupSummaryColumn `json:"details_31d"`
	DetailsAll VDBGroupSummaryColumn `json:"details_all"`
}

type VDBGroupSummaryColumn struct {
	AttestationsHead       VDBGroupSummaryColumnItem `json:"attestation_head"`
	AttestationsSource     VDBGroupSummaryColumnItem `json:"attestation_source"`
	AttestationsTarget     VDBGroupSummaryColumnItem `json:"attestation_target"`
	AttestationEfficiency  float64                   `json:"attestation_efficiency"`
	AttestationAvgInclDist float64                   `json:"attestation_avg_incl_dist"`

	SyncCommittee VDBGroupSummaryColumnItem `json:"sync"`
	Proposals     VDBGroupSummaryColumnItem `json:"proposals"`
	Slashed       VDBGroupSummaryColumnItem `json:"slashed"`

	Apr    ClElValueFloat `json:"apr"`
	Income ClElValue      `json:"income"`

	ProposalLuck Luck `json:"proposal_luck"`
	SyncLuck     Luck `json:"sync_luck"`
}

type VDBGroupSummaryColumnItem struct {
	StatusCount StatusCount     `json:"status_count"`
	Earned      decimal.Decimal `json:"earned"`
	Penalty     decimal.Decimal `json:"penalty"`
	Validators  []uint64        `json:"validators,omitempty"`
}

// ------------------------------------------------------------
// Rewards Tab

type VDBRewardsTableResponse struct {
	Paging Paging               `json:"paging"`
	Data   []VDBRewardsTableRow `json:"data"`
}

type VDBRewardsTableRow struct {
	Epoch   uint64    `json:"epoch"`
	GroupId uint64    `json:"group_id"`
	Reward  ClElValue `json:"reward"`
}

type VDBGroupRewardsResponse struct {
	AttestationSource VDBGroupRewardsDetails `json:"attestation_source"`
	AttestationTarget VDBGroupRewardsDetails `json:"attestation_target"`
	AttestationHead   VDBGroupRewardsDetails `json:"attestation_head"`
	Sync              VDBGroupRewardsDetails `json:"sync"`
	Slashing          VDBGroupRewardsDetails `json:"slashing"`
	Proposal          VDBGroupRewardsDetails `json:"proposal"`
	ProposalElReward  decimal.Decimal        `json:"proposal_el_reward"`
}

type VDBGroupRewardsDetails struct {
	StatusCount StatusCount     `json:"status_count"`
	Income      decimal.Decimal `json:"income"`
}

// Duties Modal

type VDBEpochDutiesTableResponse struct {
	Paging Paging                   `json:"paging"`
	Data   []VDBEpochDutiesTableRow `json:"data"`
}

type VDBEpochDutiesTableRow struct {
	Validator uint64 `json:"validator"`

	AttestationSource VDBEpochDutiesItem `json:"attestation_source"`
	AttestationTarget VDBEpochDutiesItem `json:"attestation_target"`
	AttestationHead   VDBEpochDutiesItem `json:"attestation_head"`
	Proposal          VDBEpochDutiesItem `json:"proposal"`
	Sync              VDBEpochDutiesItem `json:"sync"`
	Slashing          VDBEpochDutiesItem `json:"slashing"`
}

type VDBEpochDutiesItem struct {
	Status string          `json:"status" ts_type:"'success' | 'partial' | 'failed' | 'orphaned'"`
	Reward decimal.Decimal `json:"reward"`
}

// ------------------------------------------------------------
// Blocks Tab

type VDBBlocksTableResponse struct {
	Paging Paging              `json:"paging"`
	Data   []VDBBlocksTableRow `json:"data"`
}

type VDBBlocksTableRow struct {
	Proposer uint64    `json:"proposer"`
	GroupId  uint64    `json:"group_id"`
	Epoch    uint64    `json:"epoch"`
	Slot     uint64    `json:"slot"`
	Block    uint64    `json:"block"`
	Status   string    `json:"status" ts_type:"'success' | 'missed' | 'orphaned' | 'scheduled'"`
	Reward   ClElValue `json:"reward"`
}

// ------------------------------------------------------------
// Heatmap Tab

// TODO Highcharts Object

type VDBHeatmapTooltipResponse struct {
	Epoch uint64 `json:"epoch"`

	Proposers []VDBHeatmapTooltipDuty `json:"proposers"`
	Syncs     []VDBHeatmapTooltipDuty `json:"syncs"`
	Slashings []VDBHeatmapTooltipDuty `json:"slashings"`

	AttestationHead   StatusCount     `json:"attestation_head"`
	AttestationSource StatusCount     `json:"attestation_source"`
	AttestationTarget StatusCount     `json:"attestation_target"`
	AttestationIncome decimal.Decimal `json:"attestation_income"`
}

type VDBHeatmapTooltipDuty struct {
	Validator uint64 `json:"validator"`
	Status    string `json:"status" ts_type:"'success' | 'failed' | 'orphaned'"`
}

// ------------------------------------------------------------
// Deposits Tab

type VDBExecutionDepositsTableResponse struct {
	Paging Paging                         `json:"paging"`
	Data   []VDBExecutionDepositsTableRow `json:"data"`
}

type VDBExecutionDepositsTableRow struct {
	PublicKey             PubKey          `json:"public_key"`
	Index                 uint64          `json:"index"`
	GroupId               uint64          `json:"group_id"`
	Block                 uint64          `json:"block"`
	From                  Address         `json:"from"`
	Depositor             Address         `json:"depositor"`
	TxHash                Hash            `json:"tx_hash"`
	WithdrawalCredentials Hash            `json:"withdrawal_credentials"`
	Amount                decimal.Decimal `json:"amount"`
	Valid                 bool            `json:"valid"`
}

type VDBConsensusDepositsTableResponse struct {
	Paging Paging                         `json:"paging"`
	Data   []VDBConsensusDepositsTableRow `json:"data"`
}

type VDBConsensusDepositsTableRow struct {
	PublicKey            PubKey          `json:"public_key"`
	Index                uint64          `json:"index"`
	GroupId              uint64          `json:"group_id"`
	Epoch                uint64          `json:"epoch"`
	Slot                 uint64          `json:"slot"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
	Amount               decimal.Decimal `json:"amount"`
	Signature            Hash            `json:"signature"`
}

// ------------------------------------------------------------
// Withdrawals Tab

type VDBWithdrawalsTableResponse struct {
	Paging Paging                   `json:"paging"`
	Data   []VDBWithdrawalsTableRow `json:"data"`
}

type VDBWithdrawalsTableRow struct {
	Epoch     uint64          `json:"epoch"`
	Index     uint64          `json:"index"`
	GroupId   uint64          `json:"group_id"`
	Recipient Address         `json:"recipient"`
	Amount    decimal.Decimal `json:"amount"`
}

// ------------------------------------------------------------
// Manage Modal

type VDBManageValidatorsTableResponse struct {
	Paging Paging                        `json:"paging"`
	Data   []VDBManageValidatorsTableRow `json:"data"`
}

type VDBManageValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
}
