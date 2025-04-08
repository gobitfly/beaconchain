package types

import (
	"encoding/base64"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

type Base64Bytes []byte

func (b Base64Bytes) String() string {
	return hexutil.Bytes(b).String()
}

func (b Base64Bytes) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b).MarshalText()
}

func (b *Base64Bytes) UnmarshalText(text []byte) error {
	decoded, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return errors.Wrap(err, "failed to decode base64")
	}
	if len(decoded) == 0 {
		return fmt.Errorf("base64 decoded string is empty")
	}
	*b = Base64Bytes(decoded)
	return nil
}

type ElectraDeposit struct {
	Pubkey         Base64Bytes `json:"pubkey" db:"pubkey"`
	Amount         int64       `json:"amount,string" db:"amount"`
	SignatureValid bool        `json:"signature_valid" db:"signature_valid"`
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
