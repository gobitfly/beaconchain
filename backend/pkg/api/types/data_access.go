package types

type Sort[T ~int] struct {
	Column T
	Desc   bool
}

// ----------------
// Table Column Enums
// shouldn't be converted to typescript, so they are defined here

type VDBSummaryTableColumn int
// prob rather use maps
var VDBSummaryTableColumnSortNames = []string{"group", "efficiencyDay", "efficiencyWeek", "efficiencyMonth", "efficiencyTotal", "validators"}
const (
	VDBSummaryGroup VDBSummaryTableColumn = iota
	VDBSummaryEfficiencyDay
	VDBSummaryEfficiencyWeek
	VDBSummaryEfficiencyMonth
	VDBSummaryEfficiencyTotal
	VDBSummaryValidators // Sort by count, not by index
)

type VDBRewardsTableColumn int
var VDBRewardsTableColumnSortNames = []string{"epoch", "duty"}
const (
	VDBRewardEpoch VDBRewardsTableColumn = iota
	VDBRewardDuty                        // Sort by sum of percentages
)

type VDBBlocksTableColumn int
var VDBBlocksTableColumnSortNames = []string{"proposer", "group", "epoch", "slot", "block", "age", "status", "proposerReward"}
const (
	VDBBlockProposer VDBBlocksTableColumn = iota
	VDBBlockGroup
	VDBBlockEpoch
	VDBBlockSlot
	VDBBlockBlock
	VDBBlockAge
	VDBBlockStatus
	VDBBlockProposerReward
)

type VDBWithdrawalsTableColumn int
var VDBWithdrawalsTableColumnSortNames = []string{"epoch", "age", "index", "group", "recipient", "amount"}
const (
	VDBWithdrawalEpoch VDBWithdrawalsTableColumn = iota
	VDBWithdrawalAge
	VDBWithdrawalIndex
	VDBWithdrawalGroup
	VDBWithdrawalRecipient
	VDBWithdrawalAmount
)
