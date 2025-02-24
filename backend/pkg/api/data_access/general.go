package dataaccess

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/jmoiron/sqlx"
)

// retrieve (primary) ens name and optional name (=label) maintained by beaconcha.in, if present
func (d *DataAccessService) GetNamesAndEnsForAddresses(ctx context.Context, addressMap map[string]*types.Address) error {
	addresses := make([][]byte, 0, len(addressMap))
	ensMapping := make(map[string]string, len(addressMap))
	for address, data := range addressMap {
		ensMapping[address] = ""
		add, err := hexutil.Decode(address)
		if err != nil {
			return err
		}
		addresses = append(addresses, add)
		if data == nil {
			addressMap[address] = &types.Address{Hash: types.Hash(address)}
		}
	}
	// determine ENS names
	if err := db.GetEnsNamesForAddresses(ensMapping); err != nil {
		return err
	}
	for address, ens := range ensMapping {
		addressMap[address].Ens = ens
	}

	// determine names
	names := []struct {
		Address []byte `db:"address"`
		Name    string `db:"name"`
	}{}
	err := d.alloyReader.SelectContext(ctx, &names, `SELECT address, name FROM address_names WHERE address = ANY($1)`, addresses)
	if err != nil {
		return err
	}

	for _, name := range names {
		addressMap[hexutil.Encode(name.Address)].Label = name.Name
	}
	return nil
}

// helper function to sort and apply pagination to a query
// 1st param is the list of all columns necessary to sort the table deterministically; it defines their precedence and sort direction
// 2nd param is the requested sort column; it may or may not be part of the default columns (if it is, you don't have to specify the cursor limit again)
func applySortAndPagination(defaultColumns []types.SortColumn, primary types.SortColumn, cursor types.GenericCursor) ([]exp.OrderedExpression, exp.Expression, error) {
	// prepare ordering columns; always need all columns to ensure consistent ordering
	queryOrderColumns := make([]types.SortColumn, 0, len(defaultColumns))
	queryOrderColumns = append(queryOrderColumns, primary)
	// secondary sorts according to default
	for _, column := range defaultColumns {
		if column.Column == primary.Column {
			if primary.Offset == nil {
				queryOrderColumns[0].Offset = column.Offset
			}
			continue
		}
		queryOrderColumns = append(queryOrderColumns, column)
	}

	// apply ordering
	queryOrder := []exp.OrderedExpression{}
	for i := range queryOrderColumns {
		column := &queryOrderColumns[i]
		if cursor.IsReverse() {
			column.Desc = !column.Desc
		}
		colOrder := column.Column.Asc().NullsFirst()
		if column.Desc {
			colOrder = column.Column.Desc().NullsLast()
		}
		queryOrder = append(queryOrder, colOrder)
	}

	// apply cursor offsets
	var queryWhere exp.Expression
	if cursor.IsValid() {
		// reverse order to nest conditions
		for i := len(queryOrderColumns) - 1; i >= 0; i-- {
			column := queryOrderColumns[i]
			var colWhere exp.Expression

			// current convention is opposite of the psql default (ASC: nulls first, DESC: nulls last)
			if column.Desc {
				if column.Offset == nil && queryWhere == nil {
					continue
				}

				colWhere = goqu.Or(column.Column.Lt(column.Offset), column.Column.IsNull())

				if queryWhere == nil {
					queryWhere = colWhere
				} else {
					if column.Offset == nil {
						queryWhere = goqu.And(column.Column.IsNull(), queryWhere)
					} else {
						queryWhere = goqu.And(column.Column.Eq(column.Offset), queryWhere)
						queryWhere = goqu.Or(colWhere, queryWhere)
					}
				}
			} else {
				if column.Offset == nil {
					colWhere = column.Column.IsNotNull()
				} else {
					colWhere = column.Column.Gt(column.Offset)
				}

				if queryWhere == nil {
					queryWhere = colWhere
				} else {
					queryWhere = goqu.And(column.Column.Eq(column.Offset), queryWhere)
					queryWhere = goqu.Or(colWhere, queryWhere)
				}
			}
		}

		if queryWhere == nil {
			return nil, nil, fmt.Errorf("cursor given for descending order but all offset are nil meaning no data after it")
		}
	}

	return queryOrder, queryWhere, nil
}

