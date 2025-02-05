package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/contracts/oneinchoracle"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types/geth"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/types"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
)

type GethClient struct {
	endpoint     string
	rpcClient    *gethrpc.Client
	ethClient    *ethclient.Client
	chainID      *big.Int
	multiChecker *Balance
}

var CurrentGethClient *GethClient

func NewGethClient(endpoint string) (*GethClient, error) {
	log.Infof("initializing geth client at %v", endpoint)
	client := &GethClient{
		endpoint: endpoint,
	}

	rpcClient, err := gethrpc.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %v", err)
	}

	client.rpcClient = rpcClient

	ethClient, err := ethclient.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %v", err)
	}
	client.ethClient = ethClient

	client.multiChecker, err = NewBalance(common.HexToAddress("0xb1F8e55c7f64D203C1400B9D8555d050F94aDF39"), client.ethClient)
	if err != nil {
		return nil, fmt.Errorf("error initiation balance checker contract: %v", err)
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

func (client *GethClient) Close() {
	client.rpcClient.Close()
	client.ethClient.Close()
}

func (client *GethClient) GetChainID() *big.Int {
	return client.chainID
}

func (client *GethClient) GetNativeClient() *ethclient.Client {
	return client.ethClient
}

func (client *GethClient) GetRPCClient() *gethrpc.Client {
	return client.rpcClient
}

func (client *GethClient) GetBlock(number int64) (*types.Eth1Block, *types.GetBlockTimings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	start := time.Now()
	timings := &types.GetBlockTimings{}

	block, err := client.ethClient.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, nil, err
	}

	timings.Headers = time.Since(start)
	start = time.Now()

	uncles := getBlockUncles(block.Uncles())

	c := &types.Eth1Block{
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
		Transactions: []*types.Eth1Transaction{},
	}

	receipts := make([]*gethtypes.Receipt, len(block.Transactions()))
	reqs := make([]gethrpc.BatchElem, len(block.Transactions()))

	txs := block.Transactions()
	for _, tx := range txs {
		from := getGethSender(tx)
		to := getReceiver(tx)
		pbTx := &types.Eth1Transaction{
			Type:                 uint32(tx.Type()),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice().Bytes(),
			MaxPriorityFeePerGas: tx.GasTipCap().Bytes(),
			MaxFeePerGas:         tx.GasFeeCap().Bytes(),
			Gas:                  tx.Gas(),
			Value:                tx.Value().Bytes(),
			Data:                 tx.Data(),
			From:                 from,
			To:                   to,
			ChainId:              tx.ChainId().Bytes(),
			AccessList:           []*types.AccessList{},
			Hash:                 tx.Hash().Bytes(),
			Itx:                  []*types.Eth1InternalTransaction{},
		}

		c.Transactions = append(c.Transactions, pbTx)
	}

	for i := range reqs {
		reqs[i] = gethrpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{txs[i].Hash().String()},
			Result: &receipts[i],
		}
	}

	if len(reqs) > 0 {
		if err := client.rpcClient.BatchCallContext(ctx, reqs); err != nil {
			return nil, nil, fmt.Errorf("error retrieving receipts for block %v: %v", block.Number(), err)
		}
	}
	timings.Receipts = time.Since(start)

	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, nil, fmt.Errorf("error retrieving receipt %v for block %v: %v", i, block.Number(), reqs[i].Error)
		}
		if receipts[i] == nil {
			return nil, nil, fmt.Errorf("got null value for receipt %d of block %v", i, block.Number())
		}

		r := receipts[i]
		c.Transactions[i].ContractAddress = r.ContractAddress[:]
		c.Transactions[i].CommulativeGasUsed = r.CumulativeGasUsed
		c.Transactions[i].GasUsed = r.GasUsed
		c.Transactions[i].LogsBloom = r.Bloom[:]
		c.Transactions[i].Logs = getLogsFromReceipts(r.Logs)
	}

	return c, timings, nil
}

func (client *GethClient) GetLatestEth1BlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	latestBlock, err := client.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest block: %v", err)
	}

	return latestBlock.NumberU64(), nil
}

