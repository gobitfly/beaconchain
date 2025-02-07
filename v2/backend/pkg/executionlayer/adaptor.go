package executionlayer

import (
	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type IndexedBlock struct {
	ChainID         string
	Block           *types.Eth1BlockIndexed
	Transactions    []*types.Eth1TransactionIndexed
	Internals       []data.InternalWithIndexes
	ERC20Transfer   []data.TransferWithIndexes
	ERC1155Transfer []data.ERC1155TransferWithIndexes
	ERC721Transfer  []data.ERC721TransferWithIndexes
	Blobs           []data.BlobWithIndex
	Uncles          []data.UncleWithIndexes
	Withdrawals     []*types.Eth1WithdrawalIndexed
	ENS             []data.ENSLog
	Contracts       []metadataupdates.ContractUpdateWithAddress
}

// AdaptorV1 represents the current storage organisation
// it consists of a data store where indexed object are saved
// and a metadataUpdates store where balances and some other data are saved
type AdaptorV1 struct {
	data            data.Store
	metadataUpdates metadataupdates.Store
}

func NewAdaptorV1(data data.Store, metadataUpdates metadataupdates.Store) AdaptorV1 {
	return AdaptorV1{
		data:            data,
		metadataUpdates: metadataUpdates,
	}
}

func (adaptor AdaptorV1) Save(blockNumber uint64, hash []byte, block IndexedBlock) error {
	keys, err := adaptor.data.AddIndexedBlock(data.IndexedBlock{
		ChainID:         block.ChainID,
		Block:           block.Block,
		Transactions:    block.Transactions,
		Internals:       block.Internals,
		ERC20Transfer:   block.ERC20Transfer,
		ERC1155Transfer: block.ERC1155Transfer,
		ERC721Transfer:  block.ERC721Transfer,
		Blobs:           block.Blobs,
		Uncles:          block.Uncles,
		Withdrawals:     block.Withdrawals,
		ENS:             block.ENS,
	})
	if err != nil {
		return err
	}
	metadataBlock := metadataupdates.IndexedBlock{
		ChainID:       block.ChainID,
		Block:         block.Block,
		Transactions:  block.Transactions,
		Internals:     block.Internals,
		ERC20Transfer: block.ERC20Transfer,
		Blobs:         block.Blobs,
		Contracts:     block.Contracts,
		Uncles:        block.Uncles,
		Withdrawals:   block.Withdrawals,
	}
	if err := adaptor.metadataUpdates.AddIndexedBlock(blockNumber, hash, metadataBlock, keys); err != nil {
		return err
	}
	return nil
}
