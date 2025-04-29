package enums

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

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
	case "", "group_id":
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
	case "", "epoch":
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
	case "", "validator":
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
	case "", "slot":
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

type OrderableSortable interface {
	exp.Orderable
	exp.Comparable
	exp.Isable
	exp.Aliaseable
}

func (c VDBBlocksColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBBlockProposer:
		return goqu.C("validator_index")
	case VDBBlockSlot:
		return goqu.C("slot")
	case VDBBlockBlock:
		return goqu.C("exec_block_number")
	case VDBBlockStatus:
		return goqu.C("status")
	case VDBBlockProposerReward:
		return goqu.L("el_reward + cl_reward")
	default:
		return nil
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
// Validator Dashboard Withdrawals EL Table

type VDBWithdrawalsElColumn int

var _ EnumFactory[VDBWithdrawalsElColumn] = VDBWithdrawalsElColumn(0)

const (
	VDBWithdrawalElBlockQueued VDBWithdrawalsElColumn = iota
	VDBWithdrawalElAmount
)

func (c VDBWithdrawalsElColumn) Int() int {
	return int(c)
}

func (VDBWithdrawalsElColumn) NewFromString(s string) VDBWithdrawalsElColumn {
	switch s {
	case "", "block_queued", "timestamp":
		return VDBWithdrawalElBlockQueued
	case "amount":
		return VDBWithdrawalElAmount
	default:
		return VDBWithdrawalsElColumn(-1)
	}
}

func (c VDBWithdrawalsElColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBWithdrawalElBlockQueued:
		return goqu.I("block_number")
	case VDBWithdrawalElAmount:
		return goqu.I("amount")
	default:
		return nil
	}
}

var VDBWithdrawalsElColumns = struct {
	BlockQueued VDBWithdrawalsElColumn
	Amount      VDBWithdrawalsElColumn
}{
	VDBWithdrawalElBlockQueued,
	VDBWithdrawalElAmount,
}

// ----------------
// Validator Dashboard Withdrawals CL Table
type VDBWithdrawalsClColumn int

var _ EnumFactory[VDBWithdrawalsClColumn] = VDBWithdrawalsClColumn(0)

const (
	VDBWithdrawalClSlotProcessed VDBWithdrawalsClColumn = iota
	VDBWithdrawalClAmount
)

func (c VDBWithdrawalsClColumn) Int() int {
	return int(c)
}

func (VDBWithdrawalsClColumn) NewFromString(s string) VDBWithdrawalsClColumn {
	switch s {
	case "", "slot", "timestamp":
		return VDBWithdrawalClSlotProcessed
	case "amount":
		return VDBWithdrawalClAmount
	default:
		return VDBWithdrawalsClColumn(-1)
	}
}

func (c VDBWithdrawalsClColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBWithdrawalClSlotProcessed:
		return goqu.COALESCE(goqu.I("slot_queued"), goqu.I("slot_processed"))
	case VDBWithdrawalClAmount:
		return goqu.I("amount")
	default:
		return nil
	}
}

var VDBWithdrawalsClColumns = struct {
	Slot   VDBWithdrawalsClColumn
	Amount VDBWithdrawalsClColumn
}{
	VDBWithdrawalClSlotProcessed,
	VDBWithdrawalClAmount,
}

// ----------------
// Validator Dashboard EL Consolidations Table

type VDBConsolidationsElColumn int

var _ EnumFactory[VDBConsolidationsElColumn] = VDBConsolidationsElColumn(0)

const (
	VDBConsolidationElBlockProcessed VDBConsolidationsElColumn = iota
)

func (c VDBConsolidationsElColumn) Int() int {
	return int(c)
}

func (VDBConsolidationsElColumn) NewFromString(s string) VDBConsolidationsElColumn {
	switch s {
	case "", "block_processed", "timestamp":
		return VDBConsolidationElBlockProcessed
	default:
		return VDBConsolidationsElColumn(-1)
	}
}

func (c VDBConsolidationsElColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBConsolidationElBlockProcessed:
		return goqu.I("el_cr.block_number")
	default:
		return nil
	}
}

var VDBConsolidationsElColumns = struct {
	BlockProcessed VDBConsolidationsElColumn
}{
	VDBConsolidationElBlockProcessed,
}

// ----------------
// Validator Dashboard CL Consolidations Table

type VDBConsolidationsClColumn int

var _ EnumFactory[VDBConsolidationsClColumn] = VDBConsolidationsClColumn(0)

const (
	VDBConsolidationClSlot VDBConsolidationsClColumn = iota
	VDBConsolidationClAmount
)

func (c VDBConsolidationsClColumn) Int() int {
	return int(c)
}

func (VDBConsolidationsClColumn) NewFromString(s string) VDBConsolidationsClColumn {
	switch s {
	case "", "slot", "timestamp":
		return VDBConsolidationClSlot
	case "amount":
		return VDBConsolidationClAmount
	default:
		return VDBConsolidationsClColumn(-1)
	}
}

