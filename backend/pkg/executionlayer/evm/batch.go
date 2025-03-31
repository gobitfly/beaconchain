package evm

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/chain"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

// multicallerFor source https://www.multicall3.com/deployments
var multicallerFor = map[string]common.Address{
	chain.IDs.Mainnet.String():    common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	chain.IDs.Sepolia.String():    common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	chain.IDs.Gnosis.String():     common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	chain.IDs.Optimistic.String(): common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	chain.IDs.Arbitrum.String():   common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	chain.IDs.Holesky.String():    common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	// chain.IDs.Hoodi.String():    common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"), // TODO hoodi
}

// Batcher only supports eth_getBalance and eth_call for now
// feel free to add more supported methods when necessary
type Batcher interface {
	Batch(elements []BatchElement) ([]BatchResponse, error)
}

type BatcherConfig struct {
	MulticallAddress *common.Address
	Limit            int
}

func NewBatcher(chainID *big.Int, client *ethclient.Client, config BatcherConfig) Batcher {
	multicall := multicallerFor[chainID.String()]
	if config.MulticallAddress != nil {
		multicall = *config.MulticallAddress
	}
	if multicall.Cmp(common.Address{}) != 0 {
		return NewMulticallBatcher(client, multicall, config.Limit)
	}
	log.WarnWithFields(map[string]interface{}{
		"chainID": chainID,
	}, "multicall unavailable for chain, falling back to RPCBatcher")
	return NewRPCBatch(client.Client(), config.Limit)
}

type BatchElement struct {
	Method string
	To     *common.Address
	Data   []byte
}

type BatchResponse struct {
	Data   []byte
	Failed bool
}

// on mainnet it closer to 7_000 for balance retrieval
// 5_000 is good compromise between perf and potential higher gas consuming calls
// if this start to impact performances, we will need to have a limit per network (different block gas limits)
// and potentially per type of call
const multicallDefaultLimit = 5_000

// the limit is determined by the total body size
// for balance retrieval eth is closer to 300k but erc20 is closer to 130K
// 100_000 is a good compromise between perf and potential calls with longer data input
const rpcDefaultLimit = 100_000

type MulticallBatcher struct {
	multicall    *contracts.IMulticall3Caller
	multicallABI abi.ABI
	address      common.Address

	// limit to remove out of gas error that can happens otherwise
	limit int
}

func NewMulticallBatcher(client bind.ContractBackend, multicallAddress common.Address, limit int) MulticallBatcher {
	multicall, _ := contracts.NewIMulticall3Caller(multicallAddress, client)
	multicallABI, _ := abi.JSON(strings.NewReader(contracts.IMulticall3MetaData.ABI))
	if limit == 0 {
		limit = multicallDefaultLimit
	}
	return MulticallBatcher{
		multicall:    multicall,
		multicallABI: multicallABI,
		address:      multicallAddress,
		limit:        limit,
	}
}

func (m MulticallBatcher) Batch(elements []BatchElement) ([]BatchResponse, error) {
	return batchWithLimit(elements, m.limit, m.doBatch)
}

func (m MulticallBatcher) doBatch(elements []BatchElement) ([]BatchResponse, error) {
	var batch []contracts.IMulticall3Call3
	var res []BatchResponse

	for _, element := range elements {
		switch element.Method {
		case "eth_getBalance":
			// in case of eth_getBalance we redirect the call to multicall.getEthBalance
			element.To = &m.address
			data, err := m.multicallABI.Pack("getEthBalance", common.BytesToAddress(element.Data))
			if err != nil {
				return nil, fmt.Errorf("cannot pack getEthBalance: %w", err)
			}
			element.Data = data
		case "eth_call":
		default:
			return nil, fmt.Errorf("MulticallBatcher: method %s not supported", element.Method)
		}
		batch = append(batch, contracts.IMulticall3Call3{
			Target:       *element.To,
			AllowFailure: true,
			CallData:     element.Data,
		})
	}

	results, err := m.multicall.Aggregate3(nil, batch)
	if err != nil {
		return nil, fmt.Errorf("multicall failed: %v", err)
	}

	for _, elem := range results {
		res = append(res, BatchResponse{
			Data:   elem.ReturnData,
			Failed: elem.Success,
		})
	}
	return res, nil
}

type RPCBatcher struct {
	client *rpc.Client
	limit  int
}

func NewRPCBatch(client *rpc.Client, limit int) RPCBatcher {
	if limit == 0 {
		limit = rpcDefaultLimit
	}
	return RPCBatcher{
		client: client,
		limit:  limit,
	}
}

func (m RPCBatcher) Batch(elements []BatchElement) ([]BatchResponse, error) {
	return batchWithLimit(elements, m.limit, m.doBatch)
}

func (m RPCBatcher) doBatch(elements []BatchElement) ([]BatchResponse, error) {
	var batch []rpc.BatchElem
	var res []BatchResponse

	for _, element := range elements {
		result := ""
		var args []interface{}
		switch element.Method {
		case "eth_getBalance":
			args = []interface{}{hexutil.Encode(element.Data)}
		case "eth_call":
			args = []interface{}{toCallArg(ethereum.CallMsg{
				To:   element.To,
				Data: element.Data,
			})}
		default:
			return nil, fmt.Errorf("RPCBatcher: method %s not supported", element.Method)
		}
		batch = append(batch, rpc.BatchElem{
			Method: element.Method,
			Args:   append(args, "latest"),
			Result: &result,
		})
	}

	if err := m.client.BatchCall(batch); err != nil {
		return nil, fmt.Errorf("batch call failed: %v", err)
	}

	for _, elem := range batch {
		res = append(res, BatchResponse{
			Data:   common.FromHex(*elem.Result.(*string)),
			Failed: elem.Error != nil,
		})
	}
	return res, nil
}

func ERC20Supply(batcher Batcher, addresses []common.Address) ([]*big.Int, error) {
	var elements []BatchElement
	for _, address := range addresses {
		input, _ := erc20.ABI.Pack("totalSupply")
		elements = append(elements, BatchElement{
			Method: "eth_call",
			To:     &address,
			Data:   input,
		})
	}
	results, err := batcher.Batch(elements)
	if err != nil {
		return nil, err
	}
	var supplies []*big.Int
	for _, result := range results {
		supplies = append(supplies, new(big.Int).SetBytes(result.Data))
	}
	return supplies, nil
}

func BalanceForPairs(batcher Batcher, pairs []db2.Pair) ([]*big.Int, error) {
	var elements []BatchElement
	for _, pair := range pairs {
		elem := BatchElement{
			Method: "eth_getBalance",
			Data:   pair.Address.Bytes(),
		}
		if pair.Token.Cmp(common.Address{}) != 0 {
			elem.Method = "eth_call"
			elem.Data, _ = erc20.ABI.Pack("balanceOf", pair.Address)
			elem.To = &pair.Token
		}
		elements = append(elements, elem)
	}
	results, err := batcher.Batch(elements)
	if err != nil {
		return nil, err
	}
	var balances []*big.Int
	for _, result := range results {
		balances = append(balances, new(big.Int).SetBytes(result.Data))
	}
	return balances, nil
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

func batchWithLimit(elements []BatchElement, limit int, doBatch func(elements []BatchElement) ([]BatchResponse, error)) ([]BatchResponse, error) {
	var res []BatchResponse
	for i := 0; i < len(elements); i += limit {
		end := i + limit
		if end > len(elements) {
			end = len(elements)
		}
		partial, err := doBatch(elements[i:end])
		if err != nil {
			return nil, err
		}
		res = append(res, partial...)
	}
	return res, nil
}
