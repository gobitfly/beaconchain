package modules

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/klauspost/pgzip"

	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
)

type slotExporter struct {
	ModuleContext
	Client         rpc.Client
	db             edb.SlotExporterRepository
	cache          edb.ExporterCache
	FirstRun       bool
	latestEpoch    uint64
	latestSlot     uint64
	finalizedEpoch uint64
	latestProposed uint64
}

func NewSlotExporter(moduleContext ModuleContext, client rpc.Client, cache edb.ExporterCache, db edb.SlotExporterRepository) ModuleInterface {
	return &slotExporter{
		ModuleContext:  moduleContext,
		Client:         client,
		db:             db,
		cache:          cache,
		FirstRun:       true,
		latestEpoch:    0,
		latestSlot:     0,
		finalizedEpoch: 0,
		latestProposed: 0,
	}
}

var processSlotMutex = &sync.Mutex{}

func (e *slotExporter) OnHead(_ *constypes.StandardEventHeadResponse) (err error) {
	if !processSlotMutex.TryLock() {
		log.Infof("slotExporter is still running, skipping this run")
		return nil
	}
	defer processSlotMutex.Unlock()

	// cache handling
	defer func() {
		if err == nil {
			latestEpoch, err := e.cache.GetLatestEpoch()
			if err != nil {
				log.Error(err, "error retrieving latestEpoch from cache", 0)
			}

			if e.latestEpoch > 0 && latestEpoch < e.latestEpoch {
				err := e.cache.SetLatestEpoch(e.latestEpoch)
				if err != nil {
					log.Error(err, "error setting latestEpoch in cache", 0)
				}
			}

			latestSlot, err := e.cache.GetLatestSlot()
			if err != nil {
				log.Error(err, "error retrieving latestSlot from cache", 0)
			}
			if e.latestSlot > 0 && latestSlot < e.latestSlot {
				err := e.cache.SetLatestSlot(e.latestSlot)
				if err != nil {
					log.Error(err, "error setting latestSlot in cache", 0)
				}
			}

			latestFinalizedEpoch, err := e.cache.GetLatestFinalizedEpoch()
			if err != nil {
				log.Error(err, "error retrieving latestFinalizedEpoch from cache", 0)
			}
			if e.finalizedEpoch > 0 && latestFinalizedEpoch < e.finalizedEpoch {
				err := e.cache.SetLatestFinalizedEpoch(e.finalizedEpoch)
				if err != nil {
					log.Error(err, "error setting latestFinalizedEpoch in cache", 0)
				}
			}

			latestProposedSlot, err := e.cache.GetLatestProposedSlot()
			if err != nil {
				log.Error(err, "error retrieving latestProposedSlot from cache", 0)
			}
			if e.latestProposed > 0 && latestProposedSlot < e.latestProposed {
				err := e.cache.SetLatestProposedSlot(e.latestProposed)
				if err != nil {
					log.Error(err, "error setting latestProposedSlot in cache", 0)
				}
			}
		}
	}()

	// get the current chain head
	head, err := e.Client.GetChainHead()
	if err != nil {
		return fmt.Errorf("error retrieving chain head: %w", err)
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting tx: %w", err)
	}
	defer utils.Rollback(tx)

	if e.FirstRun {
		if err := e.handleFirstRun(head, tx); err != nil {
			return err
		}
		e.FirstRun = false
	}

	// at this point we know that we have a coherent list of slots in the database without any gaps

	if err := e.exportNewSlots(head, tx); err != nil {
		return err
	}

	// at this point we have all data up to the current chain head in the database

	if err := e.handleFinalizedSlots(head, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing tx: %w", err)
	}

	e.latestEpoch = utils.EpochOfSlot(head.HeadSlot)
	e.latestSlot = head.HeadSlot

	services.ReportStatus("slotExporter", "Running", nil)

	return nil
}

