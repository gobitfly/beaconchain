package executionlayer

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/db2test"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var headOfBlock1 = gethtypes.Header{Number: big.NewInt(1)}

func TestReorgWatcher(t *testing.T) {
	tests := []struct {
		name           string
		nodeBlocks     []*gethtypes.Header
		dbBlocks       []*types.Eth1Block
		depth          uint64
		lastBlockStore db2.LastBlocksStoreWriter
		reverted       []uint64
	}{
		{
			name: "no reorg",
			nodeBlocks: []*gethtypes.Header{
				&headOfBlock1,
			},
			dbBlocks: []*types.Eth1Block{
				{Number: 1, Hash: headOfBlock1.Hash().Bytes()},
			},
			depth:          0,
			lastBlockStore: noopLastBlocksStoreWriter{},
		},
		{
			name: "reorg with depth 0",
			nodeBlocks: []*gethtypes.Header{
				{Number: big.NewInt(1)},
			},
			dbBlocks: []*types.Eth1Block{
				{Number: 1, Hash: common.FromHex("01")},
			},
			depth:          0,
			lastBlockStore: noopLastBlocksStoreWriter{},
			reverted:       []uint64{1},
		},
		{
			name: "reorg with depth 2",
			nodeBlocks: []*gethtypes.Header{
				{Number: big.NewInt(1)},
				{Number: big.NewInt(2)},
				{Number: big.NewInt(3)},
				{Number: big.NewInt(4)},
				{Number: big.NewInt(5)},
			},
			dbBlocks: []*types.Eth1Block{
				{Number: 1, Hash: common.FromHex("01")},
				{Number: 2, Hash: common.FromHex("02")},
				{Number: 3, Hash: common.FromHex("03")},
				{Number: 4, Hash: common.FromHex("04")},
				{Number: 5, Hash: common.FromHex("05")},
			},
			depth:          2,
			lastBlockStore: noopLastBlocksStoreWriter{},
			reverted:       []uint64{3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newFakeClient(tt.nodeBlocks...)
			store := newStubReorgStore(tt.dbBlocks...)
			reorg := NewReorgWatcher(client, store, tt.depth, "chainID", tt.lastBlockStore)
			if err := reorg.LookForReorg(); err != nil {
				t.Fatal(err)
			}
			if got, want := store.reverted, tt.reverted; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

type noopLastBlocksStoreWriter struct{}

func (f noopLastBlocksStoreWriter) SetInBlocksTable(chainID string, number uint64) error {
	return nil
}

func (f noopLastBlocksStoreWriter) SetInDataTable(chainID string, number uint64) error {
	return nil
}

type fakeClient struct {
	headers map[uint64]*gethtypes.Header
}

func newFakeClient(headers ...*gethtypes.Header) fakeClient {
	stub := fakeClient{headers: make(map[uint64]*gethtypes.Header)}
	for _, header := range headers {
		stub.headers[header.Number.Uint64()] = header
	}
	return stub
}

func (s fakeClient) HeaderByNumber(ctx context.Context, number *big.Int) (*gethtypes.Header, error) {
	if number == nil {
		number = new(big.Int)
		for k := range s.headers {
			if k > number.Uint64() {
				number.SetUint64(k)
			}
		}
	}
	return s.headers[number.Uint64()], nil
}

type stubReorgStore struct {
	blocks   map[uint64]*types.Eth1Block
	reverted []uint64
}

func newStubReorgStore(blocks ...*types.Eth1Block) *stubReorgStore {
	stub := stubReorgStore{blocks: make(map[uint64]*types.Eth1Block)}
	for _, block := range blocks {
		stub.blocks[block.Number] = block
	}
	return &stub
}

func (s *stubReorgStore) GetBlock(chainID string, number uint64) (*types.Eth1Block, error) {
	if _, ok := s.blocks[number]; !ok {
		return nil, database.ErrNotFound
	}
	return s.blocks[number], nil
}

func (s *stubReorgStore) RevertBlock(chainID string, number uint64, blockHash []byte) error {
	delete(s.blocks, number)
	s.reverted = append(s.reverted, number)
	return nil
}

func TestReorgWithBackendAndIndexer(t *testing.T) {
	backend := th.NewBackend(t)
	chainID := strconv.Itoa(backend.ChainID)

	store, lastBlockStore := db2test.NewStoreAndCachedLastBlocks(t)
	indexer := NewIndexer(
		store,
		lastBlockStore,
		AllTransformers...,
	)

	reorg := NewReorgWatcher(backend.Client(), store, 0, chainID, lastBlockStore)

	client, err := rpc.NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	// create root block that won't be reverted
	backend.Commit()
	root := indexLastBlock(t, backend, client, indexer)

	// create block (root+1) that will be reverted
	backend.Commit()
	ethBlockToReverted := indexLastBlock(t, backend, client, indexer)

	revertedBlock, err := store.GetBlock(chainID, ethBlockToReverted.NumberU64())
	if err != nil {
		t.Fatal(err)
	}

	// revert chain to root block
	if err := backend.Fork(root.Hash()); err != nil {
		t.Fatal(err)
	}

	// create block that has the same number as revertedBlock (root+1)
	backend.Commit()

	if err := reorg.LookForReorg(); err != nil {
		t.Fatal(err)
	}

	// block should have been deleted from the store
	_, err = store.GetBlock(chainID, ethBlockToReverted.NumberU64())
	if got, want := err, database.ErrNotFound; !errors.Is(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// index last block and compare it to the reverted block
	indexLastBlock(t, backend, client, indexer)
	newBlock, err := store.GetBlock(chainID, ethBlockToReverted.NumberU64())
	if err != nil {
		t.Fatal(err)
	}
	if got, dontWant := newBlock.Hash, revertedBlock.Hash; bytes.Equal(got, dontWant) {
		t.Errorf("block hash %x should have been reverted", got)
	}
}

func indexLastBlock(t *testing.T, backend *th.BlockchainBackend, client *rpc.ErigonClient, indexer *Indexer) *gethtypes.Block {
	lastBlock, err := backend.Client().BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := indexer.Index(strconv.Itoa(backend.ChainID), client, lastBlock.NumberU64(), lastBlock.NumberU64(), 1, "geth"); err != nil {
		t.Fatal(err)
	}
	return lastBlock
}
