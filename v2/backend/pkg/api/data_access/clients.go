package dataaccess

import "github.com/gobitfly/beaconchain/pkg/api/types"

type ClientRepository interface {
	GetAllClients() ([]types.ClientInfo, error)
}

func (d *DataAccessService) GetAllClients() ([]types.ClientInfo, error) {
	// TODO @recy21
	// probably should load the clients into mem from some config when the service is created

	return []types.ClientInfo{
		// execution_layer
		{
			Id:       0,
			Name:     "Geth",
			Category: "execution_layer",
		},
		{
			Id:       1,
			Name:     "Nethermind",
			Category: "execution_layer",
		},
		{
			Id:       2,
			Name:     "Besu",
			Category: "execution_layer",
		},
		{
			Id:       3,
			Name:     "Erigon",
			Category: "execution_layer",
		},
		{
			Id:       4,
			Name:     "Reth",
			Category: "execution_layer",
		},
		// consensus_layer
		{
			Id:       5,
			Name:     "Teku",
			Category: "consensus_layer",
		},
		{
			Id:       6,
			Name:     "Prysm",
			Category: "consensus_layer",
		},
		{
			Id:       7,
			Name:     "Nimbus",
			Category: "consensus_layer",
		},
		{
			Id:       8,
			Name:     "Lighthouse",
			Category: "consensus_layer",
		},
		{
			Id:       9,
			Name:     "Lodestar",
			Category: "consensus_layer",
		},
		// other
		{
			Id:       10,
			Name:     "Rocketpool Smart Node",
			Category: "other",
		},
		{
			Id:       11,
			Name:     "MEV-Boost",
			Category: "other",
		},
	}, nil
}
