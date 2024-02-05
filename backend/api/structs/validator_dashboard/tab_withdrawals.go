package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBWithdrawals struct {
	Paging      common.Paging           `json:"paging,omitempty"`
	Total       *decimal.Decimal        `json:"total"`
	Withdrawals []VDBWithdrawalsDetails `json:"withdrawals"`
}

type VDBWithdrawalsDetails struct {
	Epoch     uint64           `json:"epoch"`
	Time      time.Time        `json:"time"`
	Index     uint64           `json:"index"`
	GroupName string           `json:"group_name,omitempty"`
	GroupId   uint64           `json:"group_id,omitempty"`
	Recipient common.Address   `json:"recipient"`
	Amount    *decimal.Decimal `json:"amount"`
}
