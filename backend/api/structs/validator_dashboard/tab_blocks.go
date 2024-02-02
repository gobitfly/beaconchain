package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBBlocks struct {
	Paging Paging             `json:"paging,omitempty"`
	Blocks []VDBBlocksDetails `json:"blocks,omitempty"`
}

type VDBBlocksDetails struct {
	Proposer  uint64           `json:"proposer"`
	GroupName string           `json:"group_name"`
	GroupId   uint64           `json:"group_id"`
	Epoch     uint64           `json:"epoch"`
	Slot      uint64           `json:"slot"`
	Block     uint64           `json:"block,omitempty"`
	Age       time.Time        `json:"time"`
	Status    string           `json:"status"` // success, missed, orphaned, scheduled
	ElReward  *decimal.Decimal `json:"el_reward,omitempty"`
	ClReward  *decimal.Decimal `json:"cl_reward,omitempty"`
}
