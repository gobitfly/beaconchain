package modules

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

var missedSlots map[uint64]bool
var missedSlotMutex = &sync.RWMutex{}

const heavyAggregateIntervalEpochs = 9 // ~ hourly

type dashboardData struct {
	ModuleContext
	log               ModuleLog
	signingDomain     []byte
	epochWriter       *epochWriter
	epochToTotal      *epochToTotalAggregator
	epochToHour       *epochToHourAggregator
	hourToDay         *hourToDayAggregator
	dayUp             *dayUpAggregator
	headEpochQueue    chan uint64
	backFillCompleted bool
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}
	temp.epochWriter = newEpochWriter(temp)
	temp.epochToTotal = newEpochToTotalAggregator(temp)
	temp.epochToHour = newEpochToHourAggregator(temp)
	temp.hourToDay = newHourToDayAggregator(temp)
	temp.dayUp = newDayUpAggregator(temp)
	temp.headEpochQueue = make(chan uint64, 1000)
	temp.backFillCompleted = false
	return temp
}

func (d *dashboardData) Init() error {
	go func() {
		for {
			done, err := d.backfillEpochData(nil)
			if err != nil {
				d.log.Fatal(err, "failed to backfill epoch data", 0)
			}

			if done {
				d.log.Infof("dashboard data up to date, starting head export")
				d.backFillCompleted = true
				break
			}
			time.Sleep(1 * time.Second)
		}

		d.processHeadQueue()
	}()

	return nil
}

