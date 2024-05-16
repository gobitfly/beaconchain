package modules

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type executionPayloadsExporter struct {
	ModuleContext ModuleContext
	ExportMutex   *sync.Mutex
}

func NewExecutionPayloadsExporter(moduleContext ModuleContext) ModuleInterface {
	return &executionPayloadsExporter{
		ModuleContext: moduleContext,
		ExportMutex:   &sync.Mutex{},
	}
}

func (d *executionPayloadsExporter) OnHead(event *constypes.StandardEventHeadResponse) (err error) {
	return nil // nop
}

func (d *executionPayloadsExporter) Init() error {
	return nil // nop
}

func (d *executionPayloadsExporter) GetName() string {
	return "ExecutionPayloads-Exporter"
}

func (d *executionPayloadsExporter) OnChainReorg(event *constypes.StandardEventChainReorg) (err error) {
	return nil // nop
}

// can take however long it wants to run, is run in a separate goroutine, so no need to worry about blocking
func (d *executionPayloadsExporter) OnFinalizedCheckpoint(event *constypes.StandardFinalizedCheckpointResponse) (err error) {
	// if mutex is locked, return early
	if !d.ExportMutex.TryLock() {
		log.Infof("execution payloads exporter is already running")
		return nil
	}
	defer d.ExportMutex.Unlock()

	err = d.maintainTable()
	if err != nil {
		return fmt.Errorf("error maintaining table: %w", err)
	}

	start := time.Now()
	// update cached view
	err = d.updateCachedView()
	if err != nil {
		return err
	}

	log.Debugf("updating execution payloads cached view took %v", time.Since(start))
	return nil
}

func (d *executionPayloadsExporter) updateCachedView() (err error) {
	err = db.CacheQuery(`
		SELECT DISTINCT ON (uvdv.dashboard_id, uvdv.group_id, b.slot)
			uvdv.dashboard_id,
			uvdv.group_id,
			b.slot,
			coalesce(rb.value / 1e18, ep.fee_recipient_reward) as reward,
			coalesce(rb.proposer_fee_recipient, b.exec_fee_recipient) as fee_recipient, 
			rb.value IS NOT NULL AS is_mev
		FROM
			blocks b
			INNER JOIN execution_payloads ep ON ep.block_hash = b.exec_block_hash
			INNER JOIN users_val_dashboards_validators uvdv ON b.proposer = uvdv.validator_index
			LEFT JOIN relays_blocks rb ON rb.exec_block_hash = b.exec_block_hash
		WHERE
			b.status = '1'
			AND b.exec_block_hash IS NOT NULL
		ORDER BY
			dashboard_id,
			group_id,
			slot DESC;
	`, "cached_proposal_rewards", []string{"dashboard_id", "slot"}, []string{"dashboard_id", "reward"})
	return err
}

// this is basically synchronous, each time it gets called it will kill the previous export and replace it with itself
func (d *executionPayloadsExporter) maintainTable() (err error) {
	blocks := struct {
		MinBlock sql.NullInt64 `db:"min"`
		MaxBlock sql.NullInt64 `db:"max"`
	}{}
	err = db.ReaderDb.Get(&blocks, `
		SELECT
			MIN(b.exec_block_number),
			MAX(b.exec_block_number)
		FROM
			blocks b
			inner JOIN execution_payloads ep ON b.exec_block_hash = ep.block_hash
		WHERE
			ep.fee_recipient_reward IS NULL and b.status='1'
	`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Infof("no missing blocks found")
			return nil
		}
		return fmt.Errorf("error getting min and max block: %w", err)
	}
	if !blocks.MinBlock.Valid || !blocks.MaxBlock.Valid {
		log.Infof("no missing blocks found")
		return nil
	}
	minBlock := uint64(blocks.MinBlock.Int64)
	maxBlock := uint64(blocks.MaxBlock.Int64)

	if minBlock == 0 {
		log.Infof("no missing blocks found")
		return nil
	}

	log.Infof("min block: %v, max block: %v", blocks.MinBlock, blocks.MaxBlock)

	// channel that will receive blocks from bigtable
	blockChan := make(chan *types.Eth1BlockIndexed, 1000)
	type Result struct {
		BlockHash          []byte
		FeeRecipientReward decimal.Decimal
	}
	resData := make([]Result, 0, maxBlock-minBlock+1)

	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	group, _ := errgroup.WithContext(ctx)

	// coroutine to process the blocks
	group.Go(func() error {
		var block *types.Eth1BlockIndexed
		for i := minBlock; i <= maxBlock; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case b, ok := <-blockChan:
				block = b
				if !ok {
					return nil
				}
			}
			// read txn reward
			raw := block.GetTxReward()
			if raw == nil {
				// use 0
				raw = []byte{0}
			}
			hash := block.GetHash()
			if hash == nil {
				return fmt.Errorf("error getting block hash for block %v", block.Number)
			}
			// convert raw (bytes) to bigint
			num := new(big.Int).SetBytes(raw)
			// convert bigint to decimal
			dec := decimal.NewFromBigInt(num, -18)
			if err != nil {
				return fmt.Errorf("error converting tx reward to decimal for block %v: %w", block.Number, err)
			}
			resData = append(resData, Result{BlockHash: hash, FeeRecipientReward: dec})
		}
		return nil
	})

	err = db.BigtableClient.StreamBlocksIndexedDescending(blockChan, maxBlock, minBlock)

	if err != nil {
		finish()
		return fmt.Errorf("error streaming blocks: %w", err)
	}

	err = group.Wait()

	if err != nil {
		return fmt.Errorf("error processing blocks: %w", err)
	}

	// update the execution_payloads table

	log.Infof("preparing copy update to temp table")

	// load data into temp table
	_, err = db.WriterDb.Exec(`
		CREATE TEMP TABLE temp_fee_recipient_reward (
			block_hash bytea PRIMARY KEY,
			fee_recipient_reward numeric
		);
	`)
	defer func() {
		_, err = db.WriterDb.Exec(`DROP TABLE IF EXISTS temp_fee_recipient_reward;`)
		if err != nil {
			log.Error(err, "error dropping temp table", 0)
		}
	}()

	if err != nil {
		return fmt.Errorf("error creating temp table: %w", err)
	}

	// prepare data for bulk insert
	dat := make([][]interface{}, len(resData))
	for i, r := range resData {
		dat[i] = []interface{}{r.BlockHash, r.FeeRecipientReward}
	}

	log.Infof("copying data to temp table")

	err = db.CopyToTable("temp_fee_recipient_reward", []string{"block_hash", "fee_recipient_reward"}, dat)
	if err != nil {
		return fmt.Errorf("error copying data to temp table: %w", err)
	}

	log.Infof("updating execution_payloads")

	// update execution_payloads using temp table
	_, err = db.WriterDb.Exec(`
		UPDATE execution_payloads ep
		SET fee_recipient_reward = t.fee_recipient_reward
		FROM temp_fee_recipient_reward t
		WHERE ep.block_hash = t.block_hash;
	`)
	if err != nil {
		return fmt.Errorf("error updating execution_payloads: %w", err)
	}

	log.Infof("finished updating execution_payloads")

	if err != nil {
		return fmt.Errorf("error caching data: %w", err)
	}

	return nil
}
