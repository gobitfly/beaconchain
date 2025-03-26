package modules

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/klauspost/pgzip"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

type SlotExporterClient interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetBlockBySlot(slot uint64) (*types.Block, error)
	GetValidatorQueue() (*types.ValidatorQueue, error)
	GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
	GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error)
}

type slotExporter struct {
	ModuleContext
	Client SlotExporterClient
	cache  edb.SlotExporterCacheRepository
	db     edb.SlotExporterDBRepository
	bt     edb.SlotExporterBTRepository

	firstRun       bool
	latestEpoch    uint64
	latestSlot     uint64
	finalizedEpoch uint64
	latestProposed uint64
}

func NewSlotExporter(moduleContext ModuleContext, cache edb.SlotExporterCacheRepository, db edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository) ModuleInterface {
	return &slotExporter{
		ModuleContext:  moduleContext,
		Client:         moduleContext.ConsClient,
		cache:          cache,
		db:             db,
		bt:             bt,
		firstRun:       true,
		latestEpoch:    0,
		latestSlot:     0,
		finalizedEpoch: 0,
		latestProposed: 0,
	}
}

var processSlotMutex = &sync.Mutex{}

func (s *slotExporter) OnHead(_ *constypes.StandardEventHeadResponse) (err error) {
	if !processSlotMutex.TryLock() {
		log.Infof("slotExporter is still running, skipping this run")
		return nil
	}
	defer processSlotMutex.Unlock()

	// cache handling
	defer func() {
		if err == nil {
			chainID := utils.Config.Chain.ClConfig.DepositChainID

			latestEpoch, err := s.cache.GetLatestEpoch(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestEpoch from cache", 0)
			}

			if s.latestEpoch > 0 && latestEpoch < s.latestEpoch {
				err := s.cache.SetLatestEpoch(chainID, s.latestEpoch)
				if err != nil {
					log.Error(err, "error setting latestEpoch in cache", 0)
				}
			}

			latestSlot, err := s.cache.GetLatestSlot(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestSlot from cache", 0)
			}
			if s.latestSlot > 0 && latestSlot < s.latestSlot {
				err := s.cache.SetLatestSlot(chainID, s.latestSlot)
				if err != nil {
					log.Error(err, "error setting latestSlot in cache", 0)
				}
			}

			latestFinalizedEpoch, err := s.cache.GetLatestFinalizedEpoch(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestFinalizedEpoch from cache", 0)
			}
			if s.finalizedEpoch > 0 && latestFinalizedEpoch < s.finalizedEpoch {
				err := s.cache.SetLatestFinalizedEpoch(chainID, s.finalizedEpoch)
				if err != nil {
					log.Error(err, "error setting latestFinalizedEpoch in cache", 0)
				}
			}

			latestProposedSlot, err := s.cache.GetLatestProposedSlot(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestProposedSlot from cache", 0)
			}
			if s.latestProposed > 0 && latestProposedSlot < s.latestProposed {
				err := s.cache.SetLatestProposedSlot(chainID, s.latestProposed)
				if err != nil {
					log.Error(err, "error setting latestProposedSlot in cache", 0)
				}
			}
		}
	}()

	// get the current chain head
	head, err := s.Client.GetChainHead()
	if err != nil {
		return fmt.Errorf("error retrieving chain head: %w", err)
	}

	tx, err := s.db.BeginTx()
	if err != nil {
		return fmt.Errorf("error starting tx: %w", err)
	}
	defer s.db.RollbackTx(tx)

	exporter := NewExporter(s.Client, s.cache, s.db, s.bt, metrics.NewMetricsCollector(), tx, s)

	if s.firstRun {
		log.Infof("performing first run consistency checks")
		if err := s.handleFirstRun(head, exporter, tx); err != nil {
			return err
		}
		s.firstRun = false
	}

	// at this point we know that we have a coherent list of slots in the database without any gaps

	if err := s.exportNewSlots(head, exporter, tx); err != nil {
		return err
	}

	// at this point we have all data up to the current chain head in the database

	if err := s.handleFinalizedSlots(head, exporter, tx); err != nil {
		return err
	}

	if err := s.db.CommitTx(tx); err != nil {
		return fmt.Errorf("error committing tx: %w", err)
	}

	s.latestEpoch = utils.EpochOfSlot(head.HeadSlot)
	s.latestSlot = head.HeadSlot

	services.ReportStatus("slotExporter", "Running", nil)

	return nil
}

