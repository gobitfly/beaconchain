package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"eth2-exporter/exporter"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

func main() {

	url := flag.String("url", "http://localhost:4000", "")

	// discordWebhookReportUrl := flag.String("discord-url", "", "")
	// discordWebhookUser := flag.String("discord-user", "", "")
	epoch := flag.Int("epoch", -1, "")
	// concurrency := flag.Int("concurrency", 1, "")

	flag.Parse()

	cfg := &types.Config{}

	err := utils.ReadConfig(cfg, "")
	if err != nil {
		logrus.Fatal(err)
	}
	utils.Config = cfg

	domain, err := utils.GetSigningDomain()
	if err != nil {
		logrus.Fatal(err)
	}

	spec := &exporter.BeaconSpecResponse{}

	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/config/spec", *url), nil, spec)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(cfg.Chain.ClConfig.DepositChainID)
	logrus.Info(cfg.Chain.ClConfig.SlotsPerEpoch)

	firstSlotOfEpoch := *epoch * int(cfg.Chain.ClConfig.SlotsPerEpoch)
	firstSlotOfPreviousEpoch := firstSlotOfEpoch - 1
	lastSlotOfEpoch := firstSlotOfEpoch + int(cfg.Chain.ClConfig.SlotsPerEpoch)

	beaconBlockData := make(map[int]*getBeaconSlotResponse)
	beaconBlockRewardData := make(map[int]*getBlockRewardsResponse)
	syncCommitteeRewardData := make(map[int]*getSyncCommitteeRewardsResponse)

	// retrieve the validator balances at the start of the epoch
	logrus.Infof("retrieving start balances using state at slot %d", firstSlotOfPreviousEpoch)
	startBalances := &getValidatorsResponse{}
	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", *url, firstSlotOfPreviousEpoch), nil, startBalances)
	if err != nil {
		logrus.Fatalf("error retrieving start balances for epoch %d: %v", *epoch, err)
	}

	// retrieve proposer assignments for the epoch in order to attribute missed slots
	logrus.Infof("retrieving proposer assignments")
	proposerAssignments := &getProposerAssignmentsResponse{}
	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", *url, *epoch), nil, proposerAssignments)

	if err != nil {
		logrus.Fatalf("error retrieving proposer assignments for epoch %d: %v", *epoch, err)
	}
	// retrieve sync committee assignments for the epoch in order to attribute missed sync assignments
	logrus.Infof("retrieving sync committee assignments")
	syncCommitteeAssignments := &getSyncCommitteeAssignmentsResponse{}
	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees?epoch=%d", *url, firstSlotOfEpoch, *epoch), nil, syncCommitteeAssignments)

	if err != nil {
		logrus.Fatalf("error retrieving proposer assignments for epoch %d: %v", *epoch, err)
	}

	// attestation rewards
	logrus.Infof("retrieving attestation rewards data")
	attestationRewards := &getAttestationRewardsResponse{}
	err = utils.HttpReq(context.Background(), http.MethodPost, fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%d", *url, *epoch), []string{}, attestationRewards)

	if err != nil {
		logrus.Fatalf("error retrieving proposer assignments for epoch %d: %v", *epoch, err)
	}

	// retrieve the data for all blocks that were proposed in this epoch
	for slot := firstSlotOfEpoch; slot <= lastSlotOfEpoch; slot++ {
		logrus.Infof("retrieving data for block at slot %d", slot)
		block := &getBeaconSlotResponse{}

		err := utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", *url, slot), nil, block)

		if err != nil {
			httpErr, ok := err.(utils.HttpReqHttpError)
			if ok {
				if httpErr.StatusCode == 404 {
					logrus.Infof("no block at slot %d", slot)
				} else {
					logrus.Fatalf("error retrieving data for slot %d: %v", slot, err)
				}
				continue
			} else {
				logrus.Fatalf("error retrieving data for slot %d: %v", slot, err)
			}
		}
		beaconBlockData[slot] = block

		blockReward := &getBlockRewardsResponse{}
		err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%d", *url, slot), nil, blockReward)
		if err != nil {
			logrus.Fatalf("error retrieving reward data for slot %d: %v", slot, err)
		}
		beaconBlockRewardData[slot] = blockReward

		syncRewards := &getSyncCommitteeRewardsResponse{}
		err = utils.HttpReq(context.Background(), http.MethodPost, fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%d", *url, slot), []string{}, syncRewards)
		if err != nil {
			logrus.Fatalf("error retrieving sync committee reward data for slot %d: %v", slot, err)
		}
		syncCommitteeRewardData[slot] = syncRewards
	}

	// retrieve the validator balances at the end of the epoch
	logrus.Infof("retrieving end balances using state at slot %d", lastSlotOfEpoch)
	endBalances := &getValidatorsResponse{}
	err = utils.HttpReq(context.Background(), http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", *url, lastSlotOfEpoch), nil, endBalances)
	if err != nil {
		logrus.Fatalf("error retrieving end balances for epoch %d: %v", *epoch, err)
	}

	validatorsData := make([]*validatorDashboardDataRow, len(endBalances.Data))

	idealAttestationRewards := make(map[string]int)
	for i, idealReward := range attestationRewards.Data.IdealRewards {
		idealAttestationRewards[idealReward.EffectiveBalance] = i
	}

	pubkeyToIndexMapEnd := make(map[string]int64)
	pubkeyToIndexMapStart := make(map[string]int64)
	// write start & end balances and slashed status
	for i := 0; i < len(validatorsData); i++ {
		validatorsData[i] = &validatorDashboardDataRow{}
		if i < len(startBalances.Data) {
			validatorsData[i].BalanceStart = mustParseInt64(startBalances.Data[i].Balance)
			pubkeyToIndexMapStart[startBalances.Data[i].Validator.Pubkey] = int64(i)
		}
		validatorsData[i].BalanceEnd = mustParseInt64(endBalances.Data[i].Balance)
		validatorsData[i].Slashed = endBalances.Data[i].Validator.Slashed

		pubkeyToIndexMapEnd[endBalances.Data[i].Validator.Pubkey] = int64(i)
	}

	// write scheduled block data
	for _, proposerAssignment := range proposerAssignments.Data {
		proposerIndex := mustParseInt(proposerAssignment.ValidatorIndex)
		validatorsData[proposerIndex].BlockScheduled++
	}

	// write scheduled sync committee data
	for _, validator := range syncCommitteeAssignments.Data.Validators {
		validatorsData[mustParseInt64(validator)].SyncScheduled = len(beaconBlockData) // take into account missed slots
	}

	// write proposer rewards data
	for _, reward := range beaconBlockRewardData {
		validatorsData[mustParseInt(reward.Data.ProposerIndex)].BlocksClReward += mustParseInt64(reward.Data.Attestations) + mustParseInt64(reward.Data.AttesterSlashings) + mustParseInt64(reward.Data.ProposerSlashings) + mustParseInt64(reward.Data.SyncAggregate)
	}

	// write sync committee reward data & sync committee execution stats
	for _, rewards := range syncCommitteeRewardData {
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
	for _, block := range beaconBlockData {
		validatorsData[mustParseInt64(block.Data.Message.ProposerIndex)].BlocksProposed++

		for depositIndex, depositData := range block.Data.Message.Body.Deposits {

			// TODO: properly verify that deposit is valid:
			// if signature is valid I count the the deposit towards the balance
			// if signature is invalid and the validator was in the state at the beginning of the epoch I count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there were no valid deposits in the block prior I DO NOT count the deposit towards the balance
			// if signature is invalid and the validator was NOT in the state at the beginning of the epoch and there was a VALID deposit in the blocks prior I DO COUNT the deposit towards the balance
			err = deposit.VerifyDepositSignature(&ethpb.Deposit_Data{
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
	for _, attestationReward := range attestationRewards.Data.TotalRewards {

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
		idealRewardsOfValidator := attestationRewards.Data.IdealRewards[idealAttestationRewards[startBalances.Data[validatorIndex].Validator.EffectiveBalance]]
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

// type beaconNodeErrorResponse struct {
// 	Code    int    `json:"code"`
// 	Message string `json:"message"`
// }

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

type getSyncCommitteeAssignmentsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Validators          []string   `json:"validators"`
		ValidatorAggregates [][]string `json:"validator_aggregates"`
	} `json:"data"`
}

type getAttestationRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		IdealRewards []struct {
			EffectiveBalance string `json:"effective_balance"`
			Head             string `json:"head"`
			Target           string `json:"target"`
			Source           string `json:"source"`
			InclusionDelay   string `json:"inclusion_delay"`
			Inactivity       string `json:"inactivity"`
		} `json:"ideal_rewards"`
		TotalRewards []struct {
			ValidatorIndex string `json:"validator_index"`
			Head           string `json:"head"`
			Target         string `json:"target"`
			Source         string `json:"source"`
			InclusionDelay string `json:"inclusion_delay"`
			Inactivity     string `json:"inactivity"`
		} `json:"total_rewards"`
	} `json:"data"`
}

type getSyncCommitteeRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		ValidatorIndex string `json:"validator_index"`
		Reward         string `json:"reward"`
	} `json:"data"`
}

type getBlockRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		ProposerIndex     string `json:"proposer_index"`
		Total             string `json:"total"`
		Attestations      string `json:"attestations"`
		SyncAggregate     string `json:"sync_aggregate"`
		ProposerSlashings string `json:"proposer_slashings"`
		AttesterSlashings string `json:"attester_slashings"`
	} `json:"data"`
}

type getValidatorsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		Index     string `json:"index"`
		Balance   string `json:"balance"`
		Status    string `json:"status"`
		Validator struct {
			Pubkey                     string `json:"pubkey"`
			WithdrawalCredentials      string `json:"withdrawal_credentials"`
			EffectiveBalance           string `json:"effective_balance"`
			Slashed                    bool   `json:"slashed"`
			ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
			ActivationEpoch            string `json:"activation_epoch"`
			ExitEpoch                  string `json:"exit_epoch"`
			WithdrawableEpoch          string `json:"withdrawable_epoch"`
		} `json:"validator"`
	} `json:"data"`
}

