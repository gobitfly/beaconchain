package types

import (
	"github.com/shopspring/decimal"
)

// count indicator per block details tab; each tab is only present if count > 0
type BlockSummary struct {
	Transactions   uint64 `json:"transactions"`
	Votes          uint64 `json:"votes"`
	Attestations   uint64 `json:"attestations"`
	Withdrawals    uint64 `json:"withdrawals"`
	BlsChanges     uint64 `json:"bls_changes"`
	VoluntaryExits uint64 `json:"voluntary_exits"`
	Blobs          uint64 `json:"blobs"`
}

type InternalGetBlockResponse ApiDataResponse[BlockSummary]

type BlockMevTag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type BlockExecutionPayload struct {
	Block                 uint64          `json:"block"`
	BlockHash             Hash            `json:"block_hash"`
	ParentHash            Hash            `json:"parent_hash"`
	PriorityFeesRecipient Address         `json:"priority_fees_recipient"`
	PriorityFees          decimal.Decimal `json:"priority_fees"`
	GasUsed               uint64          `json:"gas_used"`
	GasLimit              uint64          `json:"gas_limit"`
	BaseFeePerGas         decimal.Decimal `json:"base_fee_per_gas"`
	BaseFees              decimal.Decimal `json:"base_fees"`
	Transactions          struct {
		General uint64 `json:"general"`
		Blob    uint64 `json:"blob"`
	} `json:"transactions"`
	Time      int64  `json:"time"`
	ExtraData string `json:"extra_data,omitempty"`
	Graffiti  string `json:"graffiti"`
}

type BlockConsensusLayer struct {
	StateRoot         Hash               `json:"state_root"`
	Signature         Hash               `json:"signature"`
	RandaoReveal      Hash               `json:"randao_reveal"`
	Attestations      uint64             `json:"attestations"`
	Votes             uint64             `json:"votes"`
	VotingValidators  uint64             `json:"voting_validators"`
	VoluntaryExits    uint64             `json:"voluntary_exits"`
	AttesterSlashings uint64             `json:"attester_slashings"`
	ProposerSlashings uint64             `json:"proposer_slashings"`
	Deposits          uint64             `json:"deposit"`
	SyncCommittee     BlockSyncCommittee `json:"sync_committee"`
	Eth1Data          BlockEth1Data      `json:"eth1_data"`
}

type BlockSyncCommittee struct {
	Participation float64  `json:"participation"`
	Bits          []bool   `json:"bits"`
	SyncCommittee []uint64 `json:"sync_committee"`
	Signature     Hash     `json:"signature"`
}

type BlockEth1Data struct {
	BlockHash    Hash   `json:"block_hash"`
	DepositCount uint64 `json:"deposit_count"`
	DepositRoot  Hash   `json:"deposit_root"`
}

type BlobInfo struct {
	Count      uint64           `json:"name"`
	TxCount    uint64           `json:"tx_count"`
	GasUsed    uint64           `json:"gas_used"`
	GasPrice   *decimal.Decimal `json:"gas_price"`
	ExcessGas  uint64           `json:"excess_gas"`
	BurnedFees *decimal.Decimal `json:"burned_fees"`
}

type BlockOverview struct {
	// General
	Block uint64 `json:"block"`
	Time  int64  `json:"time"`

	// Old blocks only
	Miner    *Address         `json:"miner,omitempty"`
	Rewards  *decimal.Decimal `json:"rewards,omitempty"`
	TxFees   *decimal.Decimal `json:"tx_fees,omitempty"`
	GasUsage *decimal.Decimal `json:"gas_usage,omitempty"`
	GasLimit *struct {
		Value   uint64  `json:"value"`
		Percent float64 `json:"percent"`
	} `json:"gas_limit,omitempty"`
	LowestGasPrice *decimal.Decimal `json:"lowest_gas_price,omitempty"`
	Difficulty     *decimal.Decimal `json:"difficulty,omitempty"`
	// base + burned fee only present post EIP-1559
	BaseFee    *decimal.Decimal `json:"base_fee,omitempty"`
	BurnedFees *decimal.Decimal `json:"burned_fees,omitempty"`
	Extra      string           `json:"extra,omitempty"`
	Hash       Hash             `json:"hash,omitempty"`
	ParentHash Hash             `json:"parent_hash,omitempty"`

	// New blocks only
	MevTags                 []BlockMevTag               `json:"mev_tags,omitempty"`
	Epoch                   uint64                      `json:"epoch,omitempty"`
	Slot                    uint64                      `json:"slot,omitempty"`
	Proposer                uint64                      `json:"proposer,omitempty"`
	ProposerReward          *ClElValue[decimal.Decimal] `json:"proposer_reward,omitempty"`
	ProposerRewardRecipient *Address                    `json:"proposer_reward_recipient,omitempty"`
	Status                  *struct {
		Proposal  string `json:"proposal"`  // proposed, orphaned, missed, scheduled
		Finalized string `json:"finalized"` // finalized, justified, not_finalized
	} `json:"status,omitempty"`
	PriorityFees *decimal.Decimal `json:"priority_fees,omitempty"`
	Transactions *struct {
		General  uint64 `json:"general"`
		Internal uint64 `json:"internal"`
		Blob     uint64 `json:"blob"`
	} `json:"transactions,omitempty"`
	BlockRoot  Hash `json:"block_root,omitempty"`
	ParentRoot Hash `json:"parent_root,omitempty"`

	ExecutionPayload *BlockExecutionPayload `json:"execution_payload,omitempty"`
	ConsensusLayer   *BlockConsensusLayer   `json:"consensus_layer,omitempty"`
}

