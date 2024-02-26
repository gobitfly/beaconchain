package types

// ------------------------------------------------------------
// Slot Viz
type VDBSlotVizPassiveDuty struct {
	PendingCount uint64 `json:"pending_count"`
	SuccessCount uint64 `json:"success_count"`
	FailedCount  uint64 `json:"failed_count"`
}

type VDBSlotVizActiveDuty struct {
	Status    string `json:"status" tstype:"'success' | 'failed' | 'scheduled'"`
	Validator uint64 `json:"validator"`
	/*
		If the duty is a proposal & it's successful, the duty_object is the proposed block
		If the duty is a proposal & it failed/scheduled, the duty_object is the slot
		If the duty is a slashing & it's successful, the duty_object is the validator you slashed
		If the duty is a slashing & it failed, the duty_object is your validator that was slashed
	*/
	DutyObject uint64 `json:"duty_object"`
}

type VDBSlotVizSlot struct {
	Slot         uint64                 `json:"slot"`
	Status       string                 `json:"status" tstype:"'proposed' | 'missed' | 'scheduled' | 'orphaned'"`
	Attestations VDBSlotVizPassiveDuty  `json:"attestations,omitempty"`
	Sync         VDBSlotVizPassiveDuty  `json:"sync,omitempty"`
	Proposals    []VDBSlotVizActiveDuty `json:"proposals,omitempty"`
	Slashing     []VDBSlotVizActiveDuty `json:"slashing,omitempty"`
}
type SlotVizEpoch struct {
	Epoch    uint64           `json:"epoch"`
	State    string           `json:"state,omitempty" tstype:"'head' | 'finalized' | 'scheduled'"` // only on landing page
	Progress float64          `json:"progress,omitempty"`                                          // only on landing page
	Slots    []VDBSlotVizSlot `json:"slots,omitempty"`                                             // only on dashboard page
}

type InternalGetValidatorDashboardSlotVizResponse ApiDataResponse[[]SlotVizEpoch]
