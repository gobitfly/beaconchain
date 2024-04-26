package types

import "github.com/ethereum/go-ethereum/common/hexutil"

// /eth/v1/beacon/states/{state_id}/finality_checkpoints
type StandardFinalityCheckpointsResponse struct {
	Data struct {
		PreviousJustified struct {
			Epoch uint64        `json:"epoch,string"`
			Root  hexutil.Bytes `json:"root"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch uint64        `json:"epoch,string"`
			Root  hexutil.Bytes `json:"root"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch uint64        `json:"epoch,string"`
			Root  hexutil.Bytes `json:"root"`
		} `json:"finalized"`
	} `json:"data"`
}
