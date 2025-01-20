package modules

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/google/uuid"

	//"github.com/fjl/memsize/memsizeui"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	phase0HotfixMutex sync.Mutex
	latestSafeEpoch   atomic.Int64
	heavySemaphore    *semaphore.Weighted
	mediumSemaphore   *semaphore.Weighted
	lightSemaphore    *semaphore.Weighted
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}

	heavyWeight := utils.Config.DashboardExporter.FetchHeavyInParallel
	if heavyWeight <= 0 {
		heavyWeight = 8
	}
	mediumWeight := utils.Config.DashboardExporter.FetchMediumInParallel
	if mediumWeight <= 0 {
		mediumWeight = 18
	}
	lightWeight := utils.Config.DashboardExporter.FetchLightInParallel
	if lightWeight <= 0 {
		lightWeight = 128
	}
	// we use multiple semaphores because nodes are usually happy serving multiple light requests at once even when they are busy with heavy ones
	temp.heavySemaphore = semaphore.NewWeighted(heavyWeight)
	temp.mediumSemaphore = semaphore.NewWeighted(mediumWeight)
	temp.lightSemaphore = semaphore.NewWeighted(lightWeight)
	return temp
}

type Task struct {
	UUID     uuid.UUID `db:"uuid"`
	Hostname string    `db:"hostname"`
	Priority int64     `db:"priority"`
	StartTs  time.Time `db:"start_ts"`
	EndTs    time.Time `db:"end_ts"`
	Status   string    `db:"status"`
}

