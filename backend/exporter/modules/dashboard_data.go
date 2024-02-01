package modules

import (
	"strconv"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gobitfly/beaconchain/commons/utils"
	ctypes "github.com/gobitfly/beaconchain/consapi/types"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type dashboardData struct {
	ModuleContext
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterfaceEpoch {
	return &dashboardData{
		ModuleContext: moduleContext,
	}
}

func (d *dashboardData) Start(epoch int) {
	spec, err := d.CL.GetSpec()
	if err != nil {
		utils.LogFatal(err, "can not get spec", 0)
		return
	}

	data := d.getData(epoch, int(spec.Data.SlotsPerEpoch))
	if data == nil {
		return
	}

	domain, err := utils.GetSigningDomain()
	if err != nil {
		utils.LogFatal(err, "can not get signing domain", 0)
		return
	}

	process(data, domain)

	// todo store in db

}

type Data struct {
	startBalances            ctypes.StandardValidatorsResponse
	endBalances              ctypes.StandardValidatorsResponse
	proposerAssignments      ctypes.StandardProposerAssignmentsResponse
	syncCommitteeAssignments ctypes.StandardSyncCommitteesResponse
	attestationRewards       ctypes.StandardAttestationRewardsResponse
	beaconBlockData          map[int]*ctypes.StandardBeaconSlotResponse
	beaconBlockRewardData    map[int]*ctypes.StandardBlockRewardsResponse
	syncCommitteeRewardData  map[int]*ctypes.StandardSyncCommitteeRewardsResponse
}

func (d *dashboardData) getData(epoch, slotsPerEpoch int) *Data {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * int(slotsPerEpoch)
	firstSlotOfPreviousEpoch := firstSlotOfEpoch - 1
	lastSlotOfEpoch := firstSlotOfEpoch + int(slotsPerEpoch)

	result.beaconBlockData = make(map[int]*ctypes.StandardBeaconSlotResponse)
	result.beaconBlockRewardData = make(map[int]*ctypes.StandardBlockRewardsResponse)
	result.syncCommitteeRewardData = make(map[int]*ctypes.StandardSyncCommitteeRewardsResponse)

	// retrieve the validator balances at the start of the epoch
	logrus.Infof("retrieving start balances using state at slot %d", firstSlotOfPreviousEpoch)
	result.startBalances, err = d.CL.GetValidators(firstSlotOfPreviousEpoch)
	if err != nil {
		utils.LogError(err, "can not get validators balances", 0, map[string]interface{}{"firstSlotOfPreviousEpoch": firstSlotOfPreviousEpoch})
		return nil
	}

	// retrieve proposer assignments for the epoch in order to attribute missed slots
	logrus.Infof("retrieving proposer assignments")
	result.proposerAssignments, err = d.CL.GetPropoalAssignments(epoch)
	if err != nil {
		utils.LogError(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
	logrus.Infof("retrieving sync committee assignments")
	result.syncCommitteeAssignments, err = d.CL.GetSyncCommitteesAssignments(epoch, int64(firstSlotOfEpoch))
	if err != nil {
		utils.LogError(err, "can not get sync committee assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// attestation rewards
	logrus.Infof("retrieving attestation rewards data")
	result.attestationRewards, err = d.CL.GetAttestationRewards(epoch)
	if err != nil {
		utils.LogError(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve the data for all blocks that were proposed in this epoch
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		logrus.Infof("retrieving data for block at slot %d", slot)
		block, err := d.CL.GetSlot(slot)
		if err != nil {
			utils.LogFatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		if block.Data.Message.StateRoot == "" {
			// todo better network handling, if 404 just log info, else log error
			utils.LogError(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockData[slot] = &block

		blockReward, err := d.CL.GetPropoalRewards(slot)
		if err != nil {
			utils.LogError(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockRewardData[slot] = &blockReward

		syncRewards, err := d.CL.GetSyncRewards(slot)
		if err != nil {
			utils.LogError(err, "can not get sync committee reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.syncCommitteeRewardData[slot] = &syncRewards
	}

	// retrieve the validator balances at the end of the epoch
	logrus.Infof("retrieving end balances using state at slot %d", lastSlotOfEpoch)
	result.endBalances, err = d.CL.GetValidators(lastSlotOfEpoch)
	if err != nil {
		utils.LogError(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfEpoch": lastSlotOfEpoch})
		return nil
	}

	return &result
}

func process(data *Data, domain []byte) []*validatorDashboardDataRow {
	validatorsData := make([]*validatorDashboardDataRow, len(data.endBalances.Data))

	idealAttestationRewards := make(map[decimal.Decimal]int)
	for i, idealReward := range data.attestationRewards.Data.IdealRewards {
		idealAttestationRewards[idealReward.EffectiveBalance] = i
	}

	pubkeyToIndexMapEnd := make(map[string]int64)
	pubkeyToIndexMapStart := make(map[string]int64)
	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(data.startBalances.Data) {
			validatorsData[i].BalanceStart = data.startBalances.Data[i].Balance
			pubkeyToIndexMapStart[data.startBalances.Data[i].Validator.Pubkey] = int64(i)
		}
		validatorsData[i].BalanceEnd = data.endBalances.Data[i].Balance
		validatorsData[i].Slashed = data.endBalances.Data[i].Validator.Slashed

		pubkeyToIndexMapEnd[data.endBalances.Data[i].Validator.Pubkey] = int64(i)
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
			validatorIndex := reward.ValidatorIndex
			syncReward := reward.Reward
			validatorsData[validatorIndex].SyncReward += syncReward

			if syncReward > 0 {
				validatorsData[validatorIndex].SyncExecuted++
			}
		}
	}

	// write block specific data
	for _, block := range data.beaconBlockData {
		validatorsData[block.Data.Message.ProposerIndex].BlocksProposed++

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {

			// TODO: properly verify that deposit is valid:
			// if signature is valid I count the the deposit towards the balance
			// if signature is invalid and the validator was in the state at the beginning of the epoch I count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there were no valid deposits in the block prior I DO NOT count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there was a VALID deposit in the blocks prior I DO COUNT the deposit towards the balance

			err := utils.VerifyDepositSignature(&phase0.DepositData{
				PublicKey:             phase0.BLSPubKey(utils.MustParseHex(depositData.Data.Pubkey)),
				WithdrawalCredentials: utils.MustParseHex(depositData.Data.WithdrawalCredentials),
				Amount:                phase0.Gwei(uint64(depositData.Data.Amount)),
				Signature:             phase0.BLSSignature(utils.MustParseHex(depositData.Data.Signature)),
			}, domain)

			if err != nil {
				logrus.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, block.Data.Message.Slot, err, depositData.Data.Signature)

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

			validatorsData[validatorIndex].DepositsAmount += depositData.Data.Amount
			validatorsData[validatorIndex].DepositsCount++
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validatorIndex := withdrawal.ValidatorIndex
			validatorsData[validatorIndex].WithdrawalsAmount = validatorsData[validatorIndex].WithdrawalsAmount.Add(withdrawal.Amount)
			validatorsData[validatorIndex].WithdrawalsCount++
		}
	}

	// write attestation rewards data
	for _, attestationReward := range data.attestationRewards.Data.TotalRewards {

		validatorIndex := attestationReward.ValidatorIndex

		validatorsData[validatorIndex].AttestationsHeadReward = attestationReward.Head
		validatorsData[validatorIndex].AttestationsSourceReward = attestationReward.Source
		validatorsData[validatorIndex].AttestationsTargetReward = attestationReward.Target
		validatorsData[validatorIndex].AttestationsInactivityReward = attestationReward.Inactivity
		validatorsData[validatorIndex].AttestationsInclusionsReward = attestationReward.InclusionDelay
		validatorsData[validatorIndex].AttestationReward = validatorsData[validatorIndex].AttestationsHeadReward +
			validatorsData[validatorIndex].AttestationsSourceReward +
			validatorsData[validatorIndex].AttestationsTargetReward +
			validatorsData[validatorIndex].AttestationsInactivityReward +
			validatorsData[validatorIndex].AttestationsInclusionsReward
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[data.startBalances.Data[validatorIndex].Validator.EffectiveBalance]]
		validatorsData[validatorIndex].AttestationsIdealHeadReward = idealRewardsOfValidator.Head
		validatorsData[validatorIndex].AttestationsIdealTargetReward = idealRewardsOfValidator.Target
		validatorsData[validatorIndex].AttestationsIdealHeadReward = idealRewardsOfValidator.Head
		validatorsData[validatorIndex].AttestationsIdealInactivityReward = idealRewardsOfValidator.Inactivity
		validatorsData[validatorIndex].AttestationsIdealInclusionsReward = idealRewardsOfValidator.InclusionDelay

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

	BalanceStart decimal.Decimal // done
	BalanceEnd   decimal.Decimal // done

	DepositsCount  int   // done
	DepositsAmount int64 // done

	WithdrawalsCount  int             // done
	WithdrawalsAmount decimal.Decimal // done
}
