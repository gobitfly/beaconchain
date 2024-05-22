package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/api/types"

func (d *DataAccessService) GetLatestSlot() (uint64, error) {
	// TODO @recy21
	return d.dummy.GetLatestSlot()
}

func (d *DataAccessService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	// TODO @recy21
	return d.dummy.GetLatestExchangeRates()
}
