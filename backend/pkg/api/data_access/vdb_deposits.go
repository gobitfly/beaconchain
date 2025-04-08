package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
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
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBDepositsElColumn], search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	var err error
	var currentCursor t.ELDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ELDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ELDepositsCursor: %w", err)
		}
	}

	// filters
	searchPubkey := t.ReValidatorPublicKeyWithPrefix.MatchString(search)
	searchGroup := len(search) > 0 && !dashboardId.AggregateGroups && t.ReName.MatchString(search)
	searchIndexOrBlock := t.ReInteger.MatchString(search)

	if search != "" && !searchPubkey && !searchGroup && !searchIndexOrBlock {
		return make([]t.VDBExecutionDepositsTableRow, 0), &t.Paging{}, nil
	}

	// Custom type for log_index
	type dbResult struct {
		GroupId               sql.NullInt64 `db:"group_id"`
		PublicKey             []byte        `db:"publickey"`
		Index                 uint64        `db:"validatorindex"`
		BlockNumber           int64         `db:"block_number"`
		LogIndex              int64         `db:"log_index"`
		Timestamp             time.Time     `db:"block_ts"`
		From                  []byte        `db:"from_address"`
		Depositor             []byte        `db:"msg_sender"`
		TxHash                []byte        `db:"tx_hash"`
		WithdrawalCredentials []byte        `db:"withdrawal_credentials"`
		Amount                int64         `db:"amount"`
		Valid                 string        `db:"valid_signature"`
	}

	depositsDs := goqu.Dialect("postgres").
		From(goqu.T("eth1_deposits").As("ed")).
		Select(
			goqu.I("ed.publickey"),
			goqu.I("v.validatorindex"),
			goqu.I("ed.block_number"),
			goqu.L("COALESCE(ed.log_index, 0) AS log_index"),
			goqu.I("ed.from_address"),
			goqu.I("ed.msg_sender"),
			goqu.I("ed.tx_hash"),
			goqu.I("ed.withdrawal_credentials"),
			goqu.I("ed.amount"),
			goqu.I("ed.block_ts"),
			goqu.Case().
				When(goqu.I("ed.valid_signature").Eq(true), "valid").
				When(goqu.Func("EXISTS", goqu.
					From(goqu.T("eth1_deposits").As("ed2")).
					Select("valid_signature").
					Where(
						goqu.I("valid_signature").Eq(true),
						goqu.I("ed2.publickey").Eq(goqu.I("ed.publickey")),
						goqu.Or(
							goqu.I("ed2.block_number").Lt(goqu.I("ed.block_number")),
							goqu.And(
								goqu.I("ed2.block_number").Eq(goqu.I("ed.block_number")),
								goqu.I("ed2.log_index").Lt(goqu.I("ed.log_index")),
							),
						),
					)), "invalid_skipped").
				Else("invalid").
				As("valid_signature"),
		).
		InnerJoin(
			goqu.T("validators").As("v"),
			goqu.On(goqu.I("ed.publickey").Eq(goqu.I("v.pubkey"))),
		)

	searches := []exp.Expression{}
	if dashboardId.Validators != nil {
		// Resolve validator indices to pubkeys
		byteaArray, err := d.getValidatorPubkeys(dashboardId.Validators)
		if err != nil {
			return nil, nil, err
		}
		depositsDs = depositsDs.
			Where(goqu.L("ed.publickey = ANY(?)", byteaArray))
		if searchIndexOrBlock {
			filter, err := d.getPubIndexSearchQryFilter(search)
			if err != nil {
				return nil, nil, err
			}
			searches = append(searches, filter)
		}
		if searchPubkey {
			pubkey, err := hexutil.Decode(search)
			if err != nil {
				return nil, nil, err
			}
			searches = append(searches, goqu.I("ed.publickey").Eq(pubkey))
		}
	} else {
		depositsDs = depositsDs.
			SelectAppend(
				goqu.I("cedl.group_id"),
			).
			InnerJoin(
				goqu.T("cached_eth1_deposits_lookup").As("cedl"),
				goqu.On(
					goqu.I("ed.block_number").Eq(goqu.I("cedl.block_number")),
					goqu.I("ed.log_index").Eq(goqu.I("cedl.log_index")),
				),
			).
			Where(goqu.I("cedl.dashboard_id").Eq(dashboardId.Id))
		if searchIndexOrBlock || searchPubkey {
			var index uint64
			if searchIndexOrBlock {
				idx, err := strconv.Atoi(search)
				if err != nil {
					return nil, nil, err
				}
				index = uint64(idx)
			} else {
				validatorMapping, err := d.services.GetCurrentValidatorMapping()
				if err != nil {
					return nil, nil, err
				}
				var ok bool
				index, ok = validatorMapping.ValidatorIndices[search]
				if !ok && !searchGroup {
					return make([]t.VDBExecutionDepositsTableRow, 0), &t.Paging{}, nil
				}
			}
			// only validators with index in dashboards
			depositsDs = depositsDs.
				InnerJoin(
					goqu.T("users_val_dashboards_validators").As("uvdv"),
					goqu.On(
						goqu.I("cedl.dashboard_id").Eq(goqu.I("uvdv.dashboard_id")),
						goqu.I("cedl.group_id").Eq(goqu.I("uvdv.group_id")),
						goqu.I("v.validatorindex").Eq(goqu.I("uvdv.validator_index")),
					),
				)
			searches = append(searches, goqu.I("uvdv.validator_index").Eq(index))
		}
		if searchGroup {
			depositsDs = depositsDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
					goqu.I("cedl.dashboard_id").Eq(goqu.I("uvdg.dashboard_id")),
					goqu.I("cedl.group_id").Eq(goqu.I("uvdg.id")),
				))
			searches = append(searches,
				goqu.L("LOWER(?)", goqu.I("uvdg.name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1)+"%"),
			)
		}
	}
	if searchIndexOrBlock {
		searches = append(searches, goqu.I("ed.block_number").Eq(search))
	}
	if len(searches) > 0 {
		depositsDs = depositsDs.Where(goqu.Or(searches...))
	}

	defaultColumns := []t.SortColumn{
		{Column: goqu.I("ed.block_number"), Desc: true, Offset: currentCursor.BlockNumber},
		{Column: goqu.I("ed.log_index"), Desc: true, Offset: currentCursor.LogIndex},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBDepositsElColumns.Amount:
		offset = currentCursor.Amount
	}
	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	depositsDs = depositsDs.
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		depositsDs = depositsDs.Where(directions)
	}

	data, err := runQueryRows[[]dbResult](ctx, db.AlloyReader, depositsDs)
	if err != nil {
		return nil, nil, err
	}

	pubkeys := make([]string, len(data))
	for i, row := range data {
		pubkeys[i] = hexutil.Encode(row.PublicKey)
	}

	// need to do it manually because some pubkeys might not be in the database
	mapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current validator mapping: %w", err)
	}

	responseData := make([]t.VDBExecutionDepositsTableRow, len(data))
	addressMapping := make(map[string]*t.Address)
	fromContractStatusRequests := make([]db.ContractInteractionAtRequest, len(data))
	depositorContractStatusRequests := make([]db.ContractInteractionAtRequest, 0, len(data))
	for i, row := range data {
		responseData[i] = t.VDBExecutionDepositsTableRow{
			PublicKey:            t.PubKey(pubkeys[i]),
			Block:                uint64(row.BlockNumber),
			Timestamp:            row.Timestamp.Unix(),
			TxHash:               t.Hash(hexutil.Encode(row.TxHash)),
			WithdrawalCredential: t.Hash(hexutil.Encode(row.WithdrawalCredentials)),
			Amount:               utils.GWeiToWei(big.NewInt(row.Amount)),
			Validity:             row.Valid,
			From:                 t.Address{Hash: t.Hash(hexutil.Encode(row.From))},
		}
		addressMapping[hexutil.Encode(row.From)] = nil
		fromContractStatusRequests[i] = db.ContractInteractionAtRequest{
			Address: fmt.Sprintf("%x", row.From),
			Block:   row.BlockNumber,
			// TODO not entirely correct, would need to determine tx index and itx index of tx. But good enough for now
			TxIdx:    -1,
			TraceIdx: -1,
		}
		if row.GroupId.Valid {
			if dashboardId.AggregateGroups {
				responseData[i].GroupId = t.DefaultGroupId
			} else {
				responseData[i].GroupId = uint64(row.GroupId.Int64)
			}
		} else {
			responseData[i].GroupId = t.DefaultGroupId
		}
		if len(row.Depositor) > 0 {
			responseData[i].Depositor = t.Address{Hash: t.Hash(hexutil.Encode(row.Depositor))}
			addressMapping[hexutil.Encode(row.Depositor)] = nil
			depositorReq := fromContractStatusRequests[i]
			depositorReq.Address = fmt.Sprintf("%x", row.Depositor)
			depositorContractStatusRequests = append(depositorContractStatusRequests, depositorReq)
		} else {
			responseData[i].Depositor = responseData[i].From
		}
		if v, ok := mapping.ValidatorIndices[pubkeys[i]]; ok {
			responseData[i].Index = &v
		}
	}

	// populate address data
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, nil, err
	}
	fromContractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(fromContractStatusRequests)
	if err != nil {
		return nil, nil, err
	}
	depositorContractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(depositorContractStatusRequests)
	if err != nil {
		return nil, nil, err
	}
	var depositorIdx int
	for i := range data {
		responseData[i].From = *addressMapping[string(responseData[i].From.Hash)]
		responseData[i].From.IsContract = fromContractStatuses[i] == types.CONTRACT_CREATION || fromContractStatuses[i] == types.CONTRACT_PRESENT
		responseData[i].Depositor = *addressMapping[string(responseData[i].Depositor.Hash)]
		responseData[i].Depositor.IsContract = responseData[i].From.IsContract
		if responseData[i].Depositor.Hash != responseData[i].From.Hash {
			responseData[i].Depositor.IsContract = depositorContractStatuses[depositorIdx] == types.CONTRACT_CREATION || depositorContractStatuses[depositorIdx] == types.CONTRACT_PRESENT
			depositorIdx += 1
		}
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
		data = data[:len(data)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
		slices.Reverse(data)
	}

	p, err := utils.GetPagingFromData(data, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func (d *DataAccessService) GetValidatorDashboardClDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBDepositsClColumn], search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	var err error
	var currentCursor t.CLDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.CLDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLDepositsCursor: %w", err)
		}
	}

	// Resolve validator indices to pubkeys
	byteaArray, err := d.getValidatorPubkeys(dashboardId.Validators)
	if err != nil {
		return nil, nil, err
	}

	type dbResult struct {
		GroupId              sql.NullInt64   `db:"group_id"`
		PublicKey            []byte          `db:"publickey"`
		SlotProcessed        int64           `db:"block_slot"`
		SlotIndex            int64           `db:"request_index"`
		WithdrawalCredential []byte          `db:"withdrawalcredentials"`
		Amount               decimal.Decimal `db:"amount"`
		Signature            []byte          `db:"signature"`
		// SlotQueued           int64           `db:"slot_queued"` // BEDS-1399
		// Type                 string          `db:"type"`          // BEDS-1399
		// Status               string          `db:"status"`        // BEDS-1399
		// RejectReason         sql.NullString  `db:"reject_reason"` // BEDS-1399
	}

	depositsBridgeDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_deposits").As("bd")).
		Select(
			goqu.I("bd.publickey"),
			goqu.I("bd.block_slot"),
			goqu.I("bd.block_index").As("request_index"),
			goqu.I("bd.amount"),
			goqu.I("bd.signature"),
			goqu.I("bd.withdrawalcredentials"),
			// goqu.V(nil), // TODO BEDS-1399
			// goqu.V("manual").As("type"), // TODO BEDS-1399
			// goqu.I("status"), // TODO BEDS-1399 (or maybe check some other tables?)
			// goqu.L("reject_reason"), // TODO BEDS-1399
		).
		InnerJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("bd.block_root").Eq(goqu.I("b.blockroot")),
				goqu.L("b.status = '1'"),
			),
		)

	depositRequestsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_deposit_requests").As("bdr")).
		Select(
			goqu.I("bdr.pubkey").As("publickey"),
			goqu.I("bdr.block_slot"),
			goqu.I("bdr.request_index"),
			goqu.I("bdr.amount"),
			goqu.I("bdr.signature"),
			goqu.I("bdr.withdrawal_credentials").As("withdrawalcredentials"),
			// goqu.I("bdr.slot_queued"), // TODO BEDS-1399
			// goqu.I("bdr.type"), // TODO BEDS-1399
			// goqu.I("bdr.status"), // TODO BEDS-1399
			// goqu.I("bdr.reject_reason"), // TODO BEDS-1399
		).
		InnerJoin(
			goqu.T("blocks").As("b"),
			goqu.On(
				goqu.I("bdr.block_root").Eq(goqu.I("b.blockroot")),
				goqu.L("b.status = '1'"),
			),
		)

	if dashboardId.Validators != nil {
		depositsBridgeDs = depositsBridgeDs.
			Where(goqu.L("bd.publickey = ANY(?)", byteaArray))
		depositRequestsDs = depositRequestsDs.
			Where(goqu.L("bdr.pubkey = ANY(?)", byteaArray))
	} else {
		depositsBridgeDs = depositsBridgeDs.
			SelectAppend(
				goqu.I("cbdl.group_id"),
			).
			InnerJoin(
				goqu.T("cached_blocks_deposits_lookup").As("cbdl"),
				goqu.On(
					goqu.I("bd.block_slot").Eq(goqu.I("cbdl.block_slot")),
					goqu.I("bd.block_index").Eq(goqu.I("cbdl.block_index")),
				),
			).
			Where(goqu.I("cbdl.dashboard_id").Eq(dashboardId.Id))

		depositRequestsDs = depositRequestsDs.
			SelectAppend(
				goqu.I("cbdrl.group_id"),
			).
			InnerJoin(
				goqu.T("cached_blocks_deposit_requests_lookup").As("cbdrl"),
				goqu.On(
					goqu.I("bdr.block_slot").Eq(goqu.I("cbdrl.block_slot")),
					goqu.I("bdr.request_index").Eq(goqu.I("cbdrl.request_index")),
				),
			).
			Where(goqu.I("cbdrl.dashboard_id").Eq(dashboardId.Id))
	}

	depositsDs := depositsBridgeDs
	if d.config.ClConfig.ElectraForkEpoch < utils.MaxForkEpoch {
		depositsDs = depositsDs.
			UnionAll(depositRequestsDs)
		if currentCursor.IsValid() {
			postElectra := uint64(currentCursor.SlotProcessed)/d.config.ClConfig.SlotsPerEpoch > d.config.ClConfig.ElectraForkEpoch
			if postElectra && currentCursor.Reverse {
				depositsDs = depositRequestsDs
			} else if !postElectra && !currentCursor.Reverse {
				depositsDs = depositsBridgeDs
			}
		}
	}

	defaultSlotSortDesc := true
	if colSort.Column == enums.VDBDepositsClColumns.Slot {
		// this implements a form of multicolumn sort which we don't want to support atm, but for a time-sensitive sort it should be justified
		defaultSlotSortDesc = colSort.Desc
	}
	defaultColumns := []t.SortColumn{
		{Column: goqu.I("bd.block_slot"), Desc: defaultSlotSortDesc, Offset: currentCursor.SlotProcessed},
		{Column: goqu.I("bd.request_index"), Desc: defaultSlotSortDesc, Offset: currentCursor.SlotIndex},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBDepositsClColumns.Amount:
		offset = currentCursor.Amount
	}
	order, directions, err := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	depositsDs = goqu.Dialect("postgres").From(depositsDs.As("bd")).
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		depositsDs = depositsDs.Where(directions)
	}

	data, err := runQueryRows[[]dbResult](ctx, db.AlloyReader, depositsDs)
	if err != nil {
		return nil, nil, err
	}

	pubkeys := make([]string, len(data))
	for i, row := range data {
		pubkeys[i] = hexutil.Encode(row.PublicKey)
	}
	indices, err := d.services.GetIndexSliceFromPubkeySlice(pubkeys)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to recover indices after query: %w", err)
	}

	responseData := make([]t.VDBConsensusDepositsTableRow, len(data))
	for i, row := range data {
		responseData[i] = t.VDBConsensusDepositsTableRow{
			PublicKey: t.PubKey(pubkeys[i]),
			Index:     indices[i],
			// SlotQueued:           0,                         // TODO BEDS-1399
			SlotProcessed:        uint64(row.SlotProcessed),
			WithdrawalCredential: t.Hash(hexutil.Encode(row.WithdrawalCredential)),
			Amount:               utils.GWeiToWei(row.Amount.BigInt()),
			Signature:            t.Hash(hexutil.Encode(row.Signature)),
			// Type:                 row.Type, // TODO
			// Status:               row.Status, // TODO
		}
		responseData[i].GroupId = t.DefaultGroupId
		if row.GroupId.Valid && !dashboardId.AggregateGroups {
			responseData[i].GroupId = uint64(row.GroupId.Int64)
		}
		// TODO BEDS-1399
		/*if row.RejectReason.Valid {
			responseData[i].RejectReason = &row.RejectReason.String
		}*/
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
		data = data[:len(data)-1]
	}

	if currentCursor.IsReverse() {
		// Invert query result so response matches requested direction
		slices.Reverse(responseData)
		slices.Reverse(data)
	}

	p, err := utils.GetPagingFromData(data, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return responseData, p, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalElDeposits(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBTotalExecutionDepositsData, error) {
	responseData := t.VDBTotalExecutionDepositsData{
		TotalAmount: decimal.Zero,
	}

	// filters
	searchPubkey := t.ReValidatorPublicKeyWithPrefix.MatchString(search)
	searchGroup := len(search) > 0 && !dashboardId.AggregateGroups && t.ReName.MatchString(search)
	searchIndexOrBlock := t.ReInteger.MatchString(search)

	if search != "" && !searchPubkey && !searchGroup && !searchIndexOrBlock {
		return &responseData, nil
	}

	totalDs := goqu.Dialect("postgres").
		Select(goqu.L("COALESCE(SUM(ed.amount), 0)"))

	searches := []exp.Expression{}
	if dashboardId.Validators != nil {
		// Resolve validator indices to pubkeys
		byteaArray, err := d.getValidatorPubkeys(dashboardId.Validators)
		if err != nil {
			return nil, err
		}
		totalDs = totalDs.
			From(goqu.T("eth1_deposits").As("ed")).
			Where(goqu.L("publickey = ANY(?)", byteaArray))
		if searchPubkey {
			pubkey, err := hexutil.Decode(search)
			if err != nil {
				return nil, err
			}
			searches = append(searches, goqu.I("ed.publickey").Eq(pubkey))
		}
		if searchIndexOrBlock {
			filter, err := d.getPubIndexSearchQryFilter(search)
			if err != nil {
				return nil, err
			}
			searches = append(searches, filter)
		}
	} else {
		totalDs = totalDs.
			From(goqu.T("cached_eth1_deposits_lookup").As("ed")).
			Where(goqu.I("ed.dashboard_id").Eq(dashboardId.Id))

		if searchGroup {
			totalDs = totalDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As("uvdg"), goqu.On(
					goqu.I("ed.group_id").Eq(goqu.I("uvdg.id")),
					goqu.I("ed.dashboard_id").Eq(goqu.I("uvdg.dashboard_id")),
				))
			searches = append(searches,
				goqu.L("LOWER(?)", goqu.I("uvdg.name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1)+"%"),
			)
		}
		if searchIndexOrBlock || searchPubkey {
			var index uint64
			if searchIndexOrBlock {
				idx, err := strconv.Atoi(search)
				if err != nil {
					return nil, err
				}
				index = uint64(idx)
			} else {
				validatorMapping, err := d.services.GetCurrentValidatorMapping()
				if err != nil {
					return nil, err
				}
				var ok bool
				index, ok = validatorMapping.ValidatorIndices[search]
				if !ok && !searchGroup {
					return &responseData, nil
				}
			}
			totalDs = totalDs.
				InnerJoin(
					goqu.T("eth1_deposits"),
					goqu.On(
						goqu.I("ed.block_number").Eq(goqu.I("eth1_deposits.block_number")),
						goqu.I("ed.log_index").Eq(goqu.I("eth1_deposits.log_index")),
					),
				).
				InnerJoin(
					goqu.T("validators").As("v"),
					goqu.On(goqu.I("eth1_deposits.publickey").Eq(goqu.I("v.pubkey"))),
				).
				InnerJoin(
					goqu.T("users_val_dashboards_validators").As("uvdv"),
					goqu.On(
						goqu.I("ed.dashboard_id").Eq(goqu.I("uvdv.dashboard_id")),
						goqu.I("ed.group_id").Eq(goqu.I("uvdv.group_id")),
						goqu.I("v.validatorindex").Eq(goqu.I("uvdv.validator_index")),
					),
				)
			searches = append(searches, goqu.I("uvdv.validator_index").Eq(index))
		}
	}
	if searchIndexOrBlock {
		searches = append(searches, goqu.I("ed.block_number").Eq(search))
	}
	if len(searches) > 0 {
		totalDs = totalDs.Where(goqu.Or(searches...))
	}

	sum, err := runQuery[int64](ctx, db.AlloyReader, totalDs)
	if err != nil {
		return nil, err
	}

	responseData.TotalAmount = utils.GWeiToWei(big.NewInt(sum))
	return &responseData, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBTotalConsensusDepositsData, error) {
	// TODO add filter
	responseData := t.VDBTotalConsensusDepositsData{
		TotalAmount: decimal.Zero,
	}

	depositsTotalDs := goqu.Dialect("postgres").
		Select(goqu.L("COALESCE(SUM(amount), 0)").As("amount"))
	depositRequestsTotalDs := depositsTotalDs

	if dashboardId.Validators != nil {
		// Resolve validator indices to pubkeys
		byteaArray, err := d.getValidatorPubkeys(dashboardId.Validators)
		if err != nil {
			return nil, err
		}
		depositsTotalDs = depositsTotalDs.
			From(goqu.T("blocks_deposits").As("bd")).
			Where(goqu.L("publickey = ANY(?)", byteaArray)).
			InnerJoin(
				goqu.T("blocks").As("b"),
				goqu.On(
					goqu.I("bd.block_root").Eq(goqu.I("b.blockroot")),
					goqu.L("b.status = '1'"),
				),
			)
		depositRequestsTotalDs = depositRequestsTotalDs.
			From(goqu.T("blocks_deposit_requests").As("bdr")).
			Where(goqu.L("pubkey = ANY(?)", byteaArray)).
			InnerJoin(
				goqu.T("blocks").As("b"),
				goqu.On(
					goqu.I("bdr.block_root").Eq(goqu.I("b.blockroot")),
					goqu.L("b.status = '1'"),
				),
			)
	} else {
		depositsTotalDs = depositsTotalDs.
			From(goqu.T("cached_blocks_deposits_lookup")).
			Where(goqu.I("dashboard_id").Eq(dashboardId.Id))

		depositRequestsTotalDs = depositRequestsTotalDs.
			From(goqu.T("cached_blocks_deposit_requests_lookup")).
			Where(goqu.I("dashboard_id").Eq(dashboardId.Id))
	}

	if d.config.ClConfig.ElectraForkEpoch < utils.MaxForkEpoch {
		depositsTotalDs = goqu.Dialect("postgres").
			Select(goqu.L("COALESCE(SUM(amount), 0)")).
			From(depositsTotalDs.UnionAll(depositRequestsTotalDs))
	}

	query, params, err := depositsTotalDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL query: %w", err)
	}

	var sum int64
	err = db.AlloyReader.GetContext(ctx, &sum, query, params...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	responseData.TotalAmount = utils.GWeiToWei(big.NewInt(sum))

	return &responseData, nil
}

func (d *DataAccessService) getValidatorPubkeys(validators []t.VDBValidator) (pq.ByteaArray, error) {
	var byteaArray pq.ByteaArray

	validatorPubkeys, err := d.services.GetPubkeySliceFromIndexSlice(validators)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve validator indices to pubkeys: %w", err)
	}

	// Convert pubkeys to bytes for PostgreSQL
	byteaArray = make(pq.ByteaArray, len(validatorPubkeys))
	for i, p := range validatorPubkeys {
		byteaArray[i], _ = hexutil.Decode(p)
	}
	return byteaArray, nil
}

func (d *DataAccessService) getPubIndexSearchQryFilter(search string) (exp.Expression, error) {
	index, err := strconv.Atoi(search)
	if err != nil {
		return nil, err
	}
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}
	var pubkey string
	if index < len(validatorMapping.ValidatorPubkeys) {
		pubkey = validatorMapping.ValidatorPubkeys[index]
	}
	return goqu.I("ed.publickey").Eq(pubkey), nil
}
