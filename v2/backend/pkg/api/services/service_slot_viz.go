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

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/juliangruber/go-intersect"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var currentDutiesInfo *SyncData

var currentDataMutex = &sync.RWMutex{}

func (s *Services) startSlotVizDataService() {
	for {
		startTime := time.Now()
		err := s.updateSlotVizData() // TODO: only update data if something has changed (new head slot or new head epoch)
		if err != nil {
			log.Error(err, "error updating slotviz data", 0)
		}
		log.Infof("=== slotviz data updated in %s", time.Since(startTime))
		utils.ConstantTimeDelay(startTime, 12*time.Second)
	}
}

func (s *Services) updateSlotVizData() error {
	var dutiesInfo *SyncData
	if currentDutiesInfo == nil {
		dutiesInfo = s.initDutiesInfo()
	} else {
		dutiesInfo = s.copyAndCleanDutiesInfo()
	}

	var validatorDutiesInfo []types.ValidatorDutyInfo

	// create waiting group for concurrency
	gOuter := &errgroup.Group{}

	gOuter.SetLimit(3)

	// Get the fulfilled duties
	gOuter.Go(func() error {
		startTime := time.Now()
		var err error
		validatorDutiesInfo, err = db.GetValidatorDutiesInfo(s.getMaxValidatorDutiesInfoSlot())
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
		if currentDutiesInfo != nil && currentDutiesInfo.AssignmentsFetchedForEpoch > 0 {
			minEpoch = currentDutiesInfo.AssignmentsFetchedForEpoch + 1
		}

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

				encodedRedisCachedEpochAssignments, err := s.persistentRedisDbClient.Get(ctx, key).Result()
				if err != nil {
					if epoch == headEpoch+1 {
						log.Infof("headEpoch + 1 assignments not yet available, epoch %d", epoch)
						return nil
					}
					return errors.Wrap(err, fmt.Sprintf("error getting epoch assignments data for epoch %d", epoch))
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

					if dutiesInfo.EpochAttestationDuties[uint32(attValidator)] == nil {
						dutiesInfo.EpochAttestationDuties[uint32(attValidator)] = make(map[uint32]bool, 5)
					}

					dutiesInfo.EpochAttestationDuties[uint32(attValidator)][uint32(slot)] = false // validator has an attestation scheduled for that slot

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
				for _, validator := range duty.Validators {
					if dutiesInfo.EpochAttestationDuties[uint32(validator)] == nil {
						dutiesInfo.EpochAttestationDuties[uint32(validator)] = make(map[uint32]bool, 5)
					}
					dutiesInfo.EpochAttestationDuties[uint32(validator)][uint32(attestedSlot)] = true // validator has attested for that slot
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
				err := s.readerDb.Select(&slashedPropValidators, `
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
				attSlashings := []struct {
					Attestestation1Indices pq.Int64Array `db:"attestation1_indices"`
					Attestestation2Indices pq.Int64Array `db:"attestation2_indices"`
				}{}
				slashedValidators := []uint64{}

				err := s.readerDb.Select(&attSlashings, `
				SELECT
					attestation1_indices,
					attestation2_indices
				FROM blocks_attesterslashings
				WHERE block_slot = $1`, duty.Slot)
				if err != nil {
					return err
				}

				for _, row := range attSlashings {
					inter := intersect.Simple(row.Attestestation1Indices, row.Attestestation2Indices)
					if len(inter) == 0 {
						log.WarnWithStackTrace(nil, "No intersection found for attestation violation", 0, map[string]interface{}{"slot": duty.Slot})
					}
					for _, v := range inter {
						slashedValidators = append(slashedValidators, uint64(v.(int64)))
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
func (s *Services) GetCurrentDutiesInfo() (*SyncData, func(), error) {
	currentDataMutex.RLock()

	if currentDutiesInfo == nil {
		return nil, currentDataMutex.RUnlock, errors.New("waiting for dutiesInfo to be initialized")
	}

	return currentDutiesInfo, currentDataMutex.RUnlock, nil
}

func (s *Services) initDutiesInfo() *SyncData {
	dutiesInfo := SyncData{}
	dutiesInfo.LatestSlot = uint64(0)
	dutiesInfo.SlotStatus = make(map[uint64]int8)
	dutiesInfo.SlotBlock = make(map[uint64]uint64)
	dutiesInfo.SlotSyncParticipated = make(map[uint64]map[uint64]bool)
	dutiesInfo.SlotValiPropSlashed = make(map[uint64][]uint64)
	dutiesInfo.SlotValiAttSlashed = make(map[uint64][]uint64)
	dutiesInfo.PropAssignmentsForSlot = make(map[uint64]uint64)
	dutiesInfo.SyncAssignmentsForEpoch = make(map[uint64]map[uint64]bool)
	dutiesInfo.TotalSyncAssignmentsForEpoch = make(map[uint64][]uint64)
	dutiesInfo.EpochAttestationDuties = make(map[uint32]map[uint32]bool)
	return &dutiesInfo
}

func (s *Services) copyAndCleanDutiesInfo() *SyncData {
	// deep copy & clean
	headSlot := cache.LatestEpoch.Get() * utils.Config.Chain.ClConfig.SlotsPerEpoch
	dropBelowSlot := headSlot - 2*utils.Config.Chain.ClConfig.SlotsPerEpoch

	dutiesInfo := &SyncData{
		LatestSlot:                   currentDutiesInfo.LatestSlot,
		SlotStatus:                   make(map[uint64]int8, len(currentDutiesInfo.SlotStatus)),
		SlotBlock:                    make(map[uint64]uint64, len(currentDutiesInfo.SlotBlock)),
		SlotSyncParticipated:         make(map[uint64]map[uint64]bool, len(currentDutiesInfo.SlotSyncParticipated)),
		SlotValiPropSlashed:          make(map[uint64][]uint64, len(currentDutiesInfo.SlotValiPropSlashed)),
		SlotValiAttSlashed:           make(map[uint64][]uint64, len(currentDutiesInfo.SlotValiAttSlashed)),
		PropAssignmentsForSlot:       make(map[uint64]uint64, len(currentDutiesInfo.PropAssignmentsForSlot)),
		SyncAssignmentsForEpoch:      make(map[uint64]map[uint64]bool, len(currentDutiesInfo.SyncAssignmentsForEpoch)),
		TotalSyncAssignmentsForEpoch: make(map[uint64][]uint64, len(currentDutiesInfo.TotalSyncAssignmentsForEpoch)),
		EpochAttestationDuties:       make(map[uint32]map[uint32]bool, len(currentDutiesInfo.EpochAttestationDuties)),
		AssignmentsFetchedForEpoch:   currentDutiesInfo.AssignmentsFetchedForEpoch,
	}

	// copy SlotStatus
	for slot, v := range currentDutiesInfo.SlotStatus {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.SlotStatus[slot] = v
	}

	// copy SlotBlock
	for slot, v := range currentDutiesInfo.SlotBlock {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.SlotBlock[slot] = v
	}

	// copy SlotSyncParticipated
	for slot, v := range currentDutiesInfo.SlotSyncParticipated {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.SlotSyncParticipated[slot] = make(map[uint64]bool, len(v))

		for k2, v2 := range v {
			dutiesInfo.SlotSyncParticipated[slot][k2] = v2
		}
	}

	// copy SlotValiPropSlashed
	for slot, v := range currentDutiesInfo.SlotValiPropSlashed {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.SlotValiPropSlashed[slot] = make([]uint64, 0, len(currentDutiesInfo.SlotValiAttSlashed[slot]))
		dutiesInfo.SlotValiPropSlashed[slot] = append(dutiesInfo.SlotValiAttSlashed[slot], v...)
	}

	// copy SlotValiAttSlashed
	for slot, v := range currentDutiesInfo.SlotValiAttSlashed {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.SlotValiAttSlashed[slot] = make([]uint64, 0, len(currentDutiesInfo.SlotValiAttSlashed[slot]))
		dutiesInfo.SlotValiAttSlashed[slot] = append(dutiesInfo.SlotValiAttSlashed[slot], v...)
	}

	// copy PropAssignmentsForSlot
	for slot, v := range currentDutiesInfo.PropAssignmentsForSlot {
		if slot < dropBelowSlot {
			continue
		}
		dutiesInfo.PropAssignmentsForSlot[slot] = v
	}

	// copy SyncAssignmentsForEpoch
	for epoch, v := range currentDutiesInfo.SyncAssignmentsForEpoch {
		if epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch < dropBelowSlot {
			continue
		}
		dutiesInfo.SyncAssignmentsForEpoch[epoch] = make(map[uint64]bool, len(v))

		for k2, v2 := range v {
			dutiesInfo.SyncAssignmentsForEpoch[epoch][k2] = v2
		}
	}

	// copy TotalSyncAssignmentsForEpoch
	for epoch, v := range currentDutiesInfo.TotalSyncAssignmentsForEpoch {
		if epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch < dropBelowSlot {
			continue
		}
		dutiesInfo.TotalSyncAssignmentsForEpoch[epoch] = make([]uint64, 0, len(currentDutiesInfo.TotalSyncAssignmentsForEpoch[epoch]))
		dutiesInfo.TotalSyncAssignmentsForEpoch[epoch] = append(dutiesInfo.TotalSyncAssignmentsForEpoch[epoch], v...)
	}

	// copy EpochAttestationDuties
	for validator, v := range currentDutiesInfo.EpochAttestationDuties {
		dutiesInfo.EpochAttestationDuties[validator] = make(map[uint32]bool, len(v))

		for slot, v2 := range v {
			if slot < uint32(dropBelowSlot) {
				continue
			}
			dutiesInfo.EpochAttestationDuties[validator][slot] = v2
		}

		if len(dutiesInfo.EpochAttestationDuties[validator]) == 0 {
			delete(dutiesInfo.EpochAttestationDuties, validator)
		}
	}
	return dutiesInfo
}

func (s *Services) getMaxValidatorDutiesInfoSlot() uint64 {
	headEpoch := cache.LatestEpoch.Get()
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := uint64(0)
	if headEpoch > 1 {
		minEpoch = headEpoch - 2
	}

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
	SlotSyncParticipated         map[uint64]map[uint64]bool // slot -> validatorindex -> participated
	SlotValiPropSlashed          map[uint64][]uint64        // slot -> list of slashed indexes
	SlotValiAttSlashed           map[uint64][]uint64        // slot -> list of slashed indexes
	PropAssignmentsForSlot       map[uint64]uint64          // slot -> validatorindex
	SyncAssignmentsForEpoch      map[uint64]map[uint64]bool // epoch -> validatorindex -> assigned
	TotalSyncAssignmentsForEpoch map[uint64][]uint64        // epoch -> list of assigned indexes
	EpochAttestationDuties       map[uint32]map[uint32]bool // validatorindex -> slot -> attested
	AssignmentsFetchedForEpoch   uint64
}