func (d *dashboardData) processHeadQueue() {
	for {
		epoch := <-d.headEpochQueue

		d.log.Infof("exporting dashboard epoch data for epoch %d", epoch)
		stage := 0
		for { // retry this epoch until no errors occur
			currentFinalizedEpoch, err := d.CL.GetFinalityCheckpoints("finalized")
			if err != nil {
				d.log.Error(err, "failed to get finalized checkpoint", 0)
				time.Sleep(time.Second * 10)
				continue
			}

			if stage <= 0 {
				targetEpoch := epoch - 1
				_, err := d.backfillEpochData(&targetEpoch)
				if err != nil {
					d.log.Error(err, "failed to backfill epoch data", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 1
			}

			if stage <= 1 {
				err := d.ExportEpochData(epoch)
				if err != nil {
					d.log.Error(err, "failed to export epoch data", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 2
			}

			if stage <= 2 {
				err := d.aggregatePerEpoch(true)
				if err != nil {
					d.log.Error(err, "failed to aggregate", 0, map[string]interface{}{"epoch": epoch})
					time.Sleep(time.Second * 10)
					continue
				}
				stage = 3
			}

			// only aggregate more intense data if we are close to head finalization
			if currentFinalizedEpoch.Data.Finalized.Epoch < epoch+3 {
				if stage <= 3 {
					err := d.aggregateMid()
					if err != nil {
						d.log.Error(err, "failed to aggregate mid", 0, map[string]interface{}{"epoch": epoch})
						time.Sleep(time.Second * 10)
						continue
					}
					stage = 4
				}

				if epoch%heavyAggregateIntervalEpochs == 0 {
					if stage <= 4 {
						err := d.aggregateHeavy()
						if err != nil {
							d.log.Error(err, "failed to aggregate heavy", 0, map[string]interface{}{"epoch": epoch})
							time.Sleep(time.Second * 10)
							continue
						}
						stage = 5
					}
				}
			}

			break
		}
	}
}

type DataEpoch struct {
	Data  []*validatorDashboardDataRow
	Epoch uint64
}

// returns true if there was nothing to backfill, otherwise returns false
// if upToEpoch is nil, it will backfill until the latest finalized epoch
func (d *dashboardData) backfillEpochData(upToEpoch *uint64) (bool, error) {
	if upToEpoch == nil {
		res, err := d.CL.GetFinalityCheckpoints("finalized")
		if err != nil {
			return false, errors.Wrap(err, "failed to get finalized checkpoint")
		}
		upToEpoch = &res.Data.Finalized.Epoch
	}

	latestExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return false, errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	gaps, err := edb.GetDashboardEpochGaps(*upToEpoch, *upToEpoch-latestEpoch)
	if err != nil {
		return false, errors.Wrap(err, "failed to get epoch gaps")
	}

	if len(gaps) > 0 {
		d.log.Infof("Epoch dashboard data gaps found, backfilling gaps in the range fom epoch %d to %d", gaps[0], gaps[len(gaps)-1])

		var nextDataChan chan *DataEpoch = make(chan *DataEpoch, 1)
		go func() {
			for _, gap := range gaps {
				for {
					//backfill if needed, skip backfilling older than RetainEpochDuration
					if gap < latestExportedEpoch {
						continue
					}

					// just in case we ask again before exporting since some time may have been passed
					hasEpoch, err := edb.HasDashboardDataForEpoch(gap)
					if err != nil {
						d.log.Error(err, "failed to check if epoch has dashboard data", 0, map[string]interface{}{"epoch": gap})
						time.Sleep(time.Second * 10)
						continue
					}
					if hasEpoch {
						time.Sleep(time.Second * 1)
						continue
					}

					d.log.Infof("backfill, retreiving data for epoch %d", gap)
					data, err := d.GetEpochData(gap)
					if err != nil {
						d.log.Error(err, "failed to get epoch data", 0, map[string]interface{}{"epoch": gap})
						time.Sleep(time.Second * 10)
						continue
					}

					nextDataChan <- &DataEpoch{
						Data:  data,
						Epoch: gap,
					}

					break
				}
			}
		}()

		for {
			select {
			case data := <-nextDataChan:
				stage := 0
				for { // retry this epoch until no errors occur
					if stage <= 0 {
						err := d.epochWriter.WriteEpochData(data.Epoch, data.Data)
						if err != nil {
							d.log.Error(err, "backfill, failed to write epoch data", 0, map[string]interface{}{"epoch": data.Epoch})
							time.Sleep(time.Second * 10)
							continue
						}
						stage = 1
						d.log.Infof("backfill, wrote epoch data %d", data.Epoch)
					}

					err = d.aggregatePerEpoch(false)
					if err != nil {
						d.log.Error(err, "backfill, failed to aggregate", 0, map[string]interface{}{"epoch": data.Epoch})
						time.Sleep(time.Second * 10)
						continue
					}
					d.log.Infof("backfill, aggregated epoch data %d", data.Epoch)

					if data.Epoch >= gaps[len(gaps)-1] {
						d.log.Infof("backfilling finished for range epoch %d to %d", gaps[0], gaps[len(gaps)-1])
						return false, nil
					}

					break
				}
			case <-time.After(time.Second * 10):
			}
		}
	}
	return true, nil
}

var lastExportedHour uint64 = 0

func (d *dashboardData) aggregatePerEpoch(workingOnHead bool) error {
	start := time.Now()
	defer func() {
		d.log.Infof("all of aggregation took %v", time.Since(start))
	}()

	// important to do this before hour aggregate as hour aggregate deletes old epochs
	err := d.epochToTotal.aggregateTotal()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate total")
	}

	currentExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get last exported epoch")
	}
	currentStartBound, _ := d.epochToHour.getHourAggregateBounds(currentExportedEpoch)

	// Performance improvement for backfilling, no need to aggregate day after each epoch, we can update once per hour

	if workingOnHead || currentStartBound != lastExportedHour {
		err = d.epochToHour.aggregate1hAndClearOld() // will aggregate last hour too if it hasnt completed yet
		if err != nil {
			return errors.Wrap(err, "failed to aggregate 1h")
		}

		err := d.hourToDay.dayAggregateAndClearOld(workingOnHead)
		if err != nil {
			return errors.Wrap(err, "failed to aggregate day")
		}
		lastExportedHour = currentStartBound
	}

	return nil
}

func (d *dashboardData) aggregateMid() error {
	start := time.Now()
	defer func() {
		d.log.Infof("all of aggregation (mid) took %v", time.Since(start))
	}()

	err := d.dayUp.rolling7dAggregate()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate 7d")
	}

	err = d.dayUp.rolling30dAggregate()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate 30d")
	}

	err = d.dayUp.rolling90dAggregate()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate 90d")
	}

	return nil
}

// should not be called frequent due to the high amount of data
func (d *dashboardData) aggregateHeavy() error {
	start := time.Now()
	defer func() {
		d.log.Infof("all of aggregation (heavy) took %v", time.Since(start))
	}()

	err := d.dayUp.rolling365dAggregate()
	if err != nil {
		return errors.Wrap(err, "failed to aggregate 365d")
	}

	return nil
}

