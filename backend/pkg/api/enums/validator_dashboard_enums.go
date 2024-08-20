package enums

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
	Status         VDBBlocksColumn
	ProposerReward VDBBlocksColumn
}{
	VDBBlockProposer,
	VDBBlockSlot,
	VDBBlockBlock,
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
	Index     VDBWithdrawalsColumn
	Recipient VDBWithdrawalsColumn
	Amount    VDBWithdrawalsColumn
}{
	VDBWithdrawalEpoch,
	VDBWithdrawalSlot,
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
// Validator Dashboard Archived Reasons

type VDBArchivedReason int

var _ Enum = VDBArchivedReason(0)

const (
	VDBArchivedUser VDBArchivedReason = iota
	VDBArchivedDashboards
	VDBArchivedGroups
	VDBArchivedValidators
)

func (r VDBArchivedReason) Int() int {
	return int(r)
}

func (r VDBArchivedReason) ToString() string {
	switch r {
	case VDBArchivedUser:
		return "user"
	case VDBArchivedDashboards:
		return "dashboard_limit"
	case VDBArchivedGroups:
		return "group_limit"
	case VDBArchivedValidators:
		return "validator_limit"
	default:
		return ""
	}
}

var VDBArchivedReasons = struct {
	User       VDBArchivedReason
	Dashboards VDBArchivedReason
	Groups     VDBArchivedReason
	Validators VDBArchivedReason
}{
	VDBArchivedUser,
	VDBArchivedDashboards,
	VDBArchivedGroups,
	VDBArchivedValidators,
}

// ----------------
// Validator Reward Chart Efficiency Filter

type VDBSummaryChartEfficiencyType int

var _ EnumFactory[VDBSummaryChartEfficiencyType] = VDBSummaryChartEfficiencyType(0)

const (
	VDBSummaryChartAll VDBSummaryChartEfficiencyType = iota
	VDBSummaryChartAttestation
	VDBSummaryChartSync
	VDBSummaryChartProposal
)

func (c VDBSummaryChartEfficiencyType) Int() int {
	return int(c)
}

func (VDBSummaryChartEfficiencyType) NewFromString(s string) VDBSummaryChartEfficiencyType {
	switch s {
	case "", "all":
		return VDBSummaryChartAll
	case "attestation":
		return VDBSummaryChartAttestation
	case "sync":
		return VDBSummaryChartSync
	case "proposal":
		return VDBSummaryChartProposal
	default:
		return VDBSummaryChartEfficiencyType(-1)
	}
}

var VDBSummaryChartEfficiencyFilters = struct {
	All         VDBSummaryChartEfficiencyType
	Attestation VDBSummaryChartEfficiencyType
	Sync        VDBSummaryChartEfficiencyType
	Proposal    VDBSummaryChartEfficiencyType
}{
	VDBSummaryChartAll,
	VDBSummaryChartAttestation,
	VDBSummaryChartSync,
	VDBSummaryChartProposal,
}

// ----------------
// Validator Dashboard Rocket Pool Table

type VDBRocketPoolColumn int

var _ EnumFactory[VDBRocketPoolColumn] = VDBRocketPoolColumn(0)

const (
	VDBRocketPoolNode VDBRocketPoolColumn = iota
	VDBRocketPoolMinipools
	VDBRocketPoolCollateral
	VDBRocketPoolRpl
	VDBRocketPoolEffectiveRpl
	VDBRocketPoolRplApr
	VDBRocketPoolSmoothingPool
)

func (c VDBRocketPoolColumn) Int() int {
	return int(c)
}

func (VDBRocketPoolColumn) NewFromString(s string) VDBRocketPoolColumn {
	switch s {
	case "node":
		return VDBRocketPoolNode
	case "minipools":
		return VDBRocketPoolMinipools
	case "collateral":
		return VDBRocketPoolCollateral
	case "rpl":
		return VDBRocketPoolRpl
	case "effective_rpl":
		return VDBRocketPoolEffectiveRpl
	case "rpl_apr":
		return VDBRocketPoolRplApr
	case "smoothing_pool":
		return VDBRocketPoolSmoothingPool
	default:
		return VDBRocketPoolColumn(-1)
	}
}

var VDBRocketPoolColumns = struct {
	Node          VDBRocketPoolColumn
	Minipools     VDBRocketPoolColumn
	Collateral    VDBRocketPoolColumn
	Rpl           VDBRocketPoolColumn
	EffectiveRpl  VDBRocketPoolColumn
	RplApr        VDBRocketPoolColumn
	SmoothingPool VDBRocketPoolColumn
}{
	VDBRocketPoolNode,
	VDBRocketPoolMinipools,
	VDBRocketPoolCollateral,
	VDBRocketPoolRpl,
	VDBRocketPoolEffectiveRpl,
	VDBRocketPoolRplApr,
	VDBRocketPoolSmoothingPool,
}

// ----------------
// Validator Dashboard Rocket Pool Minipools modal

type VDBRocketPoolMinipoolsColumn int

var _ EnumFactory[VDBRocketPoolMinipoolsColumn] = VDBRocketPoolMinipoolsColumn(0)

const (
	VDBRocketPoolMinipoolsGroup VDBRocketPoolMinipoolsColumn = iota
)

func (c VDBRocketPoolMinipoolsColumn) Int() int {
	return int(c)
}

func (VDBRocketPoolMinipoolsColumn) NewFromString(s string) VDBRocketPoolMinipoolsColumn {
	switch s {
	case "group":
		return VDBRocketPoolMinipoolsGroup
	default:
		return VDBRocketPoolMinipoolsColumn(-1)
	}
}

var VDBWRocketPoolColumns = struct {
	Group VDBRocketPoolMinipoolsColumn
}{
	VDBRocketPoolMinipoolsGroup,
}
