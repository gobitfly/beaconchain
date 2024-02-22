package types

type StandardBlobSidecarsResponse struct {
	Data []struct {
		BlockRoot       bytesHexStr `json:"block_root"`
		Index           uint64      `json:"index,string"`
		Slot            uint64      `json:"slot,string"`
		BlockParentRoot bytesHexStr `json:"block_parent_root"`
		ProposerIndex   uint64      `json:"proposer_index,string"`
		KzgCommitment   bytesHexStr `json:"kzg_commitment"`
		KzgProof        bytesHexStr `json:"kzg_proof"`
		Blob            bytesHexStr `json:"blob"`
	}
}