type getProposerAssignmentsResponse struct {
	DependentRoot       string `json:"dependent_root"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Data                []struct {
		Pubkey         string `json:"pubkey"`
		ValidatorIndex string `json:"validator_index"`
		Slot           string `json:"slot"`
	} `json:"data"`
}

type getBeaconSlotResponse struct {
	Version             string `json:"version"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Finalized           bool   `json:"finalized"`
	Data                struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			Body          struct {
				RandaoReveal string `json:"randao_reveal"`
				Eth1Data     struct {
					DepositRoot  string `json:"deposit_root"`
					DepositCount string `json:"deposit_count"`
					BlockHash    string `json:"block_hash"`
				} `json:"eth1_data"`
				Graffiti          string `json:"graffiti"`
				ProposerSlashings []struct {
					SignedHeader1 struct {
						Message struct {
							Slot          string `json:"slot"`
							ProposerIndex string `json:"proposer_index"`
							ParentRoot    string `json:"parent_root"`
							StateRoot     string `json:"state_root"`
							BodyRoot      string `json:"body_root"`
						} `json:"message"`
						Signature string `json:"signature"`
					} `json:"signed_header_1"`
					SignedHeader2 struct {
						Message struct {
							Slot          string `json:"slot"`
							ProposerIndex string `json:"proposer_index"`
							ParentRoot    string `json:"parent_root"`
							StateRoot     string `json:"state_root"`
							BodyRoot      string `json:"body_root"`
						} `json:"message"`
						Signature string `json:"signature"`
					} `json:"signed_header_2"`
				} `json:"proposer_slashings"`
				AttesterSlashings []struct {
					Attestation1 struct {
						AttestingIndices []string `json:"attesting_indices"`
						Signature        string   `json:"signature"`
						Data             struct {
							Slot            string `json:"slot"`
							Index           string `json:"index"`
							BeaconBlockRoot string `json:"beacon_block_root"`
							Source          struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"source"`
							Target struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"target"`
						} `json:"data"`
					} `json:"attestation_1"`
					Attestation2 struct {
						AttestingIndices []string `json:"attesting_indices"`
						Signature        string   `json:"signature"`
						Data             struct {
							Slot            string `json:"slot"`
							Index           string `json:"index"`
							BeaconBlockRoot string `json:"beacon_block_root"`
							Source          struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"source"`
							Target struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"target"`
						} `json:"data"`
					} `json:"attestation_2"`
				} `json:"attester_slashings"`
				Attestations []struct {
					AggregationBits string `json:"aggregation_bits"`
					Signature       string `json:"signature"`
					Data            struct {
						Slot            string `json:"slot"`
						Index           string `json:"index"`
						BeaconBlockRoot string `json:"beacon_block_root"`
						Source          struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"source"`
						Target struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"target"`
					} `json:"data"`
				} `json:"attestations"`
				Deposits []struct {
					Proof []string `json:"proof"`
					Data  struct {
						Pubkey                string `json:"pubkey"`
						WithdrawalCredentials string `json:"withdrawal_credentials"`
						Amount                string `json:"amount"`
						Signature             string `json:"signature"`
					} `json:"data"`
				} `json:"deposits"`
				VoluntaryExits []struct {
					Message struct {
						Epoch          string `json:"epoch"`
						ValidatorIndex string `json:"validator_index"`
					} `json:"message"`
					Signature string `json:"signature"`
				} `json:"voluntary_exits"`
				SyncAggregate struct {
					SyncCommitteeBits      string `json:"sync_committee_bits"`
					SyncCommitteeSignature string `json:"sync_committee_signature"`
				} `json:"sync_aggregate"`
				ExecutionPayload struct {
					ParentHash    string   `json:"parent_hash"`
					FeeRecipient  string   `json:"fee_recipient"`
					StateRoot     string   `json:"state_root"`
					ReceiptsRoot  string   `json:"receipts_root"`
					LogsBloom     string   `json:"logs_bloom"`
					PrevRandao    string   `json:"prev_randao"`
					BlockNumber   string   `json:"block_number"`
					GasLimit      string   `json:"gas_limit"`
					GasUsed       string   `json:"gas_used"`
					Timestamp     string   `json:"timestamp"`
					ExtraData     string   `json:"extra_data"`
					BaseFeePerGas string   `json:"base_fee_per_gas"`
					BlockHash     string   `json:"block_hash"`
					Transactions  []string `json:"transactions"`
					Withdrawals   []struct {
						Index          string `json:"index"`
						ValidatorIndex string `json:"validator_index"`
						Address        string `json:"address"`
						Amount         string `json:"amount"`
					} `json:"withdrawals"`
				} `json:"execution_payload"`
				BlsToExecutionChanges []any `json:"bls_to_execution_changes"`
			} `json:"body"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"data"`
}

