package types

import "github.com/shopspring/decimal"

type RPNetworkStats struct {
	ClaimIntervalHours  float64         `db:"claim_interval_hours"`
	NodeOperatorRewards decimal.Decimal `db:"node_operator_rewards"`
	EffectiveRPLStaked  decimal.Decimal `db:"effective_rpl_staked"`
	RPLPrice            decimal.Decimal `db:"rpl_price"`
}

type RPInfo struct {
	Node     map[string]RPNodeInfo
	Minipool map[uint64]RPMinipoolInfo
}
type RPNodeInfo struct {
	SmoothingPoolReward map[uint64]decimal.Decimal
}
type RPMinipoolInfo struct {
	NodeFee            float64
	NodeDepositBalance decimal.Decimal
	UserDepositBalance decimal.Decimal
}
