package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBeaconHeaderResponse struct {
	Data struct {
		Root   hexutil.Bytes `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64        `json:"slot,string"`
				ProposerIndex uint64        `json:"proposer_index,string"`
				ParentRoot    hexutil.Bytes `json:"parent_root"`
				StateRoot     hexutil.Bytes `json:"state_root"`
				BodyRoot      hexutil.Bytes `json:"body_root"`
			} `json:"message"`
			Signature hexutil.Bytes `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}

type StandardBeaconHeadersResponse struct {
	Data []struct {
		Root   hexutil.Bytes `json:"root"`
		Header struct {
			Message struct {
				Slot          uint64        `json:"slot,string"`
				ProposerIndex uint64        `json:"proposer_index,string"`
				ParentRoot    hexutil.Bytes `json:"parent_root"`
				StateRoot     hexutil.Bytes `json:"state_root"`
				BodyRoot      hexutil.Bytes `json:"body_root"`
			} `json:"message"`
			Signature hexutil.Bytes `json:"signature"`
		} `json:"header"`
	} `json:"data"`
	Finalized bool `json:"finalized"`
}
