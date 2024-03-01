package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/barkimedes/go-deepcopy"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var currentDutiesInfo *SyncData

var currentDataMutex = &sync.RWMutex{}

func StartSlotVizDataService() {
	for {
		startTime := time.Now()
		err := updateSlotVizData() // TODO: only update data if something has changed (new head slot or new head epoch)
		if err != nil {
			log.Error(err, "error updating slotviz data", 0)
		}
		log.Infof("=== slotviz data updated in %s", time.Since(startTime))
		utils.ConstantTimeDelay(startTime, 12*time.Second)
	}
}

func updateSlotVizData() error {
	var dutiesInfo *SyncData
	currentDataMutex.RLock()
	if currentDutiesInfo == nil {
		dutiesInfo = initDutiesInfo()
	} else {
		dutiesInfo = copyAndCleanDutiesInfo()
	}
	currentDataMutex.RUnlock()

	var validatorDutiesInfo []types.ValidatorDutyInfo

	// create waiting group for concurrency
	gOuter := &errgroup.Group{}

	// Get the fulfilled duties
	gOuter.Go(func() error {
		startTime := time.Now()
		var err error
		validatorDutiesInfo, err = db.GetValidatorDutiesInfo(db.ReaderDb, getMaxValidatorDutiesInfoSlot())
		if err != nil {
			return errors.Wrap(err, "error getting validator duties info")
		}
		log.Debugf("getSlotsWithDuties: %s", time.Since(startTime))

		return nil
	})

	var maxEpochAssignmentsFetched uint64
	// Gather the assignments data
	{
		// Get min/max slot/epoch
		headEpoch := cache.LatestEpoch.Get()

		minEpoch := headEpoch - 2

		// if we have fetched epoch assignments before
		// dont load for this epoch again
		currentDataMutex.RLock()
		if currentDutiesInfo != nil && currentDutiesInfo.AssignmentsFetchedForEpoch > 0 {
			minEpoch = currentDutiesInfo.AssignmentsFetchedForEpoch + 1
		}
		currentDataMutex.RUnlock()

		maxEpoch := headEpoch + 1

		muxPropAssignmentsForSlot := &sync.Mutex{}
		muxAttAssignmentsForSlot := &sync.Mutex{}
		muxTotalSyncAssignmentsForEpoch := &sync.Mutex{}
		muxSyncAssignmentsForEpoch := &sync.Mutex{}

		for e := minEpoch; e <= maxEpoch; e++ {
			epoch := e
			gOuter.Go(func() error {
				startTime := time.Now()
				defer func() {
					log.Debugf("getEpochAssignments: %d %s", epoch, time.Since(startTime))
				}()
				// Get the epoch assignments data
				key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "ea", epoch)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()

				encodedRedisCachedEpochAssignments, err := db.PersistentRedisDbClient.Get(ctx, key).Result()
				if err != nil {
					return errors.Wrap(err, "error getting epoch assignments data")
				}

				var serializedAssignmentsData bytes.Buffer
				_, err = serializedAssignmentsData.Write([]byte(encodedRedisCachedEpochAssignments))
				if err != nil {
					return errors.Wrap(err, "error writing assignments data")
				}
				var decodedRedisCachedEpochAssignments types.RedisCachedEpochAssignments

				dec := gob.NewDecoder(&serializedAssignmentsData)
				err = dec.Decode(&decodedRedisCachedEpochAssignments)
				if err != nil {
					return errors.Wrap(err, "error decoding assignments data")
				}

				if decodedRedisCachedEpochAssignments.Assignments == nil {
					return nil // retry later
				}

				// Save the assignments data in maps

				// Proposals
				for slot, propValidator := range decodedRedisCachedEpochAssignments.Assignments.ProposerAssignments {
					muxPropAssignmentsForSlot.Lock()
					dutiesInfo.PropAssignmentsForSlot[slot] = propValidator
					muxPropAssignmentsForSlot.Unlock()
				}

				// Attestations
				for key, attValidator := range decodedRedisCachedEpochAssignments.Assignments.AttestorAssignments {
					keyParts := strings.Split(key, "-")
					slot, err := strconv.ParseUint(keyParts[0], 10, 64)
					if err != nil {
						return errors.Wrap(err, "error parsing slot")
					}

					muxAttAssignmentsForSlot.Lock()
					if dutiesInfo.AttAssignmentsForSlot[slot] == nil {
						dutiesInfo.AttAssignmentsForSlot[slot] = make(map[uint64]bool, 0)
					}

					dutiesInfo.AttAssignmentsForSlot[slot][attValidator] = true

					muxAttAssignmentsForSlot.Unlock()
				}

				// Syncs
				muxTotalSyncAssignmentsForEpoch.Lock()
				dutiesInfo.TotalSyncAssignmentsForEpoch[epoch] = decodedRedisCachedEpochAssignments.Assignments.SyncAssignments
				muxTotalSyncAssignmentsForEpoch.Unlock()
				muxSyncAssignmentsForEpoch.Lock()
				if dutiesInfo.SyncAssignmentsForEpoch[epoch] == nil {
					dutiesInfo.SyncAssignmentsForEpoch[epoch] = make(map[uint64]bool, 0)
				}
				for _, validator := range decodedRedisCachedEpochAssignments.Assignments.SyncAssignments {
					dutiesInfo.SyncAssignmentsForEpoch[epoch][validator] = true
				}
				muxSyncAssignmentsForEpoch.Unlock()

				if epoch > maxEpochAssignmentsFetched {
					maxEpochAssignmentsFetched = epoch
				}

				return nil
			})
		}
	}

	// wait for routines to complete
	if err := gOuter.Wait(); err != nil {
		log.Error(err, "error getting assignments data", 0)
		return err
	}

	// update max epoch assignments fetched after all assignments are fetched
	if maxEpochAssignmentsFetched > dutiesInfo.AssignmentsFetchedForEpoch {
		dutiesInfo.AssignmentsFetchedForEpoch = maxEpochAssignmentsFetched
	}

	// process extra data
	startTime := time.Now()
	for _, duty := range validatorDutiesInfo {
		if duty.Slot > dutiesInfo.LatestSlot {
			dutiesInfo.LatestSlot = duty.Slot
		}
		dutiesInfo.SlotStatus[duty.Slot] = duty.Status
		dutiesInfo.SlotBlock[duty.Slot] = duty.Block
		if duty.Status == 1 { // 1: Proposed
			// Attestations
			if duty.AttestedSlot.Valid {
				attestedSlot := uint64(duty.AttestedSlot.Int64)
				if dutiesInfo.SlotAttested[attestedSlot] == nil {
					dutiesInfo.SlotAttested[attestedSlot] = make(map[uint64]bool, 0)
				}
				for _, validator := range duty.Validators {
					dutiesInfo.SlotAttested[attestedSlot][uint64(validator)] = true
				}
			}
			// Syncs
			if dutiesInfo.SlotSyncParticipated[duty.Slot] == nil {
				dutiesInfo.SlotSyncParticipated[duty.Slot] = make(map[uint64]bool, 0)

				partValidators := utils.GetParticipatingSyncCommitteeValidators(duty.SyncAggregateBits, dutiesInfo.TotalSyncAssignmentsForEpoch[utils.EpochOfSlot(duty.Slot)])
				for _, validator := range partValidators {
					dutiesInfo.SlotSyncParticipated[duty.Slot][validator] = true
				}
			}
			// Slashings
			if duty.ProposerSlashingsCount > 0 {
				slashedPropValidators := []uint64{}
				err := db.ReaderDb.Select(&slashedPropValidators, `
					SELECT
						proposerindex
					FROM blocks_proposerslashings
					WHERE block_slot = $1`, duty.Slot)
				if err != nil {
					return err
				}
				dutiesInfo.SlotValiPropSlashed[duty.Slot] = slashedPropValidators
			}
			if duty.AttesterSlashingsCount > 0 {
				slashedAttValidators := []pq.Int64Array{}
				slashedValidators := []uint64{}

				err := db.ReaderDb.Select(&slashedAttValidators, `
					SELECT
						attestation2_indices
					FROM blocks_attesterslashings
					WHERE block_slot = $1`, duty.Slot)
				if err != nil {
					return err
				}

				for _, slashedAttValidator := range slashedAttValidators {
					for _, validator := range slashedAttValidator {
						slashedValidators = append(slashedValidators, uint64(validator))
					}
				}
				dutiesInfo.SlotValiAttSlashed[duty.Slot] = slashedValidators
			}
		}
	}
	log.Debugf("process slotduties extra data: %s", time.Since(startTime))

	// update currentDutiesInfo and hence frontend data
	currentDataMutex.Lock()
	if currentDutiesInfo == nil { // info on first iteration
		log.Infof("== slot-viz data updater initialized ==")
	}
	currentDutiesInfo = dutiesInfo
	currentDataMutex.Unlock()

	return nil
}