func (d *dashboardData) Init() error {
	// blocking loop trying to init d.latestSafeEpoch
	for {
		err := updateSafeEpoch(d)
		if err != nil {
			d.log.Error(err, "failed to update safe epoch", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	go d.insertTask()      // does all the inserting of the data
	go d.maintenanceTask() // does all the transferring of the data
	go d.rollingTask()     // does all the rolling of the data

	return nil
}

var EpochsWritten int
var FirstEpochWritten *time.Time

func (d *dashboardData) OnFinalizedCheckpoint(t *constypes.StandardFinalizedCheckpointResponse) error {
	return nil
}

func updateSafeEpoch(d *dashboardData) error {
	res, err := d.CL.GetFinalityCheckpoints("head")
	if err != nil {
		return err
	}

	finalized := res.Data.Finalized.Epoch
	safe := int64(res.Data.Finalized.Epoch) - 2

	metrics.State.WithLabelValues("dashboard_data_exporter_latest_safe_epoch").Set(float64(safe))
	metrics.State.WithLabelValues("dashboard_data_exporter_latest_finalized_epoch").Set(float64(finalized))

	d.latestSafeEpoch.Store(safe)
	return nil
}

func (d *dashboardData) GetName() string {
	return "Dashboard-Data"
}

func (d *dashboardData) GetMonitoringEventId() constants.Event {
	return constants.Event_ExporterModuleDashboardData
}

func (d *dashboardData) OnHead(event *constypes.StandardEventHeadResponse) error {
	// you may ask, why here and not OnFinalizedCheckpoint?
	// because due to our loadbalanced node architecture we sometimes receive the finalized checkpoint event
	// before the node we hit has updated its own finalized checkpoint, causing us to be off by 1 epoch sometimes
	// so we simply check more often. the request overhead is minimal anyways
	err := updateSafeEpoch(d)
	if err != nil {
		return err
	}

	return nil
}

func (d *dashboardData) OnChainReorg(event *constypes.StandardEventChainReorg) error {
	return nil
}

type MultiEpochData struct {
	// needs sorting
	epochBasedData struct {
		epochs          []uint64
		tarIndices      []int
		tarOffsets      []int
		validatorStates map[int64]constypes.LightStandardValidatorsResponse // epoch => state
		rewards         struct {
			attestationRewards      map[uint64][]constypes.AttestationReward               // epoch => validator index => reward
			attestationIdealRewards map[uint64]map[uint64]constypes.AttestationIdealReward // epoch => effective balance => reward
		}
		electraDeposits       map[uint64][]constypes.ElectraDeposit       // epoch => deposits
		electraConsolidations map[uint64][]constypes.ElectraConsolidation // epoch => consolidations
	}
	validatorBasedData struct {
		// mapping pubkey => validator index
		validatorIndices map[string]uint64
	}
	syncPeriodBasedData struct {
		// sync committee period => assignments
		SyncAssignments map[uint64][]uint64
		// sync committee period => state
		SyncStateEffectiveBalances map[uint64][]uint64
	}
	slotBasedData struct {
		blocks      map[uint64]constypes.LightAnySignedBlock // slot => block, if nil = missed. will include blocks for one more epoch than needed because attestations can be included an epoch later
		assignments struct {
			attestationAssignments map[uint64][][]uint64 // slot => committee index => validator index
			blockAssignments       map[uint64]uint64     // slot => validator index
		}
		rewards struct {
			syncCommitteeRewards map[uint64]constypes.StandardSyncCommitteeRewardsResponse // slot => sync committee rewards
			blockRewards         map[uint64]constypes.StandardBlockRewardsResponse         // slot => block reward data
		}
	}
}

// factory
func NewMultiEpochData(epochCount int) MultiEpochData {
	// allocate all maps
	data := MultiEpochData{}
	data.epochBasedData.validatorStates = make(map[int64]constypes.LightStandardValidatorsResponse, epochCount)
	data.epochBasedData.tarIndices = make([]int, epochCount)
	data.epochBasedData.tarOffsets = make([]int, epochCount)
	data.epochBasedData.rewards.attestationRewards = make(map[uint64][]constypes.AttestationReward, epochCount)
	data.epochBasedData.rewards.attestationIdealRewards = make(map[uint64]map[uint64]constypes.AttestationIdealReward, epochCount)
	data.epochBasedData.electraDeposits = make(map[uint64][]constypes.ElectraDeposit, epochCount)
	data.epochBasedData.electraConsolidations = make(map[uint64][]constypes.ElectraConsolidation, epochCount)
	slotCount := epochCount * int(utils.Config.Chain.ClConfig.SlotsPerEpoch)
	data.slotBasedData.blocks = make(map[uint64]constypes.LightAnySignedBlock, slotCount)
	data.slotBasedData.assignments.attestationAssignments = make(map[uint64][][]uint64, slotCount)
	data.slotBasedData.assignments.blockAssignments = make(map[uint64]uint64, slotCount)
	data.slotBasedData.rewards.syncCommitteeRewards = make(map[uint64]constypes.StandardSyncCommitteeRewardsResponse, slotCount)
	data.slotBasedData.rewards.blockRewards = make(map[uint64]constypes.StandardBlockRewardsResponse, slotCount)
	data.validatorBasedData.validatorIndices = make(map[string]uint64)
	return data
}

func (d *dashboardData) getDataForEpochRange(epochStart, epochEnd uint64, tar *MultiEpochData) error {
	g1 := &errgroup.Group{}
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_overall").Observe(time.Since(start).Seconds())
	}()
	// prefill epochBasedData.epochs
	for i := epochStart; i <= epochEnd; i++ {
		tar.epochBasedData.epochs = append(tar.epochBasedData.epochs, i)
	}

	timer := time.NewTicker(3 * time.Second)
	defer timer.Stop()
	go func() {
		report := func(sem *semaphore.Weighted, name string) {
			v := reflect.ValueOf(sem)
			cur := v.Elem().FieldByName("cur").Int()
			size := v.Elem().FieldByName("size").Int()
			// waiters is a struct that has a len field
			waiters := v.Elem().FieldByName("waiters").FieldByName("len").Int()
			d.log.Debugf("%s: cur: %d, size: %d, waiters: %d", name, cur, size, waiters)
		}
		for {
			_, ok := <-timer.C
			if !ok {
				return
			}
			report(d.heavySemaphore, "heavy")
			report(d.mediumSemaphore, "medium")
			report(d.lightSemaphore, "light")
		}
	}()

	// epoch based Data
	g1.Go(func() error {
		return d.fetchEpochValidatorStates(epochStart, epochEnd, tar)
	})
	// syncPeriodBasedData
	g1.Go(func() error {
		return d.fetchSyncCommitteeAssignments(epochStart, epochEnd, tar)
	})
	// blocks
	g1.Go(func() error {
		return d.fetchBlocks(epochStart, epochEnd, tar)
	})
	// block rewards
	g1.Go(func() error {
		return d.fetchBlockRewards(epochStart, epochEnd, tar)
	})
	// GetSyncRewards
	g1.Go(func() error {
		return d.fetchSyncCommitteeRewards(epochStart, epochEnd, tar)
	})
	// block assignments
	g1.Go(func() error {
		return d.fetchBlockAssignments(epochStart, epochEnd, tar)
	})
	// attestation rewards
	g1.Go(func() error {
		return d.fetchAttestationRewards(epochStart, epochEnd, tar)
	})
	// attestation assignments
	g1.Go(func() error {
		return d.fetchAttestationAssignments(epochStart, epochEnd, tar)
	})
	// electra deposits using GetDebugState
	g1.Go(func() error {
		return d.fetchElectraDeposits(epochStart, epochEnd, tar)
	})
	// electra consolidations using GetDebugState
	g1.Go(func() error {
		return d.fetchElectraConsolidations(epochStart, epochEnd, tar)
	})
	// wait for all tasks to finish
	err := g1.Wait()
	if err != nil {
		return fmt.Errorf("error in getDataForEpochRange: %w", err)
	}
	return nil
}

func (d *dashboardData) WorkaroundGetBlockHashForEpoch(epoch uint64) ([]byte, error) {
	// get the workaround block hashes
	hashes, err := db.WorkaroundGetEpochProcessedHashes(epoch)
	if err != nil {
		return nil, fmt.Errorf("can not get workaround block hashes for epoch %d: %w", epoch, err)
	}
	if len(hashes) == 0 {
		// this is bad
		return nil, fmt.Errorf("no workaround block hashes for epoch %d", epoch)
	}
	var canonicalHash []byte
	for _, hash := range hashes {
		// look up hash using beacon api
		block, err := d.CL.GetBlockHeader(fmt.Sprintf("0x%x", hash))
		if err != nil {
			// if not found, skip, means that the block was missed or orphaned
			httpErr := network.SpecificError(err)
			if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
				d.log.Tracef("skipping workaround block hash %s for epoch %d because it was not found", hexutil.Encode(hash), epoch)
				continue
			}
			return nil, fmt.Errorf("can not get block header for workaround block hash %s for epoch %d: %w", hexutil.Encode(hash), epoch, err)
		}
		if block.Data.Canonical {
			canonicalHash = block.Data.Root
			break
		}
		d.log.Tracef("skipping workaround block hash %s for epoch %d because it is not canonical", hexutil.Encode(hash), epoch)
	}
	if canonicalHash == nil {
		return nil, fmt.Errorf("no canonical workaround block hash for epoch %d", epoch)
	}
	return canonicalHash, nil
}

func (d *dashboardData) fetchElectraDeposits(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_electra_deposits_overall").Observe(time.Since(start).Seconds())
	}()
	// get electra deposits
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	for i := epochStart; i <= epochEnd; i++ {
		epoch := i
		// ignore if epoch is smaller than electra fork
		// we use <= instead of < because the first state transition that will include electra deposits is one epoch after the fork
		if epoch <= utils.Config.Chain.ClConfig.ElectraForkEpoch {
			d.log.Tracef("skipping epoch %d for electra deposits because it is before the electra fork", epoch)
			continue
		}
		g2.Go(func() error {
			// acquiring semaphore
			err := d.lightSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.lightSemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_electra_deposits_single").Observe(time.Since(start).Seconds())
			}()
			blockhash, err := d.WorkaroundGetBlockHashForEpoch(epoch)
			if err != nil {
				d.log.Error(err, "can not get workaround block hash for electra deposits", 0, map[string]interface{}{"epoch": epoch})
				return err
			}
			// get workaround deposits event using block hash
			deposits, err := db.WorkaroundGetProcessedDeposits(blockhash)
			if err != nil {
				return fmt.Errorf("can not get workaround deposits for epoch %d: %w", epoch, err)
			}
			if len(deposits) == 0 {
				// this is okay
				d.log.Tracef("no workaround deposits for epoch %d", epoch)
				return nil
			}
			// processedDeposits is now a list of deposits that were processed in the block of slot n
			d.log.Tracef("processed deposits for epoch %d: %v", epoch, deposits)
			writeMutex.Lock()
			tar.epochBasedData.electraDeposits[epoch] = deposits
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in electra deposits: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchElectraConsolidations(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_electra_consolidations_overall").Observe(time.Since(start).Seconds())
	}()
	// get electra consolidations
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	for i := epochStart; i <= epochEnd; i++ {
		epoch := i
		if epoch <= utils.Config.Chain.ClConfig.ElectraForkEpoch {
			d.log.Tracef("skipping epoch %d for electra consolidations because it is before the electra fork", epoch)
			continue
		}
		g2.Go(func() error {
			// acquiring semaphore
			err := d.lightSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.lightSemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_electra_consolidations_single").Observe(time.Since(start).Seconds())
			}()
			blockhash, err := d.WorkaroundGetBlockHashForEpoch(epoch)
			if err != nil {
				d.log.Error(err, "can not get workaround block hash for electra consolidations", 0, map[string]interface{}{"epoch": epoch})
				return err
			}
			// get workaround consolidations event using block hash
			consolidations, err := db.WorkaroundGetProcessedConsolidations(blockhash)
			if err != nil {
				return fmt.Errorf("can not get workaround consolidations for epoch %d: %w", epoch, err)
			}
			if len(consolidations) == 0 {
				// this is okay
				d.log.Tracef("no workaround consolidations for epoch %d", epoch)
				return nil
			}
			// processedConsolidations is now a list of consolidations that were processed in the block of slot n
			d.log.Tracef("processed consolidations for epoch %d: %v", epoch, consolidations)
			// write to tar
			writeMutex.Lock()
			tar.epochBasedData.electraConsolidations[epoch] = consolidations
			writeMutex.Unlock()

			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in electra consolidations: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchAttestationAssignments(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "attestationAssignments", "duration_type": "total"}).Observe(time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_attestation_assignments_overall").Observe(time.Since(start).Seconds())
	}()
	// get attestation assignments
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	for e := epochStart; e <= epochEnd; e++ {
		epoch := e
		g2.Go(func() error {
			// fetch assignment using last fetchSlot in epoch. somehow thats faster than using the first fetchSlot. dont ask why
			fetchSlot := (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1
			err := d.heavySemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.heavySemaphore.Release(1)
			start := time.Now()
			defer func() {
				//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "attestationAssignments", "duration_type": "single"}).Observe(time.Since(start).Seconds())
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_attestation_assignments_single").Observe(time.Since(start).Seconds())
			}()
			data, err := d.CL.GetCommittees(fetchSlot, nil, nil, nil)
			if err != nil {
				d.log.Error(err, "can not get attestation assignments", 0, map[string]interface{}{"slot": fetchSlot})
				return err
			}
			writeMutex.Lock()
			for _, committee := range data.Data {
				// todo replace with single alloc variant that uses config values (config has 0 when the code hits here)
				if _, ok := tar.slotBasedData.assignments.attestationAssignments[committee.Slot]; !ok {
					tar.slotBasedData.assignments.attestationAssignments[committee.Slot] = make([][]uint64, utils.Config.Chain.ClConfig.MaxCommitteesPerSlot)
				}
				if len(tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index]) == 0 {
					tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index] = make([]uint64, len(committee.Validators))
				}
				for i, valIndex := range committee.Validators {
					tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index][i] = uint64(valIndex)
				}
			}
			// use slices.DeleteFunc to remove nils from the slice
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in attestation assignments: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchAttestationRewards(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "attestationRewards", "duration_type": "total"}).Observe(time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_epoch_based_data_attestation_rewards_overall").Observe(time.Since(start).Seconds())
	}()
	// get attestation rewards
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	// once per epoch, no extra epochs needed
	for e := epochStart; e <= epochEnd; e++ {
		epoch := e
		g2.Go(func() error {
			// acquiring semaphore
			err := d.heavySemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.heavySemaphore.Release(1)
			start := time.Now()
			defer func() {
				//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "attestationRewards", "duration_type": "single"}).Observe(time.Since(start).Seconds())
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_epoch_based_data_attestation_rewards_single").Observe(time.Since(start).Seconds())
			}()
			data, err := d.CL.GetAttestationRewards(epoch)
			if err != nil {
				d.log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
				return err
			}
			// ideal
			ideal := make(map[uint64]constypes.AttestationIdealReward)
			for _, idealReward := range data.Data.IdealRewards {
				ideal[uint64(idealReward.EffectiveBalance)] = idealReward
			}
			writeMutex.Lock()
			tar.epochBasedData.rewards.attestationRewards[epoch] = data.Data.TotalRewards
			tar.epochBasedData.rewards.attestationIdealRewards[epoch] = ideal
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in attestation rewards: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchBlockAssignments(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockAssignments", "duration_type": "total"}).Observe(time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_assignments_overall").Observe(time.Since(start).Seconds())
	}()
	// get block assignments
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	for e := epochStart; e <= epochEnd; e++ {
		epoch := e
		g2.Go(func() error {
			err := d.mediumSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.mediumSemaphore.Release(1)
			start := time.Now()
			defer func() {
				//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockAssignments", "duration_type": "single"}).Observe(time.Since(start).Seconds())
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_assignments_single").Observe(time.Since(start).Seconds())
			}()
			data, err := d.CL.GetProposalAssignments(epoch)
			if err != nil {
				d.log.Error(err, "can not get block assignments", 0, map[string]interface{}{"epoch": epoch})
				return err
			}
			writeMutex.Lock()
			for _, p := range data.Data {
				tar.slotBasedData.assignments.blockAssignments[uint64(p.Slot)] = p.ValidatorIndex
			}
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in block assignments: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchSyncCommitteeRewards(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_sync_rewards_overall").Observe(time.Since(start).Seconds())
	}()
	// get sync rewards
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	lastSlotToFetch := ((epochEnd + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1
	for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
		slot := i
		epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
		// check if slot is post hardfork
		if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			d.log.Tracef("skipping sync rewards for slot %d (before altair)", slot)
			continue
		}
		g2.Go(func() error {
			// acquiring semaphore
			err := d.mediumSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.mediumSemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_sync_rewards_single").Observe(time.Since(start).Seconds())
			}()
			data, err := d.CL.GetSyncRewards(slot)
			if err != nil {
				httpErr := network.SpecificError(err)
				if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
					d.log.Tracef("no sync rewards for slot %d", slot)
					return nil
				}
				d.log.Error(err, "can not get sync rewards", 0, map[string]interface{}{"slot": slot})
				return err
			}
			writeMutex.Lock()
			tar.slotBasedData.rewards.syncCommitteeRewards[slot] = *data
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in sync rewards: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchBlockRewards(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		// metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockRewards", "duration_type": "total"}).Observe(time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_rewards_overall").Observe(time.Since(start).Seconds())
	}()
	// get block rewards
	g2 := &errgroup.Group{}
	writeMutex := &sync.Mutex{}
	// we will fetch more than the requested epoch range because there is an "median cl reward" column for missed proposals
	// slots:
	buffer := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
	firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if firstSlotToFetch >= buffer {
		firstSlotToFetch -= buffer
	}
	lastSlotToFetch := ((epochEnd + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch) + buffer - 1
	for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
		slot := i
		g2.Go(func() error {
			// acquiring semaphore
			err := d.lightSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.lightSemaphore.Release(1)
			start := time.Now()
			defer func() {
				//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockRewards", "duration_type": "single"}).Observe(time.Since(start).Seconds())
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_rewards_single").Observe(time.Since(start).Seconds())
			}()
			data, err := d.CL.GetProposalRewards(slot)
			if err != nil {
				httpErr := network.SpecificError(err)
				if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
					d.log.Infof("no block rewards for slot %d", slot)
					return nil
				}
				d.log.Error(err, "can not get block rewards", 0, map[string]interface{}{"slot": slot})
				return err
			}
			writeMutex.Lock()
			tar.slotBasedData.rewards.blockRewards[slot] = *data
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in block rewards: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchBlocks(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "slotBasedData", "duration_type": "total"}).Observe(time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_overall").Observe(time.Since(start).Seconds())
	}()
	// get blocks
	g2 := &errgroup.Group{}
	slots := make([]uint64, 0)
	// first slot of the previous epoch
	firstSlotToFetch := (epochStart) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	if epochStart == 0 {
		firstSlotToFetch = 0
	}
	lastSlotToFetch := ((epochEnd + 2) * utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1
	for i := firstSlotToFetch; i <= lastSlotToFetch; i++ {
		slots = append(slots, i)
	}
	writeMutex := &sync.Mutex{}
	tar.slotBasedData.blocks = make(map[uint64]constypes.LightAnySignedBlock, len(slots))
	for _, s := range slots {
		slot := s
		epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
		g2.Go(func() error {
			d.log.Tracef("fetching block at slot %d", slot)
			err := d.lightSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.lightSemaphore.Release(1)
			start := time.Now()
			defer func() {
				//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "slotBasedData", "duration_type": "single"}).Observe(time.Since(start).Seconds())
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_single").Observe(time.Since(start).Seconds())
			}()

			block, err := d.CL.GetSlot(slot)
			if err != nil {
				httpErr := network.SpecificError(err)
				if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
					d.log.Tracef("no block at slot %d", slot)
					return nil
				}
				d.log.Error(err, "can not get block", 0, map[string]interface{}{"slot": slot})
				return err
			}
			// header
			header, err := d.CL.GetBlockHeader(slot)
			if err != nil {
				d.log.Error(err, "can not get block header", 0, map[string]interface{}{"slot": slot})
				return err
			}
			var lightBlock constypes.LightAnySignedBlock
			lightBlock.Slot = block.Data.Message.Slot
			lightBlock.BlockRoot = header.Data.Root
			lightBlock.ParentRoot = header.Data.Header.Message.ParentRoot
			lightBlock.ProposerIndex = block.Data.Message.ProposerIndex
			lightBlock.Attestations = block.Data.Message.Body.Attestations
			// deposits
			lightBlock.Deposits = append(lightBlock.Deposits, block.Data.Message.Body.Deposits...)
			// withdrawals
			if epoch >= utils.Config.Chain.ClConfig.CapellaForkEpoch {
				for _, w := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
					lightBlock.Withdrawals = append(lightBlock.Withdrawals, constypes.LightWithdrawal{
						Amount:         w.Amount,
						ValidatorIndex: w.ValidatorIndex,
					})
				}
			}
			// AttesterSlashings
			for _, s := range block.Data.Message.Body.AttesterSlashings {
				lightBlock.SlashedIndices = append(lightBlock.SlashedIndices, s.GetSlashedIndices()...)
			}
			// ProposerSlashings
			for _, s := range block.Data.Message.Body.ProposerSlashings {
				lightBlock.SlashedIndices = append(lightBlock.SlashedIndices, s.SignedHeader1.Message.ProposerIndex)
			}
			if epoch >= utils.Config.Chain.ClConfig.AltairForkEpoch {
				// sync
				lightBlock.SyncAggregate = block.Data.Message.Body.SyncAggregate
			}
			// free up memory
			block = nil

			writeMutex.Lock()
			tar.slotBasedData.blocks[slot] = lightBlock
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in slotBasedData: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchSyncCommitteeAssignments(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_sync_period_based_data_overall").Observe(time.Since(start).Seconds())
	}()
	// get sync committee assignments
	g2 := &errgroup.Group{}
	syncPeriodAssignmentsToFetch := make([]uint64, 0)
	snycPeriodStatesToFetch := make([]uint64, 0)
	for i := epochStart; i <= epochEnd; i++ {
		if i < utils.Config.Chain.ClConfig.AltairForkEpoch {
			d.log.Tracef("skipping sync committee assignments for epoch %d (before altair)", i)
			continue
		}
		syncPeriod := utils.SyncPeriodOfEpoch(i)
		// if we dont have the assignment yet fetch it
		if len(syncPeriodAssignmentsToFetch) == 0 || syncPeriodAssignmentsToFetch[len(syncPeriodAssignmentsToFetch)-1] != syncPeriod {
			syncPeriodAssignmentsToFetch = append(syncPeriodAssignmentsToFetch, syncPeriod)
		}
		if utils.FirstEpochOfSyncPeriod(syncPeriod) == i {
			snycPeriodStatesToFetch = append(snycPeriodStatesToFetch, syncPeriod)
		}
	}
	d.log.Infof("fetching sync committee assignments and states for sync periods %v", syncPeriodAssignmentsToFetch)
	writeMutex := &sync.Mutex{}
	tar.syncPeriodBasedData.SyncAssignments = make(map[uint64][]uint64, len(syncPeriodAssignmentsToFetch))
	tar.syncPeriodBasedData.SyncStateEffectiveBalances = make(map[uint64][]uint64, len(snycPeriodStatesToFetch))
	// assignments
	for _, s := range syncPeriodAssignmentsToFetch {
		syncPeriod := s
		g2.Go(func() error {
			// acquiring semaphore
			err := d.mediumSemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.mediumSemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_sync_period_based_data_assignments_single").Observe(time.Since(start).Seconds())
			}()
			relevantSlot := utils.FirstEpochOfSyncPeriod(syncPeriod) * utils.Config.Chain.ClConfig.SlotsPerEpoch
			assignments, err := d.CL.GetSyncCommitteesAssignments(nil, relevantSlot)
			if err != nil {
				d.log.Error(err, "can not get sync committee assignments", 0, map[string]interface{}{"syncPeriod": syncPeriod})
				return err
			}
			writeMutex.Lock()
			tar.syncPeriodBasedData.SyncAssignments[syncPeriod] = make([]uint64, len(assignments.Data.Validators))
			for i, a := range assignments.Data.Validators {
				tar.syncPeriodBasedData.SyncAssignments[syncPeriod][i] = uint64(a)
			}
			writeMutex.Unlock()
			return nil
		})
	}
	// states
	for _, s := range snycPeriodStatesToFetch {
		syncPeriod := s
		g2.Go(func() error {
			slot := utils.FirstEpochOfSyncPeriod(syncPeriod) * utils.Config.Chain.ClConfig.SlotsPerEpoch
			// acquiring semaphore
			err := d.heavySemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.heavySemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_sync_period_based_data_states_single").Observe(time.Since(start).Seconds())
			}()
			valis, err := d.CL.GetValidators(slot, nil, nil)
			if err != nil {
				d.log.Error(err, "can not get sync committee state", 0, map[string]interface{}{"syncPeriod": syncPeriod})
				return err
			}
			// convert to light validators
			dat := make([]uint64, len(valis.Data))
			for i, val := range valis.Data {
				if val.Status.IsActive() {
					dat[i] = val.Validator.EffectiveBalance
				}
			}
			writeMutex.Lock()
			tar.syncPeriodBasedData.SyncStateEffectiveBalances[syncPeriod] = dat
			writeMutex.Unlock()
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in syncPeriodBasedData: %w", err)
	}
	return nil
}

func (d *dashboardData) fetchEpochValidatorStates(epochStart uint64, epochEnd uint64, tar *MultiEpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_epoch_based_data_overall").Observe(time.Since(start).Seconds())
	}()
	// get states
	g2 := &errgroup.Group{}
	slots := make([]uint64, 0)
	// first slot of the first epoch
	firstEpochToFetch := epochStart
	for i := firstEpochToFetch; i <= epochEnd+1; i++ {
		if i == 0 {
			slots = append(slots, 0)
			continue
		}
		slots = append(slots, uint64(i)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1)
	}
	writeMutex := &sync.Mutex{}
	d.log.Debugf("fetching states for epochs %d to %d using slots %v", epochStart, epochEnd, slots)
	tar.epochBasedData.validatorStates = make(map[int64]constypes.LightStandardValidatorsResponse, len(slots))
	startEpoch := int64(epochStart) - 1
	for i, s := range slots {
		slot := s
		virtualEpoch := startEpoch + int64(i)
		g2.Go(func() error {
			// acquiring semaphore
			err := d.heavySemaphore.Acquire(context.Background(), 1)
			if err != nil {
				return err
			}
			defer d.heavySemaphore.Release(1)
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_epoch_based_data_single").Observe(time.Since(start).Seconds())
			}()
			var valis *constypes.StandardValidatorsResponse
			if slot == 0 {
				valis, err = d.CL.GetValidators("genesis", nil, nil)
			} else {
				valis, err = d.CL.GetValidators(slot, nil, nil)
			}
			if err != nil {
				d.log.Error(err, "can not get validators state", 0, map[string]interface{}{"slot": slot})
				return err
			}
			// convert to light validators
			var lightValis constypes.LightStandardValidatorsResponse
			lightValis.Data = make([]constypes.LightStandardValidator, len(valis.Data))
			for i, val := range valis.Data {
				lightValis.Data[i] = constypes.LightStandardValidator{
					Index:            val.Index,
					Balance:          val.Balance,
					Status:           val.Status,
					Pubkey:           val.Validator.Pubkey,
					EffectiveBalance: val.Validator.EffectiveBalance,
					Slashed:          val.Validator.Slashed,
				}
			}
			writeMutex.Lock()
			tar.epochBasedData.validatorStates[virtualEpoch] = lightValis
			// quick update validatorBasedData.validatorIndices
			for _, val := range lightValis.Data {
				tar.validatorBasedData.validatorIndices[string(val.Pubkey)] = val.Index
			}
			writeMutex.Unlock()
			// free up memory
			valis = nil
			return nil
		})
	}
	err := g2.Wait()
	if err != nil {
		return fmt.Errorf("error in epochBasedData: %w", err)
	}
	return nil
}
