package executionlayer

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/contracts/ens"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/commons/erc1155"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

// TransformFunc describes a function that will index a specific object from the types.Eth1Block
// example: transaction, ERC20 transfer, block, ...
// It will put the indexed result into the res *IndexedBlock param
// This way all transform functions have the same signature
type TransformFunc func(chainID string, block *types.Eth1Block, res *IndexedBlock) error

var Transformers = map[string]TransformFunc{
	"TransformTx":                TransformTx,
	"TransformERC20":             TransformERC20,
	"TransformBlock":             TransformBlock,
	"TransformBlobTx":            TransformBlob,
	"TransformContract":          TransformContract,
	"TransformItx":               TransformITx,
	"TransformERC721":            TransformERC721,
	"TransformERC1155":           TransformERC1155,
	"TransformUncle":             TransformUncle,
	"TransformWithdrawals":       TransformWithdrawal,
	"TransformEnsNameRegistered": TransformEnsNameRegistered,
}

var AllTransformers = maps.Values(Transformers)

func TransformTx(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var transactions []*types.Eth1TransactionIndexed
	for _, tx := range blk.Transactions {

		to, isContract := getTxRecipient(tx)
		method := getMethodSignature(tx)
		fee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetGasPrice()), big.NewInt(int64(tx.GetGasUsed()))).Bytes()
		blobFee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetBlobGasPrice()), big.NewInt(int64(tx.GetBlobGasUsed()))).Bytes()
		indexedTx := &types.Eth1TransactionIndexed{
			Hash:               tx.GetHash(),
			BlockNumber:        block.GetNumber(),
			Time:               block.GetTime(),
			MethodId:           method,
			From:               tx.GetFrom(),
			To:                 to,
			Value:              tx.GetValue(),
			TxFee:              fee,
			GasPrice:           tx.GetGasPrice(),
			IsContractCreation: isContract,
			ErrorMsg:           "",
			BlobTxFee:          blobFee,
			BlobGasPrice:       tx.GetBlobGasPrice(),
			Status:             types.StatusType(tx.Status),
		}

		updateITxStatus(indexedTx, tx.Itx)
		transactions = append(transactions, indexedTx)
	}
	res.Transactions = transactions
	return nil
}

func TransformERC20(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	filterer, err := contracts.NewERC20Filterer(common.Address{}, nil)
	if err != nil {
		return errors.Wrap(err, "cannot ERC20 create filterer")
	}
	var transfers []data.TransferWithIndexes
	for txIndex, tx := range block.GetTransactions() {
		for logIndex, log := range tx.GetLogs() {
			if !isValidERC20Log(log) {
				continue
			}
			topics := getLogTopics(log)

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: block.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(txIndex),
				BlockHash:   common.BytesToHash(block.GetHash()),
				Index:       uint(logIndex),
				Removed:     log.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			value := getERC20TransferValue(transfer)

			indexedLog := &types.Eth1ERC20Indexed{
				ParentHash:   tx.GetHash(),
				BlockNumber:  block.GetNumber(),
				Time:         block.GetTime(),
				TokenAddress: log.Address,
				From:         transfer.From.Bytes(),
				To:           transfer.To.Bytes(),
				Value:        value,
			}
			transfers = append(transfers, data.TransferWithIndexes{
				Indexed:  indexedLog,
				TxIndex:  txIndex,
				LogIndex: logIndex,
			})
		}
	}
	res.ERC20Transfer = transfers
	return nil
}

