package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/types/geth"
)

type Eth1InternalTransactionWithPosition struct {
	types.Eth1InternalTransaction
	txPosition int
}

type NodeClient struct {
	chainID   *big.Int
	rpcClient *rpc.Client
	ethClient *ethclient.Client
}

type ClientV2 interface {
	GetBlocks(start, end int64, traceMode string) ([]*types.Eth1Block, error)
	GetBlock(number int64, traceMode string) (*types.Eth1Block, error)
}

func NewNodeClient(endpoint string) (*NodeClient, error) {
	rpcClient, err := rpc.DialOptions(context.Background(), endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting chainid of rpcclient: %w", err)
	}

	return &NodeClient{
		chainID:   chainID,
		rpcClient: rpcClient,
		ethClient: ethClient,
	}, nil
}

func (client *NodeClient) GetBlocks(start, end int64, traceMode string) ([]*types.Eth1Block, error) {
	blocks := make([]*types.Eth1Block, end-start+1)
	for i := start; i <= end; i++ {
		block, err := client.GetBlock(i, traceMode)
		if err != nil {
			return nil, err
		}
		blocks[i-start] = block
	}
	return blocks, nil
}

func (client *NodeClient) GetBlock(number int64, traceMode string) (*types.Eth1Block, error) {
	start := time.Now()
	timings := &types.GetBlockTimings{}
	mu := sync.Mutex{}

	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_block").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var traces []*Eth1InternalTransactionWithPosition
	var block *geth_types.Block
	var receipts []*geth_types.Receipt
	g := new(errgroup.Group)
	g.Go(func() error {
		b, err := client.ethClient.BlockByNumber(ctx, big.NewInt(number))
		if err != nil {
			return err
		}
		mu.Lock()
		timings.Headers = time.Since(start)
		mu.Unlock()
		block = b
		return nil
	})
	g.Go(func() error {
		if err := client.rpcClient.CallContext(ctx, &receipts, "eth_getBlockReceipts", fmt.Sprintf("0x%x", number)); err != nil {
			return fmt.Errorf("error retrieving receipts for block %v: %w", number, err)
		}
		mu.Lock()
		timings.Receipts = time.Since(start)
		mu.Unlock()
		return nil
	})
	g.Go(func() error {
		t, err := client.getTrace(traceMode, big.NewInt(number))
		if err != nil {
			return fmt.Errorf("error retrieving traces for block %v: %w", number, err)
		}
		traces = t
		mu.Lock()
		timings.Traces = time.Since(start)
		mu.Unlock()
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return parseBlock(block, receipts, traces, client.ethClient.TransactionSender), nil
}

func (client *NodeClient) getTrace(traceMode string, blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	if blockNumber.Uint64() == 0 { // genesis block is not traceable
		return nil, nil
	}
	switch traceMode {
	case "parity":
		return client.getTraceParity(blockNumber)
	case "parity/geth":
		traces, err := client.getTraceParity(blockNumber)
		if err == nil {
			return traces, nil
		}
		log.Error(err, fmt.Sprintf("error tracing block via parity style traces (%v)", blockNumber), 0)
		// fallback to geth traces
		fallthrough
	case "geth":
		return client.getTraceGeth(blockNumber)
	}
	return nil, fmt.Errorf("unknown trace mode '%s'", traceMode)
}

func (client *NodeClient) getTraceParity(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.traceParity(blockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("error tracing block via parity style traces (%v): %w", blockNumber, err)
	}

	var indexedTraces []*Eth1InternalTransactionWithPosition
	for _, trace := range traces {
		if trace.Type == "reward" {
			continue
		}
		if trace.TransactionHash == "" {
			continue
		}

		from, to, value, traceType := trace.ConvertFields()
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     traceType,
				From:     from,
				To:       to,
				Value:    value,
				ErrorMsg: trace.Error,
				Path:     fmt.Sprint(trace.TraceAddress),
			},
			txPosition: trace.TransactionPosition,
		})
	}
	return indexedTraces, nil
}

