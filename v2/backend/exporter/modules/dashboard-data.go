package modules

import (
	"fmt"
	"strconv"

	"github.com/gobitfly/beaconchain/commons"
	"github.com/gobitfly/beaconchain/commons/utils"
	"github.com/gobitfly/beaconchain/exporter/clnode"
	"github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
)

type dashboardData struct {
	ModuleContext
	Epoch int // todo remove
}

func NewDashboardDataModule(moduleContext ModuleContext, epoch int) ModuleInterface {
	return &dashboardData{
		ModuleContext: moduleContext,
		Epoch:         epoch,
	}
}

func (d *dashboardData) Start() {
	spec, err := d.CL.GetSpec()
	if err != nil {
		commons.LogFatal(err, "can not get spec", 0)
		return
	}

	data := d.getData(d.Epoch, int(spec.Data.SlotsPerEpoch))
	if data == nil {
		return
	}
	result := process(data, utils.MustParseHex(spec.Data.GenesisForkVersion))

	// todo store in db

	fmt.Printf("done %v", result)
}

type Data struct {
	startBalances            clnode.GetValidatorsResponse
	endBalances              clnode.GetValidatorsResponse
	proposerAssignments      clnode.GetProposerAssignmentsResponse
	syncCommitteeAssignments clnode.GetSyncCommitteeAssignmentsResponse
	attestationRewards       clnode.GetAttestationRewardsResponse
	beaconBlockData          map[int]*clnode.GetBeaconSlotResponse
	beaconBlockRewardData    map[int]*clnode.GetBlockRewardsResponse
	syncCommitteeRewardData  map[int]*clnode.GetSyncCommitteeRewardsResponse
}

