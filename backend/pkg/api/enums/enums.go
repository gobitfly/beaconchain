package enums

type Enum interface {
	Int() int
}

// Factory interface for creating enum values from int
type EnumFactory[T Enum] interface {
	Enum
	NewFromString(string) T
}

// ----------------
// Validator Dashboard Summary Table

type VDBSummaryTableColumn int

var _ EnumFactory[VDBSummaryTableColumn] = VDBSummaryTableColumn(0)

const (
	VDBSummaryGroup VDBSummaryTableColumn = iota
	VDBSummaryEfficiencyDay
	VDBSummaryEfficiencyWeek
	VDBSummaryEfficiencyMonth
	VDBSummaryEfficiencyTotal
	VDBSummaryValidators // Sort by count, not by index
)

func (c VDBSummaryTableColumn) Int() int {
	return int(c)
}

func (VDBSummaryTableColumn) NewFromString(s string) VDBSummaryTableColumn {
	switch s {
	case "group":
		return VDBSummaryGroup
	case "efficiency_day":
		return VDBSummaryEfficiencyDay
	case "efficiency_week":
		return VDBSummaryEfficiencyWeek
	case "efficiency_month":
		return VDBSummaryEfficiencyMonth
	case "efficiency_total":
		return VDBSummaryEfficiencyTotal
	case "validators":
		return VDBSummaryValidators
	default:
		return VDBSummaryTableColumn(-1)
	}
}

var VDBSummaryTableColumns = struct {
	Group           VDBSummaryTableColumn
	EfficiencyDay   VDBSummaryTableColumn
	EfficiencyWeek  VDBSummaryTableColumn
	EfficiencyMonth VDBSummaryTableColumn
	EfficiencyTotal VDBSummaryTableColumn
	Validators      VDBSummaryTableColumn
}{
	VDBSummaryGroup,
	VDBSummaryEfficiencyDay,
	VDBSummaryEfficiencyWeek,
	VDBSummaryEfficiencyMonth,
	VDBSummaryEfficiencyTotal,
	VDBSummaryValidators,
}

// ----------------
// Validator Dashboard Rewards Table

type VDBRewardsTableColumn int

var _ EnumFactory[VDBRewardsTableColumn] = VDBRewardsTableColumn(0)

const (
	VDBRewardEpoch VDBRewardsTableColumn = iota
	VDBRewardDuty                        // Sort by sum of percentages
)

func (c VDBRewardsTableColumn) Int() int {
	return int(c)
}

func (VDBRewardsTableColumn) NewFromString(s string) VDBRewardsTableColumn {
	switch s {
	case "epoch":
		return VDBRewardEpoch
	case "duty":
		return VDBRewardDuty
	default:
		return VDBRewardsTableColumn(-1)
	}
}

var VDBRewardsTableColumns = struct {
	Epoch VDBRewardsTableColumn
	Duty  VDBRewardsTableColumn
}{
	VDBRewardEpoch,
	VDBRewardDuty,
}

// ----------------
// Validator Dashboard Duties Table

type VDBDutiesTableColumn int

var _ EnumFactory[VDBDutiesTableColumn] = VDBDutiesTableColumn(0)

const (
	VDBDutyValidator VDBDutiesTableColumn = iota
	VDBDutyReward                         // Sort by sum of percentages
)

func (c VDBDutiesTableColumn) Int() int {
	return int(c)
}

func (VDBDutiesTableColumn) NewFromString(s string) VDBDutiesTableColumn {
	switch s {
	case "validator":
		return VDBDutyValidator
	case "reward":
		return VDBDutyReward
	default:
		return VDBDutiesTableColumn(-1)
	}
}

var VDBDutiesTableColumns = struct {
	Validator VDBDutiesTableColumn
	Reward    VDBDutiesTableColumn
}{
	VDBDutyValidator,
	VDBDutyReward,
}

// ----------------
// Validator Dashboard Blocks Table

type VDBBlocksTableColumn int

var _ EnumFactory[VDBBlocksTableColumn] = VDBBlocksTableColumn(0)

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

func (c VDBBlocksTableColumn) Int() int {
	return int(c)
}

