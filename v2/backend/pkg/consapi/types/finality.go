package types

// /eth/v1/beacon/states/{state_id}/finality_checkpoints
type StandardFinalityCheckpointsResponse struct {
	Data struct {
		PreviousJustified struct {
			Epoch uint64      `json:"epoch,string"`
			Root  bytesHexStr `json:"root"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch uint64      `json:"epoch,string"`
			Root  bytesHexStr `json:"root"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch uint64      `json:"epoch,string"`
			Root  bytesHexStr `json:"root"`
		} `json:"finalized"`
	} `json:"data"`
}
