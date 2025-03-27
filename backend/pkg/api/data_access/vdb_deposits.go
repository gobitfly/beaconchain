package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/common/hexutil"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// TODO: add default sorting
	var err error
	var currentCursor t.ELDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ELDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ELDepositsCursor: %w", err)
		}
	}

	// Resolve validator indices to pubkeys
	byteaArray, err := d.getValidatorPubkeys(dashboardId)
	if err != nil {
		return nil, nil, err
	}

	// Custom type for log_index
	var data []struct {
		GroupId               sql.NullInt64 `db:"group_id"`
		PublicKey             []byte        `db:"publickey"`
		BlockNumber           int64         `db:"block_number"`
		LogIndex              int64         `db:"log_index"`
		Timestamp             time.Time     `db:"block_ts"`
		From                  []byte        `db:"from_address"`
		Depositor             []byte        `db:"msg_sender"`
		TxHash                []byte        `db:"tx_hash"`
		WithdrawalCredentials []byte        `db:"withdrawal_credentials"`
		Amount                int64         `db:"amount"`
		Valid                 bool          `db:"valid_signature"`
	}

	depositsDs := goqu.Dialect("postgres").
		From(goqu.T("eth1_deposits").As("ed")).
		Select(
			goqu.I("ed.publickey"),
			goqu.I("ed.block_number"),
			goqu.L("COALESCE(ed.log_index, 0) AS log_index"),
			goqu.I("ed.from_address"),
			goqu.I("ed.msg_sender"),
			goqu.I("ed.tx_hash"),
			goqu.I("ed.withdrawal_credentials"),
			goqu.I("ed.amount"),
			goqu.I("ed.valid_signature"),
			goqu.I("ed.block_ts"),
		)

	if dashboardId.Validators != nil {
		depositsDs = depositsDs.
			Where(goqu.L("ed.publickey = ANY(?)", byteaArray))
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
	}

	defaultColumns := []t.SortColumn{
		{Column: goqu.I("ed.block_number"), Desc: true, Offset: currentCursor.BlockNumber},
		{Column: goqu.I("ed.log_index"), Desc: true, Offset: currentCursor.LogIndex},
	}
	order, directions, err := applySortAndPagination(defaultColumns, defaultColumns[0], currentCursor.GenericCursor)
	if err != nil {
		return nil, nil, err
	}

	depositsDs = depositsDs.
		Order(order...).
		Limit(uint(limit + 1))
	if directions != nil {
		depositsDs = depositsDs.Where(directions)
	}

	query, params, err := depositsDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare SQL query: %w", err)
	}

	err = db.AlloyReader.SelectContext(ctx, &data, query, params...)

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
			Valid:                row.Valid,
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

