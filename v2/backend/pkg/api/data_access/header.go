package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/api/types"

func (d *DataAccessService) GetLatestSlot() (uint64, error) {
	// WORKING spletka
	return d.dummy.GetLatestSlot()
}

func (d *DataAccessService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	// WORKING spletka
	return d.dummy.GetLatestExchangeRates()
}
