package data

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

const (
	maxInt                       = 9223372036854775807
	maxExecutionLayerBlockNumber = 1000000000

	txPerBlockLimit = 10_000
	withdrawalLimit = 9999999999999
	logPerTxLimit   = 100_000
	itxPerTxLimit   = 100_000
	maxUncle        = 10
)

func ERC20TransfersToItems(chainID string, transfers []TransferWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, transfer := range transfers {
		if transfer.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, transfer.TxIndex, transfer.Indexed.ParentHash)
		}
		if transfer.LogIndex > logPerTxLimit {
			return nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v tx: %x", logPerTxLimit, transfer.LogIndex, transfer.Indexed.ParentHash)
		}
		b, err := proto.Marshal(transfer.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keyERC20(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)
		key := keys[0]

		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}
	return items, nil
}

func TransactionsToItems(chainID string, transactions []*types.Eth1TransactionIndexed) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for i, transaction := range transactions {
		if i > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, i, transaction.GetHash())
		}
		b, err := proto.Marshal(transaction)
		if err != nil {
			return nil, err
		}

		keys := keyTx(chainID, transaction, i)
		key := keys[0]

		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}
	return items, nil
}

func BlockToItems(chainID string, block *types.Eth1BlockIndexed) (map[string][]database.Item, error) {
	if block == nil {
		return nil, nil
	}
	items := make(map[string][]database.Item)
	b, err := proto.Marshal(block)
	if err != nil {
		return nil, err
	}

	keys := keyBlock(chainID, block)
	key := keys[0]

	items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
	item := []database.Item{{Family: defaultFamily, Column: key}}
	for i := 1; i < len(keys); i++ {
		items[keys[i]] = item
	}

	return items, nil
}

type BlobWithIndex struct {
	Indexed *types.Eth1BlobTransactionIndexed
	TxIndex int
}