// GetCurrentDutiesInfo returns the current duties info and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the slot viz service from updating
func GetCurrentDutiesInfo() (*SyncData, func(), error) {
	currentDataMutex.RLock()

	if currentDutiesInfo == nil {
		return nil, currentDataMutex.RUnlock, errors.New("waiting for dutiesInfo to be initialized")
	}

	return currentDutiesInfo, currentDataMutex.RUnlock, nil
}

func initDutiesInfo() *SyncData {
	dutiesInfo := SyncData{}
	dutiesInfo.LatestSlot = uint64(0)
	dutiesInfo.SlotStatus = make(map[uint64]int8)
	dutiesInfo.SlotBlock = make(map[uint64]uint64)
	dutiesInfo.SlotAttested = make(map[uint64]map[uint64]bool)
	dutiesInfo.SlotSyncParticipated = make(map[uint64]map[uint64]bool)
	dutiesInfo.SlotValiPropSlashed = make(map[uint64][]uint64)
	dutiesInfo.SlotValiAttSlashed = make(map[uint64][]uint64)
	dutiesInfo.PropAssignmentsForSlot = make(map[uint64]uint64)
	dutiesInfo.AttAssignmentsForSlot = make(map[uint64]map[uint64]bool)
	dutiesInfo.SyncAssignmentsForEpoch = make(map[uint64]map[uint64]bool)
	dutiesInfo.TotalSyncAssignmentsForEpoch = make(map[uint64][]uint64)
	return &dutiesInfo
}

