package domain

import "github.com/shopspring/decimal"

type ValidatorsByIdentifier struct {
	Indices    []int
	PublicKeys [][]byte
}
type ValidatorsSelector struct {
	Identifiers          *ValidatorsByIdentifier
	DepositAddress       *EthereumAddress
	WithdrawalAddress    *EthereumAddress
	WithdrawalCredential *WithdrawalCredential
}

type ValidatorIndexCursor struct {
	Index int
}

type ValidatorBalance struct {
	ValidatorIndex     int
	ValidatorPublicKey []byte
	CurrentBalance     decimal.Decimal
	EffectiveBalance   decimal.Decimal
}
