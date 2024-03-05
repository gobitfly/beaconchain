package types

// ------------------------------------------------------------
// Slot Viz
type VDBSlotVizDuty struct {
	TotalCount uint64   `json:"total_count"`
	Validators []uint64 `json:"validators" faker:"slice_len=6"` // up to 6 validators that performed the duty, only for scheduled and failed
}

type VDBSlotVizTuple struct {
	Validator  uint64 `json:"validator"`
	DutyObject uint64 `json:"duty_object"`
}

type VDBSlotVizSlashing struct {
	TotalCount uint64            `json:"total_count"`
	Slashings  []VDBSlotVizTuple `json:"slashings" faker:"slice_len=6"` // up to 6 slashings, validator is always the slashing validator
}

type VDBSlotVizStatus[T any] struct {
	Success   *T `json:"success,omitempty"`
	Failed    *T `json:"failed,omitempty"`
	Scheduled *T `json:"scheduled,omitempty"`
}

type VDBSlotVizSlot struct {
	Slot         uint64                                `json:"slot"`
	Status       string                                `json:"status" tstype:"'proposed' | 'missed' | 'scheduled' | 'orphaned'" faker:"oneof: proposed, missed, scheduled, orphaned"`
	Proposal     *VDBSlotVizTuple                      `json:"proposal,omitempty"`
	Attestations *VDBSlotVizStatus[VDBSlotVizDuty]     `json:"attestations,omitempty"`
	Sync         *VDBSlotVizStatus[VDBSlotVizDuty]     `json:"sync,omitempty"`
	Slashing     *VDBSlotVizStatus[VDBSlotVizSlashing] `json:"slashing,omitempty"`
}
type SlotVizEpoch struct {
	Epoch    uint64           `json:"epoch"`
	State    string           `json:"state,omitempty" tstype:"'scheduled' | 'head' | 'justifying' | 'justified' | 'finalized'" faker:"oneof: scheduled, head, justifying, justified, finalized"` // only on landing page
	Progress float64          `json:"progress,omitempty"`                                                                                                                                        // only on landing page
	Slots    []VDBSlotVizSlot `json:"slots,omitempty" faker:"slice_len=32"`                                                                                                                      // only on dashboard page
}

type InternalGetValidatorDashboardSlotVizResponse ApiDataResponse[[]SlotVizEpoch]
