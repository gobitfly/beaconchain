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

type VDBOverviewBalances struct {
	Total     decimal.Decimal `json:"total"`
	Effective decimal.Decimal `json:"effective"`
	StakedEth decimal.Decimal `json:"staked_eth"`
}

type VDBOverviewData struct {
	Name                string                                     `json:"name,omitempty" extensions:"x-order=1"`
	Network             uint64                                     `json:"network"`
	Groups              []VDBOverviewGroup                         `json:"groups"`
	Validators          VDBOverviewValidators                      `json:"validators"`
	Efficiency          PeriodicValues[float64]                    `json:"efficiency"`
	Rewards             PeriodicValues[ClElValue[decimal.Decimal]] `json:"rewards"`
	Apr                 PeriodicValues[ClElValue[float64]]         `json:"apr"`
	ChartHistorySeconds ChartHistorySeconds                        `json:"chart_history_seconds"`
	Balances            VDBOverviewBalances                        `json:"balances"`
}

type GetValidatorDashboardResponse ApiDataResponse[VDBOverviewData]

type VDBPostArchivingReturnData struct {
	Id         uint64 `db:"id" json:"id"`
	IsArchived bool   `db:"is_archived" json:"is_archived"`
}

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
	GroupId                  int64                      `json:"group_id" extensions:"x-order=1"`
	Status                   VDBSummaryStatus           `json:"status"`
	Validators               VDBSummaryValidators       `json:"validators"`
	Efficiency               float64                    `json:"efficiency"`
	AverageNetworkEfficiency float64                    `json:"average_network_efficiency"`
	Attestations             StatusCount                `json:"attestations"`
	Proposals                StatusCount                `json:"proposals"`
	Reward                   ClElValue[decimal.Decimal] `json:"reward" faker:"cl_el_eth"`
}
type GetValidatorDashboardSummaryResponse ApiPagingResponse[VDBSummaryTableRow]

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
	ProposerRewards ClElValue[decimal.Decimal] `json:"proposer_rewards" faker:"cl_el_eth"`
	Attestations    decimal.Decimal            `json:"attestations" faker:"eth"`
	Sync            decimal.Decimal            `json:"sync" faker:"eth"`
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
	MissedRewards      VDBGroupSummaryMissedRewards `json:"missed_rewards"`

	Apr ClElValue[float64] `json:"apr"`

	Luck Luck `json:"luck"`

	RocketPool struct {
		Minipools  uint64  `json:"minipools"`
		Collateral float64 `json:"collateral"`
	} `json:"rocket_pool,omitempty"`
}
type GetValidatorDashboardGroupSummaryResponse ApiDataResponse[VDBGroupSummaryData]

type GetValidatorDashboardSummaryChartResponse ApiDataResponse[ChartData[int, float64]] // line chart, series id is group id

// ------------------------------------------------------------
// Summary Validators
type VDBSummaryValidator struct {
	Index       uint64   `json:"index" extensions:"x-order=1"`
	DutyObjects []uint64 `json:"duty_objects,omitempty"`
}
type VDBSummaryValidatorsData struct {
	Category   string                `json:"category" tstype:"'deposited' | 'online' | 'offline' | 'slashing' | 'slashed' | 'exited' | 'withdrawn' | 'pending' | 'exiting' | 'withdrawing' | 'sync_current' | 'sync_upcoming' | 'sync_past' | 'has_slashed' | 'got_slashed' | 'proposal_proposed' | 'proposal_missed'" faker:"oneof: deposited, online, offline, slashing, slashed, exited, withdrawn, pending, exiting, withdrawing, sync_current, sync_upcoming, sync_past, has_slashed, got_slashed, proposal_proposed, proposal_missed"`
	Validators []VDBSummaryValidator `json:"validators"`
}

type GetValidatorDashboardSummaryValidatorsResponse ApiDataResponse[[]VDBSummaryValidatorsData]

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

type GetValidatorDashboardRewardsResponse ApiPagingResponse[VDBRewardsTableRow]

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
type GetValidatorDashboardGroupRewardsResponse ApiDataResponse[VDBGroupRewardsData]

