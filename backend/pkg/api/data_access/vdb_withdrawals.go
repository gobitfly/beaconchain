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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardWithdrawals(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBWithdrawalsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBWithdrawalsTableRow, *t.Paging, error) {
	result := make([]t.VDBWithdrawalsTableRow, 0)
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
	validatorSearch, err := d.getValidatorSearch(search)
	if err != nil {
		return nil, nil, err
	}
	if validatorSearch == nil {
		// No validators found
		return result, &paging, nil
	}

	validatorGroupMap := make(map[t.VDBValidator]uint64)
	var validators []t.VDBValidator
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex t.VDBValidator `db:"validator_index"`
			GroupId        uint64         `db:"group_id"`
		}{}

		queryParams := []interface{}{dashboardId.Id}
		validatorsQuery := fmt.Sprintf(`
			SELECT
				validator_index,
				group_id
			FROM users_val_dashboards_validators
			WHERE dashboard_id = $%d
			`, len(queryParams))

		if len(validatorSearch) > 0 {
			queryParams = append(queryParams, pq.Array(validatorSearch))
			validatorsQuery += fmt.Sprintf(" AND validator_index = ANY ($%d)", len(queryParams))
		}

		err := d.alloyReader.SelectContext(ctx, &queryResult, validatorsQuery, queryParams...)
		if err != nil {
			return nil, nil, err
		}

		for _, res := range queryResult {
			groupId := res.GroupId
			if dashboardId.AggregateGroups {
				groupId = t.DefaultGroupId
			}
			validatorGroupMap[res.ValidatorIndex] = groupId
			validators = append(validators, res.ValidatorIndex)
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		validatorSearchMap := utils.SliceToMap(validatorSearch)

		for _, validator := range dashboardId.Validators {
			if _, ok := validatorSearchMap[validator]; len(validatorSearchMap) == 0 || ok {
				validatorGroupMap[validator] = t.DefaultGroupId
				validators = append(validators, validator)
			}
		}
	}

	if len(validators) == 0 {
		// Return if there are no validators
		return result, &paging, nil
	}

	// Get the withdrawals for the validators
	queryResult := []struct {
		BlockSlot       uint64 `db:"block_slot"`
		BlockNumber     uint64 `db:"exec_block_number"`
		WithdrawalIndex uint64 `db:"withdrawalindex"`
		ValidatorIndex  uint64 `db:"validatorindex"`
		Address         []byte `db:"address"`
		Amount          uint64 `db:"amount"`
	}{}

	queryParams := []interface{}{}
	withdrawalsQuery := `
		SELECT
		    w.block_slot,
		    b.exec_block_number,
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

	// Limit the query using sorting and the cursor
	orderQuery := ""
	sortColName := ""
	sortColCursor := interface{}(nil)
	switch colSort.Column {
	case enums.VDBWithdrawalsColumns.Epoch, enums.VDBWithdrawalsColumns.Slot:
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
		colSort.Column == enums.VDBWithdrawalsColumns.Slot {
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

	queryParams = append(queryParams, limit+1)
	limitQuery := fmt.Sprintf(" LIMIT $%d", len(queryParams))

	withdrawalsQuery += whereQuery + orderQuery + limitQuery

	err = d.readerDb.SelectContext(ctx, &queryResult, withdrawalsQuery, queryParams...)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting withdrawals for dashboardId: %d (%d validators): %w", dashboardId.Id, len(validators), err)
	}

	if len(queryResult) == 0 {
		// No withdrawals found
		return result, &paging, nil
	}

	// Prepare the ENS map
	addressMapping := make(map[string]*t.Address)
	contractStatusRequests := make([]db.ContractInteractionAtRequest, len(queryResult))
	for i, withdrawal := range queryResult {
		address := hexutil.Encode(withdrawal.Address)
		addressMapping[address] = nil
		contractStatusRequests[i] = db.ContractInteractionAtRequest{
			Address:  fmt.Sprintf("%x", withdrawal.Address),
			Block:    int64(withdrawal.BlockNumber),
			TxIdx:    -1,
			TraceIdx: -1,
		}
	}

	// Get the ENS names and (label) names for the addresses
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, nil, err
	}

	// Get the contract status for the addresses
	contractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(contractStatusRequests)
	if err != nil {
		return nil, nil, err
	}

	// Create the result
	cursorData := make([]t.WithdrawalsCursor, 0)
	for i, withdrawal := range queryResult {
		address := hexutil.Encode(withdrawal.Address)
		result = append(result, t.VDBWithdrawalsTableRow{
			Epoch:     withdrawal.BlockSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
			Slot:      withdrawal.BlockSlot,
			Index:     withdrawal.ValidatorIndex,
			Recipient: *addressMapping[address],
			GroupId:   validatorGroupMap[withdrawal.ValidatorIndex],
			Amount:    utils.GWeiToWei(big.NewInt(int64(withdrawal.Amount))),
		})
		result[i].Recipient.IsContract = contractStatuses[i] == types.CONTRACT_CREATION || contractStatuses[i] == types.CONTRACT_PRESENT
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

	// Find the next withdrawal if we are currently at the first page
	// If we have a prev_cursor but not enough data it means the next data is missing
	if !currentCursor.IsValid() || (currentCursor.IsReverse() && len(result) < int(limit)) {
		nextData, err := d.getNextWithdrawalRow(validators)
		if err != nil {
			return nil, nil, err
		}
		if nextData != nil {
			// Complete the next data
			nextData.GroupId = validatorGroupMap[nextData.Index]
			// TODO integrate label/ens data for "next" row
			// nextData.Recipient.Ens = addressEns[string(nextData.Recipient.Hash)]
		} else {
			// If there is no next data, add a missing estimate row
			nextData = &t.VDBWithdrawalsTableRow{
				IsMissingEstimate: true,
			}
		}
		result = append([]t.VDBWithdrawalsTableRow{*nextData}, result...)

		// Flag if above limit
		moreDataFlag = moreDataFlag || len(result) > int(limit)
		if !moreDataFlag && !currentCursor.IsValid() {
			// No paging required
			return result, &paging, nil
		}

		// Remove the last entry from data as it is only required for the check
		if moreDataFlag {
			result = result[:len(result)-1]
			cursorData = cursorData[:len(cursorData)-1]
		}
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) getNextWithdrawalRow(queryValidators []t.VDBValidator) (*t.VDBWithdrawalsTableRow, error) {
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
		return queryValidators[i] < queryValidators[j]
	})

	latestFinalized := cache.LatestFinalizedEpoch.Get()

	var nextValidator *t.VDBValidator
	for _, validator := range queryValidators {
		metadata := validatorMapping.ValidatorMetadata[validator]

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

		if (metadata.Balance > 0 && metadata.WithdrawableEpoch.Valid && metadata.WithdrawableEpoch.Int64 <= int64(epoch)) ||
			(metadata.EffectiveBalance == utils.Config.Chain.ClConfig.MaxEffectiveBalance && metadata.Balance > utils.Config.Chain.ClConfig.MaxEffectiveBalance) {
			// this validator is eligible for withdrawal, check if it is the next one
			if nextValidator == nil || validator > *stats.LatestValidatorWithdrawalIndex {
				distance, err := d.getWithdrawableCountFromCursor(validator, *stats.LatestValidatorWithdrawalIndex)
				if err != nil {
					return nil, err
				}

				timeToWithdrawal := d.getTimeToNextWithdrawal(distance)

				// it normally takes two epochs to finalize
				if !timeToWithdrawal.Before(utils.EpochToTime(epoch + (epoch - latestFinalized))) {
					// this validator has a next withdrawal
					nextValidatorInt := validator
					nextValidator = &nextValidatorInt
				}

				if nextValidator != nil && *nextValidator > *stats.LatestValidatorWithdrawalIndex {
					// the first validator after the cursor has to be the next validator
					break
				}
			}
		}
	}

	if nextValidator == nil {
		return nil, nil
	}

	nextValidatorData := validatorMapping.ValidatorMetadata[*nextValidator]

	lastWithdrawnEpochs, err := db.GetLastWithdrawalEpoch([]t.VDBValidator{*nextValidator})
	if err != nil {
		return nil, err
	}
	lastWithdrawnEpoch := lastWithdrawnEpochs[*nextValidator]

	nextDistance, err := d.getWithdrawableCountFromCursor(*nextValidator, *stats.LatestValidatorWithdrawalIndex)
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
	if nextValidatorData.WithdrawableEpoch.Valid && nextValidatorData.WithdrawableEpoch.Int64 <= int64(epoch) {
		// full withdrawal
		withdrawalAmount = nextValidatorData.Balance
	} else {
		// partial withdrawal
		withdrawalAmount = nextValidatorData.Balance - utils.Config.Chain.ClConfig.MaxEffectiveBalance
	}

	if lastWithdrawnEpoch == epoch || nextValidatorData.Balance < utils.Config.Chain.ClConfig.MaxEffectiveBalance {
		withdrawalAmount = 0
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

	nextData := &t.VDBWithdrawalsTableRow{
		Epoch: nextWithdrawalSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
		Slot:  nextWithdrawalSlot,
		Index: *nextValidator,
		Recipient: t.Address{
			Hash:       t.Hash(address.String()),
			Ens:        ens_name,
			IsContract: contractStatus[0] == types.CONTRACT_CREATION || contractStatus[0] == types.CONTRACT_PRESENT,
		},
		Amount: utils.GWeiToWei(big.NewInt(int64(withdrawalAmount))),
	}

	return nextData, nil
}

func (d *DataAccessService) GetValidatorDashboardTotalWithdrawals(ctx context.Context, dashboardId t.VDBId, search string, protocolModes t.VDBProtocolModes) (*t.VDBTotalWithdrawalsData, error) {
	result := &t.VDBTotalWithdrawalsData{
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

	queryResult := []struct {
		ValidatorIndex t.VDBValidator `db:"validator_index"`
		Epoch          uint64         `db:"epoch_end"`
		Amount         int64          `db:"acc_withdrawals_amount"`
	}{}

	withdrawalsQuery := `
			WITH validators AS (
				SELECT validator_index FROM users_val_dashboards_validators WHERE (dashboard_id = $1)
			)
			SELECT
				validator_index,
				SUM(withdrawals_amount) AS acc_withdrawals_amount,
				MAX(epoch_end) AS epoch_end
			FROM validator_dashboard_data_rolling_total FINAL
			INNER JOIN validators v ON validator_dashboard_data_rolling_total.validator_index = v.validator_index
			WHERE validator_index IN (select validator_index FROM validators)
			GROUP BY validator_index
		`

	if dashboardId.Validators != nil {
		withdrawalsQuery = `
			SELECT
				validator_index,
				SUM(withdrawals_amount) AS acc_withdrawals_amount,
				MAX(epoch_end) AS epoch_end
			from validator_dashboard_data_rolling_total FINAL
			where validator_index IN ($1)
			group by validator_index
		`
	}

	dashboardValidators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		dashboardValidators = dashboardId.Validators
	}

	if len(dashboardValidators) > 0 {
		err = d.clickhouseReader.SelectContext(ctx, &queryResult, withdrawalsQuery, dashboardValidators)
	} else {
		err = d.clickhouseReader.SelectContext(ctx, &queryResult, withdrawalsQuery, dashboardId.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("error getting total withdrawals for validators: %+v: %w", dashboardId, err)
	}

	if len(queryResult) == 0 {
		// No validators to search for
		return result, nil
	}

	var totalAmount int64
	var validators []t.VDBValidator
	lastEpoch := queryResult[0].Epoch
	lastSlot := (lastEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

	for _, res := range queryResult {
		// Calculate the total amount of withdrawals
		totalAmount += res.Amount

		// Calculate the current validators
		validators = append(validators, res.ValidatorIndex)
	}

	var latestWithdrawalsAmount int64
	err = d.readerDb.GetContext(ctx, &latestWithdrawalsAmount, `
		SELECT
			COALESCE(SUM(w.amount), 0)
		FROM
		    blocks_withdrawals w
		INNER JOIN blocks b ON w.block_slot = b.slot AND w.block_root = b.blockroot AND b.status = '1'
		WHERE w.block_slot > $1 AND w.validatorindex = ANY ($2)
		`, lastSlot, validators)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error getting latest withdrawals for validators: %+v: %w", dashboardId, err)
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
