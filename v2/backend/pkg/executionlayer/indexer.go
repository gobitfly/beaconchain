package executionlayer

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Store interface {
	Save(blockNumber uint64, hash []byte, block IndexedBlock) (err error)
}

type Indexer struct {
	store        Store
	transformers []TransformFunc
}

func NewIndexer(store Store, transformers ...TransformFunc) *Indexer {
	return &Indexer{
		store:        store,
		transformers: transformers,
	}
}

func (indexer *Indexer) IndexBlocks(chainID string, blocks []*types.Eth1Block) error {
	for _, block := range blocks {
		if err := indexer.IndexBlock(chainID, block); err != nil {
			return fmt.Errorf("blocks [%v-%v]: %w", blocks[0].Number, blocks[len(blocks)-1].Number, err)
		}
	}

	return nil
}

func (indexer *Indexer) IndexBlock(chainID string, block *types.Eth1Block) error {
	res := IndexedBlock{
		ChainID: chainID,
	}
	for _, transform := range indexer.transformers {
		err := transform(chainID, block, &res)
		if err != nil {
			return fmt.Errorf("error transforming block [%v]", block.Number)
		}
	}
	if err := indexer.store.Save(block.Number, block.Hash, res); err != nil {
		return fmt.Errorf("error saving block [%v]: %w", block.Number, err)
	}
	return nil
}