func (client *GethClient) TraceGeth(blockHash common.Hash) ([]*geth.Trace, error) {
	var res []*geth.Trace

	err := client.rpcClient.Call(&res, "debug_traceBlockByHash", blockHash, geth.Tracer)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *GethClient) GetBalances(pairs []string) ([]*types.Eth1AddressBalance, error) {
	batchElements := make([]gethrpc.BatchElem, 0, len(pairs))

	ret := make([]*types.Eth1AddressBalance, len(pairs))

	for i, pair := range pairs {
		s := strings.Split(pair, ":")

		if len(s) != 3 {
			log.Fatal(fmt.Errorf("%v has an invalid format", pair), "", 0)
		}

		if s[0] != "B" {
			log.Fatal(fmt.Errorf("%v has invalid balance update prefix", pair), "", 0)
		}

		address := s[1]
		token := s[2]
		result := ""

		ret[i] = &types.Eth1AddressBalance{
			Address: common.FromHex(address),
			Token:   common.FromHex(token),
		}

		if token == "00" {
			batchElements = append(batchElements, gethrpc.BatchElem{
				Method: "eth_getBalance",
				Args:   []interface{}{common.HexToAddress(address), "latest"},
				Result: &result,
			})
		} else {
			to := common.HexToAddress(token)
			msg := ethereum.CallMsg{
				To:   &to,
				Gas:  1000000,
				Data: common.Hex2Bytes("70a08231000000000000000000000000" + address),
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
		return nil, fmt.Errorf("error during batch request: %v", err)
	}

	for i, el := range batchElements {
		if el.Error != nil {
			log.Warnf("error in batch call: %v", el.Error) // PPR: are smart contracts that pretend to implement the erc20 standard but are somehow buggy
		}

		res := strings.TrimPrefix(*el.Result.(*string), "0x")
		ret[i].Balance = new(big.Int).SetBytes(common.Hex2Bytes(res)).Bytes()
	}

	return ret, nil
}

func (client *GethClient) GetBalancesForAddress(address string, tokenStr []string) ([]*types.Eth1AddressBalance, error) {
	opts := &bind.CallOpts{
		BlockNumber: nil,
	}

	tokens := getTokens(tokenStr)
	balancesInt, err := client.multiChecker.Balances(opts, []common.Address{common.HexToAddress(address)}, tokens)
	if err != nil {
		return nil, err
	}

	res, err := parseAddressBalance(tokens, address, balancesInt)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *GethClient) GetNativeBalance(address string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	balance, err := client.ethClient.BalanceAt(ctx, common.HexToAddress(address), nil)

	if err != nil {
		return nil, err
	}
	return balance.Bytes(), nil
}

func (client *GethClient) GetERC20TokenBalance(address string, token string) ([]byte, error) {
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

func (client *GethClient) GetERC20TokenMetadata(token []byte) (*types.ERC20Metadata, error) {
	log.Infof("retrieving metadata for token %x", token)

	oracle, err := oneinchoracle.NewOneInchOracleByChainID(client.GetChainID(), client.ethClient)
	if err != nil {
		return nil, fmt.Errorf("error initializing oneinchoracle.NewOneInchOracleByChainID: %w", err)
	}

	contract, err := contracts.NewIERC20Metadata(common.BytesToAddress(token), client.ethClient)
	if err != nil {
		return nil, fmt.Errorf("error getting token-contract: erc20.NewErc20: %w", err)
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
		if !oneinchoracle.SupportedChainId(client.GetChainID()) {
			return nil
		}
		return getRateFromOracle(oracle, token, ret)
	})

	err = g.Wait()
	if err != nil {
		return ret, err
	}

	if err == nil && len(ret.Decimals) == 0 && ret.Symbol == "" && len(ret.TotalSupply) == 0 {
		// it's possible that a token contract implements the ERC20 interfaces but does not
		// return any values; we use a backup in this case
		ret = &types.ERC20Metadata{
			Decimals:    []byte{0x0},
			Symbol:      "UNKNOWN",
			TotalSupply: []byte{0x0}}
	}

	return ret, err
}

func getGethSender(tx *gethtypes.Transaction) []byte {
	var from []byte
	sender, err := gethtypes.Sender(gethtypes.NewCancunSigner(tx.ChainId()), tx)
	if err != nil {
		from, _ = hex.DecodeString("abababababababababababababababababababab")
		log.Error(err, "error converting tx to msg", 0, map[string]interface{}{"tx": tx.Hash()})
	} else {
		from = sender.Bytes()
	}
	return from
}