func (e *slotExporter) handleFirstRun(head *types.ChainHead, tx *sqlx.Tx) error {
	// get all slots we currently have in the database
	dbSlots, err := e.db.GetAllSlots(tx)
	if err != nil {
		return fmt.Errorf("error retrieving all db slots: %w", err)
	}

	if len(dbSlots) > 0 {
		if dbSlots[0] != 0 {
			log.Infof("exporting genesis slot as it is missing in the database")
			slotExp := NewSlotExport(e.Client, 0, e.latestProposed, utils.EpochOfSlot(0) == head.HeadEpoch, tx)
			err := slotExp.ExportSlot()
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", 0, err)
			}
			dbSlots, err = e.db.GetAllSlots(tx)
			if err != nil {
				return fmt.Errorf("error retrieving all db slots: %w", err)
			}
		}
	}

	if len(dbSlots) > 1 {
		// export any gaps we might have (for whatever reason)
		for slotIndex := 1; slotIndex < len(dbSlots); slotIndex++ {
			previousSlot := dbSlots[slotIndex-1]
			currentSlot := dbSlots[slotIndex]

			if previousSlot != currentSlot-1 {
				log.Infof("slots between %v and %v are missing, exporting them", previousSlot, currentSlot)
				for slot := previousSlot + 1; slot <= currentSlot-1; slot++ {
					slotExp := NewSlotExport(e.Client, slot, e.latestProposed, false, tx)
					err := slotExp.ExportSlot()
					if err != nil {
						return fmt.Errorf("error exporting slot %v: %w", slot, err)
					}
				}
			}
		}
	}
	return nil
}

func (e *slotExporter) exportNewSlots(head *types.ChainHead, tx *sqlx.Tx) error {
	lastDbSlot, err := e.db.GetLastSlot(tx)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Infof("db is empty, export genesis slot")
			slotExp := NewSlotExport(e.Client, 0, e.latestProposed, utils.EpochOfSlot(0) == head.HeadEpoch, tx)
			err := slotExp.ExportSlot()
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", 0, err)
			}
			lastDbSlot = 0
		} else {
			return fmt.Errorf("error retrieving last slot from the db: %w", err)
		}
	}

	// check if any new slots have been added to the chain
	if lastDbSlot != head.HeadSlot {
		slotsExported := 0
		for slot := lastDbSlot + 1; slot <= head.HeadSlot; slot++ { // export any new slots
			slotExp := NewSlotExport(e.Client, slot, e.latestProposed, utils.EpochOfSlot(slot) == head.HeadEpoch, tx)
			err = slotExp.ExportSlot()
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", slot, err)
			}
			slotsExported++

			// in case of large export runs, export at most 10 epochs per tx
			if slotsExported == int(utils.Config.Chain.ClConfig.SlotsPerEpoch)*10 {
				err := tx.Commit()
				if err != nil {
					return fmt.Errorf("error committing tx: %w", err)
				}

				e.latestEpoch = utils.EpochOfSlot(slot)
				e.latestSlot = slot

				return nil
			}
		}
	}

	return nil
}