func copyAndCleanDutiesInfo() *SyncData {
	dutiesInfoCopyResult, err := deepcopy.Anything(*currentDutiesInfo)
	if err != nil {
		log.Fatal(err, "error creating deep copy of currentDutiesInfo object", 0)
	}
	dutiesInfo := dutiesInfoCopyResult.(SyncData)

	// cleaning
	headSlot := cache.LatestEpoch.Get() * utils.Config.Chain.ClConfig.SlotsPerEpoch
	dropBelowSlot := headSlot - 2*utils.Config.Chain.ClConfig.SlotsPerEpoch
	for slot := range dutiesInfo.SlotStatus {
		if slot < dropBelowSlot {
			delete(dutiesInfo.SlotStatus, slot)
			delete(dutiesInfo.SlotBlock, slot)
			delete(dutiesInfo.SlotAttested, slot)
			delete(dutiesInfo.SlotSyncParticipated, slot)
			delete(dutiesInfo.SlotValiPropSlashed, slot)
			delete(dutiesInfo.SlotValiAttSlashed, slot)
		}
	}
	return &dutiesInfo
}

func getMaxValidatorDutiesInfoSlot() uint64 {
	headEpoch := cache.LatestEpoch.Get()
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2

	/*
		Why reduce minEpoch to headEpoch - 1 after first iteration?
		- Attestations can be included until last slot of N+1 epoch (deneb), so head - 2 can not get new attestation data
		- Attestation data amount is the main culprit for the database call since it returns huge amounts of data
		- Other fields used by slotviz do not change as well (sync bits, exec block). If we at some point include changing fields for headEpoch -2
		  we should consider making this a separate call to avoid loading all attestation data again
	*/
	if currentDutiesInfo != nil && currentDutiesInfo.AssignmentsFetchedForEpoch > 0 { // if we have fetched epoch assignments before
		minEpoch = headEpoch - 1
	}

	maxValidatorDutiesInfoSlot := minEpoch * slotsPerEpoch

	return maxValidatorDutiesInfoSlot
}

type SyncData struct {
	LatestSlot                   uint64
	SlotStatus                   map[uint64]int8            // slot -> status
	SlotBlock                    map[uint64]uint64          // slot -> block
	SlotAttested                 map[uint64]map[uint64]bool // slot -> validatorindex -> attested
	SlotSyncParticipated         map[uint64]map[uint64]bool // slot -> validatorindex -> participated
	SlotValiPropSlashed          map[uint64][]uint64        // slot -> list of slashed indexes
	SlotValiAttSlashed           map[uint64][]uint64        // slot -> list of slashed indexes
	PropAssignmentsForSlot       map[uint64]uint64          // slot -> validatorindex
	AttAssignmentsForSlot        map[uint64]map[uint64]bool // slot -> validatorindex -> assigned
	SyncAssignmentsForEpoch      map[uint64]map[uint64]bool // epoch -> validatorindex -> assigned
	TotalSyncAssignmentsForEpoch map[uint64][]uint64        // epoch -> list of assigned indexes
	AssignmentsFetchedForEpoch   uint64
}