func (client *NodeClient) getTraceGeth(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.traceGeth(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error tracing block via geth style traces (%v): %w", blockNumber, err)
	}

	var indexedTraces []*Eth1InternalTransactionWithPosition
	var txPosition int
	paths := make(map[*geth.TraceCall]string)
	for i, trace := range traces {
		switch trace.Type {
		case "CREATE2":
			trace.Type = "CREATE"
		case "CREATE", "SELFDESTRUCT", "SUICIDE", "CALL", "DELEGATECALL", "STATICCALL", "CALLCODE":
		case "":
			logrus.WithFields(logrus.Fields{"type": trace.Type, "block.Number": blockNumber}).Errorf("geth style trace without type")
			spew.Dump(trace)
			continue
		default:
			spew.Dump(trace)
			logrus.Fatalf("unknown trace type %v in tx %v:%v", trace.Type, blockNumber.String(), trace.TransactionPosition)
		}
		if txPosition != trace.TransactionPosition {
			txPosition = trace.TransactionPosition
			paths = make(map[*geth.TraceCall]string)
		}
		for index, call := range trace.Calls {
			paths[call] = fmt.Sprintf("%s %d", paths[trace], index)
		}
		log.Tracef("appending trace %v to tx %d:%x from %v to %v value %v", i, blockNumber, trace.TransactionPosition, trace.From, trace.To, trace.Value)
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     strings.ToLower(trace.Type),
				From:     trace.From.Bytes(),
				To:       trace.To.Bytes(),
				Value:    common.FromHex(trace.Value),
				ErrorMsg: trace.Error,
				Path:     fmt.Sprintf("[%s]", strings.TrimPrefix(paths[trace], " ")),
			},
			txPosition: trace.TransactionPosition,
		})
	}
	return indexedTraces, nil
}

func (client *NodeClient) traceGeth(blockNumber *big.Int) ([]*geth.TraceCall, error) {
	var res []*geth.Trace

	err := client.rpcClient.Call(&res, "debug_traceBlockByNumber", hexutil.EncodeBig(blockNumber), gethTracerArg)
	if err != nil {
		return nil, err
	}

	data := make([]*geth.TraceCall, 0, 20)
	for i, r := range res {
		r.Result.TransactionPosition = i
		extractCalls(r.Result, &data)
	}

	return data, nil
}

