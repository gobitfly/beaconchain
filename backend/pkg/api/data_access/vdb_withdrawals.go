package dataaccess

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardElWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsElColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsElTableRow, *t.Paging, error) {
	responseData := make([]t.VDBWithdrawalsElTableRow, 0)
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.ELWithdrawalsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ELWithdrawalsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsElCursor: %w", err)
		}
	}

	// filters
	isValidSearchPubkey := t.ReValidatorPublicKeyWithPrefix.MatchString(search)
	isValidSearchSenderOrWithdrawer := t.ReEthereumAddress.MatchString(search)
	isValidSearchIndexOrBlock := t.ReInteger.MatchString(search)
	isValidSearchGroup := len(search) > 0 && !dashboardId.AggregateGroups && t.ReName.MatchString(search)
	isValidSearchTxHash := t.ReTransactionHash.MatchString(search)

	if isInvalidSearch(search, isValidSearchPubkey, isValidSearchSenderOrWithdrawer, isValidSearchIndexOrBlock, isValidSearchGroup, isValidSearchTxHash) {
		return responseData, &t.Paging{}, nil
	}

	ds := goqu.Dialect("postgres").
		From(goqu.T("eth1_withdrawal_requests").As("el_wr")).
		Select(
			goqu.I("v.validatorindex").As("index"),
			goqu.I("el_wr.block_number").As("block_queued"),
			goqu.I("el_wr.block_ts").As("block_queued_ts"),
			goqu.I("b.exec_block_number").As("block_processed"), // BEDS-1399
			goqu.I("b.exec_timestamp").As("block_processed_ts"), // BEDS-1399
			goqu.I("el_wr.itx_index"),                           // BEDS-1405
			goqu.I("el_wr.from_address"),                        // BEDS-1405
			goqu.I("el_wr.tx_index"),
			goqu.I("el_wr.source_address").As("withdrawer"),
			goqu.I("el_wr.tx_hash"),
			goqu.I("el_wr.fee"),    // BEDS-1405,
			goqu.I("el_wr.amount"), // BEDS-1405,
			goqu.Case().When( // BEDS-1399
				goqu.I("b.exec_block_number").Neq(nil), "processed",
			).Else(
				goqu.V("queued"),
			).As("status"), // BEDS-1399
		).
		InnerJoin(
			goqu.T("validators").As("v"),
			goqu.On(goqu.I("el_wr.validator_pubkey").Eq(goqu.I("v.pubkey"))),
		).
		// depends on BEDS-1399
		// need to match EL events to correct CL events
		// might needs some complex logic like "newest CL event after EL event which doesn't also match a previous EL event"
		LeftJoin(
			goqu.T("blocks_withdrawal_requests").As("cl_wr"),
			goqu.On(
				goqu.I("el_wr.id").Eq(goqu.I("cl_wr.id")),
			),
		).
		LeftJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("b.blockroot").Eq(goqu.I("cl_wr.block_queued_root")),
			),
		)

	searches := []exp.Expression{}
	if dashboardId.Validators != nil {
		ds = ds.
			Where(
				goqu.L("v.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
			)
	} else {
		ds = ds.
			InnerJoin(
				goqu.T("users_val_dashboards_validators").As("uvdv"),
				goqu.On(
					goqu.I("v.validatorindex").Eq(goqu.I("uvdv.validator_index")),
				),
			).
			Where(goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id))

		if isValidSearchGroup {
			ds = ds.
				InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
					goqu.I("uvdv.dashboard_id").Eq(goqu.I("uvdg.dashboard_id")),
					goqu.I("uvdv.group_id").Eq(goqu.I("uvdg.id")),
				))
			searches = append(searches,
				goqu.L("LOWER(?)", goqu.I("uvdg.name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1)+"%"),
			)
		}
	}

	if isValidSearchSenderOrWithdrawer {
		address, err := hexutil.Decode(search)
		if err != nil {
			return nil, nil, err
		}
		searches = append(searches,
			goqu.I("el_wr.source_address").Eq(address),
			// goqu.I("el_wr.from_address").Eq(address), // BEDS-1405
		)
	}
	if isValidSearchPubkey {
		pubkey, err := hexutil.Decode(search)
		if err != nil {
			return nil, nil, err
		}
		searches = append(searches, goqu.I("el_wr.validator_pubkey").Eq(pubkey))
	}
	if isValidSearchIndexOrBlock {
		searches = append(searches,
			goqu.I("el_wr.block_number").Eq(search),
			goqu.I("b.exec_block_number").Eq(search),
			goqu.I("v.validatorindex").Eq(search),
		)
	}
	if isValidSearchTxHash {
		searches = append(searches, goqu.I("el_wr.tx_hash").Eq(search))
	}
	if len(searches) > 0 {
		ds = ds.Where(goqu.Or(searches...))
	}

	defaultSlotSortDesc := true
	if colSort.Column == enums.VDBWithdrawalsElColumns.BlockQueued {
		// this implements a form of multicolumn sort which we don't want to support atm, but for a time-sensitive sort it should be justified
		defaultSlotSortDesc = colSort.Desc
	}
	defaultColumns := []t.SortColumn{
		{Column: goqu.I("block_number"), Desc: defaultSlotSortDesc, Offset: currentCursor.BlockQueued},
		{Column: goqu.I("tx_index"), Desc: defaultSlotSortDesc, Offset: currentCursor.TxIndex},
		{Column: goqu.I("itx_index"), Desc: defaultSlotSortDesc, Offset: currentCursor.ITxIndex},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBWithdrawalsElColumns.Amount:
		offset = currentCursor.Amount
	}
	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	ds = ds.
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		ds = ds.Where(directions)
	}

	type dbResult struct {
		GroupId            sql.NullInt64   `db:"group_id"`
		PublicKey          []byte          `db:"publickey"`
		Index              uint64          `db:"index"`
		BlockQueued        uint64          `db:"block_queued"`
		BlockQueuedTime    time.Time       `db:"block_queued_ts"`
		TxIndex            uint64          `db:"tx_index"`           // for cursor only
		ITxIndex           uint64          `db:"itx_index"`          // for cursor only
		BlockProcessed     sql.NullInt64   `db:"block_processed"`    // need CL queued events from BEDS-1399
		BlockProcessedTime sql.NullInt64   `db:"block_processed_ts"` // need CL queued events from BEDS-1399
		From               []byte          `db:"from_address"`
		Withdrawer         []byte          `db:"withdrawer"`
		TxHash             []byte          `db:"tx_hash"`
		Amount             decimal.Decimal `db:"amount"`
		Fee                []byte          `db:"fee"`    // BEDS-1405 // TODO change to number
		Status             string          `db:"status"` // BEDS-1399
	}

	queryResult, err := runQueryRows[[]dbResult](ctx, d.readerDb, ds)
	if err != nil {
		return nil, nil, err
	}

	if len(queryResult) == 0 {
		// No withdrawals found
		return responseData, &paging, nil
	}

	responseData = make([]t.VDBWithdrawalsElTableRow, 0, len(queryResult))
	addressMapping := make(map[string]*t.Address)
	// fromContractStatusRequests := make([]db.ContractInteractionAtRequest, len(dbRes)) // BEDS-1405
	withdrawerContractStatusRequests := make([]db.ContractInteractionAtRequest, len(queryResult))
	prepareAddressRequest := func(contractStateReqs *[]db.ContractInteractionAtRequest, addr []byte, dbRes *dbResult) t.Address {
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
	for _, res := range queryResult {
		row := t.VDBWithdrawalsElTableRow{
			Index:           res.Index,
			BlockQueued:     res.BlockQueued,
			TimestampQueued: res.BlockQueuedTime.Unix(),
			// Status:             res.Status, // BEDS-1399
			TxHash: t.Hash(hexutil.Encode(res.TxHash)),
			Amount: res.Amount,
		}
		if res.GroupId.Valid && !dashboardId.AggregateGroups {
			row.GroupId = uint64(res.GroupId.Int64)
		} else {
			row.GroupId = t.DefaultGroupId
		}
		row.Withdrawer = prepareAddressRequest(&withdrawerContractStatusRequests, res.Withdrawer, &res)
		// BEDS-1405
		/*row.From = row.Withdrawer
		if !bytes.Equal(res.Withdrawer, res.From) {
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

	// Get the ENS names and (label) names for the addresses
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, nil, err
	}

	// Get the contract status for the addresses
	withdrawerContractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(withdrawerContractStatusRequests)
	if err != nil {
		return nil, nil, err
	}

	// Create the result
	for i := range queryResult {
		// BEDS-1405
		// responseData[i].From = *addressMapping[string(responseData[i].From.Hash)]
		// responseData[i].From.IsContract = fromContractStatuses[i] == types.CONTRACT_CREATION || fromContractStatuses[i] == types.CONTRACT_PRESENT
		responseData[i].Withdrawer = *addressMapping[string(responseData[i].Withdrawer.Hash)]
		responseData[i].Withdrawer.IsContract = withdrawerContractStatuses[i] == types.CONTRACT_CREATION || withdrawerContractStatuses[i] == types.CONTRACT_PRESENT
	}

	// Flag if above limit
	moreDataFlag := len(responseData) > int(limit)

	// Remove the last entry from data as it is only required for the check
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return responseData, &paging, nil
	}
	if moreDataFlag {
		responseData = responseData[:len(responseData)-1]
		queryResult = queryResult[:len(queryResult)-1]
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(responseData)
	}

	p, err := utils.GetPagingFromData(queryResult, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func (d *DataAccessService) GetValidatorDashboardClWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsClColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsClTableRow, *t.Paging, error) {
	var currentCursor t.CLWithdrawalsCursor
	var err error
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.CLWithdrawalsCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLWithdrawalsCursor: %w", err)
		}
	}

	// filters
	isValidSearchWithdrawalAddress := t.ReEthereumAddress.MatchString(search)
	isValidSearchIndexOrSlot := t.ReInteger.MatchString(search)
	isValidSearchGroup := len(search) > 0 && !dashboardId.AggregateGroups && t.ReName.MatchString(search)
	isValidSearchPublicKey := t.ReValidatorPublicKeyWithPrefix.MatchString(search)

	if isInvalidSearch(search, isValidSearchWithdrawalAddress, isValidSearchIndexOrSlot, isValidSearchGroup, isValidSearchPublicKey) {
		return make([]t.VDBWithdrawalsClTableRow, 0), &t.Paging{}, nil
	}

	// Get the withdrawals for the validators
	type dbResult struct {
		GroupId               sql.NullInt64   `db:"group_id"`
		SlotProcessed         sql.NullInt64   `db:"slot_processed"`
		SlotQueued            sql.NullInt64   `db:"slot_queued"`
		ValidatorIndex        uint64          `db:"validatorindex"`
		Pubkey                []byte          `db:"pubkey"`
		Recipient             []byte          `db:"recipient"`
		WithdrawalCredentials []byte          `db:"withdrawal_credentials"`
		Amount                decimal.Decimal `db:"amount"`
		RejectReason          sql.NullString  `db:"reject_reason"`
		Status                string          `db:"status"`
		Type                  string          `db:"type"`
		// cursor
		Slot      uint64 `db:"slot"`
		SlotIndex uint64 `db:"index"`
	}

	// there is a pre- and a post-pectra table in db; only query from respective tables if possible to increase compatibility and simplicity
	hasPrePectraRows, hasPostPectraRows := true, false
	if d.config.ClConfig.ElectraForkEpoch < utils.MaxForkEpoch {
		hasPostPectraRows = true
		if currentCursor.IsValid() && colSort.Column == enums.VDBWithdrawalsClColumns.Slot {
			postElectra := currentCursor.Slot/d.config.ClConfig.SlotsPerEpoch > d.config.ClConfig.ElectraForkEpoch
			lookBack := colSort.Desc != currentCursor.Reverse
			if postElectra && !lookBack {
				hasPrePectraRows, hasPostPectraRows = false, true
			} else if !postElectra && lookBack {
				hasPrePectraRows, hasPostPectraRows = true, false
			}
		}
	}

	var withdrawalsDs *goqu.SelectDataset
	if hasPrePectraRows {
		withdrawalsDs, err = getWithdrawalsBridgeDs(dashboardId, search, isValidSearchWithdrawalAddress, isValidSearchIndexOrSlot, isValidSearchGroup, isValidSearchPublicKey)
		if err != nil {
			return nil, nil, err
		}
	}
	if hasPostPectraRows {
		requestsDs, err := getWithdrawalRequestsDs(dashboardId, search, isValidSearchWithdrawalAddress, isValidSearchIndexOrSlot, isValidSearchGroup, isValidSearchPublicKey)
		if err != nil {
			return nil, nil, err
		}
		if withdrawalsDs == nil {
			withdrawalsDs = requestsDs
		} else {
			withdrawalsDs = withdrawalsDs.UnionAll(requestsDs)
		}
	}

	defaultSlotSortDesc := true
	if colSort.Column == enums.VDBWithdrawalsClColumns.Slot {
		// this implements a form of multicolumn sort which we don't want to support atm, but for a time-sensitive sort it should be justified
		defaultSlotSortDesc = colSort.Desc
	}
	defaultColumns := []t.SortColumn{
		{Column: enums.VDBWithdrawalsClColumns.Slot.ToExpr(), Desc: defaultSlotSortDesc, Offset: currentCursor.Slot},
		{Column: goqu.I("index"), Desc: defaultSlotSortDesc, Offset: currentCursor.SlotIndex},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBWithdrawalsClColumns.Amount:
		offset = currentCursor.Amount
	}
	// TODO optimize by splitting into filter + sort and applying filter to subqueries
	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	withdrawalsDs = goqu.Dialect("postgres").From(withdrawalsDs.As("bw")).
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		withdrawalsDs = withdrawalsDs.Where(directions)
	}

	queryResult, err := runQueryRows[[]dbResult](ctx, d.readerDb, withdrawalsDs)
	if err != nil {
		return nil, nil, err
	}

	responseData := make([]t.VDBWithdrawalsClTableRow, 0, len(queryResult))
	for _, r := range queryResult {
		row := t.VDBWithdrawalsClTableRow{
			Index:                 r.ValidatorIndex,
			Amount:                r.Amount.Mul(decimal.NewFromInt(1e9)),
			Status:                r.Status,
			Type:                  r.Type,
			Slot:                  r.Slot,
			SlotIndex:             r.SlotIndex,
			WithdrawalCredentials: t.Hash(hexutil.Encode(r.WithdrawalCredentials)),
			PublicKey:             t.PubKey(hexutil.Encode(r.Pubkey)),
		}

		row.GroupId = t.DefaultGroupId
		if r.GroupId.Valid && !dashboardId.AggregateGroups {
			row.GroupId = uint64(r.GroupId.Int64)
		}
		// some data integrity checks TODO add more
		switch r.Status {
		case "queued":
			if r.SlotProcessed.Valid || r.RejectReason.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for queued withdrawal")
			}
			if !r.SlotQueued.Valid {
				return nil, nil, fmt.Errorf("slot_queued is not set for queued withdrawal")
			}
		case "completed":
			if !r.SlotQueued.Valid && !r.SlotProcessed.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for completed withdrawal")
			}
		case "rejected":
			if r.SlotQueued.Valid {
				return nil, nil, fmt.Errorf("unexpected field(s) set for rejected withdrawal")
			}
		default:
			return nil, nil, fmt.Errorf("unknown withdrawal status: %s", r.Status)
		}

		if r.SlotProcessed.Valid {
			slot := uint64(r.SlotProcessed.Int64)
			row.SlotProcessed = slot
		} else { //nolint: staticcheck
			// TODO estimation
		}
		if r.SlotQueued.Valid {
			slot := uint64(r.SlotQueued.Int64)
			row.SlotQueued = &slot
		}
		if r.RejectReason.Valid {
			str := mapWithdrawalRejectReasonDbToApi(r.RejectReason.String)
			if str != "" {
				row.RejectReason = &str // BEDS-1399
			}
		}
		if r.Recipient != nil {
			// TODO add ens + contract info
			row.Recipient = &t.Address{Hash: t.Hash(hexutil.Encode(r.Recipient))}
		}

		responseData = append(responseData, row)
	}

	moreDataFlag := len(responseData) > int(limit)

	// Remove the last entry from data as it is only required for the check
	if moreDataFlag {
		responseData = responseData[:len(responseData)-1]
		queryResult = queryResult[:len(queryResult)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
	}

	// Find the next withdrawal if we are currently at the first page
	// If we have a prev_cursor but not enough data it means the next data is missing
	if !currentCursor.IsValid() || (currentCursor.IsReverse() && len(responseData) < int(limit)) {
		validatorsDs := goqu.Dialect("postgres").
			Select("validatorindex").
			From(goqu.T("validators").As("v"))

		if dashboardId.Validators != nil {
			validatorsDs = validatorsDs.
				SelectAppend(
					goqu.V(t.DefaultGroupId).As("group_id"),
				).
				Where(
					goqu.L("validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
				)
		} else {
			validatorsDs = validatorsDs.
				SelectAppend(
					goqu.I("uvdv.group_id"),
				).
				InnerJoin(
					goqu.T("users_val_dashboards_validators").As("uvdv"),
					goqu.On(
						goqu.I("v.validatorindex").Eq(goqu.I("uvdv.validator_index")),
					),
				).
				Where(goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id))
		}
		// TODO add more filters
		if isValidSearchIndexOrSlot {
			validatorsDs = validatorsDs.Where(
				goqu.I("v.validatorindex").Eq(search),
			)
		}
		validatorGroups, err := runQueryRows[[]validatorGroup](ctx, d.readerDb, validatorsDs)
		if err != nil {
			return nil, nil, err
		}

		nextData, err := d.getNextWithdrawalRow(validatorGroups)
		if err != nil {
			return nil, nil, err
		}
		if nextData != nil {
			// Complete the next data TODO
			// TODO integrate label/ens data for "next" row
			// nextData.Recipient.Ens = addressEns[string(nextData.Recipient.Hash)]
		} else {
			// If there is no next data, add a missing estimate row
			nextData = &t.VDBWithdrawalsClTableRow{
				IsMissingEstimate: true,
			}
		}
		responseData = append([]t.VDBWithdrawalsClTableRow{*nextData}, responseData...)

		// Flag if above limit
		moreDataFlag = moreDataFlag || len(responseData) > int(limit)
		if !moreDataFlag && !currentCursor.IsValid() {
			// No paging required
			return responseData, &t.Paging{}, nil
		}

		// Remove the last entry from data as it is only required for the check
		if moreDataFlag {
			responseData = responseData[:len(responseData)-1]
			queryResult = queryResult[:len(queryResult)-1]
		}
	}

	paging, err := utils.GetPagingFromData(queryResult, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, paging, nil
}

func getWithdrawalsBridgeDs(dashboardId t.VDBId, search string, isValidSearchWithdrawalAddress, isValidSearchIndexOrSlot, isValidSearchGroup, isValidSearchPublicKey bool) (*goqu.SelectDataset, error) {
	var searches []exp.Expression
	withdrawalsBridgeDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_withdrawals").As("w")).
		Select(
			goqu.I("w.block_slot").As("slot_processed"),
			goqu.I("w.validatorindex"),
			goqu.I("v.pubkey"),
			goqu.I("w.address").As("recipient"),
			goqu.I("v.withdrawalcredentials").As("withdrawal_credentials"),
			goqu.I("w.amount"),
			goqu.Cast(goqu.V(nil), "DECIMAL").As("slot_queued"),
			goqu.V(nil).As("reject_reason"),
			goqu.V("completed").As("status"),
			goqu.V("auto").As("type"),
			// cursor
			goqu.I("w.block_slot").As("slot"),
			goqu.I("w.withdrawalindex").As("index"),
		).
		InnerJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("w.block_root").Eq(goqu.I("b.blockroot")),
				goqu.I("b.status").Eq("1"),
			),
		).
		InnerJoin(
			goqu.T("validators").As("v"),
			goqu.On(
				goqu.I("w.validatorindex").Eq(goqu.I("v.validatorindex")),
			),
		)

	if dashboardId.Validators != nil {
		withdrawalsBridgeDs = withdrawalsBridgeDs.Where(
			goqu.L("w.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
		)
	} else {
		withdrawalsBridgeDs = withdrawalsBridgeDs.
			SelectAppend(
				goqu.I("uvdv.group_id"),
			).
			InnerJoin(
				goqu.T("users_val_dashboards_validators").As("uvdv"),
				goqu.On(
					goqu.I("w.validatorindex").Eq(goqu.I("uvdv.validator_index")),
				),
			).
			Where(goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id))

		if isValidSearchGroup {
			withdrawalsBridgeDs = withdrawalsBridgeDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
					goqu.I("uvdv.dashboard_id").Eq(goqu.I("uvdg.dashboard_id")),
					goqu.I("uvdv.group_id").Eq(goqu.I("uvdg.id")),
				))
			s := goqu.L("LOWER(?)", goqu.I("uvdg.name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1) + "%")
			searches = append(searches, s)
		}
	}

	if isValidSearchWithdrawalAddress {
		address, err := hexutil.Decode(search)
		if err != nil {
			return nil, err
		}
		searches = append(searches, goqu.I("address").Eq(address))
	}
	if isValidSearchPublicKey {
		pubkey, err := hexutil.Decode(search)
		if err != nil {
			return nil, err
		}
		withdrawalsBridgeDs = withdrawalsBridgeDs.
			InnerJoin(
				goqu.T("validators").As("v"),
				goqu.On(
					goqu.I("v.index").Eq("w.validatorindex"),
				),
			)
		searches = append(searches, goqu.I("v.pubkey").Eq(pubkey))
	}
	if isValidSearchIndexOrSlot {
		searches = append(searches, goqu.I("block_slot").Eq(search), goqu.I("validatorindex").Eq(search))
	}

	if len(searches) > 0 {
		withdrawalsBridgeDs = withdrawalsBridgeDs.Where(goqu.Or(searches...))
	}

	return withdrawalsBridgeDs, nil
}

func getWithdrawalRequestsDs(dashboardId t.VDBId, search string, isValidSearchWithdrawalAddress, isValidSearchIndexOrSlot, isValidSearchGroup, isValidSearchPublicKey bool) (*goqu.SelectDataset, error) {
	var searches []exp.Expression
	withdrawalRequestsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_withdrawal_requests_v2").As("wr")).
		Select(
			goqu.I("wr.slot_processed"),
			goqu.I("v.validatorindex"),
			goqu.I("v.pubkey"),
			goqu.L("substring(withdrawalcredentials from 13 for 20)").As("recipient"),
			goqu.I("v.withdrawalcredentials").As("withdrawal_credentials"),
			goqu.I("wr.amount"),
			goqu.I("wr.slot_queued"),
			goqu.I("wr.reject_reason"),
			goqu.I("wr.status"),
			goqu.V("manual").As("type"),
			// cursor
			enums.VDBWithdrawalsClColumns.Slot.ToExpr().As("slot"),
			goqu.COALESCE(goqu.I("index_queued"), goqu.I("index_processed")).As("index"),
		).
		InnerJoin(
			goqu.T("validators").As("v"),
			goqu.On(
				goqu.I("wr.validator_pubkey").Eq(goqu.I("v.pubkey")),
			),
		)

	if dashboardId.Validators != nil {
		withdrawalRequestsDs = withdrawalRequestsDs.Where(
			goqu.L("v.validatorindex = ANY(?)", pq.Array(dashboardId.Validators)),
		)
	} else {
		withdrawalRequestsDs = withdrawalRequestsDs.
			SelectAppend(
				goqu.I("uvdv.group_id"),
			).
			InnerJoin(
				goqu.T("users_val_dashboards_validators").As("uvdv"),
				goqu.On(
					goqu.I("v.validatorindex").Eq(goqu.I("uvdv.validator_index")),
				),
			).
			Where(goqu.I("uvdv.dashboard_id").Eq(dashboardId.Id))

		if isValidSearchGroup {
			withdrawalRequestsDs = withdrawalRequestsDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
					goqu.I("uvdv.dashboard_id").Eq(goqu.I("uvdg.dashboard_id")),
					goqu.I("uvdv.group_id").Eq(goqu.I("uvdg.id")),
				))
			s := goqu.L("LOWER(?)", goqu.I("uvdg.name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1) + "%")
			searches = append(searches, s)
		}
	}

	if isValidSearchWithdrawalAddress {
		address, err := hexutil.Decode(search)
		if err != nil {
			return nil, err
		}
		searches = append(searches, goqu.I("withdrawalcredentials").Eq(address))
	}
	if isValidSearchPublicKey {
		pubkey, err := hexutil.Decode(search)
		if err != nil {
			return nil, err
		}
		searches = append(searches, goqu.I("v.pubkey").Eq(pubkey))
	}
	if isValidSearchIndexOrSlot {
		searches = append(searches, goqu.I("slot_processed").Eq(search), goqu.I("validatorindex").Eq(search))
	}
	if len(searches) > 0 {
		withdrawalRequestsDs = withdrawalRequestsDs.Where(goqu.Or(searches...))
	}

	return withdrawalRequestsDs, nil
}

type validatorGroup struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	GroupId        uint64 `db:"group_id"`
}

func mapWithdrawalRejectReasonDbToApi(dbReason string) string {
	switch dbReason {
	case "pending_partial_withdrawals_limit_reached":
		return "full_queue"
	case "validator_not_found":
		return "unknown_pubkey"
	case "withdrawal_credentials_invalid":
		return "no_execution_withdrawal_credentials"
	case "address_mismatch":
		return "address_mismatch"
	case "validator_not_active":
		return "inactive"
	case "validator_has_submitted_exit":
		return "exiting"
	case "validator_not_active_long_enough":
		return "too_young"
	case "pending_withdrawals_in_queue":
		return "pending_withdrawals"
	case "insufficient_effective_balance":
		return "insufficient_effective_balance"
	case "insufficient_excess_balance":
		return "excess_balance"
	case "no_compounding_withdrawal_credentials":
		return "not_compounding"

	default:
		log.Warnf("unknown error: %s", dbReason)
		return ""
	}
}

// TODO implement changes
// returns information about the next *automatic* withdrawal, if applicable (=skimming)
// 0x00 creds (genesis): never
// 0x01 creds (capella): if balance > 32 EB
// 0x02 creds (electra): if balance > 2048 EB
func (d *DataAccessService) getNextWithdrawalRow(queryValidators []validatorGroup) (*t.VDBWithdrawalsClTableRow, error) {
	if len(queryValidators) == 0 {
		return nil, nil
	}

	stats := cache.LatestStats.Get()
	if stats == nil || stats.LatestValidatorWithdrawalIndex == nil {
		return nil, errors.New("stats not available")
	}

	// Get the current validator state
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	epoch := cache.LatestEpoch.Get()

	// find subscribed validators that are active and have valid withdrawal credentials
	// order by validator index to ensure that "last withdrawal" cursor handling works
	sort.Slice(queryValidators, func(i, j int) bool {
		return queryValidators[i].ValidatorIndex < queryValidators[j].ValidatorIndex
	})

	latestFinalized := cache.LatestFinalizedEpoch.Get()

	var nextValidator *validatorGroup
	for _, validator := range queryValidators {
		metadata := validatorMapping.ValidatorMetadata[validator.ValidatorIndex]

		if !utils.IsValidWithdrawalCredentialsAddress(fmt.Sprintf("%x", metadata.WithdrawalCredentials)) {
			// Validator cannot withdraw because of invalid withdrawal credentials
			continue
		}
		if !metadata.ActivationEpoch.Valid || metadata.ActivationEpoch.Int64 > int64(epoch) {
			// Validator is not active yet
			continue
		}
		if metadata.ExitEpoch.Valid && metadata.ExitEpoch.Int64 <= int64(epoch) {
			// Validator has already exited
			continue
		}

		withdrawable := metadata.Balance > 0 && metadata.WithdrawableEpoch.Valid && metadata.WithdrawableEpoch.Int64 <= int64(epoch)
		skimmable := (metadata.EffectiveBalance == utils.GetMaxEffectiveBalanceByWithdrawalCredentials(metadata.WithdrawalCredentials) && metadata.Balance > utils.GetMaxEffectiveBalanceByWithdrawalCredentials(metadata.WithdrawalCredentials))
		latestUpdate := nextValidator == nil || validator.ValidatorIndex > *stats.LatestValidatorWithdrawalIndex
		if (withdrawable || skimmable) && latestUpdate {
			distance, err := d.getWithdrawableCountFromCursor(validator.ValidatorIndex, *stats.LatestValidatorWithdrawalIndex)
			if err != nil {
				return nil, err
			}
			// TODO this is a wrong estimate post-pectra

			timeToWithdrawal := d.getTimeToNextWithdrawal(distance)

			// it normally takes two epochs to finalize
			if !timeToWithdrawal.Before(utils.EpochToTime(epoch + (epoch - latestFinalized))) {
				// this validator has a next withdrawal
				nextValidatorInt := validator
				nextValidator = &nextValidatorInt
			}

			if nextValidator != nil && nextValidator.ValidatorIndex > *stats.LatestValidatorWithdrawalIndex {
				// the first validator after the cursor has to be the next validator
				break
			}
		}
	}

	if nextValidator == nil {
		return nil, nil
	}

	nextValidatorData := validatorMapping.ValidatorMetadata[nextValidator.ValidatorIndex]

	lastWithdrawnEpochs, err := db.GetLastWithdrawalEpoch([]t.VDBValidator{nextValidator.ValidatorIndex})
	if err != nil {
		return nil, err
	}
	lastWithdrawnEpoch := lastWithdrawnEpochs[nextValidator.ValidatorIndex]

	nextDistance, err := d.getWithdrawableCountFromCursor(nextValidator.ValidatorIndex, *stats.LatestValidatorWithdrawalIndex)
	if err != nil {
		return nil, err
	}
	nextTimeToWithdrawal := d.getTimeToNextWithdrawal(nextDistance)
	nextWithdrawalSlot := utils.TimeToSlot(uint64(nextTimeToWithdrawal.Unix()))

	address, err := utils.GetAddressOfWithdrawalCredentials(nextValidatorData.WithdrawalCredentials)
	if err != nil {
		return nil, err
	}

	var withdrawalAmount uint64
	if lastWithdrawnEpoch != epoch && nextValidatorData.Balance > utils.GetMaxEffectiveBalanceByWithdrawalCredentials(nextValidatorData.WithdrawalCredentials) {
		withdrawalAmount = nextValidatorData.Balance
		if !(nextValidatorData.WithdrawableEpoch.Valid && nextValidatorData.WithdrawableEpoch.Int64 <= int64(epoch)) {
			// partial withdrawal
			withdrawalAmount -= utils.GetMaxEffectiveBalanceByWithdrawalCredentials(nextValidatorData.WithdrawalCredentials)
		}
	}

	ens_name, err := db.GetEnsNameForAddress(*address, utils.SlotToTime(nextWithdrawalSlot))
	if err != sql.ErrNoRows {
		return nil, err
	}

	contractStatusReq := []db.ContractInteractionAtRequest{{
		Address: fmt.Sprintf("%x", address),
		Block:   -1,
	}}
	contractStatus, err := d.bigtable.GetAddressContractInteractionsAt(contractStatusReq)
	if err != nil {
		return nil, err
	}

	nextData := &t.VDBWithdrawalsClTableRow{
		SlotProcessed: nextWithdrawalSlot,
		Index:         nextValidator.ValidatorIndex,
		GroupId:       nextValidator.GroupId,
		Recipient: &t.Address{
			Hash:       t.Hash(address.String()),
			Ens:        ens_name,
			IsContract: contractStatus[0] == types.CONTRACT_CREATION || contractStatus[0] == types.CONTRACT_PRESENT,
		},
		Amount: utils.GWeiToWei(big.NewInt(int64(withdrawalAmount))),
	}

	return nextData, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalElWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalExecutionWithdrawalsData, error) {
	result := &t.VDBTotalExecutionWithdrawalsData{
		TotalAmount: decimal.NewFromBigInt(big.NewInt(0), 0),
	}

	// Analyze the search term
	validatorSearch, err := d.getValidatorSearch(search)
	if err != nil {
		return nil, err
	}
	if validatorSearch == nil {
		// No validators found
		return result, nil
	}

	withdrawalsDs := goqu.Dialect("postgres").
		From(goqu.T("eth1_withdrawal_requests").As("w")).
		Select(
			goqu.SUM(goqu.I("w.amount")).As("acc_withdrawals_amount"),
		)

	sum, err := runQuery[int64](ctx, d.readerDb, withdrawalsDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error getting total el withdrawals for validators: %w", err)
	}

	result.TotalAmount = utils.GWeiToWei(big.NewInt(sum))
	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalClWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalConsensusWithdrawalsData, error) {
	result := &t.VDBTotalConsensusWithdrawalsData{
		TotalAmount: decimal.NewFromBigInt(big.NewInt(0), 0),
	}

	// Analyze the search term
	validatorSearch, err := d.getValidatorSearch(search)
	if err != nil {
		return nil, err
	}
	if validatorSearch == nil {
		// No validators found
		return result, nil
	}

	type queryResult struct {
		ValidatorIndex t.VDBValidator `db:"validator_index"`
		Epoch          uint64         `db:"epoch_end"`
		Amount         int64          `db:"acc_withdrawals_amount"`
	}

	withdrawalsDs := goqu.Dialect("postgres").
		From(goqu.L("validator_dashboard_data_rolling_total FINAL")).
		Select(
			goqu.I("validator_index"),
			goqu.SUM(goqu.I("withdrawals_amount")).As("acc_withdrawals_amount"),
			goqu.MAX(goqu.I("epoch_end")).As("epoch_end"),
		).
		GroupBy(goqu.I("validator_index"))

	if dashboardId.Validators == nil {
		withdrawalsDs = withdrawalsDs.
			With("validators", goqu.L("(SELECT validator_index FROM users_val_dashboards_validators WHERE (dashboard_id = ?))", dashboardId.Id)).
			InnerJoin(
				goqu.T("validators").As("v"), goqu.On(
					goqu.I("validator_dashboard_data_rolling_total.validator_index").Eq(goqu.I("v.validator_index")),
				),
			).
			Where(
				goqu.I("validator_index").In(goqu.L("SELECT validator_index FROM validators")),
			)
	} else {
		withdrawalsDs = withdrawalsDs.
			Where(goqu.I("validator_index").In(pq.Array(dashboardId.Validators)))
	}

	res, err := runQueryRows[[]queryResult](ctx, d.clickhouseReader, withdrawalsDs)
	if err != nil {
		return nil, fmt.Errorf("error getting total cl withdrawals for validators: %+v: %w", dashboardId, err)
	}

	if len(res) == 0 {
		// No validators to search for
		return result, nil
	}

	var totalAmount int64
	var validators []t.VDBValidator
	lastEpoch := res[0].Epoch
	lastSlot := (lastEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

	for _, res := range res {
		// Calculate the total amount of withdrawals
		totalAmount += res.Amount

		// Calculate the current validators
		validators = append(validators, res.ValidatorIndex)
	}

	latestWithdrawalsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_withdrawals").As("w")).
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("w.amount")), 0),
		).
		InnerJoin(
			goqu.T("blocks").As("b"), goqu.On(
				goqu.I("w.block_slot").Eq(goqu.I("b.slot")),
				goqu.I("w.block_root").Eq(goqu.I("b.blockroot")),
				goqu.I("b.status").Eq("1"),
			),
		).
		Where(
			goqu.I("w.block_slot").Gt(lastSlot),
			goqu.L("w.validatorindex = ANY(?)", pq.Array(validators)),
		)

	latestWithdrawalsAmount, err := runQuery[int64](ctx, d.readerDb, latestWithdrawalsDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error getting latest cl withdrawals for validators: %+v: %w", dashboardId, err)
	}

	totalAmount += latestWithdrawalsAmount
	result.TotalAmount = utils.GWeiToWei(big.NewInt(totalAmount))
	return result, nil
}

func (d *DataAccessService) getValidatorSearch(search string) ([]t.VDBValidator, error) {
	validatorSearch := make([]t.VDBValidator, 0)

	if search != "" {
		if strings.HasPrefix(search, "0x") && (utils.IsHash(search) || utils.IsEth1Address(search)) {
			search = strings.ToLower(search)

			validatorMapping, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				return nil, err
			}

			if utils.IsHash(search) {
				if index, ok := validatorMapping.ValidatorIndices[search]; ok {
					validatorSearch = append(validatorSearch, index)
				} else {
					// No validator index for pubkey found, return empty results
					return nil, nil
				}
			} else {
				// Get the withdrawal credentials of the address
				address, err := hexutil.Decode(search)
				if err != nil {
					return nil, fmt.Errorf("failed to decode search term %s as address: %w", search, err)
				}
				withdrawalCredentials := utils.GetWithdrawalCredentialsOfAddress(common.BytesToAddress(address))

				for index, metadata := range validatorMapping.ValidatorMetadata {
					if bytes.Equal(withdrawalCredentials, metadata.WithdrawalCredentials) {
						validatorSearch = append(validatorSearch, t.VDBValidator(index))
					}
				}

				if len(validatorSearch) == 0 {
					// No validator index for withdrawal credentials found, return empty results
					return nil, nil
				}
			}
		} else if index, err := strconv.ParseUint(search, 10, 64); err == nil {
			validatorSearch = append(validatorSearch, index)
		} else {
			// No allowed search term found, return empty results
			return nil, nil
		}
	}

	return validatorSearch, nil
}
