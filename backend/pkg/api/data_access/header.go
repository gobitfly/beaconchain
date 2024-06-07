package dataaccess

import (
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
)

func (d *DataAccessService) GetLatestSlot() (uint64, error) {
	latestSlot := cache.LatestSlot.Get()
	return latestSlot, nil
}

func (d *DataAccessService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	result := []t.EthConversionRate{}

	availableCurrencies := price.GetAvailableCurrencies()
	for _, code := range availableCurrencies {
		if code == "ETH" {
			// Don't return ETH/ETH info
			continue
		}
		rate := price.GetPrice("ETH", code)
		result = append(result, t.EthConversionRate{
			Currency: price.GetCurrencyLabel(code),
			Code:     code,
			Symbol:   price.GetCurrencySymbol(code),
			Rate:     rate,
		})
	}

	return result, nil
}
