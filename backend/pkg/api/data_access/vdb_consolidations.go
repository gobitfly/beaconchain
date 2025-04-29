package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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
	isValidSearchSenderOrConsolidator := t.ReEthereumAddress.MatchString(search)
	isValidSearchIndexOrBlock := t.ReInteger.MatchString(search)

	if isInvalidSearch(search, isValidSearchSenderOrConsolidator, isValidSearchIndexOrBlock) {
		return make([]t.VDBConsolidationsElTableRow, 0), &t.Paging{}, nil
	}

	ds := goqu.Dialect("postgres").
		From(goqu.T("eth1_consolidation_requests").As("el_cr")).
		Select(
			goqu.I("vs.validatorindex").As("source_index"),
			goqu.I("vt.validatorindex").As("target_index"),
			goqu.I("el_cr.block_number").As("block_queued"),
			goqu.I("el_cr.block_ts").As("block_queued_ts"),
			goqu.I("b.exec_block_number").As("block_processed"),
			goqu.I("b.exec_timestamp").As("block_processed_ts"),
			goqu.I("el_cr.itx_index"),
			goqu.I("el_cr.from_address"),
			goqu.I("el_cr.tx_index"),
			goqu.I("el_cr.source_address").As("consolidator"),
			goqu.I("el_cr.tx_hash"),
			goqu.I("el_cr.fee"),
			goqu.Case().When(
				goqu.I("b.exec_block_number").Neq(nil), "processed",
			).Else(
				goqu.V("queued"),
			).As("status"),
		).
		InnerJoin(
			goqu.T("validators").As("vs"),
			goqu.On(goqu.I("el_cr.source_pubkey").Eq(goqu.I("vs.pubkey"))),
		).
		InnerJoin(
			goqu.T("validators").As("vt"),
			goqu.On(goqu.I("el_cr.target_pubkey").Eq(goqu.I("vt.pubkey"))),
		).
		LeftJoin(
			goqu.T("blocks_consolidation_requests_v2").As("cl_cr"),
			goqu.On(
				goqu.I("el_cr.id").Eq(goqu.I("cl_cr.id")),
			),
		).
		LeftJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("b.blockroot").Eq(goqu.I("cl_cr.block_queued_root")),
			),
		)

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
	if isValidSearchSenderOrConsolidator {
		address, err := hexutil.Decode(search)
		if err != nil {
			return nil, nil, err
		}
		searches = append(searches,
			goqu.I("el_cr.source_address").Eq(address),
			// goqu.I("el_cr.from_address").Eq(address), // BEDS-1405
		)
	}
	if isValidSearchIndexOrBlock {
		searches = append(
			searches, goqu.I("el_cr.block_number").Eq(search),
			goqu.I("vs.validatorindex").Eq(search),
			goqu.I("vt.validatorindex").Eq(search),
		)
	}
	if len(searches) > 0 {
		ds = ds.Where(goqu.Or(searches...))
	}

	defaultSlotSortDesc := true
	if colSort.Column == enums.VDBConsolidationsElColumns.BlockProcessed {
		// this implements a form of multicolumn sort which we don't want to support atm, but for a time-sensitive sort it should be justified
		defaultSlotSortDesc = colSort.Desc
	}
	defaultColumns := []t.SortColumn{
		{Column: goqu.I("el_cr.block_number"), Desc: defaultSlotSortDesc, Offset: currentCursor.BlockProcessed},
		{Column: goqu.I("el_cr.tx_index"), Desc: defaultSlotSortDesc, Offset: currentCursor.TxIndex},   // should be log/itx_index; BEDS-1405
		{Column: goqu.I("el_cr.itx_index"), Desc: defaultSlotSortDesc, Offset: currentCursor.ITxIndex}, // should be log/itx_index; BEDS-1405
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
		SourceIndex        uint64        `db:"source_index"`
		TargetIndex        uint64        `db:"target_index"`
		BlockQueued        uint64        `db:"block_queued"`
		BlockQueuedTime    time.Time     `db:"block_queued_ts"`
		TxIndex            uint64        `db:"tx_index"`           // for cursor only
		ITxIndex           uint64        `db:"itx_index"`          // for cursor only // BEDS-1405
		BlockProcessed     sql.NullInt64 `db:"block_processed"`    // need CL queued events from BEDS-1399
		BlockProcessedTime sql.NullInt64 `db:"block_processed_ts"` // need CL queued events from BEDS-1399
		From               []byte        `db:"from_address"`       // BEDS-1405 (might get from BT for now)
		Consolidator       []byte        `db:"consolidator"`
		TxHash             []byte        `db:"tx_hash"`
		Fee                []byte        `db:"fee"` // BEDS-1405 // TODO change to number
		Status             string        `db:"status"`
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
			TxIndexQueued:   res.TxIndex,
			ITxIndexQueued:  res.ITxIndex,
			TimestampQueued: res.BlockQueuedTime.Unix(),
			// Status:             res.Status, // BEDS-1399
			TxHash: t.Hash(hexutil.Encode(res.TxHash)),
		}
		row.Consolidator = prepareAddressRequest(&consolidatorContractStatusRequests, res.Consolidator, &res)
		// BEDS-1405
		/*row.From = row.Consolidator
		if !bytes.Equal(res.Consolidator, res.From) {
			row.From = prepareAddressRequest(&fromContractStatusRequests, res.From, &res)
		}*/

		row.Fee = decimal.NewFromBigInt(new(big.Int).SetBytes(res.Fee), 0)

		if res.BlockProcessedTime.Valid { // BEDS-1399
			row.Status = "processed"
			row.BlockProcessed = uint64(res.BlockProcessed.Int64)
			row.TimestampProcessed = res.BlockProcessedTime.Int64
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
	var currentCursor t.CLConsolidationsCursor
	var err error
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.CLConsolidationsCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLConsolidationsCursor: %w", err)
		}
	}

	// filters
	isValidSearchIndexOrSlot := t.ReInteger.MatchString(search)
	if isInvalidSearch(search, isValidSearchIndexOrSlot) {
		return make([]t.VDBConsolidationsClTableRow, 0), &t.Paging{}, nil
	}

	consolidationsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_consolidation_requests_v2").As("bcr")).
		Select(
			goqu.I("slot_processed"),
			goqu.I("slot_queued"),
			goqu.I("vs.validatorindex").As("source"),
			goqu.I("vt.validatorindex").As("target"),
			goqu.L("amount_consolidated::decimal").As("amount"),
			goqu.I("bcr.status"),
			goqu.I("bcr.reject_reason"),
			goqu.I("bcr.id"),
			// cursor
			enums.VDBConsolidationsClColumns.Slot.ToExpr().As("slot"),
			goqu.COALESCE(goqu.I("index_queued"), goqu.I("index_processed")).As("index"),
		).
		InnerJoin(
			goqu.T("validators").As("vs"),
			goqu.On(goqu.I("bcr.source_pubkey").Eq(goqu.I("vs.pubkey"))),
		).
		InnerJoin(
			goqu.T("validators").As("vt"),
			goqu.On(goqu.I("bcr.target_pubkey").Eq(goqu.I("vt.pubkey"))),
		)

	if dashboardId.Validators != nil {
		consolidationsDs = consolidationsDs.
			Where(goqu.Or(
				goqu.L("vs.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
				goqu.L("vt.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
			))
	} else {
		consolidationsDs = consolidationsDs.
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
	if isValidSearchIndexOrSlot {
		searches = append(searches,
			goqu.I("slot_queued").Eq(search),
			goqu.I("slot_processed").Eq(search),
			goqu.I("vs.validatorindex").Eq(search),
			goqu.I("vt.validatorindex").Eq(search),
		)
	}

	if len(searches) > 0 {
		consolidationsDs = consolidationsDs.Where(goqu.Or(searches...))
	}

	defaultSlotSortDesc := true
	if colSort.Column == enums.VDBConsolidationsClColumns.Slot {
		// this implements a form of multicolumn sort which we don't want to support atm, but for a time-sensitive sort it should be justified
		defaultSlotSortDesc = colSort.Desc
	}
	defaultColumns := []t.SortColumn{
		{Column: enums.VDBConsolidationsClColumns.Slot.ToExpr(), Desc: defaultSlotSortDesc, Offset: currentCursor.Slot},
		{Column: goqu.COALESCE(goqu.I("index_queued"), goqu.I("index_processed")), Desc: defaultSlotSortDesc, Offset: currentCursor.SlotIndex},
	}

	var offset any
	switch colSort.Column {
	case enums.VDBConsolidationsClColumns.Amount:
		if currentCursor.Amount != nil {
			offset = currentCursor.Amount
		}
	}

	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}
	consolidationsDs = consolidationsDs.
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		consolidationsDs = consolidationsDs.Where(directions)
	}

	type dbResult struct {
		Source        uint64           `db:"source"`
		Target        uint64           `db:"target"`
		SlotProcessed sql.NullInt64    `db:"slot_processed"`
		SlotQueued    sql.NullInt64    `db:"slot_queued"`
		Status        string           `db:"status"`
		RejectReason  sql.NullString   `db:"reject_reason"`
		Amount        *decimal.Decimal `db:"amount"`
		Id            uint64           `db:"id"`
		Slot          uint64           `db:"slot"`
		SlotIndex     uint64           `db:"index"`
	}

	res, err := runQueryRows[[]dbResult](ctx, d.readerDb, consolidationsDs)
	if err != nil {
		return nil, nil, err
	}

	responseData := make([]t.VDBConsolidationsClTableRow, 0, len(res))

	for _, r := range res {
		row := t.VDBConsolidationsClTableRow{
			Source: r.Source,
			Target: r.Target,
			Status: r.Status,
			Id:     r.Id,
		}

		// some data integrity checks TODO add more
		switch r.Status {
		case "queued":
			if r.SlotProcessed.Valid || r.Amount != nil || r.RejectReason.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for queued consolidation")
			}
			if !r.SlotQueued.Valid {
				return nil, nil, fmt.Errorf("slot_queued is not set for queued consolidation")
			}
		case "completed":
			if !r.SlotQueued.Valid || !r.SlotProcessed.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for completed consolidation")
			}
		case "rejected":
			if r.SlotQueued.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for rejected consolidation")
			}
		default:
			return nil, nil, fmt.Errorf("unknown consolidation status: %s", r.Status)
		}

		if r.SlotProcessed.Valid {
			slot := uint64(r.SlotProcessed.Int64)
			row.SlotProcessed = &slot
		} else { //nolint: staticcheck
			// TODO estimation
		}
		if r.SlotQueued.Valid {
			slot := uint64(r.SlotQueued.Int64)
			row.SlotQueued = &slot
		}
		if r.Amount != nil {
			amt := r.Amount.Mul(decimal.NewFromUint64(1e9))
			row.Amount = &amt
		}
		if r.RejectReason.Valid {
			str := mapConsolidationRejectReasonDbToApi(r.RejectReason.String)
			if str != "" {
				row.RejectReason = &str // BEDS-1399
			}
		}

		responseData = append(responseData, row)
	}

	var paging t.Paging
	moreDataFlag := len(res) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return responseData, &paging, nil
	}
	if moreDataFlag {
		// Remove the last entry as it is only required for the more data flag
		responseData = responseData[:len(responseData)-1]
		res = res[:len(res)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
		slices.Reverse(res)
	}

	p, err := utils.GetPagingFromData(res, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func mapConsolidationRejectReasonDbToApi(dbReason string) string {
	switch dbReason {
	case "source_target_pubkey_equal":
		return "source_equals_target"
	case "source_pubkey_not_found":
		return "source_unknown_pubkey"
	case "target_pubkey_not_found":
		return "target_unknown_pubkey"
	case "source_withdrawal_credentials_invalid":
		return "source_no_execution_withdrawal_credentials"
	case "source_address_mismatch":
		return "source_address_mismatch"
	case "target_withdrawal_credentials_not_compounding":
		return "target_not_compounding"
	case "source_not_active":
		return "source_inactive"
	case "target_not_active":
		return "target_inactive"
	case "source_exiting":
		return "source_exiting"
	case "target_exiting":
		return "target_exiting"
	case "source_not_active_long_enough":
		return "source_too_young"
	case "pending_balance_to_withdraw_not_zero":
		return "source_pending_withdrawals"
	case "full_queue":
		return "full_queue"
	case "insufficient_consolidation_churn_limit":
		return "insufficient_consolidation_churn"
	case "source_slashed":
		return "source_slashed"

	case "activation_epoch_overflow", "pending_balance_to_withdraw_error", "compute_consolidation_epoch_error":
		log.Warn("prysm client error")
		return ""
	default:
		log.Warnf("unknown consolidation reject reason: '%s'", dbReason)
		return ""
	}
}
