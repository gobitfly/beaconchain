package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBlobSidecarsResponse struct {
	Data []BlobSidecarsData
}

type BlobSidecarsData struct {
	Index                       uint64                  `json:"index,string"`
	Blob                        hexutil.Bytes           `json:"blob"`
	KzgCommitment               hexutil.Bytes           `json:"kzg_commitment"`
	KzgProof                    hexutil.Bytes           `json:"kzg_proof"`
	SignedBlockHeader           BlobSidecarsBlockHeader `json:"signed_block_header"`
	KzgCommitmentInclusionProof []hexutil.Bytes         `json:"kzg_commitment_inclusion_proof"`
}

type BlobSidecarsBlockHeader struct {
	Message   BlobSidecarsMessage `json:"message"`
	Signature hexutil.Bytes       `json:"signature"`
}

type BlobSidecarsMessage struct {
	Slot          uint64        `json:"slot,string"`
	ProposerIndex uint64        `json:"proposer_index,string"`
	ParentRoot    hexutil.Bytes `json:"parent_root"`
	StateRoot     hexutil.Bytes `json:"state_root"`
	BodyRoot      hexutil.Bytes `json:"body_root"`
}
