package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	endpoint     string
	rpcClient    *gethrpc.Client
	ethClient    *ethclient.Client
	chainID      *big.Int
	multiChecker *Balance
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

	client.multiChecker, err = NewBalance(common.HexToAddress("0xb1F8e55c7f64D203C1400B9D8555d050F94aDF39"), client.ethClient)
	if err != nil {
		return nil, fmt.Errorf("error initiation balance checker contract: %w", err)
	}

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
	// we cannot trust block.Hash(), some chain (gnosis) have extra field that are included in the hash computation
	// so extract it from the receipts or from the node again if no receipt (it should be very rare)
	var blockHash common.Hash
	if len(receipts) != 0 {
		blockHash = receipts[0].BlockHash
	} else {
		var res minimalBlock
		if err := client.rpcClient.CallContext(ctx, &res, "eth_getBlockByNumber", fmt.Sprintf("0x%x", number), false); err != nil {
			return nil, nil, fmt.Errorf("error retrieving blockHash %v: %w", number, err)
		}
		blockHash = common.HexToHash(res.Hash)
	}

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
	if len(receipts) != len(block.Transactions()) {
		return nil, nil, fmt.Errorf("block %s receipts length [%d] mismatch with transactions length [%d]", block.Number(), len(receipts), len(block.Transactions()))
	}
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
				sender, err := client.ethClient.TransactionSender(context.Background(), tx, blockHash, uint(txPosition))
				if err != nil {
					sender = common.HexToAddress("abababababababababababababababababababab")
					log.Error(err, "error converting tx to msg", 0, map[string]interface{}{"tx": tx.Hash()})
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
		Hash:        blockHash.Bytes(),
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

func (client *ErigonClient) GetBalances(pairs []*types.Eth1AddressBalance, addressIndex, tokenIndex int) ([]*types.Eth1AddressBalance, error) {
	batchElements := make([]gethrpc.BatchElem, 0, len(pairs))

	ret := make([]*types.Eth1AddressBalance, len(pairs))

	for i, pair := range pairs {
		result := ""

		ret[i] = &types.Eth1AddressBalance{
			Address: pair.Address,
			Token:   pair.Token,
		}

		// log.LogInfo("retrieving balance for %x / %x", ret[i].Address, ret[i].Token)

		if len(pair.Token) < 20 {
			batchElements = append(batchElements, gethrpc.BatchElem{
				Method: "eth_getBalance",
				Args:   []interface{}{common.BytesToAddress(pair.Address), "latest"},
				Result: &result,
			})
		} else {
			to := common.BytesToAddress(pair.Token)
			msg := ethereum.CallMsg{
				To:   &to,
				Gas:  1000000,
				Data: common.Hex2Bytes(fmt.Sprintf("70a08231000000000000000000000000%x", pair.Address)),
			}

			batchElements = append(batchElements, gethrpc.BatchElem{
				Method: "eth_call",
				Args:   []interface{}{toCallArg(msg), "latest"},
				Result: &result,
			})
		}
	}

	err := client.rpcClient.BatchCall(batchElements)
	if err != nil {
		return nil, fmt.Errorf("error during batch request: %w", err)
	}

	for i, el := range batchElements {
		if el.Error != nil {
			log.Warnf("error in batch call: %v", el.Error) // PPR: are smart contracts that pretend to implement the erc20 standard but are somehow buggy
		}

		res := strings.TrimPrefix(*el.Result.(*string), "0x")
		ret[i].Balance = new(big.Int).SetBytes(common.FromHex(res)).Bytes()

		// log.LogInfo("retrieved balance %x / %x: %x (%v)", ret[i].Address, ret[i].Token, ret[i].Balance, *el.Result.(*string))
	}

	return ret, nil
}

func (client *ErigonClient) GetBalancesForAddresse(address string, tokenStr []string) ([]*types.Eth1AddressBalance, error) {
	opts := &bind.CallOpts{
		BlockNumber: nil,
	}

	tokens := make([]common.Address, 0, len(tokenStr))

	for _, token := range tokenStr {
		tokens = append(tokens, common.HexToAddress(token))
	}
	balancesInt, err := client.multiChecker.Balances(opts, []common.Address{common.HexToAddress(address)}, tokens)
	if err != nil {
		return nil, err
	}

	res := make([]*types.Eth1AddressBalance, len(tokenStr))
	for tokenIdx := range tokens {
		res[tokenIdx] = &types.Eth1AddressBalance{
			Address: common.FromHex(address),
			Token:   common.FromHex(string(tokens[tokenIdx].Bytes())),
			Balance: balancesInt[tokenIdx].Bytes(),
		}
	}

	return res, nil
}

func (client *ErigonClient) GetNativeBalance(address string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	balance, err := client.ethClient.BalanceAt(ctx, common.HexToAddress(address), nil)

	if err != nil {
		return nil, err
	}
	return balance.Bytes(), nil
}

func (client *ErigonClient) GetERC20TokenBalance(address string, token string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	to := common.HexToAddress(token)
	balance, err := client.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &to,
		Gas:  1000000,
		Data: common.Hex2Bytes("70a08231000000000000000000000000" + address),
	}, nil)

	if err != nil && !strings.HasPrefix(err.Error(), "execution reverted") {
		return nil, err
	}
	return balance, nil
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
	})

	g.Go(func() error {
		totalSupply, err := contract.TotalSupply(nil)
		if err != nil {
			return fmt.Errorf("error retrieving total supply: %w", err)
		}
		ret.TotalSupply = totalSupply.Bytes()
		return nil
	})

	g.Go(func() error {
		decimals, err := contract.Decimals(nil)
		if err != nil {
			return fmt.Errorf("error retrieving decimals: %w", err)
		}
		ret.Decimals = big.NewInt(int64(decimals)).Bytes()
		return nil
	})

	g.Go(func() error {
		rate, err := oracle.GetRateToEth(nil, common.BytesToAddress(token), false)
		if err != nil {
			return fmt.Errorf("error calling oneinchoracle.GetRateToEth: %w", err)
		}
		ret.Price = rate.Bytes()
		return nil
	})

	err = g.Wait()
	if err != nil {
		return ret, err
	}

	if err == nil && len(ret.Decimals) == 0 && ret.Symbol == "" && len(ret.TotalSupply) == 0 {
		// it's possible that a token contract implements the ERC20 interfaces but does not return any values; we use a backup in this case
		ret = &types.ERC20Metadata{
			Decimals:    []byte{0x0},
			Symbol:      "UNKNOWN",
			TotalSupply: []byte{0x0}}
	}

	return ret, err
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
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