func (c VDBConsolidationsClColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBConsolidationClSlot:
		return goqu.COALESCE(goqu.I("slot_queued"), goqu.I("slot_processed"))
	case VDBConsolidationClAmount:
		return goqu.C("amount_consolidated")
	default:
		return nil
	}
}

var VDBConsolidationsClColumns = struct {
	Slot   VDBConsolidationsClColumn
	Amount VDBConsolidationsClColumn
}{
	VDBConsolidationClSlot,
	VDBConsolidationClAmount,
}

// Validator Dashboard EL Deposits Table

type VDBDepositsElColumn int

var _ EnumFactory[VDBDepositsElColumn] = VDBDepositsElColumn(0)

const (
	VDBDepositElBlock VDBDepositsElColumn = iota
	VDBDepositElAmount
)

func (c VDBDepositsElColumn) Int() int {
	return int(c)
}

func (VDBDepositsElColumn) NewFromString(s string) VDBDepositsElColumn {
	switch s {
	case "", "block", "timestamp":
		return VDBDepositElBlock
	case "amount":
		return VDBDepositElAmount
	default:
		return VDBDepositsElColumn(-1)
	}
}

func (c VDBDepositsElColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBDepositElBlock:
		return goqu.I("ed.block_number")
	case VDBDepositElAmount:
		return goqu.I("ed.amount")
	default:
		return nil
	}
}

var VDBDepositsElColumns = struct {
	Block  VDBDepositsElColumn
	Amount VDBDepositsElColumn
}{
	VDBDepositElBlock,
	VDBDepositElAmount,
}

// ----------------
// Validator Dashboard CL Deposits Table

type VDBDepositsClColumn int

var _ EnumFactory[VDBDepositsClColumn] = VDBDepositsClColumn(0)

const (
	VDBDepositClSlot VDBDepositsClColumn = iota
	VDBDepositClAmount
)

func (c VDBDepositsClColumn) Int() int {
	return int(c)
}

func (VDBDepositsClColumn) NewFromString(s string) VDBDepositsClColumn {
	switch s {
	case "", "slot", "timestamp":
		return VDBDepositClSlot
	case "amount":
		return VDBDepositClAmount
	default:
		return VDBDepositsClColumn(-1)
	}
}

func (c VDBDepositsClColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBDepositClSlot:
		return goqu.COALESCE(goqu.I("slot_queued"), goqu.I("slot_processed"))
	case VDBDepositClAmount:
		return goqu.C("amount")
	default:
		return nil
	}
}

var VDBDepositsClColumns = struct {
	Slot   VDBDepositsClColumn
	Amount VDBDepositsClColumn
}{
	VDBDepositClSlot,
	VDBDepositClAmount,
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
	case "", "index":
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
	VDBRocketPoolNode      VDBRocketPoolColumn = iota
	VDBRocketPoolMinipools                     // might be supported later
	VDBRocketPoolCollateral
	VDBRocketPoolEffectiveRpl
	VDBRocketPoolSmoothingPool
)

func (c VDBRocketPoolColumn) Int() int {
	return int(c)
}

func (VDBRocketPoolColumn) NewFromString(s string) VDBRocketPoolColumn {
	switch s {
	case "", "node":
		return VDBRocketPoolNode
	case "collateral":
		return VDBRocketPoolCollateral
	case "minipools":
		return VDBRocketPoolMinipools
	case "effective_rpl":
		return VDBRocketPoolEffectiveRpl
	case "smoothing_pool":
		return VDBRocketPoolSmoothingPool
	default:
		return VDBRocketPoolColumn(-1)
	}
}

func (c VDBRocketPoolColumn) ToExpr() OrderableSortable {
	switch c {
	case VDBRocketPoolNode:
		return goqu.T("n").Col("address")
	case VDBRocketPoolCollateral:
		return goqu.C("rpl_stake")
	case VDBRocketPoolEffectiveRpl:
		return goqu.C("effective_rpl_stake")
	case VDBRocketPoolSmoothingPool:
		return goqu.C("smoothing_pool_opted_in")
	case VDBRocketPoolMinipools:
		return goqu.L("COUNT(mp.address)")

	default:
		return nil
	}
}

var VDBRocketPoolColumns = struct {
	Node          VDBRocketPoolColumn
	Minipools     VDBRocketPoolColumn
	Collateral    VDBRocketPoolColumn
	EffectiveRpl  VDBRocketPoolColumn
	SmoothingPool VDBRocketPoolColumn
}{
	VDBRocketPoolNode,
	VDBRocketPoolMinipools,
	VDBRocketPoolCollateral,
	VDBRocketPoolEffectiveRpl,
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
	case "", "group_id":
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
