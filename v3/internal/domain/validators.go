package domain

import "github.com/shopspring/decimal"

type ValidatorIndex = int
type PublicKey = []byte
type ValidatorsByIdentifier struct {
	Indices    []ValidatorIndex
	PublicKeys []PublicKey
}
type ValidatorsSelector struct {
	Identifiers          *ValidatorsByIdentifier
	DepositAddress       *EthereumAddress
	WithdrawalAddress    *EthereumAddress
	WithdrawalCredential *WithdrawalCredential
}

type ValidatorIndexCursor struct {
	Index ValidatorIndex `json:"i"`
}

type ValidatorBalance struct {
	ValidatorIndex     ValidatorIndex
	ValidatorPublicKey PublicKey
	CurrentBalance     decimal.Decimal
	EffectiveBalance   decimal.Decimal
}

type ValidatorOverview struct {
	ValidatorIndex             ValidatorIndex
	ValidatorPublicKey         PublicKey
	Slashed                    bool
	Online                     *bool
	WithdrawalCredential       WithdrawalCredential
	ActivationEligibilityEpoch *int
	ActivationEpoch            *int
	ExitEpoch                  *int
	WithdrawableEpoch          *int
}
