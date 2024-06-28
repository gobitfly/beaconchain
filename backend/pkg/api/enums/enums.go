package enums

import "time"

type Enum interface {
	Int() int
}

// Factory interface for creating enum values from strings
type EnumFactory[T Enum] interface {
	Enum
	NewFromString(string) T
}

func IsInvalidEnum(e Enum) bool {
	return e.Int() == -1
}

// ----------------
// Validator Dashboard Summary Table

type VDBSummaryColumn int

var _ EnumFactory[VDBSummaryColumn] = VDBSummaryColumn(0)

const (
	VDBSummaryGroup VDBSummaryColumn = iota
	VDBSummaryValidators
	VDBSummaryEfficiency
	VDBSummaryAttestations
	VDBSummaryProposals
	VDBSummaryReward
)

func (c VDBSummaryColumn) Int() int {
	return int(c)
}

func (VDBSummaryColumn) NewFromString(s string) VDBSummaryColumn {
	switch s {
	case "group_id":
		return VDBSummaryGroup
	case "validators":
		return VDBSummaryValidators
	case "efficiency":
		return VDBSummaryEfficiency
	case "attestations":
		return VDBSummaryAttestations
	case "proposals":
		return VDBSummaryProposals
	case "reward":
		return VDBSummaryReward
	default:
		return VDBSummaryColumn(-1)
	}
}

var VDBSummaryColumns = struct {
	Group        VDBSummaryColumn
	Validators   VDBSummaryColumn
	Efficiency   VDBSummaryColumn
	Attestations VDBSummaryColumn
	Proposals    VDBSummaryColumn
	Reward       VDBSummaryColumn
}{
	VDBSummaryGroup,
	VDBSummaryValidators,
	VDBSummaryEfficiency,
	VDBSummaryAttestations,
	VDBSummaryProposals,
	VDBSummaryReward,
}

// ----------------
// Validator Dashboard Rewards Table

type VDBRewardsColumn int

var _ EnumFactory[VDBRewardsColumn] = VDBRewardsColumn(0)

const (
	VDBRewardEpoch VDBRewardsColumn = iota
)

func (c VDBRewardsColumn) Int() int {
	return int(c)
}

func (VDBRewardsColumn) NewFromString(s string) VDBRewardsColumn {
	switch s {
	case "epoch":
		return VDBRewardEpoch
	default:
		return VDBRewardsColumn(-1)
	}
}

var VDBRewardsColumns = struct {
	Epoch VDBRewardsColumn
}{
	VDBRewardEpoch,
}

// ----------------
// Validator Dashboard Duties Table

type VDBDutiesColumn int

var _ EnumFactory[VDBDutiesColumn] = VDBDutiesColumn(0)

const (
	VDBDutyValidator VDBDutiesColumn = iota
	VDBDutyReward                    // Sort by sum of percentages
)

func (c VDBDutiesColumn) Int() int {
	return int(c)
}

func (VDBDutiesColumn) NewFromString(s string) VDBDutiesColumn {
	switch s {
	case "validator":
		return VDBDutyValidator
	case "reward":
		return VDBDutyReward
	default:
		return VDBDutiesColumn(-1)
	}
}

var VDBDutiesColumns = struct {
	Validator VDBDutiesColumn
	Reward    VDBDutiesColumn
}{
	VDBDutyValidator,
	VDBDutyReward,
}

// ----------------
// Validator Dashboard Blocks Table

type VDBBlocksColumn int

var _ EnumFactory[VDBBlocksColumn] = VDBBlocksColumn(0)

const (
	VDBBlockSlot VDBBlocksColumn = iota // default
	VDBBlockProposer
	VDBBlockBlock
	VDBBlockAge
	VDBBlockStatus
	VDBBlockProposerReward
)

func (c VDBBlocksColumn) Int() int {
	return int(c)
}

func (VDBBlocksColumn) NewFromString(s string) VDBBlocksColumn {
	switch s {
	case "proposer":
		return VDBBlockProposer
	case "slot":
		return VDBBlockSlot
	case "block":
		return VDBBlockBlock
	case "age":
		return VDBBlockAge
	case "status":
		return VDBBlockStatus
	case "reward":
		return VDBBlockProposerReward
	default:
		return VDBBlocksColumn(-1)
	}
}

