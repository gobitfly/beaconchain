package dataaccess

import "github.com/gobitfly/beaconchain/pkg/api/types"

type ClientRepository interface {
	GetAllClients() ([]types.ClientInfo, error)
}

func (d *DataAccessService) GetAllClients() ([]types.ClientInfo, error) {
	// TODO @recy21
	// probably should load the clients into mem from some config when the service is created

	return []types.ClientInfo{
		// Execution Clients
		{
			Id:       0,
			Name:     "Geth",
			Category: "Execution Clients",
		},
		{
			Id:       1,
			Name:     "Nethermind",
			Category: "Execution Clients",
		},
		{
			Id:       2,
			Name:     "Besu",
			Category: "Execution Clients",
		},
		{
			Id:       3,
			Name:     "Erigon",
			Category: "Execution Clients",
		},
		{
			Id:       4,
			Name:     "Reth",
			Category: "Execution Clients",
		},
		// Consensus Clients
		{
			Id:       5,
			Name:     "Teku",
			Category: "Consensus Clients",
		},
		{
			Id:       6,
			Name:     "Prysm",
			Category: "Consensus Clients",
		},
		{
			Id:       7,
			Name:     "Nimbus",
			Category: "Consensus Clients",
		},
		{
			Id:       8,
			Name:     "Lighthouse",
			Category: "Consensus Clients",
		},
		{
			Id:       9,
			Name:     "Lodestar",
			Category: "Consensus Clients",
		},
		// Other
		{
			Id:       10,
			Name:     "Rocketpool Smart Node",
			Category: "Other",
		},
		{
			Id:       11,
			Name:     "MEV-Boost",
			Category: "Other",
		},
	}, nil
}