func TransformBlock(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	idx := types.Eth1BlockIndexed{
		Hash:       block.GetHash(),
		ParentHash: block.GetParentHash(),
		UncleHash:  block.GetUncleHash(),
		Coinbase:   block.GetCoinbase(),
		Difficulty: block.GetDifficulty(),
		Number:     block.GetNumber(),
		GasLimit:   block.GetGasLimit(),
		GasUsed:    block.GetGasUsed(),
		Time:       block.GetTime(),
		BaseFee:    block.GetBaseFee(),
		// Duration:               uint64(block.GetTime().AsTime().Unix() - previous.GetTime().AsTime().Unix()),
		UncleCount:       uint64(len(block.GetUncles())),
		TransactionCount: uint64(len(block.GetTransactions())),
		// BaseFeeChange:          new(big.Int).Sub(new(big.Int).SetBytes(block.GetBaseFee()), new(big.Int).SetBytes(previous.GetBaseFee())).Bytes(),
		// BlockUtilizationChange: new(big.Int).Sub(new(big.Int).Div(big.NewInt(int64(block.GetGasUsed())), big.NewInt(int64(block.GetGasLimit()))), new(big.Int).Div(big.NewInt(int64(previous.GetGasUsed())), big.NewInt(int64(previous.GetGasLimit())))).Bytes(),
		BlobGasUsed:   block.GetBlobGasUsed(),
		ExcessBlobGas: block.GetExcessBlobGas(),
	}

	blockUncleReward := calculateBlockUncleReward(block, chainID)
	idx.UncleReward = blockUncleReward.Bytes()

	var maxGasPrice *big.Int
	var minGasPrice *big.Int
	txReward := big.NewInt(0)

	for _, t := range block.GetTransactions() {
		price := new(big.Int).SetBytes(t.GasPrice)

		if minGasPrice == nil {
			minGasPrice = price
		}
		if maxGasPrice == nil {
			maxGasPrice = price
		}

		if price.Cmp(maxGasPrice) > 0 {
			maxGasPrice = price
		}

		if price.Cmp(minGasPrice) < 0 {
			minGasPrice = price
		}

		txFee := calculateTxFee(t, block.BaseFee)
		txReward.Add(txReward, txFee)

		for _, itx := range t.Itx {
			if !isValidItx(itx) {
				continue
			}
			idx.InternalTransactionCount++
		}

		if isBlobTx(t.GetType()) {
			idx.BlobTransactionCount++
		}
	}

	idx.TxReward = txReward.Bytes()

	if maxGasPrice != nil {
		idx.LowestGasPrice = minGasPrice.Bytes()
	}
	if minGasPrice != nil {
		idx.HighestGasPrice = maxGasPrice.Bytes()
	}

	idx.Mev = calculateMevFromBlock(block).Bytes() // deprecated but we still write the value to keep all blocks consistent
	res.Block = &idx
	return nil
}

func TransformBlob(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var blobs []data.BlobWithIndex
	for i, tx := range block.Transactions {
		if !isBlobTx(tx.Type) {
			// skip non blob-txs
			continue
		}
		fee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetGasPrice()), big.NewInt(int64(tx.GetGasUsed()))).Bytes()
		blobFee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetBlobGasPrice()), big.NewInt(int64(tx.GetBlobGasUsed()))).Bytes()
		indexedTx := &types.Eth1BlobTransactionIndexed{
			Hash:                tx.GetHash(),
			BlockNumber:         block.GetNumber(),
			Time:                block.GetTime(),
			From:                tx.GetFrom(),
			To:                  tx.GetTo(),
			Value:               tx.GetValue(),
			TxFee:               fee,
			GasPrice:            tx.GetGasPrice(),
			BlobTxFee:           blobFee,
			BlobGasPrice:        tx.GetBlobGasPrice(),
			ErrorMsg:            tx.GetErrorMsg(),
			BlobVersionedHashes: tx.GetBlobVersionedHashes(),
		}
		blobs = append(blobs, data.BlobWithIndex{
			Indexed: indexedTx,
			TxIndex: i,
		})
	}
	res.Blobs = blobs
	return nil
}

func TransformContract(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var contracts []metadataupdates.ContractUpdateWithAddress
	for i, tx := range block.GetTransactions() {
		for j, itx := range tx.GetItx() {
			if itx.GetType() == "create" || itx.GetType() == "suicide" {
				contractUpdate := &types.IsContractUpdate{
					IsContract: itx.GetType() == "create",
					// also use success status of enclosing transaction, as even successful sub-calls can still be reverted later in the tx
					Success: itx.GetErrorMsg() == "" && tx.GetErrorMsg() == "",
				}
				address := getContractAddress(itx)
				ts, err := encodeIsContractUpdateTs(block.GetNumber(), uint64(i), uint64(j))
				if err != nil {
					return nil, nil, fmt.Errorf("error generating bigtable isContract timestamp: %w", err)
				}
				updates = append(updates, metadataupdates.ContractUpdateWithAddress{
					Indexed:   contractUpdate,
					Address:   address,
					Timestamp: ts,
				})
			}
		}
	}
	res.Contracts = contracts
	return nil
}

