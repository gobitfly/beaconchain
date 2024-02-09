package api

import (
	"github.com/shopspring/decimal"
)

type VDBSlotViz struct {
	CurrentSlot uint64            `json:"current_slot"`
	Epochs      []VDBSlotVizEpoch `json:"epochs"`
}

type VDBSlotVizEpoch struct {
	Number uint64           `json:"number"`
	State  string           `json:"state"` // head, finalized, scheduled
	Slots  []VDBSlotVizSlot `json:"slots"`
}

type VDBSlotVizSlot struct {
	Number uint64           `json:"number"`
	Status string           `json:"status"` // proposed, missed, scheduled, orphaned
	Duties []VDBSlotVizDuty `json:"duties"`
}

type VDBSlotVizDuty struct {
	Type            string          `json:"type"` // proposal, attestation, sync, slashing
	PendingCount    uint64          `json:"pending_count"`
	SuccessCount    uint64          `json:"success_count"`
	SuccessEarnings decimal.Decimal `json:"success_earnings"`
	FailedCount     uint64          `json:"failed_count"`
	FailedEarnings  decimal.Decimal `json:"failed_earnings"`
	Validator       uint64          `json:"validator"`
	Block           uint64          `json:"block"`
}