var VDBBlocksColumns = struct {
	Proposer       VDBBlocksColumn
	Slot           VDBBlocksColumn
	Block          VDBBlocksColumn
	Age            VDBBlocksColumn
	Status         VDBBlocksColumn
	ProposerReward VDBBlocksColumn
}{
	VDBBlockProposer,
	VDBBlockSlot,
	VDBBlockBlock,
	VDBBlockAge,
	VDBBlockStatus,
	VDBBlockProposerReward,
}

// ----------------
// Validator Dashboard Withdrawals Table

type VDBWithdrawalsColumn int

var _ EnumFactory[VDBWithdrawalsColumn] = VDBWithdrawalsColumn(0)

const (
	VDBWithdrawalEpoch VDBWithdrawalsColumn = iota
	VDBWithdrawalSlot
	VDBWithdrawalAge
	VDBWithdrawalIndex
	VDBWithdrawalRecipient
	VDBWithdrawalAmount
)

func (c VDBWithdrawalsColumn) Int() int {
	return int(c)
}

func (VDBWithdrawalsColumn) NewFromString(s string) VDBWithdrawalsColumn {
	switch s {
	case "epoch":
		return VDBWithdrawalEpoch
	case "slot":
		return VDBWithdrawalSlot
	case "age":
		return VDBWithdrawalAge
	case "index":
		return VDBWithdrawalIndex
	case "recipient":
		return VDBWithdrawalRecipient
	case "amount":
		return VDBWithdrawalAmount
	default:
		return VDBWithdrawalsColumn(-1)
	}
}

var VDBWithdrawalsColumns = struct {
	Epoch     VDBWithdrawalsColumn
	Slot      VDBWithdrawalsColumn
	Age       VDBWithdrawalsColumn
	Index     VDBWithdrawalsColumn
	Recipient VDBWithdrawalsColumn
	Amount    VDBWithdrawalsColumn
}{
	VDBWithdrawalEpoch,
	VDBWithdrawalSlot,
	VDBWithdrawalAge,
	VDBWithdrawalIndex,
	VDBWithdrawalRecipient,
	VDBWithdrawalAmount,
}

// ----------------
// Validator Dashboard Manage Validators Table

type VDBManageValidatorsColumn int

var _ EnumFactory[VDBManageValidatorsColumn] = VDBManageValidatorsColumn(0)

const (
	VDBManageValidatorsIndex VDBManageValidatorsColumn = iota
	VDBManageValidatorsPublicKey
	VDBManageValidatorsBalance
	VDBManageValidatorsStatus
	VDBManageValidatorsWithdrawalCredential
)

func (c VDBManageValidatorsColumn) Int() int {
	return int(c)
}

func (VDBManageValidatorsColumn) NewFromString(s string) VDBManageValidatorsColumn {
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
		return VDBManageValidatorsColumn(-1)
	}
}

var VDBManageValidatorsColumns = struct {
	Index                VDBManageValidatorsColumn
	PublicKey            VDBManageValidatorsColumn
	Balance              VDBManageValidatorsColumn
	Status               VDBManageValidatorsColumn
	WithdrawalCredential VDBManageValidatorsColumn
}{
	VDBManageValidatorsIndex,
	VDBManageValidatorsPublicKey,
	VDBManageValidatorsBalance,
	VDBManageValidatorsStatus,
	VDBManageValidatorsWithdrawalCredential,
}

// ----------------
// Postgres sort direction enum
// SortOrder represents the sorting order, either ascending or descending.
type SortOrder int

// Constants for the sorting order.
const (
	ASC SortOrder = iota
	DESC
)

// String method converts SortOrder to string representation.
func (s SortOrder) String() string {
	if s == ASC {
		return "ASC"
	}
	return "DESC"
}

// Invert method inverts the sorting order.
func (s SortOrder) Invert() SortOrder {
	if s == ASC {
		return DESC
	}
	return ASC
}