func (s *slotExporter) handleFirstRun(head *types.ChainHead, exporter *exporter, tx *sqlx.Tx) error {
	// get all slots we currently have in the database
	dbSlots, err := s.db.GetAllSlots(tx)
	if err != nil {
		return fmt.Errorf("error retrieving all db slots: %w", err)
	}
	log.Info("retrieved all exported slots from the database")

	if len(dbSlots) > 0 {
		if dbSlots[0] != 0 {
			log.Infof("exporting genesis slot as it is missing in the database")
			isHeadEpoch := utils.EpochOfSlot(0) == head.HeadEpoch
			err := exporter.ExportSlot(0, isHeadEpoch)
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", 0, err)
			}
			dbSlots, err = s.db.GetAllSlots(tx)
			if err != nil {
				return fmt.Errorf("error retrieving all db slots: %w", err)
			}
		}
	}

	if len(dbSlots) > 1 {
		log.Info("performing gap checks")
		// export any gaps we might have (for whatever reason)
		for slotIndex := 1; slotIndex < len(dbSlots); slotIndex++ {
			previousSlot := dbSlots[slotIndex-1]
			currentSlot := dbSlots[slotIndex]

			if previousSlot != currentSlot-1 {
				log.Infof("slots between %v and %v are missing, exporting them", previousSlot, currentSlot)
				for slot := previousSlot + 1; slot <= currentSlot-1; slot++ {
					err := exporter.ExportSlot(slot, false)
					if err != nil {
						return fmt.Errorf("error exporting slot %v: %w", slot, err)
					}
				}
			}
		}
	}
	return nil
}

func (s *slotExporter) exportNewSlots(head *types.ChainHead, exporter *exporter, tx *sqlx.Tx) error {
	lastDbSlot, err := s.db.GetLastSlot(tx)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Infof("db is empty, export genesis slot")
			isHeadEpoch := utils.EpochOfSlot(0) == head.HeadEpoch
			err := exporter.ExportSlot(0, isHeadEpoch)
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
			isHeadEpoch := utils.EpochOfSlot(slot) == head.HeadEpoch
			err = exporter.ExportSlot(slot, isHeadEpoch)
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

				s.latestEpoch = utils.EpochOfSlot(slot)
				s.latestSlot = slot

				return nil
			}
		}
	}

	return nil
}

