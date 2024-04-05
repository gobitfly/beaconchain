package types

type StandardBeaconHeaderResponse struct {
	Data struct {
		Root   bytesHexStr `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64      `json:"slot,string"`
				ProposerIndex uint64      `json:"proposer_index,string"`
				ParentRoot    bytesHexStr `json:"parent_root"`
				StateRoot     bytesHexStr `json:"state_root"`
				BodyRoot      bytesHexStr `json:"body_root"`
			} `json:"message"`
			Signature bytesHexStr `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}

type StandardBeaconHeadersResponse struct {
	Data []struct {
		Root   bytesHexStr `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64      `json:"slot,string"`
				ProposerIndex uint64      `json:"proposer_index,string"`
				ParentRoot    bytesHexStr `json:"parent_root"`
				StateRoot     bytesHexStr `json:"state_root"`
				BodyRoot      bytesHexStr `json:"body_root"`
			} `json:"message"`
			Signature bytesHexStr `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}
