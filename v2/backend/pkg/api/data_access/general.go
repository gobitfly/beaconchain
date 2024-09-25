package dataaccess

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
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
// 1st param defines default column precedence and direction
// 2nd param defines requested primary sort
// TODO pagination
func applySortAndPagination(defaultColumnsOrder []t.SortColumn, primary t.SortColumn) []exp.OrderedExpression {
	// prepare ordering columns; always need all columns to ensure consistent ordering
	queryOrderColumns := make([]t.SortColumn, len(defaultColumnsOrder))
	queryOrderColumns = append(queryOrderColumns, primary)
	// secondary sorts according to default
	for _, column := range defaultColumnsOrder {
		if column.Column != primary.Column {
			queryOrderColumns = append(queryOrderColumns, column)
		}
	}

	// apply ordering
	queryColumns := []exp.OrderedExpression{}
	for _, column := range queryOrderColumns {
		col := goqu.C(column.Column).Asc()
		if column.Desc {
			col = goqu.C(column.Column).Desc()
		}
		queryColumns = append(queryColumns, col)
	}
	return queryColumns
}
