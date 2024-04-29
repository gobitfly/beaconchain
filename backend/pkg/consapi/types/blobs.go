package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBlobSidecarsResponse struct {
	Data []struct {
		BlockRoot       hexutil.Bytes `json:"block_root"`
		Index           uint64        `json:"index,string"`
		Slot            uint64        `json:"slot,string"`
		BlockParentRoot hexutil.Bytes `json:"block_parent_root"`
		ProposerIndex   uint64        `json:"proposer_index,string"`
		KzgCommitment   hexutil.Bytes `json:"kzg_commitment"`
		KzgProof        hexutil.Bytes `json:"kzg_proof"`
		Blob            hexutil.Bytes `json:"blob"`
	}
}
