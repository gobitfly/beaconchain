package dataaccess

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
)

type DataAccessInterface interface {
	GetUserDashboards(userId uint64) (t.DashboardData, error)

	CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error)
	GetValidatorDashboardOverview(userId uint64, dashboardId uint64) (t.VDBOverviewData, error)
	GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.SlotVizEpoch, error)

	GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error)
	GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummaryData, error)

	GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error)

	CloseDataAccessService()
}

type DataAccessService struct {
	dummy DummyService

	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client
}

func NewDataAccessService(cfg *types.Config) DataAccessService {
	// Create the data access service
	dataAccessService := DataAccessService{
		dummy: NewDummyService()}

	// Initialize the database
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		//TODO adjust db functions to be able to set local reader/writer without setting the global ones

		db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		})

		dataAccessService.readerDb = db.ReaderDb
		dataAccessService.writerDb = db.WriterDb
	}()

	// Initialize the bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		dataAccessService.bigtable = bt
	}()

	// Initialize the tiered cache (redis)
	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	// Initialize the persistent redis client
	wg.Add(1)
	go func() {
		defer wg.Done()
		rdc := redis.NewClient(&redis.Options{
			Addr:        utils.Config.RedisSessionStoreEndpoint,
			ReadTimeout: time.Second * 20,
		})

		if err := rdc.Ping(context.Background()).Err(); err != nil {
			log.Fatal(err, "error connecting to persistent redis store", 0)
		}
		dataAccessService.persistentRedisDbClient = rdc
	}()

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
	}

	// Return the result
	return dataAccessService
}

func (d DataAccessService) CloseDataAccessService() {
	if d.readerDb != nil {
		d.readerDb.Close()
	}
	if d.writerDb != nil {
		d.writerDb.Close()
	}
	if d.bigtable != nil {
		d.bigtable.Close()
	}
}

func (d DataAccessService) GetUserDashboards(userId uint64) (t.DashboardData, error) {
	// TODO @recy21
	return d.dummy.GetUserDashboards(userId)
}

func (d DataAccessService) CreateValidatorDashboard(userId uint64, name string, network uint64) (t.VDBPostReturnData, error) {
	// TODO @recy21
	return d.dummy.CreateValidatorDashboard(userId, name, network)
}

func (d DataAccessService) GetValidatorDashboardOverview(userId uint64, dashboardId uint64) (t.VDBOverviewData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardOverview(userId, dashboardId)
}

