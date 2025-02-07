package executionlayer

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/chain"
	"github.com/gobitfly/beaconchain/pkg/commons/contracts/ens"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/commons/erc1155"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var (
	alice        = []byte("alice")
	aliceAddress = common.BytesToAddress(leftPad(alice, 20))
	bob          = []byte("bob")
	bobAddress   = common.BytesToAddress(leftPad(bob, 20))
	contract     = []byte("contract")
	usdc         = []byte("usdc")
)

func TestTransformTX(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1TransactionIndexed
	}{
		{
			name: "normal transaction",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						From: alice,
						To:   bob,
					},
				},
			},
			want: []*types.Eth1TransactionIndexed{
				{
					From: alice,
					To:   bob,
				},
			},
		},
		{
			name: "transaction with contract",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						ContractAddress: contract,
					},
				},
			},
			want: []*types.Eth1TransactionIndexed{
				{
					IsContractCreation: true,
					To:                 contract,
				},
			},
		},
		{
			name: "transaction with method",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Data: []byte("123456789"),
					},
				},
			},
			want: []*types.Eth1TransactionIndexed{
				{
					MethodId: []byte("1234"),
				},
			},
		},
		{
			name: "success transaction",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Status: 1,
					},
				},
			},
			want: []*types.Eth1TransactionIndexed{
				{
					Status: types.StatusType_SUCCESS,
				},
			},
		},
		{
			name: "partially failed transaction",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Status: 1,
						Itx: []*types.Eth1InternalTransaction{
							{
								ErrorMsg: "fail",
							},
						},
					},
				},
			},
			want: []*types.Eth1TransactionIndexed{
				{
					Status: types.StatusType_PARTIAL,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformTx("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Transactions), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Transactions[i].From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Transactions[i].To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Transactions[i].Status, indexed.Status; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Transactions[i].IsContractCreation, indexed.IsContractCreation; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Transactions[i].MethodId, indexed.MethodId; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformERC20(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1ERC20Indexed
	}{
		{
			name: "normal transfer",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: usdc,
								Data:    leftPad([]byte{1}, 32),
								Topics: [][]byte{
									erc20.TransferTopic.Bytes(),
									alice,
									bob,
								},
							},
						},
					},
				},
			},
			want: []*types.Eth1ERC20Indexed{
				{
					TokenAddress: usdc,
					From:         aliceAddress.Bytes(),
					To:           bobAddress.Bytes(),
					Value:        []byte{1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformERC20("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.ERC20Transfer), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.ERC20Transfer[i].Indexed.From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC20Transfer[i].Indexed.To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC20Transfer[i].Indexed.TokenAddress, indexed.TokenAddress; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC20Transfer[i].Indexed.Value, indexed.Value; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformERC721(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1ERC721Indexed
	}{
		{
			name: "normal transfer",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: contract,
								Topics: [][]byte{
									erc721.TransferTopic.Bytes(),
									alice,
									bob,
									leftPad([]byte{1}, 32),
								},
							},
						},
					},
				},
			},
			want: []*types.Eth1ERC721Indexed{
				{
					TokenAddress: contract,
					From:         aliceAddress.Bytes(),
					To:           bobAddress.Bytes(),
					TokenId:      []byte{1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformERC721("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.ERC721Transfer), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.ERC721Transfer[i].Indexed.From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC721Transfer[i].Indexed.To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC721Transfer[i].Indexed.TokenAddress, indexed.TokenAddress; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC721Transfer[i].Indexed.TokenId, indexed.TokenId; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformERC1155(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.ETh1ERC1155Indexed
	}{
		{
			name: "transfer single",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: contract,
								Data: bytes.Join([][]byte{
									leftPad([]byte{1}, 32),
									leftPad([]byte{2}, 32),
								}, nil),
								Topics: [][]byte{
									erc1155.TransferSingleTopic.Bytes(),
									alice,
									alice,
									bob,
								},
							},
						},
					},
				},
			},
			want: []*types.ETh1ERC1155Indexed{
				{
					TokenAddress: contract,
					Operator:     aliceAddress.Bytes(),
					From:         aliceAddress.Bytes(),
					To:           bobAddress.Bytes(),
					TokenId:      []byte{1},
					Value:        []byte{2},
				},
			},
		},
		// TODO add transfer bulk when bug fixed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformERC1155("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.ERC1155Transfer), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.ERC1155Transfer[i].Indexed.From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC1155Transfer[i].Indexed.To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC1155Transfer[i].Indexed.TokenAddress, indexed.TokenAddress; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC1155Transfer[i].Indexed.TokenId, indexed.TokenId; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.ERC1155Transfer[i].Indexed.Value, indexed.Value; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestBlob(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1BlobTransactionIndexed
	}{
		{
			name: "normal blob",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Type: gethtypes.BlobTxType,
						From: alice,
						To:   bob,
					},
				},
			},
			want: []*types.Eth1BlobTransactionIndexed{
				{
					From: alice,
					To:   bob,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformBlob("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Blobs), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Blobs[i].Indexed.From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Blobs[i].Indexed.To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Blobs[i].Indexed.Value, indexed.Value; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformITx(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1InternalTransactionIndexed
	}{
		{
			name: "normal internal",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								From:  alice,
								To:    bob,
								Value: []byte{1},
							},
						},
					},
				},
			},
			want: []*types.Eth1InternalTransactionIndexed{
				{
					From: alice,
					To:   bob,
				},
			},
		},
		{
			name: "ignore without value",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Value: []byte{0},
							},
						},
					},
				},
			},
		},
		{
			name: "ignore without top call - geth",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Value: []byte{0},
								Path:  "0",
							},
						},
					},
				},
			},
		},
		{
			name: "ignore without top call - parity",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Value: []byte{0},
								Path:  "[]",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformITx("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Internals), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Internals[i].Indexed.From, indexed.From; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Internals[i].Indexed.To, indexed.To; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformContracts(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []metadataupdates.ContractUpdateWithAddress
	}{
		{
			name: "create",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Type: "create",
								To:   contract,
							},
						},
					},
				},
			},
			want: []metadataupdates.ContractUpdateWithAddress{
				{
					Address: contract,
					Indexed: &types.IsContractUpdate{
						IsContract: true,
						Success:    true,
					},
				},
			},
		},
		{
			name: "failed",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Type:     "create",
								To:       contract,
								ErrorMsg: "failed",
							},
						},
					},
				},
			},
			want: []metadataupdates.ContractUpdateWithAddress{
				{
					Address: contract,
					Indexed: &types.IsContractUpdate{
						IsContract: true,
						Success:    false,
					},
				},
			},
		},
		{
			name: "suicide",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								Type: "suicide",
								From: contract,
							},
						},
					},
				},
			},
			want: []metadataupdates.ContractUpdateWithAddress{
				{
					Address: contract,
					Indexed: &types.IsContractUpdate{
						IsContract: false,
						Success:    true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformContract("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Contracts), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Contracts[i].Address, indexed.Address; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Contracts[i].Indexed.IsContract, indexed.Indexed.IsContract; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Contracts[i].Indexed.Success, indexed.Indexed.Success; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformBlock(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  *types.Eth1BlockIndexed
	}{
		{
			name: "normal",
			block: &types.Eth1Block{
				Coinbase: alice,
				Number:   1,
			},
			want: &types.Eth1BlockIndexed{
				Coinbase: alice,
				Number:   1,
			},
		},
		{
			name: "with tx",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						GasPrice: []byte{1},
					},
					{
						GasPrice: []byte{2},
					},
				},
			},
			want: &types.Eth1BlockIndexed{
				HighestGasPrice: []byte{2},
				LowestGasPrice:  []byte{1},
			},
		},
		{
			name: "with blob",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Type: gethtypes.BlobTxType,
					},
				},
			},
			want: &types.Eth1BlockIndexed{
				BlobTransactionCount: 1,
			},
		},
		{
			name: "with mev",
			block: &types.Eth1Block{
				Coinbase: alice,
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								To:    alice,
								Value: []byte{1},
							},
						},
					},
				},
			},
			want: &types.Eth1BlockIndexed{
				Coinbase: alice,
				Mev:      []byte{1},
			},
		},
		{
			name: "with uncle",
			block: &types.Eth1Block{
				Coinbase: alice,
				Number:   1,
				Uncles: []*types.Eth1Block{
					{
						Number: 1,
					},
				},
				Difficulty: []byte{32},
			},
			want: &types.Eth1BlockIndexed{
				Coinbase:    alice,
				Number:      1,
				UncleReward: big.NewInt(62500000000000000).Bytes(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformBlock("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := res.Block.Number, tt.want.Number; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := res.Block.Coinbase, tt.want.Coinbase; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := res.Block.UncleReward, tt.want.UncleReward; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := res.Block.HighestGasPrice, tt.want.HighestGasPrice; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := res.Block.LowestGasPrice, tt.want.LowestGasPrice; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestTransformUncle(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*data.UncleWithIndexes
	}{
		{
			name: "normal uncle",
			block: &types.Eth1Block{
				Number: 1,
				Uncles: []*types.Eth1Block{
					{
						Number:   42,
						Coinbase: alice,
					},
				},
			},
			want: []*data.UncleWithIndexes{
				{
					Indexed: &types.Eth1UncleIndexed{
						BlockNumber: 1,
						Number:      42,
					},
					Coinbase: alice,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformUncle("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Uncles), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Uncles[i].Indexed.BlockNumber, indexed.Indexed.BlockNumber; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Uncles[i].Indexed.Number, indexed.Indexed.Number; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Uncles[i].Coinbase, indexed.Coinbase; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformWithdrawal(t *testing.T) {
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []*types.Eth1WithdrawalIndexed
	}{
		{
			name: "normal withdrawal",
			block: &types.Eth1Block{
				Number: 42,
				Withdrawals: []*types.Eth1Withdrawal{
					{
						Index:          1,
						ValidatorIndex: 2,
						Address:        alice,
						Amount:         []byte{1},
					},
				},
			},
			want: []*types.Eth1WithdrawalIndexed{
				{
					BlockNumber:    42,
					Index:          1,
					ValidatorIndex: 2,
					Address:        alice,
					Amount:         []byte{1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformWithdrawal("", tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.Withdrawals), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if got, want := res.Withdrawals[i].BlockNumber, indexed.BlockNumber; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Withdrawals[i].Index, indexed.Index; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Withdrawals[i].ValidatorIndex, indexed.ValidatorIndex; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Withdrawals[i].Address, indexed.Address; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := res.Withdrawals[i].Amount, indexed.Amount; !bytes.Equal(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestTransformENS(t *testing.T) {
	name := "ensName"
	chainID := chain.IDs.Mainnet
	tests := []struct {
		name  string
		block *types.Eth1Block
		want  []data.ENSLog
	}{
		{
			name: "ens registry - new resolver",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumRegistry).Bytes(),
								Data:    leftPad([]byte{}, 32),
								Topics: [][]byte{
									ens.RegistryNewResolverTopic.Bytes(),
									leftPad([]byte("node"), 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Node: (*[32]byte)(leftPad([]byte("node"), 32)),
				},
			},
		},
		{
			name: "ens registry - new owner",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumRegistry).Bytes(),
								Data:    leftPad(alice, 32),
								Topics: [][]byte{
									ens.RegistryNewOwnerTopic.Bytes(),
									leftPad([]byte{}, 32),
									leftPad([]byte{}, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Owner: &aliceAddress,
				},
			},
		},
		{
			name: "ens registry - new ttl",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumRegistry).Bytes(),
								Data:    leftPad([]byte{}, 32),
								Topics: [][]byte{
									ens.RegistryNewTTLTopic.Bytes(),
									leftPad([]byte("node"), 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Node: (*[32]byte)(leftPad([]byte("node"), 32)),
				},
			},
		},
		{
			name: "ens registrar controller - NameRegistered",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumRegistrarController).Bytes(),
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000080" +
											"0000000000000000000000000000000000000000000000000000000000000000" + // baseCost
											"0000000000000000000000000000000000000000000000000000000000000000" + // premium
											"0000000000000000000000000000000000000000000000000000000000000000" + // expires
											"0000000000000000000000000000000000000000000000000000000000000007",
									),
									rightPad([]byte(name)),
								}, nil),
								Topics: [][]byte{
									ens.RegistrarControllerNameRegisteredTopic.Bytes(),
									leftPad([]byte{}, 32),
									leftPad(alice, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Name:  &name,
					Owner: &aliceAddress,
				},
			},
		},
		{
			name: "ens registrar controller - NameRenewed",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumRegistrarController).Bytes(),
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000060" +
											"0000000000000000000000000000000000000000000000000000000000000000" + // cost
											"0000000000000000000000000000000000000000000000000000000000000000" + // expires
											"0000000000000000000000000000000000000000000000000000000000000007",
									),
									rightPad([]byte(name)),
								}, nil),
								Topics: [][]byte{
									ens.RegistrarControllerNameRenewedTopic.Bytes(),
									leftPad([]byte{}, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Name: &name,
				},
			},
		},
		{
			name: "old ens registrar controller - NameRegistered",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumOldEnsRegistrarController).Bytes(),
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000060" +
											"0000000000000000000000000000000000000000000000000000000000000000" + // cost
											"0000000000000000000000000000000000000000000000000000000000000000" + // expires
											"0000000000000000000000000000000000000000000000000000000000000007",
									),
									rightPad([]byte(name)),
								}, nil),
								Topics: [][]byte{
									ens.OldRegistrarControllerNameRegisteredTopic.Bytes(),
									leftPad([]byte{}, 32),
									leftPad(alice, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Name:  &name,
					Owner: &aliceAddress,
				},
			},
		},
		{
			name: "old ens registrar controller - NameRenewed",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Address: common.HexToAddress(ens.EthereumOldEnsRegistrarController).Bytes(),
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000060" +
											"0000000000000000000000000000000000000000000000000000000000000000" + // cost
											"0000000000000000000000000000000000000000000000000000000000000000" + // expires
											"0000000000000000000000000000000000000000000000000000000000000007",
									),
									rightPad([]byte(name)),
								}, nil),
								Topics: [][]byte{
									ens.OldRegistrarControllerNameRenewedTopic.Bytes(),
									leftPad([]byte{}, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Name: &name,
				},
			},
		},
		{
			name: "ens public resolver - NameChanged",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000020" +
											"0000000000000000000000000000000000000000000000000000000000000007",
									),
									rightPad([]byte(name)),
								}, nil),
								Topics: [][]byte{
									ens.PublicResolverNameChangedTopic.Bytes(),
									leftPad([]byte{}, 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Name: &name,
				},
			},
		},
		{
			name: "ens public resolver - AddressChanged",
			block: &types.Eth1Block{
				Transactions: []*types.Eth1Transaction{
					{
						Logs: []*types.Eth1Log{
							{
								Data: bytes.Join([][]byte{
									common.Hex2Bytes(
										"0000000000000000000000000000000000000000000000000000000000000040" +
											"0000000000000000000000000000000000000000000000000000000000000000" + // coinType
											"0000000000000000000000000000000000000000000000000000000000000000", // address
									),
								}, nil),
								Topics: [][]byte{
									ens.PublicResolverAddressChangedTopic.Bytes(),
									leftPad([]byte("node"), 32),
								},
							},
						},
					},
				},
			},
			want: []data.ENSLog{
				{
					Node: (*[32]byte)(leftPad([]byte("node"), 32)),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res IndexedBlock
			if err := TransformEnsNameRegistered(chainID.String(), tt.block, &res); err != nil {
				t.Fatal(err)
			}
			if got, want := len(res.ENS), len(tt.want); got != want {
				t.Fatalf("got %v, want %v", got, want)
			}
			for i, indexed := range tt.want {
				if indexed.Node != nil {
					if got, want := *res.ENS[i].Node, *indexed.Node; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
				if indexed.Owner != nil {
					if got, want := *res.ENS[i].Owner, *indexed.Owner; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
				if indexed.Name != nil {
					if got, want := *res.ENS[i].Name, *indexed.Name; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
			}
		})
	}
}

func leftPad(b []byte, length int) []byte {
	for len(b) != length {
		b = append([]byte{0}, b...)
	}
	return b
}

func rightPad(b []byte) []byte {
	for len(b) != 32 {
		b = append(b, 0)
	}
	return b
}

func TestTransformer_FromList(t *testing.T) {
	tests := []struct {
		name    string
		want    TransformFunc
		wantErr bool
	}{
		{
			name: "TransformBlock",
			want: TransformBlock,
		},
		{
			name: "TransformTx",
			want: TransformTx,
		},
		{
			name: "TransformBlobTx",
			want: TransformBlob,
		},
		{
			name: "TransformItx",
			want: TransformITx,
		},
		{
			name: "TransformERC20",
			want: TransformERC20,
		},
		{
			name: "TransformERC721",
			want: TransformERC721,
		},
		{
			name: "TransformERC1155",
			want: TransformERC1155,
		},
		{
			name: "TransformWithdrawals",
			want: TransformWithdrawal,
		},
		{
			name: "TransformUncle",
			want: TransformUncle,
		},
		{
			name: "TransformEnsNameRegistered",
			want: TransformEnsNameRegistered,
		},
		{
			name: "TransformContract",
			want: TransformContract,
		},
		{
			name:    "invalid",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformerFromList([]string{tt.name})
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("got %v, want nil", err)
			}
			if got, want := got[0], tt.want; reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
