package db2

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ConsensusRepository interface {
	SaveValidatorTags(data []types.SSVExporterData) error
	DeleteValidatorTags() error
	DeleteInvalidTags() error
}

type ConsensusDB struct {
	ReaderDb *sqlx.DB
	WriterDb *sqlx.DB
}

func NewConsensusRepository(reader, writer *sqlx.DB) *ConsensusDB {
	return &ConsensusDB{
		ReaderDb: reader,
		WriterDb: writer,
	}
}

func (c *ConsensusDB) SaveValidatorTags(data []types.SSVExporterData) error {
	index := 1
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([][]byte, 0, len(data)*index)

	for i, d := range data {
		pubkey, err := hex.DecodeString(strings.Replace(d.Publickey, "0x", "", -1))
		if err != nil {
			log.Error(err, "error decoding public key", 0)
			continue
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'ssv')", i*index+1))
		valueArgs = append(valueArgs, pubkey)
	}

	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(
		fmt.Sprintf(`
            INSERT INTO validator_tags (publickey, tag)
            VALUES %s
            ON CONFLICT (publickey, tag) DO NOTHING`,
			strings.Join(valueStrings, ",")), pq.ByteaArray(valueArgs))

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) DeleteValidatorTags() error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	for {
		res, err := tx.Exec(`DELETE FROM validator_tags WHERE publickey IN (SELECT publickey FROM validator_tags WHERE publickey NOT IN (SELECT pubkey FROM validators) LIMIT 1000)`)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *ConsensusDB) DeleteInvalidTags() error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	for {
		res, err := tx.Exec(`DELETE FROM validator_tags WHERE publickey IN (SELECT publickey FROM validator_tags WHERE tag = 'ssv' LIMIT 1000)`)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
