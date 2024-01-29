package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type MevTag struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	PublicLink  string `json:"public_url"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type Block struct {
	// Old blocks
	Miner          Address          `json:"miner,omitempty"`
	MinerName      string           `json:"miner_name,omitempty"`
	Rewards        *decimal.Decimal `json:"rewards,omitempty"`
	TxFees         *decimal.Decimal `json:"tx_fees,omitempty"`
	GasUsage       *decimal.Decimal `json:"gas_usage,omitempty"`
	GasLimit       uint64           `json:"gas_limit,omitempty"`
	LowestGasPrice *decimal.Decimal `json:"lowest_gas_price,omitempty"`
	// Time
	Difficulty *decimal.Decimal `json:"difficulty,omitempty"`
	BaseFee    *decimal.Decimal `json:"base_fee,omitempty"`
	BurnedFees *decimal.Decimal `json:"burned_fees,omitempty"`
	Extra      string           `json:"extra,omitempty"`
	Hash       Hash             `json:"hash,omitempty"`
	ParentHash Hash             `json:"parent_hash,omitempty"`

	// New blocks
	MevTags                     []MevTag         `json:"mev_tags,omitempty"`
	Epoch                       uint64           `json:"epoch,omitempty"`
	Slot                        uint64           `json:"slot,omitempty"`
	Block                       uint64           `json:"block"`
	Proposer                    uint64           `json:"proposer,omitempty"`
	ProposerRewardEL            *decimal.Decimal `json:"proposer_reward_el,omitempty"`
	ProposerRewardCL            *decimal.Decimal `json:"proposer_reward_cl,omitempty"`
	ProposerRewardRecipient     Address          `json:"proposer_reward_recipient,omitempty"`
	ProposerRewardRecipientName string           `json:"proposer_reward_recipient_name,omitempty"`
	Status                      string           `json:"status,omitempty"`
	Time                        time.Time        `json:"time,omitempty"`
	PriorityFees                *decimal.Decimal `json:"priority_fees,omitempty"`
	Transactions                uint16           `json:"transactions,omitempty"`
	TransactionsInternal        uint16           `json:"transcations_internal,omitempty"`
	TransactionsBlobs           uint32           `json:"transcations_blob,omitempty"`
	BlockRoot                   Hash             `json:"block_root,omitempty"`
	ParentRoot                  Hash             `json:"parent_root,omitempty"`

	ExecutionPayload ExecutionPayloadInfo `json:"execution_payload,omitempty"`
	ConsensusLayer   ConsensusLayerInfo   `json:"consensus_layer,omitempty"`
	SyncCommittee    SyncCommitteeInfo    `json:"sync_committee,omitempty"`
	Eth1Data         Eth1DataInfo         `json:"eth1_data,omitempty"`
}

type ExecutionPayloadInfo struct {
	// Block
	BlockHash     Hash             `json:"block_hash"`
	ParentHash    Hash             `json:"parent_hash"`
	PriorityFees  *decimal.Decimal `json:"priority_fees"`
	Recipient     Address          `json:"priority_fees_recipient"`
	RecipientName string           `json:"priority_fees_recipient_name,omitempty"`
	GasUsed       uint64           `json:"gas_used"`
	GasLimit      uint64           `json:"gas_limit"`
	BaseFeePerGas *decimal.Decimal `json:"base_fee_per_gas"`
	BaseFees      *decimal.Decimal `json:"base_fees"`
	// Transactions
	// Time
	ExtraData string `json:"extra_data,omitempty"`
	Graffiti  string `json:"graffiti"`
}

type ConsensusLayerInfo struct {
	StateRoot         Hash   `json:"state_root"`
	Signature         Hash   `json:"signature"`
	RandaoReveal      Hash   `json:"randao_reveal"`
	Attestations      uint64 `json:"attestations"`
	Votes             uint64 `json:"votes"`
	VotingValidators  uint64 `json:"voting_validators"`
	VoluntaryExits    uint64 `json:"voluntary_exits"`
	AttesterSlashings uint64 `json:"attester_slashings"`
	ProposerSlashings uint64 `json:"proposer_slashings"`
	Deposits          uint64 `json:"deposit"`
}

type SyncCommitteeInfo struct {
	Participation float64  `json:"participation"`
	Bits          []byte   `json:"bits"`
	SyncCommittee []uint64 `json:"sync_committee"`
	Signature     []byte   `json:"signature"`
}

type Eth1DataInfo struct {
	BlockHash    string `json:"block_hash"`
	DepositCount string `json:"deposit_count"`
	DepositRoot  []byte `json:"deposit_root"`
}

type BlobInfo struct {
	Count      uint64           `json:"name"`
	TxCount    uint64           `json:"tx_count"`
	GasUsed    uint64           `json:"gas_used"`
	GasPrice   *decimal.Decimal `json:"gas_price"`
	ExcessGas  uint64           `json:"excess_gas"`
	BurnedFees *decimal.Decimal `json:"burned_fees"`
}
