package data

import (
	"context"
	"encoding/hex"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var (
	alice = common.HexToAddress("0x000000000000000000000000000000000000abba")
	bob   = common.HexToAddress("0x000000000000000000000000000000000000beef")
	carl  = common.HexToAddress("0x000000000000000000000000000000000000cafe")
	usdc  = common.HexToAddress("0x000000000000000000000000000000000000dead")
)

func TestStore(t *testing.T) {
	client, admin := databasetest.NewBigTable(t)

	s, err := database.NewBigTableWithClient(context.Background(), client, admin, Schema)
	if err != nil {
		t.Fatal(err)
	}
	store := Store{
		db: database.Wrap(s, Table),
	}

	tests := []struct {
		name           string
		txs            map[string][][]*types.Eth1TransactionIndexed // map[chainID][block][txPosition]*types.Eth1TransactionIndexed
		transfers      map[string][][]TransferWithIndexes
		limit          int64
		opts           []Option
		addresses      []common.Address
		expectedHashes []string
	}{
		{
			name: "one sender one chain ID",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", alice, bob, "", 1)},
					{newTx("hash3", alice, bob, "", 2)},
				},
			},
			limit:          1,
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3", "hash2", "hash1"},
		},
		{
			name: "two sender one chain ID",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", carl, bob, "", 1)},
				},
			},
			limit:          2,
			addresses:      []common.Address{alice, carl},
			expectedHashes: []string{"hash2", "hash1"},
		},
		{
			name: "two sender one chain ID with limit",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", carl, bob, "", 1)},
					{newTx("hash3", alice, bob, "", 2)},
					{newTx("hash4", carl, bob, "", 3)},
				},
			},
			limit:          2,
			addresses:      []common.Address{alice, carl},
			expectedHashes: []string{"hash4", "hash3", "hash2", "hash1"},
		},
		{
			name: "two sender each on one chain",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash3", alice, bob, "", 2)},
				},
				"2": {
					{newTx("hash2", carl, bob, "", 1)},
					{newTx("hash4", carl, bob, "", 3)},
				},
			},
			limit:          2,
			addresses:      []common.Address{alice, carl},
			expectedHashes: []string{"hash4", "hash3", "hash2", "hash1"},
		},
		{
			name: "two sender both on two chain",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash3", carl, bob, "", 2)},
				},
				"2": {
					{newTx("hash2", carl, bob, "", 1)},
					{newTx("hash4", alice, bob, "", 3)},
				},
			},
			limit:          2,
			addresses:      []common.Address{alice, carl},
			expectedHashes: []string{"hash4", "hash3", "hash2", "hash1"},
		},
		{
			name: "by method",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "foo", 0)},
					{newTx("hash2", alice, bob, "bar", 1)},
					{newTx("hash3", carl, bob, "foo", 2)},
				},
			},
			limit:          1,
			opts:           []Option{IgnoreTransfers(), ByMethod(hex.EncodeToString([]byte("foo")))},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash1"},
		},
		{
			name: "by time range",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", alice, bob, "", 1)},
					{newTx("hash3", alice, bob, "", 2)},
				},
			},
			limit:          1,
			opts:           []Option{WithTimeRange(timestamppb.New(t0), timestamppb.New(t0.Add(1*time.Second)))},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash2", "hash1"},
		},
		{
			name: "by sender",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", bob, alice, "", 1)},
					{newTx("hash3", alice, bob, "", 2)},
				},
			},
			limit:          1,
			opts:           []Option{OnlySent()},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3", "hash1"},
		},
		{
			name: "by receiver",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash2", bob, alice, "", 1)},
					{newTx("hash3", alice, bob, "", 2)},
				},
			},
			limit:          1,
			opts:           []Option{OnlyReceived()},
			addresses:      []common.Address{bob},
			expectedHashes: []string{"hash3", "hash1"},
		}, {
			name: "only transfers",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash2", alice, bob, "", 1)},
				},
			},
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash1", alice, bob, common.Address{}, 0)},
					{newTransfer("hash3", alice, bob, common.Address{}, 2)},
				},
			},
			limit:          1,
			opts:           []Option{IgnoreTransactions()},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3", "hash1"},
		},
		{
			name: "only txs",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash3", alice, bob, "", 2)},
				},
			},
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash2", alice, bob, common.Address{}, 1)},
				},
			},
			limit:          1,
			opts:           []Option{IgnoreTransfers()},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3", "hash1"},
		},
		{
			name: "mix of both",
			txs: map[string][][]*types.Eth1TransactionIndexed{
				"1": {
					{newTx("hash1", alice, bob, "", 0)},
					{newTx("hash3", alice, bob, "", 2)},
					{newTx("hash5", alice, bob, "", 4)},
					{newTx("hash7", alice, bob, "", 6)},
					{newTx("hash9", alice, bob, "", 8)},
				},
			},
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash2", alice, bob, common.Address{}, 1)},
					{newTransfer("hash4", alice, bob, common.Address{}, 3)},
					{newTransfer("hash6", alice, bob, common.Address{}, 5)},
					{newTransfer("hash8", alice, bob, common.Address{}, 7)},
					{newTransfer("hash10", alice, bob, common.Address{}, 9)},
				},
			},
			limit:          2,
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash10", "hash9", "hash8", "hash7", "hash6", "hash5", "hash4", "hash3", "hash2", "hash1"},
		},
		{
			name: "by asset with time range",
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash1", alice, bob, usdc, 0)},
					{newTransfer("hash2", alice, bob, usdc, 1)},
					{newTransfer("hash3", alice, bob, usdc, 2)},
					{newTransfer("hash4", alice, bob, usdc, 3)},
					{newTransfer("hash5", alice, bob, usdc, 4)},
				},
			},
			limit:          2,
			opts:           []Option{IgnoreTransactions(), ByAsset(usdc), WithTimeRange(timestamppb.New(t0.Add(1*time.Second)), timestamppb.New(t0.Add(3*time.Second)))},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash4", "hash3", "hash2"},
		},
		{
			name: "by asset and sender with time range",
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash1", alice, bob, usdc, 0)},
					{newTransfer("hash2", bob, alice, usdc, 1)},
					{newTransfer("hash3", alice, bob, usdc, 2)},
					{newTransfer("hash4", bob, alice, usdc, 3)},
					{newTransfer("hash5", alice, bob, usdc, 4)},
				},
			},
			limit:          2,
			opts:           []Option{IgnoreTransactions(), OnlySent(), ByAsset(usdc), WithTimeRange(timestamppb.New(t0.Add(1*time.Second)), timestamppb.New(t0.Add(3*time.Second)))},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3"},
		},
		{
			name: "by asset and receiver with time range",
			transfers: map[string][][]TransferWithIndexes{
				"1": {
					{newTransfer("hash1", bob, alice, usdc, 0)},
					{newTransfer("hash2", alice, bob, usdc, 1)},
					{newTransfer("hash3", bob, alice, usdc, 2)},
					{newTransfer("hash4", alice, bob, usdc, 3)},
					{newTransfer("hash5", bob, alice, usdc, 4)},
				},
			},
			limit:          2,
			opts:           []Option{IgnoreTransactions(), OnlyReceived(), ByAsset(usdc), WithTimeRange(timestamppb.New(t0.Add(1*time.Second)), timestamppb.New(t0.Add(3*time.Second)))},
			addresses:      []common.Address{alice},
			expectedHashes: []string{"hash3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() { _ = s.Clear() }()
			for chainID, blocks := range tt.txs {
				for _, txs := range blocks {
					if err := store.AddBlockTransactions(chainID, txs); err != nil {
						t.Fatal(err)
					}
				}
			}
			for chainID, blocks := range tt.transfers {
				for _, transfers := range blocks {
					if err := store.AddBlockERC20Transfers(chainID, transfers); err != nil {
						t.Fatal(err)
					}
				}
			}
			chainIDs := append(maps.Keys(tt.txs), maps.Keys(tt.transfers)...)
			sort.Strings(chainIDs)
			chainIDs = slices.Compact(chainIDs)
			var suffix map[string]map[string]string
			for i := int64(0); i < int64(len(tt.expectedHashes))/tt.limit; i++ {
				txs, newSuffix, err := store.Get(chainIDs, tt.addresses, suffix, tt.limit, tt.opts...)
				if err != nil {
					t.Fatalf("tx %d: %v", i, err)
				}
				if len(txs) == 0 {
					t.Fatalf("tx %d: no transactions found", i)
				}
				if got, want := int64(len(txs)), tt.limit; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				for j := int64(0); j < tt.limit; j++ {
					if got, want := string(txs[j].Hash), tt.expectedHashes[i*tt.limit+j]; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
				suffix = newSuffix
			}
		})
	}
}

var t0 = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func newTx(hash string, from, to common.Address, method string, delta int64) *types.Eth1TransactionIndexed {
	return &types.Eth1TransactionIndexed{
		Hash:               []byte(hash),
		BlockNumber:        0,
		Time:               timestamppb.New(t0.Add(time.Duration(delta) * time.Second)),
		MethodId:           []byte(method),
		From:               from.Bytes(),
		To:                 to.Bytes(),
		Value:              nil,
		TxFee:              nil,
		GasPrice:           nil,
		IsContractCreation: false,
		InvokesContract:    false,
		ErrorMsg:           "",
		BlobTxFee:          nil,
		BlobGasPrice:       nil,
	}
}

func newTransfer(hash string, from, to, contract common.Address, delta int64) TransferWithIndexes {
	return TransferWithIndexes{
		Indexed: &types.Eth1ERC20Indexed{
			ParentHash:   []byte(hash),
			BlockNumber:  0,
			TokenAddress: contract.Bytes(),
			Time:         timestamppb.New(t0.Add(time.Duration(delta) * time.Second)),
			From:         from.Bytes(),
			To:           to.Bytes(),
			Value:        nil,
		},
		TxIndex:  0,
		LogIndex: 0,
	}
}
