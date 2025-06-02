package types

import (
	"github.com/shopspring/decimal"
)

// ------------------------------------------------------------
// Overview

type VDBOverviewGroup struct {
	Id    uint64 `json:"id"`
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

type ValidatorBalances struct {
	Total            decimal.Decimal `json:"total"`
	EffectiveCurrent decimal.Decimal `json:"effective_current"` // on-chain
	EffectiveLatest  decimal.Decimal `json:"effective_latest"`  // from premium perks pov: exited validators are counted with their latest eb
	StakedEth        decimal.Decimal `json:"staked_eth"`
}

type VDBOverviewData struct {
	Name                string                                     `json:"name,omitempty" extensions:"x-order=1"`
	Network             uint64                                     `json:"network"`
	Groups              []VDBOverviewGroup                         `json:"groups"`
	Validators          ValidatorStateCounts                       `json:"validators"`
	Efficiency          PeriodicValues[*float64]                   `json:"efficiency"`
	Rewards             PeriodicValues[ClElValue[decimal.Decimal]] `json:"rewards"`
	Apr                 PeriodicValues[ClElValue[*float64]]        `json:"apr"`
	ChartHistorySeconds ChartHistorySeconds                        `json:"chart_history_seconds"`
	Balances            ValidatorBalances                          `json:"balances"`
	IsAboveEbLimit      bool                                       `json:"is_above_effective_balance_limit"` // refers to owner; relevant for shared dashboards
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
	Efficiency               *float64                   `json:"efficiency"`
	AverageNetworkEfficiency float64                    `json:"average_network_efficiency"`
	Attestations             StatusCount                `json:"attestations"`
	Proposals                StatusCount                `json:"proposals"`
	Reward                   ClElValue[decimal.Decimal] `json:"reward" faker:"cl_el_eth"`
}
type GetValidatorDashboardSummaryResponse ApiPagingResponse[VDBSummaryTableRow]

type VDBGroupSummaryColumnItem struct {
	StatusCount    StatusCount `json:"status_count"`
	Validators     []uint64    `json:"validators,omitempty"` // fill with up to 3 validator indexes
	ValidatorCount uint64      `json:"validator_count"`      // number of distinct validators
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
	Efficiency *float64                   `json:"efficiency"`
	Balances   ValidatorBalances          `json:"balances"`
	Rewards    ClElValue[decimal.Decimal] `json:"rewards"`

	AttestationsHead       StatusCount `json:"attestations_head"`
	AttestationsSource     StatusCount `json:"attestations_source"`
	AttestationsTarget     StatusCount `json:"attestations_target"`
	AttestationEfficiency  *float64    `json:"attestation_efficiency"`
	AttestationAvgInclDist float64     `json:"attestation_avg_incl_dist"`

	SyncCommittee          VDBGroupSummaryColumnItem    `json:"sync"`
	SyncCommitteeCount     VDBGroupSummarySyncCount     `json:"sync_count"`
	SyncEfficiency         *float64                     `json:"sync_efficiency"`
	Slashings              VDBGroupSummaryColumnItem    `json:"slashings"`                // Failed slashings are count of validators in the group that were slashed
	ProposalValidators     []uint64                     `json:"proposal_validators"`      // fill with up to 3 validator indexes
	ProposalValidatorCount uint64                       `json:"proposal_validator_count"` // number of distinct validators
	ProposalEfficiency     *float64                     `json:"proposal_efficiency"`
	MissedRewards          VDBGroupSummaryMissedRewards `json:"missed_rewards"`

	Apr ClElValue[*float64] `json:"apr"`

	Luck Luck `json:"luck"`

	RocketPool struct {
		Minipools  uint64  `json:"minipools"`
		Collateral float64 `json:"collateral"`
	} `json:"rocket_pool,omitempty"`
}
type GetValidatorDashboardGroupSummaryResponse ApiDataResponse[VDBGroupSummaryData]

type GetValidatorDashboardSummaryChartResponse ApiDataResponse[ChartData[int, *float64]] // line chart, series id is group id

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

	ProposalStatusCount         StatusCount     `json:"proposal_status_count"`
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
	PublicKey            PubKey          `json:"public_key" faker:"pubkey"`
	Index                *uint64         `json:"index,omitempty"`
	GroupId              uint64          `json:"group_id"`
	Block                uint64          `json:"block"`
	BlockIndex           uint64          `json:"block_index"` // unique
	Timestamp            int64           `json:"timestamp" faker:"past_timestamp"`
	From                 Address         `json:"-"` // TODO enable again
	Depositor            Address         `json:"depositor"`
	TxHash               Hash            `json:"tx_hash" faker:"tx_hash"`
	WithdrawalCredential Hash            `json:"withdrawal_credential" faker:"withdrawal_credentials"`
	Amount               decimal.Decimal `json:"amount" faker:"eth"`
	Validity             string          `json:"validity" tstype:"'valid' | 'invalid' | 'invalid_skipped'" faker:"oneof: valid, invalid, invalid_skipped"`
}
type GetValidatorDashboardExecutionLayerDepositsResponse ApiPagingResponse[VDBExecutionDepositsTableRow]