func (s *slotExporter) handleFinalizedSlots(head *types.ChainHead, exporter *exporter, tx *sqlx.Tx) error {
	// check if any non-finalized slot has changed by comparing it with the node
	dbNonFinalSlots, err := s.db.GetAllNonFinalizedSlots()
	if err != nil {
		return fmt.Errorf("error retrieving all non finalized slots from the db: %w", err)
	}

	for _, dbSlot := range dbNonFinalSlots {
		header, err := s.Client.GetBlockHeader(dbSlot.Slot)
		if err != nil {
			return fmt.Errorf("error retrieving block root for slot %v: %w", dbSlot.Slot, err)
		}

		nodeSlotFinalized := dbSlot.Slot <= head.FinalizedSlot

		if nodeSlotFinalized != dbSlot.Finalized {
			log.Infof("checking slot %d for finalization / reorgs", dbSlot.Slot)

			// slot has finalized, mark it in the db
			if header != nil && bytes.Equal(dbSlot.BlockRoot, header.Data.Root) {
				// no reorg happened, simply mark the slot as final
				log.Infof("setting slot %v as finalized (proposed)", dbSlot.Slot)
				err := s.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, dbSlot.Status, tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (proposed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) < 32 {
				// no reorg happened, mark the slot as missed
				log.Infof("setting slot %v as finalized (missed)", dbSlot.Slot)
				err := s.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "2", tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (missed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) == 32 {
				// slot has been orphaned, mark the slot as orphaned
				log.Infof("setting slot %v as finalized (orphaned)", dbSlot.Slot)
				err := s.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
			} else if header != nil && !bytes.Equal(header.Data.Root, dbSlot.BlockRoot) {
				// we have a different block root for the slot in the db, mark the currently present one as orphaned and write the new one
				log.Infof("setting slot %v as orphaned and exporting new slot", dbSlot.Slot)
				err := s.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
				isHeadEpoch := utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch
				err = exporter.ExportSlot(dbSlot.Slot, isHeadEpoch)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}

			// epoch transition slot has finalized, update epoch status
			if dbSlot.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 && dbSlot.Slot > utils.Config.Chain.ClConfig.SlotsPerEpoch-1 {
				epoch := utils.EpochOfSlot(dbSlot.Slot)
				epochParticipationStats, err := s.Client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				} else {
					log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
					err := s.db.UpdateEpochStatus(epochParticipationStats, tx)
					if epochParticipationStats.Finalized && epochParticipationStats.Epoch > s.finalizedEpoch {
						s.finalizedEpoch = epochParticipationStats.Epoch
					}

					if err != nil {
						return err
					}

					log.Infof("exporting validation queue")
					queue, err := s.Client.GetValidatorQueue()
					if err != nil {
						return fmt.Errorf("error retrieving validator queue data: %w", err)
					}

					err = s.db.SaveValidatorQueue(queue, tx)
					if err != nil {
						return fmt.Errorf("error saving validator queue data: %w", err)
					}
				}
			}
		} else { // check if a late slot has been proposed in the meantime
			// TODO: reenable once holesky is close to recovery
			if utils.Config.Chain.Id != 17000 {
				if len(dbSlot.BlockRoot) < 32 && header != nil { // we have no slot in the db, but the node has a slot, export it
					log.Infof("exporting new slot %v", dbSlot.Slot)
					isHeadEpoch := utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch
					err := exporter.ExportSlot(dbSlot.Slot, isHeadEpoch)
					if err != nil {
						return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
					}
				}
			}
		}
	}

	return nil
}

type exporter struct {
	Client  SlotExporterClient
	cache   edb.SlotExporterCacheRepository
	db      edb.SlotExporterDBRepository
	bt      edb.SlotExporterBTRepository
	metrics metrics.MetricsRepository
	dbTx    *sqlx.Tx

	slotExporter *slotExporter
}

func NewExporter(client SlotExporterClient, cache edb.SlotExporterCacheRepository, db edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository, metrics metrics.MetricsRepository, dbTx *sqlx.Tx, slotExporter *slotExporter) *exporter {
	return &exporter{
		Client:       client,
		cache:        cache,
		db:           db,
		bt:           bt,
		metrics:      metrics,
		dbTx:         dbTx,
		slotExporter: slotExporter,
	}
}

func (s *exporter) ExportSlot(slot uint64, headEpoch bool) error {
	isFirstSlotOfEpoch := slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0
	epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch

	if isFirstSlotOfEpoch {
		log.Infof("exporting slot %v (epoch transition into epoch %v)", slot, epoch)
	} else {
		log.Infof("exporting slot %v", slot)
	}
	start := time.Now()

	// retrieve the data for the slot from the node
	// the first slot of an epoch will also contain all validator duties for the whole epoch
	block, err := s.Client.GetBlockBySlot(slot)
	if err != nil {
		return fmt.Errorf("error retrieving data for slot %v: %w", slot, err)
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
		if s.slotExporter != nil {
			if s.slotExporter.latestProposed < block.Slot {
				s.slotExporter.latestProposed = block.Slot
			}
		}
	}
	s.metrics.ObserveTaskDuration("slot_exporter_export_slot", time.Since(start))

	if block.EpochAssignments != nil { // export the epoch assignments as they are included in the first slot of an epoch
		if err := s.exportEpochAssignments(block, headEpoch); err != nil {
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
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		s.metrics.ObserveTaskDuration("slot_exporter_export_duties", time.Since(timeStart))
	}(timeStart)

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

	g := errgroup.Group{}

	// save sync & attestation duties to bigtable
	g.Go(func() error {
		err := s.bt.SaveAttestationDuties(attDuties)
		if err != nil {
			return fmt.Errorf("error exporting attestations to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})
	g.Go(func() error {
		err := s.bt.SaveSyncCommitteeDuties(syncDuties)
		if err != nil {
			return fmt.Errorf("error exporting sync committee duties to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *exporter) exportEpochAssignments(block *types.Block, isHeadEpoch bool) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		s.metrics.ObserveTaskDuration("slot_exporter_export_epoch", time.Since(timeStart))
	}(timeStart)

	epoch := utils.EpochOfSlot(block.Slot)
	chainID := utils.Config.Chain.ClConfig.DepositChainID

	if epoch > utils.Config.ClConfig.ElectraForkEpoch {
		log.Infof("checking that events have been loaded for epoch %v", epoch)
		firstSlot := (epoch - 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
		lastSlot := (epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1

		exported, err := s.db.HasEventsForEpoch(firstSlot, lastSlot)
		if err != nil {
			return fmt.Errorf("error retrieving events for epoch %v: %w", epoch, err)
		}
		if !exported {
			return fmt.Errorf("events for epoch %v have not been loaded yet", epoch)
			// log.Infof("ERROR: events for epoch %v have not been loaded yet, RE-EXPORT events manually!!!", epoch)
		} else {
			log.Infof("events for epoch %v have been loaded, transforming consolidations & deposits", epoch)

			firstSlot := (epoch - 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
			lastSlot := (epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1

			switchToCompoundingRequestsProcessed, err := s.db.TransformSwitchToCompoundingRequests(firstSlot, lastSlot, s.dbTx)
			if err != nil {
				return fmt.Errorf("error transforming consolidation requests for epoch %v: %w", epoch, err)
			}
			log.Infof("transformed switch to compounding requests for epoch %v, processed %d requests", epoch, switchToCompoundingRequestsProcessed)

			consolidationRequestsProcessed, err := s.db.TransformConsolidationRequests(firstSlot, lastSlot, s.dbTx)
			if err != nil {
				return fmt.Errorf("error transforming consolidation requests for epoch %v: %w", epoch, err)
			}
			log.Infof("transformed consolidations for epoch %v, processed %d requests", epoch, consolidationRequestsProcessed)

			depositRequestsProcessed, err := s.db.TransformDepositRequests(firstSlot, lastSlot, s.dbTx)
			if err != nil {
				return fmt.Errorf("error transforming deposit requests for epoch %v: %w", epoch, err)
			}
			log.Infof("transformed deposits for epoch %v, processed %d requests", epoch, depositRequestsProcessed)

			removedExcessBalanceProcessed, err := s.db.TransformRemovedExcessBalanceEvents(firstSlot, lastSlot, s.dbTx)
			if err != nil {
				return fmt.Errorf("error transforming removed excess balance events for epoch %v: %w", epoch, err)
			}
			log.Infof("transformed removed excess balance events for epoch %v, processed %d events", epoch, removedExcessBalanceProcessed)
		}
	}

	log.Infof("exporting duties & balances for epoch %v", epoch)

	g := errgroup.Group{}

	// store epoch assignments in redis
	g.Go(func() error {
		return s.saveEpochAssignmentsToRedis(block, epoch, chainID, isHeadEpoch)
	})

	// save attestation duties to bigtable
	g.Go(func() error {
		return s.saveEpochAssigmentsToBigtable(block, epoch)
	})

	// save the validator balances to bigtable
	g.Go(func() error {
		err := s.bt.SaveValidatorBalances(epoch, block.Validators)
		if err != nil {
			return fmt.Errorf("error exporting validator balances to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})

	// if we are exporting the head epoch, update the validator db table
	if isHeadEpoch {
		if err := s.ExportValidatorData(block.Validators, epoch, chainID); err != nil {
			return err
		}
	}

	var epochParticipationStats *types.ValidatorParticipation
	if epoch > 0 {
		g.Go(func() error {
			// retrieve the epoch participation stats
			var err error
			epochParticipationStats, err = s.Client.GetValidatorParticipation(epoch - 1)
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
	err = s.db.SaveEpoch(epoch, block.Validators, s.dbTx)
	if err != nil {
		return fmt.Errorf("error saving epoch data: %w", err)
	}

	if epoch > 0 && epochParticipationStats != nil {
		log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
		err := s.db.UpdateEpochStatus(epochParticipationStats, s.dbTx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *exporter) saveEpochAssigmentsToBigtable(block *types.Block, epoch uint64) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		s.metrics.ObserveTaskDuration("slot_exporter_export_epoch_assignments_to_bigtable", time.Since(timeStart))
	}(timeStart)

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

	// save attestation duties to bigtable
	err := s.bt.SaveAttestationDuties(attDutiesEpoch)
	if err != nil {
		return fmt.Errorf("error exporting attestation assignments to bigtable for slot %v: %w", block.Slot, err)
	}

	// save sync committee duties to bigtable
	err = s.bt.SaveSyncCommitteeDuties(syncDutiesEpoch)
	if err != nil {
		return fmt.Errorf("error exporting sync committee assignments to bigtable for slot %v: %w", block.Slot, err)
	}

	return nil
}

func (s *exporter) saveEpochAssignmentsToRedis(block *types.Block, epoch, chainID uint64, isHeadEpoch bool) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		s.metrics.ObserveTaskDuration("slot_exporter_export_epoch_assignments_to_redis", time.Since(timeStart))
	}(timeStart)

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

	expirationTime := utils.EpochToTime(epoch + 7) // keep it for at least 7 epochs in the cache
	expirationDuration := time.Until(expirationTime)

	if expirationDuration.Seconds() < 0 || expirationDuration.Hours() > 2 {
		log.Warnf("NOT writing assignments data for epoch %v to redis because a TTL < 0 or TTL > 2h: %v", epoch, expirationDuration)
	} else {
		log.Infof("writing assignments data for epoch %v to redis with a TTL of %v", epoch, expirationDuration)
		err = s.cache.SetEpochAssignments(chainID, epoch, serializedAssignmentsData.Bytes(), expirationDuration)
		if err != nil {
			return fmt.Errorf("error writing assignments data to redis for epoch %v: %w", epoch, err)
		}
		// publish the event to inform the api about the new data (todo)
		// db.PersistentRedisDbClient.Publish(context.Background(), fmt.Sprintf("%d:slotViz", utils.Config.Chain.ClConfig.DepositChainID), fmt.Sprintf("%s:%d", "ea", epoch)).Err()
		log.Infof("writing current epoch assignments to redis completed")
	}

	if isHeadEpoch {
		nextEpoch := epoch + 1
		nextEpochAssignments, err := s.Client.GetEpochAssignments(nextEpoch)
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

		expirationTime := utils.EpochToTime(nextEpoch + 7) // keep it for at least 7 epochs in the cache
		expirationDuration := time.Until(expirationTime)

		if expirationDuration.Seconds() < 0 || expirationDuration.Hours() > 2 {
			log.Warnf("NOT writing assignments data for head+1 epoch (%v) to redis because a TTL < 0 or TTL > 2h: %v", nextEpoch, expirationDuration)
		} else {
			log.Infof("writing assignments data for head+1 epoch (%v) to redis with a TTL of %v", nextEpoch, expirationDuration)
			err = s.cache.SetEpochAssignments(chainID, nextEpoch, serializedAssignmentsData.Bytes(), expirationDuration)
			if err != nil {
				return fmt.Errorf("error writing assignments data for head+1 epoch to redis for epoch %v: %w", nextEpoch, err)
			}
		}
	}

	return nil
}

func (s *exporter) ExportValidatorData(validators []*types.Validator, epoch, chainID uint64) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		s.metrics.ObserveTaskDuration("slot_exporter_export_epoch_validators_data", time.Since(timeStart))
	}(timeStart)

	g := errgroup.Group{}

	// this function sets exports the validator status into the db
	// and also updates the status field in the validators array
	err := s.SaveValidators(validators)
	if err != nil {
		return fmt.Errorf("error saving validators for epoch %v: %w", epoch, err)
	}

	var genesisBalances map[uint64][]*types.ValidatorBalance
	if epoch == 0 {
		var err error
		indices := make([]uint64, 0, len(validators))

		for _, validator := range validators {
			indices = append(indices, validator.Index)
		}
		genesisBalances, err = s.bt.GetValidatorBalanceHistory(indices, 0, 0)
		if err != nil {
			return fmt.Errorf("error retrieving genesis validator balances: %w", err)
		}
	}

	vl, err := s.db.GetValidatorsWithMissingBalances(10000, s.dbTx)
	if err != nil {
		return fmt.Errorf("error retrieving validators with missing balances: %w", err)
	}

	balanceCache := make(map[uint64]map[uint64]uint64) // cache balances by epoch
	currentActivationEpoch := uint64(0)

	balanceStart := time.Now()
	for _, validator := range vl {
		if validator.ActivationEpoch > epoch {
			continue
		}

		if validator.ActivationEpoch != currentActivationEpoch {
			log.Infof("removing epoch %v from the activation epoch balance cache", currentActivationEpoch)
			delete(balanceCache, currentActivationEpoch) // remove old items from the map
			currentActivationEpoch = validator.ActivationEpoch
		}

		var balance map[uint64][]*types.ValidatorBalance
		if validator.ActivationEpoch == 0 {
			balance = genesisBalances
		} else {
			balance, err = s.bt.GetValidatorBalanceHistory([]uint64{validator.ValidatorIndex}, validator.ActivationEpoch, validator.ActivationEpoch)
			if err != nil {
				return fmt.Errorf("error retrieving validator balance history: %w", err)
			}
		}

		foundBalance := uint64(0)
		if balance[validator.ValidatorIndex] == nil || len(balance[validator.ValidatorIndex]) == 0 {
			log.Warnf("no activation epoch balance found for validator %v for epoch %v in bigtable, trying node", validator.ValidatorIndex, validator.ActivationEpoch)

			if balanceCache[validator.ActivationEpoch] == nil {
				balances, err := s.Client.GetBalancesForEpoch(int64(validator.ActivationEpoch))
				if err != nil {
					return fmt.Errorf("error retrieving balances for epoch %d: %v", validator.ActivationEpoch, err)
				}
				balanceCache[validator.ActivationEpoch] = balances
			}
			foundBalance = balanceCache[validator.ActivationEpoch][validator.ValidatorIndex]
		} else {
			foundBalance = balance[validator.ValidatorIndex][0].Balance
		}

		log.Infof("retrieved activation epoch balance of %v for validator %v", foundBalance, validator.ValidatorIndex)

		err = s.db.UpdateActivationEpochBalance(validator.ValidatorIndex, foundBalance, s.dbTx)
		if err != nil {
			return fmt.Errorf("error saving activation epoch balance for validator %v: %w", validator.ValidatorIndex, err)
		}
	}
	log.Infof("updating validator activation epoch balance completed, took %v", time.Since(balanceStart))

	err = s.db.AnalyzeValidatorsTable(s.dbTx)
	if err != nil {
		return fmt.Errorf("error analyzing validators table: %w", err)
	}

	// also update the queue deposit table once every epoch
	g.Go(func() error {
		err = s.db.UpdateQueueDeposits(s.dbTx)
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
			Mapping: make([]*types.CachedValidator, len(validators)),
		}

		activationMapping := make(map[int][]uint64)
		start := time.Now()

		for _, v := range validators {
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
		log.Infof("writing validator mappping to redis with no TTL")
		err = s.cache.SetValidatorMapping(chainID, compressedValidatorMapping.Bytes(), 0)
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

	if config.ClConfig.ElectraForkEpoch != nil && *config.ClConfig.ElectraForkEpoch < utils.MaxForkEpoch {
		// update cached view of consensus deposit requests
		g.Go(func() error {
			start := time.Now()
			err := s.db.CacheBlockDepositRequestsLookup()
			if err != nil {
				return fmt.Errorf("error updating cached view of consensus deposit requests: %w", err)
			}
			log.Infof("updating cached view of consensus deposit requests took %s", time.Since(start))
			return nil
		})
	}

	return g.Wait()
}

func (s *exporter) SaveValidators(validators []*types.Validator) error {
	currentState, err := s.db.GetValidatorsCurrentState(s.dbTx)
	if err != nil {
		return fmt.Errorf("error retrieving current validator state: %w", err)
	}

	for ; ; time.Sleep(time.Second) { // wait till the last attestation in memory cache has been populated by the exporter
		s.bt.GetLastAttestationCacheMux().Lock()
		if s.bt.GetLastAttestationCache() != nil {
			s.bt.GetLastAttestationCacheMux().Unlock()
			break
		}
		s.bt.GetLastAttestationCacheMux().Unlock()
		log.Infof("waiting until LastAttestation in memory cache is available")
	}

	currentStateMap := make(map[uint64]*types.Validator, len(currentState))
	latestBlock := uint64(0)

	// safely access and update the latestBlock and currentStateMap
	s.bt.GetLastAttestationCacheMux().Lock()
	for _, v := range currentState {
		if s.bt.GetLastAttestationCache()[v.Index] > latestBlock {
			latestBlock = s.bt.GetLastAttestationCache()[v.Index]
		}
		currentStateMap[v.Index] = v
	}
	s.bt.GetLastAttestationCacheMux().Unlock()

	lastGlobalAttestedEpoch := int64(latestBlock / utils.Config.Chain.ClConfig.SlotsPerEpoch)
	latestEpoch := latestBlock / utils.Config.Chain.ClConfig.SlotsPerEpoch

	log.Info("updating validator status and metadata")

	valiudatorUpdateTs := time.Now()
	validatorStatusCounts := make(map[string]int)
	validatorStatusUpdateMap := make(map[string][]uint64)
	updates := 0
	var queries strings.Builder

	for _, v := range validators {
		// exchange farFutureEpoch with the corresponding max sql value
		if v.WithdrawableEpoch == edb.FarFutureEpoch {
			v.WithdrawableEpoch = edb.MaxSqlNumber
		}
		if v.ExitEpoch == edb.FarFutureEpoch {
			v.ExitEpoch = edb.MaxSqlNumber
		}
		if v.ActivationEligibilityEpoch == edb.FarFutureEpoch {
			v.ActivationEligibilityEpoch = edb.MaxSqlNumber
		}
		if v.ActivationEpoch == edb.FarFutureEpoch {
			v.ActivationEpoch = edb.MaxSqlNumber
		}

		c := currentStateMap[v.Index]

		if c == nil {
			if v.Index%1000 == 0 {
				log.Infof("validator %v is new", v.Index)
			}

			err := s.db.SaveNewValidator(v, s.dbTx)
			if err != nil {
				log.Error(err, "error saving new validator", 0, map[string]interface{}{"index": v.Index})
			}
			validatorStatusCounts[v.Status]++
		} else {
			// safely read the last attestation slot for the validator
			s.bt.GetLastAttestationCacheMux().Lock()
			lastAttestationSlot := s.bt.GetLastAttestationCache()[v.Index]
			lastValidatorAttestedEpoch := int64(lastAttestationSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
			offline := lastGlobalAttestedEpoch-lastValidatorAttestedEpoch > 1 // validator has not attested in the last two epochs
			s.bt.GetLastAttestationCacheMux().Unlock()

			if v.ExitEpoch <= latestEpoch && v.Slashed {
				v.Status = string(constypes.DbSlashed)
			} else if v.ExitEpoch <= latestEpoch {
				v.Status = string(constypes.DbExited)
			} else if v.ActivationEligibilityEpoch == edb.MaxSqlNumber {
				v.Status = string(constypes.DbDeposited)
			} else if v.ActivationEpoch > latestEpoch {
				v.Status = string(constypes.DbPending)
			} else if v.Slashed && v.ActivationEpoch < latestEpoch && offline {
				v.Status = string(constypes.DbSlashingOffline)
			} else if v.Slashed {
				v.Status = string(constypes.DbSlashingOnline)
			} else if v.ExitEpoch < edb.MaxSqlNumber && offline {
				v.Status = string(constypes.DbExitingOffline)
			} else if v.ExitEpoch < edb.MaxSqlNumber {
				v.Status = string(constypes.DbExitingOnline)
			} else if v.ActivationEpoch < latestEpoch && offline {
				v.Status = string(constypes.DbActiveOffline)
			} else {
				v.Status = string(constypes.DbActiveOnline)
			}

			validatorStatusCounts[v.Status]++

			if c.Status != v.Status {
				log.Debugf("Status changed for validator %v from %v to %v", v.Index, c.Status, v.Status)
				log.Debugf("v.ActivationEpoch %v, latestEpoch %v, lastAttestationSlots[v.Index] %v, lastGlobalAttestedEpoch: %v, lastValidatorAttestedEpoch: %v", v.ActivationEpoch, latestEpoch, lastAttestationSlot, lastGlobalAttestedEpoch, lastValidatorAttestedEpoch)
				if validatorStatusUpdateMap[v.Status] == nil {
					validatorStatusUpdateMap[v.Status] = make([]uint64, 0)
				}
				validatorStatusUpdateMap[v.Status] = append(validatorStatusUpdateMap[v.Status], c.Index)
			}

			updateCount, updateQueries, err := s.db.PrepareValidatorsUpdate(c, v, s.dbTx)
			if err != nil {
				return fmt.Errorf("error preparing validators update: %w", err)
			}

			updates += updateCount
			queries.WriteString(updateQueries)
		}
	}

	log.Infof("processing validator updates for %d status entry", len(validatorStatusUpdateMap))
	err = s.db.UpdateValidatorsStatus(validatorStatusUpdateMap, s.dbTx)
	if err != nil {
		return fmt.Errorf("error saving validators status: %w", err)
	}

	if updates > 0 {
		err := s.db.UpdateValidators(queries.String(), updates, s.dbTx)
		if err != nil {
			return fmt.Errorf("error saving validators update: %w", err)
		}
	}

	log.Infof("updating validator status and metadata completed, took %v", time.Since(valiudatorUpdateTs))

	return nil
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
