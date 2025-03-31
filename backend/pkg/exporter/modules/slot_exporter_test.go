package modules

import (
	"crypto/rand"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"testing"
	"time"
)

func TestPectraQueue(t *testing.T) {
	epoch := uint64(3120)

	cfg := &types.Config{}
	cfg.Chain.ClConfig.SlotsPerEpoch = 32                                 // uint64(spec.Data.SlotsPerEpoch)
	cfg.Chain.ClConfig.SecondsPerSlot = 12                                // uint64(spec.Data.SecondsPerSlot)
	cfg.Chain.ClConfig.ChurnLimitQuotient = 65536                         // uint64(spec.Data.ChurnLimitQuotient)
	cfg.Chain.ClConfig.MinPerEpochChurnLimit = 4                          // uint64(spec.Data.MinPerEpochChurnLimit)
	cfg.Chain.ClConfig.EffectiveBalanceIncrement = 1000000000             // uint64(spec.Data.EffectiveBalanceIncrement)
	cfg.Chain.ClConfig.MaxPerEpochActivationExitChurnLimit = 256000000000 // uint64(spec.Data.MaxPerEpochActivationExitChurnLimit)
	utils.Config = cfg

	validators := &constypes.StandardValidatorsResponse{
		Finalized:           false,
		ExecutionOptimistic: false,
	}

	// add 10000 random validators to the set
	validatorSetSize := 1000000
	for i := 0; i < validatorSetSize; i++ {

		validators.Data = append(validators.Data, constypes.StandardValidator{
			Index: uint64(i),
			Validator: struct {
				Pubkey                     hexutil.Bytes `json:"pubkey"`
				WithdrawalCredentials      hexutil.Bytes `json:"withdrawal_credentials"`
				EffectiveBalance           uint64        `json:"effective_balance,string"`
				Slashed                    bool          `json:"slashed"`
				ActivationEligibilityEpoch uint64        `json:"activation_eligibility_epoch,string"`
				ActivationEpoch            uint64        `json:"activation_epoch,string"`
				ExitEpoch                  uint64        `json:"exit_epoch,string"`
				WithdrawableEpoch          uint64        `json:"withdrawable_epoch,string"`
			}{
				Pubkey:           generateRandomBytes(t, 42),
				ActivationEpoch:  epoch - 225,
				ExitEpoch:        epoch + 225,
				EffectiveBalance: 32 * params.GWei,
			},
		})
	}

	deposits := constypes.StandardBeaconPendingDepositsResponse{
		Finalized:           true,
		ExecutionOptimistic: false,
	}
	depositSetSize := 500
	for i := 0; i < depositSetSize; i++ {
		deposits.Data = append(deposits.Data, struct {
			Pubkey                hexutil.Bytes `json:"pubkey"`
			WithdrawalCredentials hexutil.Bytes `json:"withdrawal_credentials"`
			Amount                uint64        `json:"amount,string"`
			Signature             hexutil.Bytes `json:"signature"`
			Slot                  uint64        `json:"slot,string"`
		}{
			Pubkey: generateRandomBytes(t, 42),
			Amount: 32 * params.GWei,
		})
	}

	topUpSetSize := 50
	for i := 0; i < topUpSetSize; i++ {
		deposits.Data = append(deposits.Data, struct {
			Pubkey                hexutil.Bytes `json:"pubkey"`
			WithdrawalCredentials hexutil.Bytes `json:"withdrawal_credentials"`
			Amount                uint64        `json:"amount,string"`
			Signature             hexutil.Bytes `json:"signature"`
			Slot                  uint64        `json:"slot,string"`
		}{
			Pubkey: validators.Data[0].Validator.Pubkey, // simulate topping up the validators in question
			Amount: 32 * params.GWei,
		})
	}

	t.Logf("retrieved state with %d pending deposits", len(deposits.Data))

	// calculate the total active effective balance
	// active = all validators with an activation epoch >= the current epoch
	// and an exit epoch < the current epoch
	totalActiveEffectiveBalance := uint64(0)

	// at the same time build up an index for easy pubkey -> index lookup
	pubkeyToIndexMap := make(map[string]uint64)
	for _, v := range validators.Data {
		pubkeyToIndexMap[v.Validator.Pubkey.String()] = v.Index
		if epoch >= v.Validator.ActivationEpoch && epoch < v.Validator.ExitEpoch {
			totalActiveEffectiveBalance += v.Validator.EffectiveBalance
		}
	}
	t.Logf("total active effective balance: %d", totalActiveEffectiveBalance)

	// figure out the amount of validators & ether in the deposit queue
	enteringEthAmount := uint64(0)
	enteringValidatorCount := uint64(0)
	for _, deposit := range deposits.Data {
		enteringEthAmount += deposit.Amount
		// only increment entering new validators for validators that don't have an index yet
		// other deposits are top-ups for existing validators
		if _, found := pubkeyToIndexMap[deposit.Pubkey.String()]; !found {
			enteringValidatorCount++
		}
	}
	t.Logf("entering validator count: %d", enteringValidatorCount)
	t.Logf("entering eth amount: %d", enteringEthAmount)
	if enteringEthAmount != 17600000000000 {
		t.Fatalf("entering eth amount should be 17600000000000 but got %d", enteringEthAmount)
	}

	// retrieve the dynamic eth churn per epoch (valid for activation & exits)
	etherChurnByEpoch := GetActivationExitChurnLimit(totalActiveEffectiveBalance)
	etherChurnByDay := etherChurnByEpoch * utils.EpochsPerDay()

	t.Logf("etherChurnByEpoch: %d", etherChurnByEpoch)
	if etherChurnByEpoch != 256000000000 {
		t.Fatalf("ether churn by epoch should be 256000000000 but got %d", etherChurnByEpoch)
	}
	t.Logf("etherChurnByDay: %d", etherChurnByDay)

	depositQueueTimeSec := (float64(enteringEthAmount) / float64(etherChurnByDay)) * 60 * 60 * 24
	depositQueueTime := time.Duration(depositQueueTimeSec) * time.Second
	t.Logf("deposit queue time: %v", depositQueueTime)

	depositsEpochTime := enteringEthAmount / etherChurnByEpoch
	t.Logf("deposits epoch time: %v", depositsEpochTime)

	if depositsEpochTime != 68 {
		t.Fatalf("deposits epoch time should be 68 but got %d", depositsEpochTime)
	}
}

func generateRandomBytes(t *testing.T, n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		t.Fatalf("failed to generate random bytes: %v", err)
	}

	return b
}