func (d *DataAccessService) GetValidatorDashboardClDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	// TODO: add default sorting
	var err error
	var currentCursor t.CLDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.CLDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLDepositsCursor: %w", err)
		}
	}

	// Resolve validator indices to pubkeys
	byteaArray, err := d.getValidatorPubkeys(dashboardId)
	if err != nil {
		return nil, nil, err
	}

	// Custom type for block_index
	type dbResult struct {
		GroupId              sql.NullInt64 `db:"group_id"`
		PublicKey            []byte        `db:"publickey"`
		Slot                 int64         `db:"block_slot"`
		SlotIndex            int64         `db:"block_index"`
		WithdrawalCredential []byte        `db:"withdrawalcredentials"`
		Amount               int64         `db:"amount"`
		Signature            []byte        `db:"signature"`
	}

	depositsBridgeDs := goqu.Dialect("postgres").
		From(goqu.T("blocks_deposits").As("bd")).
		Select(
			goqu.I("bd.publickey"),
			goqu.I("bd.block_slot"),
			goqu.I("bd.block_index"),
			goqu.I("bd.amount"),
			goqu.I("bd.signature"),
			goqu.I("bd.withdrawalcredentials"),
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
			goqu.I("bdr.pubkey"),
			goqu.I("bdr.block_slot"),
			goqu.I("bdr.request_index"),
			goqu.I("bdr.amount"),
			goqu.I("bdr.signature"),
			goqu.I("bdr.withdrawal_credentials"),
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
			postElectra := uint64(currentCursor.Slot)/d.config.ClConfig.SlotsPerEpoch > d.config.ClConfig.ElectraForkEpoch
			if postElectra && currentCursor.Reverse {
				depositsDs = depositRequestsDs
			} else if !postElectra && !currentCursor.Reverse {
				depositsDs = depositsBridgeDs
			}
		}
	}

	defaultColumns := []t.SortColumn{
		{Column: goqu.I("bd.block_slot"), Desc: true, Offset: currentCursor.Slot},
		{Column: goqu.I("bd.block_index"), Desc: true, Offset: currentCursor.SlotIndex},
	}
	order, directions, err := applySortAndPagination(defaultColumns, defaultColumns[0], currentCursor.GenericCursor)
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
			PublicKey:            t.PubKey(pubkeys[i]),
			Index:                indices[i],
			Epoch:                utils.EpochOfSlot(uint64(row.Slot)),
			Slot:                 uint64(row.Slot),
			WithdrawalCredential: t.Hash(hexutil.Encode(row.WithdrawalCredential)),
			Amount:               utils.GWeiToWei(big.NewInt(row.Amount)),
			Signature:            t.Hash(hexutil.Encode(row.Signature)),
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

func (d *DataAccessService) GetValidatorDashboardTotalElDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalExecutionDepositsData, error) {
	responseData := t.VDBTotalExecutionDepositsData{
		TotalAmount: decimal.Zero,
	}

	totalDs := goqu.Dialect("postgres").
		Select(goqu.L("COALESCE(SUM(amount), 0)"))

	if dashboardId.Validators != nil {
		// Resolve validator indices to pubkeys
		byteaArray, err := d.getValidatorPubkeys(dashboardId)
		if err != nil {
			return nil, err
		}
		totalDs = totalDs.
			From(goqu.T("eth1_deposits").As("ed")).
			Where(goqu.L("publickey = ANY(?)", byteaArray))
	} else {
		totalDs = totalDs.
			From(goqu.T("cached_eth1_deposits_lookup")).
			Where(goqu.I("dashboard_id").Eq(dashboardId.Id)).
			GroupBy(goqu.I("dashboard_id"))
	}

	query, params, err := totalDs.Prepared(true).ToSQL()
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

func (d *DataAccessService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error) {
	responseData := t.VDBTotalConsensusDepositsData{
		TotalAmount: decimal.Zero,
	}

	depositsTotalDs := goqu.Dialect("postgres").
		Select(goqu.L("COALESCE(SUM(amount), 0)").As("amount"))
	depositRequestsTotalDs := depositsTotalDs

	if dashboardId.Validators != nil {
		// Resolve validator indices to pubkeys
		byteaArray, err := d.getValidatorPubkeys(dashboardId)
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
			Where(goqu.I("dashboard_id").Eq(dashboardId.Id)).
			GroupBy(goqu.I("dashboard_id"))

		depositRequestsTotalDs = depositRequestsTotalDs.
			From(goqu.T("cached_blocks_deposit_requests_lookup")).
			Where(goqu.I("dashboard_id").Eq(dashboardId.Id)).
			GroupBy(goqu.I("dashboard_id"))
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

func (d *DataAccessService) getValidatorPubkeys(dashboardId t.VDBId) (pq.ByteaArray, error) {
	var byteaArray pq.ByteaArray

	if dashboardId.Validators != nil {
		validatorPubkeys, err := d.services.GetPubkeySliceFromIndexSlice(dashboardId.Validators)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve validator indices to pubkeys: %w", err)
		}

		// Convert pubkeys to bytes for PostgreSQL
		byteaArray = make(pq.ByteaArray, len(validatorPubkeys))
		for i, p := range validatorPubkeys {
			byteaArray[i], _ = hexutil.Decode(p)
		}
	}
	return byteaArray, nil
}
