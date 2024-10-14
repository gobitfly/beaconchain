package db2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var ttl = 2 * time.Second

type EthClient interface {
	ethereum.ChainReader
	ethereum.ContractCaller
	bind.ContractBackend
	ethereum.ChainStateReader
}

type BigTableEthRaw struct {
	db RawStore

	chainID uint64

	// cache to store link between block hash and number
	// ethclient.Client.BlockByNumber retrieves the uncles by hash
	// so we need a way to access it simply
	// we also could use postgres db
	hashToNumber sync.Map
}

func NewBigTableEthRaw(db RawStore, chainID uint64) *BigTableEthRaw {
	return &BigTableEthRaw{
		db:           db,
		chainID:      chainID,
		hashToNumber: sync.Map{},
	}
}

func (r *BigTableEthRaw) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	var messages []*jsonrpcMessage
	var isSingle bool
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&messages); err != nil {
		isSingle = true
		message := new(jsonrpcMessage)
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	var resps []*jsonrpcMessage
	for _, message := range messages {
		resp, err := r.handle(request.Context(), message)
		if err != nil {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewBufferString(err.Error())),
				StatusCode: http.StatusBadRequest,
			}, nil
		}
		resps = append(resps, resp)
	}

	respBody, _ := makeBody(isSingle, resps)
	return &http.Response{
		Body:       respBody,
		StatusCode: http.StatusOK,
	}, nil
}

func (r *BigTableEthRaw) handle(ctx context.Context, message *jsonrpcMessage) (*jsonrpcMessage, error) {
	var args []interface{}
	if err := json.Unmarshal(message.Params, &args); err != nil {
		return nil, err
	}

	var respBody []byte
	switch message.Method {
	case "eth_getBlockByNumber":
		// we decode only big.Int maybe we should also handle "latest"
		block, err := hexutil.DecodeBig(args[0].(string))
		if err != nil {
			return nil, err
		}

		b, err := r.BlockByNumber(ctx, block)
		if err != nil {
			return nil, err
		}
		respBody = b

	case "debug_traceBlockByNumber":
		block, err := hexutil.DecodeBig(args[0].(string))
		if err != nil {
			return nil, err
		}

		b, err := r.TraceBlockByNumber(ctx, block)
		if err != nil {
			return nil, err
		}
		respBody = b

	case "eth_getBlockReceipts":
		block, err := hexutil.DecodeBig(args[0].(string))
		if err != nil {
			return nil, err
		}

		b, err := r.BlockReceipts(ctx, block)
		if err != nil {
			return nil, err
		}
		respBody = b

	case "eth_getUncleByBlockHashAndIndex":
		number, exist := r.hashToNumber.Load(args[0].(string))
		if !exist {
			return nil, fmt.Errorf("cannot find hash '%s' in cache", args[0].(string))
		}

		index, err := hexutil.DecodeBig(args[1].(string))
		if err != nil {
			return nil, err
		}
		b, err := r.UncleByBlockNumberAndIndex(ctx, number.(*big.Int), index.Int64())
		if err != nil {
			return nil, err
		}
		respBody = b
	}
	var resp jsonrpcMessage
	_ = json.Unmarshal(respBody, &resp)
	if len(respBody) == 0 {
		resp.Version = message.Version
		resp.Result = []byte("[]")
	}
	resp.ID = message.ID
	return &resp, nil
}

func makeBody(isSingle bool, messages []*jsonrpcMessage) (io.ReadCloser, error) {
	var b []byte
	var err error
	if isSingle {
		b, err = json.Marshal(messages[0])
	} else {
		b, err = json.Marshal(messages)
	}
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

func (r *BigTableEthRaw) BlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.db.ReadBlock(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	// retrieve the block hash for caching purpose
	var mini MinimalBlock
	if err := json.Unmarshal(block.Block, &mini); err != nil {
		return nil, err
	}
	r.hashToNumber.Store(mini.Result.Hash, number)
	go func(hash string) {
		time.Sleep(ttl)
		r.hashToNumber.Delete(hash)
	}(mini.Result.Hash)
	return block.Block, nil
}

func (r *BigTableEthRaw) BlockReceipts(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.db.ReadBlock(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Receipts, nil
}

func (r *BigTableEthRaw) TraceBlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.db.ReadBlock(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Traces, nil
}

func (r *BigTableEthRaw) UncleByBlockNumberAndIndex(ctx context.Context, number *big.Int, index int64) ([]byte, error) {
	block, err := r.db.ReadBlock(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}

	// len of one uncle without tx is 1473
	if len(block.Uncles) < 2000 {
		return block.Uncles, nil
	}
	// we have two uncles so we need to retrieve the good index
	var uncles []*jsonrpcMessage
	_ = json.Unmarshal(block.Uncles, &uncles)
	return json.Marshal(uncles[index])
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
