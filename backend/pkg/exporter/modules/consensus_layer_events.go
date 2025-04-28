package modules

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules/cl_transformers"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

var epochsPerBatch uint64 = 128
var slotsPerBatch int = 128 // 4 epochs worth

// the fork plays down as follows:
// 		process_slots()
// 				process_slot() 	// slot: 29
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 30 65628
//
// 		process_slots()
// 				process_slot()  // slot: 30
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 31
//
// 		process_slots()
// 				process_slot()	// slot: 31
// 				process_epoch() // 				slot + 1 % 32 == 0
// 				incr_slot() 	// slot++
// 				upgrade_state(epochs)	// 		slot % 32 == 0			ElectraEvents(slot=31)	EpochProcessed(slot=31)		epoch 0 (we cheat and set the slots to -1 the actual slot. needed anyways because the blockroot in the events will still be of slot -1, since process_block didnt run yet)
// 		process_block() 		// slot: 32 							ElectraEvents(slot=32)								epoch |
//
// 		process_slots()
// 				process_slot() 	// slot: 32								ElectraEvents(slot=32)								epoch |
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 33								ElectraEvents(slot=33)								epoch |
//
// 		process_slots()
// 				process_slot() 	// slot: 33								ElectraEvents(slot=33)								epoch |
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 34								ElectraEvents(slot=34)								epoch |
// 		…
// 		process_slots()
// 				process_slot()	// slot: 63								ElectraEvents(slot=63)								epoch |
// 				process_epoch() // 				slot + 1 % 32 == 0		ElectraEvents(slot=63)	EpochProcessed(slot=63)		epoch |
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 64								ElectraEvents(slot=64)								epoch ||
//
// 		process_slots()
// 				process_slot() 	// slot: 64								ElectraEvents(slot=64)								epoch ||
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 65								ElectraEvents(slot=65)								epoch ||
// 		…
// 		process_slots()
// 				process_slot()	// slot: 63								ElectraEvents(slot=95)								epoch ||
// 				process_epoch() // 				slot + 1 % 32 == 0		ElectraEvents(slot=95)	EpochProcessed(slot=95)		epoch ||
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 64								ElectraEvents(slot=96)								epoch |||
//
// 		process_slots()
// 				process_slot() 	// slot: 64								ElectraEvents(slot=96)								epoch |||
// 				incr_slot() 	// slot++
// 		process_block() 		// slot: 65								ElectraEvents(slot=97)								epoch |||

type consensusLayerEventsIndexer struct {
	client consapi.ClientInt
	db     db2.ConsensusRepository
	delay  time.Duration
	ctx    context.Context
	config *consensusLayerEventsConfig
}

type consensusLayerEventsConfig struct {
	SlotsPerEpoch    uint64
	ElectraForkEpoch uint64
}

func newConsensusLayerEventsIndexer(ctx context.Context, client consapi.ClientInt, db db2.ConsensusRepository, config *consensusLayerEventsConfig) consensusLayerEventsIndexer {
	return consensusLayerEventsIndexer{
		client: client,
		db:     db,
		delay:  time.Second * 12,
		ctx:    ctx,
		config: config,
	}
}

func (c *consensusLayerEventsIndexer) Index() {
	// create ticker with interval set to e.delay
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lastRun := time.Now().Add(-c.delay)
	for {
		select {
		case <-c.ctx.Done():
			log.Info("consensus layer events index loop cancelled")
			return
		case <-ticker.C:
			if time.Since(lastRun) < c.delay {
				log.Tracef("skipping consensus layer events index loop, last run was %v ago", time.Since(lastRun))
				continue
			}
			partialRum, err := c.IndexEvents()
			if err != nil {
				log.Error(err, "error indexing consensus layer events", 0)
				lastRun = time.Now()
				continue
			}
			log.Debugf("consensus layer events index loop ran successfully")
			if !partialRum {
				lastRun = time.Now()
			}
		}
	}
}

