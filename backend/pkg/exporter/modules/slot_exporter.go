package modules

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/klauspost/pgzip"

	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
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

type slotExporterData struct {
	ModuleContext
	Client   SlotExporterClient
	cache    edb.SlotExporterCacheRepository
	db       edb.SlotExporterDBRepository
	bt       edb.SlotExporterBTRepository
	FirstRun bool
}

func NewSlotExporter(moduleContext ModuleContext, cache edb.SlotExporterCacheRepository, db edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository) ModuleInterface {
	return &slotExporterData{
		ModuleContext: moduleContext,
		Client:        moduleContext.ConsClient,
		cache:         cache,
		db:            db,
		bt:            bt,
		FirstRun:      true,
	}
}

var latestEpoch, latestSlot, finalizedEpoch, latestProposed uint64 // holy shit these should really be in the slotExporterData struct or some other module might accidentally corrupt them

var processSlotMutex = &sync.Mutex{}

func (d *slotExporterData) OnHead(_ *constypes.StandardEventHeadResponse) (err error) {
	if !processSlotMutex.TryLock() {
		log.Infof("slotExporter is still running, skipping this run")
		return nil
	}
	defer processSlotMutex.Unlock()

	latestEpoch, latestSlot, finalizedEpoch, latestProposed = 0, 0, 0, 0
	// cache handling
	defer func() {
		if err == nil {
			chainID := utils.Config.Chain.ClConfig.DepositChainID

			cacheLatestEpoch, err := d.cache.GetLatestEpoch(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestEpoch from cache", 0)
			}

			if latestEpoch > 0 && cacheLatestEpoch < latestEpoch {
				err := d.cache.SetLatestEpoch(chainID, latestEpoch)
				if err != nil {
					log.Error(err, "error setting latestEpoch in cache", 0)
				}
			}

			cacheLatestSlot, err := d.cache.GetLatestSlot(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestSlot from cache", 0)
			}
			if latestSlot > 0 && cacheLatestSlot < latestSlot {
				err := d.cache.SetLatestSlot(chainID, latestSlot)
				if err != nil {
					log.Error(err, "error setting latestSlot in cache", 0)
				}
			}

			cacheLatestFinalizedEpoch, err := d.cache.GetLatestFinalizedEpoch(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestFinalizedEpoch from cache", 0)
			}
			if finalizedEpoch > 0 && cacheLatestFinalizedEpoch < finalizedEpoch {
				err := d.cache.SetLatestFinalizedEpoch(chainID, finalizedEpoch)
				if err != nil {
					log.Error(err, "error setting latestFinalizedEpoch in cache", 0)
				}
			}

			cacheLatestProposedSlot, err := d.cache.GetLatestProposedSlot(chainID)
			if err != nil {
				log.Error(err, "error retrieving latestProposedSlot from cache", 0)
			}
			if latestProposed > 0 && cacheLatestProposedSlot < latestProposed {
				err := d.cache.SetLatestProposedSlot(chainID, latestProposed)
				if err != nil {
					log.Error(err, "error setting latestProposedSlot in cache", 0)
				}
			}
		}
	}()

	// get the current chain head
	head, err := d.Client.GetChainHead()

	if err != nil {
		return fmt.Errorf("error retrieving chain head: %w", err)
	}

	tx, err := d.db.BeginTx()
	if err != nil {
		return fmt.Errorf("error starting tx: %w", err)
	}
	defer d.db.RollbackTx(tx)

	if d.FirstRun {
		log.Infof("performing first run consistency checks")
		// get all slots we currently have in the database
		dbSlots, err := d.db.GetAllSlots(tx)
		if err != nil {
			return fmt.Errorf("error retrieving all db slots: %w", err)
		}
		log.Info("retrieved all exported slots from the database")

		if len(dbSlots) > 0 {
			if dbSlots[0] != 0 {
				log.Infof("exporting genesis slot as it is missing in the database")
				err := ExportSlot(d.Client, 0, utils.EpochOfSlot(0) == head.HeadEpoch, d.cache, d.db, d.bt, tx)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", 0, err)
				}
				dbSlots, err = d.db.GetAllSlots(tx)
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
						err := ExportSlot(d.Client, slot, false, d.cache, d.db, d.bt, tx)

						if err != nil {
							return fmt.Errorf("error exporting slot %v: %w", slot, err)
						}
					}
				}
			}
		}
	}
	d.FirstRun = false

	// at this point we know that we have a coherent list of slots in the database without any gaps
	lastDbSlot, err := d.db.GetLastSlot(tx)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Infof("db is empty, export genesis slot")
			err := ExportSlot(d.Client, 0, utils.EpochOfSlot(0) == head.HeadEpoch, d.cache, d.db, d.bt, tx)
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
			err := ExportSlot(d.Client, slot, utils.EpochOfSlot(slot) == head.HeadEpoch, d.cache, d.db, d.bt, tx)
			if err != nil {
				return fmt.Errorf("error exporting slot %v: %w", slot, err)
			}
			slotsExported++

			// in case of large export runs, export at most 10 epochs per tx
			if slotsExported == int(utils.Config.Chain.ClConfig.SlotsPerEpoch)*10 {
				err := d.db.CommitTx(tx)
				if err != nil {
					return fmt.Errorf("error committing tx: %w", err)
				}

				latestEpoch = utils.EpochOfSlot(slot)
				latestSlot = slot

				return nil
			}
		}
	}

	// at this point we have all data up to the current chain head in the database

	// check if any non-finalized slot has changed by comparing it with the node
	dbNonFinalSlots, err := d.db.GetAllNonFinalizedSlots()
	if err != nil {
		return fmt.Errorf("error retrieving all non finalized slots from the db: %w", err)
	}
	for _, dbSlot := range dbNonFinalSlots {
		nodeSlotFinalized := dbSlot.Slot <= head.FinalizedSlot

		var header *constypes.StandardBeaconHeaderResponse

		if nodeSlotFinalized != dbSlot.Finalized {
			log.Infof("checking slot %d for finalization / reorgs", dbSlot.Slot)
			header, err = d.Client.GetBlockHeader(dbSlot.Slot)

			if err != nil {
				return fmt.Errorf("error retrieving block root for slot %v: %w", dbSlot.Slot, err)
			}

			// slot has finalized, mark it in the db
			if header != nil && bytes.Equal(dbSlot.BlockRoot, header.Data.Root) {
				// no reorg happened, simply mark the slot as final
				log.Infof("setting slot %v as finalized (proposed)", dbSlot.Slot)
				err := d.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, dbSlot.Status, tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (proposed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) < 32 {
				// no reorg happened, mark the slot as missed
				log.Infof("setting slot %v as finalized (missed)", dbSlot.Slot)
				err := d.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "2", tx)
				if err != nil {
					return fmt.Errorf("error setting slot %v as finalized (missed): %w", dbSlot.Slot, err)
				}
			} else if header == nil && len(dbSlot.BlockRoot) == 32 {
				// slot has been orphaned, mark the slot as orphaned
				log.Infof("setting slot %v as finalized (orphaned)", dbSlot.Slot)
				err := d.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
			} else if header != nil && !bytes.Equal(header.Data.Root, dbSlot.BlockRoot) {
				// we have a different block root for the slot in the db, mark the currently present one as orphaned and write the new one
				log.Infof("setting slot %v as orphaned and exporting new slot", dbSlot.Slot)
				err := d.db.SetSlotFinalizationAndStatus(dbSlot.Slot, nodeSlotFinalized, "3", tx)
				if err != nil {
					return fmt.Errorf("error setting block %v as finalized (orphaned): %w", dbSlot.Slot, err)
				}
				err = ExportSlot(d.Client, dbSlot.Slot, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, d.cache, d.db, d.bt, tx)
				if err != nil {
					return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
				}
			}

			// epoch transition slot has finalized, update epoch status
			if dbSlot.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 && dbSlot.Slot > utils.Config.Chain.ClConfig.SlotsPerEpoch-1 {
				epoch := utils.EpochOfSlot(dbSlot.Slot)
				epochParticipationStats, err := d.Client.GetValidatorParticipation(epoch - 1)
				if err != nil {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				} else {
					log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
					err := d.db.UpdateEpochStatus(epochParticipationStats, tx)
					if epochParticipationStats.Finalized && epochParticipationStats.Epoch > finalizedEpoch {
						finalizedEpoch = epochParticipationStats.Epoch
					}

					if err != nil {
						return err
					}

					log.Infof("exporting validation queue")
					queue, err := d.Client.GetValidatorQueue()
					if err != nil {
						return fmt.Errorf("error retrieving validator queue data: %w", err)
					}

					err = d.db.SaveValidatorQueue(queue, tx)
					if err != nil {
						return fmt.Errorf("error saving validator queue data: %w", err)
					}
				}
			}
		} else {
			// check if a late slot has been proposed in the meantime
			// TODO: reenable once holesky is close to recovery
			if utils.Config.Chain.Id != 17000 {
				if len(dbSlot.BlockRoot) < 32 && header != nil { // we have no slot in the db, but the node has a slot, export it
					log.Infof("exporting new slot %v", dbSlot.Slot)
					err := ExportSlot(d.Client, dbSlot.Slot, utils.EpochOfSlot(dbSlot.Slot) == head.HeadEpoch, d.cache, d.db, d.bt, tx)
					if err != nil {
						return fmt.Errorf("error exporting slot %v: %w", dbSlot.Slot, err)
					}
				}
			}
		}
	}

	err = d.db.CommitTx(tx)
	if err != nil {
		return fmt.Errorf("error committing tx: %w", err)
	}

	latestEpoch = utils.EpochOfSlot(head.HeadSlot)
	latestSlot = head.HeadSlot

	services.ReportStatus("slotExporter", "Running", nil)

	return nil
}

