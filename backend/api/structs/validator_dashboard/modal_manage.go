package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBManageTable struct {
	Default VDBManageTableGroup   `json:"default"`
	Groups  []VDBManageTableGroup `json:"groups,omitempty"`
}

type VDBManageTableGroup struct {
	Name  string `json:"name,omitempty"`
	Id    uint64 `json:"id,omitempty"`
	Count uint64 `json:"count"`
}

type VDBManageDetails struct {
	Paging     common.Paging          `json:"paging,omitempty"`
	Validators []VDBManageDetailsItem `json:"validators"`
}

type VDBManageDetailsItem struct {
	Index                 uint64           `json:"index,omitempty"`
	Pubkey                common.PubKey    `json:"pubkey"`
	Balance               *decimal.Decimal `json:"balance"`
	Status                string           `json:"status"` // active, deposited, pending, inactive
	WithdrawalCredentials common.Hash      `json:"withdrawal_credentials"`
	// pending
	QueuePosition  uint64    `json:"queue_position,omitempty"`
	ActivationTime time.Time `json:"activation_time,omitempty"`
}
