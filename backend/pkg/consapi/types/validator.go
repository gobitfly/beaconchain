package types

import "github.com/ethereum/go-ethereum/common/hexutil"

const (
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
)

// eth/v1/beacon/states/{state_id}/validators
type StandardValidatorsResponse struct {
	//ExecutionOptimistic bool                `json:"execution_optimistic"`
	//Finalized           bool                `json:"finalized"`
	Data []StandardValidator `json:"data"`
}

type LightStandardValidatorsResponse struct {
	Epoch uint64
	Data  []LightStandardValidator
}

type UltraLightStandardValidatorsResponse struct {
	Data []UltraLightStandardValidator
}
type UltraLightStandardValidator struct {
	Index            uint64
	EffectiveBalance uint64
	Status           ValidatorStatus
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
		ActivationEligibilityEpoch uint64        `json:"activation_eligibility_epoch,string"` // drop, not used
		ActivationEpoch            uint64        `json:"activation_epoch,string"`             // drop, not used
		ExitEpoch                  uint64        `json:"exit_epoch,string"`                   // drop, not used
		WithdrawableEpoch          uint64        `json:"withdrawable_epoch,string"`           // drop, not used
	} `json:"validator"`
}

type LightStandardValidator struct {
	Index            uint64
	Balance          uint64
	Status           ValidatorStatus
	Pubkey           hexutil.Bytes
	EffectiveBalance uint64
	Slashed          bool
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
