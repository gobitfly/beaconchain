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
			ChainId:           1,
			Name:              "ethereum",
			NotificationsName: "mainnet",
		},
		{
			ChainId:           100,
			Name:              "gnosis",
			NotificationsName: "gnosis",
		},
		{
			ChainId:           17000,
			Name:              "holesky",
			NotificationsName: "holesky",
		},
		{
			ChainId:           560048,
			Name:              "hoodi",
			NotificationsName: "hoodi",
		},
		{
			ChainId:           11155111,
			Name:              "sepolia",
			NotificationsName: "sepolia",
		},
		{
			ChainId:           7088110746,
			Name:              "pectra-devnet-5",
			NotificationsName: "pectra-devnet-5",
		},
		{
			ChainId:           7072151312,
			Name:              "pectra-devnet-6",
			NotificationsName: "pectra-devnet-6",
		},
		{
			ChainId:           3151908,
			Name:              "local-devnet",
			NotificationsName: "local-devnet",
		},
	}, nil
}
