package dataaccess

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
)

func (d *DataAccessService) GetLabelsAndEnsForAddresses(ctx context.Context, addressMap map[string]*types.Address) error {
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

	// determine tags
	tags := []struct {
		Address []byte `db:"address"`
		Tag     string `db:"tag"`
	}{}
	err := d.alloyReader.SelectContext(ctx, &tags, `SELECT address, tag FROM address_tags WHERE address = ANY($1)`, addresses)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		addressMap[hexutil.Encode(tag.Address)].Label = tag.Tag
	}
	return nil
}
