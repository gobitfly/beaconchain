package raw

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/jsonrpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types/geth"
)

type FullGethBlock struct {
	block    *types.Block
	receipts []*types.Receipt
	traces   []*geth.Trace
}

var GethParse = func(rawBlock *FullBlockData) (*types.Block, []*types.Receipt, []*geth.Trace, error) {
	var blockResp, receiptsResp, tracesResp jsonrpc.Message
	_ = json.Unmarshal(rawBlock.Receipts, &receiptsResp)
	_ = json.Unmarshal(rawBlock.Block, &blockResp)
	_ = json.Unmarshal(rawBlock.Traces, &tracesResp)

	var unclesResp []jsonrpc.Message
	_ = json.Unmarshal(rawBlock.Uncles, &unclesResp)

	block, err := parseEthBlock(blockResp.Result, unclesResp)
	if err != nil {
		return nil, nil, nil, err
	}

	var receipts []*types.Receipt
	var traces []*geth.Trace
	if len(block.Transactions()) != 0 {
		if err := json.Unmarshal(receiptsResp.Result, &receipts); err != nil {
			return nil, nil, nil, err
		}

		if err := json.Unmarshal(tracesResp.Result, &traces); err != nil {
			return nil, nil, nil, err
		}

		for i := 0; i < len(block.Transactions()); i++ {
			if traces[i].TxHash != "" {
				break
			}
			// manually insert the hash in case it is missing
			// ie: old block traces don't include the hash
			traces[i].TxHash = receipts[i].TxHash.Hex()
		}
		// manually insert the transaction position
		for i := 0; i < len(traces); i++ {
			calls := []*geth.TraceCall{traces[i].Result}
			for len(calls) != 0 {
				calls[0].TransactionPosition = i
				calls = append(calls[1:], calls[0].Calls...)
			}
		}
	}

	return block, receipts, traces, nil
}

type rpcBlock struct {
	Hash         common.Hash         `json:"hash"`
	Transactions []rpcTransaction    `json:"transactions"`
	UncleHashes  []common.Hash       `json:"uncles"`
	Withdrawals  []*types.Withdrawal `json:"withdrawals,omitempty"`
	Requests     []*types.Request    `json:"requests,omitempty"`
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

// parseEthBlock is a copy of ethclient.Client.getBlock
// modified to work the with raw db
// https://github.com/ethereum/go-ethereum/blob/v1.14.11/ethclient/ethclient.go#L129
func parseEthBlock(raw json.RawMessage, rawUncles []jsonrpc.Message) (*types.Block, error) {
	// Decode header and transactions.
	var head *types.Header
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}

	// When the block is not found, the API returns JSON null.
	if head == nil {
		return nil, ethereum.NotFound
	}

	var body rpcBlock
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	// Quick-verify transaction and uncle lists. This mostly helps with debugging the server.
	if head.UncleHash == types.EmptyUncleHash && len(body.UncleHashes) > 0 {
		return nil, errors.New("server returned non-empty uncle list but block header indicates no uncles")
	}
	if head.UncleHash != types.EmptyUncleHash && len(body.UncleHashes) == 0 {
		return nil, errors.New("server returned empty uncle list but block header indicates uncles")
	}
	if head.TxHash == types.EmptyTxsHash && len(body.Transactions) > 0 {
		return nil, errors.New("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.TxHash != types.EmptyTxsHash && len(body.Transactions) == 0 {
		return nil, errors.New("server returned empty transaction list but block header indicates transactions")
	}
	// Load uncles because they are not included in the block response.
	uncles := make([]*types.Header, len(body.UncleHashes))
	for i := 0; i < len(body.UncleHashes); i++ {
		err := json.Unmarshal(rawUncles[i].Result, &uncles[i])
		if err != nil {
			return nil, err
		}
	}
	// Fill the sender cache of transactions in the block.
	txs := make([]*types.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		if tx.From != nil {
			setSenderFromDBSigner(tx.tx, *tx.From, body.Hash)
		}
		txs[i] = tx.tx
	}
	return types.NewBlockWithHeader(head).WithBody(
		types.Body{
			Transactions: txs,
			Uncles:       uncles,
			Withdrawals:  body.Withdrawals,
			Requests:     body.Requests,
		}), nil
}

// SenderFromDBSigner is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
// copy of senderFromServer
// https://github.com/ethereum/go-ethereum/blob/v1.14.11/ethclient/signer.go#L30
type SenderFromDBSigner struct {
	addr      common.Address
	Blockhash common.Hash
}

var errNotCached = errors.New("sender not cached")

func setSenderFromDBSigner(tx *types.Transaction, addr common.Address, block common.Hash) {
	// Use types.Sender for side-effect to store our signer into the cache.
	_, _ = types.Sender(&SenderFromDBSigner{addr, block}, tx)
}

func (s *SenderFromDBSigner) Equal(other types.Signer) bool {
	os, ok := other.(*SenderFromDBSigner)
	return ok && os.Blockhash == s.Blockhash
}

func (s *SenderFromDBSigner) Sender(tx *types.Transaction) (common.Address, error) {
	if s.addr == (common.Address{}) {
		return common.Address{}, errNotCached
	}
	return s.addr, nil
}

func (s *SenderFromDBSigner) ChainID() *big.Int {
	panic("can't sign with SenderFromDBSigner")
}
func (s *SenderFromDBSigner) Hash(tx *types.Transaction) common.Hash {
	panic("can't sign with SenderFromDBSigner")
}
func (s *SenderFromDBSigner) SignatureValues(tx *types.Transaction, sig []byte) (R, S, V *big.Int, err error) {
	panic("can't sign with SenderFromDBSigner")
}
