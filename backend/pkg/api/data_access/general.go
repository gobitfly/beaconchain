package dataaccess

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
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
// 2nd param is the requested sort column; it may or may not be part of the default columns
func applySortAndPagination(defaultColumns []types.SortColumn, primary types.SortColumn, cursor types.GenericCursor) ([]exp.OrderedExpression, exp.Expression) {
	// prepare ordering columns; always need all columns to ensure consistent ordering
	queryOrderColumns := make([]types.SortColumn, 0, len(defaultColumns))
	queryOrderColumns = append(queryOrderColumns, primary)
	// secondary sorts according to default
	for _, column := range defaultColumns {
		if column.Column != primary.Column {
			queryOrderColumns = append(queryOrderColumns, column)
		}
	}

	// apply ordering
	queryOrder := []exp.OrderedExpression{}
	for i := range queryOrderColumns {
		if cursor.IsReverse() {
			queryOrderColumns[i].Desc = !queryOrderColumns[i].Desc
		}
		column := queryOrderColumns[i]
		colOrder := goqu.C(column.Column).Asc()
		if column.Desc {
			colOrder = goqu.C(column.Column).Desc()
		}
		queryOrder = append(queryOrder, colOrder)
	}

	// apply cursor offsets
	var queryWhere exp.Expression
	if cursor.IsValid() {
		// reverse order to nest conditions
		for i := len(queryOrderColumns) - 1; i >= 0; i-- {
			column := queryOrderColumns[i]
			colWhere := goqu.C(column.Column).Gt(column.Offset)
			if column.Desc {
				colWhere = goqu.C(column.Column).Lt(column.Offset)
			}

			equal := goqu.C(column.Column).Eq(column.Offset)
			if queryWhere == nil {
				queryWhere = equal
			} else {
				queryWhere = goqu.And(equal, queryWhere)
			}
			queryWhere = goqu.Or(colWhere, queryWhere)
		}
	}

	return queryOrder, queryWhere
}
