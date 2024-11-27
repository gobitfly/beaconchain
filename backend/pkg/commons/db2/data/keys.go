package data

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

const (
	maxInt = 9223372036854775807
	// 	maxExecutionLayerBlockNumber = 1000000000

	txPerBlockLimit = 10_000
	logPerTxLimit   = 100_000
)

func reversePaddedIndex(i int, maxValue int) string {
	if i > maxValue {
		log.Fatal(nil, fmt.Sprintf("padded index %v is greater than the max index of %v", i, maxValue), 0)
	}
	// TODO probably a bug here
	// TODO -1 means that the result will be 100, 99, 01
	// TODO meanings index 0 (100) will be placed before index 1 (99)
	// TODO it will be index 1 (99), ..., index 90 (10), index 0 (100), index 91 (09)
	length := fmt.Sprintf("%d", len(fmt.Sprintf("%d", maxValue))-1)
	fmtStr := "%0" + length + "d"
	return fmt.Sprintf(fmtStr, maxValue-i)
}

func reversePaddedTimestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		log.Fatal(nil, fmt.Sprintf("unknown timestamp: %v", timestamp), 0)
	}
	return fmt.Sprintf("%019d", maxInt-timestamp.Seconds)
}

// commented for now but we will use it soon
// func reversedPaddedBlockNumber(blockNumber uint64) string {
//	return fmt.Sprintf("%09d", maxExecutionLayerBlockNumber-blockNumber)
//}

