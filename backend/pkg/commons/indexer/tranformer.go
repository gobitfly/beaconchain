package indexer

import (
	"bytes"
	"maps"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadata"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Transformer struct {
	cache metadata.Cache
}

func NewTransformer(cache metadata.Cache) *Transformer {
	return &Transformer{
		cache: cache,
	}
}

func (t *Transformer) Tx(chainID string, blk *types.Eth1Block) (map[string][]database.Item, map[string][]database.Item, error) {
	/*	startTime := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("bt_transform_tx").Observe(time.Since(startTime).Seconds())
		}()*/

	updateMetadata := make(map[string][]database.Item)
	var transactions []*types.Eth1TransactionIndexed
	for _, tx := range blk.Transactions {
		to := tx.GetTo()
		isContract := false
		if !bytes.Equal(tx.GetContractAddress(), common.Address{}.Bytes()) {
			to = tx.GetContractAddress()
			isContract = true
		}
		method := make([]byte, 0)
		if len(tx.GetData()) > 3 {
			method = tx.GetData()[:4]
		}

		fee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetGasPrice()), big.NewInt(int64(tx.GetGasUsed()))).Bytes()
		blobFee := new(big.Int).Mul(new(big.Int).SetBytes(tx.GetBlobGasPrice()), big.NewInt(int64(tx.GetBlobGasUsed()))).Bytes()
		indexedTx := &types.Eth1TransactionIndexed{
			Hash:               tx.GetHash(),
			BlockNumber:        blk.GetNumber(),
			Time:               blk.GetTime(),
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
		for _, itx := range tx.Itx {
			if itx.ErrorMsg != "" {
				indexedTx.ErrorMsg = itx.ErrorMsg
				if indexedTx.Status == types.StatusType_SUCCESS {
					indexedTx.Status = types.StatusType_PARTIAL
				}
				break
			}
		}

		transactions = append(transactions, indexedTx)

		// mark sender and recipient for balance update
		maps.Copy(updateMetadata, metadata.MarkBalanceUpdate(chainID, indexedTx.From, []byte{0x0}, t.cache))
		maps.Copy(updateMetadata, metadata.MarkBalanceUpdate(chainID, indexedTx.To, []byte{0x0}, t.cache))
	}
	update, err := data.BlockTransactionsToItemsV2(chainID, transactions)
	if err != nil {
		return nil, nil, err
	}

	return update, updateMetadata, nil
}

func (t *Transformer) ERC20(chainID string, blk *types.Eth1Block) (map[string][]database.Item, map[string][]database.Item, error) {
	/*	startTime := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("bt_transform_erc20").Observe(time.Since(startTime).Seconds())
		}()*/

	updateMetadata := make(map[string][]database.Item)
	filterer, err := erc20.NewErc20Filterer(common.Address{}, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot ERC20 create filterer")
	}
	var transfers []data.TransferWithIndexes
	for txIndex, tx := range blk.GetTransactions() {
		for logIndex, log := range tx.GetLogs() {
			if len(log.GetTopics()) != 3 || !bytes.Equal(log.GetTopics()[0], erc20.TransferTopic) {
				continue
			}

			topics := make([]common.Hash, 0, len(log.GetTopics()))

			for _, lTopic := range log.GetTopics() {
				topics = append(topics, common.BytesToHash(lTopic))
			}

			ethLog := gethtypes.Log{
				Address:     common.BytesToAddress(log.GetAddress()),
				Data:        log.Data,
				Topics:      topics,
				BlockNumber: blk.GetNumber(),
				TxHash:      common.BytesToHash(tx.GetHash()),
				TxIndex:     uint(txIndex),
				BlockHash:   common.BytesToHash(blk.GetHash()),
				Index:       uint(logIndex),
				Removed:     log.GetRemoved(),
			}

			transfer, _ := filterer.ParseTransfer(ethLog)
			if transfer == nil {
				continue
			}

			var value []byte
			if transfer.Value != nil {
				value = transfer.Value.Bytes()
			}

			indexedLog := &types.Eth1ERC20Indexed{
				ParentHash:   tx.GetHash(),
				BlockNumber:  blk.GetNumber(),
				Time:         blk.GetTime(),
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

			// mark sender and recipient for balance update
			maps.Copy(updateMetadata, metadata.MarkBalanceUpdate(chainID, indexedLog.From, indexedLog.TokenAddress, t.cache))
			maps.Copy(updateMetadata, metadata.MarkBalanceUpdate(chainID, indexedLog.To, indexedLog.TokenAddress, t.cache))
		}
	}
	update, err := data.BlockERC20TransfersToItemsV2(chainID, transfers)
	if err != nil {
		return nil, nil, err
	}
	return update, updateMetadata, nil
}