type VDBConsensusDepositsTableRow struct {
	PublicKey            PubKey          `json:"public_key" faker:"pubkey"`
	Index                uint64          `json:"index"`
	GroupId              uint64          `json:"group_id"`
	SlotQueued           *uint64         `json:"slot_queued,omitempty"`
	SlotProcessed        *uint64         `json:"slot_processed,omitempty"`
	WithdrawalCredential Hash            `json:"withdrawal_credential" faker:"withdrawal_credentials"`
	Amount               decimal.Decimal `json:"amount" faker:"eth"`
	Signature            Hash            `json:"signature"`
	Type                 string          `json:"type" tstype:"'manual' | 'auto'" faker:"oneof: manual, auto"`
	Status               string          `json:"status" tstype:"'queued' | 'completed' | 'rejected'" faker:"oneof: queued, completed, rejected"`
	RejectReason         *string         `json:"reject_reason,omitempty" tstype:"'invalid_signature'" faker:"oneof: invalid_signature"`
	// unique
	Slot      uint64 `json:"slot"`
	SlotIndex int64  `json:"slot_index"`
}
type GetValidatorDashboardConsensusLayerDepositsResponse ApiPagingResponse[VDBConsensusDepositsTableRow]

type VDBTotalExecutionDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount" faker:"eth"`
}

type GetValidatorDashboardTotalExecutionDepositsResponse ApiDataResponse[VDBTotalExecutionDepositsData]

type VDBTotalConsensusDepositsData struct {
	TotalAmount decimal.Decimal `json:"total_amount" faker:"eth"`
}

type GetValidatorDashboardTotalConsensusDepositsResponse ApiDataResponse[VDBTotalConsensusDepositsData]

// ------------------------------------------------------------
// Withdrawals Tab
type VDBWithdrawalsElTableRow struct {
	BlockQueued        uint64          `json:"block_queued"`
	TimestampQueued    int64           `json:"timestamp_queued" faker:"past_timestamp"`
	TxIndexQueued      uint64          `json:"tx_index_queued"`
	ITxIndexQueued     uint64          `json:"itx_index_queued"`
	BlockProcessed     *uint64         `json:"block_processed,omitempty"`
	TimestampProcessed *int64          `json:"timestamp_processed,omitempty" faker:"past_timestamp"`
	Index              uint64          `json:"index"`
	TxHash             Hash            `json:"tx_hash" faker:"tx_hash"`
	GroupId            uint64          `json:"group_id"`
	From               Address         `json:"-"`
	Withdrawer         Address         `json:"withdrawer"`
	Amount             decimal.Decimal `json:"amount" faker:"eth"`
	Status             string          `json:"status" tstype:"'queued' | 'processed'" faker:"oneof: queued, processed"`
	Fee                decimal.Decimal `json:"fee" faker:"eth"`
}
type GetValidatorDashboardExecutionLayerWithdrawalsResponse ApiPagingResponse[VDBWithdrawalsElTableRow]

