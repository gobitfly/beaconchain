package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBSummaryTable struct {
	Paging          common.Paging          `json:"paging,omitempty"`
	TotalEfficiency VDBSummaryTableGroup   `json:"total_efficiency,omitempty"`
	Groups          []VDBSummaryTableGroup `json:"groups,omitempty"`
}

type VDBSummaryTableGroup struct {
	Id   uint64 `json:"id,omitempty"`
	Name string `json:"name"`

	Efficiency24h float64 `json:"efficiency_24h"`
	Efficiency7d  float64 `json:"efficiency_7d"`
	Efficiency31d float64 `json:"efficiency_31d"`
	EfficiencyAll float64 `json:"efficiency_all"`

	Validators []uint64 `json:"validators"`
}

type VDBSummaryDetails struct {
	AttestationsHead       VDBSummaryDetailsItem `json:"att_head"`
	AttestationsSource     VDBSummaryDetailsItem `json:"att_source"`
	AttestationsTarget     VDBSummaryDetailsItem `json:"att_target"`
	AttestationEfficiency  float64               `json:"att_efficiency"`
	AttestationAvgInclDist float64               `json:"att_avg_incl_dist"`

	SyncCommittee VDBSummaryDetailsItem `json:"sync,omitempty"`
	Proposals     VDBSummaryDetailsItem `json:"proposals,omitempty"`

	ElApr VDBSummaryDetailsApr `json:"el_apr"`
	ClApr VDBSummaryDetailsApr `json:"cl_apr"`

	ProposalLuck VDBSummaryDetailsLuck `json:"proposal_luck"`
	SyncLuck     VDBSummaryDetailsLuck `json:"sync_luck"`
}

type VDBSummaryDetailsItem struct {
	Success    uint64           `json:"success"`
	Failed     uint64           `json:"failed"`
	Earned     *decimal.Decimal `json:"earned"`
	Penalty    *decimal.Decimal `json:"penalty"`
	Validators []uint64         `json:"validators,omitempty"`
}

type VDBSummaryDetailsApr struct {
	Value   *decimal.Decimal `json:"value"`
	Percent float64          `json:"percent"`
}

type VDBSummaryDetailsLuck struct {
	Percent  float64        `json:"percent"`
	Expected *time.Time     `json:"expected"`
	Average  *time.Duration `json:"average"`
}

type VDBSummaryChart struct {
	Intervals []VDBSummaryChartInterval `json:"intervals,omitempty"`
}

type VDBSummaryChartInterval struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	StartEpoch uint64    `json:"start_epoch"`
	EndEpoch   uint64    `json:"end_epoch"`

	TotalEfficiency VDBSummaryChartGroup   `json:"total_efficiency,omitempty"`
	Groups          []VDBSummaryChartGroup `json:"groups,omitempty"`
}

type VDBSummaryChartGroup struct {
	Name       string  `json:"name,omitempty"`
	Efficiency float64 `json:"efficiency"`
}
