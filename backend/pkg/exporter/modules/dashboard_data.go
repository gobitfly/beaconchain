package modules

import (
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

type dashboardData struct {
	ModuleContext
	log           ModuleLog
	signingDomain []byte
	epochWriter   *epochWriter
	epochToHour   *epochToHourAggregator
	hourToDay     *hourToDayAggregator
	dayToWeek     *dayToWeeklyAggregator
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	temp := &dashboardData{
		ModuleContext: moduleContext,
	}
	temp.log = ModuleLog{module: temp}
	temp.epochWriter = newEpochWriter(temp)
	temp.epochToHour = newEpochToHourAggregator(temp)
	temp.hourToDay = newHourToDayAggregator(temp)
	temp.dayToWeek = newDayToWeeklyAggregator(temp)
	return temp
}

func (d *dashboardData) Init() error {
	go func() {
		d.backfillEpochData()
	}()

	return nil
}

var backFillMutex = &sync.Mutex{}

// returns false if there is already an active backfill happening, otherwise returns true
func (d *dashboardData) backfillEpochData() bool {
	if !backFillMutex.TryLock() {
		return false
	}
	defer backFillMutex.Unlock()

	res, err := d.CL.GetFinalityCheckpoints("finalized")
	if err != nil {
		d.log.Error(err, "failed to get finalized checkpoint", 0)
		return true
	}

	gaps, err := edb.GetDashboardEpochGaps(res.Data.Finalized.Epoch, d.epochWriter.getRetentionEpochDuration())
	if err != nil {
		d.log.Error(err, "failed to get epoch gaps", 0)
		return true
	}

	if len(gaps) > 0 {
		d.log.Infof("Epoch dashboard data gaps found, backfilling fom epoch %d to %d", gaps[0], gaps[len(gaps)-1])

		for _, gap := range gaps {
			//backfill if needed, skip backfilling older than RetainEpochDuration/2 since the time it would take to backfill exceeds the retention period anyway
			if gap < res.Data.Finalized.Epoch-d.epochWriter.getRetentionEpochDuration() {
				continue
			}

			// just in case we ask again before exporting since some time may have been passed
			hasEpoch, err := edb.HasDashboardDataForEpoch(gap)
			if err != nil {
				d.log.Error(err, "failed to check if epoch has dashboard data", 0, map[string]interface{}{"epoch": gap})
				continue
			}
			if hasEpoch {
				continue
			}

			d.log.Infof("backfilling epoch %d", gap)
			err = d.ExportEpochData(gap)
			if err != nil {
				d.log.Error(err, "failed to export dashboard epoch data", 0, map[string]interface{}{"epoch": gap})
			}
			d.log.Infof("backfill completed for epoch %d", gap)

			d.aggregate()
		}
	}
	d.log.Infof("backfill finished")
	return true
}

// called every epoch so far
func (d *dashboardData) aggregate() {
	start := time.Now()
	defer func() {
		d.log.Infof("all of aggregation took %v", time.Since(start))
	}()

	// TODO decide what to aggregate how often
	d.epochToHour.aggregate1h()
	d.hourToDay.dayAggregate()
	d.dayToWeek.rolling7dAggregate()
	d.dayToWeek.rolling31dAggregate()
}

func (d *dashboardData) OnFinalizedCheckpoint(_ *constypes.StandardFinalizedCheckpointResponse) error {
	// TODO: dont export head epochs as long as backfill is not finished, makes the code simpler and easier to understand
	// and reduced complex edge cases

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

	d.log.Infof("exporting dashboard epoch data for epoch %d", res.Data.Finalized.Epoch)

	err = d.ExportEpochData(res.Data.Finalized.Epoch)
	if err != nil {
		return err
	}

	// We call backfill here to retry any failed "OnFinalizedCheckpoint" epoch exports
	// since the init backfill only fixes gaps at the start of exporter but does not fix
	// gaps that occur while operating (for example node not available for a brief moment)
	// aggregate if backfill is not running in parallel elsewhere
	if d.backfillEpochData() {
		d.aggregate()
	}

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

	result := d.process(data, d.signingDomain)
	d.log.Infof("processed data for epoch %d in %v", epoch, time.Since(start))

	start = time.Now()
	err := d.epochWriter.writeEpochData(epoch, result)
	if err != nil {
		return err
	}
	d.log.Infof("wrote data for epoch %d in %v", epoch, time.Since(start))

	d.log.Infof("successfully wrote dashboard epoch data for epoch %d in %v", epoch, time.Since(totalStart))
	return nil
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
	lastSlotOfEpoch := firstSlotOfEpoch + slotsPerEpoch - 1

	result.beaconBlockData = make(map[uint64]*constypes.StandardBeaconSlotResponse)
	result.beaconBlockRewardData = make(map[uint64]*constypes.StandardBlockRewardsResponse)
	result.syncCommitteeRewardData = make(map[uint64]*constypes.StandardSyncCommitteeRewardsResponse)
	result.attestationAssignments = make(map[string]uint64)

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(4)

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
		for slot := lastSlotOfEpoch; slot >= lastSlotOfEpoch-utils.Config.Chain.ClConfig.SlotsPerEpoch; slot -= utils.Config.Chain.ClConfig.SlotsPerEpoch {
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

func (d *dashboardData) process(data *Data, domain []byte) []*validatorDashboardDataRow {
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
				validatorsData[i].AttestationsScheduled = 1
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
		validatorsData[proposerIndex].BlockScheduled++
	}

	// write scheduled sync committee data
	for _, validator := range data.syncCommitteeAssignments.Data.Validators {
		validatorsData[mustParseInt64(validator)].SyncScheduled = len(data.beaconBlockData) // take into account missed slots
	}

	// write proposer rewards data
	for _, reward := range data.beaconBlockRewardData {
		validatorsData[reward.Data.ProposerIndex].BlocksClReward += reward.Data.Attestations + reward.Data.AttesterSlashings + reward.Data.ProposerSlashings + reward.Data.SyncAggregate
	}

	// write sync committee reward data & sync committee execution stats
	for _, rewards := range data.syncCommitteeRewardData {
		for _, reward := range rewards.Data {
			validator_index := reward.ValidatorIndex
			syncReward := reward.Reward
			validatorsData[validator_index].SyncReward += syncReward

			if syncReward > 0 {
				validatorsData[validator_index].SyncExecuted++
			}
		}
	}

	// write block specific data
	for _, block := range data.beaconBlockData {
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed++

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
				if validatorsData[pubkeyToIndexMapEnd[depositData.Data.Pubkey]].DepositsCount > 0 {
					d.log.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapStart[depositData.Data.Pubkey]; ok {
					d.log.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					d.log.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
			}

			validator_index := pubkeyToIndexMapEnd[depositData.Data.Pubkey]

			validatorsData[validator_index].DepositsAmount += depositData.Data.Amount
			validatorsData[validator_index].DepositsCount++
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validator_index := withdrawal.ValidatorIndex
			validatorsData[validator_index].WithdrawalsAmount += withdrawal.Amount
			validatorsData[validator_index].WithdrawalsCount++
		}

		for _, attestation := range block.Data.Message.Body.Attestations {
			aggregationBits := bitfield.Bitlist(attestation.AggregationBits)

			for i := uint64(0); i < aggregationBits.Len(); i++ {
				if aggregationBits.BitAt(i) {
					validator_index, found := data.attestationAssignments[utils.FormatAttestorAssignmentKey(attestation.Data.Slot, attestation.Data.Index, i)]
					if !found { // This should never happen!
						d.log.Error(fmt.Errorf("validator not found in attestation assignments"), "validator not found in attestation assignments", 0, map[string]interface{}{"slot": attestation.Data.Slot, "index": attestation.Data.Index, "i": i})
						continue
					}
					validatorsData[validator_index].InclusionDelaySum = int64(block.Data.Message.Slot - attestation.Data.Slot - 1)

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

		validatorsData[validator_index].AttestationsHeadReward = attestationReward.Head
		validatorsData[validator_index].AttestationsSourceReward = attestationReward.Source
		validatorsData[validator_index].AttestationsTargetReward = attestationReward.Target
		validatorsData[validator_index].AttestationsInactivityReward = attestationReward.Inactivity
		validatorsData[validator_index].AttestationsInclusionsReward = attestationReward.InclusionDelay
		validatorsData[validator_index].AttestationReward = validatorsData[validator_index].AttestationsHeadReward +
			validatorsData[validator_index].AttestationsSourceReward +
			validatorsData[validator_index].AttestationsTargetReward +
			validatorsData[validator_index].AttestationsInactivityReward +
			validatorsData[validator_index].AttestationsInclusionsReward
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[int64(data.startBalances.Data[validator_index].Validator.EffectiveBalance)]]
		validatorsData[validator_index].AttestationsIdealHeadReward = idealRewardsOfValidator.Head
		validatorsData[validator_index].AttestationsIdealTargetReward = idealRewardsOfValidator.Target
		validatorsData[validator_index].AttestationsIdealSourceReward = idealRewardsOfValidator.Source
		validatorsData[validator_index].AttestationsIdealInactivityReward = idealRewardsOfValidator.Inactivity
		validatorsData[validator_index].AttestationsIdealInclusionsReward = idealRewardsOfValidator.InclusionDelay

		validatorsData[validator_index].AttestationIdealReward = validatorsData[validator_index].AttestationsIdealHeadReward +
			validatorsData[validator_index].AttestationsIdealSourceReward +
			validatorsData[validator_index].AttestationsIdealTargetReward +
			validatorsData[validator_index].AttestationsIdealInactivityReward +
			validatorsData[validator_index].AttestationsIdealInclusionsReward

		if attestationReward.Head > 0 {
			validatorsData[validator_index].AttestationHeadExecuted = 1
			validatorsData[validator_index].AttestationsExecuted = 1
		}
		if attestationReward.Source > 0 {
			validatorsData[validator_index].AttestationSourceExecuted = 1
			validatorsData[validator_index].AttestationsExecuted = 1
		}
		if attestationReward.Target > 0 {
			validatorsData[validator_index].AttestationTargetExecuted = 1
			validatorsData[validator_index].AttestationsExecuted = 1
		}
	}

	// TODO: el reward data

	return validatorsData
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
	AttestationsSourceReward          int64 //done
	AttestationsTargetReward          int64 //done
	AttestationsHeadReward            int64 //done
	AttestationsInactivityReward      int64 //done
	AttestationsInclusionsReward      int64 //done
	AttestationReward                 int64 //done
	AttestationsIdealSourceReward     int64 //done
	AttestationsIdealTargetReward     int64 //done
	AttestationsIdealHeadReward       int64 //done
	AttestationsIdealInactivityReward int64 //done
	AttestationsIdealInclusionsReward int64 //done
	AttestationIdealReward            int64 //done

	AttestationsScheduled     int8 //done
	AttestationsExecuted      int8 //done
	AttestationHeadExecuted   int8 //done
	AttestationSourceExecuted int8 //done
	AttestationTargetExecuted int8 //done

	BlockScheduled int8    // done
	BlocksProposed int8    // done
	BlockChance    float64 // done

	BlocksClReward int64 // done
	BlocksElReward decimal.Decimal

	SyncScheduled int     // done
	SyncExecuted  int8    // done
	SyncReward    int64   // done
	SyncChance    float64 // done

	Slashed bool // done

	BalanceStart uint64 // done
	BalanceEnd   uint64 // done

	DepositsCount  int8   // done
	DepositsAmount uint64 // done

	WithdrawalsCount  int8   // done
	WithdrawalsAmount uint64 // done

	InclusionDelaySum     int64 // done
	OptimalInclusionDelay int64 // done
}
