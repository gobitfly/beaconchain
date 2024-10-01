package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBlobSidecarsResponse struct {
	Data []struct {
		Index             uint64        `json:"index,string"`
		Blob              hexutil.Bytes `json:"blob"`
		KzgCommitment     hexutil.Bytes `json:"kzg_commitment"`
		KzgProof          hexutil.Bytes `json:"kzg_proof"`
		SignedBlockHeader struct {
			Message struct {
				Slot          uint64        `json:"slot,string"`
				ProposerIndex uint64        `json:"proposer_index,string"`
				ParentRoot    hexutil.Bytes `json:"parent_root"`
				StateRoot     hexutil.Bytes `json:"state_root"`
				BodyRoot      hexutil.Bytes `json:"body_root"`
			} `json:"message"`
			Signature hexutil.Bytes `json:"signature"`
		} `json:"signed_block_header"`
		KzgCommitmentInclusionProof []hexutil.Bytes `json:"kzg_commitment_inclusion_proof"`
	}
}
