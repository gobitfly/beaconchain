package executionlayer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type EthClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*gethtypes.Header, error)
}

type ReorgStore interface {
	GetBlock(chainID string, number uint64) (*types.Eth1Block, error)
	RevertBlock(chainID string, number uint64, blockHash []byte) error
}

type ReorgWatcher struct {
	client         EthClient
	store          ReorgStore
	depth          uint64
	chainID        string
	lastBlockStore db2.LastBlocksStoreWriter
}

func NewReorgWatcher(client EthClient, store ReorgStore, depth uint64, chainID string, lastBlockStore db2.LastBlocksStoreWriter) *ReorgWatcher {
	return &ReorgWatcher{
		client:         client,
		store:          store,
		depth:          depth,
		chainID:        chainID,
		lastBlockStore: lastBlockStore,
	}
}

func (r *ReorgWatcher) LookForReorg() error {
	head, err := r.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}

	// ensure we will not try to retrieve blocks that do not exist
	depth := r.depth
	if r.depth > head.Number.Uint64() {
		depth = head.Number.Uint64()
	}

	ctx := context.Background()

	// for each block check if block node hash and block db hash match
	for i := head.Number.Uint64() - depth; i <= head.Number.Uint64(); i++ {
		nodeBlock, err := r.client.HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return err
		}
		dbBlock, err := r.store.GetBlock(r.chainID, i)
		if err != nil {
			// exit if we hit a block that is not yet in the db
			// it means that we never processed that block or that we revert that block
			if errors.Is(err, database.ErrNotFound) {
				return nil
			}
			return err
		}

		if bytes.Equal(nodeBlock.Hash().Bytes(), dbBlock.Hash) {
			continue
		}
		log.Warnf("found incosistency at height %v, node block hash: %x, db block hash: %x", i, nodeBlock.Hash().Bytes(), dbBlock.Hash)
		if err := r.handleReorg(i, head.Number.Uint64()); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReorgWatcher) handleReorg(number uint64, head uint64) error {
	// first we set the cached marker of the last block in the blocks/data table to the block prior to the forked one
	if number > 0 {
		previousBlock := number - 1
		if err := r.lastBlockStore.SetInBlocksTable(r.chainID, previousBlock); err != nil {
			return fmt.Errorf("error setting last block [%v] in blocks table: %w", previousBlock, err)
		}
		if err := r.lastBlockStore.SetInDataTable(r.chainID, previousBlock); err != nil {
			return fmt.Errorf("error setting last block [%v] in data table: %w", previousBlock, err)
		}
	}
	// delete all blocks starting from the fork block up to the latest block in the db
	for j := number; j <= head; j++ {
		dbBlock, err := r.store.GetBlock(r.chainID, j)
		if err != nil {
			// exit if we hit a block that is not yet in the db
			if errors.Is(err, database.ErrNotFound) {
				return nil
			}
			return err
		}
		log.Infof("deleting block at height %v with hash %x", dbBlock.Number, dbBlock.Hash)
		if err := r.store.RevertBlock(r.chainID, dbBlock.Number, dbBlock.Hash); err != nil {
			return err
		}
	}
	return nil
}
