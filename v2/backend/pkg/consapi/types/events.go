package types

import "github.com/gobitfly/beaconchain/pkg/consapi/utils"

type EventTopic string

const (
	EventHead  EventTopic = "head"
	EventBlock EventTopic = "block"
	// EventAttestation                 EventTopic = "attestation"
	// EventVoluntaryExit               EventTopic = "voluntary_exit"
	// EventBlsToExecutionChange        EventTopic = "bls_to_execution_change"
	EventFinalizedCheckpoint EventTopic = "finalized_checkpoint"
	EventChainReorg          EventTopic = "chain_reorg"
	// EventContributionAndProof        EventTopic = "contribution_and_proof"
	// EventLightClientFinalityUpdate   EventTopic = "light_client_finality_update"
	// EventLightClientOptimisticUpdate EventTopic = "light_client_optimistic_update"
	// EventPayloadAttributes           EventTopic = "payload_attributes"
)

type EventResponse struct {
	Event EventTopic
	Data  []byte
	Error error
}

// Helper to get Head response type, returns nil if it is not a head event
func (e EventResponse) Head() (*StandardEventHeadResponse, error) {
	if e.Event != EventHead {
		return nil, nil
	}
	return utils.Unmarshal[StandardEventHeadResponse](e.Data, e.Error)
}

// Helper to get Block response type, returns nil if it is not a block event
func (e EventResponse) Block() (*StandardEventBlockResponse, error) {
	if e.Event != EventBlock {
		return nil, nil
	}
	return utils.Unmarshal[StandardEventBlockResponse](e.Data, e.Error)
}

// Helper to get ChainReorg response type, returns nil if it is not a chain reorg event
func (e EventResponse) ChainReorg() (*StandardEventChainReorg, error) {
	if e.Event != EventChainReorg {
		return nil, nil
	}
	return utils.Unmarshal[StandardEventChainReorg](e.Data, e.Error)
}

// Helper to get FinalizedCheckpoint response type, returns nil if it is not a finalized checkpoint event
func (e EventResponse) FinalizedCheckpoint() (*StandardFinalizedCheckpointResponse, error) {
	if e.Event != EventFinalizedCheckpoint {
		return nil, nil
	}
	return utils.Unmarshal[StandardFinalizedCheckpointResponse](e.Data, e.Error)
}

type StandardEventHeadResponse struct {
	Slot                      uint64      `json:"slot,string"`
	Block                     string      `json:"block"`
	State                     bytesHexStr `json:"state"`
	EpochTransition           bool        `json:"epoch_transition"`
	PreviousDutyDependentRoot bytesHexStr `json:"previous_duty_dependent_root"`
	CurrentDutyDependentRoot  bytesHexStr `json:"current_duty_dependent_root"`
	ExecutionOptimistic       bool        `json:"execution_optimistic"`
}

type StandardEventBlockResponse struct {
	Slot                uint64      `json:"slot,string"`
	Block               bytesHexStr `json:"block"`
	ExecutionOptimistic bool        `json:"execution_optimistic"`
}

type StandardEventChainReorg struct {
	Slot                uint64      `json:"slot,string"`
	Depth               uint64      `json:"depth,string"`
	OldHeadBlock        bytesHexStr `json:"old_head_block"`
	NewHeadBlock        bytesHexStr `json:"new_head_block"`
	OldHeadState        bytesHexStr `json:"old_head_state"`
	NewHeadState        bytesHexStr `json:"new_head_state"`
	Epoch               uint64      `json:"epoch,string"`
	ExecutionOptimistic bool        `json:"execution_optimistic"`
}

type StandardFinalizedCheckpointResponse struct {
	Block               bytesHexStr `json:"block"`
	State               bytesHexStr `json:"state"`
	Epoch               uint64      `json:"epoch,string"`
	ExecutionOptimistic bool        `json:"execution_optimistic"`
}
