package types

import "github.com/ethereum/go-ethereum/common/hexutil"

const (
	// Node statuses
	PendingInitialized ValidatorStatus = "pending_initialized"
	PendingQueued      ValidatorStatus = "pending_queued"
	ActiveOngoing      ValidatorStatus = "active_ongoing"
	ActiveExiting      ValidatorStatus = "active_exiting"
	ActiveSlashed      ValidatorStatus = "active_slashed"
	ExitedUnslashed    ValidatorStatus = "exited_unslashed"
	ExitedSlashed      ValidatorStatus = "exited_slashed"
	WithdrawalPossible ValidatorStatus = "withdrawal_possible"
	WithdrawalDone     ValidatorStatus = "withdrawal_done"
	Active             ValidatorStatus = "active"
	Pending            ValidatorStatus = "pending"
	Exited             ValidatorStatus = "exited"
	Withdrawal         ValidatorStatus = "withdrawal"

	// Db statuses
	DbSlashed         ValidatorDbStatus = "slashed"
	DbExited          ValidatorDbStatus = "exited"
	DbDeposited       ValidatorDbStatus = "deposited"
	DbPending         ValidatorDbStatus = "pending"
	DbSlashingOffline ValidatorDbStatus = "slashing_offline"
	DbSlashingOnline  ValidatorDbStatus = "slashing_online"
	DbExitingOffline  ValidatorDbStatus = "exiting_offline"
	DbExitingOnline   ValidatorDbStatus = "exiting_online"
	DbActiveOffline   ValidatorDbStatus = "active_offline"
	DbActiveOnline    ValidatorDbStatus = "active_online"
)

// eth/v1/beacon/states/{state_id}/validators
type StandardValidatorsResponse struct {
	ExecutionOptimistic bool                `json:"execution_optimistic"`
	Finalized           bool                `json:"finalized"`
	Data                []StandardValidator `json:"data"`
}

// eth/v1/beacon/states/{state_id}/validators/{validator_id}
type StandardSingleValidatorsResponse struct {
	ExecutionOptimistic bool              `json:"execution_optimistic"`
	Finalized           bool              `json:"finalized"`
	Data                StandardValidator `json:"data"`
}

type StandardValidator struct {
	Index     uint64          `json:"index,string"`
	Balance   uint64          `json:"balance,string"`
	Status    ValidatorStatus `json:"status"`
	Validator struct {
		Pubkey                     hexutil.Bytes `json:"pubkey"`
		WithdrawalCredentials      hexutil.Bytes `json:"withdrawal_credentials"`
		EffectiveBalance           uint64        `json:"effective_balance,string"`
		Slashed                    bool          `json:"slashed"`
		ActivationEligibilityEpoch uint64        `json:"activation_eligibility_epoch,string"`
		ActivationEpoch            uint64        `json:"activation_epoch,string"`
		ExitEpoch                  uint64        `json:"exit_epoch,string"`
		WithdrawableEpoch          uint64        `json:"withdrawable_epoch,string"`
	} `json:"validator"`
}

// /eth/v1/validator/duties/proposer/{epoch}
type StandardProposerAssignmentsResponse struct {
	DependentRoot       hexutil.Bytes `json:"dependent_root"`
	ExecutionOptimistic bool          `json:"execution_optimistic"`
	Data                []struct {
		Pubkey         hexutil.Bytes `json:"pubkey"`
		ValidatorIndex uint64        `json:"validator_index,string"`
		Slot           int64         `json:"slot,string"`
	} `json:"data"`
}

// /eth/v1/beacon/states/{state_id}/validator_balances
type StandardValidatorBalancesResponse struct {
	Data []struct {
		Index   uint64 `json:"index,string"`
		Balance uint64 `json:"balance,string"`
	} `json:"data"`
}
