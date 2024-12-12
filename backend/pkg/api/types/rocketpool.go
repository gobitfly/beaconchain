package types

import (
	"time"

	"github.com/shopspring/decimal"
)

type RPNetworkStats struct {
	ClaimIntervalHours  float64         `db:"claim_interval_hours"`
	NodeOperatorRewards decimal.Decimal `db:"node_operator_rewards"`
	EffectiveRPLStaked  decimal.Decimal `db:"effective_rpl_staked"`
	RPLPrice            decimal.Decimal `db:"rpl_price"`
	Ts                  time.Time       `db:"ts"`
}

type RPInfo struct {
	Minipool             map[uint64]RPMinipoolInfo
	SmoothingPoolAddress []byte
}

type RPMinipoolInfo struct {
	NodeFee              float64
	NodeDepositBalance   decimal.Decimal
	UserDepositBalance   decimal.Decimal
	SmoothingPoolRewards map[uint64]decimal.Decimal
}
