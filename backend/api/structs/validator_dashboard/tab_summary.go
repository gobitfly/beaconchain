package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBSummaryTable struct {
	TotalEfficiency VDBSummaryGroup   `json:"total_efficiency"`
	Groups          []VDBSummaryGroup `json:"groups"`
}

type VDBSummaryGroup struct {
	GroupName string `json:"group_name"`
	GroupId   uint64 `json:"group_id"`

	Efficiency24h float64 `json:"efficiency_24h"`
	Efficiency7d  float64 `json:"efficiency_7d"`
	Efficiency31d float64 `json:"efficiency_31d"`
	EfficiencyAll float64 `json:"efficiency_all"`

	Validators []uint64 `json:"validators"`
}

type VDBSummaryDetails struct {
	Details24h VDBSummaryDetailsColumn `json:"details_24h"`
	Details7d  VDBSummaryDetailsColumn `json:"details_7d"`
	Details31d VDBSummaryDetailsColumn `json:"details_31d"`
	DetailsAll VDBSummaryDetailsColumn `json:"details_all"`
}

type VDBSummaryDetailsColumn struct {
	AttestationsHead       VDBSummaryDetailsItem `json:"att_head"`
	AttestationsSource     VDBSummaryDetailsItem `json:"att_source"`
	AttestationsTarget     VDBSummaryDetailsItem `json:"att_target"`
	AttestationEfficiency  float64               `json:"att_efficiency"`
	AttestationAvgInclDist float64               `json:"att_avg_incl_dist"`

	SyncCommittee VDBSummaryDetailsItem `json:"sync"`
	Proposals     VDBSummaryDetailsItem `json:"proposals"`

	ElApr VDBSummaryDetailsApr `json:"el_apr"`
	ClApr VDBSummaryDetailsApr `json:"cl_apr"`

	ProposalLuck VDBSummaryDetailsLuck `json:"proposal_luck"`
	SyncLuck     VDBSummaryDetailsLuck `json:"sync_luck"`
}

type VDBSummaryDetailsItem struct {
	Success    uint64          `json:"success"`
	Failed     uint64          `json:"failed"`
	Earned     decimal.Decimal `json:"earned"`
	Penalty    decimal.Decimal `json:"penalty"`
	Validators []uint64        `json:"validators,omitempty"`
}

type VDBSummaryDetailsApr struct {
	Value   decimal.Decimal `json:"value"`
	Percent float64         `json:"percent"`
}

type VDBSummaryDetailsLuck struct {
	Percent  float64       `json:"percent"`
	Expected time.Time     `json:"expected"`
	Average  time.Duration `json:"average"`
}
