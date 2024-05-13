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
	CurrentSlot   uint64              `json:"current_slot"`
	ExchangeRates []EthConversionRate `json:"exchange_rates" faker:"len=3"`
}

type InternalGetLatestStateResponse ApiDataResponse[LatestStateData]
