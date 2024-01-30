package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBSummary struct {
	Paging          Paging            `json:"paging,omitempty"`
	TotalEfficiency VDBSummaryGroup   `json:"total_efficiency,omitempty"`
	Groups          []VDBSummaryGroup `json:"groups,omitempty"`
}

type VDBSummaryGroup struct {
	Id         uint64         `json:"id,omitempty"`
	Name       string         `json:"name"`
	Summary24h VDBSummaryInfo `json:"summary_24h"`
	Summary7d  VDBSummaryInfo `json:"summary_7d"`
	Summary31d VDBSummaryInfo `json:"summary_31d"`
	SummaryAll VDBSummaryInfo `json:"summary_all"`
	Validators []uint64       `json:"validators"`

	// chart
	DataPoints []VDBSummaryDataPoint `json:"data_points,omitempty"`
}

type VDBSummaryInfo struct {
	EfficiencyTotal float64 `json:"efficiency_total"`

	AttestationsHead       VDBSummaryItem `json:"att_head"`
	AttestationsSource     VDBSummaryItem `json:"att_source"`
	AttestationsTarget     VDBSummaryItem `json:"att_target"`
	AttestationEfficiency  float64        `json:"att_efficiency"`
	AttestationAvgInclDist float64        `json:"att_avg_incl_dist"`

	SyncCommittee VDBSummaryItem `json:"sync,omitempty"`
	Proposals     VDBSummaryItem `json:"proposals,omitempty"`

	AprElValue   *decimal.Decimal `json:"apr_el_value"`
	AprElPercent float64          `json:"apr_el_percent"`
	AprClValue   *decimal.Decimal `json:"apr_cl_value"`
	AprClPercent float64          `json:"apr_cl_percent"`

	LuckProposalPercent  float64        `json:"luck_proposal_percent"`
	LuckProposalExpected *time.Time     `json:"luck_proposal_expected"`
	LuckProposalAverage  *time.Duration `json:"luck_proposal_average"`
	LuckSyncPercent      float64        `json:"luck_sync_percent"`
	LuckSyncExpected     *time.Time     `json:"luck_sync_expected"`
	LuckSyncAverage      *time.Duration `json:"luck_sync_average"`
}

type VDBSummaryItem struct {
	Success    uint64           `json:"success"`
	Failed     uint64           `json:"failed"`
	Earned     *decimal.Decimal `json:"earned"`
	Penalty    *decimal.Decimal `json:"penalty"`
	Validators []uint64         `json:"validators,omitempty"`
}

type VDBSummaryDataPoint struct {
	Time       time.Time `json:"time"`
	Epoch      uint64    `json:"epoch"`
	Efficiency float64   `json:"efficiency"`
}