type VDBWithdrawalsClTableRow struct {
	SlotQueued            *uint64         `json:"slot_queued,omitempty"` // does not exist pre-pectra
	SlotProcessed         uint64          `json:"slot_processed"`
	Index                 uint64          `json:"index"`
	PublicKey             PubKey          `json:"public_key" faker:"pubkey"`
	WithdrawalCredentials Hash            `json:"withdrawal_credentials"`
	GroupId               uint64          `json:"group_id"`
	Recipient             *Address        `json:"recipient,omitempty"`
	Amount                decimal.Decimal `json:"amount" faker:"eth"`
	Type                  string          `json:"type" tstype:"'auto' | 'manual'" faker:"oneof: auto, manual"`
	Status                string          `json:"status" tstype:"'queued' | 'completed' | 'rejected'" faker:"oneof: queued, completed, rejected"`
	RejectReason          *string         `json:"reject_reason,omitempty" tstype:"'full_queue' | 'unknown_pubkey' | 'no_execution_withdrawal_credentials' | 'address_mismatch' | 'inactive' | 'exiting' | 'too_young' | 'pending_withdrawals' | 'not_compounding' | 'insufficient_effective_balance' | 'excess_balance'" faker:"oneof: full_queue, unknown_pubkey, no_execution_withdrawal_credentials, address_mismatch, inactive, exiting, too_young, pending_withdrawals, not_compounding, insufficient_effective_balance, excess_balance"`
	// unique
	Slot      uint64 `json:"slot"`
	SlotIndex int64  `json:"slot_index"`
}
type GetValidatorDashboardConsensusLayerWithdrawalsResponse ApiPagingResponse[VDBWithdrawalsClTableRow]

type VDBTotalExecutionWithdrawalsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type GetValidatorDashboardTotalExecutionWithdrawalsResponse ApiDataResponse[VDBTotalExecutionWithdrawalsData]

