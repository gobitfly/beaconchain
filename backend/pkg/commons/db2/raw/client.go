package raw

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/jsonrpc"
)

var ErrNotFoundInCache = fmt.Errorf("cannot find hash in cache")
var ErrMethodNotSupported = fmt.Errorf("method not supported")

type StoreReader interface {
	ReadBlockByNumber(chainID uint64, number int64) (*FullBlockData, error)
	ReadBlockByHash(chainID uint64, hash string) (*FullBlockData, error)
	ReadBlocksByNumber(chainID uint64, start, end int64) ([]*FullBlockData, error)
}

type WithFallback struct {
	roundTripper http.RoundTripper
	fallback     http.RoundTripper
}

func NewWithFallback(roundTripper, fallback http.RoundTripper) *WithFallback {
	return &WithFallback{
		roundTripper: roundTripper,
		fallback:     fallback,
	}
}

func (r WithFallback) RoundTrip(request *http.Request) (*http.Response, error) {
	resp, err := r.roundTripper.RoundTrip(request)
	if err == nil {
		// no fallback needed
		return resp, nil
	}

	var e1 *json.SyntaxError
	if !errors.As(err, &e1) &&
		!errors.Is(err, ErrNotFoundInCache) &&
		!errors.Is(err, ErrMethodNotSupported) &&
		!errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	return r.fallback.RoundTrip(request)
}

type StoreRoundTripper struct {
	store   StoreReader
	chainID uint64
}

func NewBigTableEthRaw(store StoreReader, chainID uint64) *StoreRoundTripper {
	return &StoreRoundTripper{
		store:   store,
		chainID: chainID,
	}
}

func (r *StoreRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(body))
	}()
	var messages []*jsonrpc.Message
	var isSingle bool
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&messages); err != nil {
		isSingle = true
		message := new(jsonrpc.Message)
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	var resps []*jsonrpc.Message
	for _, message := range messages {
		resp, err := r.handle(request.Context(), message)
		if err != nil {
			return nil, err
		}
		resps = append(resps, resp)
	}

	respBody, _ := makeBody(isSingle, resps)
	return &http.Response{
		Body:       respBody,
		StatusCode: http.StatusOK,
	}, nil
}

func (r *StoreRoundTripper) handle(ctx context.Context, message *jsonrpc.Message) (*jsonrpc.Message, error) {
	var args []interface{}
	err := json.Unmarshal(message.Params, &args)
	if err != nil {
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

		respBody, err = r.BlockByNumber(ctx, block)
		if err != nil {
			return nil, err
		}

	case "debug_traceBlockByNumber":
		block, err := hexutil.DecodeBig(args[0].(string))
		if err != nil {
			return nil, err
		}

		respBody, err = r.TraceBlockByNumber(ctx, block)
		if err != nil {
			return nil, err
		}

	case "eth_getBlockReceipts":
		block, err := hexutil.DecodeBig(args[0].(string))
		if err != nil {
			return nil, err
		}

		respBody, err = r.BlockReceipts(ctx, block)
		if err != nil {
			return nil, err
		}

	case "eth_getUncleByBlockHashAndIndex":
		index, err := hexutil.DecodeBig(args[1].(string))
		if err != nil {
			return nil, err
		}
		respBody, err = r.UncleByBlockHashAndIndex(ctx, args[0].(string), index.Int64())
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrMethodNotSupported
	}
	var resp jsonrpc.Message
	_ = json.Unmarshal(respBody, &resp)
	if len(respBody) == 0 {
		resp.Version = message.Version
		resp.Result = []byte("[]")
	}
	resp.ID = message.ID
	return &resp, nil
}

func makeBody(isSingle bool, messages []*jsonrpc.Message) (io.ReadCloser, error) {
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

func (r *StoreRoundTripper) BlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.store.ReadBlockByNumber(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Block, nil
}

func (r *StoreRoundTripper) BlockReceipts(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.store.ReadBlockByNumber(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Receipts, nil
}

func (r *StoreRoundTripper) TraceBlockByNumber(ctx context.Context, number *big.Int) ([]byte, error) {
	block, err := r.store.ReadBlockByNumber(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	return block.Traces, nil
}

func (r *StoreRoundTripper) UncleByBlockNumberAndIndex(ctx context.Context, number *big.Int, index int64) ([]byte, error) {
	block, err := r.store.ReadBlockByNumber(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}

	var uncles []*jsonrpc.Message
	if err := json.Unmarshal(block.Uncles, &uncles); err != nil {
		var uncle *jsonrpc.Message
		if err := json.Unmarshal(block.Uncles, &uncle); err != nil {
			return nil, fmt.Errorf("cannot unmarshal uncle: %w", err)
		}
		return json.Marshal(uncle)
	}
	return json.Marshal(uncles[index])
}

func (r *StoreRoundTripper) UncleByBlockHashAndIndex(ctx context.Context, hash string, index int64) ([]byte, error) {
	block, err := r.store.ReadBlockByHash(r.chainID, hash)
	if err != nil {
		return nil, err
	}

	var uncles []*jsonrpc.Message
	if err := json.Unmarshal(block.Uncles, &uncles); err != nil {
		var uncle *jsonrpc.Message
		if err := json.Unmarshal(block.Uncles, &uncle); err != nil {
			return nil, fmt.Errorf("cannot unmarshal uncle: %w", err)
		}
		return json.Marshal(uncle)
	}
	return json.Marshal(uncles[index])
}
