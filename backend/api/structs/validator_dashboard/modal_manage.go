package api

import (
	"time"

	common "github.com/gobitfly/beaconchain/api/structs"
	"github.com/shopspring/decimal"
)

type VDBManageTable struct {
	Groups []VDBManageTableGroup `json:"groups"`
}

type VDBManageTableGroup struct {
	Validators []VDBManageTableValidator `json:"validators"`
}

type VDBManageTableValidator struct {
	Index                 uint64          `json:"index"`
	Pubkey                common.PubKey   `json:"pubkey"`
	GroupName             string          `json:"group_name,omitempty"`
	GroupId               uint64          `json:"group_id,omitempty"`
	Balance               decimal.Decimal `json:"balance"`
	Status                string          `json:"status"` // active, deposited, pending, inactive
	WithdrawalCredentials common.Hash     `json:"withdrawal_credentials"`
	// pending
	QueuePosition  uint64    `json:"queue_position"`
	ActivationTime time.Time `json:"activation_time"`
}
