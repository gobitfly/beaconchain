package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ElectraDeposit struct {
	Pubkey hexutil.Bytes `json:"pubkey" db:"pubkey"`
	Amount uint64        `json:"amount,string" db:"amount"`
}

type ElectraConsolidation struct {
	SourceValidatorIndex uint64 `json:"source_index,string" db:"source_index"`
	TargetValidatorIndex uint64 `json:"target_index,string" db:"target_index"`
	Amount               uint64 `db:"amount"`
}

type ElectraExcessBalance struct {
	ValidatorIndex uint64 `json:"validator_index,string" db:"validator_index"`
	Amount         uint64 `json:"amount,string" db:"amount"`
}
