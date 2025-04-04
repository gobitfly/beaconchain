package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
)

func (d *DataAccessService) GetValidatorDashboardExecutionLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsElColumn], search string, limit uint64) ([]t.VDBConsolidationsElTableRow, *t.Paging, error) {
	var err error
	var currentCursor t.ELConsolidationsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ELConsolidationsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ELConsolidationsCursor: %w", err)
		}
	}

	// filters
	searchSenderOrConsolidator := t.ReEthereumAddress.MatchString(search)
	searchIndexOrBlock := t.ReInteger.MatchString(search)

	if search != "" && !searchSenderOrConsolidator && !searchIndexOrBlock {
		return make([]t.VDBConsolidationsElTableRow, 0), &t.Paging{}, nil
	}

	ds := goqu.Dialect("postgres").
		From(goqu.T("eth1_consolidation_requests").As("el_cr")).
		Select(
			goqu.I("vs.validatorindex").As("source_index"),
			goqu.I("vt.validatorindex").As("target_index"),
			goqu.I("el_cr.block_number").As("block_queued"),
			goqu.I("el_cr.block_ts").As("block_queued_ts"),
			// goqu.I("el_cr.block_processed"),
			// goqu.I("el_cr.block_processed_ts"),
			// goqu.L("COALESCE(el_cr.log_index, 0) AS log_index"), // BEDS-1405
			// goqu.L("el_cr.from_address"), // BEDS-1405
			goqu.I("el_cr.tx_index"),
			goqu.I("el_cr.source_address").As("consolidator"),
			goqu.I("el_cr.tx_hash"),
			// goqu.I("el_cr.fee"), // BEDS-1405,
			// goqu.Case().As("status"), // BEDS-1399
		).
		InnerJoin(
			goqu.T("validators").As("vs"),
			goqu.On(goqu.I("el_cr.source_pubkey").Eq(goqu.I("vs.pubkey"))),
		).
		InnerJoin(
			goqu.T("validators").As("vt"),
			goqu.On(goqu.I("el_cr.target_pubkey").Eq(goqu.I("vt.pubkey"))),
		)
		// depends on BEDS-1399
		// need to match EL events to correct CL events
		// might needs some complex logic like "newest CL event after EL event which doesn't also match a previous EL event"
		/*LeftJoin(
			goqu.T("blocks_consolidation_requests").As("cl_cr"),
			goqu.On(
				?
			),
		)*/

	if dashboardId.Validators != nil {
		ds = ds.
			Where(goqu.Or(
				goqu.L("vs.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
				goqu.L("vt.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
			))
	} else {
		ds = ds.
			InnerJoin(
				goqu.T("users_val_dashboards_validators").As("uvdv"),
				goqu.On(goqu.Or(
					goqu.I("vs.validatorindex").Eq(goqu.I("uvdv.validator_index")),
					goqu.I("vt.validatorindex").Eq(goqu.I("uvdv.validator_index")),
				)),
			).
			Where(goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id))
	}

	searches := []exp.Expression{}
	if searchSenderOrConsolidator {
		address, err := hexutil.Decode(search)
		if err != nil {
			return nil, nil, err
		}
		searches = append(searches,
			goqu.I("el_cr.source_address").Eq(address),
			// goqu.I("el_cr.from_address").Eq(address), // BEDS-1405
		)
	}
	if searchIndexOrBlock {
		searches = append(
			searches, goqu.I("el_cr.block_number").Eq(search),
			goqu.I("vs.validatorindex").Eq(search),
			goqu.I("vt.validatorindex").Eq(search),
		)
	}
	if len(searches) > 0 {
		ds = ds.Where(goqu.Or(searches...))
	}

	defaultColumns := []t.SortColumn{
		{Column: goqu.I("el_cr.block_number"), Desc: true, Offset: currentCursor.BlockProcessed},
		{Column: goqu.I("el_cr.tx_index"), Desc: true, Offset: currentCursor.TxIndex}, // should be log/itx_index; BEDS-1405
	}
	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	ds = ds.
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		ds = ds.Where(directions)
	}

	type elDbResult struct {
		SourceIndex     uint64    `db:"source_index"`
		TargetIndex     uint64    `db:"target_index"`
		BlockQueued     uint64    `db:"block_queued"`
		BlockQueuedTime time.Time `db:"block_queued_ts"`
		TxIndex         uint64    `db:"tx_index"` // for cursor only
		// ITxIndex          uint64         `db:"itx_index"` 		  // for cursor only // BEDS-1405
		BlockProcessed     sql.NullInt64 `db:"block_processed"`    // need CL queued events from BEDS-1399
		BlockProcessedTime sql.NullTime  `db:"block_processed_ts"` // need CL queued events from BEDS-1399
		// From              []byte         `db:"from_address"` 	  // BEDS-1405 (might get from BT for now)
		Consolidator []byte `db:"consolidator"`
		TxHash       []byte `db:"tx_hash"`
		// Fee               decimal.Decimal   `db:"fee"` 			  // BEDS-1405
		// Status            string   `db:"status"` 			      // BEDS-1399
	}
	dbRes, err := runQueryRows[[]elDbResult](ctx, d.readerDb, ds)
	if err != nil {
		return nil, nil, err
	}

	responseData := make([]t.VDBConsolidationsElTableRow, 0, len(dbRes))
	addressMapping := make(map[string]*t.Address)
	// fromContractStatusRequests := make([]db.ContractInteractionAtRequest, len(dbRes)) // BEDS-1405
	consolidatorContractStatusRequests := make([]db.ContractInteractionAtRequest, len(dbRes))
	prepareAddressRequest := func(contractStateReqs *[]db.ContractInteractionAtRequest, addr []byte, dbRes *elDbResult) t.Address {
		req := db.ContractInteractionAtRequest{
			Address:  fmt.Sprintf("%x", addr),
			Block:    -1,
			TxIdx:    -1,
			TraceIdx: -1,
		}
		if dbRes.BlockProcessed.Valid {
			req.Block = dbRes.BlockProcessed.Int64
		}
		*contractStateReqs = append(*contractStateReqs, req)
		addrStr := hexutil.Encode(addr)
		addressMapping[addrStr] = nil
		return t.Address{Hash: t.Hash(addrStr)}
	}
	for _, res := range dbRes {
		row := t.VDBConsolidationsElTableRow{
			Source:          res.SourceIndex,
			Target:          res.TargetIndex,
			BlockQueued:     res.BlockQueued,
			TimestampQueued: res.BlockQueuedTime.Unix(),
			// Status:             res.Status, // BEDS-1399
			TxHash: t.Hash(hexutil.Encode(res.TxHash)),
			// Fee:               res.Fee, // BEDS-1405
		}
		row.Consolidator = prepareAddressRequest(&consolidatorContractStatusRequests, res.Consolidator, &res)
		// BEDS-1405
		/*row.From = row.Consolidator
		if !bytes.Equal(res.Consolidator, res.From) {
			row.From = prepareAddressRequest(&fromContractStatusRequests, res.From, &res)
		}*/

		if res.BlockProcessedTime.Valid { // BEDS-1399
			row.Status = "processed"
			row.BlockProcessed = uint64(res.BlockProcessed.Int64)
			row.TimestampProcessed = res.BlockProcessedTime.Time.Unix()
		} else {
			row.Status = "queued"
			// TODO implement estimate

			/*slotDs := goqu.Dialect("postgres").
				From(goqu.T("blocks").As("b")).
				Select(
					goqu.I("b.slot").As("slot"),
				).
				Where(
					goqu.I("b.exec_block_number").Eq(res.BlockQueued),
					goqu.I("b.status").Eq("1"),
				)

			slot_queued, err := runQuery[uint64](ctx, d.readerDb, slotDs)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get slot for block %d: %w", res.BlockQueued, err)
			}

			var slot_estimate uint64

			slotDistanceEstimate := slot_estimate - slot_queued
			row.BlockProcessed = row.BlockQueued + slotDistanceEstimate
			row.TimestampProcessed = row.TimestampQueued + int64(utils.Config.ClConfig.SecondsPerSlot*slotDistanceEstimate)*/
		}

		responseData = append(responseData, row)
	}

	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, nil, err
	}
	// BEDS-1405
	/*fromContractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(fromContractStatusRequests)
	if err != nil {
		return nil, nil, err
	}*/
	depositorContractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(consolidatorContractStatusRequests)
	if err != nil {
		return nil, nil, err
	}
	for i := range dbRes {
		// BEDS-1405
		// responseData[i].From = *addressMapping[string(responseData[i].From.Hash)]
		// responseData[i].From.IsContract = fromContractStatuses[i] == types.CONTRACT_CREATION || fromContractStatuses[i] == types.CONTRACT_PRESENT
		responseData[i].Consolidator = *addressMapping[string(responseData[i].Consolidator.Hash)]
		responseData[i].Consolidator.IsContract = depositorContractStatuses[i] == types.CONTRACT_CREATION || depositorContractStatuses[i] == types.CONTRACT_PRESENT
	}

	var paging t.Paging
	moreDataFlag := len(responseData) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return responseData, &paging, nil
	}
	if moreDataFlag {
		// Remove the last entry as it is only required for the more data flag
		responseData = responseData[:len(responseData)-1]
		dbRes = dbRes[:len(dbRes)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
	}

	p, err := utils.GetPagingFromData(dbRes, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func (d *DataAccessService) GetValidatorDashboardConsensusLayerConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsClColumn], search string, limit uint64) ([]t.VDBConsolidationsClTableRow, *t.Paging, error) {
	// WIP
	/*var currentCursor t.CLConsolidationsCursor
	var err error
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.CLConsolidationsCursor](cursor); err != nil {
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
	return nil, &t.Paging{}, nil
}
