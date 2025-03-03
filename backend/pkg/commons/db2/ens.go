package db2

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

type ENS struct {
	NameHash  [32]byte       `db:"name_hash"`
	Name      string         `db:"ens_name"`
	Address   common.Address `db:"address"`
	IsPrimary bool           `db:"is_primary_name"`
	Expires   time.Time      `db:"valid_to"`
}

type ENSStore struct {
	db *sqlx.DB
}

func NewENSStore(db *sqlx.DB) *ENSStore {
	return &ENSStore{
		db: db,
	}
}

func (store ENSStore) GetENSNameFromHash(nameHash [32]byte) (string, error) {
	var name string
	err := store.db.Get(&name, `
					SELECT
						ens_name
					FROM ens
					WHERE name_hash = $1
					`, nameHash[:])
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return name, nil
}

func (store ENSStore) SetENS(ens ENS) error {
	_, err := store.db.Exec(`
	INSERT INTO ens (
		name_hash, 
		ens_name, 
		address,
		is_primary_name, 
		valid_to)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT 
		(name_hash) 
	DO UPDATE SET 
		ens_name = excluded.ens_name,
		address = excluded.address,
		is_primary_name = excluded.is_primary_name,
		valid_to = excluded.valid_to
	`, ens.NameHash[:], ens.Name, ens.Address.Bytes(), ens.IsPrimary, ens.Expires)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
			log.Warnf("could not insert ens name [%v]: %v", ens.Name, err)
			return nil
		}
		return fmt.Errorf("error writing ens data for name [%v]: %w", ens.Name, err)
	}
	return nil
}

func (store ENSStore) DeleteENS(name string) error {
	_, err := store.db.Exec(`
	DELETE FROM ens
	WHERE
		ens_name = $1
	;`, name)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
			log.Warnf("could not delete ens name [%v]: %v", name, err)
			return nil
		}
		return fmt.Errorf("error deleting ens name [%v]: %v", name, err)
	}
	return nil
}

func (store ENSStore) GetNamesForAddress(address common.Address) ([]string, error) {
	var names []string
	err := store.db.Select(&names, `SELECT ens_name FROM ens WHERE address = $1 AND is_primary_name AND valid_to >= now()`, address.Bytes())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return names, nil
}

func (store ENSStore) GetAllENS() ([]ENS, error) {
	type dbENS struct {
		NameHash  []byte         `db:"name_hash"`
		Name      string         `db:"ens_name"`
		Address   common.Address `db:"address"`
		IsPrimary bool           `db:"is_primary_name"`
		Expires   time.Time      `db:"valid_to"`
	}
	var dbENSs []dbENS
	err := store.db.Select(&dbENSs, `SELECT * FROM ens`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var all []ENS
	for _, ens := range dbENSs {
		all = append(all, ENS{
			NameHash:  toByte32(ens.NameHash),
			Name:      ens.Name,
			Address:   ens.Address,
			IsPrimary: ens.IsPrimary,
			Expires:   ens.Expires,
		})
	}
	return all, nil
}
