package raw

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/chain"
)

type noOpSigner struct{}

func (s noOpSigner) Equal(other types.Signer) bool {
	panic("can't compare with noOpSigner")
}
func (s noOpSigner) Sender(tx *types.Transaction) (common.Address, error) {
	panic("can't get sender with noOpSigner")
}
func (s noOpSigner) ChainID() *big.Int {
	panic("can't sign with noOpSigner")
}
func (s noOpSigner) Hash(tx *types.Transaction) common.Hash {
	panic("can't sign with noOpSigner")
}
func (s noOpSigner) SignatureValues(tx *types.Transaction, sig []byte) (R, S, V *big.Int, err error) {
	panic("can't sign with noOpSigner")
}

// SenderFromDBSigner is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
// inspired by senderFromServer from go-ethereum
// https://github.com/ethereum/go-ethereum/blob/v1.14.11/ethclient/signer.go#L30
type SenderFromDBSigner struct {
	addr      common.Address
	blockHash common.Hash
	noOpSigner
}

var errNotCached = errors.New("sender not cached")

func (s *SenderFromDBSigner) Equal(other types.Signer) bool {
	os, ok := other.(*SenderFromDBSigner)
	return ok && os.blockHash == s.blockHash
}

func (s *SenderFromDBSigner) Sender(tx *types.Transaction) (common.Address, error) {
	if s.addr == (common.Address{}) {
		return common.Address{}, errNotCached
	}
	return s.addr, nil
}

// setSender doesn't rely on tx.ChainId() because it introduces problems for testing
// tx.ChainId() come from the signature, and we cannot sign for some sender for certain chain ID
// see tests for examples
func setSender(chainID *big.Int, tx *types.Transaction, addr common.Address, block common.Hash) {
	switch {
	case chainID.Cmp(chain.IDs.Optimistic) == 0:
		_, _ = types.Sender(&OptimisticSigner{addr: &addr, blockHash: block}, tx)
	default:
		_, _ = types.Sender(&SenderFromDBSigner{addr: addr, blockHash: block}, tx)
	}
}

func SignerForChainID(chainID *big.Int, blockHash common.Hash) types.Signer {
	switch {
	case chainID.Cmp(chain.IDs.Optimistic) == 0:
		return &OptimisticSigner{
			blockHash: blockHash,
		}
	default:
		return &SenderFromDBSigner{
			blockHash: blockHash,
		}
	}
}

type OptimisticSigner struct {
	addr      *common.Address
	blockHash common.Hash
	noOpSigner
}

func (s *OptimisticSigner) Equal(other types.Signer) bool {
	os, ok := other.(*OptimisticSigner)
	return ok && os.blockHash == s.blockHash
}

func (s *OptimisticSigner) Sender(tx *types.Transaction) (common.Address, error) {
	// optimism can have transaction from address 0
	// https://docs.optimism.io/stack/transactions/deposit-flow#l1-processing
	if s.addr == nil {
		return common.Address{}, errNotCached
	}
	return *s.addr, nil
}
