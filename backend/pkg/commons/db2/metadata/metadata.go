package metadata

import (
	"fmt"
	"math/big"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
)

type Balance struct {
	metadataupdates.Pair
	Value *big.Int
}
type Store struct {
	db database.Database
}

func NewStore(db database.Database) Store {
	return Store{
		db: db,
	}
}

func (store Store) UpdateBalance(chainID string, balances []Balance) error {
	updates := make(map[string][]database.Item)
	for _, balance := range balances {
		key := fmt.Sprintf("%s:%x", chainID, balance.Address)
		updates[key] = append(updates[key], database.Item{
			Family: accountFamily,
			Column: fmt.Sprintf("B:%x", balance.Token),
			Data:   balance.Value.Bytes(),
		})
	}
	return store.db.BulkAdd(updates)
}
