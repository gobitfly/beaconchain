package dataaccess

import (
	"context"

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
