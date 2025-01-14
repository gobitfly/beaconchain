package raw

import (
	"encoding/json"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/chain"
	"github.com/gobitfly/beaconchain/pkg/commons/hexutil"
)

type minimalBlock struct {
	Result struct {
		Hash         string `json:"hash"`
		Transactions []struct {
			Hash string `json:"hash"`
		} `json:"transactions"`
	} `json:"result"`
}
type minimalReceipts struct {
	Result []minimalReceipt `json:"result"`
}
type minimalReceiptResp struct {
	Result minimalReceipt `json:"result"`
}
type minimalReceipt struct {
	BlockHash string `json:"blockHash"`
	Hash      string `json:"transactionHash"`
}
type minimalTraces struct {
	Result []struct {
		TxHash string `json:"txHash"`
	} `json:"result"`
}

func validateBlock(fullBlock FullBlockData) error {
	if len(fullBlock.Block) == 0 || len(fullBlock.BlockTxs) != 0 && len(fullBlock.Receipts) == 0 || len(fullBlock.BlockTxs) != 0 && len(fullBlock.Traces) == 0 {
		return fmt.Errorf("empty data block=%d receipts=%d traces=%d", len(fullBlock.Block), len(fullBlock.Receipts), len(fullBlock.Traces))
	}
	var block minimalBlock
	_ = json.Unmarshal(fullBlock.Block, &block)

	var receipts minimalReceipts
	_ = json.Unmarshal(fullBlock.Receipts, &receipts)
	// Arbitrum did not support eth_getBlockReceipts at one point
	// so the receipt data is the list of all eth_getTransactionReceipt responses
	if fullBlock.ChainID == chain.IDs.Arbitrum.Uint64() {
		receipts = normalizeReceipts(fullBlock.Receipts)
	}

	var traces minimalTraces
	_ = json.Unmarshal(fullBlock.Traces, &traces)

	if len(block.Result.Transactions) != len(receipts.Result) || len(block.Result.Transactions) != len(traces.Result) {
		return fmt.Errorf("mismatch transaction count block=%d receipt=%d trace=%d", len(block.Result.Transactions), len(receipts.Result), len(traces.Result))
	}

	if len(block.Result.Transactions) == 0 {
		return nil
	}

	if block.Result.Hash != receipts.Result[0].BlockHash {
		return fmt.Errorf("mismatch block hash block=%s receipt=%s", block.Result.Hash, receipts.Result[0].BlockHash)
	}

	for i := 0; i < len(block.Result.Transactions); i++ {
		tx, receipt, trace := block.Result.Transactions[i].Hash, block.Result.Transactions[i].Hash, traces.Result[i].TxHash
		if tx != receipt || (trace != "" && tx != trace) {
			return fmt.Errorf("mismatch transaction hash at index=%d block=%s receipt=%s trace=%s", i, tx, receipt, trace)
		}
	}

	return nil
}

// normalizeReceipts parse a list of eth_getTransactionReceipt response and transform it to something similar to
// eth_getBlockReceipts response
func normalizeReceipts(rawReceipts hexutil.Bytes) minimalReceipts {
	var receipts []minimalReceiptResp
	_ = json.Unmarshal(rawReceipts, &receipts)
	var normalizedReceipts []minimalReceipt
	for _, receipt := range receipts {
		normalizedReceipts = append(normalizedReceipts, receipt.Result)
	}
	return minimalReceipts{
		Result: normalizedReceipts,
	}
}
