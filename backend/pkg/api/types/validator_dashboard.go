package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// ------------------------------------------------------------
// Overview

type VDBOverviewValidators struct {
	Total   uint64 `json:"total"`
	Active  uint64 `json:"active"`
	Pending uint64 `json:"pending"`
	Exited  uint64 `json:"exited"`
	Slashed uint64 `json:"slashed"`
}

type VDBOverviewEfficiency struct {
	Total       float64 `json:"total"`
	Attestation float64 `json:"attestation"`
	Proposal    float64 `json:"proposal"`
	Sync        float64 `json:"sync"`
}

type VDBOverviewGroup struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type VDBOverviewData struct {
	Groups     []VDBOverviewGroup                  `json:"groups"`
	Validators VDBOverviewValidators               `json:"validators"`
	Efficiency VDBOverviewEfficiency               `json:"efficiency"`
	Rewards    PeriodicClElValues[decimal.Decimal] `json:"rewards"`
	Luck       Luck                                `json:"luck"`
	Apr        PeriodicClElValues[float64]         `json:"apr"`
}

type VDBOverviewResponse ApiDataResponse[VDBOverviewData]

// ------------------------------------------------------------
// Slot Viz
type VDBSlotVizSlot struct {
	Slot   uint64           `json:"slot"`
	Status string           `json:"status" tstype:"'proposed' | 'missed' | 'scheduled' | 'orphaned'"`
	Duties []VDBSlotVizDuty `json:"duties"`
}
type VDBSlotVizEpoch struct {
	Epoch uint64           `json:"epoch"`
	State string           `json:"state" tstype:"'head' | 'finalized' | 'scheduled'"`
	Slots []VDBSlotVizSlot `json:"slots"`
}

type VDBSlotVizDuty struct {
	Type            string          `json:"type" tstype:"'proposal' | 'attestation' | 'sync' | 'slashing'"`
	PendingCount    uint64          `json:"pending_count"`
	SuccessCount    uint64          `json:"success_count"`
	SuccessEarnings decimal.Decimal `json:"success_earnings"`
	FailedCount     uint64          `json:"failed_count"`
	FailedEarnings  decimal.Decimal `json:"failed_earnings"`
	Block           uint64          `json:"block,omitempty"`
	Validator       uint64          `json:"validator,omitempty"`
}

type VDBSlotVizResponse ApiDataResponse[VDBSlotVizEpoch]

// ------------------------------------------------------------
// Summary Tab
type VDBSummaryTableResponse ApiPagingResponse[VDBSummaryTableRow]

type VDBSummaryTableRow struct {
	GroupId uint64 `json:"group_id"`

	EfficiencyDay   float64 `json:"efficiency_day"`
	EfficiencyWeek  float64 `json:"efficiency_week"`
	EfficiencyMonth float64 `json:"efficiency_month"`
	EfficiencyTotal float64 `json:"efficiency_total"`

	Validators []uint64 `json:"validators"`
}

type VDBGroupSummaryResponse ApiDataResponse[VDBGroupSummaryData]

type VDBGroupSummaryData struct {
	DetailsDay   VDBGroupSummaryColumn `json:"details_day"`
	DetailsWeek  VDBGroupSummaryColumn `json:"details_week"`
	DetailsMonth VDBGroupSummaryColumn `json:"details_month"`
	DetailsTotal VDBGroupSummaryColumn `json:"details_total"`
}

type VDBGroupSummaryColumn struct {
	AttestationsHead       VDBGroupSummaryColumnItem `json:"attestation_head"`
	AttestationsSource     VDBGroupSummaryColumnItem `json:"attestation_source"`
	AttestationsTarget     VDBGroupSummaryColumnItem `json:"attestation_target"`
	AttestationEfficiency  float64                   `json:"attestation_efficiency"`
	AttestationAvgInclDist float64                   `json:"attestation_avg_incl_dist"`

	SyncCommittee VDBGroupSummaryColumnItem `json:"sync"`
	Proposals     VDBGroupSummaryColumnItem `json:"proposals"`
	Slashed       VDBGroupSummaryColumnItem `json:"slashed"` // Failed slashings are count of validators in the group that were slashed

	Apr    ClElValue[float64]         `json:"apr"`
	Income ClElValue[decimal.Decimal] `json:"income"`

	Luck Luck `json:"luck"`
}

