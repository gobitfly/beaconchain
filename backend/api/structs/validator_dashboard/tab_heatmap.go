package api

import (
	"github.com/shopspring/decimal"
)

type VDBHeatmapDetails struct {
	Epoch uint64 `json:"epoch"`

	Proposers []VDBHeatmapExtraValidatorDuty `json:"proposers"`
	Syncs     []VDBHeatmapExtraValidatorDuty `json:"syncs"`
	Slashings []VDBHeatmapExtraValidatorDuty `json:"slashings"`

	AttHead   VDBHeatmapAttestations `json:"att_head"`
	AttSource VDBHeatmapAttestations `json:"att_source"`
	AttTarget VDBHeatmapAttestations `json:"att_target"`
	AttIncome decimal.Decimal        `json:"att_income"`
}

type VDBHeatmapAttestations struct {
	Success uint64 `json:"success"`
	Failed  uint64 `json:"failed"`
}

type VDBHeatmapExtraValidatorDuty struct {
	Index  uint64 `json:"index"`
	Status string `json:"status"` // success, failed, orphaned
}
