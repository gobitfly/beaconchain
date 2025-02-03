package rpc

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/contracts/oneinchoracle"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func getReceiver(tx *gethtypes.Transaction) []byte {
	if tx.To() != nil {
		return tx.To().Bytes()
	}
	return nil
}

func getTokens(tokenStr []string) []common.Address {
	tokens := make([]common.Address, 0, len(tokenStr))

	for _, token := range tokenStr {
		tokens = append(tokens, common.HexToAddress(token))
	}
	return tokens
}

func getBlockUncles(blkUncles []*gethtypes.Header) []*types.Eth1Block {
	uncles := make([]*types.Eth1Block, len(blkUncles))
	for i, uncle := range blkUncles {
		uncles[i] = &types.Eth1Block{
			Hash:        uncle.Hash().Bytes(),
			ParentHash:  uncle.ParentHash.Bytes(),
			UncleHash:   uncle.UncleHash.Bytes(),
			Coinbase:    uncle.Coinbase.Bytes(),
			Root:        uncle.Root.Bytes(),
			TxHash:      uncle.TxHash.Bytes(),
			ReceiptHash: uncle.ReceiptHash.Bytes(),
			Difficulty:  uncle.Difficulty.Bytes(),
			Number:      uncle.Number.Uint64(),
			GasLimit:    uncle.GasLimit,
			GasUsed:     uncle.GasUsed,
			Time:        timestamppb.New(time.Unix(int64(uncle.Time), 0)),
			Extra:       uncle.Extra,
			MixDigest:   uncle.MixDigest.Bytes(),
			Bloom:       uncle.Bloom.Bytes(),
		}
	}
	return uncles
}

func getBlockWithdrawals(blkWithdrawals gethtypes.Withdrawals) []*types.Eth1Withdrawal {
	withdrawals := make([]*types.Eth1Withdrawal, len(blkWithdrawals))
	for i, withdrawal := range blkWithdrawals {
		withdrawals[i] = &types.Eth1Withdrawal{
			Index:          withdrawal.Index,
			ValidatorIndex: withdrawal.Validator,
			Address:        withdrawal.Address.Bytes(),
			Amount:         new(big.Int).SetUint64(withdrawal.Amount).Bytes(),
		}
	}
	return withdrawals
}

func getLogsFromReceipts(logs []*gethtypes.Log) []*types.Eth1Log {
	eth1Logs := make([]*types.Eth1Log, len(logs))
	for i, log := range logs {
		topics := make([][]byte, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Bytes()
		}
		eth1Logs[i] = &types.Eth1Log{
			Address: log.Address.Bytes(),
			Data:    log.Data,
			Removed: log.Removed,
			Topics:  topics,
		}
	}
	return eth1Logs
}

func getContractSymbol(contract *contracts.IERC20Metadata, ret *types.ERC20Metadata) error {
	symbol, err := contract.Symbol(nil)
	if err != nil {
		if strings.Contains(err.Error(), "abi") {
			ret.Symbol = "UNKNOWN"
			return nil
		}

		return fmt.Errorf("error retrieving symbol: %w", err)
	}

	ret.Symbol = symbol
	return nil
}

func getContractTotalSupply(contract *contracts.IERC20Metadata, ret *types.ERC20Metadata) error {
	totalSupply, err := contract.TotalSupply(nil)
	if err != nil {
		return fmt.Errorf("error retrieving total supply: %w", err)
	}
	ret.TotalSupply = totalSupply.Bytes()
	return nil
}

func getContractDecimals(contract *contracts.IERC20Metadata, ret *types.ERC20Metadata) error {
	decimals, err := contract.Decimals(nil)
	if err != nil {
		return fmt.Errorf("error retrieving decimals: %w", err)
	}
	ret.Decimals = big.NewInt(int64(decimals)).Bytes()
	return nil
}

func getRateFromOracle(oracle *oneinchoracle.OneinchOracle, token []byte, ret *types.ERC20Metadata) error {
	rate, err := oracle.GetRateToEth(nil, common.BytesToAddress(token), false)
	if err != nil {
		return fmt.Errorf("error calling oneinchoracle.GetRateToEth: %w", err)
	}
	ret.Price = rate.Bytes()
	return nil
}

func parseAddressBalance(tokens []common.Address, address string, balances []*big.Int) []*types.Eth1AddressBalance {
	res := make([]*types.Eth1AddressBalance, len(tokens))
	for tokenIdx := range tokens {
		res[tokenIdx] = &types.Eth1AddressBalance{
			Address: common.FromHex(address),
			Token:   common.FromHex(string(tokens[tokenIdx].Bytes())),
			Balance: balances[tokenIdx].Bytes(),
		}
	}
	return res
}
