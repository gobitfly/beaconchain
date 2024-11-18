package data

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

const (
	maxInt                       = 9223372036854775807
	maxExecutionLayerBlockNumber = 1000000000

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

func reversedPaddedBlockNumber(blockNumber uint64) string {
	return fmt.Sprintf("%09d", maxExecutionLayerBlockNumber-blockNumber)
}

func keyTx(chainID string, hash []byte) string {
	format := "<chainID>:TX:<hash>"
	replacer := strings.NewReplacer("<chainID>", chainID, "<hash>", fmt.Sprintf("%x", hash))
	return replacer.Replace(format)
}

func keyTxSent(chainID string, tx *types.Eth1TransactionIndexed, index int) string {
	format := "<chainID>:I:TX:<from>:TO:<to>:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<from>", fmt.Sprintf("%x", tx.From),
		"<to>", fmt.Sprintf("%x", tx.To),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<time>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxReceived(chainID string, tx *types.Eth1TransactionIndexed, index int) string {
	format := "<chainID>:I:TX:<to>:FROM:<from>:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<to>", fmt.Sprintf("%x", tx.To),
		"<from>", fmt.Sprintf("%x", tx.From),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxTime(chainID string, tx *types.Eth1TransactionIndexed, address []byte, index int) string {
	format := "<chainID>:I:TX:<address>:TIME:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxBlock(chainID string, tx *types.Eth1TransactionIndexed, address []byte, index int) string {
	format := "<chainID>:I:TX:<address>:BLOCK:<block>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<block>", reversedPaddedBlockNumber(tx.BlockNumber),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxMethod(chainID string, tx *types.Eth1TransactionIndexed, address []byte, index int) string {
	format := "<chainID>:I:TX:<address>:METHOD:<method>:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<method>", fmt.Sprintf("%x", tx.MethodId),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxError(chainID string, tx *types.Eth1TransactionIndexed, address []byte, index int) string {
	format := "<chainID>:I:TX:<address>:ERROR:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyTxContractCreation(chainID string, tx *types.Eth1TransactionIndexed, address []byte, index int) string {
	format := "<chainID>:I:TX:<address>:CONTRACT:<time>:<index>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	return replacer.Replace(format)
}

func keyERC20(chainID string, hash []byte, logIndex int) string {
	format := "<chainID>:ERC20:<hash>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", hash),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20Time(chainID string, tx *types.Eth1ERC20Indexed, address []byte, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<address>:TIME:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<address>", fmt.Sprintf("%x", address),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20ContractAllTime(chainID string, tx *types.Eth1ERC20Indexed, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<contract>:ALL:TIME:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<contract>", fmt.Sprintf("%x", tx.TokenAddress),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20ContractTime(chainID string, tx *types.Eth1ERC20Indexed, address []byte, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<contract>:<address>:TIME:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<contract>", fmt.Sprintf("%x", tx.TokenAddress),
		"<address>", fmt.Sprintf("%x", address),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20To(chainID string, tx *types.Eth1ERC20Indexed, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<from>:TO:<to>:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<from>", fmt.Sprintf("%x", tx.From),
		"<to>", fmt.Sprintf("%x", tx.To),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20From(chainID string, tx *types.Eth1ERC20Indexed, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<to>:FROM:<from>:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<to>", fmt.Sprintf("%x", tx.To),
		"<from>", fmt.Sprintf("%x", tx.From),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20Sent(chainID string, tx *types.Eth1ERC20Indexed, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<from>:TOKEN_SENT:<contract>:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<from>", fmt.Sprintf("%x", tx.From),
		"<contract>", fmt.Sprintf("%x", tx.TokenAddress),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}

func keyERC20Received(chainID string, tx *types.Eth1ERC20Indexed, txIndex int, logIndex int) string {
	format := "<chainID>:I:ERC20:<to>:TOKEN_RECEIVED:<contract>:<time>:<index>:<logIndex>"
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<to>", fmt.Sprintf("%x", tx.To),
		"<contract>", fmt.Sprintf("%x", tx.TokenAddress),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	return replacer.Replace(format)
}