type VDBGroupSummaryColumnItem struct {
	StatusCount StatusCount     `json:"status_count"`
	Earned      decimal.Decimal `json:"earned"`
	Penalty     decimal.Decimal `json:"penalty"`
	Validators  []uint64        `json:"validators,omitempty"`
}

type VDBSummaryChartResponse ApiDataResponse[[]HighchartsSeries]

// ------------------------------------------------------------
// Rewards Tab
type VDBRewardsTableResponse ApiPagingResponse[VDBRewardsTableRow]

type VDBRewardsTableRow struct {
	Epoch   uint64                     `json:"epoch"`
	Duty    VDBRewardesTableDuty       `json:"duty"`
	GroupId uint64                     `json:"group_id"`
	Reward  ClElValue[decimal.Decimal] `json:"reward"`
}

type VDBRewardesTableDuty struct {
	Attestation float64 `json:"attestation"`
	Proposal    float64 `json:"proposal"`
	Sync        float64 `json:"sync"`
	Slashing    uint64  `json:"slashing"`
}

type VDBGroupRewardsResponse ApiDataResponse[VDBGroupRewardsData]

type VDBGroupRewardsData struct {
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

type VDBRewardsChartResponse ApiDataResponse[[]HighchartsSeries]

// Duties Modal
type VDBEpochDutiesTableResponse ApiPagingResponse[VDBEpochDutiesTableRow]

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
	Status string          `json:"status" tstype:"'success' | 'partial' | 'failed' | 'orphaned'"`
	Reward decimal.Decimal `json:"reward"`
}

// ------------------------------------------------------------
// Blocks Tab
type VDBBlocksTableResponse ApiPagingResponse[VDBBlocksTableRow]

type VDBBlocksTableRow struct {
	Proposer uint64                     `json:"proposer"`
	GroupId  uint64                     `json:"group_id"`
	Epoch    uint64                     `json:"epoch"`
	Slot     uint64                     `json:"slot"`
	Block    uint64                     `json:"block"`
	Status   string                     `json:"status" tstype:"'success' | 'missed' | 'orphaned' | 'scheduled'"`
	Reward   ClElValue[decimal.Decimal] `json:"reward"`
}

// ------------------------------------------------------------
// Heatmap Tab
type VDBHeatmapResponse ApiDataResponse[VDBHeatmap]

type VDBHeatmap struct {
	Epochs   []uint64         `json:"epochs"`    // X-Axis Categories
	GroupIds []uint64         `json:"group_ids"` // Y-Axis Categories
	Data     []VDBHeatmapCell `json:"data"`
}

type VDBHeatmapCell struct {
	X     uint64  `json:"x" ts_doc:"Epoch"`
	Y     uint64  `json:"y" ts_doc:"Group ID"`
	Value float64 `json:"value"` // Attestaton Rewards
}

type VDBHeatmapTooltipResponse ApiDataResponse[VDBHeatmapTooltipData]

type VDBHeatmapTooltipData struct {
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
	Status    string `json:"status" tstype:"'success' | 'failed' | 'orphaned'"`
}

// ------------------------------------------------------------
// Deposits Tab
type VDBExecutionDepositsTableResponse ApiPagingResponse[VDBExecutionDepositsTableRow]

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

type VDBConsensusDepositsTableResponse ApiPagingResponse[VDBConsensusDepositsTableRow]

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
type VDBWithdrawalsTableResponse ApiPagingResponse[VDBWithdrawalsTableRow]

type VDBWithdrawalsTableRow struct {
	Epoch     uint64          `json:"epoch"`
	Index     uint64          `json:"index"`
	GroupId   uint64          `json:"group_id"`
	Recipient Address         `json:"recipient"`
	Amount    decimal.Decimal `json:"amount"`
}

// ------------------------------------------------------------
// Manage Modal
type VDBManageValidatorsTableResponse ApiPagingResponse[VDBManageValidatorsTableRow]

type VDBManageValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
}

// ------------------------------------------------------------
// Misc. Responses
type VDBPostData struct {
	Id        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	Name      string    `json:"name"`
	Network   uint64    `json:"network"`
	CreatedAt time.Time `json:"created_at"`
}
