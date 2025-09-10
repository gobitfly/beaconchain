package types

type PostSearchRequest struct {
	Input    string        `json:"input"`
	Networks []interface{} `json:"networks,omitempty" tstype:"(number | string)[]"`
	Types    []string      `json:"types,omitempty" tstype:"('validator_by_index' | 'validator_by_public_key' | 'validator_list' | 'validators_by_deposit_address' | 'validators_by_withdrawal_credential' | 'validators_by_graffiti' | 'address' | 'address_by_ens_name' | 'ens_name' | 'transaction' | 'block' | 'epoch' | 'token' | 'slot' | 'slot_by_block_root' | 'slot_by_state_root')[]"`
}

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

type SearchValidatorsByWithdrawalCredential struct {
	EnsName              string `json:"ens_name,omitempty"`
	WithdrawalCredential string `json:"withdrawal_credential"`
	Count                uint64 `json:"count"`
}

type SearchValidatorsByGraffiti struct {
	Graffiti string `json:"graffiti"`
	Hex      string `json:"hex"`
	Count    uint64 `json:"count"`
}

type SearchAddress struct {
	Address Address `json:"address"`
}

type SearchEnsName struct {
	EnsName string `json:"ens_name"`
}

type SearchTransaction struct {
	TransactionHash Hash `json:"transaction_hash"`
}

type SearchBlock struct {
	BlockNumber uint64 `json:"block_number"`
}

type SearchSlot struct {
	Slot uint64 `json:"slot"`
}

type SearchEpoch struct {
	Epoch uint64 `json:"epoch"`
}

type SearchToken struct {
	Address Address `json:"address"`
	Token   string  `json:"token" tstype:"'ERC20' | 'ERC721' | 'ERC1155'"` // currently only erc20 tokens can be found
}

type SearchResult struct {
	Type    string      `json:"type"`
	ChainId uint64      `json:"chain_id"`
	Value   interface{} `json:"value"`
}

type InternalPostSearchResponse struct {
	Data []SearchResult `json:"data" tstype:"({ type: 'validator'; chain_id: number; value: SearchValidator } | { type: 'validator_list'; chain_id: number; value: SearchValidatorList } | { type: 'validators_by_deposit_address'; chain_id: number; value: SearchValidatorsByDepositAddress } | { type: 'validators_by_withdrawal_credential'; chain_id: number; value: SearchValidatorsByWithdrawalCredential } | { type: 'validators_by_graffiti'; chain_id: number; value: SearchValidatorsByGraffiti } | { type: 'address'; chain_id: number; value: SearchAddress } | { type: 'transaction'; chain_id: number; value: SearchTransaction } | { type: 'block'; chain_id: number; value: SearchBlock } | { type: 'epoch'; chain_id: number; value: SearchEpoch } | { type: 'token'; chain_id: number; value: SearchToken } | { type: 'slot'; chain_id: number; value: SearchSlot } | { type: 'ens_name'; chain_id: number; value: SearchEnsName })[]"`
}
