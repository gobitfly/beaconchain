package types

type SearchValidator struct {
	Index     uint64
	PublicKey Hash
}

type SearchValidatorsByDepositEnsName struct {
	EnsName    string
	Validators []uint64
}

type SearchValidatorsByDepositAddress struct {
	Address    Hash
	Validators []uint64
}

type SearchValidatorsByWithdrwalCredential struct {
	WithdrawalCredential Hash
	Validators           []uint64
}

type SearchValidatorsByWithrawalEnsName struct {
	EnsName    string
	Validators []uint64
}

type SearchValidatorsByGraffiti struct {
	Graffiti   string
	Validators []uint64
}
