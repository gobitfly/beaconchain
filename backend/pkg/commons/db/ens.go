package db

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// pass invalid time to get latest data
func GetEnsNameForAddress(address common.Address, validUntil time.Time) (name string, err error) {
	if validUntil.IsZero() {
		validUntil = time.Now()
	}
	err = ReaderDb.Get(&name, `
	SELECT ens_name
	FROM ens
	WHERE
		address = $1 AND
		is_primary_name AND
		valid_to >= $2
	;`, address.Bytes(), validUntil)
	return name, err
}

func GetEnsNamesForAddresses(addressMap map[string]string) error {
	if len(addressMap) == 0 {
		return nil
	}
	type pair struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}
	dbAddresses := []pair{}
	addresses := make([][]byte, 0, len(addressMap))
	for address := range addressMap {
		add, err := hexutil.Decode(address)
		if err != nil {
			return err
		}
		addresses = append(addresses, add)
	}

	err := ReaderDb.Select(&dbAddresses, `
	SELECT address, ens_name
	FROM ens
	WHERE
		address = ANY($1) AND
		is_primary_name AND
		valid_to >= now()
	;`, addresses)
	if err != nil {
		return err
	}
	for _, foundling := range dbAddresses {
		addressMap[hexutil.Encode(foundling.Address)] = foundling.EnsName
	}
	return nil
}
