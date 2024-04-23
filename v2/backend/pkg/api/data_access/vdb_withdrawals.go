package dataaccess

import (
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func (d *DataAccessService) GetValidatorDashboardWithdrawals(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	var paging t.Paging

	// Initialize the cursor
	var currentCursor t.WithdrawalsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.WithdrawalsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsCursor: %w", err)
		}
	}

	// Prepare the sorting
	sortSearchDirection := ">"
	sortSearchOrder := " ASC"
	if (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse()) {
		sortSearchDirection = "<"
		sortSearchOrder = " DESC"
	}

	// Analyze the search term
	indexSearch := int64(-1)
	if search != "" {
		if utils.IsHash(search) || utils.IsEth1Address(search) {
			// Ensure that we have a "0x" prefix for the search term
			if !strings.HasPrefix(search, "0x") {
				search = "0x" + search
			}
			search = strings.ToLower(search)
			if utils.IsHash(search) {
				// Get the current validator state to convert pubkey to index
				validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
				defer releaseLock()
				if err != nil {
					return nil, nil, err
				}
				if index, ok := validatorMapping.ValidatorIndices[search]; ok {
					indexSearch = int64(*index)
				} else {
					// No validator index for pubkey found, return empty results
					return nil, &paging, nil
				}
			}
		} else if index, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(index)
		} else {
			// No allowed search term found, return empty results
			return nil, &paging, nil
		}
	}

	validatorGroupMap := make(map[uint64]uint64)
	var validators []uint64
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			GroupId        uint64 `db:"group_id"`
		}{}

		queryArgs := []interface{}{dashboardId.Id}
		validatorsQuery := fmt.Sprintf(`
			SELECT 
				validator_index,
				group_id
			FROM users_val_dashboards_validators
			WHERE dashboard_id = $%d
			`, len(queryArgs))

		if indexSearch != -1 {
			queryArgs = append(queryArgs, indexSearch)
			validatorsQuery += fmt.Sprintf(" AND validator_index = $%d", len(queryArgs))
		}

		err := d.alloyReader.Select(&queryResult, validatorsQuery, queryArgs...)
		if err != nil {
			return nil, nil, err
		}

		for _, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = res.GroupId
			validators = append(validators, res.ValidatorIndex)
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		for _, validator := range dashboardId.Validators {
			if indexSearch != -1 && validator.Index != uint64(indexSearch) {
				continue
			}
			validatorGroupMap[validator.Index] = t.DefaultGroupId
			validators = append(validators, validator.Index)
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return nil, &paging, nil
	}

	// Get the withdrawals for the validators
	queryResult := []struct {
		BlockSlot       uint64 `db:"block_slot"`
		WithdrawalIndex uint64 `db:"withdrawalindex"`
		ValidatorIndex  uint64 `db:"validatorindex"`
		Address         []byte `db:"address"`
		Amount          uint64 `db:"amount"`
	}{}

	queryParams := []interface{}{}
	withdrawalsQuery := `
		SELECT
		    w.block_slot,
			w.withdrawalindex,
			w.validatorindex,
			w.address,
			w.amount
		FROM
		    blocks_withdrawals w
		INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
		`

	// Limit the query to relevant validators
	queryParams = append(queryParams, pq.Array(validators))
	whereQuery := fmt.Sprintf(`
		WHERE
		    validatorindex = ANY ($%d)`, len(queryParams))

	// Limit the query using the search term if it is an address
	if utils.IsEth1Address(search) {
		searchAddress, err := hexutil.Decode(search)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode search term %s as address: %w", search, err)
		}
		queryParams = append(queryParams, searchAddress)
		whereQuery += fmt.Sprintf(" AND w.address = $%d", len(queryParams))
	}

	// Limit the query using sorting and the cursor
	orderQuery := ""
	sortColName := ""
	sortColCursor := interface{}(nil)
	switch colSort.Column {
	case enums.VDBWithdrawalsColumns.Epoch, enums.VDBWithdrawalsColumns.Slot, enums.VDBWithdrawalsColumns.Age:
	case enums.VDBWithdrawalsColumns.Index:
		sortColName = "w.validatorindex"
		sortColCursor = currentCursor.Index
	case enums.VDBWithdrawalsColumns.Recipient:
		sortColName = "w.address"
		sortColCursor = currentCursor.Recipient
	case enums.VDBWithdrawalsColumns.Amount:
		sortColName = "w.amount"
		sortColCursor = currentCursor.Amount
	}

	if colSort.Column == enums.VDBWithdrawalsColumns.Epoch ||
		colSort.Column == enums.VDBWithdrawalsColumns.Slot ||
		colSort.Column == enums.VDBWithdrawalsColumns.Age {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			queryParams = append(queryParams, currentCursor.Slot, currentCursor.WithdrawalIndex)
			whereQuery += fmt.Sprintf(" AND (w.block_slot%[1]s$%[2]d OR (w.block_slot=$%[2]d AND w.withdrawalindex%[1]s$%[3]d))",
				sortSearchDirection, len(queryParams)-1, len(queryParams))
		}
		orderQuery = fmt.Sprintf(" ORDER BY w.block_slot %[1]s, w.withdrawalindex %[1]s", sortSearchOrder)
	} else {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			queryParams = append(queryParams, sortColCursor, currentCursor.Slot, currentCursor.WithdrawalIndex)

			// The additional WHERE requirement is
			// WHERE sortColName>cursor OR (sortColName=cursor AND (block_slot>cursor OR (block_slot=cursor AND withdrawalindex>cursor)))
			// with the > flipped if the sort is descending
			whereQuery += fmt.Sprintf(" AND (%[1]s%[2]s$%[3]d OR (%[1]s=$%[3]d AND (w.block_slot%[2]s$%[4]d OR (w.block_slot=$%[4]d AND w.withdrawalindex%[2]s$%[5]d))))",
				sortColName, sortSearchDirection, len(queryParams)-2, len(queryParams)-1, len(queryParams))
		}
		// The ordering is
		// ORDER BY sortColName ASC, block_slot ASC, withdrawalindex ASC
		// with the ASC flipped if the sort is descending
		orderQuery = fmt.Sprintf(" ORDER BY %[1]s %[2]s, w.block_slot %[2]s, w.withdrawalindex %[2]s",
			sortColName, sortSearchOrder)
	}

	limitQuery := fmt.Sprintf(" LIMIT %d", limit+1)

	withdrawalsQuery += whereQuery + orderQuery + limitQuery

	err = d.readerDb.Select(&queryResult, withdrawalsQuery, queryParams...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &paging, nil
		}
		return nil, nil, fmt.Errorf("error getting withdrawals for validators: %+v: %w", validators, err)
	}

	// Prepare the ENS map
	addressEns := make(map[string]string)
	for _, withdrawal := range queryResult {
		address := hexutil.Encode(withdrawal.Address)
		addressEns[address] = ""
	}

	// Get the ENS names for the addresses
	if err := db.GetEnsNamesForAddresses(addressEns); err != nil {
		return nil, nil, err
	}

	// Create the result
	result := make([]t.VDBWithdrawalsTableRow, 0)
	cursorData := make([]t.WithdrawalsCursor, 0)
	for _, withdrawal := range queryResult {
		address := hexutil.Encode(withdrawal.Address)
		result = append(result, t.VDBWithdrawalsTableRow{
			Epoch:   withdrawal.BlockSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
			Slot:    withdrawal.BlockSlot,
			Index:   withdrawal.ValidatorIndex,
			GroupId: validatorGroupMap[withdrawal.ValidatorIndex],
			Recipient: t.Address{
				Hash: t.Hash(address),
				Ens:  addressEns[address],
			},
			Amount: utils.GWeiToWei(big.NewInt(int64(withdrawal.Amount))),
		})
		cursorData = append(cursorData, t.WithdrawalsCursor{
			Slot:            withdrawal.BlockSlot,
			WithdrawalIndex: withdrawal.WithdrawalIndex,
			Index:           withdrawal.ValidatorIndex,
			Recipient:       withdrawal.Address,
			Amount:          withdrawal.Amount,
		})
	}

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entry from data as it is only required for the check
	if moreDataFlag {
		result = result[:len(result)-1]
		cursorData = cursorData[:len(cursorData)-1]
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(result)
		slices.Reverse(cursorData)
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}