func (d DataAccessService) GetValidatorDashboardSlotViz(dashboardId uint64) ([]t.SlotVizEpoch, error) {
	// TODO: Get the validators from the dashboardId

	dummyValidators := []uint64{900000, 900001, 900002, 900003, 900004, 900005, 900006, 900007, 900008, 900009}
	ValidatorsMap := make(map[uint64]bool)
	for _, v := range dummyValidators {
		ValidatorsMap[v] = true
	}

	// Get min/max slot/epoch
	headEpoch := cache.LatestEpoch.Get()
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch

	minEpoch := headEpoch - 2
	maxEpoch := headEpoch + 1

	minSlot := minEpoch * slotsPerEpoch

	// Gather the assignments data
	propAssignmentsForSlot := make(map[uint64]uint64)
	attAssignmentsForSlot := make(map[uint64]map[uint64]bool)
	totalSyncAssignmentsForEpoch := make(map[uint64][]uint64)
	syncAssignmentsForEpoch := make(map[uint64]map[uint64]bool)
	for epoch := minEpoch; epoch <= maxEpoch; epoch++ {
		// Get the epoch assignments data
		key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "ea", epoch)
		encodedRedisCachedEpochAssignments, err := d.persistentRedisDbClient.Get(context.Background(), key).Result()
		if err != nil {
			return nil, err
		}

		var serializedAssignmentsData bytes.Buffer
		_, err = serializedAssignmentsData.Write([]byte(encodedRedisCachedEpochAssignments))
		if err != nil {
			return nil, err
		}
		var decodedRedisCachedEpochAssignments types.RedisCachedEpochAssignments

		dec := gob.NewDecoder(&serializedAssignmentsData)
		err = dec.Decode(&decodedRedisCachedEpochAssignments)
		if err != nil {
			return nil, err
		}

		// Save the assignments data in maps

		// Proposals
		for slot, propValidator := range decodedRedisCachedEpochAssignments.Assignments.ProposerAssignments {
			// Only add results for validators we care about
			if _, isValid := ValidatorsMap[propValidator]; isValid {
				propAssignmentsForSlot[slot] = propValidator
			}
		}

		// Attestations
		for key, attValidator := range decodedRedisCachedEpochAssignments.Assignments.AttestorAssignments {
			keyParts := strings.Split(key, "-")
			slot, err := strconv.ParseUint(keyParts[0], 10, 64)
			if err != nil {
				return nil, err
			}

			if attAssignmentsForSlot[slot] == nil {
				attAssignmentsForSlot[slot] = make(map[uint64]bool, 0)
			}
			// Only add results for validators we care about
			if _, isValid := ValidatorsMap[attValidator]; isValid {
				attAssignmentsForSlot[slot][attValidator] = true
			}
		}

		// Syncs
		totalSyncAssignmentsForEpoch[epoch] = decodedRedisCachedEpochAssignments.Assignments.SyncAssignments
		if syncAssignmentsForEpoch[epoch] == nil {
			syncAssignmentsForEpoch[epoch] = make(map[uint64]bool, 0)
		}
		for _, validator := range decodedRedisCachedEpochAssignments.Assignments.SyncAssignments {
			// Only add results for validators we care about
			if _, isValid := ValidatorsMap[validator]; isValid {
				syncAssignmentsForEpoch[epoch][validator] = true
			}
		}
	}

	// Get the fullfilled duties
	validatorDuties, err := db.GetValidatorDuties(d.readerDb, minSlot)
	if err != nil {
		return nil, err
	}

	// Restructure proposal status, attestations and sync duties
	latestSlot := uint64(0)
	slotStatus := make(map[uint64]uint64)
	slotBlock := make(map[uint64]uint64)
	slotAttested := make(map[uint64]map[uint64]bool)
	slotSyncParticipated := make(map[uint64]map[uint64]bool)
	for _, duty := range validatorDuties {
		if duty.Slot > latestSlot {
			latestSlot = duty.Slot
		}
		slotStatus[duty.Slot] = duty.Status
		slotBlock[duty.Slot] = duty.Block
		if duty.Status == 1 { // 1: Proposed
			if duty.AttestedSlot.Valid {
				attestedSlot := uint64(duty.AttestedSlot.Int64)
				if slotAttested[attestedSlot] == nil {
					slotAttested[attestedSlot] = make(map[uint64]bool, 0)
				}
				for _, validator := range duty.Validators {
					slotAttested[attestedSlot][uint64(validator)] = true
				}
			}

			if slotSyncParticipated[duty.Slot] == nil {
				slotSyncParticipated[duty.Slot] = make(map[uint64]bool, 0)

				partValidators := utils.GetParticipatingSyncCommitteeValidators(duty.SyncAggregateBits, totalSyncAssignmentsForEpoch[utils.EpochOfSlot(duty.Slot)])
				for _, validator := range partValidators {
					slotSyncParticipated[duty.Slot][validator] = true
				}
			}
		}
	}

	slotVizEpochs := make([]t.SlotVizEpoch, maxEpoch-minEpoch+1)
	for epochIdx := uint64(0); epochIdx <= maxEpoch-minEpoch; epochIdx++ {
		epoch := maxEpoch - epochIdx

		// Set the epoch number
		slotVizEpochs[epochIdx].Epoch = epoch

		// Set the slots
		slotVizEpochs[epochIdx].Slots = make([]t.VDBSlotVizSlot, slotsPerEpoch)
		for slotIdx := uint64(0); slotIdx < slotsPerEpoch; slotIdx++ {
			// Set the slot number
			slot := epoch*slotsPerEpoch + slotIdx
			slotVizEpochs[epochIdx].Slots[slotIdx].Slot = slot

			// Set the slot status
			status := "scheduled"
			switch slotStatus[slot] {
			case 0, 2:
				status = "missed"
			case 1:
				status = "proposed"
			case 3:
				status = "orphaned"
			}
			slotVizEpochs[epochIdx].Slots[slotIdx].Status = status

			// Get the proposals for the slot
			if _, ok := propAssignmentsForSlot[slot]; ok {
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposals.Validator = propAssignmentsForSlot[slot]

				status := "scheduled"
				dutyObject := slot
				switch slotStatus[slot] {
				case 0, 2:
					status = "failed"
				case 1, 3:
					status = "success"
					dutyObject = slotBlock[slot]
				}
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposals.Status = status
				slotVizEpochs[epochIdx].Slots[slotIdx].Proposals.DutyObject = dutyObject
			}

			// Get the attestation summary for the slot
			for validator := range attAssignmentsForSlot[slot] {
				if slot > latestSlot {
					slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.PendingCount++
				} else if _, ok := slotAttested[slot][validator]; ok {
					slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.SuccessCount++
				} else {
					slotVizEpochs[epochIdx].Slots[slotIdx].Attestations.FailedCount++
				}
			}

			// Get the sync summary for the slot
			for validator := range syncAssignmentsForEpoch[epoch] {
				if slot > latestSlot {
					slotVizEpochs[epochIdx].Slots[slotIdx].Sync.PendingCount++
				} else if _, ok := slotSyncParticipated[slot][validator]; ok {
					slotVizEpochs[epochIdx].Slots[slotIdx].Sync.SuccessCount++
				} else {
					slotVizEpochs[epochIdx].Slots[slotIdx].Sync.FailedCount++
				}
			}

			// TODO: Get the slashings for the slot
		}

	}

	return slotVizEpochs, nil
}

func (d DataAccessService) GetValidatorDashboardSummary(dashboardId uint64, cursor string, sort []t.Sort[t.VDBSummaryTableColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
}

func (d DataAccessService) GetValidatorDashboardGroupSummary(dashboardId uint64, groupId uint64) (t.VDBGroupSummaryData, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardGroupSummary(dashboardId, groupId)
}

func (d DataAccessService) GetValidatorDashboardBlocks(dashboardId uint64, cursor string, sort []t.Sort[t.VDBBlocksTableColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, t.Paging, error) {
	// TODO @recy21
	return d.dummy.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
}