func (d *dashboardData) OnFinalizedCheckpoint(_ *constypes.StandardFinalizedCheckpointResponse) error {
	if !d.backFillCompleted {
		return nil // nothing to do, wait for backfill to finish
	}

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Hence we just listen for new justified epochs here and fetch the latest finalized one from the node
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	res, err := d.CL.GetFinalityCheckpoints("finalized")
	if err != nil {
		return err
	}

	latestExported, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return err
	}

	if latestExported != 0 {
		if res.Data.Finalized.Epoch <= latestExported {
			d.log.Infof("dashboard epoch data already exported for epoch %d", res.Data.Finalized.Epoch)
			return nil
		}
	}

	d.headEpochQueue <- res.Data.Finalized.Epoch

	return nil
}

func (d *dashboardData) GetName() string {
	return "Dashboard-Data"
}

func (d *dashboardData) OnHead(event *constypes.StandardEventHeadResponse) error {
	return nil
}

func (d *dashboardData) OnChainReorg(event *constypes.StandardEventChainReorg) error {
	return nil
}

func (d *dashboardData) ExportEpochDataNonSequential(epoch uint64) error {
	missedSlotMutex.Lock()
	missedSlots = nil
	missedSlotMutex.Unlock()
	return d.ExportEpochData(epoch)
}

/*
Use this function when sequentially getting data of new epochs, for example in the exporter.
For random non sequential epoch data use ExportEpochDataNonSequential
*/
func (d *dashboardData) ExportEpochData(epoch uint64) error {
	totalStart := time.Now()
	data := d.getData(epoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if data == nil {
		return errors.New("can not get data")
	}
	d.log.Infof("retrieved data for epoch %d in %v", epoch, time.Since(totalStart))

	start := time.Now()
	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return err
		}
		d.signingDomain = domain
	}

	result, err := d.process(data, d.signingDomain)
	if err != nil {
		return err
	}
	d.log.Infof("processed data for epoch %d in %v", epoch, time.Since(start))

	start = time.Now()
	err = d.epochWriter.WriteEpochData(epoch, result)
	if err != nil {
		return err
	}
	d.log.Infof("wrote data for epoch %d in %v", epoch, time.Since(start))

	d.log.Infof("successfully wrote dashboard epoch data for epoch %d in %v", epoch, time.Since(totalStart))
	return nil
}

func (d *dashboardData) GetEpochData(epoch uint64) ([]*validatorDashboardDataRow, error) {
	data := d.getData(epoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if data == nil {
		return nil, errors.New("can not get data")
	}

	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return nil, err
		}
		d.signingDomain = domain
	}

	return d.process(data, d.signingDomain)
}

type Data struct {
	startBalances            *constypes.StandardValidatorsResponse
	endBalances              *constypes.StandardValidatorsResponse
	proposerAssignments      *constypes.StandardProposerAssignmentsResponse
	syncCommitteeAssignments *constypes.StandardSyncCommitteesResponse
	attestationRewards       *constypes.StandardAttestationRewardsResponse
	beaconBlockData          map[uint64]*constypes.StandardBeaconSlotResponse
	beaconBlockRewardData    map[uint64]*constypes.StandardBlockRewardsResponse
	syncCommitteeRewardData  map[uint64]*constypes.StandardSyncCommitteeRewardsResponse
	attestationAssignments   map[string]uint64
}