func getSlot(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getValidators(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getCommittees(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/committees", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getSyncCommittees(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getPropoalAssignments(url string, httpClient *http.Client, epoch int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", url, epoch)
	return genericRequest("GET", requestUrl, httpClient)
}

func getPropoalRewards(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%d", url, slot)
	return genericRequest("GET", requestUrl, httpClient)
}

func getSyncRewards(url string, httpClient *http.Client, slot int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%d", url, slot)
	return genericRequest("POST", requestUrl, httpClient)
}

func getAttestationRewards(url string, httpClient *http.Client, epoch int) ([]byte, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%d", url, epoch)
	return genericRequest("POST", requestUrl, httpClient)
}

func genericRequest(method string, requestUrl string, httpClient *http.Client) ([]byte, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {

		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		if res.StatusCode == http.StatusBadRequest {
			return nil, nil
		}
		if res.StatusCode == http.StatusInternalServerError {
			return nil, nil
		}
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	// logrus.Info(string(resString))

	if strings.Contains(string(resString), `"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return compress(resString), nil
}

func compress(src []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(src)
	if err != nil {
		logrus.Fatalf("error writing to gzip writer: %v", err)
	}
	if err := zw.Close(); err != nil {
		logrus.Fatalf("error closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

func decompress(src []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		logrus.Fatalf("error creating gzip reader: %v", err)
	}

	data, err := io.ReadAll(zr)
	if err != nil {
		logrus.Fatalf("error reading from gzip reader: %v", err)
	}
	return data
}
