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
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/contracts/oneinchoracle"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/types/geth"
)

type ErigonClient struct {
	endpoint  string
	rpcClient *gethrpc.Client
	ethClient *ethclient.Client
	chainID   *big.Int
}

var CurrentErigonClient *ErigonClient

func NewErigonClient(endpoint string) (*ErigonClient, error) {
	log.Infof("initializing erigon client at %v", endpoint)
	client := &ErigonClient{
		endpoint: endpoint,
	}

	rpcClient, err := gethrpc.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}
	client.rpcClient = rpcClient

	ethClient, err := ethclient.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}
	client.ethClient = ethClient

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	chainID, err := client.ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting chainid of rpcclient: %w", err)
	}
	client.chainID = chainID

	return client, nil
}

func (client *ErigonClient) Close() {
	client.rpcClient.Close()
	client.ethClient.Close()
}

func (client *ErigonClient) GetChainID() *big.Int {
	return client.chainID
}

func (client *ErigonClient) GetNativeClient() *ethclient.Client {
	return client.ethClient
}

func (client *ErigonClient) GetRPCClient() *gethrpc.Client {
	return client.rpcClient
}

type minimalBlock struct {
	Hash string `json:"hash"`
}

func (client *ErigonClient) GetBlock(number int64, traceMode string) (*types.Eth1Block, *types.GetBlockTimings, error) {
	start := time.Now()
	timings := &types.GetBlockTimings{}
	mu := sync.Mutex{}

	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_block").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var traces []*Eth1InternalTransactionWithPosition
	var block *gethtypes.Block
	var receipts []*gethtypes.Receipt
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
		return nil, nil, err
	}

	blockHash, err := client.getBlockHash(number, receipts)
	if err != nil {
		return nil, nil, err
	}

	transactions := make([]*types.Eth1Transaction, len(block.Transactions()))
	traceIndex := 0
	if len(receipts) != len(block.Transactions()) {
		return nil, nil, fmt.Errorf("block %s receipts length [%d] mismatch with transactions length [%d]", block.Number(), len(receipts), len(block.Transactions()))
	}
	for txPosition, receipt := range receipts {
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
			To:                   getReceiver(tx),
			From:                 client.getSender(tx, blockHash, txPosition),
			ChainId:              tx.ChainId().Bytes(),
			AccessList:           []*types.AccessList{},
			Hash:                 tx.Hash().Bytes(),
			ContractAddress:      receipt.ContractAddress[:],
			CommulativeGasUsed:   receipt.CumulativeGasUsed,
			GasUsed:              receipt.GasUsed,
			LogsBloom:            receipt.Bloom[:],
			Status:               receipt.Status,
			Logs:                 getLogsFromReceipts(receipt.Logs),
			Itx:                  getInternalTxs(traceIndex, traces, txPosition),
			MaxFeePerBlobGas:     getMaxFeePerBlobGas(tx),
			BlobVersionedHashes:  getBlobVersionedHashes(tx),
			BlobGasPrice:         getBlobGasPrice(receipt),
			BlobGasUsed:          receipt.BlobGasUsed,
		}
	}

	return &types.Eth1Block{
		Hash:          blockHash.Bytes(),
		ParentHash:    block.ParentHash().Bytes(),
		UncleHash:     block.UncleHash().Bytes(),
		Coinbase:      block.Coinbase().Bytes(),
		Root:          block.Root().Bytes(),
		TxHash:        block.TxHash().Bytes(),
		ReceiptHash:   block.ReceiptHash().Bytes(),
		Difficulty:    block.Difficulty().Bytes(),
		Number:        block.NumberU64(),
		GasLimit:      block.GasLimit(),
		GasUsed:       block.GasUsed(),
		Time:          timestamppb.New(time.Unix(int64(block.Time()), 0)),
		Extra:         block.Extra(),
		MixDigest:     block.MixDigest().Bytes(),
		Bloom:         block.Bloom().Bytes(),
		BaseFee:       getBaseFee(block),
		Uncles:        getBlockUncles(block.Uncles()),
		Transactions:  transactions,
		Withdrawals:   getBlockWithdrawals(block.Withdrawals()),
		BlobGasUsed:   getBlobGasUsed(block),
		ExcessBlobGas: getExcessBlobGas(block),
	}, timings, nil
}

func (client *ErigonClient) GetBlockNumberByHash(hash string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	block, err := client.ethClient.BlockByHash(ctx, common.HexToHash(hash))
	if err != nil {
		return 0, err
	}
	return block.NumberU64(), nil
}

func (client *ErigonClient) GetLatestEth1BlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	latestBlock, err := client.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest block: %w", err)
	}

	return latestBlock.NumberU64(), nil
}

