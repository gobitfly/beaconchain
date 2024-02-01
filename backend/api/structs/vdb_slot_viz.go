package api

import (
	"github.com/shopspring/decimal"
)

type VDBSlotViz struct {
	Epochs []VDBSlotVizEpoch `json:"epochs"`
}

type VDBSlotVizEpoch struct {
	Number uint64           `json:"number"`
	Slots  []VDBSlotVizSlot `json:"slots"`
}

type VDBSlotVizSlot struct {
	Number uint64           `json:"number"`
	Status string           `json:"status"` // proposed, missed, scheduled, orphaned
	Duties []VDBSlotVizDuty `json:"duties"`
}

type VDBSlotVizDuty struct {
	Status    string           `json:"status"` // success, failed, pending
	Type      string           `json:"type"`   // proposal, attestation, sync, slashing
	Validator uint64           `json:"validator,omitempty"`
	Value     *decimal.Decimal `json:"value,omitempty"`
	Block     uint64           `json:"block,omitempty"`
}