// Generic function to execute a query
// Ensures T is a slice in compile-time
// Retrieves multiple rows and stores them in a slice
// Used when the query returns multiple results (e.g., SELECT * FROM users)
// The destination must be a slice ([]T)
func runQueryRows[T ~[]E, E any](ctx context.Context, db *sqlx.DB, ds *goqu.SelectDataset) (T, error) {
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error preparing query: %w", err)
	}

	var result T // slice of type T
	err = db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return result, fmt.Errorf("error executing query: %w", err)
	}

	return result, nil
}

// Generic function to execute a query
// Retrieves a single row and stores it in a struct or variable
// Used when the query is expected to return only one row (e.g., SELECT * FROM users WHERE id = ?)
// The destination must be a single struct or variable (T)
func runQuery[T any](ctx context.Context, db *sqlx.DB, ds *goqu.SelectDataset) (T, error) {
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error preparing query: %w", err)
	}

	var result T
	err = db.GetContext(ctx, &result, query, args...)
	if err != nil {
		return result, fmt.Errorf("error executing query: %w", err)
	}

	return result, nil
}

// returns the effective balances of the provided validators
// if onlyActive = true: executed from the vdb premium limits pov, i.e. exited validators account for the EB at exit time
func (d *DataAccessService) GetValidatorsEffectiveBalances(ctx context.Context, validators []t.VDBValidator, onlyActive bool) (map[t.VDBValidator]uint64, error) {
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	// active
	effectiveBalances := make(map[t.VDBValidator]uint64)
	var validatorExitEpochs []exp.Expression
	for _, validator := range validators {
		if len(validatorMapping.ValidatorMetadata) <= int(validator) {
			return nil, fmt.Errorf("validator index %d not found in validator mapping", validator)
		}
		status := constypes.ValidatorDbStatus(validatorMapping.ValidatorMetadata[validator].Status)
		if !onlyActive &&
			(status == constypes.DbSlashed || status == constypes.DbExited) &&
			validatorMapping.ValidatorMetadata[validator].EffectiveBalance == 0 {
			// exited & balance withdrawn, need to query latest EB before exit
			if !validatorMapping.ValidatorMetadata[validator].ExitEpoch.Valid {
				return nil, fmt.Errorf("validator %d has no exit epoch", validator)
			}
			// goqu can't pass tuples directly
			epochTs := utils.EpochToTime(uint64(validatorMapping.ValidatorMetadata[validator].ExitEpoch.Int64)).Unix()
			validatorExitEpochs = append(validatorExitEpochs, goqu.L("(fromUnixTimestamp(?), ?)", epochTs, validator))
		} else {
			effectiveBalances[validator] = validatorMapping.ValidatorMetadata[validator].EffectiveBalance
		}
	}

	// exited
	if len(validatorExitEpochs) > 0 {
		ds := goqu.Dialect("postgres").
			Select(
				// goqu.SUM(goqu.I("balance_effective_end")).As("balance_effective_end"),
				goqu.I("validator_index"),
				goqu.I("balance_effective_end"),
			).
			From("validator_dashboard_data_epoch").
			Where(
				goqu.L("(epoch_timestamp, validator_index)").In(validatorExitEpochs),
			)
		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return nil, err
		}

		ebsBeforeExit := []struct {
			ValidatorIndex   uint64 `db:"validator_index"`
			EffectiveBalance uint64 `db:"balance_effective_end"`
		}{}
		err = d.clickhouseReader.SelectContext(ctx, &ebsBeforeExit, query, args...)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		for _, eb := range ebsBeforeExit {
			effectiveBalances[eb.ValidatorIndex] = eb.EffectiveBalance
		}
	}
	return effectiveBalances, nil
}

func (d *DataAccessService) GetValidatorsEffectiveBalanceTotal(ctx context.Context, validators []t.VDBValidator, onlyActive bool) (uint64, error) {
	validatorEbs, err := d.GetValidatorsEffectiveBalances(ctx, validators, onlyActive)
	if err != nil {
		return 0, err
	}

	var totalEb uint64
	for _, validator := range validators {
		if eb, ok := validatorEbs[validator]; !ok {
			return 0, fmt.Errorf("effective balance for validator %d not found", validator)
		} else {
			totalEb += eb
		}
	}
	return totalEb, nil
}