func extractCalls(r *geth.TraceCall, d *[]*geth.TraceCall) {
	if r == nil {
		return
	}
	*d = append(*d, r)

	if r.Calls == nil {
		return
	}
	for _, c := range r.Calls {
		c.TransactionPosition = r.TransactionPosition
		extractCalls(c, d)
	}
}

func (client *ErigonClient) TraceGeth(blockNumber *big.Int) ([]*geth.TraceCall, error) {
	var res []*geth.Trace

	err := client.rpcClient.Call(&res, "debug_traceBlockByNumber", hexutil.EncodeBig(blockNumber), geth.Tracer)
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

type ParityTraceResult struct {
	Action struct {
		CallType      string `json:"callType"`
		From          string `json:"from"`
		Gas           string `json:"gas"`
		Input         string `json:"input"`
		To            string `json:"to"`
		Value         string `json:"value"`
		Init          string `json:"init"`
		Address       string `json:"address"`
		Balance       string `json:"balance"`
		RefundAddress string `json:"refundAddress"`
		Author        string `json:"author"`
		RewardType    string `json:"rewardType"`
	} `json:"action"`
	BlockHash   string `json:"blockHash"`
	BlockNumber int    `json:"blockNumber"`
	Error       string `json:"error"`
	Result      struct {
		GasUsed string `json:"gasUsed"`
		Code    string `json:"code"`
		Output  string `json:"output"`
		Address string `json:"address"`
	} `json:"result"`

	Subtraces           int     `json:"subtraces"`
	TraceAddress        []int64 `json:"traceAddress"`
	TransactionHash     string  `json:"transactionHash"`
	TransactionPosition int     `json:"transactionPosition"`
	Type                string  `json:"type"`
}

func (trace *ParityTraceResult) ConvertFields() ([]byte, []byte, []byte, string) {
	var from, to, value []byte
	tx_type := trace.Type

	switch trace.Type {
	case "create":
		from = common.FromHex(trace.Action.From)
		to = common.FromHex(trace.Result.Address)
		value = common.FromHex(trace.Action.Value)
	case "suicide":
		from = common.FromHex(trace.Action.Address)
		to = common.FromHex(trace.Action.RefundAddress)
		value = common.FromHex(trace.Action.Balance)
	case "call":
		from = common.FromHex(trace.Action.From)
		to = common.FromHex(trace.Action.To)
		value = common.FromHex(trace.Action.Value)
		tx_type = trace.Action.CallType
	default:
		spew.Dump(trace)
		log.Fatal(nil, "unknown trace type", 0, map[string]interface{}{"trace type": trace.Type, "tx hash": trace.TransactionHash})
	}
	return from, to, value, tx_type
}

func (client *ErigonClient) TraceParity(blockNumber uint64) ([]*ParityTraceResult, error) {
	var res []*ParityTraceResult

	err := client.rpcClient.Call(&res, "trace_block", blockNumber)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *ErigonClient) TraceParityTx(txHash string) ([]*ParityTraceResult, error) {
	var res []*ParityTraceResult

	err := client.rpcClient.Call(&res, "trace_transaction", txHash)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *ErigonClient) GetERC20TokenMetadata(token []byte) (*types.ERC20Metadata, error) {
	log.Infof("retrieving metadata for token %x", token)

	oracle, err := oneinchoracle.NewOneInchOracleByChainID(client.GetChainID(), client.ethClient)
	if err != nil {
		return nil, err
	}

	contract, err := contracts.NewIERC20Metadata(common.BytesToAddress(token), client.ethClient)
	if err != nil {
		return nil, err
	}

	g := new(errgroup.Group)

	ret := &types.ERC20Metadata{}

	g.Go(func() error {
		return getERC20ContractSymbol(contract, ret)
	})

	g.Go(func() error {
		return getERC20ContractTotalSupply(contract, ret)
	})

	g.Go(func() error {
		return getERC20ContractDecimals(contract, ret)
	})

	g.Go(func() error {
		return getRateFromOracle(oracle, token, ret)
	})

	err = g.Wait()
	if err != nil {
		return ret, err
	}

	if len(ret.Decimals) == 0 && ret.Symbol == "" && len(ret.TotalSupply) == 0 {
		// it's possible that a token contract implements the ERC20 interfaces but does not
		// return any values; we use a backup in this case
		ret = &types.ERC20Metadata{
			Decimals:    []byte{0x0},
			Symbol:      "UNKNOWN",
			TotalSupply: []byte{0x0}}
	}

	return ret, err
}

type Eth1InternalTransactionWithPosition struct {
	types.Eth1InternalTransaction
	txPosition int
}

func (client *ErigonClient) getTrace(traceMode string, blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
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
		log.Error(err, "error tracing block via parity style traces", 0, map[string]interface{}{"blockNumber": blockNumber.String()})
		// fallback to geth traces
		fallthrough
	case "geth":
		return client.getTraceGeth(blockNumber)
	}
	return nil, fmt.Errorf("unknown trace mode '%s'", traceMode)
}