func (d *dashboardData) getData(epoch, slotsPerEpoch uint64) *Data {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * slotsPerEpoch

	firstSlotOfPreviousEpoch := firstSlotOfEpoch - 1
	if firstSlotOfEpoch == 0 {
		firstSlotOfPreviousEpoch = 0
	}
	lastSlotOfEpoch := firstSlotOfEpoch + slotsPerEpoch - 1

	result.beaconBlockData = make(map[uint64]*constypes.StandardBeaconSlotResponse)
	result.beaconBlockRewardData = make(map[uint64]*constypes.StandardBlockRewardsResponse)
	result.syncCommitteeRewardData = make(map[uint64]*constypes.StandardSyncCommitteeRewardsResponse)
	result.attestationAssignments = make(map[string]uint64)

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(4)

	totalStart := time.Now()
	errGroup.Go(func() error {
		// retrieve the validator balances at the start of the epoch
		start := time.Now()
		result.startBalances, err = d.CL.GetValidators(firstSlotOfPreviousEpoch, nil, nil)
		if err != nil {
			d.log.Error(err, "can not get validators balances", 0, map[string]interface{}{"firstSlotOfPreviousEpoch": firstSlotOfPreviousEpoch})
			return err
		}
		d.log.Infof("retrieved start balances using state at slot %d in %v", firstSlotOfPreviousEpoch, time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve proposer assignments for the epoch in order to attribute missed slots
		start := time.Now()
		result.proposerAssignments, err = d.CL.GetPropoalAssignments(epoch)
		if err != nil {
			d.log.Error(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
			return err
		}
		d.log.Infof("retrieved proposer assignments in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
		start := time.Now()
		result.syncCommitteeAssignments, err = d.CL.GetSyncCommitteesAssignments(epoch, int64(firstSlotOfEpoch))
		if err != nil {
			d.log.Error(err, "can not get sync committee assignments", 0, map[string]interface{}{"epoch": epoch})
			return nil
		}
		d.log.Infof("retrieved sync committee assignments in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve the attestation committees
		start := time.Now()
		// As of dencun you can attest up until the end of the following epoch
		min := lastSlotOfEpoch - utils.Config.Chain.ClConfig.SlotsPerEpoch
		if lastSlotOfEpoch < utils.Config.Chain.ClConfig.SlotsPerEpoch {
			min = 0
		}

		for slot := lastSlotOfEpoch; slot >= min; slot -= utils.Config.Chain.ClConfig.SlotsPerEpoch {
			data, err := d.CL.GetCommittees(slot, nil, nil, nil)
			if err != nil {
				d.log.Error(err, "can not get attestation assignments", 0, map[string]interface{}{"slot": slot})
				return nil
			}

			for _, committee := range data.Data {
				for i, valIndex := range committee.Validators {
					valIndexU64, err := strconv.ParseUint(valIndex, 10, 64)
					if err != nil {
						d.log.Error(err, "can not parse validator index", 0, map[string]interface{}{"slot": committee.Slot, "committee": committee.Index, "index": i})
						continue
					}
					k := utils.FormatAttestorAssignmentKey(committee.Slot, committee.Index, uint64(i))
					result.attestationAssignments[k] = valIndexU64
				}
			}

			if slot < utils.Config.Chain.ClConfig.SlotsPerEpoch {
				break // special case for epoch 0
			}
		}
		d.log.Infof("retrieved attestation assignments in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// attestation rewards
		start := time.Now()
		result.attestationRewards, err = d.CL.GetAttestationRewards(epoch)
		if err != nil {
			d.log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
			return nil
		}
		d.log.Infof("retrieved attestation rewards data in %v", time.Since(start))
		return nil
	})

	errGroup.Go(func() error {
		// retrieve the validator balances at the end of the epoch
		start := time.Now()
		result.endBalances, err = d.CL.GetValidators(lastSlotOfEpoch, nil, nil)
		if err != nil {
			d.log.Error(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfEpoch": lastSlotOfEpoch})
			return nil
		}
		d.log.Infof("retrieved end balances using state at slot %d in %v", lastSlotOfEpoch, time.Since(start))
		return nil
	})

	err = errGroup.Wait()
	if err != nil {
		return nil
	}

	d.log.Infof("retrieved all data for epoch %d in %v", epoch, time.Since(totalStart))

	missedSlotMutex.Lock()
	if missedSlots == nil {
		missedSlots = make(map[uint64]bool)
		if firstSlotOfEpoch > slotsPerEpoch { // handle case for first epoch
			// get missed slots of last epoch for optimal inclusion distance
			for slot := firstSlotOfEpoch - slotsPerEpoch; slot <= lastSlotOfEpoch-slotsPerEpoch; slot++ {
				_, err := d.CL.GetSlot(slot)
				if err != nil {
					httpErr, _ := network.SpecificError(err)
					if httpErr != nil && httpErr.StatusCode == 404 {
						missedSlots[slot] = true
						continue // missed
					}
					d.log.Fatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
					continue
				}
			}
		}
	}
	missedSlotMutex.Unlock()

	// retrieve the data for all blocks that were proposed in this epoch
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		//d.log.Infof("retrieving data for block at slot %d", slot)
		block, err := d.CL.GetSlot(slot)
		if err != nil {
			httpErr, _ := network.SpecificError(err)
			if httpErr != nil && httpErr.StatusCode == 404 {
				missedSlotMutex.Lock()
				missedSlots[slot] = true
				missedSlotMutex.Unlock()
				continue // missed
			}
			d.log.Fatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		if block.Data.Message.StateRoot == "" {
			// todo better network handling, if 404 just log info, else log error
			d.log.Error(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockData[slot] = block

		blockReward, err := d.CL.GetPropoalRewards(slot)
		if err != nil {
			d.log.Error(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockRewardData[slot] = blockReward

		syncRewards, err := d.CL.GetSyncRewards(slot)
		if err != nil {
			d.log.Error(err, "can not get sync committee reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.syncCommitteeRewardData[slot] = syncRewards
	}

	// clean up old missed slots
	missedSlotMutex.Lock()
	if len(missedSlots) > 0 {
		newMissedSlots := make(map[uint64]bool, 0)
		for slot := range missedSlots {
			if slot < firstSlotOfEpoch-slotsPerEpoch {
				continue
			}
			newMissedSlots[slot] = true
		}
		missedSlots = newMissedSlots
	}
	missedSlotMutex.Unlock()

	return &result
}

func (d *dashboardData) process(data *Data, domain []byte) ([]*validatorDashboardDataRow, error) {
	validatorsData := make([]*validatorDashboardDataRow, len(data.endBalances.Data))

	idealAttestationRewards := make(map[int64]int)
	for i, idealReward := range data.attestationRewards.Data.IdealRewards {
		idealAttestationRewards[idealReward.EffectiveBalance] = i
	}

	pubkeyToIndexMapEnd := make(map[string]int64)
	pubkeyToIndexMapStart := make(map[string]int64)
	activeCount := 0
	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(data.startBalances.Data) {
			validatorsData[i].BalanceStart = data.startBalances.Data[i].Balance
			pubkeyToIndexMapStart[data.startBalances.Data[i].Validator.Pubkey] = int64(i)

			if data.startBalances.Data[i].Status.IsActive() {
				activeCount++
				validatorsData[i].AttestationsScheduled = sql.NullInt16{Int16: 1, Valid: true}
			}
		}
		validatorsData[i].BalanceEnd = data.endBalances.Data[i].Balance
		validatorsData[i].Slashed = data.endBalances.Data[i].Validator.Slashed

		pubkeyToIndexMapEnd[data.endBalances.Data[i].Validator.Pubkey] = int64(i)
	}

	// slotsPerSyncCommittee :=  * float64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
	for validator_index := range validatorsData {
		validatorsData[validator_index].SyncChance = float64(utils.Config.Chain.ClConfig.SyncCommitteeSize) / float64(activeCount) / float64(utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod)
		validatorsData[validator_index].BlockChance = float64(utils.Config.Chain.ClConfig.SlotsPerEpoch) / float64(activeCount)
	}

	// write scheduled block data
	for _, proposerAssignment := range data.proposerAssignments.Data {
		proposerIndex := proposerAssignment.ValidatorIndex
		validatorsData[proposerIndex].BlockScheduled.Int16++
		validatorsData[proposerIndex].BlockScheduled.Valid = true
	}

	// write scheduled sync committee data
	for _, validator := range data.syncCommitteeAssignments.Data.Validators {
		validatorsData[mustParseInt64(validator)].SyncScheduled.Int32 = int32(len(data.beaconBlockData)) // take into account missed slots
		validatorsData[mustParseInt64(validator)].SyncScheduled.Valid = true
	}

	// write proposer rewards data
	for _, reward := range data.beaconBlockRewardData {
		validatorsData[reward.Data.ProposerIndex].BlocksClReward.Int64 += reward.Data.Attestations + reward.Data.AttesterSlashings + reward.Data.ProposerSlashings + reward.Data.SyncAggregate
		validatorsData[reward.Data.ProposerIndex].BlocksClReward.Valid = true
	}

	// write sync committee reward data & sync committee execution stats
	for _, rewards := range data.syncCommitteeRewardData {
		for _, reward := range rewards.Data {
			validator_index := reward.ValidatorIndex
			syncReward := reward.Reward
			validatorsData[validator_index].SyncReward.Int64 += syncReward
			validatorsData[validator_index].SyncReward.Valid = true

			if syncReward > 0 {
				validatorsData[validator_index].SyncExecuted.Int32++
				validatorsData[validator_index].SyncExecuted.Valid = true
			}
		}
	}

	// write block specific data
	for _, block := range data.beaconBlockData {
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Int16++
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed.Valid = true

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {
			// TODO: properly verify that deposit is valid:
			// if signature is valid I count the deposit towards the balance
			// if signature is invalid and the validator was in the state at the beginning of the epoch I count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there were no valid deposits in the block prior I DO NOT count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there was a VALID deposit in the blocks prior I DO COUNT the deposit towards the balance

			err := utils.VerifyDepositSignature(&phase0.DepositData{
				PublicKey:             phase0.BLSPubKey(utils.MustParseHex(depositData.Data.Pubkey)),
				WithdrawalCredentials: depositData.Data.WithdrawalCredentials,
				Amount:                phase0.Gwei(depositData.Data.Amount),
				Signature:             phase0.BLSSignature(depositData.Data.Signature),
			}, domain)

			if err != nil {
				d.log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, block.Data.Message.Slot, err, depositData.Data.Signature), "", 0)

				// if the validator hat a valid deposit prior to the current one, count the invalid towards the balance
				if validatorsData[pubkeyToIndexMapEnd[depositData.Data.Pubkey]].DepositsCount.Int16 > 0 {
					d.log.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapStart[depositData.Data.Pubkey]; ok {
					d.log.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					d.log.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
				return nil, err
			}

			validator_index := pubkeyToIndexMapEnd[depositData.Data.Pubkey]

			validatorsData[validator_index].DepositsAmount.Int64 += int64(depositData.Data.Amount)
			validatorsData[validator_index].DepositsAmount.Valid = true

			validatorsData[validator_index].DepositsCount.Int16++
			validatorsData[validator_index].DepositsCount.Valid = true
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validator_index := withdrawal.ValidatorIndex
			validatorsData[validator_index].WithdrawalsAmount.Int64 += int64(withdrawal.Amount)
			validatorsData[validator_index].WithdrawalsAmount.Valid = true

			validatorsData[validator_index].WithdrawalsCount.Int16++
			validatorsData[validator_index].WithdrawalsCount.Valid = true
		}

		for _, attestation := range block.Data.Message.Body.Attestations {
			aggregationBits := bitfield.Bitlist(attestation.AggregationBits)

			for i := uint64(0); i < aggregationBits.Len(); i++ {
				if aggregationBits.BitAt(i) {
					validator_index, found := data.attestationAssignments[utils.FormatAttestorAssignmentKey(attestation.Data.Slot, attestation.Data.Index, i)]
					if !found { // This should never happen!
						d.log.Error(fmt.Errorf("validator not found in attestation assignments"), "validator not found in attestation assignments", 0, map[string]interface{}{"slot": attestation.Data.Slot, "index": attestation.Data.Index, "i": i})
						return nil, fmt.Errorf("validator not found in attestation assignments")
					}
					validatorsData[validator_index].InclusionDelaySum = sql.NullInt64{
						Int64: int64(block.Data.Message.Slot - attestation.Data.Slot - 1),
						Valid: true,
					}

					optimalInclusionDistance := 0
					missedSlotMutex.RLock()
					for i := attestation.Data.Slot + 1; i < block.Data.Message.Slot; i++ {
						if _, ok := missedSlots[i]; ok {
							optimalInclusionDistance++
						} else {
							break
						}
					}
					missedSlotMutex.RUnlock()

					validatorsData[validator_index].OptimalInclusionDelay = int64(optimalInclusionDistance)
				}
			}
		}
	}

	// write attestation rewards data
	for _, attestationReward := range data.attestationRewards.Data.TotalRewards {
		validator_index := attestationReward.ValidatorIndex

		validatorsData[validator_index].AttestationsHeadReward = sql.NullInt64{Int64: attestationReward.Head, Valid: true}
		validatorsData[validator_index].AttestationsSourceReward = sql.NullInt64{Int64: attestationReward.Source, Valid: true}
		validatorsData[validator_index].AttestationsTargetReward = sql.NullInt64{Int64: attestationReward.Target, Valid: true}
		validatorsData[validator_index].AttestationsInactivityReward = sql.NullInt64{Int64: attestationReward.Inactivity, Valid: true}
		validatorsData[validator_index].AttestationsInclusionsReward = sql.NullInt64{Int64: attestationReward.InclusionDelay, Valid: true}
		validatorsData[validator_index].AttestationReward = sql.NullInt64{
			Int64: attestationReward.Head + attestationReward.Source + attestationReward.Target + attestationReward.Inactivity + attestationReward.InclusionDelay,
			Valid: true,
		}
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[int64(data.startBalances.Data[validator_index].Validator.EffectiveBalance)]]
		validatorsData[validator_index].AttestationsIdealHeadReward = sql.NullInt64{Int64: idealRewardsOfValidator.Head, Valid: true}
		validatorsData[validator_index].AttestationsIdealTargetReward = sql.NullInt64{Int64: idealRewardsOfValidator.Target, Valid: true}
		validatorsData[validator_index].AttestationsIdealSourceReward = sql.NullInt64{Int64: idealRewardsOfValidator.Source, Valid: true}
		validatorsData[validator_index].AttestationsIdealInactivityReward = sql.NullInt64{Int64: idealRewardsOfValidator.Inactivity, Valid: true}
		validatorsData[validator_index].AttestationsIdealInclusionsReward = sql.NullInt64{Int64: idealRewardsOfValidator.InclusionDelay, Valid: true}

		validatorsData[validator_index].AttestationIdealReward = sql.NullInt64{
			Int64: idealRewardsOfValidator.Head + idealRewardsOfValidator.Source + idealRewardsOfValidator.Target + idealRewardsOfValidator.Inactivity + idealRewardsOfValidator.InclusionDelay,
			Valid: true,
		}

		if attestationReward.Head > 0 {
			validatorsData[validator_index].AttestationHeadExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
		if attestationReward.Source > 0 {
			validatorsData[validator_index].AttestationSourceExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
		if attestationReward.Target > 0 {
			validatorsData[validator_index].AttestationTargetExecuted = sql.NullInt16{Int16: 1, Valid: true}
			validatorsData[validator_index].AttestationsExecuted = sql.NullInt16{Int16: 1, Valid: true}
		}
	}

	// TODO: el reward data

	return validatorsData, nil
}

func mustParseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	r, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		panic(err)
	}
	return r
}

type validatorDashboardDataRow struct {
	AttestationsSourceReward          sql.NullInt64 //done
	AttestationsTargetReward          sql.NullInt64 //done
	AttestationsHeadReward            sql.NullInt64 //done
	AttestationsInactivityReward      sql.NullInt64 //done
	AttestationsInclusionsReward      sql.NullInt64 //done
	AttestationReward                 sql.NullInt64 //done
	AttestationsIdealSourceReward     sql.NullInt64 //done
	AttestationsIdealTargetReward     sql.NullInt64 //done
	AttestationsIdealHeadReward       sql.NullInt64 //done
	AttestationsIdealInactivityReward sql.NullInt64 //done
	AttestationsIdealInclusionsReward sql.NullInt64 //done
	AttestationIdealReward            sql.NullInt64 //done

	AttestationsScheduled     sql.NullInt16 //done
	AttestationsExecuted      sql.NullInt16 //done
	AttestationHeadExecuted   sql.NullInt16 //done
	AttestationSourceExecuted sql.NullInt16 //done
	AttestationTargetExecuted sql.NullInt16 //done

	BlockScheduled sql.NullInt16 // done
	BlocksProposed sql.NullInt16 // done
	BlockChance    float64       // done

	BlocksClReward sql.NullInt64 // done
	BlocksElReward decimal.Decimal

	SyncScheduled sql.NullInt32 // done
	SyncExecuted  sql.NullInt32 // done
	SyncReward    sql.NullInt64 // done
	SyncChance    float64       // done

	Slashed bool // done

	BalanceStart uint64 // done
	BalanceEnd   uint64 // done

	DepositsCount  sql.NullInt16 // done
	DepositsAmount sql.NullInt64 // done

	WithdrawalsCount  sql.NullInt16 // done
	WithdrawalsAmount sql.NullInt64 // done

	InclusionDelaySum     sql.NullInt64 // done
	OptimalInclusionDelay int64         // done
}
