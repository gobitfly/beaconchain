package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBHeatmap struct {
	Epochs []VDBBlocksEpoch `json:"epochs,omitempty"`
}

type VDBBlocksEpoch struct {
	Number uint64               `json:"number"`
	Time   time.Time            `json:"time"`
	Groups []VDBHeatmapOverview `json:"groups,omitempty"`
}

type VDBHeatmapOverview struct {
	Name     string           `json:"name"`
	Value    *decimal.Decimal `json:"value"`
	Proposal bool             `json:"proposal,omitempty"`
	Sync     bool             `json:"sync,omitempty"`
	Slashing bool             `json:"slashing,omitempty"`
}

type VDBHeatmapDetails struct {
	Proposers []VDBHeatmapExtraValidatorDuty `json:"proposers,omitempty"`
	Syncs     []VDBHeatmapExtraValidatorDuty `json:"syncs,omitempty"`
	Slashings []VDBHeatmapExtraValidatorDuty `json:"slashings,omitempty"`

	AttHead   VDBHeatmapAttestations `json:"att_head"`
	AttSource VDBHeatmapAttestations `json:"att_source"`
	AttTarget VDBHeatmapAttestations `json:"att_target"`
	AttValue  *decimal.Decimal       `json:"att_value"`
}

type VDBHeatmapAttestations struct {
	Success uint64 `json:"success"`
	Failed  uint64 `json:"failed"`
}

type VDBHeatmapExtraValidatorDuty struct {
	Index  uint64 `json:"index"`
	Status string `json:"status"` // success, failed, orphaned
}
