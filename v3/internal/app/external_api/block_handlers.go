package app

import (
	"context"
	"fmt"

	model "github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/common/islices"
	"github.com/gobitfly/beaconchain-backend/internal/common/pagination"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

func (service *ApiService) GetBlockTransactions(ctx context.Context, in model.GetBlockTransactionsRequestObject) (model.GetBlockTransactionsResponseObject, error) {
	data, paging, err := pagination.Handle(
		in.Params.Cursor,
		in.Params.PageSize,
		transformBlockTransactionToCursor,
		func(cursor *domain.BlockTransactionCursor, pageSize int) ([]domain.BlockTransaction, error) {
			return mockGetBlockTransactions(ctx, in.BlockNumber, cursor, pageSize)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}

	response := model.BlockTransactions{
		Data:   islices.Transform(data, transformBlockTransactionToModel),
		Paging: paging,
	}
	return model.GetBlockTransactions200JSONResponse(response), nil
}

func transformBlockTransactionToCursor(tx domain.BlockTransaction) domain.BlockTransactionCursor {
	return domain.BlockTransactionCursor{
		Idx: tx.Idx,
	}
}

func transformBlockTransactionToModel(tx domain.BlockTransaction) model.BlockTransaction {
	return model.BlockTransaction{
		Hash: tx.Hash,
	}
}

// this is what the repo func signature should look like
func mockGetBlockTransactions(_ context.Context, block int, cursor *domain.BlockTransactionCursor, pageSize int) ([]domain.BlockTransaction, error) {
	txs := []domain.BlockTransaction{}
	startIdx := 0
	if cursor != nil {
		// in actual repo, you'd set `WHERE idx > cursor.Idx` in this block
		startIdx = cursor.Idx + 1
	}
	// in actual repo, you'd query the db here
	for i := range pageSize {
		txs = append(txs, domain.BlockTransaction{
			Hash: fmt.Sprintf("0xabc, block: %d, tx index: %d", block, i+startIdx),
			Idx:  i + startIdx,
		})
	}
	return txs, nil
}
