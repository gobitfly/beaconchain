package modules

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type dashboardData struct {
	ModuleContext
}

func NewDashboardDataModule(moduleContext ModuleContext) ModuleInterface {
	return &dashboardData{
		ModuleContext: moduleContext,
	}
}

func (d *dashboardData) Init() error {
	// todo aggregator
	return nil
}

var onProcessingMutex = &sync.Mutex{}

func (d *dashboardData) OnFinalizedCheckpoint(_ *constypes.StandardFinalizedCheckpointResponse) error {
	onProcessingMutex.Lock()
	defer onProcessingMutex.Unlock()

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
			log.Infof("dashboard epoch data already exported for epoch %d", res.Data.Finalized.Epoch)
			return nil
		}

		// todo think about backfilling process

		// backfill if needed, skip backfilling older than RetainEpochDuration/3 since the time it would take to backfill exceeds the retention period anyway
		// if res.Data.Finalized.Epoch-latestExported > 0 && res.Data.Finalized.Epoch-latestExported < RetainEpochDuration/3 {
		// 	// backfill first
		// 	log.Infof("backfilling dashboard epoch data from epoch %d to %d", latestExported+1, res.Data.Finalized.Epoch-1)
		// 	for epoch := res.Data.Finalized.Epoch - 1; epoch > latestExported; epoch-- {
		// 		err := d.exportEpochData(int(epoch))
		// 		if err != nil {
		// 			log.Error(err, "failed to export dashboard epoch data", 0, map[string]interface{}{"epoch": epoch})
		// 		}
		// 	}
		// }
	}

	err = d.exportEpochData(int(res.Data.Finalized.Epoch))
	if err != nil {
		return err
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

func (d *dashboardData) exportEpochData(epoch int) error {
	spec, err := d.CL.GetSpec()
	if err != nil {
		return err
	}

	start := time.Now()
	data := d.getData(epoch, int(spec.Data.SlotsPerEpoch))
	if data == nil {
		return errors.New("can not get data")
	}
	log.Infof("retrieved data for epoch %d in %v", epoch, time.Since(start))

	start = time.Now()
	domain, err := utils.GetSigningDomain()
	if err != nil {
		return err
	}

	result := process(data, domain)
	log.Infof("processed data for epoch %d in %v", epoch, time.Since(start))

	start = time.Now()
	err = WriteEpochData(epoch, result)
	if err != nil {
		return err
	}
	log.Infof("wrote data for epoch %d in %v", epoch, time.Since(start))

	log.Infof("successfully wrote dashboard epoch data for epoch %d", epoch)
	return nil
}

type Data struct {
	startBalances            *constypes.StandardValidatorsResponse
	endBalances              *constypes.StandardValidatorsResponse
	proposerAssignments      *constypes.StandardProposerAssignmentsResponse
	syncCommitteeAssignments *constypes.StandardSyncCommitteesResponse
	attestationRewards       *constypes.StandardAttestationRewardsResponse
	beaconBlockData          map[int]*constypes.StandardBeaconSlotResponse
	beaconBlockRewardData    map[int]*constypes.StandardBlockRewardsResponse
	syncCommitteeRewardData  map[int]*constypes.StandardSyncCommitteeRewardsResponse
}

func (d *dashboardData) getData(epoch, slotsPerEpoch int) *Data {
	var result Data
	var err error

	firstSlotOfEpoch := epoch * slotsPerEpoch
	firstSlotOfPreviousEpoch := firstSlotOfEpoch - 1
	lastSlotOfEpoch := firstSlotOfEpoch + slotsPerEpoch

	result.beaconBlockData = make(map[int]*constypes.StandardBeaconSlotResponse)
	result.beaconBlockRewardData = make(map[int]*constypes.StandardBlockRewardsResponse)
	result.syncCommitteeRewardData = make(map[int]*constypes.StandardSyncCommitteeRewardsResponse)

	// retrieve the validator balances at the start of the epoch
	log.Infof("retrieving start balances using state at slot %d", firstSlotOfPreviousEpoch)
	result.startBalances, err = d.CL.GetValidators(firstSlotOfPreviousEpoch, nil, nil)

	if err != nil {
		log.Error(err, "can not get validators balances", 0, map[string]interface{}{"firstSlotOfPreviousEpoch": firstSlotOfPreviousEpoch})
		return nil
	}

	// retrieve proposer assignments for the epoch in order to attribute missed slots
	log.Infof("retrieving proposer assignments")
	result.proposerAssignments, err = d.CL.GetPropoalAssignments(epoch)
	if err != nil {
		log.Error(err, "can not get proposer assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
	log.Infof("retrieving sync committee assignments")
	result.syncCommitteeAssignments, err = d.CL.GetSyncCommitteesAssignments(epoch, int64(firstSlotOfEpoch))
	if err != nil {
		log.Error(err, "can not get sync committee assignments", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// attestation rewards
	log.Infof("retrieving attestation rewards data")
	result.attestationRewards, err = d.CL.GetAttestationRewards(uint64(epoch))

	if err != nil {
		log.Error(err, "can not get attestation rewards", 0, map[string]interface{}{"epoch": epoch})
		return nil
	}

	// retrieve the data for all blocks that were proposed in this epoch
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		log.Infof("retrieving data for block at slot %d", slot)
		block, err := d.CL.GetSlot(slot)
		if err != nil {
			httpErr, _ := network.SpecificError(err)
			if httpErr != nil && httpErr.StatusCode == 404 {
				continue // missed
			}
			log.Fatal(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		if block.Data.Message.StateRoot == "" {
			// todo better network handling, if 404 just log info, else log error
			log.Error(err, "can not get block data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockData[slot] = block

		blockReward, err := d.CL.GetPropoalRewards(slot)
		if err != nil {
			log.Error(err, "can not get block reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.beaconBlockRewardData[slot] = blockReward

		syncRewards, err := d.CL.GetSyncRewards(slot)
		if err != nil {
			log.Error(err, "can not get sync committee reward data", 0, map[string]interface{}{"slot": slot})
			continue
		}
		result.syncCommitteeRewardData[slot] = syncRewards
	}

	// retrieve the validator balances at the end of the epoch
	log.Infof("retrieving end balances using state at slot %d", lastSlotOfEpoch)
	result.endBalances, err = d.CL.GetValidators(lastSlotOfEpoch, nil, nil)

	if err != nil {
		log.Error(err, "can not get validators balances", 0, map[string]interface{}{"lastSlotOfEpoch": lastSlotOfEpoch})
		return nil
	}

	return &result
}

func process(data *Data, domain []byte) []*validatorDashboardDataRow {
	validatorsData := make([]*validatorDashboardDataRow, len(data.endBalances.Data))

	idealAttestationRewards := make(map[int64]int)
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
				log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, block.Data.Message.Slot, err, depositData.Data.Signature), "", 0)

				// if the validator hat a valid deposit prior to the current one, count the invalid towards the balance
				if validatorsData[pubkeyToIndexMapEnd[depositData.Data.Pubkey]].DepositsCount > 0 {
					log.Infof("validator had a valid deposit in some earlier block of the epoch, count the invalid towards the balance")
				} else if _, ok := pubkeyToIndexMapStart[depositData.Data.Pubkey]; ok {
					log.Infof("validator had a valid deposit in some block prior to the current epoch, count the invalid towards the balance")
				} else {
					log.Infof("validator did not have a prior valid deposit, do not count the invalid towards the balance")
					continue
				}
			}

			validatorIndex := pubkeyToIndexMapEnd[depositData.Data.Pubkey]

			validatorsData[validatorIndex].DepositsAmount += depositData.Data.Amount
			validatorsData[validatorIndex].DepositsCount++
		}

		for _, withdrawal := range block.Data.Message.Body.ExecutionPayload.Withdrawals {
			validatorIndex := withdrawal.ValidatorIndex
			validatorsData[validatorIndex].WithdrawalsAmount += withdrawal.Amount
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
		idealRewardsOfValidator := data.attestationRewards.Data.IdealRewards[idealAttestationRewards[int64(data.startBalances.Data[validatorIndex].Validator.EffectiveBalance)]]
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

const PartitionEpochWidth = 3
const RetainEpochDuration = 300 // in epochs

func getPartitionRange(epoch int) (int, int) {
	startOfPartition := epoch / PartitionEpochWidth * PartitionEpochWidth // inclusive
	endOfPartition := startOfPartition + PartitionEpochWidth              // exclusive
	return startOfPartition, endOfPartition
}

func WriteEpochData(epoch int, data []*validatorDashboardDataRow) error {

	// Create table if needed
	startOfPartition, endOfPartition := getPartitionRange(epoch)

	err := createEpochPartition(startOfPartition, endOfPartition)
	if err != nil {
		return errors.Wrap(err, "failed to create epoch partition")
	}

	conn, err := db.AlloyWriter.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())

		if err != nil {
			return errors.Wrap(err, "error starting transaction")
		}

		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Error(err, "error rolling back transaction", 0)
			}
		}()

		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"dashboard_data_epoch"}, []string{
			"validatorindex",
			"epoch",
			"attestations_source_reward",
			"attestations_target_reward",
			"attestations_head_reward",
			"attestations_inactivity_reward",
			"attestations_inclusion_reward",
			"attestations_reward",
			"attestations_ideal_source_reward",
			"attestations_ideal_target_reward",
			"attestations_ideal_head_reward",
			"attestations_ideal_inactivity_reward",
			"attestations_ideal_inclusion_reward",
			"attestations_ideal_reward",
			"blocks_scheduled",
			"blocks_proposed",
			"blocks_cl_reward",
			"blocks_el_reward",
			"sync_scheduled",
			"sync_executed",
			"sync_rewards",
			"slashed",
			"balance_start",
			"balance_end",
			"deposits_count",
			"deposits_amount",
			"withdrawals_count",
			"withdrawals_amount",
		}, pgx.CopyFromSlice(len(data), func(i int) ([]interface{}, error) {
			return []interface{}{
				data[i].Index,
				epoch,
				data[i].AttestationsSourceReward,
				data[i].AttestationsTargetReward,
				data[i].AttestationsHeadReward,
				data[i].AttestationsInactivityReward,
				data[i].AttestationsInclusionsReward,
				data[i].AttestationReward,
				data[i].AttestationsIdealSourceReward,
				data[i].AttestationsIdealTargetReward,
				data[i].AttestationsIdealHeadReward,
				data[i].AttestationsIdealInactivityReward,
				data[i].AttestationsIdealInclusionsReward,
				data[i].AttestationIdealReward,
				data[i].BlockScheduled,
				data[i].BlocksProposed,
				data[i].BlocksClReward,
				data[i].BlocksElReward,
				data[i].SyncScheduled,
				data[i].SyncExecuted,
				data[i].SyncReward,
				data[i].Slashed,
				data[i].BalanceStart,
				data[i].BalanceEnd,
				data[i].DepositsCount,
				data[i].DepositsAmount,
				data[i].WithdrawalsCount,
				data[i].WithdrawalsAmount,
			}, nil
		}))

		if err != nil {
			return errors.Wrap(err, "error copying data")
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return errors.Wrap(err, "error committing transaction")
		}
		return nil
	})

	//Clear old partitions
	//todo delete in aggregator, not here
	for i := 0; ; i++ {
		startOfPartition, endOfPartition := getPartitionRange(epoch - RetainEpochDuration - i)
		finished, err := deleteEpochPartition(startOfPartition, endOfPartition)
		if err != nil {
			return err
		}

		if finished {
			break
		}
	}

	return nil
}

func createEpochPartition(epochFrom, epochTo int) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS dashboard_data_epoch_%d_%d 
		PARTITION OF dashboard_data_epoch
			FOR VALUES FROM (%[1]d) TO (%[2]d)
		`,
		epochFrom, epochTo,
	))
	return err
}

// Returns finished, error
func deleteEpochPartition(epochFrom, epochTo int) (bool, error) {
	st, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		DROP TABLE IF EXISTS dashboard_data_epoch_%d_%d
		`,
		epochFrom, epochTo,
	))
	if err != nil {
		return false, err
	}
	rowsAffected, err := st.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return true, nil
	}

	return false, nil
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
	BlocksElReward decimal.Decimal

	SyncScheduled int   // done
	SyncExecuted  int   // done
	SyncReward    int64 // done

	Slashed bool // done

	BalanceStart uint64 // done
	BalanceEnd   uint64 // done

	DepositsCount  int    // done
	DepositsAmount uint64 // done

	WithdrawalsCount  int    // done
	WithdrawalsAmount uint64 // done
}
