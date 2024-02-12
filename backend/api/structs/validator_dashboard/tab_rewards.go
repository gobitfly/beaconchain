package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBRewardsTable struct {
	Epochs []VDBRewardsTableEpoch `json:"epochs"`
}

type VDBRewardsTableEpoch struct {
	Number uint64               `json:"number"`
	Time   time.Time            `json:"time"`
	Groups []VDBRewardsTableRow `json:"groups"`
}

type VDBRewardsTableRow struct {
	GroupName string `json:"group_name"`
	GroupId   uint64 `json:"group_id"`

	TotalReward decimal.Decimal `json:"total_reward"`
	ElReward    decimal.Decimal `json:"el_reward"`
	ClReward    decimal.Decimal `json:"cl_reward"`
}

type VDBRewardsDetails struct {
	AttestationSource VDBRewardsDetailsDutyInfo `json:"attestation_source"`
	AttestationTarget VDBRewardsDetailsDutyInfo `json:"attestation_target"`
	AttestationHead   VDBRewardsDetailsDutyInfo `json:"attestation_head"`
	Proposal          VDBRewardsDetailsDutyInfo `json:"proposal"`
	ProposalElReward  decimal.Decimal           `json:"proposal_el_reward"`
	Sync              VDBRewardsDetailsDutyInfo `json:"sync"`
	Slashing          VDBRewardsDetailsDutyInfo `json:"slashing"`
}

type VDBRewardsDetailsDutyInfo struct {
	Success     uint64          `json:"count"`
	Failed      uint64          `json:"failed"`
	ClReward    decimal.Decimal `json:"cl_reward"`
	TotalReward decimal.Decimal `json:"total_reward"`
}