type InternalGetBlockOverviewResponse ApiDataResponse[BlockOverview]

type BlockTransactionTableRow struct {
	Success  bool            `json:"success"`
	TxHash   Hash            `json:"tx_hash"`
	Method   string          `json:"method"`
	Block    uint64          `json:"block"`
	Age      uint64          `json:"age"`
	From     ContractAddress `json:"from"`
	Type     string          `json:"type"` // in, out, self, contract
	To       ContractAddress `json:"to"`
	Value    decimal.Decimal `json:"value"`
	GasPrice decimal.Decimal `json:"gas_price"`
	TxFee    decimal.Decimal `json:"tx_fee"`
}

type InternalGetBlockTransactionsResponse ApiDataResponse[[]BlockTransactionTableRow]

type BlockVoteTableRow struct {
	AllocatedSlot   uint64   `json:"allocated_slot"`
	Committee       uint64   `json:"committee"`
	IncludedInBlock uint64   `json:"included_in_block"`
	Validators      []uint64 `json:"validators"`
}

type InternalGetBlockVotesResponse ApiDataResponse[[]BlockVoteTableRow]

type EpochInfo struct {
	Epoch     uint64 `json:"epoch"`
	BlockRoot Hash   `json:"block_root"`
}

type BlockAttestationTableRow struct {
	Slot            uint64    `json:"slot"`
	CommitteeIndex  uint64    `json:"committee_index"`
	AggregationBits []bool    `json:"aggregation_bits"`
	Validators      []uint64  `json:"validators"`
	BeaconBlockRoot Hash      `json:"beacon_block_root"`
	Source          EpochInfo `json:"source"`
	Target          EpochInfo `json:"target"`
	Signature       Hash      `json:"signature"`
}

type InternalGetBlockAttestationsResponse ApiDataResponse[[]BlockAttestationTableRow]

type BlockWithdrawalTableRow struct {
	// TODO
}

type InternalGetBlockWtihdrawalsResponse ApiDataResponse[[]BlockWithdrawalTableRow]

type BlockBlsChangeTableRow struct {
	Index                uint64          `json:"index"`
	Signature            Hash            `json:"signature"`
	BlsPubkey            Hash            `json:"bls_pubkey"`
	NewWithdrawalAddress ContractAddress `json:"new_withdrawal_address"`
}

type InternalGetBlockBlsChangeResponse ApiDataResponse[[]BlockBlsChangeTableRow]

type BlockVoluntaryExitTableRow struct {
	Validator uint64 `json:"validator"`
	Signature Hash   `json:"signature"`
}

type InternalGetBlockVoluntaryExitsResponse ApiDataResponse[[]BlockVoluntaryExitTableRow]

type BlockBlobTableRow struct {
	VersionedHash   Hash   `json:"versioned_hash"`
	Commitment      Hash   `json:"commitment"`
	Proof           Hash   `json:"proof"`
	Size            uint64 `json:"size"`
	TransactionHash Hash   `json:"transaction_hash"`
	Block           uint64 `json:"block"`
	Data            []byte `json:"data"`
}

type InternalGetBlockBlobsResponse ApiDataResponse[[]BlockBlobTableRow]