type GetValidatorDashboardRewardsChartResponse ApiDataResponse[ChartData[int, decimal.Decimal]] // bar chart, series id is group id, property is 'el' or 'cl'

// Duties Modal

type VDBEpochDutiesTableRow struct {
	Validator uint64                 `json:"validator" extensions:"x-order=1"`
	Duties    ValidatorHistoryDuties `json:"duties"`
}
type GetValidatorDashboardDutiesResponse ApiPagingResponse[VDBEpochDutiesTableRow]

// ------------------------------------------------------------
// Blocks Tab
type VDBBlocksTableRow struct {
	Proposer        uint64                      `json:"proposer" extensions:"x-order=1"`
	GroupId         uint64                      `json:"group_id" extensions:"x-order=2"`
	Epoch           uint64                      `json:"epoch" extensions:"x-order=3"`
	Slot            uint64                      `json:"slot" extensions:"x-order=4"`
	Block           *uint64                     `json:"block,omitempty" extensions:"x-order=5"`
	Status          string                      `json:"status" tstype:"'success' | 'missed' | 'orphaned' | 'scheduled'" faker:"oneof: success, missed, orphaned, scheduled"`
	RewardRecipient *Address                    `json:"reward_recipient,omitempty"`
	Reward          *ClElValue[decimal.Decimal] `json:"reward,omitempty"`
	Graffiti        *string                     `json:"graffiti,omitempty"`
}
type GetValidatorDashboardBlocksResponse ApiPagingResponse[VDBBlocksTableRow]

// ------------------------------------------------------------
// Heatmap Tab

type VDBHeatmapEvents struct {
	Proposal bool `json:"proposal"`
	Slash    bool `json:"slash"`
	Sync     bool `json:"sync"`
}
type VDBHeatmapCell struct {
	X int64  `json:"x" extensions:"x-order=1"` // Timestamp
	Y uint64 `json:"y" extensions:"x-order=2"` // Group ID

	Value  float64           `json:"value" extensions:"x-order=3"` // Attestaton Rewards
	Events *VDBHeatmapEvents `json:"events,omitempty"`
}
type VDBHeatmap struct {
	Timestamps  []int64          `json:"timestamps" extensions:"x-order=1"` // X-Axis Categories (unix timestamp)
	GroupIds    []uint64         `json:"group_ids" extensions:"x-order=2"`  // Y-Axis Categories
	Data        []VDBHeatmapCell `json:"data" extensions:"x-order=3"`
	Aggregation string           `json:"aggregation" tstype:"'epoch' | 'hourly' | 'daily' | 'weekly'" faker:"oneof: epoch, hourly, daily, weekly"`
}
type GetValidatorDashboardHeatmapResponse ApiDataResponse[VDBHeatmap]

type VDBHeatmapTooltipData struct {
	Timestamp int64 `json:"timestamp" extensions:"x-order=1"`

	Proposers StatusCount `json:"proposers"`
	Syncs     uint64      `json:"syncs"`
	Slashings StatusCount `json:"slashings"`

	AttestationsHead      StatusCount     `json:"attestations_head"`
	AttestationsSource    StatusCount     `json:"attestations_source"`
	AttestationsTarget    StatusCount     `json:"attestations_target"`
	AttestationIncome     decimal.Decimal `json:"attestation_income"`
	AttestationEfficiency float64         `json:"attestation_efficiency"`
}
type GetValidatorDashboardGroupHeatmapResponse ApiDataResponse[VDBHeatmapTooltipData]

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
type GetValidatorDashboardExecutionLayerDepositsResponse ApiPagingResponse[VDBExecutionDepositsTableRow]

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
type GetValidatorDashboardConsensusLayerDepositsResponse ApiPagingResponse[VDBConsensusDepositsTableRow]

type VDBTotalExecutionDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type GetValidatorDashboardTotalExecutionDepositsResponse ApiDataResponse[VDBTotalExecutionDepositsData]

type VDBTotalConsensusDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type GetValidatorDashboardTotalConsensusDepositsResponse ApiDataResponse[VDBTotalConsensusDepositsData]

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
type GetValidatorDashboardWithdrawalsResponse ApiPagingResponse[VDBWithdrawalsTableRow]

type VDBTotalWithdrawalsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type GetValidatorDashboardTotalWithdrawalsResponse ApiDataResponse[VDBTotalWithdrawalsData]

// ------------------------------------------------------------
// Rocket Pool Tab
type VDBRocketPoolTableRow struct {
	Node   Address `json:"node" extensions:"x-order=1"`
	Staked struct {
		Eth decimal.Decimal `json:"eth"`
		Rpl decimal.Decimal `json:"rpl"`
	} `json:"staked"`
	Minipools struct {
		Total uint64 `json:"total"`
		Leb16 uint64 `json:"leb_16"`
		Leb8  uint64 `json:"leb_8"`
	} `json:"minipools"`
	Collateral    PercentageDetails[decimal.Decimal] `json:"collateral"`
	AvgCommission float64                            `json:"avg_commission"`
	Rpl           struct {
		Claimed   decimal.Decimal `json:"claimed"`
		Unclaimed decimal.Decimal `json:"unclaimed"`
	} `json:"rpl"`
	EffectiveRpl   decimal.Decimal `json:"effective_rpl"`
	RplApr         float64         `json:"rpl_apr"`
	RplAprUpdateTs int64           `json:"rpl_apr_update_ts"`
	RplEstimate    decimal.Decimal `json:"rpl_estimate"`
	SmoothingPool  struct {
		IsOptIn   bool            `json:"is_opt_in"`
		Claimed   decimal.Decimal `json:"claimed"`
		Unclaimed decimal.Decimal `json:"unclaimed"`
	} `json:"smoothing_pool"`
}
type GetValidatorDashboardRocketPoolResponse ApiPagingResponse[VDBRocketPoolTableRow]

type GetValidatorDashboardTotalRocketPoolResponse ApiDataResponse[VDBRocketPoolTableRow]

type VDBNodeRocketPoolData struct {
	Timezone      string          `json:"timezone"`
	RefundBalance decimal.Decimal `json:"refund_balance"`
	DepositCredit decimal.Decimal `json:"deposit_credit"`
	RplStake      struct {
		Min decimal.Decimal `json:"min"`
		Max decimal.Decimal `json:"max"`
	} `json:"rpl_stake"`
}

type GetValidatorDashboardNodeRocketPoolResponse ApiDataResponse[VDBNodeRocketPoolData]

type VDBRocketPoolMinipoolsTableRow struct {
	Node             Address         `json:"node"`
	ValidatorIndex   uint64          `json:"validator_index"`
	MinipoolStatus   string          `json:"minipool_status" tstype:"'initialized' | 'prelaunch' | 'staking' | 'withdrawable' | 'dissolved'" faker:"oneof: initialized, prelaunch, staking, withdrawable, dissolved"`
	ValidatorStatus  string          `json:"validator_status" tstype:"'slashed' | 'exited' | 'deposited' | 'pending' | 'slashing_offline' | 'slashing_online' | 'exiting_offline' | 'exiting_online' | 'active_offline' | 'active_online'" faker:"oneof: slashed, exited, deposited, pending, slashing_offline, slashing_online, exiting_offline, exiting_online, active_offline, active_online"`
	GroupId          uint64          `json:"group_id"`
	Deposit          decimal.Decimal `json:"deposit"`
	Commission       float64         `json:"commission"`
	CreatedTimestamp int64           `json:"created_timestamp"`
	Penalties        uint64          `json:"penalties"`
}
type GetValidatorDashboardRocketPoolMinipoolsResponse ApiPagingResponse[VDBRocketPoolMinipoolsTableRow]

// ------------------------------------------------------------
// Manage Modal
type VDBManageValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status" tstype:"'slashed' | 'exited' | 'deposited' | 'pending' | 'slashing_offline' | 'slashing_online' | 'exiting_offline' | 'exiting_online' | 'active_offline' | 'active_online'" faker:"oneof: slashed, exited, deposited, pending, slashing_offline, slashing_online, exiting_offline, exiting_online, active_offline, active_online"`
	QueuePosition        *uint64         `json:"queue_position,omitempty"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
}

type GetValidatorDashboardValidatorsResponse ApiPagingResponse[VDBManageValidatorsTableRow]

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