// lost keyTxBlock, keyTxError, keyTxContractCreation
// key are sorted side, with, chainID, type, method
func transactionKeys(chainID string, transaction *types.Eth1TransactionIndexed, index int) (string, []string) {
	main := "TX:<chainID>:<hash>"
	baseKeys := []string{
		"all:<address>",
		"all:with:<address>:<with>",
		"all:chainID:<address>:<chainID>",
		"all:with:chainID:<address>:<with>:<chainID>",
	}
	fromToKeys := []string{
		"in:<to>",
		"in:chainID:<to>:<chainID>",
		"in:with:<to>:<from>",
		"in:with:chainID:<to>:<from>:<chainID>",
		"out:<from>",
		"out:chainID:<from>:<chainID>",
		"out:with:<from>:<to>",
		"out:with:chainID:<from>:<to>:<chainID>",
	}
	baseTxKeys := []string{
		"all:TX:<address>",
		"all:TX:method:<address>:<method>",
		"all:chainID:TX:<address>:<chainID>",
		"all:chainID:TX:method:<address>:<chainID>:<method>",
		"all:with:TX:<address>:<with>",
		"all:with:TX:method:<address>:<with>",
		"all:with:chainID:TX:<address>:<with>:<chainID>",
		"all:with:chainID:TX:method:<address>:<with>:<chainID>",
	}
	fromToTxKeys := []string{
		"in:TX:<to>",
		"in:TX:method:<to>:<method>",
		"in:chainID:TX:<to>:<chainID>",
		"in:chainID:TX:method:<to>:<chainID>:<method>",
		"in:with:TX:<to>:<from>",
		"in:with:TX:method:<to>:<from>:<method>",
		"in:with:chainID:TX:<to>:<from>:<chainID>",
		"in:with:chainID:TX:method:<to>:<from>:<chainID>:<method>",
		"out:TX:<from>",
		"out:TX:method:<from>:<method>",
		"out:chainID:TX:<from>:<chainID>",
		"out:chainID:TX:method:<from>:<chainID>:<method>",
		"out:with:TX:<from>:<to>",
		"out:with:TX:method:<from>:<to>:<method>",
		"out:with:chainID:TX:<from>:<to>:<chainID>",
		"out:with:chainID:TX:method:<from>:<to>:<chainID>:<method>",
	}
	replacer := strings.NewReplacer(
		"<hash>", toHex(transaction.Hash),
		"<from>", toHex(transaction.From),
		"<to>", toHex(transaction.To),
		"<chainID>", chainID,
		"<time>", reversePaddedTimestamp(transaction.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
		"<method>", fmt.Sprintf("%x", transaction.MethodId),
	)
	keys := append(fromToKeys, fromToTxKeys...)
	for _, format := range append(baseKeys, baseTxKeys...) {
		keys = append(keys,
			strings.ReplaceAll(strings.ReplaceAll(format, "<address>", "<from>"), "<with>", "<to>"),
			strings.ReplaceAll(strings.ReplaceAll(format, "<address>", "<to>"), "<with>", "<from>"),
		)
	}
	id := ":<time>:<index>"
	for i := range keys {
		keys[i] = replacer.Replace(keys[i] + id)
	}

	return replacer.Replace(main), keys
}

// key are sorted side (+optional other address), chainID, type, asset
func transferKeys(chainID string, transaction *types.Eth1ERC20Indexed, index int, logIndex int) (string, []string) {
	main := "ERC20:<chainID>:<hash>"
	baseKeys := []string{
		"all:<address>",
		"all:chainID:<address>:<chainID>",
		"all:chainID:<address>:<chainID>",
		"all:with:chainID:<address>:<with>:<chainID>",
	}
	fromToKeys := []string{
		"in:<to>",
		"in:chainID:<to>:<chainID>",
		"in:with:<to>:<from>",
		"in:with:chainID:<to>:<from>:<chainID>",
		"out:<from>",
		"out:chainID:<from>:<chainID>",
		"out:with:<from>:<to>",
		"out:with:chainID:<from>:<to>:<chainID>",
	}
	baseTxKeys := []string{
		"all:ERC20:<address>",
		"all:ERC20:asset:<address>:<asset>",
		"all:chainID:ERC20:<address>:<chainID>",
		"all:chainID:ERC20:asset:<address>:<chainID>:<asset>",
		"all:with:ERC20:<address>:<with>",
		"all:with:ERC20:method:<address>:<with>",
		"all:with:chainID:ERC20:<address>:<with>:<chainID>",
		"all:with:chainID:ERC20:method:<address>:<with>:<chainID>",
	}
	fromToTxKeys := []string{
		"in:ERC20:<to>",
		"in:ERC20:asset:<to>:<asset>",
		"in:chainID:ERC20:<to>:<chainID>",
		"in:chainID:ERC20:asset:<to>:<chainID>:<asset>",
		"in:with:ERC20:<to>:<from>",
		"in:with:ERC20:asset:<to>:<from>:<asset>",
		"in:with:chainID:ERC20:<to>:<from>:<chainID>",
		"in:with:chainID:ERC20:asset:<to>:<from>:<chainID>:<asset>",
		"out:ERC20:<from>",
		"out:ERC20:asset:<from>:<asset>",
		"out:chainID:ERC20:<from>:<chainID>",
		"out:chainID:ERC20:asset:<from>:<chainID>:<asset>",
		"out:with:ERC20:<from>:<to>",
		"out:with:ERC20:asset:<from>:<to>:<asset>",
		"out:with:chainID:ERC20:<from>:<to>:<chainID>",
		"out:with:chainID:ERC20:asset:<from>:<to>:<chainID>:<asset>",
	}
	replacer := strings.NewReplacer(
		"<hash>", toHex(transaction.ParentHash),
		"<from>", toHex(transaction.From),
		"<to>", toHex(transaction.To),
		"<chainID>", chainID,
		"<time>", reversePaddedTimestamp(transaction.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
		"<asset>", toHex(transaction.TokenAddress),
	)
	keys := append(fromToKeys, fromToTxKeys...)
	for _, format := range append(baseKeys, baseTxKeys...) {
		keys = append(keys,
			strings.ReplaceAll(strings.ReplaceAll(format, "<address>", "<from>"), "<with>", "<to>"),
			strings.ReplaceAll(strings.ReplaceAll(format, "<address>", "<to>"), "<with>", "<from>"),
		)
	}
	id := ":<time>:<index>:<logIndex>"
	for i := range keys {
		keys[i] = replacer.Replace(keys[i] + id)
	}

	return replacer.Replace(main), keys
}
