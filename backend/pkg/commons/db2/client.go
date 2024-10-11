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
	tr http.RoundTripper
	db RawStore

	// cache to store link between block hash and number
	// ethclient.Client.BlockByNumber retrieves the uncles by hash
	// so we need a way to access it simply
	// we also could use postgres db
	hashToNumber sync.Map
}

func NewBigTableEthRaw(tr http.RoundTripper, db RawStore) *BigTableEthRaw {
	return &BigTableEthRaw{
		tr:           tr,
		db:           db,
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
	var resps []jsonrpcMessage
	for _, message := range messages {
		var respBody []byte
		switch message.Method {
		case "eth_getBlockByNumber":
			var args []interface{}
			if err := json.Unmarshal(message.Params, &args); err != nil {
				return nil, err
			}

			// we decode only big.Int maybe we should also handle "latest"
			block, err := hexutil.DecodeBig(args[0].(string))
			if err != nil {
				return nil, err
			}

			b, err := r.BlockByNumber(request.Context(), block)
			if err != nil {
				return nil, err
			}
			respBody = b

		case "debug_traceBlockByNumber":
			var args []interface{}
			if err := json.Unmarshal(message.Params, &args); err != nil {
				return nil, err
			}

			block, err := hexutil.DecodeBig(args[0].(string))
			if err != nil {
				return nil, err
			}

			b, err := r.TraceBlockByNumber(request.Context(), block)
			if err != nil {
				return nil, err
			}
			respBody = b

		case "eth_getBlockReceipts":
			var args []interface{}
			if err := json.Unmarshal(message.Params, &args); err != nil {
				return nil, err
			}

			block, err := hexutil.DecodeBig(args[0].(string))
			if err != nil {
				return nil, err
			}

			b, err := r.BlockReceipts(request.Context(), block)
			if err != nil {
				return nil, err
			}
			respBody = b

		case "eth_getUncleByBlockHashAndIndex":
			var args []interface{}
			if err := json.Unmarshal(message.Params, &args); err != nil {
				return nil, err
			}
			number, exist := r.hashToNumber.Load(args[0].(string))
			if !exist {
				return nil, fmt.Errorf("cannot find hash '%s' in cache", args[0].(string))
			}

			index, err := hexutil.DecodeBig(args[1].(string))
			if err != nil {
				return nil, err
			}
			// TODO handle index
			b, err := r.UncleByBlockNumberAndIndex(request.Context(), number.(*big.Int), index.Int64())
			if err != nil {
				return nil, err
			}
			respBody = b
		}
		if len(respBody) == 0 {
			continue
		}
		var resp jsonrpcMessage
		_ = json.Unmarshal(respBody, &resp)
		resp.ID = message.ID
		resps = append(resps, resp)
	}

	if len(resps) == 0 {
		return r.tr.RoundTrip(request)
	}
	return &http.Response{
		Body:       makeBody(isSingle, resps),
		StatusCode: http.StatusOK,
	}, nil
}

func makeBody(isSingle bool, messages []jsonrpcMessage) io.ReadCloser {
	var b []byte
	if isSingle {
		b, _ = json.Marshal(messages[0])
	} else {
		b, _ = json.Marshal(messages)
	}
	return io.NopCloser(bytes.NewReader(b))
}

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

func (r *BigTableEthRaw) BlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.db.ReadBlock(1, number.Int64())
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
	block, err := r.db.ReadBlock(1, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Receipts, nil
}

func (r *BigTableEthRaw) TraceBlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.db.ReadBlock(1, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Traces, nil
}

func (r *BigTableEthRaw) UncleByBlockNumberAndIndex(ctx context.Context, number *big.Int, index int64) ([]byte, error) {
	block, err := r.db.ReadBlock(1, number.Int64())
	if err != nil {
		return nil, err
	}
	// TODO handle index
	return block.Uncles, nil
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
