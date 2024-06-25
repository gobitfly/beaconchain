package types

import (
	"github.com/shopspring/decimal"
)

// ------------------------------------------------------------
// Overview
type VDBOverviewValidators struct {
	Online  uint64 `json:"online"`
	Offline uint64 `json:"offline"`
	Pending uint64 `json:"pending"`
	Exited  uint64 `json:"exited"`
	Slashed uint64 `json:"slashed"`
}

type VDBOverviewGroup struct {
	Id    uint64 `json:"id"`
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

type VDBOverviewData struct {
	Name       string                                     `json:"name,omitempty"`
	Groups     []VDBOverviewGroup                         `json:"groups"`
	Validators VDBOverviewValidators                      `json:"validators"`
	Efficiency PeriodicValues[float64]                    `json:"efficiency"`
	Rewards    PeriodicValues[ClElValue[decimal.Decimal]] `json:"rewards"`
	Apr        PeriodicValues[ClElValue[float64]]         `json:"apr"`
}

type InternalGetValidatorDashboardResponse ApiDataResponse[VDBOverviewData]

// ------------------------------------------------------------
// Summary Tab

type VDBSummaryStatus struct {
	UpcomingSyncCount uint64 `json:"next_sync_count"`
	CurrentSyncCount  uint64 `json:"current_sync_count"`
	SlashedCount      uint64 `json:"slashed_count"`
}
type VDBSummaryValidators struct {
	Online  uint64 `json:"online"`
	Offline uint64 `json:"offline"`
	Exited  uint64 `json:"exited"`
}

type VDBSummaryTableRow struct {
	GroupId                  int64                      `json:"group_id"`
	Status                   VDBSummaryStatus           `json:"status"`
	Validators               VDBSummaryValidators       `json:"validators"`
	Efficiency               float64                    `json:"efficiency"`
	AverageNetworkEfficiency float64                    `json:"average_network_efficiency"`
	Attestations             StatusCount                `json:"attestations"`
	Proposals                StatusCount                `json:"proposals"`
	Reward                   ClElValue[decimal.Decimal] `json:"reward" faker:"cl_el_eth"`
}
type InternalGetValidatorDashboardSummaryResponse ApiPagingResponse[VDBSummaryTableRow]

type VDBGroupSummaryColumnItem struct {
	StatusCount StatusCount `json:"status_count"`
	Validators  []uint64    `json:"validators,omitempty"`
}

type VDBGroupSummarySyncCount struct {
	CurrentValidators  uint64 `json:"current_validators"`
	UpcomingValidators uint64 `json:"upcoming_validators"`
	PastPeriods        uint64 `json:"past_periods"`
}

type VDBGroupSummaryMissedRewards struct {
	ProposerRewards ClElValue[decimal.Decimal] `json:"proposer_rewards"`
	Attestations    decimal.Decimal            `json:"attestations"`
	Sync            decimal.Decimal            `json:"sync"`
}
type VDBGroupSummaryData struct {
	AttestationsHead       StatusCount `json:"attestations_head"`
	AttestationsSource     StatusCount `json:"attestations_source"`
	AttestationsTarget     StatusCount `json:"attestations_target"`
	AttestationEfficiency  float64     `json:"attestation_efficiency"`
	AttestationAvgInclDist float64     `json:"attestation_avg_incl_dist"`

	SyncCommittee      VDBGroupSummaryColumnItem    `json:"sync"`
	SyncCommitteeCount VDBGroupSummarySyncCount     `json:"sync_count"`
	Slashings          VDBGroupSummaryColumnItem    `json:"slashings"` // Failed slashings are count of validators in the group that were slashed
	ProposalValidators []uint64                     `json:"proposal_validators"`
	MissedRewards      VDBGroupSummaryMissedRewards `json:"missed_rewards" faker:"missed_rewards"`

	Apr ClElValue[float64] `json:"apr"`

	Luck Luck `json:"luck"`
}
type InternalGetValidatorDashboardGroupSummaryResponse ApiDataResponse[VDBGroupSummaryData]

type InternalGetValidatorDashboardSummaryChartResponse ApiDataResponse[ChartData[int, float64]] // line chart, series id is group id

// ------------------------------------------------------------
// Summary Validators
type VDBSummaryValidator struct {
	Index       uint64   `json:"index"`
	DutyObjects []uint64 `json:"duty_objects,omitempty"`
}
type VDBSummaryValidatorsData struct {
	Category   string                `json:"category" tstype:"'online' | 'offline' | 'pending' | 'deposited' | 'sync_current' | 'sync_upcoming' | 'sync_past' | 'has_slashed' | 'got_slashed' | 'proposal_proposed' | 'proposal_missed'" faker:"oneof: online, offline, pending, deposited, sync_current, sync_upcoming, sync_past, has_slashed, got_slashed, proposal_proposed, proposal_missed"`
	Validators []VDBSummaryValidator `json:"validators"`
}

type InternalGetValidatorDashboardSummaryValidatorsResponse ApiDataResponse[[]VDBSummaryValidatorsData]

// ------------------------------------------------------------
// Rewards Tab
type VDBRewardsTableDuty struct {
	Attestation *float64 `json:"attestation,omitempty"`
	Proposal    *float64 `json:"proposal,omitempty"`
	Sync        *float64 `json:"sync,omitempty"`
	Slashing    *uint64  `json:"slashing,omitempty"`
}

type VDBRewardsTableRow struct {
	Epoch   uint64                     `json:"epoch"`
	Duty    VDBRewardsTableDuty        `json:"duty"`
	GroupId int64                      `json:"group_id"`
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
	Inactivity         VDBGroupRewardsDetails `json:"inactivity"`
	Proposal           VDBGroupRewardsDetails `json:"proposal"`

	ProposalElReward            decimal.Decimal `json:"proposal_el_reward"`
	ProposalClAttIncReward      decimal.Decimal `json:"proposal_cl_att_inc_reward"`
	ProposalClSyncIncReward     decimal.Decimal `json:"proposal_cl_sync_inc_reward"`
	ProposalClSlashingIncReward decimal.Decimal `json:"proposal_cl_slashing_inc_reward"`
}
type InternalGetValidatorDashboardGroupRewardsResponse ApiDataResponse[VDBGroupRewardsData]

type InternalGetValidatorDashboardRewardsChartResponse ApiDataResponse[ChartData[int, decimal.Decimal]] // bar chart, series id is group id, property is 'el' or 'cl'

// Duties Modal

type VDBEpochDutiesTableRow struct {
	Validator uint64                 `json:"validator"`
	Duties    ValidatorHistoryDuties `json:"duties"`
}
type InternalGetValidatorDashboardDutiesResponse ApiPagingResponse[VDBEpochDutiesTableRow]

// ------------------------------------------------------------
// Blocks Tab
type VDBBlocksTableRow struct {
	Proposer        uint64                      `json:"proposer"`
	GroupId         uint64                      `json:"group_id"`
	Epoch           uint64                      `json:"epoch"`
	Slot            uint64                      `json:"slot"`
	Status          string                      `json:"status" tstype:"'success' | 'missed' | 'orphaned' | 'scheduled'" faker:"oneof: success, missed, orphaned, scheduled"`
	Block           *uint64                     `json:"block,omitempty"`
	RewardRecipient *Address                    `json:"reward_recipient,omitempty"`
	Reward          *ClElValue[decimal.Decimal] `json:"reward,omitempty"`
	Graffiti        *string                     `json:"graffiti,omitempty"`
}
type InternalGetValidatorDashboardBlocksResponse ApiPagingResponse[VDBBlocksTableRow]

// ------------------------------------------------------------
// Heatmap Tab

type VDBHeatmapEvents struct {
	Proposal bool `json:"proposal"`
	Slash    bool `json:"slash"`
	Sync     bool `json:"sync"`
}
type VDBHeatmapCell struct {
	X int64  `json:"x"` // Timestamp
	Y uint64 `json:"y"` // Group ID

	Value  float64           `json:"value"` // Attestaton Rewards
	Events *VDBHeatmapEvents `json:"events,omitempty"`
}
type VDBHeatmap struct {
	Timestamps  []int64          `json:"timestamps"` // X-Axis Categories (unix timestamp)
	GroupIds    []uint64         `json:"group_ids"`  // Y-Axis Categories
	Data        []VDBHeatmapCell `json:"data"`
	Aggregation string           `json:"aggregation" tstype:"'epoch' | 'day'" faker:"oneof: epoch, day"`
}
type InternalGetValidatorDashboardHeatmapResponse ApiDataResponse[VDBHeatmap]

type VDBHeatmapTooltipData struct {
	Timestamp int64 `json:"timestamp"` // epoch or day

	Proposers StatusCount `json:"proposers"`
	Syncs     uint64      `json:"syncs"`
	Slashings StatusCount `json:"slashings"`

	AttestationsHead      StatusCount     `json:"attestations_head"`
	AttestationsSource    StatusCount     `json:"attestations_source"`
	AttestationsTarget    StatusCount     `json:"attestations_target"`
	AttestationIncome     decimal.Decimal `json:"attestation_income"`
	AttestationEfficiency float64         `json:"attestation_efficiency"`
}
type InternalGetValidatorDashboardGroupHeatmapResponse ApiDataResponse[VDBHeatmapTooltipData]

// ------------------------------------------------------------
// Deposits Tab
type VDBExecutionDepositsTableRow struct {
	PublicKey            PubKey          `json:"public_key"`
	Index                *uint64         `json:"index,omitempty"`
	GroupId              uint64          `json:"group_id"`
	Block                uint64          `json:"block"`
	Timestamp            int64           `json:"timestamp"`
	From                 Address         `json:"from"`
	Depositor            Address         `json:"depositor"`
	TxHash               Hash            `json:"tx_hash"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
	Amount               decimal.Decimal `json:"amount"`
	Valid                bool            `json:"valid"`
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

type VDBTotalExecutionDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type InternalGetValidatorDashboardTotalExecutionDepositsResponse ApiDataResponse[VDBTotalExecutionDepositsData]

type VDBTotalConsensusDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type InternalGetValidatorDashboardTotalConsensusDepositsResponse ApiDataResponse[VDBTotalConsensusDepositsData]

// ------------------------------------------------------------
// Withdrawals Tab
type VDBWithdrawalsTableRow struct {
	Epoch             uint64          `json:"epoch"`
	Slot              uint64          `json:"slot"`
	Index             uint64          `json:"index"`
	GroupId           uint64          `json:"group_id"`
	Recipient         Address         `json:"recipient"`
	Amount            decimal.Decimal `json:"amount"`
	IsMissingEstimate bool            `json:"is_missing_estimate"`
}
type InternalGetValidatorDashboardWithdrawalsResponse ApiPagingResponse[VDBWithdrawalsTableRow]

type VDBTotalWithdrawalsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type InternalGetValidatorDashboardTotalWithdrawalsResponse ApiDataResponse[VDBTotalWithdrawalsData]

// ------------------------------------------------------------
// Manage Modal
type VDBManageValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status" tstype:"'pending' | 'online' | 'offline' | 'exiting' | 'exited' | 'slashed' | 'withdrawn'" faker:"oneof: pending, online, offline, exiting, exited, slashed, withdrawn"`
	QueuePosition        *uint64         `json:"queue_position,omitempty"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
}

type InternalGetValidatorDashboardValidatorsResponse ApiPagingResponse[VDBManageValidatorsTableRow]

// ------------------------------------------------------------
// Misc.
type VDBPostReturnData struct {
	Id        uint64 `db:"id" json:"id"`
	UserID    uint64 `db:"user_id" json:"user_id"`
	Name      string `db:"name" json:"name"`
	Network   uint64 `db:"network" json:"network"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type VDBPostCreateGroupData struct {
	Id   uint64 `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type VDBPostValidatorsData struct {
	PublicKey string `json:"public_key"`
	GroupId   uint64 `json:"group_id"`
}
