package types

import "github.com/shopspring/decimal"

// /eth/v2/beacon/blocks/{block_id}
type StandardBeaconSlotResponse struct {
	Version             string         `json:"version"`
	ExecutionOptimistic bool           `json:"execution_optimistic"`
	Finalized           bool           `json:"finalized"`
	Data                AnySignedBlock `json:"data"`
}

type AnySignedBlock struct {
	Message struct {
		Slot          uint64 `json:"slot,string"`
		ProposerIndex uint64 `json:"proposer_index,string"`
		ParentRoot    string `json:"parent_root"`
		StateRoot     string `json:"state_root"`
		Body          struct {
			RandaoReveal      string             `json:"randao_reveal"`
			Eth1Data          Eth1Data           `json:"eth1_data"`
			Graffiti          string             `json:"graffiti"`
			ProposerSlashings []ProposerSlashing `json:"proposer_slashings"`
			AttesterSlashings []AttesterSlashing `json:"attester_slashings"`
			Attestations      []Attestation      `json:"attestations"`
			Deposits          []Deposit          `json:"deposits"`
			VoluntaryExits    []VoluntaryExit    `json:"voluntary_exits"`

			// not present in phase0 blocks
			SyncAggregate *SyncAggregate `json:"sync_aggregate,omitempty"`

			// not present in phase0/altair blocks
			ExecutionPayload *ExecutionPayload `json:"execution_payload"`

			// present only after capella
			SignedBLSToExecutionChange []*SignedBLSToExecutionChange `json:"bls_to_execution_changes"`

			// present only after deneb
			BlobKZGCommitments []bytesHexStr `json:"blob_kzg_commitments"`
		} `json:"body"`
	} `json:"message"`
	Signature bytesHexStr `json:"signature"`
}

type ProposerSlashing struct {
	SignedHeader1 struct {
		Message struct {
			Slot          uint64 `json:"slot,string"`
			ProposerIndex uint64 `json:"proposer_index,string"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			BodyRoot      string `json:"body_root"`
		} `json:"message"`
		Signature bytesHexStr `json:"signature"`
	} `json:"signed_header_1"`
	SignedHeader2 struct {
		Message struct {
			Slot          uint64 `json:"slot,string"`
			ProposerIndex uint64 `json:"proposer_index,string"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			BodyRoot      string `json:"body_root"`
		} `json:"message"`
		Signature bytesHexStr `json:"signature"`
	} `json:"signed_header_2"`
}

type AttesterSlashing struct {
	Attestation1 struct {
		AttestingIndices []Uint64Str `json:"attesting_indices"`
		Signature        bytesHexStr `json:"signature"`
		Data             struct {
			Slot            uint64 `json:"slot,string"`
			Index           uint64 `json:"index,string"`
			BeaconBlockRoot string `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64 `json:"epoch,string"`
				Root  string `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64 `json:"epoch,string"`
				Root  string `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_1"`
	Attestation2 struct {
		AttestingIndices []Uint64Str `json:"attesting_indices"`
		Signature        bytesHexStr `json:"signature"`
		Data             struct {
			Slot            uint64 `json:"slot,string"`
			Index           uint64 `json:"index,string"`
			BeaconBlockRoot string `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64 `json:"epoch,string"`
				Root  string `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64 `json:"epoch,string"`
				Root  string `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_2"`
}

type Attestation struct {
	AggregationBits string      `json:"aggregation_bits"`
	Signature       bytesHexStr `json:"signature"`
	Data            struct {
		Slot            uint64 `json:"slot,string"`
		Index           uint64 `json:"index,string"`
		BeaconBlockRoot string `json:"beacon_block_root"`
		Source          struct {
			Epoch uint64 `json:"epoch,string"`
			Root  string `json:"root"`
		} `json:"source"`
		Target struct {
			Epoch uint64 `json:"epoch,string"`
			Root  string `json:"root"`
		} `json:"target"`
	} `json:"data"`
}

type Deposit struct {
	Proof []string `json:"proof"`
	Data  struct {
		Pubkey                string          `json:"pubkey"`
		WithdrawalCredentials bytesHexStr     `json:"withdrawal_credentials"`
		Amount                decimal.Decimal `json:"amount"`
		Signature             bytesHexStr     `json:"signature"`
	} `json:"data"`
}

type VoluntaryExit struct {
	Message struct {
		Epoch          uint64 `json:"epoch,string"`
		ValidatorIndex uint64 `json:"validator_index,string"`
	} `json:"message"`
	Signature bytesHexStr `json:"signature"`
}

type Eth1Data struct {
	DepositRoot  string `json:"deposit_root"`
	DepositCount uint64 `json:"deposit_count,string"`
	BlockHash    string `json:"block_hash"`
}

type SyncAggregate struct {
	SyncCommitteeBits      string `json:"sync_committee_bits"`
	SyncCommitteeSignature string `json:"sync_committee_signature"`
}

// https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
// https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
type ExecutionPayload struct {
	ParentHash    bytesHexStr   `json:"parent_hash"`
	FeeRecipient  bytesHexStr   `json:"fee_recipient"`
	StateRoot     bytesHexStr   `json:"state_root"`
	ReceiptsRoot  bytesHexStr   `json:"receipts_root"`
	LogsBloom     bytesHexStr   `json:"logs_bloom"`
	PrevRandao    bytesHexStr   `json:"prev_randao"`
	BlockNumber   uint64        `json:"block_number,string"`
	GasLimit      uint64        `json:"gas_limit,string"`
	GasUsed       uint64        `json:"gas_used,string"`
	Timestamp     uint64        `json:"timestamp,string"`
	ExtraData     bytesHexStr   `json:"extra_data"`
	BaseFeePerGas uint64        `json:"base_fee_per_gas,string"`
	BlockHash     bytesHexStr   `json:"block_hash"`
	Transactions  []bytesHexStr `json:"transactions"`
	// present only after capella
	Withdrawals []WithdrawalPayload `json:"withdrawals"`
	// present only after deneb
	BlobGasUsed   uint64 `json:"blob_gas_used,string"`
	ExcessBlobGas uint64 `json:"excess_blob_gas,string"`
}

type WithdrawalPayload struct {
	Index          uint64          `json:"index,string"`
	ValidatorIndex uint64          `json:"validator_index,string"`
	Address        bytesHexStr     `json:"address"`
	Amount         decimal.Decimal `json:"amount"`
}

type SignedBLSToExecutionChange struct {
	Message struct {
		ValidatorIndex     uint64      `json:"validator_index,string"`
		FromBlsPubkey      bytesHexStr `json:"from_bls_pubkey"`
		ToExecutionAddress bytesHexStr `json:"to_execution_address"`
	} `json:"message"`
	Signature bytesHexStr `json:"signature"`
}