func TransformITx(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var transactions []data.InternalWithIndexes
	for i, tx := range block.GetTransactions() {
		for j, itx := range tx.GetItx() {
			if !isValidItx(itx) {
				continue
			}

			indexed := &types.Eth1InternalTransactionIndexed{
				ParentHash:  tx.GetHash(),
				BlockNumber: block.GetNumber(),
				Time:        block.GetTime(),
				Type:        itx.GetType(),
				From:        itx.GetFrom(),
				To:          itx.GetTo(),
				Value:       itx.GetValue(),
			}
			if itx.GetType() == "delegatecall" {
				continue
			}
			transactions = append(transactions, data.InternalWithIndexes{
				Indexed:       indexed,
				TxIndex:       i,
				InternalIndex: j,
				Path:          itx.Path,
			})
		}
	}
	res.Internals = transactions
	return nil
}

func TransformERC1155(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	filterer, err := contracts.NewERC1155Filterer(common.Address{}, nil)
	if err != nil {
		return errors.Wrap(err, "cannot ERC1155 create filterer")
	}
	var transfers []data.ERC1155TransferWithIndexes
	for txIndex, tx := range block.GetTransactions() {
		for logIndex, log := range tx.GetLogs() {
			if !isValidERC1155Log(log) {
				continue
			}

			topics := getLogTopics(log)

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: block.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(txIndex),
				BlockHash:   common.BytesToHash(block.GetHash()),
				Index:       uint(logIndex),
				Removed:     log.GetRemoved(),
			}

			transferBatch, _ := filterer.ParseTransferBatch(ethLog)
			transferSingle, _ := filterer.ParseTransferSingle(ethLog)
			if transferBatch == nil && transferSingle == nil {
				continue
			}
			indexedLog := &types.ETh1ERC1155Indexed{}
			if transferBatch != nil {
				ids := getERC1155TransferIDs(transferBatch.Ids)
				values := getERC1155TransferValues(transferBatch.Values)

				// TODO - Tangui this is probably a bug, only the last transfer will be saved
				for ti := range ids {
					indexedLog.BlockNumber = block.GetNumber()
					indexedLog.Time = block.GetTime()
					indexedLog.ParentHash = tx.GetHash()
					indexedLog.From = transferBatch.From.Bytes()
					indexedLog.To = transferBatch.To.Bytes()
					indexedLog.Operator = transferBatch.Operator.Bytes()
					indexedLog.TokenId = ids[ti]
					indexedLog.Value = values[ti]
					indexedLog.TokenAddress = log.GetAddress()
				}
			} else if transferSingle != nil {
				indexedLog.BlockNumber = block.GetNumber()
				indexedLog.Time = block.GetTime()
				indexedLog.ParentHash = tx.GetHash()
				indexedLog.From = transferSingle.From.Bytes()
				indexedLog.To = transferSingle.To.Bytes()
				indexedLog.Operator = transferSingle.Operator.Bytes()
				indexedLog.TokenId = transferSingle.Id.Bytes()
				indexedLog.Value = transferSingle.Value.Bytes()
				indexedLog.TokenAddress = log.GetAddress()
			}
			transfers = append(transfers, data.ERC1155TransferWithIndexes{
				Indexed:  indexedLog,
				TxIndex:  txIndex,
				LogIndex: logIndex,
			})
		}
	}
	res.ERC1155Transfer = transfers
	return nil
}

func TransformERC721(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	filterer, err := contracts.NewERC721Filterer(common.Address{}, nil)
	if err != nil {
		return errors.Wrap(err, "cannot ER721 create filterer")
	}
	var transfers []data.ERC721TransferWithIndexes
	for txIndex, tx := range block.GetTransactions() {
		for logIndex, log := range tx.GetLogs() {
			if !isValidERC721Log(log) {
				continue
			}

			topics := getLogTopics(log)

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: block.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(txIndex),
				BlockHash:   common.BytesToHash(block.GetHash()),
				Index:       uint(logIndex),
				Removed:     log.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			tokenId := getTokenID(transfer)

			indexedLog := &types.Eth1ERC721Indexed{
				ParentHash:   tx.GetHash(),
				BlockNumber:  block.GetNumber(),
				Time:         block.GetTime(),
				TokenAddress: log.Address,
				From:         transfer.From.Bytes(),
				To:           transfer.To.Bytes(),
				TokenId:      tokenId.Bytes(),
			}
			transfers = append(transfers, data.ERC721TransferWithIndexes{
				Indexed:  indexedLog,
				TxIndex:  txIndex,
				LogIndex: logIndex,
			})
		}
	}
	res.ERC721Transfer = transfers
	return nil
}

