package dataaccess

import (
	"context"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type BlockRepository interface {
	GetBlock(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error)
	GetBlockOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error)
	GetBlockTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error)
	GetBlockVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error)
	GetBlockAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error)
	GetBlockWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error)
	GetBlockBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error)
	GetBlockVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error)
	GetBlockBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error)
}

func (d *DataAccessService) GetBlock(ctx context.Context, chainId, block uint64) (*t.BlockSummary, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlock(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockOverview(ctx context.Context, chainId, block uint64) (*t.BlockOverview, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockOverview(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockTransactions(ctx context.Context, chainId, block uint64) ([]t.BlockTransactionTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockTransactions(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockVotes(ctx context.Context, chainId, block uint64) ([]t.BlockVoteTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockVotes(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockAttestations(ctx context.Context, chainId, block uint64) ([]t.BlockAttestationTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockAttestations(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockWithdrawals(ctx context.Context, chainId, block uint64) ([]t.BlockWithdrawalTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockWithdrawals(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockBlsChanges(ctx context.Context, chainId, block uint64) ([]t.BlockBlsChangeTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockBlsChanges(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockVoluntaryExits(ctx context.Context, chainId, block uint64) ([]t.BlockVoluntaryExitTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockVoluntaryExits(ctx, chainId, block)
}

func (d *DataAccessService) GetBlockBlobs(ctx context.Context, chainId, block uint64) ([]t.BlockBlobTableRow, error) {
	// @DATA-ACCESS
	return d.dummy.GetBlockBlobs(ctx, chainId, block)
}
