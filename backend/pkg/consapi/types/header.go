package types

type StandardBeaconHeaderResponse struct {
	Data struct {
		Root   string `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64 `json:"slot,string"`
				ProposerIndex uint64 `json:"proposer_index,string"`
				ParentRoot    string `json:"parent_root"`
				StateRoot     string `json:"state_root"`
				BodyRoot      string `json:"body_root"`
			} `json:"message"`
			Signature bytesHexStr `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}

type StandardBeaconHeadersResponse struct {
	Data []struct {
		Root   string `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64 `json:"slot,string"`
				ProposerIndex uint64 `json:"proposer_index,string"`
				ParentRoot    string `json:"parent_root"`
				StateRoot     string `json:"state_root"`
				BodyRoot      string `json:"body_root"`
			} `json:"message"`
			Signature bytesHexStr `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}
