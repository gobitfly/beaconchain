package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type VDBManage struct {
	Paging Paging           `json:"paging,omitempty"`
	Groups []VDBManageGroup `json:"groups"`
}

type VDBManageGroup struct {
	Name       string               `json:"name"`
	Id         uint64               `json:"id"`
	Validators []VDBManageValidator `json:"validators"`
}

type VDBManageValidator struct {
	Index                 uint64           `json:"index,omitempty"`
	Pubkey                PubKey           `json:"pubkey"`
	Balance               *decimal.Decimal `json:"balance"`
	Status                string           `json:"status"` // active, deposited, pending, inactive
	WithdrawalCredentials Hash             `json:"withdrawal_credentials"`
	// pending
	QueuePosition  uint64    `json:"queue_position,omitempty"`
	ActivationTime time.Time `json:"activation_time,omitempty"`
}
