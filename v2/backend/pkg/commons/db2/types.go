package db2

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type ContractUpdateWithAddress struct {
	Indexed       *types.IsContractUpdate
	Address       []byte
	TxIndex       int
	InternalIndex int
}

type Pair struct {
	Address common.Address
	Token   common.Address
}

type Balance struct {
	Pair
	Value *big.Int
}

type IndexedBlock struct {
	ChainID               string
	Hash                  []byte
	Number                uint64
	Block                 *types.Eth1BlockIndexed
	Transactions          []*types.Eth1TransactionIndexed
	Internals             []InternalWithIndexes
	ERC20Transfer         []TransferWithIndexes
	ERC1155Transfer       []ERC1155TransferWithIndexes
	ERC721Transfer        []ERC721TransferWithIndexes
	Blobs                 []BlobWithIndex
	Uncles                []UncleWithIndexes
	Withdrawals           []*types.Eth1WithdrawalIndexed
	ENS                   []ENSLog
	Contracts             []ContractUpdateWithAddress
	ConsolidationRequests []ConsolidationRequest
	WithdrawalRequests    []WithdrawalRequest
}

type BridgeQueueRequest struct {
	TxHash         []byte
	TxIndex        int
	ItxIndex       int
	BlockNumber    uint64
	BlockTimestamp time.Time
	From           []byte
	Fee            int64
}

type ConsolidationRequest struct {
	SourceAddress []byte
	SourcePubKey  []byte
	TargetPubKey  []byte
	BridgeQueueRequest
}

type WithdrawalRequest struct {
	SourceAddress   []byte
	ValidatorPubKey []byte
	Amount          uint64
	BridgeQueueRequest
}

type TransferWithIndexes struct {
	Indexed  *types.Eth1ERC20Indexed
	TxIndex  int
	LogIndex int
}

type InternalWithIndexes struct {
	Indexed       *types.Eth1InternalTransactionIndexed
	TxIndex       int
	InternalIndex int
	Path          string
}

type ERC1155TransferWithIndexes struct {
	Indexed  *types.ETh1ERC1155Indexed
	TxIndex  int
	LogIndex int
}

type ERC721TransferWithIndexes struct {
	Indexed  *types.Eth1ERC721Indexed
	TxIndex  int
	LogIndex int
}

type BlobWithIndex struct {
	Indexed *types.Eth1BlobTransactionIndexed
	TxIndex int
}

type UncleWithIndexes struct {
	Indexed   *types.Eth1UncleIndexed
	Index     int
	Coinbase  []byte
	BlockTime *timestamppb.Timestamp
}

type ENSLog struct {
	Node  *[32]byte
	Name  *string
	Owner *common.Address
}