func (c *consensusLayerEventsIndexer) IndexEvents() (bool, error) {
	didPartialRun := false
	// get latest finalized epoch from node
	res, err := c.client.GetFinalityCheckpoints("head")
	if err != nil {
		return false, errors.Wrap(err, "error retrieving latest finalized epoch from node")
	}
	finalized := res.Data.Finalized.Epoch
	if finalized == 0 || finalized >= math.MaxInt64 {
		log.Debugf("no epoch finalized yet, skipping index")
		return false, nil
	}
	finalized--
	if finalized < c.config.ElectraForkEpoch {
		log.Debugf("finalized epoch %d is less than electra fork epoch %d", finalized, c.config.ElectraForkEpoch)
		return false, nil
	}
	// get last indexed block from db
	lastIndexedEpoch, err := c.db.GetLastIndexedConsensusLayerEventsEpoch()
	if err != nil {
		return false, errors.Wrap(err, "error retrieving last indexed epoch from db")
	}
	if lastIndexedEpoch >= int64(finalized) {
		log.Debugf("last indexed epoch %d is greater than or equal to finalized epoch %d", lastIndexedEpoch, finalized)
		return false, nil
	}
	// start & end epoch are both inclusive
	minForkEpoch := uint64(max(0, int64(c.config.ElectraForkEpoch)-1))
	startEpoch := uint64(max(minForkEpoch, uint64(lastIndexedEpoch+1)))
	endEpoch := uint64(max(c.config.ElectraForkEpoch, finalized))

	if endEpoch-startEpoch > epochsPerBatch {
		log.Warnf("end epoch %d is greater than start epoch %d + epochs per batch %d", endEpoch, startEpoch, epochsPerBatch)
		endEpoch = startEpoch + epochsPerBatch
		didPartialRun = true
	}
	// check that the epoch before startEpoch has been exported. we wont run transformer against this epoch, but we need to know that it exists
	if startEpoch > minForkEpoch {
		tailEpoch := startEpoch - 1
		log.Infof("checking that tail epoch %d has been exported", tailEpoch)
		tailEvents, err := c.GetEpochProcessedEvents(tailEpoch, tailEpoch)
		if err != nil {
			return false, errors.Wrapf(err, "error gathering epoch processed events for tail epoch %d", tailEpoch)
		}
		tailBlockRoots, err := c.GetEpochBlockRoots(tailEvents)
		if err != nil {
			return false, errors.Wrapf(err, "error gathering epoch block roots for tail epoch %d", tailEpoch)
		}
		log.Debugf("verified that tail epoch %d has event with correct block root %x", tailEpoch, tailBlockRoots[tailEpoch].EpochBlockRoot)
	}
	epochProcessedEvents, err := c.GetEpochProcessedEvents(startEpoch, endEpoch)
	if err != nil {
		return false, errors.Wrap(err, "error gathering epoch processed events")
	}
	// at this point we know there is at least one epoch processed event per epoch in the db
	epochBlockRoots, err := c.GetEpochBlockRoots(epochProcessedEvents)
	if err != nil {
		return false, errors.Wrap(err, "error gathering epoch block roots")
	}
	// at this point we know that there is an epoch process event for each epoch in the db that matches the nodes data
	// we use a quick function to turn the map of epochBlockRoots into a slice of structs that each contain the slot and the block root
	// this is so we can abstract the transformation of the events to a more general approach
	// the events will still have to be processed in order by the transformers, but the idea that the transformers themselves can be stateless
	eventFilters := make([]types.ConsensusLayerEventFilter, 0)
	seenSlots := make(map[uint64]bool)
	for _, ebr := range epochBlockRoots {
		for slot, blockRoot := range ebr.SlotBlockRoots {
			if seenSlots[slot] {
				// should be impossible, but just in case
				return false, fmt.Errorf("found duplicate slot %d for epoch %d", slot, ebr.Epoch)
			}
			seenSlots[slot] = true
			eventFilters = append(eventFilters, types.ConsensusLayerEventFilter{
				Slot:      slot,
				BlockRoot: blockRoot,
			})
		}
	}
	// sort because maps are not ordered
	slices.SortFunc(eventFilters, func(a, b types.ConsensusLayerEventFilter) int {
		return int(a.Slot) - int(b.Slot)
	})
	// now we have a sorted list of event filters, we can process the events
	transformers := []cl_transformers.EventTransformer{
		&cl_transformers.DepositQueuedEventTransformer{},
		&cl_transformers.DepositRejectedPreQueueEventTransformer{},
		&cl_transformers.DepositPostponedEventTransformer{},
		&cl_transformers.DepositProcessedEventTransformer{},
		&cl_transformers.DepositRejectedPostQueueEventTransformer{},
		&cl_transformers.ConsolidationQueuedEventTransformer{},
		&cl_transformers.ConsolidationRejectedPreQueueEventTransformer{},
		&cl_transformers.ConsolidationProcessedEventTransformer{},
		&cl_transformers.ConsolidationRejectedPostQueueEventTransformer{},
		&cl_transformers.WithdrawalQueuedEventTransformer{},
		&cl_transformers.WithdrawalRejectedPreQueueEventTransformer{},
		&cl_transformers.WithdrawalProcessedEventTransformer{},
		&cl_transformers.WithdrawalRejectedPostQueueEventTransformer{},
		//&cl_transformers.RemovedExcessBalanceEventTransformer{},
		//&cl_transformers.ExitRequestProcessedEventTransformer{},
		&cl_transformers.SwitchToCompoundingEventTransformer{},
	}
	// the processing boils down to the following basically:
	// chunk the filters into chunks. for each chunk, fetch all possible events from the db using the filters,
	// then pass all events to all transformers. the key thing to watch is:
	// - do not run transformers in parallel, they need to be run in order, as they might depend on each other (*Processed depends on *Queued, etc)
	tx, err := c.db.GetWriterTx()
	if err != nil {
		return false, errors.Wrap(err, "error getting db transaction")
	}
	defer utils.Rollback(tx)
	for filterChunk := range slices.Chunk(eventFilters, slotsPerBatch) {
		// get all events for the chunk
		events, err := c.db.GetConsensusLayerEventsForFilters(filterChunk)
		if err != nil {
			return false, errors.Wrap(err, "error retrieving events from db")
		}
		if len(events) == 0 {
			log.Debugf("no events found for chunk %d-%d", filterChunk[0].Slot, filterChunk[len(filterChunk)-1].Slot)
			continue
		}
		log.Debugf("found %d events for chunk %d-%d", len(events), filterChunk[0].Slot, filterChunk[len(filterChunk)-1].Slot)
		// we do not sort our events because the transformers will do that for us
		for _, transformer := range transformers {
			log.Tracef("transforming events using transformer %T", transformer)
			c, err := transformer.Transform(tx, events)
			if err != nil {
				return false, errors.Wrapf(err, "error transforming events using transformer %T", transformer)
			}
			if c != 0 {
				log.Debugf("transformed %d events using transformer %T", c, transformer)
			}
		}
	}
	// todo: update the last indexed epoch in the db
	if err := c.db.UpdateConsensusLayerExporterMetadata(tx, maps.Values(epochBlockRoots)); err != nil {
		return false, errors.Wrap(err, "error updating last indexed epoch in db")
	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		return false, errors.Wrap(err, "error committing db transaction")
	}

	return didPartialRun, nil
}

