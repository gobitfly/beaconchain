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
			DbName:   "geth",
			Category: "execution_layer",
		},
		{
			Id:       1,
			Name:     "Nethermind",
			DbName:   "nethermind",
			Category: "execution_layer",
		},
		{
			Id:       2,
			Name:     "Besu",
			DbName:   "besu",
			Category: "execution_layer",
		},
		{
			Id:       3,
			Name:     "Erigon",
			DbName:   "erigon",
			Category: "execution_layer",
		},
		{
			Id:       4,
			Name:     "Reth",
			DbName:   "reth",
			Category: "execution_layer",
		},
		// consensus_layer
		{
			Id:       5,
			Name:     "Teku",
			DbName:   "teku",
			Category: "consensus_layer",
		},
		{
			Id:       6,
			Name:     "Prysm",
			DbName:   "prysm",
			Category: "consensus_layer",
		},
		{
			Id:       7,
			Name:     "Nimbus",
			DbName:   "nimbus",
			Category: "consensus_layer",
		},
		{
			Id:       8,
			Name:     "Lighthouse",
			DbName:   "lighthouse",
			Category: "consensus_layer",
		},
		{
			Id:       9,
			Name:     "Lodestar",
			DbName:   "lodestar",
			Category: "consensus_layer",
		},
		// other
		{
			Id:       10,
			Name:     "Rocketpool Smart Node",
			DbName:   "rocketpool",
			Category: "other",
		},
		{
			Id:       11,
			Name:     "MEV-Boost",
			DbName:   "mev-boost",
			Category: "other",
		},
	}, nil
}