var SortOrderColumns = struct {
	Asc  SortOrder
	Desc SortOrder
}{
	ASC,
	DESC,
}

// ----------------
// Time Periods

type TimePeriod int

const (
	AllTime TimePeriod = iota
	Last1h
	Last24h
	Last7d
	Last30d
	Last365d
)

func (t TimePeriod) Int() int {
	return int(t)
}

func (TimePeriod) NewFromString(s string) TimePeriod {
	switch s {
	case "", "all_time":
		return AllTime
	case "1h":
		return Last1h
	case "24h":
		return Last24h
	case "7d":
		return Last7d
	case "30d":
		return Last30d
	case "365d":
		return Last365d
	default:
		return TimePeriod(-1)
	}
}

var TimePeriods = struct {
	AllTime  TimePeriod
	Last1h   TimePeriod
	Last24h  TimePeriod
	Last7d   TimePeriod
	Last30d  TimePeriod
	Last365d TimePeriod
}{
	AllTime,
	Last1h,
	Last24h,
	Last7d,
	Last30d,
	Last365d,
}

func (t TimePeriod) Duration() time.Duration {
	day := 24 * time.Hour
	switch t {
	case Last1h:
		return time.Hour
	case Last24h:
		return day
	case Last7d:
		return 7 * day
	case Last30d:
		return 30 * day
	case Last365d:
		return 365 * day
	default:
		return 0
	}
}

// ----------------
// Validator Duties

type ValidatorDuty int

const (
	DutyNone ValidatorDuty = iota
	DutySync
	DutyProposal
	DutySlashed
)

func (d ValidatorDuty) Int() int {
	return int(d)
}

func (ValidatorDuty) NewFromString(s string) ValidatorDuty {
	switch s {
	case "":
		return DutyNone
	case "sync":
		return DutySync
	case "proposal":
		return DutyProposal
	case "slashed":
		return DutySlashed
	default:
		return ValidatorDuty(-1)
	}
}

var ValidatorDuties = struct {
	None     ValidatorDuty
	Sync     ValidatorDuty
	Proposal ValidatorDuty
	Slashed  ValidatorDuty
}{
	DutyNone,
	DutySync,
	DutyProposal,
	DutySlashed,
}

// ----------------
// Validator Dashboard Summary Table

type ValidatorStatus int

var _ EnumFactory[ValidatorStatus] = ValidatorStatus(0)

const (
	ValidatorStatusDeposited ValidatorStatus = iota
	ValidatorStatusPending
	ValidatorStatusOffline
	ValidatorStatusOnline
	ValidatorStatusSlashed
	ValidatorStatusExited
)

func (vs ValidatorStatus) Int() int {
	return int(vs)
}

func (ValidatorStatus) NewFromString(s string) ValidatorStatus {
	switch s {
	case "deposited":
		return ValidatorStatusDeposited
	case "pending":
		return ValidatorStatusPending
	case "offline":
		return ValidatorStatusOffline
	case "online":
		return ValidatorStatusOnline
	case "slashed":
		return ValidatorStatusSlashed
	case "exited":
		return ValidatorStatusExited
	default:
		return ValidatorStatus(-1)
	}
}

func (vs ValidatorStatus) ToString() string {
	switch vs {
	case ValidatorStatusDeposited:
		return "deposited"
	case ValidatorStatusPending:
		return "pending"
	case ValidatorStatusOffline:
		return "offline"
	case ValidatorStatusOnline:
		return "online"
	case ValidatorStatusSlashed:
		return "slashed"
	case ValidatorStatusExited:
		return "exited"
	default:
		return ""
	}
}

var ValidatorStatuses = struct {
	Deposited ValidatorStatus
	Pending   ValidatorStatus
	Offline   ValidatorStatus
	Online    ValidatorStatus
	Slashed   ValidatorStatus
	Exited    ValidatorStatus
}{
	ValidatorStatusDeposited,
	ValidatorStatusPending,
	ValidatorStatusOffline,
	ValidatorStatusOnline,
	ValidatorStatusSlashed,
	ValidatorStatusExited,
}