func ExportSlot(client SlotExporterClient, slot uint64, isHeadEpoch bool, cache edb.SlotExporterCacheRepository, exporterdb edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository, tx *sqlx.Tx) error {
	isFirstSlotOfEpoch := slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0
	epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	chainID := utils.Config.Chain.ClConfig.DepositChainID

	if isFirstSlotOfEpoch {
		log.Infof("exporting slot %v (epoch transition into epoch %v)", slot, epoch)
	} else {
		log.Infof("exporting slot %v", slot)
	}
	start := time.Now()

	// retrieve the data for the slot from the node
	// the first slot of an epoch will also contain all validator duties for the whole epoch
	block, err := client.GetBlockBySlot(slot)
	if err != nil {
		return fmt.Errorf("error retrieving data for slot %v: %w", slot, err)
	}

	// for the slot itself start by preparing the duties for export to bigtable
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
		err = bt.SaveAttestationDuties(attDuties)
		if err != nil {
			return fmt.Errorf("error exporting attestations to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})
	g.Go(func() error {
		err = bt.SaveSyncCommitteeDuties(syncDuties)
		if err != nil {
			return fmt.Errorf("error exporting sync committee duties to bigtable for slot %v: %w", block.Slot, err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	// save the block data to the db
	err = exporterdb.SaveBlock(block, false, tx)
	if err != nil {
		return fmt.Errorf("error saving slot to the db: %w", err)
	}

	if block.Status == 1 {
		if latestProposed < block.Slot {
			latestProposed = block.Slot
		}
	}

	if block.EpochAssignments != nil { // export the epoch assignments as they are included in the first slot of an epoch
		epoch := utils.EpochOfSlot(block.Slot)
		if epoch > utils.Config.ClConfig.ElectraForkEpoch {
			log.Infof("checking that events have been loaded for epoch %v", epoch)
			exported, err := exporterdb.HasEventsForEpoch(epoch)
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

				switchToCompoundingRequestsProcessed, err := exporterdb.TransformSwitchToCompoundingRequests(firstSlot, lastSlot, tx)
				if err != nil {
					return fmt.Errorf("error transforming consolidation requests for epoch %v: %w", epoch, err)
				}
				log.Infof("transformed switch to compounding requests for epoch %v, processed %d requests", epoch, switchToCompoundingRequestsProcessed)

				consolidationRequestsProcessed, err := exporterdb.TransformConsolidationRequests(firstSlot, lastSlot, tx)
				if err != nil {
					return fmt.Errorf("error transforming consolidation requests for epoch %v: %w", epoch, err)
				}
				log.Infof("transformed consolidations for epoch %v, processed %d requests", epoch, consolidationRequestsProcessed)

				depositRequestsProcessed, err := exporterdb.TransformDepositRequests(firstSlot, lastSlot, tx)
				if err != nil {
					return fmt.Errorf("error transforming deposit requests for epoch %v: %w", epoch, err)
				}
				log.Infof("transformed deposits for epoch %v, processed %d requests", epoch, depositRequestsProcessed)

				removedExcessBalanceProcessed, err := exporterdb.TransformRemovedExcessBalanceEvents(firstSlot, lastSlot, tx)
				if err != nil {
					return fmt.Errorf("error transforming removed excess balance events for epoch %v: %w", epoch, err)
				}
				log.Infof("transformed removed excess balance events for epoch %v, processed %d events", epoch, removedExcessBalanceProcessed)
			}
		}

		log.Infof("exporting duties & balances for epoch %v", epoch)
		// prepare the duties for export to bigtable
		syncDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex]bool)
		attDutiesEpoch := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if syncDutiesEpoch[types.Slot(slot)] == nil {
				syncDutiesEpoch[types.Slot(slot)] = make(map[types.ValidatorIndex]bool)
			}
			for _, validatorIndex := range block.EpochAssignments.SyncAssignments {
				syncDutiesEpoch[types.Slot(slot)][types.ValidatorIndex(validatorIndex)] = false
			}
		}

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
				err = cache.SetEpochAssignments(chainID, epoch, serializedAssignmentsData.Bytes(), expirationDuration)
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

				expirationTime := utils.EpochToTime(nextEpoch + 7) // keep it for at least 7 epochs in the cache
				expirationDuration := time.Until(expirationTime)
				if expirationDuration.Seconds() < 0 || expirationDuration.Hours() > 2 {
					log.Warnf("NOT writing assignments data for head+1 epoch (%v) to redis because a TTL < 0 or TTL > 2h: %v", nextEpoch, expirationDuration)
				} else {
					log.Infof("writing assignments data for head+1 epoch (%v) to redis with a TTL of %v", nextEpoch, expirationDuration)
					err = cache.SetEpochAssignments(chainID, nextEpoch, serializedAssignmentsData.Bytes(), expirationDuration)
					if err != nil {
						return fmt.Errorf("error writing assignments data for head+1 epoch to redis for epoch %v: %w", nextEpoch, err)
					}
				}
			}

			return nil
		})

		// save all duties to bigtable
		g.Go(func() error {
			err := bt.SaveAttestationDuties(attDutiesEpoch)
			if err != nil {
				return fmt.Errorf("error exporting attestation assignments to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})
		g.Go(func() error {
			err := bt.SaveSyncCommitteeDuties(syncDutiesEpoch)
			if err != nil {
				return fmt.Errorf("error exporting sync committee assignments to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})

		// save the validator balances to bigtable
		g.Go(func() error {
			err := bt.SaveValidatorBalances(epoch, block.Validators)
			if err != nil {
				return fmt.Errorf("error exporting validator balances to bigtable for slot %v: %w", block.Slot, err)
			}
			return nil
		})

		// if we are exporting the head epoch, update the validator db table
		if isHeadEpoch {
			err := ExportValidatorData(block.Validators, epoch, chainID, cache, client, exporterdb, bt, tx)
			if err != nil {
				return fmt.Errorf("error saving validators for epoch %v: %w", epoch, err)
			}

			// also update the queue deposit table once every epoch
			g.Go(func() error {
				err = exporterdb.UpdateQueueDeposits(tx)
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
				log.Infof("writing validator mappping to redis with no TTL")
				err = cache.SetValidatorMapping(chainID, compressedValidatorMapping.Bytes(), 0)
				if err != nil {
					return fmt.Errorf("error writing validator mapping to redis for epoch %v: %w", epoch, err)
				}
				log.Infof("writing validator mapping to redis done, took %s", time.Since(start))
				return nil
			})

			// update cached view of consensus deposits
			// possible bug: at this point the export tx is not yet committed, so the query will read
			// stale data
			g.Go(func() error {
				start := time.Now()
				err := exporterdb.CacheBlockDepositLookup()
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
					err := exporterdb.CacheBlockDepositRequestsLookup()
					if err != nil {
						return fmt.Errorf("error updating cached view of consensus deposit requests: %w", err)
					}
					log.Infof("updating cached view of consensus deposit requests took %s", time.Since(start))
					return nil
				})
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
		err = g.Wait()
		if err != nil {
			return err
		}

		// save the epoch metadata to the database
		err = exporterdb.SaveEpoch(epoch, block.Validators, tx)
		if err != nil {
			return fmt.Errorf("error saving epoch data: %w", err)
		}

		if epoch > 0 && epochParticipationStats != nil {
			log.Infof("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
			err := exporterdb.UpdateEpochStatus(epochParticipationStats, tx)

			if err != nil {
				return err
			}
		}

		// time.Sleep(time.Minute)
	}

	// time.Sleep(time.Second)

	log.InfoWithFields(
		log.Fields{
			"slot":      block.Slot,
			"blockRoot": fmt.Sprintf("%x", block.BlockRoot),
			"duration":  time.Since(start),
		}, "! export of slot completed")

	return nil
}

func ExportValidatorData(validators []*types.Validator, epoch, chainID uint64, cache edb.SlotExporterCacheRepository, client SlotExporterClient, exporterdb edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository, tx *sqlx.Tx) error {
	g := errgroup.Group{}

	// this function sets exports the validator status into the db
	// and also updates the status field in the validators array
	err := SaveValidators(validators, exporterdb, bt, tx)
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
		genesisBalances, err = bt.GetValidatorBalanceHistory(indices, 0, 0)
		if err != nil {
			return fmt.Errorf("error retrieving genesis validator balances: %w", err)
		}
	}

	vl, err := exporterdb.GetValidatorsWithMissingBalances(10000, tx)
	if err != nil {
		return fmt.Errorf("error retrieving validators with missing balances: %w", err)
	}

	balanceCache := make(map[uint64]map[uint64]uint64) // cache balances by epoch
	currentActivationEpoch := uint64(0)

	timeStart := time.Now()
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
			balance, err = bt.GetValidatorBalanceHistory([]uint64{validator.ValidatorIndex}, validator.ActivationEpoch, validator.ActivationEpoch)
			if err != nil {
				return fmt.Errorf("error retrieving validator balance history: %w", err)
			}
		}

		foundBalance := uint64(0)
		if balance[validator.ValidatorIndex] == nil || len(balance[validator.ValidatorIndex]) == 0 {
			log.Warnf("no activation epoch balance found for validator %v for epoch %v in bigtable, trying node", validator.ValidatorIndex, validator.ActivationEpoch)

			if balanceCache[validator.ActivationEpoch] == nil {
				balances, err := client.GetBalancesForEpoch(int64(validator.ActivationEpoch))
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

		err = exporterdb.UpdateActivationEpochBalance(validator.ValidatorIndex, foundBalance, tx)
		if err != nil {
			return fmt.Errorf("error saving activation epoch balance for validator %v: %w", validator.ValidatorIndex, err)
		}
	}
	log.Infof("updating validator activation epoch balance completed, took %v", time.Since(timeStart))

	err = exporterdb.AnalyzeValidatorsTable(tx)
	if err != nil {
		return fmt.Errorf("error analyzing validators table: %w", err)
	}

	// also update the queue deposit table once every epoch
	g.Go(func() error {
		err = exporterdb.UpdateQueueDeposits(tx)
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
		err = cache.SetValidatorMapping(chainID, compressedValidatorMapping.Bytes(), 0)
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
		err := exporterdb.CacheBlockDepositLookup()
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
			err := exporterdb.CacheBlockDepositRequestsLookup()
			if err != nil {
				return fmt.Errorf("error updating cached view of consensus deposit requests: %w", err)
			}
			log.Infof("updating cached view of consensus deposit requests took %s", time.Since(start))
			return nil
		})
	}

	return g.Wait()
}

func SaveValidators(validators []*types.Validator, exporterdb edb.SlotExporterDBRepository, bt edb.SlotExporterBTRepository, tx *sqlx.Tx) error {
	currentState, err := exporterdb.GetValidatorsCurrentState(tx)
	if err != nil {
		return fmt.Errorf("error retrieving current validator state: %w", err)
	}

	for ; ; time.Sleep(time.Second) { // wait till the last attestation in memory cache has been populated by the exporter
		bt.GetLastAttestationCacheMux().Lock()
		if bt.GetLastAttestationCache() != nil {
			bt.GetLastAttestationCacheMux().Unlock()
			break
		}
		bt.GetLastAttestationCacheMux().Unlock()
		log.Infof("waiting until LastAttestation in memory cache is available")
	}

	currentStateMap := make(map[uint64]*types.Validator, len(currentState))
	latestBlock := uint64(0)

	// safely access and update the latestBlock and currentStateMap
	bt.GetLastAttestationCacheMux().Lock()
	for _, v := range currentState {
		if bt.GetLastAttestationCache()[v.Index] > latestBlock {
			latestBlock = bt.GetLastAttestationCache()[v.Index]
		}
		currentStateMap[v.Index] = v
	}
	bt.GetLastAttestationCacheMux().Unlock()

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

			err := exporterdb.SaveNewValidator(v, tx)
			if err != nil {
				log.Error(err, "error saving new validator", 0, map[string]interface{}{"index": v.Index})
			}
			validatorStatusCounts[v.Status]++
		} else {
			// safely read the last attestation slot for the validator
			bt.GetLastAttestationCacheMux().Lock()
			lastAttestationSlot := bt.GetLastAttestationCache()[v.Index]
			lastValidatorAttestedEpoch := int64(lastAttestationSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
			offline := lastGlobalAttestedEpoch-lastValidatorAttestedEpoch > 1 // validator has not attested in the last two epochs
			bt.GetLastAttestationCacheMux().Unlock()

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

			updateCount, updateQueries, err := exporterdb.PrepareValidatorsUpdate(c, v, tx)
			if err != nil {
				return fmt.Errorf("error preparing validators update: %w", err)
			}

			updates += updateCount
			queries.WriteString(updateQueries)
		}
	}

	log.Infof("processing validator updates for %d status entry", len(validatorStatusUpdateMap))
	err = exporterdb.UpdateValidatorsStatus(validatorStatusUpdateMap, tx)
	if err != nil {
		return fmt.Errorf("error saving validators status: %w", err)
	}

	if updates > 0 {
		err := exporterdb.UpdateValidators(queries.String(), updates, tx)
		if err != nil {
			return fmt.Errorf("error saving validators update: %w", err)
		}
	}

	log.Infof("updating validator status and metadata completed, took %v", time.Since(valiudatorUpdateTs))

	return nil
}

func (d *slotExporterData) Init() error {
	// // Initialize the slot exporter by doing 20 sync runs
	// for i := 0; i < 20; i++ {
	// 	err := d.OnHead(nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (d *slotExporterData) GetName() string {
	return "Slot-Exporter"
}

func (d *slotExporterData) GetMonitoringEventId() constants.Event {
	return constants.Event_ExporterModuleSlotExporter
}

func (d *slotExporterData) OnChainReorg(event *constypes.StandardEventChainReorg) (err error) {
	return nil // nop
}

func (d *slotExporterData) OnFinalizedCheckpoint(event *constypes.StandardFinalizedCheckpointResponse) (err error) {
	return nil // nop
}
