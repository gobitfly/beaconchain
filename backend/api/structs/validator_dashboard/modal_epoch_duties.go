package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBEpochDuties struct {
	Paging     Paging                    `json:"paging,omitempty"`
	Validators []VDBEpochDutiesValidator `json:"validators,omitempty"`
}

type VDBEpochDutiesValidator struct {
	Index       uint64    `json:"index"`
	Latest_Time time.Time `json:"latest_time"`

	AttestationSource VDBEpochDutiesItem `json:"attestation_source,omitempty"`
	AttestationTarget VDBEpochDutiesItem `json:"attestation_target,omitempty"`
	AttestationHead   VDBEpochDutiesItem `json:"attestation_head,omitempty"`
	Proposal          VDBEpochDutiesItem `json:"proposal,omitempty"`
	Sync              VDBEpochDutiesItem `json:"sync,omitempty"`
	Slashing          VDBEpochDutiesItem `json:"slashing,omitempty"`

	TotalValue *decimal.Decimal `json:"total_value,omitempty"`
}

type VDBEpochDutiesItem struct {
	Status string           `json:"status"` // success, mixed, failed, orphaned
	Value  *decimal.Decimal `json:"value,omitempty"`
}
