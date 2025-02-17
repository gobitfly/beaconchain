package executionlayer

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func TestIndexerWithBigTable(t *testing.T) {
	btClient, btAdmin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), btClient, btAdmin, db2.Schema)
	if err != nil {
		t.Fatal(err)
	}

	backend := th.NewBackend(t)
	client, err := rpc.NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		transformers []TransformFunc
		action       func(*testing.T) error
		dataKeys     []string
		updatesKeys  []string
		metadataKeys []string
	}{
		{
			name: "transaction",
			transformers: []TransformFunc{
				TransformTx,
			},
			action: func(t *testing.T) error {
				temp := th.CreateEOA(t)
				return backend.Client().SendTransaction(context.Background(), backend.MakeTx(t, backend.BankAccount, &temp.From, big.NewInt(1), nil))
			},
			dataKeys: []string{
				":I:TX:",
			},
			updatesKeys: []string{
				":B:",
			},
		},
		{
			name: "block",
			transformers: []TransformFunc{
				TransformBlock,
			},
			action: func(t *testing.T) error {
				return nil
			},
			dataKeys: []string{
				":I:B:",
			},
			updatesKeys: []string{
				":BLOCK:",
			},
		},
		{
			name: "erc20 transfer",
			transformers: []TransformFunc{
				TransformERC20,
			},
			action: func(t *testing.T) error {
				temp := th.CreateEOA(t)
				_, usdt := backend.DeployERC20(t, "usdt", "usdt", backend.BankAccount.From)
				_, err := usdt.Mint(backend.BankAccount.TransactOpts, temp.From, big.NewInt(1))
				return err
			},
			dataKeys: []string{
				":I:ERC20:",
			},
			updatesKeys: []string{
				":B:",
			},
		},
		{
			name: "erc721 transfer",
			transformers: []TransformFunc{
				TransformERC721,
			},
			action: func(t *testing.T) error {
				temp := th.CreateEOA(t)
				_, token := backend.DeployERC721(t, "name", "symbol")
				_, err := token.Mint(backend.BankAccount.TransactOpts, temp.From, big.NewInt(1))
				return err
			},
			dataKeys: []string{
				":I:ERC721:",
			},
		},
		{
			name: "erc1155 transfer",
			transformers: []TransformFunc{
				TransformERC1155,
			},
			action: func(t *testing.T) error {
				temp := th.CreateEOA(t)
				_, token := backend.DeployToken1155(t)
				_, err := token.Mint(backend.BankAccount.TransactOpts, temp.From, big.NewInt(1), big.NewInt(1), nil)
				return err
			},
			dataKeys: []string{
				":I:ERC1155:",
			},
		},
		{
			name: "contract",
			transformers: []TransformFunc{
				TransformContract,
			},
			action: func(t *testing.T) error {
				_, _, _, err := contracts.DeployERC20(backend.BankAccount.TransactOpts, backend.Client(), "name", "symbol")
				return err
			},
			metadataKeys: []string{":S:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() { _ = bt.Clear() }()

			indexer := NewIndexer(
				db2.NewStoreV1(
					database.Wrap(bt, db2.DataTable),
					database.Wrap(bt, db2.MetadataTable),
					database.Wrap(bt, db2.UpdatesTable),
					db2.NoopCache{},
				),
				tt.transformers...,
			)
			if err := tt.action(t); err != nil {
				t.Fatal(err)
			}
			backend.Commit()
			lastBlock, err := backend.Client().BlockNumber(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			block, _, err := client.GetBlock(int64(lastBlock), "geth")
			if err != nil {
				t.Fatal(err)
			}
			if err := indexer.IndexBlocks(fmt.Sprintf("%d", backend.ChainID), []*types.Eth1Block{block}); err != nil {
				t.Fatal(err)
			}

			dataRows, err := bt.Read(db2.DataTable, "")
			if err != nil {
				t.Fatal(err)
			}
			if err := rowsContains(dataRows, tt.dataKeys); err != nil {
				t.Error(err)
			}

			updatesRows, err := bt.Read(db2.UpdatesTable, "")
			if err != nil {
				t.Fatal(err)
			}
			if err := rowsContains(updatesRows, tt.updatesKeys); err != nil {
				t.Error(err)
			}

			metadataRows, err := bt.Read(db2.MetadataTable, "")
			if err != nil {
				t.Fatal(err)
			}
			if err := rowsContains(metadataRows, tt.metadataKeys); err != nil {
				t.Error(err)
			}
		})
	}
}

func rowsContains(rows []database.Row, keys []string) error {
	for _, key := range keys {
		found := false
		for _, row := range rows {
			if strings.Contains(row.Key, key) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("failed to find key %s", key)
		}
	}
	return nil
}
