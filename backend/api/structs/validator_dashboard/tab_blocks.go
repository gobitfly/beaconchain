package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBBlocks struct {
	Blocks []VDBBlocksDetails `json:"blocks"`
}

type VDBBlocksDetails struct {
	Proposer  uint64          `json:"proposer"`
	GroupName string          `json:"group_name"`
	GroupId   uint64          `json:"group_id"`
	Epoch     uint64          `json:"epoch"`
	Slot      uint64          `json:"slot"`
	Block     uint64          `json:"block"`
	Age       time.Time       `json:"time"`
	Status    string          `json:"status"` // success, missed, orphaned, scheduled
	ElReward  decimal.Decimal `json:"el_reward"`
	ClReward  decimal.Decimal `json:"cl_reward"`
}
