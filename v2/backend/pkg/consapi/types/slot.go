package types

import "github.com/ethereum/go-ethereum/common/hexutil"

// /eth/v2/beacon/blocks/{block_id}
type StandardBeaconSlotResponse struct {
	Version             string         `json:"version"`
	ExecutionOptimistic bool           `json:"execution_optimistic"`
	Finalized           bool           `json:"finalized"`
	Data                AnySignedBlock `json:"data"`
}

type AnySignedBlock struct {
	Message struct {
		Slot          uint64        `json:"slot,string"`
		ProposerIndex uint64        `json:"proposer_index,string"`
		ParentRoot    hexutil.Bytes `json:"parent_root"`
		StateRoot     hexutil.Bytes `json:"state_root"`
		Body          struct {
			RandaoReveal      hexutil.Bytes      `json:"randao_reveal"`
			Eth1Data          Eth1Data           `json:"eth1_data"`
			Graffiti          hexutil.Bytes      `json:"graffiti"`
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
			BlobKZGCommitments []hexutil.Bytes `json:"blob_kzg_commitments"`
		} `json:"body"`
	} `json:"message"`
	Signature hexutil.Bytes `json:"signature"`
}

type ProposerSlashing struct {
	SignedHeader1 struct {
		Message struct {
			Slot          uint64        `json:"slot,string"`
			ProposerIndex uint64        `json:"proposer_index,string"`
			ParentRoot    hexutil.Bytes `json:"parent_root"`
			StateRoot     hexutil.Bytes `json:"state_root"`
			BodyRoot      hexutil.Bytes `json:"body_root"`
		} `json:"message"`
		Signature hexutil.Bytes `json:"signature"`
	} `json:"signed_header_1"`
	SignedHeader2 struct {
		Message struct {
			Slot          uint64        `json:"slot,string"`
			ProposerIndex uint64        `json:"proposer_index,string"`
			ParentRoot    hexutil.Bytes `json:"parent_root"`
			StateRoot     hexutil.Bytes `json:"state_root"`
			BodyRoot      hexutil.Bytes `json:"body_root"`
		} `json:"message"`
		Signature hexutil.Bytes `json:"signature"`
	} `json:"signed_header_2"`
}

type AttesterSlashing struct {
	Attestation1 struct {
		AttestingIndices []Uint64Str   `json:"attesting_indices"`
		Signature        hexutil.Bytes `json:"signature"`
		Data             struct {
			Slot            uint64        `json:"slot,string"`
			Index           uint16        `json:"index,string"`
			BeaconBlockRoot hexutil.Bytes `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64        `json:"epoch,string"`
				Root  hexutil.Bytes `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64        `json:"epoch,string"`
				Root  hexutil.Bytes `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_1"`
	Attestation2 struct {
		AttestingIndices []Uint64Str   `json:"attesting_indices"`
		Signature        hexutil.Bytes `json:"signature"`
		Data             struct {
			Slot            uint64        `json:"slot,string"`
			Index           uint16        `json:"index,string"`
			BeaconBlockRoot hexutil.Bytes `json:"beacon_block_root"`
			Source          struct {
				Epoch uint64        `json:"epoch,string"`
				Root  hexutil.Bytes `json:"root"`
			} `json:"source"`
			Target struct {
				Epoch uint64        `json:"epoch,string"`
				Root  hexutil.Bytes `json:"root"`
			} `json:"target"`
		} `json:"data"`
	} `json:"attestation_2"`
}

func (a *AttesterSlashing) GetSlashedIndices() []uint64 {
	commonIndices := make([]uint64, 0)
	indexMap := make(map[int32]bool)

	for _, index := range a.Attestation1.AttestingIndices {
		indexMap[int32(index)] = true
	}

	for _, index := range a.Attestation2.AttestingIndices {
		if indexMap[int32(index)] {
			commonIndices = append(commonIndices, uint64(index))
		}
	}
	return commonIndices
}

type Attestation struct {
	AggregationBits hexutil.Bytes `json:"aggregation_bits"`
	Signature       hexutil.Bytes `json:"signature"`
	Data            struct {
		Slot            uint64        `json:"slot,string"`
		Index           uint16        `json:"index,string"`
		BeaconBlockRoot hexutil.Bytes `json:"beacon_block_root"`
		Source          struct {
			Epoch uint64        `json:"epoch,string"`
			Root  hexutil.Bytes `json:"root"`
		} `json:"source"`
		Target struct {
			Epoch uint64        `json:"epoch,string"`
			Root  hexutil.Bytes `json:"root"`
		} `json:"target"`
	} `json:"data"`
}

type Deposit struct {
	Proof []hexutil.Bytes `json:"proof"`
	Data  struct {
		Pubkey                hexutil.Bytes `json:"pubkey"`
		WithdrawalCredentials hexutil.Bytes `json:"withdrawal_credentials"`
		Amount                uint64        `json:"amount,string"`
		Signature             hexutil.Bytes `json:"signature"`
	} `json:"data"`
}

type VoluntaryExit struct {
	Message struct {
		Epoch          uint64 `json:"epoch,string"`
		ValidatorIndex uint64 `json:"validator_index,string"`
	} `json:"message"`
	Signature hexutil.Bytes `json:"signature"`
}

type Eth1Data struct {
	DepositRoot  hexutil.Bytes `json:"deposit_root"`
	DepositCount uint64        `json:"deposit_count,string"`
	BlockHash    hexutil.Bytes `json:"block_hash"`
}

type SyncAggregate struct {
	SyncCommitteeBits      hexutil.Bytes `json:"sync_committee_bits"`
	SyncCommitteeSignature hexutil.Bytes `json:"sync_committee_signature"`
}

// https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
// https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
type ExecutionPayload struct {
	ParentHash    hexutil.Bytes   `json:"parent_hash"`
	FeeRecipient  hexutil.Bytes   `json:"fee_recipient"`
	StateRoot     hexutil.Bytes   `json:"state_root"`
	ReceiptsRoot  hexutil.Bytes   `json:"receipts_root"`
	LogsBloom     hexutil.Bytes   `json:"logs_bloom"`
	PrevRandao    hexutil.Bytes   `json:"prev_randao"`
	BlockNumber   uint64          `json:"block_number,string"`
	GasLimit      uint64          `json:"gas_limit,string"`
	GasUsed       uint64          `json:"gas_used,string"`
	Timestamp     uint64          `json:"timestamp,string"`
	ExtraData     hexutil.Bytes   `json:"extra_data"`
	BaseFeePerGas uint64          `json:"base_fee_per_gas,string"`
	BlockHash     hexutil.Bytes   `json:"block_hash"`
	Transactions  []hexutil.Bytes `json:"transactions"`
	// present only after capella
	Withdrawals []WithdrawalPayload `json:"withdrawals"`
	// present only after deneb
	BlobGasUsed   uint64 `json:"blob_gas_used,string"`
	ExcessBlobGas uint64 `json:"excess_blob_gas,string"`
}

type WithdrawalPayload struct {
	Index          uint64        `json:"index,string"`
	ValidatorIndex uint64        `json:"validator_index,string"`
	Address        hexutil.Bytes `json:"address"`
	Amount         uint64        `json:"amount,string"`
}

type SignedBLSToExecutionChange struct {
	Message struct {
		ValidatorIndex     uint64        `json:"validator_index,string"`
		FromBlsPubkey      hexutil.Bytes `json:"from_bls_pubkey"`
		ToExecutionAddress hexutil.Bytes `json:"to_execution_address"`
	} `json:"message"`
	Signature hexutil.Bytes `json:"signature"`
}
