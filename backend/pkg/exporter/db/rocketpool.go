package db

import (
	"fmt"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RocketpoolDBRepository interface {
	SaveRocketPoolDAOMembers(valueStrings []string, valueArgs []interface{}) error
	DeleteRocketPoolDAOMembers(addresses [][]byte) error
}

type RocketpoolDB struct {
	WriterDb *sqlx.DB
}

func NewRocketpoolDB(writerDb *sqlx.DB) *RocketpoolDB {
	return &RocketpoolDB{
		WriterDb: writerDb,
	}
}

func (rp *RocketpoolDB) SaveRocketPoolDAOMembers(valueStrings []string, valueArgs []interface{}) error {
	tx, err := rp.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(
		fmt.Sprintf(`
			INSERT INTO rocketpool_dao_members (
				rocketpool_storage_address,
				address,
				id,
				url,
				joined_time,
				last_proposal_time,
				rpl_bond_amount,
				unbonded_validator_count
			)
			VALUES %s
			ON CONFLICT (rocketpool_storage_address, address) DO UPDATE SET
				id = excluded.id,
				url = excluded.url,
				joined_time = excluded.joined_time,
				last_proposal_time = excluded.last_proposal_time,
				rpl_bond_amount = excluded.rpl_bond_amount,
				unbonded_validator_count = excluded.unbonded_validator_count
			`, strings.Join(valueStrings, ",")), valueArgs...)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (rp *RocketpoolDB) DeleteRocketPoolDAOMembers(addresses [][]byte) error {
	tx, err := rp.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`
	DELETE FROM rocketpool_dao_members 
	WHERE NOT address = ANY($1)`, pq.ByteaArray(addresses))

	if err != nil {
		return err
	}

	return tx.Commit()
}
