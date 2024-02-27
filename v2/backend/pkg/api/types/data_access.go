package types

type Sort[T ColEnum] struct {
	Column T
	Desc   bool
}

// ----------------
// Table Column Enums
// shouldn't be converted to typescript, so they are defined here
type ColEnum interface {
	GetColNames() []string
}

// Define a generic type constraint that extends ColEnum and includes a method to instantiate itself.
type ColEnumFactory[T ColEnum] interface {
    ColEnum
    NewFromIndex(i int) T
}


type VDBSummaryTableColumn int
// prob rather use maps
const (
	VDBSummaryGroup VDBSummaryTableColumn = iota
	VDBSummaryEfficiencyDay
	VDBSummaryEfficiencyWeek
	VDBSummaryEfficiencyMonth
	VDBSummaryEfficiencyTotal
	VDBSummaryValidators // Sort by count, not by index
)
func (VDBSummaryTableColumn) GetColNames() []string {
	return []string{"group_id", "efficiency_day", "efficiency_week", "efficiency_month", "efficiency_total", "validators"}
}
func (VDBSummaryTableColumn) NewFromIndex(i int) VDBSummaryTableColumn {
    return VDBSummaryTableColumn(i)
}

type VDBRewardsTableColumn int
const (
	VDBRewardEpoch VDBRewardsTableColumn = iota
	VDBRewardDuty                        // Sort by sum of percentages
)
func (c VDBRewardsTableColumn) GetColNames() []string {
	return []string{"epoch", "duty"}
}

func (c VDBRewardsTableColumn) NewFromIndex(i int) VDBRewardsTableColumn {
    return VDBRewardsTableColumn(i)
}

type VDBDutiesTableColumn int
const (
	VDBDutyValidator VDBDutiesTableColumn = iota
	VDBDutyReward                        // Sort by sum of percentages
)
func (VDBDutiesTableColumn) GetColNames() []string {
	return []string{"validator", "reward"}
}
func (VDBDutiesTableColumn) NewFromIndex(i int) VDBDutiesTableColumn {
    return VDBDutiesTableColumn(i)
}

type VDBBlocksTableColumn int
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
func (VDBBlocksTableColumn) GetColNames() []string {
	return []string{"proposer", "group_id", "epoch", "slot", "block", "age", "status", "reward"}
}
func (VDBBlocksTableColumn) NewFromIndex(i int) VDBBlocksTableColumn {
    return VDBBlocksTableColumn(i)
}

type VDBWithdrawalsTableColumn int
const (
	VDBWithdrawalEpoch VDBWithdrawalsTableColumn = iota
	VDBWithdrawalAge
	VDBWithdrawalIndex
	VDBWithdrawalGroup
	VDBWithdrawalRecipient
	VDBWithdrawalAmount
)
func (c VDBWithdrawalsTableColumn) GetColNames() []string {
	return []string{"epoch", "age", "index", "group_id", "recipient", "amount"}
}
func (VDBWithdrawalsTableColumn) NewFromIndex(i int) VDBWithdrawalsTableColumn {
    return VDBWithdrawalsTableColumn(i)
}


// TODO
type VDBValidatorsColumn int
var VDBValidatorsColumnSortNames = []string{}
func (c VDBValidatorsColumn) GetColNames() []string {
	return VDBValidatorsColumnSortNames
}
func (VDBValidatorsColumn) NewFromIndex(i int) VDBValidatorsColumn {
    return VDBValidatorsColumn(i)
}
