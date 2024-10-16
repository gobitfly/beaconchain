package types

import "github.com/shopspring/decimal"

type MobileBundleData struct {
	BundleUrl                string `json:"bundle_url,omitempty"`
	HasNativeUpdateAvailable bool   `json:"has_native_update_available"`
}

type GetMobileLatestBundleResponse ApiDataResponse[MobileBundleData]

type MobileWidgetData struct {
	ValidatorStateCounts ValidatorStateCounts `json:"validator_state_counts"`
	Last24hIncome        decimal.Decimal      `json:"last_24h_income" faker:"eth"`
	Last7dIncome         decimal.Decimal      `json:"last_7d_income" faker:"eth"`
	Last30dApr           float64              `json:"last_30d_apr"`
	Last30dEfficiency    float64              `json:"last_30d_efficiency"`
	NetworkEfficiency    float64              `json:"network_efficiency"`
	RplPrice             decimal.Decimal      `json:"rpl_price" faker:"eth"`
	RplApr               float64              `json:"rpl_apr"`
}

type InternalGetValidatorDashboardMobileWidgetResponse ApiDataResponse[MobileWidgetData]

type MobileValidatorDashboardValidatorsRocketPool struct {
	DepositAmount     decimal.Decimal `json:"deposit_Amount"`
	Commision         float64         `json:"commision"` // percentage, 0-1
	Status            string          `json:"status"  tstype:"'staking' | 'dissolved' | 'prelaunch' | 'initialized' | 'withdrawable'" faker:"oneof: staking, dissolved, prelaunch, initialized, withdrawable"`
	PenaltyCount      uint64          `json:"penalty_count"`
	IsInSmoothingPool bool            `json:"is_in_smokaothing_pool"`
}
type MobileValidatorDashboardValidatorsTableRow struct {
	Index                uint64          `json:"index"`
	PublicKey            PubKey          `json:"public_key"`
	GroupId              uint64          `json:"group_id"`
	Balance              decimal.Decimal `json:"balance"`
	Status               string          `json:"status" tstype:"'slashed' | 'exited' | 'deposited' | 'pending' | 'slashing_offline' | 'slashing_online' | 'exiting_offline' | 'exiting_online' | 'active_offline' | 'active_online'" faker:"oneof: slashed, exited, deposited, pending, slashing_offline, slashing_online, exiting_offline, exiting_online, active_offline, active_online"`
	QueuePosition        *uint64         `json:"queue_position,omitempty"`
	WithdrawalCredential Hash            `json:"withdrawal_credential"`
	// additional mobile fields
	IsInSyncCommittee bool                                          `json:"is_in_sync_committee"`
	Efficiency        float64                                       `json:"efficiency"`
	RocketPool        *MobileValidatorDashboardValidatorsRocketPool `json:"rocket_pool,omitempty"`
}

type InternalGetValidatorDashboardMobileValidatorsResponse ApiPagingResponse[MobileValidatorDashboardValidatorsTableRow]
