package api

import (
	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBEpochDuties struct {
	Validators []VDBEpochDutiesValidator `json:"validators"`
}

type VDBEpochDutiesValidator struct {
	Index uint64 `json:"index"`

	AttestationSource VDBEpochDutiesItem `json:"attestation_source"`
	AttestationTarget VDBEpochDutiesItem `json:"attestation_target"`
	AttestationHead   VDBEpochDutiesItem `json:"attestation_head"`
	Proposal          VDBEpochDutiesItem `json:"proposal"`
	Sync              VDBEpochDutiesItem `json:"sync"`
	Slashing          VDBEpochDutiesItem `json:"slashing"`

	TotalReward decimal.Decimal `json:"total_reward"`
}

type VDBEpochDutiesItem struct {
	Status string          `json:"status"` // success, mixed, failed, orphaned
	Reward decimal.Decimal `json:"reward"`
}
