package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

func (d *DataAccessService) GetValidatorDashboardExecutionLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsElColumn], search string, limit uint64) ([]t.VDBConsolidationsElTableRow, *t.Paging, error) {
	return nil, nil, nil
}

func (d *DataAccessService) GetValidatorDashboardConsensusLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsClColumn], search string, limit uint64) ([]t.VDBConsolidationsClTableRow, *t.Paging, error) {
	// WIP
	/*var currentCursor t.ConsolidationsCursor
	var err error
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.ConsolidationsCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as BlocksCursor: %w", err)
		}
	}

	consolidationsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_consolidation_requests").As("bcr")).
		Select(
			goqu.I("block_slot").As("slot"),
			goqu.L("block_slot / ?", d.config.ClConfig.SlotsPerEpoch).As("epoch"),
			goqu.I("source_index").As("source"),
			goqu.I("target_index").As("target"),
			goqu.L("amount_consolidated::decimal * ?", 1e9).As("amount"),
		).
		InnerJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("bcr.block_root").Eq(goqu.I("b.blockroot")),
				goqu.L("b.status = '1'"),
			),
		)

	defaultColumns := []t.SortColumn{
		{Column: goqu.I("block_slot"), Desc: true, Offset: currentCursor.Slot},
		{Column: goqu.I("request_index"), Desc: true, Offset: currentCursor.ConsolidationIndex},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBConsolidationsColumns.Epoch:
		colSort.Column = enums.VDBConsolidationsColumns.Slot
		fallthrough
	case enums.VDBConsolidationsColumns.Slot:
		offset = currentCursor.Slot
	case enums.VDBConsolidationsColumns.Index:
		offset = currentCursor.ConsolidationIndex
	}

	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}
	consolidationsDs = consolidationsDs.
		Order(order...)
	if directions != nil {
		consolidationsDs = consolidationsDs.Where(directions)
	}
	res, err := runQueryRows[[]t.VDBConsolidationsTableRow](ctx, db.ReaderDb, consolidationsDs)
	return res, &t.Paging{}, err*/
	return nil, nil, nil
}
