package types

// search types to be used between the data access layer and the api layer, shouldn't be exported to typescript

type SearchValidator struct {
	Index     uint64 `json:"index"`
	PublicKey string `json:"public_key"`
}

type SearchValidatorList struct {
	Validators []uint64 `json:"validators"`
}

type SearchValidatorsByDepositAddress struct {
	EnsName        string `json:"ens_name,omitempty"`
	DepositAddress string `json:"deposit_address"`
	Count          uint64 `json:"count"`
}

type SearchValidatorsByWithdrwalCredential struct {
	EnsName              string `json:"ens_name,omitempty"`
	WithdrawalCredential string `json:"withdrawal_credential"`
	Count                uint64 `json:"count"`
}

type SearchValidatorsByGraffiti struct {
	Graffiti string `json:"graffiti"`
	Count    uint64 `json:"count"`
}

type SearchResult struct {
	Type    string      `json:"type"`
	ChainId uint64      `json:"chain_id"`
	Value   interface{} `json:"value"`
}

type InternalPostSearchResponse struct {
	Data []SearchResult `json:"data" tstype:"({ type: 'validator'; chain_id: number; value: SearchValidator } | { type: 'validator_list'; chain_id: number; value: SearchValidatorList } | { type: 'validators_by_deposit_address'; chain_id: number; value: SearchValidatorsByDepositAddress } | { type: 'validators_by_withdrawal_credential'; chain_id: number; value: SearchValidatorsByWithdrwalCredential } | { type: 'validators_by_graffiti'; chain_id: number; value: SearchValidatorsByGraffiti })[]"`
}
