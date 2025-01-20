package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type StandardBeaconHeaderResponse struct {
	Data      BeaconHeaderData `json:"data"`
	Finalized bool             `json:"finalized"`
}

type StandardBeaconHeadersResponse struct {
	Data      []BeaconHeaderData `json:"data"`
	Finalized bool               `json:"finalized"`
}

type BeaconHeaderData struct {
	Root      hexutil.Bytes       `json:"root"`
	Header    BeaconHeaderMessage `json:"header"`
	Canonical bool                `json:"canonical"`
}

type BeaconHeaderMessage struct {
	Message   BeaconHeaderMessageData `json:"message"`
	Signature hexutil.Bytes           `json:"signature"`
}
type BeaconHeaderMessageData struct {
	Slot          uint64        `json:"slot,string"`
	ProposerIndex uint64        `json:"proposer_index,string"`
	ParentRoot    hexutil.Bytes `json:"parent_root"`
	StateRoot     hexutil.Bytes `json:"state_root"`
	BodyRoot      hexutil.Bytes `json:"body_root"`
}
