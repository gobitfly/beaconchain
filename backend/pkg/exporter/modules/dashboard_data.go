package modules

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/google/uuid"

	//"github.com/fjl/memsize/memsizeui"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	headEpochQueue    chan uint64
	backFillCompleted bool
	phase0HotfixMutex sync.Mutex
	latestSafeEpoch   atomic.Int64
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// This channel is used to queue up epochs from chain head that need to be exported
	temp.headEpochQueue = make(chan uint64, 100)

	// Indicates whether the initial backfill - which is checked when starting the exporter - has completed
	// and the exporter can start listening for new head epochs to be processed
	temp.backFillCompleted = false

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
		blocks      map[uint64]constypes.LightAnySignedBlock // slotOffset => block, if nil = missed. will include blocks for one more epoch than needed because attestations can be included an epoch later
		assignments struct {
			attestationAssignments map[uint64][][]uint64 // slotOffset => committee index => validator index
			blockAssignments       map[uint64]uint64     // slotOffset => validator index
		}
		rewards struct {
			syncCommitteeRewards map[uint64]constypes.StandardSyncCommitteeRewardsResponse // slotOffset => sync committee rewards
			blockRewards         map[uint64]constypes.StandardBlockRewardsResponse         // slotOffset => block reward data
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
	heavyRequestsSemMap := sync.Map{}
	weights := make(map[string]int64)
	weights["heavy"] = 8
	weights["medium"] = 18
	weights["light"] = 128
	orderedKeyList := []string{"heavy", "medium", "light"}
	for k, v := range weights {
		a := semaphore.NewWeighted(v)
		heavyRequestsSemMap.Store(k, a)
	}
	// debug timer that prints the size of the queue for each node every 10 seconds
	// should be stopped once function is done
	timer := time.NewTicker(3 * time.Second)
	defer timer.Stop()
	go func() {
		for {
			_, ok := <-timer.C
			if !ok {
				return
			}
			for _, k := range orderedKeyList {
				heavyRequestsSem, _ := heavyRequestsSemMap.Load(k)
				// read cur, size, len(waiters) using reflection
				v := reflect.ValueOf(heavyRequestsSem)
				cur := v.Elem().FieldByName("cur").Int()
				size := v.Elem().FieldByName("size").Int()
				// waiters is a struct that has a len field
				waiters := v.Elem().FieldByName("waiters").FieldByName("len").Int()
				d.log.Debugf("%s: cur: %d, size: %d, waiters: %d", k, cur, size, waiters)
			}
		}
	}()

	// epoch based Data
	g1.Go(func() error {
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
			slot := uint64(s)
			virtualEpoch := startEpoch + int64(i)
			g2.Go(func() error {
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
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
	})
	// syncPeriodBasedData
	g1.Go(func() error {
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
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
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
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
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
	})

	// blocks
	g1.Go(func() error {
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
			slots = append(slots, uint64(i))
		}
		writeMutex := &sync.Mutex{}
		tar.slotBasedData.blocks = make(map[uint64]constypes.LightAnySignedBlock, len(slots))
		for _, s := range slots {
			slot := uint64(s)
			epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
			g2.Go(func() error {
				d.log.Tracef("fetching block at slot %d", slot)
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("light")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				defer func() {
					//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "slotBasedData", "duration_type": "single"}).Observe(time.Since(start).Seconds())
					metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_single").Observe(time.Since(start).Seconds())
				}()

				block, err := d.CL.GetSlot(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
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
	})
	// block rewards
	g1.Go(func() error {
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
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("light")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				defer func() {
					//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockRewards", "duration_type": "single"}).Observe(time.Since(start).Seconds())
					metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_rewards_single").Observe(time.Since(start).Seconds())
				}()
				data, err := d.CL.GetPropoalRewards(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
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
	})
	// GetSyncRewards
	g1.Go(func() error {
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
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				defer func() {
					metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_sync_rewards_single").Observe(time.Since(start).Seconds())
				}()
				data, err := d.CL.GetSyncRewards(slot)
				if err != nil {
					httpErr := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
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
	})
	// block assignments
	g1.Go(func() error {
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
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("medium")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
				start := time.Now()
				defer func() {
					//metrics.TaskDuration.With(prometheus.Labels{"pkg": "exporter", "module": "dashboard_data", "function": "getDataForEpochRange", "task": "blockAssignments", "duration_type": "single"}).Observe(time.Since(start).Seconds())
					metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_fetch_slot_based_data_block_assignments_single").Observe(time.Since(start).Seconds())
				}()
				data, err := d.CL.GetPropoalAssignments(epoch)
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
	})

	// attestation rewards
	g1.Go(func() error {
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
				// aquiring semaphore
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
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
	})
	// attestation assignments
	g1.Go(func() error {
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
				heavyRequestsSem, _ := heavyRequestsSemMap.Load("heavy")
				err := heavyRequestsSem.(*semaphore.Weighted).Acquire(context.Background(), 1)
				if err != nil {
					return err
				}
				defer heavyRequestsSem.(*semaphore.Weighted).Release(1)
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
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot] = make([][]uint64, committee.Index+1)
					}
					// if not long enough, extend
					if l := len(tar.slotBasedData.assignments.attestationAssignments[committee.Slot]); l < int(committee.Index)+1 {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot] = append(
							tar.slotBasedData.assignments.attestationAssignments[committee.Slot],
							make([][]uint64, int(committee.Index)+1-l)...,
						)
					}
					// if not enough space for validators, allocate
					if l := len(tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index]); l < len(committee.Validators) {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index] = make([]uint64, len(committee.Validators))
					}
					// ass
					for i, valIndex := range committee.Validators {
						tar.slotBasedData.assignments.attestationAssignments[committee.Slot][committee.Index][i] = uint64(valIndex)
					}
				}
				writeMutex.Unlock()
				return nil
			})
		}
		err := g2.Wait()
		if err != nil {
			return fmt.Errorf("error in attestation assignments: %w", err)
		}
		return nil
	})

	err := g1.Wait()
	if err != nil {
		return fmt.Errorf("error in getDataForEpochRange: %w", err)
	}
	return nil
}
