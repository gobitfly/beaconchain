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
	Id   uint64 `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type VDBOverviewData struct {
	Groups     []VDBOverviewGroup                  `json:"groups"`
	Validators VDBOverviewValidators               `json:"validators"`
	Efficiency VDBOverviewEfficiency               `json:"efficiency"`
	Rewards    PeriodicClElValues[decimal.Decimal] `json:"rewards"`
	Luck       Luck                                `json:"luck"`
	Apr        PeriodicClElValues[float64]         `json:"apr"`
}

type InternalGetValidatorDashboardResponse ApiDataResponse[VDBOverviewData]

// ------------------------------------------------------------
// Summary Tab
type VDBSummaryTableRow struct {
	GroupId uint64 `json:"group_id"`

	EfficiencyLast24h float64 `json:"efficiency_last_24h"`
	EfficiencyLast7d  float64 `json:"efficiency_last_7d"`
	EfficiencyLast31d float64 `json:"efficiency_last_31d"`
	EfficiencyAllTime float64 `json:"efficiency_all_time"`

	Validators []uint64 `json:"validators"`
}
type InternalGetValidatorDashboardSummaryResponse ApiPagingResponse[VDBSummaryTableRow]

type VDBGroupSummaryColumnItem struct {
	StatusCount StatusCount `json:"status_count"`
	Validators  []uint64    `json:"validators,omitempty"`
}
type VDBGroupSummaryColumn struct {
	AttestationsHead       VDBGroupSummaryColumnItem `json:"attestations_head"`
	AttestationsSource     VDBGroupSummaryColumnItem `json:"attestations_source"`
	AttestationsTarget     VDBGroupSummaryColumnItem `json:"attestations_target"`
	AttestationCount       StatusCount               `json:"attestation_count"`
	AttestationEfficiency  float64                   `json:"attestation_efficiency"`
	AttestationAvgInclDist float64                   `json:"attestation_avg_incl_dist"`

	SyncCommittee VDBGroupSummaryColumnItem `json:"sync"`
	Proposals     VDBGroupSummaryColumnItem `json:"proposals"`
	Slashed       VDBGroupSummaryColumnItem `json:"slashed"` // Failed slashings are count of validators in the group that were slashed

	Apr    ClElValue[float64]         `json:"apr"`
	Income ClElValue[decimal.Decimal] `json:"income"`

	Luck Luck `json:"luck"`
}
type VDBGroupSummaryData struct {
	Last24h VDBGroupSummaryColumn `json:"last_24h"`
	Last7d  VDBGroupSummaryColumn `json:"last_7d"`
	Last31d VDBGroupSummaryColumn `json:"last_31d"`
	AllTime VDBGroupSummaryColumn `json:"all_time"`
}
type InternalGetValidatorDashboardGroupSummaryResponse ApiDataResponse[VDBGroupSummaryData]

type InternalGetValidatorDashboardSummaryChartResponse ApiDataResponse[ChartData[int]] // line chart, series id is group id, no stack

// ------------------------------------------------------------
// Rewards Tab
type VDBRewardesTableDuty struct {
	Attestation float64 `json:"attestation"`
	Proposal    float64 `json:"proposal"`
	Sync        float64 `json:"sync"`
	Slashing    uint64  `json:"slashing"`
}

type VDBRewardsTableRow struct {
	Epoch   uint64                     `json:"epoch"`
	Duty    VDBRewardesTableDuty       `json:"duty"`
	GroupId uint64                     `json:"group_id"`
	Reward  ClElValue[decimal.Decimal] `json:"reward"`
}

type InternalGetValidatorDashboardRewardsResponse ApiPagingResponse[VDBRewardsTableRow]

type VDBGroupRewardsDetails struct {
	StatusCount StatusCount     `json:"status_count"`
	Income      decimal.Decimal `json:"income"`
}
type VDBGroupRewardsData struct {
	AttestationsSource VDBGroupRewardsDetails `json:"attestations_source"`
	AttestationsTarget VDBGroupRewardsDetails `json:"attestations_target"`
	AttestationsHead   VDBGroupRewardsDetails `json:"attestations_head"`
	Sync               VDBGroupRewardsDetails `json:"sync"`
	Slashing           VDBGroupRewardsDetails `json:"slashing"`
	Proposal           VDBGroupRewardsDetails `json:"proposal"`
	ProposalElReward   decimal.Decimal        `json:"proposal_el_reward"`
}
type InternalGetValidatorDashboardGroupRewardsResponse ApiDataResponse[VDBGroupRewardsData]

type InternalGetValidatorDashboardRewardsChartResponse ApiDataResponse[ChartData[int]] // bar chart, series id is group id, stack is 'execution' or 'consensus'

// Duties Modal
type VDBEpochDutiesItem struct {
	Status string          `json:"status" tstype:"'success' | 'partial' | 'failed' | 'orphaned'"`
	Reward decimal.Decimal `json:"reward"`
}
type VDBEpochDutiesTableRow struct {
	Validator uint64 `json:"validator"`

	AttestationsSource VDBEpochDutiesItem `json:"attestations_source"`
	AttestationsTarget VDBEpochDutiesItem `json:"attestations_target"`
	AttestationsHead   VDBEpochDutiesItem `json:"attestations_head"`
	Proposal           VDBEpochDutiesItem `json:"proposal"`
	Sync               VDBEpochDutiesItem `json:"sync"`
	Slashing           VDBEpochDutiesItem `json:"slashing"`
}
type InternalGetValidatorDashboardDutiesResponse ApiPagingResponse[VDBEpochDutiesTableRow]

// ------------------------------------------------------------
// Blocks Tab
type VDBBlocksTableRow struct {
	Proposer        uint64                     `json:"proposer"`
	GroupId         uint64                     `json:"group_id"`
	Epoch           uint64                     `json:"epoch"`
	Slot            uint64                     `json:"slot"`
	Block           uint64                     `json:"block"`
	Status          string                     `json:"status" tstype:"'success' | 'missed' | 'orphaned' | 'scheduled'"`
	RewardRecipient Address                    `json:"reward_recipient"`
	Reward          ClElValue[decimal.Decimal] `json:"reward"`
	Graffiti        string                     `json:"graffiti"`
}
type InternalGetValidatorDashboardBlocksResponse ApiPagingResponse[VDBBlocksTableRow]

// ------------------------------------------------------------
// Heatmap Tab
type VDBHeatmapCell struct {
	X     uint64  `json:"x" ts_doc:"Epoch"`
	Y     uint64  `json:"y" ts_doc:"Group ID"`
	Value float64 `json:"value"` // Attestaton Rewards
}
type VDBHeatmap struct {
	Epochs   []uint64         `json:"epochs"`    // X-Axis Categories
	GroupIds []uint64         `json:"group_ids"` // Y-Axis Categories
	Data     []VDBHeatmapCell `json:"data"`
}
type InternalGetValidatorDashboardHeatmapResponse ApiDataResponse[VDBHeatmap]

type VDBHeatmapTooltipDuty struct {
	Validator uint64 `json:"validator"`
	Status    string `json:"status" tstype:"'success' | 'failed' | 'orphaned'"`
}
type VDBHeatmapTooltipData struct {
	Epoch uint64 `json:"epoch"`

	Proposers []VDBHeatmapTooltipDuty `json:"proposers"`
	Syncs     []VDBHeatmapTooltipDuty `json:"syncs"`
	Slashings []VDBHeatmapTooltipDuty `json:"slashings"`

	AttestationsHead   StatusCount     `json:"attestations_head"`
	AttestationsSource StatusCount     `json:"attestations_source"`
	AttestationsTarget StatusCount     `json:"attestations_target"`
	AttestationIncome  decimal.Decimal `json:"attestation_income"`
}
type InternalGetValidatorDashboardGroupHeatmapResponse ApiDataResponse[VDBHeatmapTooltipData]

// ------------------------------------------------------------
// Deposits Tab
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
type InternalGetValidatorDashboardExecutionLayerDepositsResponse ApiPagingResponse[VDBExecutionDepositsTableRow]

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
type InternalGetValidatorDashboardConsensusLayerDepositsResponse ApiPagingResponse[VDBConsensusDepositsTableRow]

// ------------------------------------------------------------
// Withdrawals Tab
type VDBWithdrawalsTableRow struct {
	Epoch     uint64          `json:"epoch"`
	Index     uint64          `json:"index"`
	GroupId   uint64          `json:"group_id"`
	Recipient Address         `json:"recipient"`
	Amount    decimal.Decimal `json:"amount"`
}
type InternalGetValidatorDashboardWithdrawalsResponse ApiPagingResponse[VDBWithdrawalsTableRow]

// ------------------------------------------------------------
// Manage Modal
type VDBManageValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status" tstype:"'pending' | 'online' | 'offline' | 'exiting' | 'exited' | 'slashed' | 'withdrawn'" faker:"oneof: pending, online, offline, exiting, exited, slashed, withdrawn"`
	QueuePosition        uint64          `json:"queue_position,omitempty"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
}

type InternalGetValidatorDashboardValidatorsResponse ApiPagingResponse[VDBManageValidatorsTableRow]

// ------------------------------------------------------------
// Misc.
type VDBPostReturnData struct {
	Id        uint64    `db:"id" json:"id"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Network   uint64    `db:"network" json:"network"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type VDBPostValidatorsData struct {
	PublicKey string `json:"public_key"`
	GroupId   uint64 `json:"group_id"`
}

type VDBPostPublicIdData struct {
	PublicId      string `json:"public_id"`
	Name          string `json:"name"`
	ShareSettings struct {
		GroupNames bool `json:"group_names"`
	} `json:"share_settings"`
}
