package types

// ------------------------------
// various types that are used for frontend configs

// EthConversionRate is the exchange rate of ETH to a specific currency
type EthConversionRate struct {
	Currency string  `json:"currency" faker:"oneof:United States Dollar,British Pound,Euro"`
	Code     string  `json:"code" faker:"oneof:USD,GBP,EUR"`
	Symbol   string  `json:"symbol" faker:"oneof:£,$,€"`
	Rate     float64 `json:"rate" faker:"amount"`
}

type LatestStateData struct {
	LatestSlot    uint64              `json:"current_slot"`
	ExchangeRates []EthConversionRate `json:"exchange_rates" faker:"slice_len=3"`
}

type InternalGetLatestStateResponse ApiDataResponse[LatestStateData]

type RocketPoolData struct {
	LatestUpdateSlot uint64 `json:"latest_update_slot"`
	EthRates         struct {
		Rpl  float64 `json:"rpl"`
		Reth float64 `json:"reth"`
	} `json:"eth_rates"`
}

type InternalGetRocketPoolResponse ApiDataResponse[RocketPoolData]
