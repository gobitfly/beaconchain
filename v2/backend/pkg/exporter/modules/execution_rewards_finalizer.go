package modules

import (
	"fmt"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type executionRewardsFinalizer struct {
	ModuleContext
	ExportMutex *sync.Mutex
	CooldownTs  time.Time
}

func NewExecutionRewardFinalizer(moduleContext ModuleContext) ModuleInterface {
	return &executionRewardsFinalizer{
		ModuleContext: moduleContext,
		ExportMutex:   &sync.Mutex{},
	}
}

func (d *executionRewardsFinalizer) Init() error {
	return nil
}

func (d *executionRewardsFinalizer) GetName() string {
	return "ExecutionRewards-Finalizer"
}

func (d *executionRewardsFinalizer) GetMonitoringEventId() constants.Event {
	return constants.Event_ExporterModuleELRewardsFinalizer
}

func (d *executionRewardsFinalizer) OnChainReorg(event *constypes.StandardEventChainReorg) (err error) {
	return nil // nop
}

func (d *executionRewardsFinalizer) OnFinalizedCheckpoint(event *constypes.StandardFinalizedCheckpointResponse) (err error) {
	return nil // nop
}

func (d *executionRewardsFinalizer) OnHead(event *constypes.StandardEventHeadResponse) (err error) {
	if time.Now().Before(d.CooldownTs) {
		log.Warnf("execution rewards finalizer is on cooldown till %s", d.CooldownTs)
	}
	// if mutex is locked, return early
	if !d.ExportMutex.TryLock() {
		log.Infof("execution rewards finalizer is already running")
		return nil
	}
	defer d.ExportMutex.Unlock()
	err = d.maintainTable()
	if err != nil {
		d.CooldownTs = time.Now().Add(1 * time.Minute)
		return fmt.Errorf("error maintaining table: %w", err)
	}
	return nil
}

func (d *executionRewardsFinalizer) maintainTable() (err error) {
	var lastExportedSlot int64
	err = db.ReaderDb.Get(&lastExportedSlot, `
		SELECT
			coalesce(MAX(slot), -1)
		FROM
			execution_rewards_finalized
	`)
	if err != nil {
		return fmt.Errorf("error getting last exported slot: %w", err)
	}
	// get latest finalized slot
	var latestFinalizedSlot int64
	err = db.ReaderDb.Get(&latestFinalizedSlot, `
		SELECT
			max(slot)
		FROM
			blocks
		WHERE
			status = '1' AND finalized = true
	`)

	if err != nil {
		return fmt.Errorf("error getting finalized-slot: %w", err)
	}

	// limit to prevent overloading
	// gnosis has a 5 second slot window, so to prevent hammering the db scale the batch size by the slot time
	batch := int64(10_000 * utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if latestFinalizedSlot-lastExportedSlot > batch {
		latestFinalizedSlot = lastExportedSlot + batch
	}

	if latestFinalizedSlot <= lastExportedSlot {
		log.Debugf("no new finalized slots to export")
		return nil
	}
	// sanity check, check if any non-missed block has a fee_recipient_reward that is NULL
	var count struct {
		Total   int64 `db:"total"`
		NonNull int64 `db:"non_null"`
	}
	gc := goqu.Dialect("postgres").From("blocks").
		Select(
			goqu.Func("count", goqu.Star()).As("total"),
			goqu.Func("count", goqu.I("ep.fee_recipient_reward")).As("non_null"),
		).
		LeftJoin(
			goqu.T("execution_payloads").As("ep"),
			goqu.On(goqu.I("ep.block_hash").Eq(goqu.I("blocks.exec_block_hash"))),
		).
		Where(
			goqu.I("slot").Gt(lastExportedSlot),
			goqu.I("slot").Lte(latestFinalizedSlot),
			goqu.I("status").Eq("1"),
		)
	query, args, err := gc.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	err = db.ReaderDb.Get(&count, query, args...)
	if err != nil {
		return fmt.Errorf("error getting count of non-missed blocks: %w", err)
	}
	if count.Total != count.NonNull {
		return fmt.Errorf("only %v out of %v blocks have non-null fee_recipient_reward", count.NonNull, count.Total)
	}
	log.Infof("finalized rewards = last exported slot: %v, latest finalized slot: %v", lastExportedSlot, latestFinalizedSlot)
	start := time.Now()
	ds := goqu.Dialect("postgres").Insert("execution_rewards_finalized").FromQuery(
		goqu.From(goqu.T("blocks").As("b")).
			LeftJoin(
				goqu.T("execution_payloads").As("ep"),
				goqu.On(goqu.I("ep.block_hash").Eq(goqu.I("b.exec_block_hash"))),
			).
			LeftJoin(
				goqu.T("relays_blocks").As("rb"),
				goqu.On(goqu.I("rb.exec_block_hash").Eq(goqu.I("b.exec_block_hash"))),
			).
			Select(
				goqu.I("b.epoch").As("epoch"),
				goqu.I("b.slot").As("slot"),
				goqu.I("b.proposer").As("proposer"),
				goqu.Func("max", goqu.COALESCE(goqu.I("rb.value"), goqu.L("ep.fee_recipient_reward * '1e18'::numeric"), goqu.L("0::numeric"))).As("value"),
			).
			Where(
				goqu.I("b.slot").Gt(lastExportedSlot),
				goqu.I("b.slot").Lte(latestFinalizedSlot),
				goqu.I("b.status").Eq("1"),
			).
			GroupBy(
				goqu.I("b.epoch"), goqu.I("b.slot"), goqu.I("b.proposer"),
			),
	).OnConflict(goqu.DoUpdate("slot", goqu.Record{
		"value":    goqu.I("excluded.value"),
		"proposer": goqu.I("excluded.proposer"),
	}))

	log.Debugf("writing execution rewards finalized data")

	query, args, err = ds.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	_, err = db.WriterDb.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error inserting data: %w", err)
	}
	log.Infof("execution rewards finalized data written in %v", time.Since(start))

	return nil
}