func (e *slotExporter) handleFinalizedSlots(head *types.ChainHead, tx *sqlx.Tx) error {
	// check if any non-finalized slot has changed by comparing it with the node
	dbNonFinalSlots, err := e.db.GetAllNonFinalizedSlots()
	if err != nil {
		return fmt.Errorf("error retrieving all non finalized slots from the db: %w", err)
	}

	for _, dbSlot := range dbNonFinalSlots {
		header, err := e.Client.GetBlockHeader(dbSlot.Slot)
		if err != nil {
			return fmt.Errorf("error retrieving block root for slot %v: %w", dbSlot.Slot, err)
		}

		nodeSlotFinalized := dbSlot.Slot <= head.FinalizedSlot

		if nodeSlotFinalized != dbSlot.Finalized {
			log.Infof("checking slot %d for finalization / reorgs", dbSlot.Slot)
			header, err = e.Client.GetBlockHeader(dbSlot.Slot)

			if err != nil {
				return fmt.Errorf("error retrieving block root for slot %v: %w", dbSlot.Slot, err)
			}

			// slot has finalized, mark it in the db
			if header != nil && bytes.Equal(dbSlot.BlockRoot, header.Data.Root) {
				// no reorg happened, simply mark the slot as final
				log.Infof("setting slot %v as finalized (proposed)", dbSlot.Slot)
				err := e.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, dbSlot.Status, tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (proposed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) < 32 {
				// no reorg happened, mark the slot as missed
				log.Infof("setting slot %v as finalized (missed)", dbSlot.Slot)
				err := e.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "2", tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (missed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) == 32 {
				// slot has been orphaned, mark the slot as orphaned
				log.Infof("setting slot %v as finalized (orphaned)", dbSlot.Slot)
				err := e.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
			} else if header != nil && !bytes.Equal(header.Data.Root, dbSlot.BlockRoot) {
				// we have a different block root for the slot in the db, mark the currently present one as orphaned and write the new one
				log.Infof("setting slot %v as orphaned and exporting new slot", dbSlot.Slot)
				err := e.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
				slotExp := NewSlotExport(e.Client, dbSlot.Slot, e.latestProposed, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, tx)
				err = slotExp.ExportSlot()
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}

			// epoch transition slot has finalized, update epoch status
			if dbSlot.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 && dbSlot.Slot > utils.Config.Chain.ClConfig.SlotsPerEpoch-1 {
				epoch := utils.EpochOfSlot(dbSlot.Slot)
				epochParticipationStats, err := e.Client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				} else {
					log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
					err := e.db.UpdateEpochStatus(epochParticipationStats, tx)
					if epochParticipationStats.Finalized && epochParticipationStats.Epoch > e.finalizedEpoch {
						e.finalizedEpoch = epochParticipationStats.Epoch
					}

					if err != nil {
						return err
					}

					log.Infof("exporting validation queue")
					queue, err := e.Client.GetValidatorQueue()
					if err != nil {
						return fmt.Errorf("error retrieving validator queue data: %w", err)
					}

					err = e.db.SaveValidatorQueue(queue, tx)
					if err != nil {
						return fmt.Errorf("error saving validator queue data: %w", err)
					}
				}
			}
		} else { // check if a late slot has been proposed in the meantime
			if len(dbSlot.BlockRoot) < 32 && header != nil { // we have no slot in the db, but the node has a slot, export it
				log.Infof("exporting new slot %v", dbSlot.Slot)
				slotExp := NewSlotExport(e.Client, dbSlot.Slot, e.latestProposed, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, tx)
				err := slotExp.ExportSlot()
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}
		}
	}

	return nil
}

type exporter struct {
	Client rpc.Client
	db     edb.SlotExporterRepository
	dbTx   *sqlx.Tx

	slot           uint64
	latestProposed uint64
	headEpoch      bool
}

func NewSlotExport(client rpc.Client, slot, latestProposed uint64, headEpoch bool, dbTx *sqlx.Tx) *exporter {
	return &exporter{
		Client:         client,
		dbTx:           dbTx,
		slot:           slot,
		latestProposed: latestProposed,
		headEpoch:      headEpoch,
	}
}

func (s *exporter) ExportSlot() error {
	isFirstSlotOfEpoch := s.slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0
	epoch := s.slot / utils.Config.Chain.ClConfig.SlotsPerEpoch

	if isFirstSlotOfEpoch {
		log.Infof("exporting slot %v (epoch transition into epoch %v)", s.slot, epoch)
	} else {
		log.Infof("exporting slot %v", s.slot)
	}
	start := time.Now()

	// retrieve the data for the slot from the node
	// the first slot of an epoch will also contain all validator duties for the whole epoch
	block, err := s.Client.GetBlockBySlot(s.slot)
	if err != nil {
		return fmt.Errorf("error retrieving data for slot %v: %w", s.slot, err)
	}

	// for the slot itself start by preparing the duties for export to bigtable
	if err := s.exportDuties(block); err != nil {
		return err
	}

	// save the block data to the db
	if err := s.db.SaveBlock(block, false, s.dbTx); err != nil {
		return fmt.Errorf("error saving slot to the db: %w", err)
	}

	if block.Status == 1 {
		if s.latestProposed < block.Slot {
			s.latestProposed = block.Slot
		}
	}

	if block.EpochAssignments != nil { // export the epoch assignments as they are included in the first slot of an epoch
		if err := s.exportEpochAssignments(s.Client, block, s.headEpoch, s.dbTx); err != nil {
			return err
		}
	}

	log.InfoWithFields(
		log.Fields{
			"slot":      block.Slot,
			"blockRoot": fmt.Sprintf("%x", block.BlockRoot),
			"duration":  time.Since(start),
		}, "! export of slot completed")

	return nil
}

func (s *exporter) exportDuties(block *types.Block) error {
	syncDuties := make(map[types.Slot]map[types.ValidatorIndex]bool)
	syncDuties[types.Slot(block.Slot)] = make(map[types.ValidatorIndex]bool)

	for validator, duty := range block.SyncDuties {
		syncDuties[types.Slot(block.Slot)][validator] = duty
	}

	attDuties := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
	for validator, attestedSlots := range block.AttestationDuties {
		for _, attestedSlot := range attestedSlots {
			if attDuties[attestedSlot] == nil {
				attDuties[attestedSlot] = make(map[types.ValidatorIndex][]types.Slot)
			}
			if attDuties[attestedSlot][validator] == nil {
				attDuties[attestedSlot][validator] = make([]types.Slot, 0, 10)
			}
			attDuties[attestedSlot][validator] = append(attDuties[attestedSlot][validator], types.Slot(block.Slot))
		}
	}

	// save sync & attestation duties to bigtable
	err := db.BigtableClient.SaveAttestationDuties(attDuties)
	if err != nil {
		return fmt.Errorf("error exporting attestations to bigtable for slot %v: %w", block.Slot, err)
	}
	err = db.BigtableClient.SaveSyncCommitteeDuties(syncDuties)
	if err != nil {
		return fmt.Errorf("error exporting sync committee duties to bigtable for slot %v: %w", block.Slot, err)
	}

	return nil
}

func (s *exporter) exportEpochAssignments(client rpc.Client, block *types.Block, isHeadEpoch bool, tx *sqlx.Tx) error {
	epoch := utils.EpochOfSlot(block.Slot)

	log.Infof("exporting duties & balances for epoch %v", epoch)

	// prepare the duties for export to bigtable
	syncDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex]bool)
	for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
		if syncDutiesEpoch[types.Slot(slot)] == nil {
			syncDutiesEpoch[types.Slot(slot)] = make(map[types.ValidatorIndex]bool)
		}
		for _, validatorIndex := range block.EpochAssignments.SyncAssignments {
			syncDutiesEpoch[types.Slot(slot)][types.ValidatorIndex(validatorIndex)] = false
		}
	}

	attDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
	for key, validatorIndex := range block.EpochAssignments.AttestorAssignments {
		keySplit := strings.Split(key, "-")
		attestedSlot, err := strconv.ParseUint(keySplit[0], 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing attested slot from attestation key: %w", err)
		}

		if attDutiesEpoch[types.Slot(attestedSlot)] == nil {
			attDutiesEpoch[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
		}

		attDutiesEpoch[types.Slot(attestedSlot)][types.ValidatorIndex(validatorIndex)] = []types.Slot{}
	}

	g := errgroup.Group{}

	// store epoch assignments in redis
	g.Go(func() error {
		return s.saveEpochAssignmentsToRedis(client, block, epoch, isHeadEpoch)
	})

	// save attestation duties to bigtable
	g.Go(func() error {
		err := db.BigtableClient.SaveAttestationDuties(attDutiesEpoch)
		if err != nil {
			return fmt.Errorf("error exporting attestation assignments to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})

	// save sync committee duties to bigtable
	g.Go(func() error {
		err := db.BigtableClient.SaveSyncCommitteeDuties(syncDutiesEpoch)
		if err != nil {
			return fmt.Errorf("error exporting sync committee assignments to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})

	// save the validator balances to bigtable
	g.Go(func() error {
		return db.BigtableClient.SaveValidatorBalances(epoch, block.Validators)
	})

	// if we are exporting the head epoch, update the validator db table
	if isHeadEpoch {
		if err := s.exportValidatorData(client, block, epoch, tx); err != nil {
			return err
		}
	}

	var epochParticipationStats *types.ValidatorParticipation
	if epoch > 0 {
		g.Go(func() error {
			// retrieve the epoch participation stats
			var err error
			epochParticipationStats, err = client.GetValidatorParticipation(epoch - 1)
			if err != nil {
				return fmt.Errorf("error retrieving epoch participation statistics: %w", err)
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}

	// save the epoch metadata to the database
	err = s.db.SaveEpoch(epoch, block.Validators, client, tx)
	if err != nil {
		return fmt.Errorf("error saving epoch data: %w", err)
	}

	if epoch > 0 && epochParticipationStats != nil {
		log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
		err := s.db.UpdateEpochStatus(epochParticipationStats, tx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *exporter) saveEpochAssignmentsToRedis(client rpc.Client, block *types.Block, epoch uint64, isHeadEpoch bool) error {
	redisCachedEpochAssignments := &types.RedisCachedEpochAssignments{
		Epoch:       types.Epoch(epoch),
		Assignments: block.EpochAssignments,
	}

	var serializedAssignmentsData bytes.Buffer
	enc := gob.NewEncoder(&serializedAssignmentsData)
	err := enc.Encode(redisCachedEpochAssignments)
	if err != nil {
		return fmt.Errorf("error serializing assignments to gob for slot %v: %w", block.Slot, err)
	}

	key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "ea", epoch)
	expirationTime := utils.EpochToTime(epoch + 7) // keep it for at least 7 epochs in the cache
	expirationDuration := time.Until(expirationTime)

	if expirationDuration.Seconds() < 0 || expirationDuration.Hours() > 2 {
		log.Warnf("NOT writing assignments data for epoch %v to redis because a TTL < 0 or TTL > 2h: %v", epoch, expirationDuration)
	} else {
		log.Infof("writing assignments data for epoch %v to redis with a TTL of %v", epoch, expirationDuration)
		err = db.PersistentRedisDbClient.Set(context.Background(), key, serializedAssignmentsData.Bytes(), expirationDuration).Err()
		if err != nil {
			return fmt.Errorf("error writing assignments data to redis for epoch %v: %w", epoch, err)
		}
		// publish the event to inform the api about the new data (todo)
		// db.PersistentRedisDbClient.Publish(context.Background(), fmt.Sprintf("%d:slotViz", utils.Config.Chain.ClConfig.DepositChainID), fmt.Sprintf("%s:%d", "ea", epoch)).Err()
		log.Infof("writing current epoch assignments to redis completed")
	}

	if isHeadEpoch {
		nextEpoch := epoch + 1
		nextEpochAssignments, err := client.GetEpochAssignments(nextEpoch)
		if err != nil {
			return fmt.Errorf("error retrieving epoch assignments for head+1 epoch: %v", err)
		}

		redisCachedNextEpochAssignments := &types.RedisCachedEpochAssignments{
			Epoch:       types.Epoch(nextEpoch),
			Assignments: nextEpochAssignments,
		}

		var serializedAssignmentsData bytes.Buffer
		enc := gob.NewEncoder(&serializedAssignmentsData)
		err = enc.Encode(redisCachedNextEpochAssignments)
		if err != nil {
			return fmt.Errorf("error serializing assignments to gob for head+1 epoch %v: %w", block.Slot, err)
		}

		key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "ea", nextEpoch)
		expirationTime := utils.EpochToTime(nextEpoch + 7) // keep it for at least 7 epochs in the cache
		expirationDuration := time.Until(expirationTime)

		if expirationDuration.Seconds() < 0 || expirationDuration.Hours() > 2 {
			log.Warnf("NOT writing assignments data for head+1 epoch (%v) to redis because a TTL < 0 or TTL > 2h: %v", nextEpoch, expirationDuration)
		} else {
			log.Infof("writing assignments data for head+1 epoch (%v) to redis with a TTL of %v", nextEpoch, expirationDuration)
			err = db.PersistentRedisDbClient.Set(context.Background(), key, serializedAssignmentsData.Bytes(), expirationDuration).Err()
			if err != nil {
				return fmt.Errorf("error writing assignments data for head+1 epoch to redis for epoch %v: %w", nextEpoch, err)
			}
		}
	}

	return nil
}

func (s *exporter) exportValidatorData(client rpc.Client, block *types.Block, epoch uint64, tx *sqlx.Tx) error {
	g := errgroup.Group{}

	// this function sets exports the validator status into the db
	// and also updates the status field in the validators array
	err := s.db.SaveValidators(epoch, block.Validators, client, 10000, tx)
	if err != nil {
		return fmt.Errorf("error saving validators for epoch %v: %w", epoch, err)
	}

	// also update the queue deposit table once every epoch
	g.Go(func() error {
		err = s.db.UpdateQueueDeposits(tx)
		if err != nil {
			return fmt.Errorf("error updating queue deposits cache: %w", err)
		}
		return nil
	})

	// store validator mapping in redis
	g.Go(func() error {
		// generate mapping
		RedisCachedValidatorsMapping := &types.RedisCachedValidatorsMapping{
			Epoch:   types.Epoch(epoch),
			Mapping: make([]*types.CachedValidator, len(block.Validators)),
		}

		activationMapping := make(map[int][]uint64)
		start := time.Now()

		for _, v := range block.Validators {
			r := types.CachedValidator{
				PublicKey:             v.PublicKey,
				Status:                v.Status,
				WithdrawalCredentials: v.WithdrawalCredentials,
				Balance:               v.Balance,
				EffectiveBalance:      v.EffectiveBalance,
				Slashed:               v.Slashed,
			}
			if v.ActivationEpoch != edb.MaxSqlNumber {
				r.ActivationEpoch = sql.NullInt64{Int64: int64(v.ActivationEpoch), Valid: true}
			}
			if v.ActivationEligibilityEpoch != edb.MaxSqlNumber {
				r.ActivationEligibilityEpoch = sql.NullInt64{Int64: int64(v.ActivationEligibilityEpoch), Valid: true}
			}
			if v.ExitEpoch != edb.MaxSqlNumber {
				r.ExitEpoch = sql.NullInt64{Int64: int64(v.ExitEpoch), Valid: true}
			}
			if v.WithdrawableEpoch != edb.MaxSqlNumber {
				r.WithdrawableEpoch = sql.NullInt64{Int64: int64(v.WithdrawableEpoch), Valid: true}
			}
			RedisCachedValidatorsMapping.Mapping[v.Index] = &r
			if v.Status == "pending" {
				a := int(v.ActivationEligibilityEpoch)
				activationMapping[a] = append(activationMapping[a], v.Index)
			}
		}
		log.Debugf("filled validator mapping, took: %s", time.Since(start))

		start = time.Now()
		// need to sort as activations don't necessarily have to be in order
		keys := maps.Keys(activationMapping)
		sort.Ints(keys)
		var i int64
		for _, a := range keys {
			// don't need to sort as we our validator array is indeed in order
			for _, vi := range activationMapping[a] {
				RedisCachedValidatorsMapping.Mapping[vi].Queues.ActivationIndex = sql.NullInt64{Int64: i, Valid: true}
				i++
			}
		}
		log.Debugf("calculated activation queue indexes, took: %s", time.Since(start))

		// gob struct
		start = time.Now()
		var serializedValidatorMapping bytes.Buffer
		enc := gob.NewEncoder(&serializedValidatorMapping)
		err := enc.Encode(RedisCachedValidatorsMapping)
		if err != nil {
			return fmt.Errorf("error serializing validator mapping to gob for epoch %v: %w", epoch, err)
		}
		log.Debugf("encoding validator mapping into gob took %s", time.Since(start))

		// compress using pgzip
		start = time.Now()
		var compressedValidatorMapping bytes.Buffer
		w, err := pgzip.NewWriterLevel(&compressedValidatorMapping, pgzip.BestCompression)
		if err != nil {
			return fmt.Errorf("failed to create pgzip writer for epoch %v: %w", epoch, err)
		}
		err = w.SetConcurrency(500_000, 10)
		if err != nil {
			return fmt.Errorf("failed to set concurrency for pgzip writer for epoch %v: %w", epoch, err)
		}
		_, err = w.Write(serializedValidatorMapping.Bytes())
		if err != nil {
			return fmt.Errorf("error decompressing validator mapping using pgzip for epoch %v: %w", epoch, err)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("error closing pgzip writer for epoch %v: %w", epoch, err)
		}
		log.Debugf("compressing validator mapping using pgzip took %s", time.Since(start))

		// load into redis
		start = time.Now()
		key := fmt.Sprintf("%d:%s", utils.Config.Chain.ClConfig.DepositChainID, "vm")
		log.Infof("writing validator mappping to redis with no TTL")
		err = db.PersistentRedisDbClient.Set(context.Background(), key, compressedValidatorMapping.Bytes(), 0).Err()
		if err != nil {
			return fmt.Errorf("error writing validator mapping to redis for epoch %v: %w", epoch, err)
		}
		log.Infof("writing validator mapping to redis done, took %s", time.Since(start))
		return nil
	})

	// update cached view of consensus desposits
	// possible bug: at this point the export tx is not yet committed, so the query will read
	// stale data
	g.Go(func() error {
		start := time.Now()
		err := s.db.CacheBlockDepositLookup()
		if err != nil {
			return fmt.Errorf("error updating cached view of consensus deposits: %w", err)
		}
		log.Infof("updating cached view of consensus deposits took %s", time.Since(start))
		return nil
	})

	return g.Wait()
}

func (d *slotExporter) Init() error {
	return nil
}

func (d *slotExporter) GetName() string {
	return "Slot-Exporter"
}

func (d *slotExporter) GetMonitoringEventId() constants.Event {
	return constants.Event_ExporterModuleSlotExporter
}

func (d *slotExporter) OnChainReorg(event *constypes.StandardEventChainReorg) (err error) {
	return nil // nop
}

func (d *slotExporter) OnFinalizedCheckpoint(event *constypes.StandardFinalizedCheckpointResponse) (err error) {
	return nil // nop
}
