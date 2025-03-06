package consensus

import (
	"database/sql"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func (c *consensusRepository) GetRelays() ([]types.Relay, error) {
	var relays []types.Relay
	err := c.ReaderDb.Select(&relays, `
		SELECT tag_id, endpoint, public_link, is_censoring, is_ethical, export_failure_count, last_export_try_ts, last_export_success_ts
		FROM relays`)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return relays, nil
}

func (c *consensusRepository) UpdateRelay(tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
		UPDATE relays SET
			export_failure_count = 0,
			last_export_try_ts = NOW() AT TIME ZONE 'utc',
			last_export_success_ts = NOW() AT TIME ZONE 'utc'
		WHERE tag_id = $1 AND endpoint = $2`, tagID, endpoint)

	return err
}

func (c *consensusRepository) UpdateRelayLastExportTry(tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
			UPDATE relays SET
				last_export_try_ts = (NOW() AT TIME ZONE 'utc')
			WHERE tag_id = $1 AND endpoint = $2`, tagID, endpoint)

	return err
}

func (c *consensusRepository) UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error {
	_, err := c.WriterDb.Exec(`
	UPDATE relays SET
		export_failure_count = $1,
		last_export_try_ts = (NOW() AT TIME ZONE 'utc')
	WHERE tag_id = $2 AND endpoint = $3`, exportFailureCount+1, tagID, endpoint)

	return err
}

func (c *consensusRepository) GetFirstRelayBlock(tagID string) (types.RelayBlock, error) {
	var block types.RelayBlock
	err := c.ReaderDb.Get(&block, `
		SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient
		FROM relays_blocks
		WHERE tag_id=$1
		ORDER BY block_slot ASC
		LIMIT 1`, tagID)

	return block, err
}

func (c *consensusRepository) GetLastRelayBlock(tagID string) (types.RelayBlock, error) {
	var block types.RelayBlock
	err := c.ReaderDb.Get(&block, `
		SELECT tag_id, block_slot, block_root, exec_block_hash, value, builder_pubkey, proposer_pubkey, proposer_fee_recipient
		FROM relays_blocks
		WHERE tag_id=$1
		ORDER BY block_slot DESC
		LIMIT 1`, tagID)

	return block, err
}

func (c *consensusRepository) SaveBlockTagsAndRelays(tagID string, payload types.BidTrace) error {
	tx, err := c.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	// first insert the tag into the blocks_tags table
	_, err = tx.Exec(`
	INSERT INTO blocks_tags
	SELECT blocks.slot, blocks.blockroot, $1
	FROM blocks
	WHERE blocks.slot = $2 AND blocks.exec_block_hash = $3
	ON CONFLICT DO NOTHING`, tagID, payload.Slot,
		utils.MustParseHex(payload.BlockHash))

	if err != nil {
		log.Error(fmt.Errorf("failed to insert payload into blocks_tags table"), "", 0, map[string]interface{}{"relay": tagID})
		return err
	}

	// save relays
	_, err = tx.Exec(`
		INSERT INTO relays_blocks
		(
			tag_id,
			block_slot,
			block_root,
			exec_block_hash,
			value,
			builder_pubkey,
			proposer_pubkey,
			proposer_fee_recipient
		)
		SELECT
			$1,	blocks.slot, blocks.blockroot, blocks.exec_block_hash, $4, $5, $6, $7
		FROM blocks
		WHERE
			blocks.slot = $2 and
			blocks.exec_block_hash = $3
		ON CONFLICT (block_slot, block_root, tag_id) DO NOTHING`,
		tagID, payload.Slot, payload.Value,
		utils.MustParseHex(payload.BlockHash),
		utils.MustParseHex(payload.BuilderPubkey),
		utils.MustParseHex(payload.ProposerPubkey),
		utils.MustParseHex(payload.ProposerFeeRecipient))

	if err != nil {
		log.Error(fmt.Errorf("failed to insert payload into relays_blocks table"), "", 0, map[string]interface{}{"relay": tagID})
		return err
	}

	return tx.Commit()
}
