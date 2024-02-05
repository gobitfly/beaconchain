package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBRewardsTable struct {
	Paging common.Paging          `json:"paging,omitempty"`
	Epochs []VDBRewardsTableEpoch `json:"epochs,omitempty"`
}

type VDBRewardsTableEpoch struct {
	Number     uint64                 `json:"number"`
	Time       time.Time              `json:"time"`
	EpochTotal VDBRewardsTableGroup   `json:"epoch_total,omitempty"`
	Groups     []VDBRewardsTableGroup `json:"groups,omitempty"`
}

type VDBRewardsTableGroup struct {
	Name string `json:"name"`
	Id   uint64 `json:"id"`

	TotalReward *decimal.Decimal `json:"total_reward"`
	ElReward    *decimal.Decimal `json:"el_reward"`
	ClReward    *decimal.Decimal `json:"cl_reward"`
}

type VDBRewardsDetails struct {
	AttestationSource VDBRewardsDetailsDutyInfo `json:"attestation_source,omitempty"`
	AttestationTarget VDBRewardsDetailsDutyInfo `json:"attestation_target,omitempty"`
	AttestationHead   VDBRewardsDetailsDutyInfo `json:"attestation_head,omitempty"`
	Proposal          VDBRewardsDetailsDutyInfo `json:"proposal,omitempty"`
	Sync              VDBRewardsDetailsDutyInfo `json:"sync,omitempty"`
	Slashing          VDBRewardsDetailsDutyInfo `json:"slashing,omitempty"`
}

type VDBRewardsDetailsDutyInfo struct {
	Success     uint64           `json:"count"`
	Failed      uint64           `json:"failed"`
	ClReward    *decimal.Decimal `json:"cl_reward"`
	TotalReward *decimal.Decimal `json:"total_reward"`
	// proposals
	ElReward            *decimal.Decimal `json:"el_reward,omitempty"`
	TotalProposerReward *decimal.Decimal `json:"total_proposer_reward,omitempty"`
}

type VDBRewardsChart struct {
	Intervals []VDBRewardsChartInterval `json:"intervals,omitempty"`
}

type VDBRewardsChartInterval struct {
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	StartEpoch uint64    `json:"start_epoch"`
	EndEpoch   uint64    `json:"end_epoch"`

	TotalIncome VDBRewardsChartGroup   `json:"total_income"`
	Groups      []VDBRewardsChartGroup `json:"groups"`
}

type VDBRewardsChartGroup struct {
	Name        string           `json:"name,omitempty"`
	ElIncome    *decimal.Decimal `json:"el_income,omitempty"`
	ClIncome    *decimal.Decimal `json:"cl_income,omitempty"`
	TotalIncome *decimal.Decimal `json:"total_income,omitempty"`
}
