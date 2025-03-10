package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBeaconStateResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Balances   []Uint64Str `json:"balances"`
		Validators []struct {
			Pubkey                     hexutil.Bytes `json:"pubkey"`
			WithdrawalCredentials      hexutil.Bytes `json:"withdrawal_credentials"`
			EffectiveBalance           uint64        `json:"effective_balance,string"`
			Slashed                    bool          `json:"slashed"`
			ActivationEligibilityEpoch uint64        `json:"activation_eligibility_epoch,string"`
			ActivationEpoch            uint64        `json:"activation_epoch,string"`
			ExitEpoch                  uint64        `json:"exit_epoch,string"`
			WithdrawalbleEpoch         uint64        `json:"withdrawable_epoch,string"`
		} `json:"validators"`
		PendingConsolidations []struct {
			SourceIndex uint64 `json:"source_index,string"`
			TargetIndex uint64 `json:"target_index,string"`
		} `json:"pending_consolidations"`
		PendingPartialWithdrawals []struct {
			ValidatorIndex     uint64 `json:"validator_index,string"`
			Amount             uint64 `json:"amount,string"`
			WithdrawableEpopch uint64 `json:"withdrawable_epoch,string"`
		} `json:"pending_partial_withdrawals"`
		PendingDeposits []struct {
			Pubkey                hexutil.Bytes `json:"pubkey"`
			WithdrawalCredentials hexutil.Bytes `json:"withdrawal_credentials"`
			Amount                uint64        `json:"amount,string"`
			Signature             hexutil.Bytes `json:"signature"`
			Slot                  uint64        `json:"slot,string"`
		} `json:"pending_deposits"`
	} `json:"data"`
}