func (client *NodeClient) traceParity(blockNumber uint64) ([]*ParityTraceResult, error) {
	var res []*ParityTraceResult

	err := client.rpcClient.Call(&res, "trace_block", blockNumber)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type RawStoreClient struct {
	chainID *big.Int
	store   raw.StoreReader
}

func NewRawStoreClient(chainID *big.Int, store raw.StoreReader) *RawStoreClient {
	return &RawStoreClient{
		chainID: chainID,
		store:   store,
	}
}

func (client *RawStoreClient) GetBlocks(start, end int64, traceMode string) ([]*types.Eth1Block, error) {
	rawBlocks, err := client.store.ReadBlocksByNumber(client.chainID.Uint64(), start, end)
	if err != nil {
		return nil, err
	}
	var blocks []*types.Eth1Block
	for _, rawBlock := range rawBlocks {
		block, receipts, traces, err := raw.GethParse(rawBlock)
		if err != nil {
			return nil, err
		}
		parsedBlock := parseBlock(block, receipts, parseTraceGeth(traces), func(_ context.Context, tx *geth_types.Transaction, block common.Hash, _ uint) (common.Address, error) {
			// Try to load the address from the cache.
			return geth_types.Sender(&raw.SenderFromDBSigner{Blockhash: block}, tx)
		})
		blocks = append(blocks, parsedBlock)
	}
	return blocks, nil
}

func (client *RawStoreClient) GetBlock(number int64, traceMode string) (*types.Eth1Block, error) {
	rawBlock, err := client.store.ReadBlockByNumber(client.chainID.Uint64(), number)
	if err != nil {
		return nil, err
	}
	block, receipts, traces, err := raw.GethParse(rawBlock)
	if err != nil {
		return nil, err
	}
	return parseBlock(block, receipts, parseTraceGeth(traces), func(_ context.Context, tx *geth_types.Transaction, block common.Hash, _ uint) (common.Address, error) {
		// Try to load the address from the cache.
		return geth_types.Sender(&raw.SenderFromDBSigner{Blockhash: block}, tx)
	}), nil
}

type MultiClient struct {
	node     *NodeClient
	rawStore *RawStoreClient
}

func NewMultiClient(endpoint string, store raw.StoreReader) (*MultiClient, error) {
	nodeClient, err := NewNodeClient(endpoint)
	if err != nil {
		return nil, err
	}
	return &MultiClient{
		node:     nodeClient,
		rawStore: NewRawStoreClient(nodeClient.chainID, store),
	}, nil
}

func (m MultiClient) GetBlocks(start, end int64, traceMode string) ([]*types.Eth1Block, error) {
	return m.rawStore.GetBlocks(start, end, traceMode)
}

func (m MultiClient) GetBlock(number int64, traceMode string) (*types.Eth1Block, error) {
	return m.rawStore.GetBlock(number, traceMode)
}

type TransactionSenderFunc func(ctx context.Context, tx *geth_types.Transaction, block common.Hash, index uint) (common.Address, error)

func parseBlock(block *geth_types.Block, receipts []*geth_types.Receipt, traces []*Eth1InternalTransactionWithPosition, txSender TransactionSenderFunc) *types.Eth1Block {
	withdrawals := make([]*types.Eth1Withdrawal, len(block.Withdrawals()))
	for i, withdrawal := range block.Withdrawals() {
		withdrawals[i] = &types.Eth1Withdrawal{
			Index:          withdrawal.Index,
			ValidatorIndex: withdrawal.Validator,
			Address:        withdrawal.Address.Bytes(),
			Amount:         new(big.Int).SetUint64(withdrawal.Amount).Bytes(),
		}
	}

	transactions := make([]*types.Eth1Transaction, len(block.Transactions()))
	traceIndex := 0
	for txPosition, receipt := range receipts {
		logs := make([]*types.Eth1Log, len(receipt.Logs))
		for i, log := range receipt.Logs {
			topics := make([][]byte, len(log.Topics))
			for j, topic := range log.Topics {
				topics[j] = topic.Bytes()
			}
			logs[i] = &types.Eth1Log{
				Address: log.Address.Bytes(),
				Data:    log.Data,
				Removed: log.Removed,
				Topics:  topics,
			}
		}

		var internals []*types.Eth1InternalTransaction
		for ; traceIndex < len(traces) && traces[traceIndex].txPosition == txPosition; traceIndex++ {
			internals = append(internals, &traces[traceIndex].Eth1InternalTransaction)
		}

		tx := block.Transactions()[txPosition]
		transactions[txPosition] = &types.Eth1Transaction{
			Type:                 uint32(tx.Type()),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice().Bytes(),
			MaxPriorityFeePerGas: tx.GasTipCap().Bytes(),
			MaxFeePerGas:         tx.GasFeeCap().Bytes(),
			Gas:                  tx.Gas(),
			Value:                tx.Value().Bytes(),
			Data:                 tx.Data(),
			To: func() []byte {
				if tx.To() != nil {
					return tx.To().Bytes()
				}
				return nil
			}(),
			From: func() []byte {
				// this won't make a request in most cases as the sender is already present in the cache
				// context https://github.com/ethereum/go-ethereum/blob/v1.14.11/ethclient/ethclient.go#L268
				sender, err := txSender(context.Background(), tx, block.Hash(), uint(txPosition))
				if err != nil {
					sender = common.HexToAddress("abababababababababababababababababababab")
					logrus.Errorf("could not retrieve tx sender %v: %v", tx.Hash(), err)
				}
				return sender.Bytes()
			}(),
			ChainId:            tx.ChainId().Bytes(),
			AccessList:         []*types.AccessList{},
			Hash:               tx.Hash().Bytes(),
			ContractAddress:    receipt.ContractAddress[:],
			CommulativeGasUsed: receipt.CumulativeGasUsed,
			GasUsed:            receipt.GasUsed,
			LogsBloom:          receipt.Bloom[:],
			Status:             receipt.Status,
			Logs:               logs,
			Itx:                internals,
			MaxFeePerBlobGas: func() []byte {
				if tx.BlobGasFeeCap() != nil {
					return tx.BlobGasFeeCap().Bytes()
				}
				return nil
			}(),
			BlobVersionedHashes: func() (b [][]byte) {
				for _, h := range tx.BlobHashes() {
					b = append(b, h.Bytes())
				}
				return b
			}(),
			BlobGasPrice: func() []byte {
				if receipt.BlobGasPrice != nil {
					return receipt.BlobGasPrice.Bytes()
				}
				return nil
			}(),
			BlobGasUsed: receipt.BlobGasUsed,
		}
	}

	uncles := make([]*types.Eth1Block, len(block.Uncles()))
	for i, uncle := range block.Uncles() {
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

	return &types.Eth1Block{
		Hash:        block.Hash().Bytes(),
		ParentHash:  block.ParentHash().Bytes(),
		UncleHash:   block.UncleHash().Bytes(),
		Coinbase:    block.Coinbase().Bytes(),
		Root:        block.Root().Bytes(),
		TxHash:      block.TxHash().Bytes(),
		ReceiptHash: block.ReceiptHash().Bytes(),
		Difficulty:  block.Difficulty().Bytes(),
		Number:      block.NumberU64(),
		GasLimit:    block.GasLimit(),
		GasUsed:     block.GasUsed(),
		Time:        timestamppb.New(time.Unix(int64(block.Time()), 0)),
		Extra:       block.Extra(),
		MixDigest:   block.MixDigest().Bytes(),
		Bloom:       block.Bloom().Bytes(),
		BaseFee: func() []byte {
			if block.BaseFee() != nil {
				return block.BaseFee().Bytes()
			}
			return nil
		}(),
		Uncles:       uncles,
		Transactions: transactions,
		Withdrawals:  withdrawals,
		BlobGasUsed: func() uint64 {
			blobGasUsed := block.BlobGasUsed()
			if blobGasUsed != nil {
				return *blobGasUsed
			}
			return 0
		}(),
		ExcessBlobGas: func() uint64 {
			excessBlobGas := block.ExcessBlobGas()
			if excessBlobGas != nil {
				return *excessBlobGas
			}
			return 0
		}(),
	}
}

func parseTraceGeth(traces []*geth.Trace) []*Eth1InternalTransactionWithPosition {
	var indexedTraces []*Eth1InternalTransactionWithPosition
	var txPosition int
	paths := make(map[*geth.TraceCall]string)
	for i, trace := range traces {
		switch trace.Result.Type {
		case "CREATE2":
			trace.Result.Type = "CREATE"
		case "CREATE", "SELFDESTRUCT", "SUICIDE", "CALL", "DELEGATECALL", "STATICCALL", "CALLCODE":
		case "":
			logrus.WithFields(logrus.Fields{"type": trace.Result.Type, "tx.hash": trace.TxHash}).Errorf("geth style trace without type")
			spew.Dump(trace)
			continue
		default:
			spew.Dump(trace)
			logrus.Fatalf("unknown trace type %v in tx %v", trace.Result.Type, trace.TxHash)
		}
		if txPosition != trace.Result.TransactionPosition {
			txPosition = trace.Result.TransactionPosition
			paths = make(map[*geth.TraceCall]string)
		}
		for index, call := range trace.Result.Calls {
			paths[call] = fmt.Sprintf("%s %d", paths[trace.Result], index)
		}
		log.Tracef("appending trace %v to tx %s from %v to %v value %v", i, trace.TxHash, trace.Result.From, trace.Result.To, trace.Result.Value)
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     strings.ToLower(trace.Result.Type),
				From:     trace.Result.From.Bytes(),
				To:       trace.Result.To.Bytes(),
				Value:    common.FromHex(trace.Result.Value),
				ErrorMsg: trace.Result.Error,
				Path:     fmt.Sprintf("[%s]", strings.TrimPrefix(paths[trace.Result], " ")),
			},
			txPosition: trace.Result.TransactionPosition,
		})
	}
	return indexedTraces
}
