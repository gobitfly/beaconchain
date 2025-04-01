package db2

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
)

type ConsensusRepository interface {
	GetLatestEpoch() (uint64, error)
	GetDepositsCountForBlockSlot() (uint64, error)
	SaveBlockDeposits(validatorIndex uint64, pubkey []byte, withdrawalCredentials []byte, balance uint64) error
	UpdateBlockDepositsSignature() error
	UpdateBlockDepositCount(count int) error
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

// GetLatestEpoch will return the latest epoch from the database
func (c *ConsensusDB) GetLatestEpoch() (uint64, error) {
	var epoch uint64
	err := c.WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %w", err)
	}

	return epoch, nil
}

func (c *ConsensusDB) GetDepositsCountForBlockSlot() (uint64, error) {
	var count uint64
	err := c.WriterDb.Get(&count, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *ConsensusDB) SaveBlockDeposits(vIndex uint64, vPubkey, vWithdrawalCredentials []byte, vBalance uint64) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
	VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
		vIndex, vPubkey, vWithdrawalCredentials, vBalance, []byte{0x0},
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) UpdateBlockDepositsSignature() error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	// hydrate the eth1 deposit signature for all genesis validators that have a corresponding eth1 deposit
	_, err = tx.Exec(`
		UPDATE blocks_deposits
		SET signature = a.signature
		FROM (
			SELECT DISTINCT ON(publickey) publickey, signature
			FROM eth1_deposits
			WHERE valid_signature = true) AS a
		WHERE block_slot = 0 AND blocks_deposits.publickey = a.publickey AND blocks_deposits.signature = '\x'`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *ConsensusDB) UpdateBlockDepositCount(count int) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", count)

	if err != nil {
		return err
	}

	return tx.Commit()
}
