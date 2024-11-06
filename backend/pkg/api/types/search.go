package types

// search types to be used between the data access layer and the api layer, shouldn't be exported to typescript

type SearchValidator struct {
	Index     uint64
	PublicKey []byte
}

type SearchValidatorsByDepositEnsName struct {
	EnsName string
	Address []byte
	Count   uint64
}

type SearchValidatorsByDepositAddress struct {
	Address []byte
	Count   uint64
}

type SearchValidatorsByWithdrwalCredential struct {
	WithdrawalCredential []byte
	Count                uint64
}

type SearchValidatorsByWithrawalEnsName struct {
	EnsName string
	Address []byte
	Count   uint64
}

type SearchValidatorsByGraffiti struct {
	Graffiti string
	Count    uint64
}
