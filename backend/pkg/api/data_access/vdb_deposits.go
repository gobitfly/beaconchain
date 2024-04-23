package dataaccess

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardElDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBExecutionDepositsTableRow, *t.Paging, error) {
	// WORKING @invis
	return d.dummy.GetValidatorDashboardElDeposits(dashboardId, cursor, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardClDeposits(dashboardId t.VDBId, cursor string, search string, limit uint64) ([]t.VDBConsensusDepositsTableRow, *t.Paging, error) {
	var err error
	currentDirection := enums.DESC // TODO: expose over parameter
	var currentCursor t.CLDepositsCursor

	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.CLDepositsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as CLDepositsCursor: %w", err)
		}
	}

	var byteaArray pq.ByteaArray

	// Resolve validator indices to pubkeys
	if dashboardId.Validators != nil {
		validatorsArray := make([]uint64, len(dashboardId.Validators))
		for i, v := range dashboardId.Validators {
			validatorsArray[i] = v.Index
		}
		validatorPubkeys, err := d.services.GetPubkeysOfValidatorIndexSlice(validatorsArray)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve validator indices to pubkeys: %w", err)
		}

		// Convert pubkeys to bytes for PostgreSQL
		byteaArray = make(pq.ByteaArray, len(validatorPubkeys))
		for i, p := range validatorPubkeys {
			byteaArray[i], _ = hexutil.Decode(p)
		}
	}

	// Custom type for block_index
	var data []struct {
		GroupId              sql.NullInt64   `db:"group_id"`
		PublicKey            []byte          `db:"publickey"`
		Slot                 int64           `db:"block_slot"`
		SlotIndex            int64           `db:"block_index"`
		WithdrawalCredential []byte          `db:"withdrawalcredentials"`
		Amount               decimal.Decimal `db:"amount"`
		Signature            []byte          `db:"signature"`
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
				LEFT JOIN blocks_deposits bd ON bd.block_slot = cbdl.block_slot
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

	err = db.AlloyReader.Select(&data, query+filterFragment, params...)

	if err != nil {
		return nil, nil, err
	}

	pubkeys := make([]string, len(data))
	for i, row := range data {
		pubkeys[i] = hexutil.Encode(row.PublicKey)
	}
	indices, err := d.services.GetValidatorIndexOfPubkeySlice(pubkeys)
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
			Amount:               row.Amount,
			Signature:            t.Hash(hexutil.Encode(row.Signature)),
		}
		if row.GroupId.Valid {
			responseData[i].GroupId = uint64(row.GroupId.Int64)
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