func (client *ErigonClient) getTraceParity(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.TraceParity(blockNumber.Uint64())
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

func (client *ErigonClient) getTraceGeth(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.TraceGeth(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error tracing block via geth style traces (%v): %w", blockNumber, err)
	}

	var indexedTraces []*Eth1InternalTransactionWithPosition
	var txPosition int
	paths := make(map[*geth.TraceCall]string)
	for _, trace := range traces {
		switch trace.Type {
		case "CREATE2":
			trace.Type = "CREATE"
		case "CREATE", "SELFDESTRUCT", "SUICIDE", "CALL", "DELEGATECALL", "STATICCALL", "CALLCODE":
		case "":
			log.Error(fmt.Errorf("geth style trace without type"), "", 0, map[string]interface{}{"type": trace.Type, "block.Number": blockNumber.String()})
			spew.Dump(trace)
			continue
		default:
			spew.Dump(trace)
			log.Fatal(nil, "unknown trace type", 0, map[string]interface{}{"trace type": trace.Type, "block": blockNumber.String(), "tx_index": trace.TransactionPosition})
		}
		if txPosition != trace.TransactionPosition {
			txPosition = trace.TransactionPosition
			paths = make(map[*geth.TraceCall]string)
		}
		for index, call := range trace.Calls {
			paths[call] = fmt.Sprintf("%s %d", paths[trace], index)
		}

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

func (client *ErigonClient) getBlockHash(blockNumber int64, receipts []*gethtypes.Receipt) (common.Hash, error) {
	var blockHash common.Hash
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// we cannot trust block.Hash(), some chain (gnosis) have extra field that are included in the hash computation
	// so extract it from the receipts or from the node again if no receipt (it should be very rare)
	if len(receipts) != 0 {
		blockHash = receipts[0].BlockHash
	} else {
		var res minimalBlock
		if err := client.rpcClient.CallContext(ctx, &res, "eth_getBlockByNumber", fmt.Sprintf("0x%x", blockNumber), false); err != nil {
			return common.Hash{}, fmt.Errorf("error retrieving blockHash %v: %w", blockNumber, err)
		}
		blockHash = common.HexToHash(res.Hash)
	}

	return blockHash, nil
}

func (client *ErigonClient) getSender(tx *gethtypes.Transaction, blockHash common.Hash, txPosition int) []byte {
	// this won't make a request in most cases as the sender is already present in the cache
	// context https://github.com/ethereum/go-ethereum/blob/v1.14.11/ethclient/ethclient.go#L268
	sender, err := client.ethClient.TransactionSender(context.Background(), tx, blockHash, uint(txPosition))
	if err != nil {
		sender = common.HexToAddress("abababababababababababababababababababab")
		log.Error(err, "error converting tx to msg", 0, map[string]interface{}{"tx": tx.Hash()})
	}
	return sender.Bytes()
}

func getInternalTxs(traceIndex int, traces []*Eth1InternalTransactionWithPosition, txPosition int) []*types.Eth1InternalTransaction {
	var internals []*types.Eth1InternalTransaction
	for ; traceIndex < len(traces) && traces[traceIndex].txPosition == txPosition; traceIndex++ {
		internals = append(internals, &traces[traceIndex].Eth1InternalTransaction)
	}
	return internals
}

func getMaxFeePerBlobGas(tx *gethtypes.Transaction) []byte {
	if tx.BlobGasFeeCap() != nil {
		return tx.BlobGasFeeCap().Bytes()
	}
	return nil
}

func getBlobVersionedHashes(tx *gethtypes.Transaction) [][]byte {
	var hashes [][]byte
	for _, h := range tx.BlobHashes() {
		hashes = append(hashes, h.Bytes())
	}
	return hashes
}

func getBlobGasPrice(receipt *gethtypes.Receipt) []byte {
	if receipt.BlobGasPrice != nil {
		return receipt.BlobGasPrice.Bytes()
	}
	return nil
}

func getBaseFee(block *gethtypes.Block) []byte {
	if block.BaseFee() != nil {
		return block.BaseFee().Bytes()
	}
	return nil
}

func getBlobGasUsed(block *gethtypes.Block) uint64 {
	blobGasUsed := block.BlobGasUsed()
	if blobGasUsed != nil {
		return *blobGasUsed
	}
	return 0
}

func getExcessBlobGas(block *gethtypes.Block) uint64 {
	excessBlobGas := block.ExcessBlobGas()
	if excessBlobGas != nil {
		return *excessBlobGas
	}
	return 0
}
