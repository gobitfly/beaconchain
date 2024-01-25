package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type Block struct {
	Number        uint64           `json:"number"`
	TxCount       uint16           `json:"tx_count"`
	UncleCount    uint16           `json:"uncle_count,omitempty"`
	Hash          Hash             `json:"hash"`
	ParentHash    Hash             `json:"parent_hash"`
	TxFees        *decimal.Decimal `json:"tx_fees"`
	GasUsed       uint64           `json:"gas_used"`
	GasLimit      uint64           `json:"gas_limit"`
	Ts            time.Time        `json:"ts"`
	State         string           `json:"state"`
	Extra         string           `json:"extra,omitempty"`
	DepositsCount uint64           `json:"deposits_count,omitempty"`

	Pow     PoWInfo     `json:"pow,omitempty"`
	Eip1559 Eip1559Info `json:"eip1559,omitempty"`
	Pos     PosInfo     `json:"pos,omitempty"`
	Blob    BlobInfo    `json:"blob,omitempty"` // eip 4844
}

type PoWInfo struct {
	Miner          Address          `json:"miner"`
	LowestGasPrice *decimal.Decimal `json:"lowest_gas_price"`
	Difficulty     *decimal.Decimal `json:"difficulty"`
}

type Eip1559Info struct {
	BurnedFees   *decimal.Decimal `json:"burned_fees"`
	BurnedTxFees *decimal.Decimal `json:"burned_tx_fees"`
	BaseFee      *decimal.Decimal `json:"base_fee"`
}

type PosInfo struct {
	Epoch              uint64  `json:"epoch"`
	EpochFinalized     bool    `json:"epoch_finalized"`
	PrevEpochFinalized bool    `json:"prev_epoch_finalized"`
	Proposer           uint64  `json:"proposer"`
	ProposerName       string  `json:"proposer_name,omitempty"`
	FeeRecipient       Address `json:"fee_recipient"`
	BlockRoot          Hash    `json:"block_root"`
	ParentRoot         Hash    `json:"parent_root"`
	StateRoot          Hash    `json:"state_root"`
	Signature          Hash    `json:"signature"`
	RandaoReveal       Hash    `json:"randao_reveal"`
	Graffiti           string  `json:"graffiti"`

	Eth1Data struct {
		DepositRoot   string `json:"deposit_root"`
		DepositsCount string `json:"deposits_count"`
		Hash          string `json:"hash"`
	} `json:"eth1data"`

	SyncAggregate struct {
		Bits          []byte  `json:"bits"`
		Signature     []byte  `json:"signature"`
		Participation float64 `json:"participation"`
	} `json:"sync_aggregate"`

	ProposerSlashingsCount uint64 `json:"proposer_slashings_count,omitempty"`
	AttesterSlashingsCount uint64 `json:"attester_slashings_count,omitempty"`
	AttestationsCount      uint64 `json:"attestations_count"`
	DepositCount           uint64 `json:"deposit_count"`
	WithdrawalCount        uint64 `json:"withdrawal_count"`
	BLSChangeCount         uint64 `json:"bls_change_count"`
	VoluntaryExitsCount    uint64 `json:"voluntary_exits_count"`
	VotesCount             uint64 `json:"votes_count"`
	VotingValidatorsCount  uint64 `json:"voting_validators_count"`

	MevTags          []MevTag         `json:"mev_tags,omitempty"`
	MevReward        *decimal.Decimal `json:"mev_reward"`
	MevBribe         *decimal.Decimal `json:"mev_bribe"`
	MevIsValid       bool             `json:"mev_is_valid"`
	MevRecipient     Address          `json:"mev_recipient"`
	MevRecipientName string           `json:"mev_recipient_name,omitempty"`
}

type MevTag struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	PublicLink  string `json:"public_url"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type BlobInfo struct {
	Count      uint64           `json:"name"`
	TxCount    uint64           `json:"tx_count"`
	GasUsed    uint64           `json:"gas_used"`
	GasPrice   *decimal.Decimal `json:"gas_price"`
	ExcessGas  uint64           `json:"excess_gas"`
	BurnedFees *decimal.Decimal `json:"burned_fees"`
}