func (d *dashboardData) getData(epoch, slotsPerEpoch int) *Data {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * int(slotsPerEpoch)
	firstSlotOfPreviousEpoch := firstSlotOfEpoch - 1
	lastSlotOfEpoch := firstSlotOfEpoch + int(slotsPerEpoch)

	result.beaconBlockData = make(map[int]*clnode.GetBeaconSlotResponse)
	result.beaconBlockRewardData = make(map[int]*clnode.GetBlockRewardsResponse)
	result.syncCommitteeRewardData = make(map[int]*clnode.GetSyncCommitteeRewardsResponse)

	// retrieve the validator balances at the start of the epoch
	logrus.Infof("retrieving start balances using state at slot %d", firstSlotOfPreviousEpoch)
	result.startBalances, err = d.CL.GetValidators(firstSlotOfPreviousEpoch)
	if err != nil {
		commons.LogError(err, "can not get validators balances", 0, map[string]interface{}{"firstSlotOfPreviousEpoch": firstSlotOfPreviousEpoch})
		return nil
	}

	// retrieve proposer assignments for the epoch in order to attribute missed slots
	logrus.Infof("retrieving proposer assignments")
	result.proposerAssignments, err = d.CL.GetPropoalAssignments(epoch)
	if err != nil {
		commons.LogError(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
	logrus.Infof("retrieving sync committee assignments")
	result.syncCommitteeAssignments, err = d.CL.GetSyncCommitteesAssignments(epoch, int64(firstSlotOfEpoch))
	if err != nil {
		commons.LogError(err, "can not get sync committee assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// attestation rewards
	logrus.Infof("retrieving attestation rewards data")
	result.attestationRewards, err = d.CL.GetAttestationRewards(epoch)
	if err != nil {
		commons.LogError(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve the data for all blocks that were proposed in this epoch
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		logrus.Infof("retrieving data for block at slot %d", slot)
		block, err := d.CL.GetSlot(slot)
		if err != nil {
			commons.LogFatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		if block.Data.Message.Slot == "" {
			// todo better network handling, if 404 just log info, else log error
			commons.LogError(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockData[slot] = &block

		blockReward, err := d.CL.GetPropoalRewards(slot)
		if err != nil {
			commons.LogError(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockRewardData[slot] = &blockReward

		syncRewards, err := d.CL.GetSyncRewards(slot)
		if err != nil {
			commons.LogError(err, "can not get sync committee reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.syncCommitteeRewardData[slot] = &syncRewards
	}

	// retrieve the validator balances at the end of the epoch
	logrus.Infof("retrieving end balances using state at slot %d", lastSlotOfEpoch)
	result.endBalances, err = d.CL.GetValidators(lastSlotOfEpoch)
	if err != nil {
		commons.LogError(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfEpoch": lastSlotOfEpoch})
		return nil
	}

	return &result
}

func process(data *Data, domain []byte) []*validatorDashboardDataRow {
	validatorsData := make([]*validatorDashboardDataRow, len(data.endBalances.Data))

	idealAttestationRewards := make(map[string]int)
	for i, idealReward := range data.attestationRewards.Data.IdealRewards {
		idealAttestationRewards[idealReward.EffectiveBalance] = i
	}

	pubkeyToIndexMapEnd := make(map[string]int64)
	pubkeyToIndexMapStart := make(map[string]int64)
	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(data.startBalances.Data) {
			validatorsData[i].BalanceStart = mustParseInt64(data.startBalances.Data[i].Balance)
			pubkeyToIndexMapStart[data.startBalances.Data[i].Validator.Pubkey] = int64(i)
		}
		validatorsData[i].BalanceEnd = mustParseInt64(data.endBalances.Data[i].Balance)
		validatorsData[i].Slashed = data.endBalances.Data[i].Validator.Slashed

		pubkeyToIndexMapEnd[data.endBalances.Data[i].Validator.Pubkey] = int64(i)
	}

	// write scheduled block data
	for _, proposerAssignment := range data.proposerAssignments.Data {
		proposerIndex := mustParseInt(proposerAssignment.ValidatorIndex)
		validatorsData[proposerIndex].BlockScheduled++
	}

	// write scheduled sync committee data
	for _, validator := range data.syncCommitteeAssignments.Data.Validators {
		validatorsData[mustParseInt64(validator)].SyncScheduled = len(data.beaconBlockData) // take into account missed slots
	}

	// write proposer rewards data
	for _, reward := range data.beaconBlockRewardData {
		validatorsData[mustParseInt(reward.Data.ProposerIndex)].BlocksClReward += mustParseInt64(reward.Data.Attestations) + mustParseInt64(reward.Data.AttesterSlashings) + mustParseInt64(reward.Data.ProposerSlashings) + mustParseInt64(reward.Data.SyncAggregate)
	}

	// write sync committee reward data & sync committee execution stats
	for _, rewards := range data.syncCommitteeRewardData {
		for _, reward := range rewards.Data {
			validatorIndex := mustParseInt(reward.ValidatorIndex)
			syncReward := mustParseInt64(reward.Reward)
			validatorsData[validatorIndex].SyncReward += syncReward

			if syncReward > 0 {
				validatorsData[validatorIndex].SyncExecuted++
			}
		}
	}

	// write block specific data
	for _, block := range data.beaconBlockData {
		validatorsData[mustParseInt64(block.Data.Message.ProposerIndex)].BlocksProposed++

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {

			// TODO: properly verify that deposit is valid:
			// if signature is valid I count the the deposit towards the balance
			// if signature is invalid and the validator was in the state at the beginning of the epoch I count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there were no valid deposits in the block prior I DO NOT count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there was a VALID deposit in the blocks prior I DO COUNT the deposit towards the balance
			err := deposit.VerifyDepositSignature(&ethpb.Deposit_Data{
				PublicKey:             utils.MustParseHex(depositData.Data.Pubkey),
				WithdrawalCredentials: utils.MustParseHex(depositData.Data.WithdrawalCredentials),
				Amount:                uint64(mustParseInt64(depositData.Data.Amount)),
				Signature:             utils.MustParseHex(depositData.Data.Signature),
			}, domain)

			if err != nil {
				logrus.Errorf("deposit at index %d in slot %s is invalid: %v (signature: %s)", depositIndex, block.Data.Message.Slot, err, depositData.Data.Signature)

				// if the validator hat a valid deposit prior to the current one, count the invalid towards the balance
				if validatorsData[pubkeyToIndexMapEnd[depositData.Data.Pubkey]].DepositsCount > 0 {
					logrus.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapStart[depositData.Data.Pubkey]; ok {
					logrus.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					logrus.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
			}

			validatorIndex := pubkeyToIndexMapEnd[depositData.Data.Pubkey]

			validatorsData[validatorIndex].DepositsAmount += mustParseInt64(depositData.Data.Amount)
			validatorsData[validatorIndex].DepositsCount++
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validatorIndex := mustParseInt64(withdrawal.ValidatorIndex)
			validatorsData[validatorIndex].WithdrawalsAmount += mustParseInt64(withdrawal.Amount)
			validatorsData[validatorIndex].WithdrawalsCount++
		}
	}

	// write attestation rewards data
	for _, attestationReward := range data.attestationRewards.Data.TotalRewards {

		validatorIndex := mustParseInt(attestationReward.ValidatorIndex)

		validatorsData[validatorIndex].AttestationsHeadReward = mustParseInt64(attestationReward.Head)
		validatorsData[validatorIndex].AttestationsSourceReward = mustParseInt64(attestationReward.Source)
		validatorsData[validatorIndex].AttestationsTargetReward = mustParseInt64(attestationReward.Target)
		validatorsData[validatorIndex].AttestationsInactivityReward = mustParseInt64(attestationReward.Inactivity)
		validatorsData[validatorIndex].AttestationsInclusionsReward = mustParseInt64(attestationReward.InclusionDelay)
		validatorsData[validatorIndex].AttestationReward = validatorsData[validatorIndex].AttestationsHeadReward +
			validatorsData[validatorIndex].AttestationsSourceReward +
			validatorsData[validatorIndex].AttestationsTargetReward +
			validatorsData[validatorIndex].AttestationsInactivityReward +
			validatorsData[validatorIndex].AttestationsInclusionsReward
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[data.startBalances.Data[validatorIndex].Validator.EffectiveBalance]]
		validatorsData[validatorIndex].AttestationsIdealHeadReward = mustParseInt64(idealRewardsOfValidator.Head)
		validatorsData[validatorIndex].AttestationsIdealTargetReward = mustParseInt64(idealRewardsOfValidator.Target)
		validatorsData[validatorIndex].AttestationsIdealHeadReward = mustParseInt64(idealRewardsOfValidator.Head)
		validatorsData[validatorIndex].AttestationsIdealInactivityReward = mustParseInt64(idealRewardsOfValidator.Inactivity)
		validatorsData[validatorIndex].AttestationsIdealInclusionsReward = mustParseInt64(idealRewardsOfValidator.InclusionDelay)

		validatorsData[validatorIndex].AttestationIdealReward = validatorsData[validatorIndex].AttestationsIdealHeadReward +
			validatorsData[validatorIndex].AttestationsIdealSourceReward +
			validatorsData[validatorIndex].AttestationsIdealTargetReward +
			validatorsData[validatorIndex].AttestationsIdealInactivityReward +
			validatorsData[validatorIndex].AttestationsIdealInclusionsReward
	}

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
func mustParseInt(s string) int {
	if s == "" {
		return 0
	}

	r, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		panic(err)
	}
	return int(r)
}

type validatorDashboardDataRow struct {
	Index uint64

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

	BlockScheduled int // done
	BlocksProposed int // done

	BlocksClReward int64 // done
	BlocksElReward int64

	SyncScheduled int   // done
	SyncExecuted  int   // done
	SyncReward    int64 // done

	Slashed bool // done

	BalanceStart int64 // done
	BalanceEnd   int64 // done

	DepositsCount  int   // done
	DepositsAmount int64 // done

	WithdrawalsCount  int   // done
	WithdrawalsAmount int64 // done
}