func TransformUncle(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var uncles []data.UncleWithIndexes
	for i, uncle := range block.Uncles {
		reward := calculateUncleReward(block, uncle, chainID)

		uncleIndexed := types.Eth1UncleIndexed{
			Number:      uncle.GetNumber(),
			BlockNumber: block.GetNumber(),
			GasLimit:    uncle.GetGasLimit(),
			GasUsed:     uncle.GetGasUsed(),
			BaseFee:     uncle.GetBaseFee(),
			Difficulty:  uncle.GetDifficulty(),
			Time:        uncle.GetTime(),
			Reward:      reward.Bytes(),
		}
		uncles = append(uncles, data.UncleWithIndexes{
			Indexed:   &uncleIndexed,
			Index:     i,
			Coinbase:  uncle.GetCoinbase(),
			BlockTime: block.GetTime(),
		})
	}
	res.Uncles = uncles
	return nil
}

func TransformWithdrawal(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	var withdrawals []*types.Eth1WithdrawalIndexed
	for _, withdrawal := range block.Withdrawals {
		withdrawals = append(withdrawals, &types.Eth1WithdrawalIndexed{
			BlockNumber:    block.Number,
			Index:          withdrawal.Index,
			ValidatorIndex: withdrawal.ValidatorIndex,
			Address:        withdrawal.Address,
			Amount:         withdrawal.Amount,
			Time:           block.Time,
		})
	}
	res.Withdrawals = withdrawals
	return nil
}

func TransformEnsNameRegistered(chainID string, block *types.Eth1Block, res *IndexedBlock) error {
	ensContractAddresses := ens.ENSContractFor(chainID)
	var ensLogs []data.ENSLog
	for i, tx := range block.GetTransactions() {
		for j, txLog := range tx.GetLogs() {
			ensContract := ensContractAddresses[common.BytesToAddress(txLog.Address).String()]

			topics := getLogTopics(txLog)

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(txLog.GetAddress()),
				Topics:      topics,
				Data:        txLog.Data,
				BlockNumber: block.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(i),
				BlockHash:   common.BytesToHash(block.GetHash()),
				Index:       uint(j),
				Removed:     txLog.GetRemoved(),
			}
			var ensLog data.ENSLog
			// TODO there is probably a better way to do this
			for _, lTopic := range txLog.GetTopics() {
				switch ensContract {
				case "Registry":
					filterer, _ := ens.NewENSRegistryFilterer(common.Address{}, nil)
					switch {
					case bytes.Equal(lTopic, ens.RegistryNewResolverTopic.Bytes()):
						r, _ := filterer.ParseNewResolver(ethLog)
						ensLog.Node = &r.Node
					case bytes.Equal(lTopic, ens.RegistryNewOwnerTopic.Bytes()):
						r, _ := filterer.ParseNewOwner(ethLog)
						ensLog.Owner = &r.Owner
					case bytes.Equal(lTopic, ens.RegistryNewTTLTopic.Bytes()):
						r, _ := filterer.ParseNewTTL(ethLog)
						ensLog.Node = &r.Node
					}
				case "ETHRegistrarController":
					filterer, _ := ens.NewENSETHRegistrarControllerFilterer(common.Address{}, nil)
					switch {
					case bytes.Equal(lTopic, ens.RegistrarControllerNameRegisteredTopic.Bytes()):
						r, _ := filterer.ParseNameRegistered(ethLog)
						if err := verifyName(r.Name); err != nil {
							continue
						}
						ensLog.Name = &r.Name
						ensLog.Owner = &r.Owner
					case bytes.Equal(lTopic, ens.RegistrarControllerNameRenewedTopic.Bytes()):
						r, _ := filterer.ParseNameRenewed(ethLog)
						if err := verifyName(r.Name); err != nil {
							continue
						}
						ensLog.Name = &r.Name
					}
				case "OldEnsRegistrarController":
					filterer, _ := ens.NewENSOldRegistrarControllerFilterer(common.Address{}, nil)
					switch {
					case bytes.Equal(lTopic, ens.OldRegistrarControllerNameRegisteredTopic.Bytes()):
						r, _ := filterer.ParseNameRegistered(ethLog)
						if err := verifyName(r.Name); err != nil {
							continue
						}
						ensLog.Name = &r.Name
						ensLog.Owner = &r.Owner
					case bytes.Equal(lTopic, ens.OldRegistrarControllerNameRenewedTopic.Bytes()):
						r, _ := filterer.ParseNameRenewed(ethLog)
						if err := verifyName(r.Name); err != nil {
							continue
						}
						ensLog.Name = &r.Name
					}
				default:
					// try to parse public resolver events
					// in that case we are not sure that the log is emitted from a public resolver
					// checking the topic is not enough collision is common for events topic
					// TODO this might be a source of bug, someone could create a fake ens contract and emit the same
					// events and we will parse it like if it was the truth
					filterer, _ := ens.NewENSPublicResolver(common.Address{}, nil)
					switch {
					case bytes.Equal(lTopic, ens.PublicResolverNameChangedTopic.Bytes()):
						r, err := filterer.ParseNameChanged(ethLog)
						if err != nil {
							continue
						}
						if err := verifyName(r.Name); err != nil {
							continue
						}
						ensLog.Name = &r.Name
					case bytes.Equal(lTopic, ens.PublicResolverAddressChangedTopic.Bytes()):
						r, err := filterer.ParseAddressChanged(ethLog)
						if err != nil {
							continue
						}
						ensLog.Node = &r.Node
					}
				}
			}
			ensLogs = append(ensLogs, ensLog)
		}
	}
	res.ENS = ensLogs
	return nil
}

