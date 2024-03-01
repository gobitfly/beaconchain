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

// Factory interface for creating enum values from int
type ColEnumFactory[T ColEnum] interface {
	ColEnum
	NewFromInt(i int) T
}

type VDBSummaryTableColumn int

const (
	VDBSummaryGroup VDBSummaryTableColumn = iota
	VDBSummaryEfficiencyDay
	VDBSummaryEfficiencyWeek
	VDBSummaryEfficiencyMonth
	VDBSummaryEfficiencyTotal
	VDBSummaryValidators // Sort by count, not by index
)

func (VDBSummaryTableColumn) GetColNames() []string {
	return []string{"group", "efficiency_day", "efficiency_week", "efficiency_month", "efficiency_total", "validators"}
}
func (VDBSummaryTableColumn) NewFromInt(i int) VDBSummaryTableColumn {
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

func (c VDBRewardsTableColumn) NewFromInt(i int) VDBRewardsTableColumn {
	return VDBRewardsTableColumn(i)
}

type VDBDutiesTableColumn int

const (
	VDBDutyValidator VDBDutiesTableColumn = iota
	VDBDutyReward                         // Sort by sum of percentages
)

func (VDBDutiesTableColumn) GetColNames() []string {
	return []string{"validator", "reward"}
}
func (VDBDutiesTableColumn) NewFromInt(i int) VDBDutiesTableColumn {
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
	return []string{"proposer", "group", "epoch", "slot", "block", "age", "status", "reward"}
}
func (VDBBlocksTableColumn) NewFromInt(i int) VDBBlocksTableColumn {
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
	return []string{"epoch", "age", "index", "group", "recipient", "amount"}
}
func (VDBWithdrawalsTableColumn) NewFromInt(i int) VDBWithdrawalsTableColumn {
	return VDBWithdrawalsTableColumn(i)
}

type VDBManageValidatorsTableColumn int

const (
	VDBManageValidatorsIndex VDBManageValidatorsTableColumn = iota
	VDBManageValidatorsPublicKey
	VDBManageValidatorsBalance
	VDBManageValidatorsStatus
	VDBManageValidatorsWithdrawalCredential
)

var VDBValidatorsColumnSortNames = []string{"index", "public_key", "balance", "status", "withdrawal_credential"}

func (c VDBManageValidatorsTableColumn) GetColNames() []string {
	return VDBValidatorsColumnSortNames
}
func (VDBManageValidatorsTableColumn) NewFromInt(i int) VDBManageValidatorsTableColumn {
	return VDBManageValidatorsTableColumn(i)
}