func BlobToItems(chainID string, blobs []BlobWithIndex) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, blob := range blobs {
		if blob.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, blob.TxIndex, blob.Indexed.GetHash())
		}
		b, err := proto.Marshal(blob.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keysBlob(chainID, blob.Indexed, blob.TxIndex)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

type InternalWithIndexes struct {
	Indexed       *types.Eth1InternalTransactionIndexed
	TxIndex       int
	InternalIndex int
	Path          string
}

func InternalsToItems(chainID string, internals []InternalWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, internal := range internals {
		if internal.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, internal.TxIndex, internal.Indexed.ParentHash)
		}
		if internal.InternalIndex > itxPerTxLimit {
			return nil, fmt.Errorf("unexpected number of internal transactions in block expected at most %d but got: %v, tx: %x", itxPerTxLimit, internal.InternalIndex, internal.Indexed.ParentHash)
		}

		b, err := proto.Marshal(internal.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keysInternalTransaction(chainID, internal.Indexed, internal.TxIndex, internal.InternalIndex)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

type ERC1155TransferWithIndexes struct {
	Indexed  *types.ETh1ERC1155Indexed
	TxIndex  int
	LogIndex int
}

func ERC1155TransfersToItems(chainID string, transfers []ERC1155TransferWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, transfer := range transfers {
		if transfer.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, transfer.TxIndex, transfer.Indexed.ParentHash)
		}
		if transfer.LogIndex > logPerTxLimit {
			return nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v tx: %x", logPerTxLimit, transfer.LogIndex, transfer.Indexed.ParentHash)
		}

		b, err := proto.Marshal(transfer.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keys1155(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

type ERC721TransferWithIndexes struct {
	Indexed  *types.Eth1ERC721Indexed
	TxIndex  int
	LogIndex int
}

func ERC721TransfersToItems(chainID string, transfers []ERC721TransferWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, transfer := range transfers {
		if transfer.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, tx: %x", txPerBlockLimit, transfer.TxIndex, transfer.Indexed.ParentHash)
		}
		if transfer.LogIndex > logPerTxLimit {
			return nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v tx: %x", logPerTxLimit, transfer.LogIndex, transfer.Indexed.ParentHash)
		}

		b, err := proto.Marshal(transfer.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keys721(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

type UncleWithIndexes struct {
	Indexed   *types.Eth1UncleIndexed
	Index     int
	Coinbase  []byte
	BlockTime *timestamppb.Timestamp
}

func UnclesToItems(chainID string, uncles []UncleWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	if len(uncles) > maxUncle {
		return nil, fmt.Errorf("unexpected number of uncles in block expected at most %d but got: %v", maxUncle, len(uncles))
	}
	for _, uncle := range uncles {
		b, err := proto.Marshal(uncle.Indexed)
		if err != nil {
			return nil, err
		}

		keys := keysUncle(chainID, uncle.Indexed, uncle.Index, uncle.Coinbase, uncle.BlockTime)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

func WithdrawalToItems(chainID string, withdrawals []*types.Eth1WithdrawalIndexed) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, withdrawal := range withdrawals {
		b, err := proto.Marshal(withdrawal)
		if err != nil {
			return nil, err
		}

		keys := keysWithdrawal(chainID, withdrawal)
		key := keys[0]
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		item := []database.Item{{Family: defaultFamily, Column: key}}
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = item
		}
	}

	return items, nil
}

type ENSLog struct {
	Node  *[32]byte
	Name  *string
	Owner *common.Address
}

func ENSToItems(chainID string, logs []ENSLog) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, withdrawal := range logs {
		keys := keysENS(chainID, withdrawal)
		for i := 1; i < len(keys); i++ {
			items[keys[i]] = []database.Item{{Family: defaultFamily, Column: keys[i]}}
		}
	}
	return items, nil
}

func reversePaddedIndex(i int, maxValue int) string {
	if i > maxValue {
		log.Fatal(nil, fmt.Sprintf("padded index %v is greater than the max index of %v", i, maxValue), 0)
	}
	// TODO bug here
	// -1 means that the result will be 100, 99, 01
	// meanings index 0 (100) will be placed before index 1 (99)
	// it will be index 1 (99), ..., index 90 (10), index 0 (100), index 91 (09)
	// see TestReversePaddedIndex for more details
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

func keyTx(chainID string, tx *types.Eth1TransactionIndexed, index int) []string {
	formats := []string{
		"<chainID>:TX:<hash>",

		"<chainID>:I:TX:<from>:TO:<to>:<time>:<index>",
		"<chainID>:I:TX:<to>:FROM:<from>:<time>:<index>",

		"<chainID>:I:TX:<from>:TIME:<time>:<index>",
		"<chainID>:I:TX:<from>:BLOCK:<block>:<index>",
		"<chainID>:I:TX:<from>:METHOD:<method>:<time>:<index>",

		"<chainID>:I:TX:<to>:TIME:<time>:<index>",
		"<chainID>:I:TX:<to>:BLOCK:<block>:<index>",
		"<chainID>:I:TX:<to>:METHOD:<method>:<time>:<index>",
	}
	formatsErr := []string{
		"<chainID>:I:TX:<from>:ERROR:<time>:<index>",
		"<chainID>:I:TX:<to>:ERROR:<time>:<index>",
	}
	formatsContracts := []string{
		"<chainID>:I:TX:<from>:CONTRACT:<time>:<index>",
		"<chainID>:I:TX:<to>:CONTRACT:<time>:<index>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", tx.Hash),
		"<from>", fmt.Sprintf("%x", tx.From),
		"<to>", fmt.Sprintf("%x", tx.To),
		"<block>", reversedPaddedBlockNumber(tx.BlockNumber),
		"<method>", fmt.Sprintf("%x", tx.MethodId),
		"<time>", reversePaddedTimestamp(tx.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	if tx.ErrorMsg != "" {
		for _, format := range formatsErr {
			results = append(results, replacer.Replace(format))
		}
	}
	if tx.IsContractCreation {
		for _, format := range formatsContracts {
			results = append(results, replacer.Replace(format))
		}
	}
	return results
}

func keyERC20(chainID string, transfer *types.Eth1ERC20Indexed, txIndex int, logIndex int) []string {
	formats := []string{
		"<chainID>:ERC20:<hash>:<logIndex>",

		"<chainID>:I:ERC20:<contract>:ALL:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC20:<from>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC20:<contract>:<from>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC20:<to>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC20:<contract>:<to>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC20:<from>:TO:<to>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC20:<to>:FROM:<from>:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC20:<from>:TOKEN_SENT:<contract>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC20:<to>:TOKEN_RECEIVED:<contract>:<time>:<index>:<logIndex>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", transfer.ParentHash),
		"<contract>", fmt.Sprintf("%x", transfer.TokenAddress),
		"<from>", fmt.Sprintf("%x", transfer.From),
		"<to>", fmt.Sprintf("%x", transfer.To),
		"<index>", reversePaddedIndex(txIndex, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
		"<time>", reversePaddedTimestamp(transfer.Time),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keyBlock(chainID string, block *types.Eth1BlockIndexed) []string {
	formats := []string{
		"<chainID>:B:<block>",
		"<chainID>:I:B:<coinbase>:TIME:<time>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<coinbase>", fmt.Sprintf("%x", block.Coinbase),
		"<block>", reversedPaddedBlockNumber(block.Number),
		"<time>", reversePaddedTimestamp(block.Time),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keysBlob(chainID string, blob *types.Eth1BlobTransactionIndexed, index int) []string {
	formats := []string{
		"<chainID>:BTX:<hash>",

		"<chainID>:I:BTX:<from>:TO:<to>:<time>:<index>",
		"<chainID>:I:BTX:<from>:TIME:<time>:<index>",
		"<chainID>:I:BTX:<from>:BLOCK:<block>:<index>",

		"<chainID>:I:BTX:<to>:FROM:<from>:<time>:<index>",
		"<chainID>:I:BTX:<to>:TIME:<time>:<index>",
		"<chainID>:I:BTX:<to>:BLOCK:<block>:<index>",
	}
	formatsErr := []string{
		"<chainID>:I:BTX:<from>:ERROR:<time>:<index>",
		"<chainID>:I:BTX:<to>:ERROR:<time>:<index>",
	}

	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", blob.Hash),
		"<from>", fmt.Sprintf("%x", blob.From),
		"<to>", fmt.Sprintf("%x", blob.To),
		"<block>", reversedPaddedBlockNumber(blob.BlockNumber),
		"<time>", reversePaddedTimestamp(blob.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	if blob.ErrorMsg != "" {
		for _, format := range formatsErr {
			results = append(results, replacer.Replace(format))
		}
	}
	return results
}

func keysInternalTransaction(chainID string, internal *types.Eth1InternalTransactionIndexed, index int, internalIndex int) []string {
	formats := []string{
		"<chainID>:ITX:<hash>:<internalIndex>",

		"<chainID>:I:ITX:<from>:TO:<to>:<time>:<index>:<internalIndex>",
		"<chainID>:I:ITX:<to>:FROM:<from>:<time>:<index>:<internalIndex>",
		"<chainID>:I:ITX:<from>:TIME:<time>:<index>:<internalIndex>",
		"<chainID>:I:ITX:<to>:TIME:<time>:<index>:<internalIndex>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", internal.ParentHash),
		"<from>", fmt.Sprintf("%x", internal.From),
		"<to>", fmt.Sprintf("%x", internal.To),
		"<time>", reversePaddedTimestamp(internal.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
		"<internalIndex>", reversePaddedIndex(internalIndex, itxPerTxLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keys1155(chainID string, internal *types.ETh1ERC1155Indexed, index int, logIndex int) []string {
	formats := []string{
		"<chainID>:ERC1155:<hash>:<logIndex>",

		"<chainID>:I:ERC1155:<from>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC1155:<to>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC1155:<token>:ALL:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC1155:<token>:<from>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC1155:<token>:<to>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC1155:<from>:TO:<to>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC1155:<to>:FROM:<from>:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC1155:<from>:TOKEN_SENT:<token>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC1155:<to>:TOKEN_RECEIVED:<token>:<time>:<index>:<logIndex>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", internal.ParentHash),
		"<from>", fmt.Sprintf("%x", internal.From),
		"<to>", fmt.Sprintf("%x", internal.To),
		"<token>", fmt.Sprintf("%x", internal.TokenAddress),
		"<time>", reversePaddedTimestamp(internal.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keys721(chainID string, internal *types.Eth1ERC721Indexed, index int, logIndex int) []string {
	formats := []string{
		"<chainID>:ERC721:<hash>:<logIndex>",

		"<chainID>:I:ERC721:<from>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC721:<to>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC721:<token>:ALL:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC721:<token>:<from>:TIME:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC721:<token>:<to>:TIME:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC721:<from>:TO:<to>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC721:<to>:FROM:<from>:<time>:<index>:<logIndex>",

		"<chainID>:I:ERC721:<from>:TOKEN_SENT:<token>:<time>:<index>:<logIndex>",
		"<chainID>:I:ERC721:<to>:TOKEN_RECEIVED:<token>:<time>:<index>:<logIndex>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<hash>", fmt.Sprintf("%x", internal.ParentHash),
		"<from>", fmt.Sprintf("%x", internal.From),
		"<to>", fmt.Sprintf("%x", internal.To),
		"<token>", fmt.Sprintf("%x", internal.TokenAddress),
		"<time>", reversePaddedTimestamp(internal.Time),
		"<index>", reversePaddedIndex(index, txPerBlockLimit),
		"<logIndex>", reversePaddedIndex(logIndex, logPerTxLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keysUncle(chainID string, uncle *types.Eth1UncleIndexed, index int, coinbase []byte, blockTime *timestamppb.Timestamp) []string {
	formats := []string{
		"<chainID>:U:<block>:<index>",
		"<chainID>:I:U:<coinbase>:TIME:<time>:<index>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<block>", reversedPaddedBlockNumber(uncle.BlockNumber),
		"<coinbase>", fmt.Sprintf("%x", coinbase),
		"<time>", reversePaddedTimestamp(blockTime),
		"<index>", reversePaddedIndex(index, maxUncle),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keysWithdrawal(chainID string, withdrawal *types.Eth1WithdrawalIndexed) []string {
	formats := []string{
		"<chainID>:W:<block>:<index>",
		"<chainID>:I:W:<address>:TIME:<time>:<index>",
	}
	replacer := strings.NewReplacer(
		"<chainID>", chainID,
		"<block>", reversedPaddedBlockNumber(withdrawal.BlockNumber),
		"<address>", fmt.Sprintf("%x", withdrawal.Address),
		"<time>", reversePaddedTimestamp(withdrawal.Time),
		"<index>", reversePaddedIndex(int(withdrawal.Index), withdrawalLimit),
	)
	var results []string
	for _, format := range formats {
		results = append(results, replacer.Replace(format))
	}
	return results
}

func keysENS(chainID string, log ENSLog) []string {
	var keys []string
	if log.Node != nil {
		keys = append(keys, fmt.Sprintf("%s:ENS:V:H:%x", chainID, log.Node))
	}
	if log.Owner != nil {
		keys = append(keys, fmt.Sprintf("%s:ENS:V:A:%x", chainID, log.Owner))
	}
	if log.Name != nil {
		keys = append(keys, fmt.Sprintf("%s:ENS:V:N:%x", chainID, log.Name))
	}
	return keys
}