func TransformerFromList(names []string) ([]TransformFunc, error) {
	var transforms []TransformFunc
	for _, name := range names {
		transform, ok := Transformers[name]
		if !ok {
			return nil, fmt.Errorf("invalid transformer flag %v", name)
		}
		transforms = append(transforms, transform)
	}
	return transforms, nil
}

func eth1BlockReward(chainID string, blockNumber uint64, difficulty []byte) *big.Int {
	// no block rewards for PoS blocks
	// holesky genesis block has difficulty 1 and zero block reward (launched with pos)
	if len(difficulty) == 0 || (len(difficulty) == 1 && difficulty[0] == 1) {
		return big.NewInt(0)
	}

	el := elConfigForChainID(chainID)
	if blockNumber < el.ByzantiumBlock.Uint64() {
		return big.NewInt(5e+18)
	} else if blockNumber < el.ConstantinopleBlock.Uint64() {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}
}

func elConfigForChainID(chainID string) *params.ChainConfig {
	switch chainID {
	case "1":
		return params.MainnetChainConfig
	}
	return &params.ChainConfig{
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
	}
}

func verifyName(name string) error {
	// limited by max capacity of db (caused by btrees of indexes); tests showed maximum of 2684 (added buffer)
	if len(name) > 2048 {
		return fmt.Errorf("name too long: %v", name)
	}
	return nil
}

func isValidERC20Log(log *types.Eth1Log) bool {
	return len(log.GetTopics()) == 3 && bytes.Equal(log.GetTopics()[0], erc20.TransferTopic.Bytes())
}

func isValidERC721Log(log *types.Eth1Log) bool {
	return len(log.GetTopics()) == 4 && bytes.Equal(log.GetTopics()[0], erc721.TransferTopic.Bytes())
}

func isValidERC1155Log(log *types.Eth1Log) bool {
	if len(log.GetTopics()) != 4 {
		return false
	}

	// check if first topic matches TransferSingle or TransferBatch topic
	firstTopic := log.GetTopics()[0]
	isTransferBulk := bytes.Equal(firstTopic, erc1155.TransferBulkTopic.Bytes())
	isTransferSingle := bytes.Equal(firstTopic, erc1155.TransferSingleTopic.Bytes())

	return isTransferBulk || isTransferSingle
}

func isBlobTx(txType uint32) bool {
	return txType == gethtypes.BlobTxType
}

func isValidItx(itx *types.Eth1InternalTransaction) bool {
	// skip top level and empty calls
	if itx.Path == "[]" || itx.Path == "0" || bytes.Equal(itx.Value, []byte{0x0}) {
		return false
	}
	return true
}

func getLogTopics(log *types.Eth1Log) []common.Hash {
	topics := make([]common.Hash, 0, len(log.GetTopics()))
	for _, lTopic := range log.GetTopics() {
		topics = append(topics, common.BytesToHash(lTopic))
	}
	return topics
}

func getTxRecipient(tx *types.Eth1Transaction) ([]byte, bool) {
	to := tx.GetTo()
	isContract := false
	if tx.GetContractAddress() != nil && !bytes.Equal(tx.GetContractAddress(), common.Address{}.Bytes()) {
		to = tx.GetContractAddress()
		isContract = true
	}
	return to, isContract
}

