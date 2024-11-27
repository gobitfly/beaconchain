package indexer

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadata"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type TransformFunc func(chainID string, blk *types.Eth1Block) (map[string][]database.Item, map[string][]database.Item, error)

type Indexer struct {
	store        data.Store
	transformers []TransformFunc
}

func New(store data.Store, transformers ...TransformFunc) *Indexer {
	return &Indexer{
		store:        store,
		transformers: transformers,
	}
}

func (indexer *Indexer) IndexBlocksWithTransformers(chainID string, blocks []*types.Eth1Block) error {
	updates := make(map[string][]database.Item)
	updatesMetadata := make(map[string][]database.Item)
	for _, block := range blocks {
		for _, transform := range indexer.transformers {
			update, updateMetadata, err := transform(chainID, block)
			if err != nil {
				return fmt.Errorf("error transforming block [%v]", block.Number)
			}
			maps.Copy(updates, update)

			if updateMetadata != nil {
				maps.Copy(updatesMetadata, updateMetadata)
			}

			if len(update) > 0 {
				metaKeys := strings.Join(maps.Keys(update), ",") // save block keys in order to be able to handle chain reorgs
				maps.Copy(updatesMetadata, metadata.BlockKeysMutation(chainID, block.Number, block.Hash, metaKeys))
			}
		}
	}

	if len(updates) > 0 {
		err := indexer.store.AddItems(updates)
		if err != nil {
			return fmt.Errorf("error writing blocks [%v-%v] to bigtable data table: %w", blocks[0].Number, blocks[len(blocks)-1].Number, err)
		}
	}

	// todo
	/*	if len(bulkMutsMetadataUpdate.Keys) > 0 {
		err := bigtable.WriteBulk(&bulkMutsMetadataUpdate, bigtable.tableMetadataUpdates, DEFAULT_BATCH_INSERTS)
		if err != nil {
			return fmt.Errorf("error writing blocks [%v-%v] to bigtable metadata updates table: %w", blocks[0].Number, blocks[len(blocks)-1].Number, err)
		}
	}*/

	return nil
}
