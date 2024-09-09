package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardElDeposits(ctx context.Context, dashboardId t.VDBId, cursor string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	var err error
	currentDirection := enums.DESC // TODO: expose over parameter
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

	query := `
			SELECT
				ed.publickey,
				ed.block_number,
				ed.log_index,
				ed.from_address,
				ed.msg_sender,
				ed.tx_hash,
				ed.withdrawal_credentials,
				ed.amount,
				ed.valid_signature,
				ed.block_ts
		`

	var filter interface{}
	if dashboardId.Validators != nil {
		query += `
			FROM
				eth1_deposits ed
			WHERE
				ed.publickey = ANY ($1)`
		filter = byteaArray
	} else {
		query += `
			, cedl.group_id
			FROM
				cached_eth1_deposits_lookup cedl
			INNER JOIN eth1_deposits ed ON ed.block_number = cedl.block_number AND ed.log_index = cedl.log_index
			WHERE
				cedl.dashboard_id = $1`
		filter = dashboardId.Id
	}

	params := []interface{}{filter}
	filterFragment := ` ORDER BY ed.block_number DESC, ed.log_index DESC`
	if currentCursor.IsValid() {
		filterFragment = ` AND (ed.block_number < $2 or (ed.block_number = $2 and ed.log_index < $3)) ` + filterFragment
		params = append(params, currentCursor.BlockNumber, currentCursor.LogIndex)
	}

	if currentDirection == enums.ASC && !currentCursor.IsReverse() || currentDirection == enums.DESC && currentCursor.IsReverse() {
		filterFragment = strings.Replace(strings.Replace(filterFragment, "<", ">", -1), "DESC", "ASC", -1)
	}

	if dashboardId.Validators == nil {
		filterFragment = strings.Replace(filterFragment, "ed.", "cedl.", -1)
	}

	params = append(params, limit+1)
	filterFragment += fmt.Sprintf(" LIMIT $%d", len(params))

	err = db.AlloyReader.SelectContext(ctx, &data, query+filterFragment, params...)

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
	for i, row := range data {
		responseData[i] = t.VDBExecutionDepositsTableRow{
			PublicKey:            t.PubKey(pubkeys[i]),
			Block:                uint64(row.BlockNumber),
			Timestamp:            row.Timestamp.Unix(),
			From:                 t.Address{Hash: t.Hash(hexutil.Encode(row.From))},
			TxHash:               t.Hash(hexutil.Encode(row.TxHash)),
			WithdrawalCredential: t.Hash(hexutil.Encode(row.WithdrawalCredentials)),
			Amount:               utils.GWeiToWei(big.NewInt(row.Amount)),
			Valid:                row.Valid,
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
		} else {
			responseData[i].Depositor = responseData[i].From
		}
		if v, ok := mapping.ValidatorIndices[pubkeys[i]]; ok {
			responseData[i].Index = &v
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
	var err error
	currentDirection := enums.DESC // TODO: expose over parameter
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
	var data []struct {
		GroupId              sql.NullInt64 `db:"group_id"`
		PublicKey            []byte        `db:"publickey"`
		Slot                 int64         `db:"block_slot"`
		SlotIndex            int64         `db:"block_index"`
		WithdrawalCredential []byte        `db:"withdrawalcredentials"`
		Amount               int64         `db:"amount"`
		Signature            []byte        `db:"signature"`
	}

	query := `
			SELECT
				bd.publickey,
				bd.block_slot,
				bd.block_index,
				bd.amount,
				bd.signature,
				bd.withdrawalcredentials
		`

	var filter interface{}
	if dashboardId.Validators != nil {
		query += `
			FROM
				blocks_deposits bd
			WHERE
				bd.publickey = ANY ($1)`
		filter = byteaArray
	} else {
		query += `
			, cbdl.group_id
			FROM
				cached_blocks_deposits_lookup cbdl
				INNER JOIN blocks_deposits bd ON bd.block_slot = cbdl.block_slot
					AND bd.block_index = cbdl.block_index
			WHERE
				cbdl.dashboard_id = $1`
		filter = dashboardId.Id
	}

	params := []interface{}{filter}
	filterFragment := ` ORDER BY bd.block_slot DESC, bd.block_index DESC`
	if currentCursor.IsValid() {
		filterFragment = ` AND (bd.block_slot < $2 or (bd.block_slot = $2 and bd.block_index < $3)) ` + filterFragment
		params = append(params, currentCursor.Slot, currentCursor.SlotIndex)
	}

	if currentDirection == enums.ASC && !currentCursor.IsReverse() || currentDirection == enums.DESC && currentCursor.IsReverse() {
		filterFragment = strings.Replace(strings.Replace(filterFragment, "<", ">", -1), "DESC", "ASC", -1)
	}

	if dashboardId.Validators == nil {
		filterFragment = strings.Replace(filterFragment, "bd.", "cbdl.", -1)
	}

	params = append(params, limit+1)
	filterFragment += fmt.Sprintf(" LIMIT $%d", len(params))

	err = db.AlloyReader.SelectContext(ctx, &data, query+filterFragment, params...)

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

	// Resolve validator indices to pubkeys
	byteaArray, err := d.getValidatorPubkeys(dashboardId)
	if err != nil {
		return nil, err
	}

	query := `
			SELECT
				COALESCE(SUM(amount), 0)
		`

	var filter interface{}
	if dashboardId.Validators != nil {
		query += `
			FROM
				eth1_deposits
			WHERE
				publickey = ANY ($1)`
		filter = byteaArray
	} else {
		query += `
			FROM
				cached_eth1_deposits_lookup
			WHERE
				dashboard_id = $1
			GROUP BY
				dashboard_id`
		filter = dashboardId.Id
	}

	var data int64
	err = db.AlloyReader.GetContext(ctx, &data, query, filter)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	responseData.TotalAmount = utils.GWeiToWei(big.NewInt(data))

	return &responseData, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalClDeposits(ctx context.Context, dashboardId t.VDBId) (*t.VDBTotalConsensusDepositsData, error) {
	responseData := t.VDBTotalConsensusDepositsData{
		TotalAmount: decimal.Zero,
	}

	// Resolve validator indices to pubkeys
	byteaArray, err := d.getValidatorPubkeys(dashboardId)
	if err != nil {
		return nil, err
	}

	query := `
			SELECT
				COALESCE(SUM(amount), 0)
		`

	var filter interface{}
	if dashboardId.Validators != nil {
		query += `
			FROM
				blocks_deposits
			WHERE
				publickey = ANY ($1)`
		filter = byteaArray
	} else {
		query += `
			FROM
				cached_blocks_deposits_lookup
			WHERE
				dashboard_id = $1
			GROUP BY
				dashboard_id`
		filter = dashboardId.Id
	}

	var data int64
	err = db.AlloyReader.GetContext(ctx, &data, query, filter)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	responseData.TotalAmount = utils.GWeiToWei(big.NewInt(data))

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