func getTokenID(transfer *contracts.ERC721Transfer) *big.Int {
	tokenId := new(big.Int)
	if transfer.TokenId != nil {
		tokenId = transfer.TokenId
	}
	return tokenId
}

// extracts the first 4 bytes of the data field for method signature
func getMethodSignature(tx *types.Eth1Transaction) []byte {
	method := make([]byte, 0)
	if len(tx.GetData()) > 3 {
		method = tx.GetData()[:4]
	}
	return method
}

func getContractAddress(itx *types.Eth1InternalTransaction) []byte {
	address := itx.GetTo()
	if itx.GetType() == "suicide" {
		address = itx.GetFrom()
	}
	return address
}

func getERC20TransferValue(transfer *contracts.ERC20Transfer) []byte {
	var value []byte
	if transfer.Value != nil {
		value = transfer.Value.Bytes()
	}
	return value
}

func getERC1155TransferIDs(idList []*big.Int) [][]byte {
	ids := make([][]byte, 0, len(idList))
	for _, id := range idList {
		ids = append(ids, id.Bytes())
	}
	return ids
}

func getERC1155TransferValues(values []*big.Int) [][]byte {
	v := make([][]byte, 0, len(values))
	for _, val := range values {
		v = append(v, val.Bytes())
	}
	return v
}

func updateITxStatus(indexedTx *types.Eth1TransactionIndexed, internals []*types.Eth1InternalTransaction) {
	for _, itx := range internals {
		if itx.ErrorMsg != "" {
			indexedTx.ErrorMsg = itx.ErrorMsg
			if indexedTx.Status == types.StatusType_SUCCESS {
				indexedTx.Status = types.StatusType_PARTIAL
			}
			break
		}
	}
}

// calculates tx fee and priority fee
func calculateTxFee(t *types.Eth1Transaction, baseFee []byte) *big.Int {
	txFee := new(big.Int).Mul(new(big.Int).SetBytes(t.GasPrice), big.NewInt(int64(t.GasUsed)))

	if len(baseFee) > 0 {
		effectiveGasPrice := math.BigMin(new(big.Int).Add(new(big.Int).SetBytes(t.MaxPriorityFeePerGas), new(big.Int).SetBytes(baseFee)), new(big.Int).SetBytes(t.MaxFeePerGas))
		proposerGasPricePart := new(big.Int).Sub(effectiveGasPrice, new(big.Int).SetBytes(baseFee))

		if proposerGasPricePart.Cmp(big.NewInt(0)) >= 0 {
			txFee = new(big.Int).Mul(proposerGasPricePart, big.NewInt(int64(t.GasUsed)))
		} else {
			log.Error(fmt.Errorf("error minerGasPricePart is below 0 for tx %v: %v", t.Hash, proposerGasPricePart), "", 0)
			txFee = big.NewInt(0)
		}
	}
	return txFee
}

// calculates the total value of uncle rewards for a block
func calculateBlockUncleReward(block *types.Eth1Block, chainID string) *big.Int {
	uncleRewards := big.NewInt(0)
	reward := new(big.Int)

	for _, uncle := range block.Uncles {
		if len(block.Difficulty) == 0 { // no uncle rewards in PoS
			return uncleRewards // return 0
		}

		reward.Add(big.NewInt(int64(uncle.GetNumber())), big.NewInt(8))
		reward.Sub(reward, big.NewInt(int64(block.GetNumber())))
		reward.Mul(reward, eth1BlockReward(chainID, block.GetNumber(), block.Difficulty))
		reward.Div(reward, big.NewInt(8))

		reward.Div(eth1BlockReward(chainID, block.GetNumber(), block.Difficulty), big.NewInt(32))
		uncleRewards.Add(uncleRewards, reward)
	}

	return uncleRewards
}

// calculates the reward for a single uncle block
func calculateUncleReward(block, uncle *types.Eth1Block, chainID string) *big.Int {
	reward := new(big.Int)

	if len(block.Difficulty) > 0 {
		reward.Add(big.NewInt(int64(uncle.GetNumber())), big.NewInt(8))
		reward.Sub(reward, big.NewInt(int64(block.GetNumber())))
		reward.Mul(reward, eth1BlockReward(chainID, block.GetNumber(), block.Difficulty))
		reward.Div(reward, big.NewInt(8))

		reward.Div(eth1BlockReward(chainID, block.GetNumber(), block.Difficulty), big.NewInt(32))
	}
	return reward
}