func (c *consensusLayerEventsIndexer) GetEpochProcessedEvents(startEpoch, endEpoch uint64) (map[uint64][]types.ConsensusLayerEvent, error) {
	log.Debugf("getting epoch processed events for epochs %d to %d", startEpoch, endEpoch)
	if startEpoch < uint64(max(0, int64(c.config.ElectraForkEpoch)-1)) {
		// we allow a lookback of 1 epoch, this is relevant for the upgrade_state event.
		// not applicable if electra gets activated during genesis
		return nil, errors.New("start epoch is before electra fork epoch")
	}
	if endEpoch < startEpoch {
		return nil, errors.New("end epoch is before start epoch")
	}
	res := make(map[uint64][]types.ConsensusLayerEvent)
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		slot := uint64(max(0, int64((epoch+1)*c.config.SlotsPerEpoch)-1))
		epochProcessedEvents, err := c.db.GetConsensusLayerEventsOfType(slot, types.EpochProcessedEventName)
		if err != nil {
			return nil, errors.Wrapf(err, "error retrieving epoch processed events for slot %d", slot)
		}
		if len(epochProcessedEvents) == 0 {
			return nil, errors.Errorf("no epoch processed events for epoch %d", epoch)
		}
		log.Debugf("found %d epoch processed events for slot %d", len(epochProcessedEvents), slot)
		res[epoch] = epochProcessedEvents
	}
	return res, nil
}

func (c *consensusLayerEventsIndexer) GetEpochBlockRoots(epochProcessedEvents map[uint64][]types.ConsensusLayerEvent) (map[uint64]types.EpochBlockRoots, error) {
	log.Debugf("getting epoch block roots for epochs %v", maps.Keys(epochProcessedEvents))
	res := make(map[uint64]types.EpochBlockRoots)
outer:
	for epoch, events := range epochProcessedEvents {
		if len(events) == 0 {
			return nil, errors.Errorf("no epoch processed events for epoch %d", epoch)
		}
		ebr, err := c.GetBlockRootsForEpoch(epoch)
		if err != nil {
			return nil, fmt.Errorf("error getting block roots for epoch %d from node: %w", epoch, err)
		}
		// log.Debugf("found hashes for epoch %d using node: %v", epoch, ebr)
		for _, event := range events {
			if bytes.Equal(event.BlockRoot, ebr.EpochBlockRoot) {
				log.Debugf("found epoch processed event for epoch %d with block root %x that matches node data", epoch, event.BlockRoot)
				res[epoch] = *ebr
				continue outer // continue to the next epoch
			}
		}
		return nil, fmt.Errorf("no epoch processed event for epoch %d found with that matches node block hash %x", epoch, ebr.EpochBlockRoot)
	}
	return res, nil
}