func (VDBBlocksTableColumn) NewFromString(s string) VDBBlocksTableColumn {
	switch s {
	case "proposer":
		return VDBBlockProposer
	case "group":
		return VDBBlockGroup
	case "epoch":
		return VDBBlockEpoch
	case "slot":
		return VDBBlockSlot
	case "block":
		return VDBBlockBlock
	case "age":
		return VDBBlockAge
	case "status":
		return VDBBlockStatus
	case "proposer_reward":
		return VDBBlockProposerReward
	default:
		return VDBBlocksTableColumn(-1)
	}
}

var VDBBlocksTableColumns = struct {
	Proposer       VDBBlocksTableColumn
	Group          VDBBlocksTableColumn
	Epoch          VDBBlocksTableColumn
	Slot           VDBBlocksTableColumn
	Block          VDBBlocksTableColumn
	Age            VDBBlocksTableColumn
	Status         VDBBlocksTableColumn
	ProposerReward VDBBlocksTableColumn
}{
	VDBBlockProposer,
	VDBBlockGroup,
	VDBBlockEpoch,
	VDBBlockSlot,
	VDBBlockBlock,
	VDBBlockAge,
	VDBBlockStatus,
	VDBBlockProposerReward,
}

// ----------------
// Validator Dashboard Withdrawals Table

type VDBWithdrawalsTableColumn int

var _ EnumFactory[VDBWithdrawalsTableColumn] = VDBWithdrawalsTableColumn(0)

const (
	VDBWithdrawalEpoch VDBWithdrawalsTableColumn = iota
	VDBWithdrawalAge
	VDBWithdrawalIndex
	VDBWithdrawalGroup
	VDBWithdrawalRecipient
	VDBWithdrawalAmount
)

func (c VDBWithdrawalsTableColumn) Int() int {
	return int(c)
}

func (VDBWithdrawalsTableColumn) NewFromString(s string) VDBWithdrawalsTableColumn {
	switch s {
	case "epoch":
		return VDBWithdrawalEpoch
	case "age":
		return VDBWithdrawalAge
	case "index":
		return VDBWithdrawalIndex
	case "group":
		return VDBWithdrawalGroup
	case "recipient":
		return VDBWithdrawalRecipient
	case "amount":
		return VDBWithdrawalAmount
	default:
		return VDBWithdrawalsTableColumn(-1)
	}
}

var VDBWithdrawalsTableColumns = struct {
	Epoch     VDBWithdrawalsTableColumn
	Age       VDBWithdrawalsTableColumn
	Index     VDBWithdrawalsTableColumn
	Group     VDBWithdrawalsTableColumn
	Recipient VDBWithdrawalsTableColumn
	Amount    VDBWithdrawalsTableColumn
}{
	VDBWithdrawalEpoch,
	VDBWithdrawalAge,
	VDBWithdrawalIndex,
	VDBWithdrawalGroup,
	VDBWithdrawalRecipient,
	VDBWithdrawalAmount,
}

// ----------------
// Validator Dashboard Manage Validators Table

type VDBManageValidatorsTableColumn int

var _ EnumFactory[VDBManageValidatorsTableColumn] = VDBManageValidatorsTableColumn(0)

const (
	VDBManageValidatorsIndex VDBManageValidatorsTableColumn = iota
	VDBManageValidatorsPublicKey
	VDBManageValidatorsBalance
	VDBManageValidatorsStatus
	VDBManageValidatorsWithdrawalCredential
)

func (c VDBManageValidatorsTableColumn) Int() int {
	return int(c)
}

func (VDBManageValidatorsTableColumn) NewFromString(s string) VDBManageValidatorsTableColumn {
	switch s {
	case "index":
		return VDBManageValidatorsIndex
	case "public_key":
		return VDBManageValidatorsPublicKey
	case "balance":
		return VDBManageValidatorsBalance
	case "status":
		return VDBManageValidatorsStatus
	case "withdrawal_credential":
		return VDBManageValidatorsWithdrawalCredential
	default:
		return VDBManageValidatorsTableColumn(-1)
	}
}

var VDBManageValidatorsTableColumns = struct {
	Index                VDBManageValidatorsTableColumn
	PublicKey            VDBManageValidatorsTableColumn
	Balance              VDBManageValidatorsTableColumn
	Status               VDBManageValidatorsTableColumn
	WithdrawalCredential VDBManageValidatorsTableColumn
}{
	VDBManageValidatorsIndex,
	VDBManageValidatorsPublicKey,
	VDBManageValidatorsBalance,
	VDBManageValidatorsStatus,
	VDBManageValidatorsWithdrawalCredential,
}
