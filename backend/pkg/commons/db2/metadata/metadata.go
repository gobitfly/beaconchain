package metadata

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
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

func (store Store) UpdateToken(chainID string, tokens []*types.ERC20TokenPrice) error {
	updates := make(map[string][]database.Item)
	for _, token := range tokens {
		key := fmt.Sprintf("%s:%x", chainID, token.Token)
		updates[key] = append(updates[key], database.Item{
			Family: erc20MetadataFamily,
			Column: erc20ColumnPrice,
			Data:   token.Price,
		})
		updates[key] = append(updates[key], database.Item{
			Family: erc20MetadataFamily,
			Column: erc20ColumnTotalSupply,
			Data:   token.TotalSupply,
		})
	}
	return store.db.BulkAdd(updates)
}

func (store Store) TokenPrice(chainID string, token common.Address) (*types.ERC20TokenPrice, error) {
	key := fmt.Sprintf("%s:%x", chainID, token.Bytes())
	row, err := store.db.GetRow(key)
	if err != nil {
		return nil, err
	}
	return &types.ERC20TokenPrice{
		Token:       token.Bytes(),
		Price:       row.Values[fmt.Sprintf("%s:%s", erc20MetadataFamily, erc20ColumnPrice)],
		TotalSupply: row.Values[fmt.Sprintf("%s:%s", erc20MetadataFamily, erc20ColumnTotalSupply)],
	}, nil
}