func (c *consensusLayerEventsIndexer) GetBlockRootsForEpoch(epoch uint64) (*types.EpochBlockRoots, error) {
	epochHashes := &types.EpochBlockRoots{
		Epoch:          epoch,
		SlotBlockRoots: make(map[uint64][]byte),
	}
	slotStart := epoch * c.config.SlotsPerEpoch
	slotEnd := (epoch+1)*c.config.SlotsPerEpoch - 1
	// get slot hashes of the current epoch with max concurrency
	var eg errgroup.Group
	muttex := &sync.Mutex{}
	for slot := slotStart; slot <= slotEnd; slot++ {
		slot := slot
		eg.Go(func() error {
			header, err := c.client.GetBlockHeader(slot)
			log.Tracef("got block header for epoch %d, slot %d", epoch, slot)
			if err != nil {
				if network.SpecificError(err) != nil && network.SpecificError(err).StatusCode == http.StatusNotFound {
					// if the block is not found, skip
					log.Tracef("skipping slot %d for epoch %d because it is not found", slot, epoch)
					return nil
				}
				return fmt.Errorf("can not get block header for epoch %d, slot %d: %w", epoch, slot, err)
			}
			if !header.Data.Canonical {
				log.Debugf("skipping slot %d for epoch %d because it is not canonical", slot, epoch)
				return nil
			}
			if !header.Finalized {
				return errors.Errorf("block header for epoch %d, slot %d is not finalized", epoch, slot)
			}
			muttex.Lock()
			defer muttex.Unlock()
			epochHashes.SlotBlockRoots[slot] = header.Data.Root
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting block roots for epoch %d: %w", epoch, err)
	}
	log.Tracef("found %d block roots for epoch %d", len(epochHashes.SlotBlockRoots), epoch)
	// if we have at least one block, set the one with the highest slot as our epoch block root
	if len(epochHashes.SlotBlockRoots) > 0 {
		// find the highest slot
		slots := maps.Keys(epochHashes.SlotBlockRoots)
		slices.Sort(slots)
		highestSlot := slots[len(slots)-1]
		epochHashes.EpochBlockRoot = epochHashes.SlotBlockRoots[highestSlot]
		epochHashes.SlotBlockRoots[slotEnd] = epochHashes.SlotBlockRoots[highestSlot] // this is because events emitted in the last slot of the epoch will have this block root
		return epochHashes, nil
	}
	lowestOverallSlot := max(0, (epoch-4)*c.config.SlotsPerEpoch)
	// the above takes care over most of the overhead during the happy path. the below is less efficient but only triggers if an entire epoch has no block
	for slot := slotStart; slot >= lowestOverallSlot; slot-- {
		header, err := c.client.GetBlockHeader(slot)
		log.Tracef("got block header for epoch %d, slot %d", epoch, slot)
		if err != nil {
			if network.SpecificError(err) != nil && network.SpecificError(err).StatusCode == http.StatusNotFound {
				// if the block is not found, skip
				log.Tracef("skipping slot %d for epoch %d because it is not found", slot, epoch)
				continue
			}
			return nil, fmt.Errorf("can not get block header for epoch %d, slot %d: %w", epoch, slot, err)
		}
		if !header.Data.Canonical {
			log.Debugf("skipping slot %d for epoch %d because it is not canonical", slot, epoch)
			continue
		}
		if epochHashes.EpochBlockRoot == nil {
			// the first block we find is the last block of the epoch. in an ideal scenario this would be the last block of the epoch (or slotStart)
			log.Tracef("found epoch hash for epoch %d, slot %d: %s", epoch, slot, hexutil.Encode(header.Data.Root))
			epochHashes.SlotBlockRoots[slotEnd] = header.Data.Root // this is because events emitted in the last slot of the epoch will have this block root
			epochHashes.EpochBlockRoot = header.Data.Root          // easier access from outside
			break
		}
	}
	// case for when we find no block before hitting the lookback limit
	if len(epochHashes.EpochBlockRoot) == 0 {
		return nil, fmt.Errorf("found no epoch hash for epoch %d", epoch)
	}
	return epochHashes, nil
}
