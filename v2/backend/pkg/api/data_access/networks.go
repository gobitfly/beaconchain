package dataaccess

import "github.com/gobitfly/beaconchain/pkg/api/types"

type NetworkRepository interface {
	GetAllNetworks() ([]types.NetworkInfo, error)
}

func (d *DataAccessService) GetAllNetworks() ([]types.NetworkInfo, error) {
	// TODO @recy21
	// probably should load the networks into mem from some config when the service is created

	return []types.NetworkInfo{
		{
			ChainId: 1,
			Name:    "ethereum",
		},
		{
			ChainId: 100,
			Name:    "gnosis",
		},
	}, nil
}