type VDBTotalConsensusWithdrawalsData struct {
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type GetValidatorDashboardTotalConsensusWithdrawalsResponse ApiDataResponse[VDBTotalConsensusWithdrawalsData]

// ------------------------------------------------------------
// Consolidations Tab
type VDBConsolidationsElTableRow struct {
	From               Address         `json:"-"`
	Consolidator       Address         `json:"consolidator"`
	Source             uint64          `json:"source"`
	Target             uint64          `json:"target"`
	BlockQueued        uint64          `json:"block_queued"`
	TimestampQueued    int64           `json:"timestamp_queued" faker:"past_timestamp"`
	TxIndexQueued      uint64          `json:"tx_index_queued"`
	ITxIndexQueued     uint64          `json:"itx_index_queued"`
	BlockProcessed     *uint64         `json:"block_processed,omitempty"`
	TimestampProcessed *int64          `json:"timestamp_processed,omitempty" faker:"past_timestamp"`
	Status             string          `json:"status" tstype:"'queued' | 'processed'" faker:"oneof: queued, processed"`
	TxHash             Hash            `json:"tx_hash" faker:"tx_hash"`
	Fee                decimal.Decimal `json:"fee" faker:"eth"`
}
type GetValidatorDashboardExecutionLayerConsolidationsResponse ApiPagingResponse[VDBConsolidationsElTableRow]
type VDBConsolidationsClTableRow struct {
	Source        uint64           `json:"source"`
	Target        uint64           `json:"target"`
	SlotQueued    *uint64          `json:"slot_queued,omitempty"`
	SlotProcessed *uint64          `json:"slot_processed,omitempty"`
	Status        string           `json:"status" tstype:"'queued' | 'completed' | 'rejected'" faker:"oneof: queued, completed, rejected"`
	RejectReason  *string          `json:"reject_reason,omitempty" tstype:"'source_equals_target' | 'full_queue' | 'insufficient_consolidation_churn' | 'source_unknown_pubkey' | 'target_unknown_pubkey' | 'source_no_execution_withdrawal_credentials' | 'source_address_mismatch' | 'target_not_compounding' | 'source_inactive' | 'target_inactive' | 'source_exiting' | 'target_exiting' | 'source_too_young' | 'source_pending_withdrawals' | 'source_slashed'" faker:"oneof: source_equals_target, full_queue, insufficient_consolidation_churn, source_unknown_pubkey, target_unknown_pubkey, source_no_execution_withdrawal_credentials, source_address_mismatch, target_not_compounding, source_inactive, target_inactive, source_exiting, target_exiting, source_too_young, source_pending_withdrawals, source_slashed"`
	Amount        *decimal.Decimal `json:"amount,omitempty" faker:"eth"`
	Id            uint64           `json:"id"`
	Slot          uint64           `json:"slot"`
	SlotIndex     uint64           `json:"slot_index"`
}
type GetValidatorDashboardConsensusLayerConsolidationsResponse ApiPagingResponse[VDBConsolidationsClTableRow]

// ------------------------------------------------------------
// Rocket Pool Tab
type VDBRocketPoolTableRow struct {
	Address                []byte                             `json:"-"`
	Node                   Address                            `json:"node" extensions:"x-order=1"`
	StakedEth              decimal.Decimal                    `json:"staked_eth"`
	StakedRpl              decimal.Decimal                    `json:"staked_rpl"`
	MinipoolsTotal         uint64                             `json:"minipools_count_total"`
	MinipoolsLeb16         uint64                             `json:"minipools_count_leb_16"`
	MinipoolsLeb8          uint64                             `json:"minipools_count_leb_8"`
	Collateral             PercentageDetails[decimal.Decimal] `json:"collateral"`
	AvgCommission          float64                            `json:"avg_commission"`
	RplClaimed             decimal.Decimal                    `json:"rpl_claimed"`
	RplUnclaimed           decimal.Decimal                    `json:"rpl_unclaimed"`
	EffectiveRpl           decimal.Decimal                    `json:"effective_rpl"`
	RplApr                 float64                            `json:"rpl_apr"`
	RplAprUpdateTs         int64                              `json:"rpl_apr_update_ts"`
	RplEstimate            decimal.Decimal                    `json:"rpl_estimate"`
	SmoothingpoolOptIn     bool                               `json:"smoothingpool_opt_in"`
	SmoothingpoolClaimed   decimal.Decimal                    `json:"smoothingpool_claimed"`
	SmoothingpoolUnclaimed decimal.Decimal                    `json:"smoothingpool_unclaimed"`
	NodeDepositBalance     decimal.Decimal                    `json:"node_deposit_balance"`
	UserDepositBalance     decimal.Decimal                    `json:"user_deposit_balance"`

	Timezone      string          `json:"timezone"`
	RefundBalance decimal.Decimal `json:"refund_balance"`
	DepositCredit decimal.Decimal `json:"deposit_credit"`
}

type GetValidatorDashboardRocketPoolResponse ApiPagingResponse[VDBRocketPoolTableRow]

type GetValidatorDashboardTotalRocketPoolResponse ApiDataResponse[VDBRocketPoolTableRow]

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
	Index   uint64 `json:"index"`
	GroupId uint64 `json:"group_id"`
}

// helper for frontend
type PostValidatorDashboardValidatorsRequest struct {
	GroupId              uint64        `json:"group_id,omitempty" x-nullable:"true"`
	Validators           []interface{} `json:"validators,omitempty" tstype:"(number | string)[]"`
	DepositAddress       string        `json:"deposit_address,omitempty"`
	WithdrawalCredential string        `json:"withdrawal_credential,omitempty"`
	Graffiti             string        `json:"graffiti,omitempty"`
}

type PostValidatorDashboardGroupsRequest struct {
	Name string `json:"name"`
}
