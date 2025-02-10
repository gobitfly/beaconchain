package executionlayer

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/erc1155"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/erc721"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var (
	alice    = []byte("alice")
	bob      = []byte("bob")
	john     = []byte("john")
	contract = []byte("contract")
	usdc     = []byte("usdc")
	tokenID  = []byte("tokenID")
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

// TestIsValidERC20Log tests the isValidERC20Log function to verify
// if the log is a valid ERC20 transfer log
func TestIsValidERC20Log(t *testing.T) {
	tests := []struct {
		name     string
		log      *types.Eth1Log
		expected bool
	}{
		{
			name: "valid ERC20 transfer log",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc20.TransferTopic.Bytes(),
					alice,
					bob,
				},
			},
			expected: true,
		},
		{
			name: "invalid ERC20 transfer log with incorrect event topic",
			log: &types.Eth1Log{
				Topics: [][]byte{
					common.HexToHash("0x0123").Bytes(),
					alice,
					bob,
				},
			},
			expected: false,
		},
		{
			name: "invalid ERC20 transfer log with too little topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc20.TransferTopic.Bytes(),
					alice,
				},
			},
			expected: false,
		},
		{
			name: "invalid ERC20 transfer log with too many topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc20.TransferTopic.Bytes(),
					alice,
					bob,
					john,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidERC20Log(tt.log)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidERC721Log tests the isValidERC721Log function to verify
// if the log is a valid ERC721 transfer log
func TestIsValidERC721Log(t *testing.T) {
	tests := []struct {
		name     string
		log      *types.Eth1Log
		expected bool
	}{
		{
			name: "valid ERC721 transfer log",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc721.TransferTopic.Bytes(),
					alice,
					bob,
					tokenID,
				},
			},
			expected: true,
		},
		{
			name: "invalid ERC721 transfer log with incorrect topic",
			log: &types.Eth1Log{
				Topics: [][]byte{
					common.HexToHash("0x1234").Bytes(),
					alice,
					bob,
					tokenID,
				},
			},
			expected: false,
		},
		{
			name: "invalid ERC721 transfer log with too little topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc721.TransferTopic.Bytes(),
					alice,
					bob,
				},
			},
			expected: false,
		},
		{
			name: "invalid ERC721 transfer log with too many topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc721.TransferTopic.Bytes(),
					alice,
					bob,
					tokenID,
					john,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidERC721Log(tt.log)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidERC1155Log tests the isValidERC1155Log function to verify
// if the log is a valid ERC1155 transfer log
func TestIsValidERC1155Log(t *testing.T) {
	tests := []struct {
		name     string
		log      *types.Eth1Log
		expected bool
	}{
		{
			name: "valid TransferSingleTopic log",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferSingleTopic.Bytes(),
					alice,
					bob,
					john,
				},
			},
			expected: true,
		},
		{
			name: "valid TransferBulkTopic log",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferBulkTopic.Bytes(),
					alice,
					bob,
					john,
				},
			},
			expected: true,
		},
		{
			name: "invalid log with invalid topic",
			log: &types.Eth1Log{
				Topics: [][]byte{
					common.HexToHash("0x1234").Bytes(),
					alice,
					bob,
					john,
				},
			},
			expected: false,
		},
		{
			name: "invalid TransferSingleTopic log with too little topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferSingleTopic.Bytes(),
					alice,
					bob,
				},
			},
			expected: false,
		},
		{
			name: "invalid TransferBulkTopic log with too little topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferBulkTopic.Bytes(),
					alice,
					bob,
				},
			},
			expected: false,
		},
		{
			name: "invalid TransferSingleTopic log with too many topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferSingleTopic.Bytes(),
					alice,
					bob,
					john,
					contract,
				},
			},
			expected: false,
		},
		{
			name: "invalid TransferBulkTopic log with too many topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					erc1155.TransferBulkTopic.Bytes(),
					alice,
					bob,
					john,
					contract,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidERC1155Log(tt.log)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidItx tests the isValidItx function to verify
// if the internal transactions are valid and can be indexed
func TestIsValidItx(t *testing.T) {
	tests := []struct {
		name     string
		itx      *types.Eth1InternalTransaction
		expected bool
	}{
		{
			name: "valid internal transaction",
			itx: &types.Eth1InternalTransaction{
				Path:  "0,1",
				Value: big.NewInt(100).Bytes(),
			},
			expected: true,
		},
		{
			name: "invalid internal transaction with path '0'",
			itx: &types.Eth1InternalTransaction{
				Path:  "0",
				Value: big.NewInt(100).Bytes(),
			},
			expected: false,
		},
		{
			name: "invalid internal transaction with path is '[]'",
			itx: &types.Eth1InternalTransaction{
				Path:  "[]",
				Value: big.NewInt(100).Bytes(),
			},
			expected: false,
		},
		{
			name: "invalid internal transaction with value 0",
			itx: &types.Eth1InternalTransaction{
				Path:  "0,1",
				Value: []byte{0x0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidItx(tt.itx)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetLogTopics tests the getLogTopics function to verify
// if the function correctly extracts the topics from the log
func TestGetLogTopics(t *testing.T) {
	topic1 := "0x1234"
	topic2 := "0x2345"
	topic3 := "0x3456"
	tests := []struct {
		name     string
		log      *types.Eth1Log
		expected []common.Hash
	}{
		{
			name: "log with topics",
			log: &types.Eth1Log{
				Topics: [][]byte{
					common.HexToHash(topic1).Bytes(),
					common.HexToHash(topic2).Bytes(),
					common.HexToHash(topic3).Bytes(),
				},
			},
			expected: []common.Hash{
				common.HexToHash(topic1),
				common.HexToHash(topic2),
				common.HexToHash(topic3),
			},
		},
		{
			name: "log with no topics",
			log: &types.Eth1Log{
				Topics: [][]byte{},
			},
			expected: []common.Hash{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLogTopics(tt.log)
			if len(result) != len(tt.expected) {
				t.Errorf("got %v topics, want %v topics", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("got topic %v, want %v", result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestGetTxRecipient tests the getTxRecipient function to verify
// if the function correctly extracts the recipients from the transaction
func TestGetTxRecipient(t *testing.T) {
	tests := []struct {
		name               string
		tx                 *types.Eth1Transaction
		expectedTo         []byte
		expectedIsContract bool
	}{
		{
			name: "normal transaction with recipient",
			tx: &types.Eth1Transaction{
				To:              alice,
				ContractAddress: common.Address{}.Bytes(),
			},
			expectedTo:         alice,
			expectedIsContract: false,
		},
		{
			name: "contract creation transaction",
			tx: &types.Eth1Transaction{
				ContractAddress: contract,
			},
			expectedTo:         contract,
			expectedIsContract: true,
		},
		{
			name: "transaction with no recipient",
			tx: &types.Eth1Transaction{
				To:              common.Address{}.Bytes(),
				ContractAddress: common.Address{}.Bytes(),
			},
			expectedTo:         common.Address{}.Bytes(),
			expectedIsContract: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to, isContract := getTxRecipient(tt.tx)
			if !bytes.Equal(to, tt.expectedTo) {
				t.Errorf("got %v, want %v", to, tt.expectedTo)
			}
			if isContract != tt.expectedIsContract {
				t.Errorf("got %v, want %v", isContract, tt.expectedIsContract)
			}
		})
	}
}

// TestGetTokenID tests the getTokenID function to verify
// if the function correctly extracts the token ID from the transfer
func TestGetTokenID(t *testing.T) {
	tests := []struct {
		name     string
		transfer *contracts.ERC721Transfer
		expected *big.Int
	}{
		{
			name: "transfer with token ID",
			transfer: &contracts.ERC721Transfer{
				TokenId: big.NewInt(123),
			},
			expected: big.NewInt(123),
		},
		{
			name: "transfer with empty token ID",
			transfer: &contracts.ERC721Transfer{
				TokenId: nil,
			},
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTokenID(tt.transfer)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetMethodsSignature tests the getMethodSignature function to verify
// if the function correctly extracts the method signature from the transaction
func TestGetMethodSignature(t *testing.T) {
	tests := []struct {
		name     string
		tx       *types.Eth1Transaction
		expected []byte
	}{
		{
			name: "transaction with data longer than 4 bytes",
			tx: &types.Eth1Transaction{
				Data: []byte("123456789"),
			},
			expected: []byte("1234"),
		},
		{
			name: "transaction with data exactly 4 bytes",
			tx: &types.Eth1Transaction{
				Data: []byte("1234"),
			},
			expected: []byte("1234"),
		},
		{
			name: "transaction with data shorter than 4 bytes",
			tx: &types.Eth1Transaction{
				Data: []byte("12"),
			},
			expected: []byte{},
		},
		{
			name: "transaction with no data",
			tx: &types.Eth1Transaction{
				Data: []byte{},
			},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := getMethodSignature(tt.tx)
			if !bytes.Equal(method, tt.expected) {
				t.Errorf("got %v, want %v", method, tt.expected)
			}
		})
	}
}

// TestGetContractAddress tests the getContractAddress function to verify
// if the function correctly extracts the contract address from the internal transaction
func TestGetContractAddress(t *testing.T) {
	tests := []struct {
		name     string
		itx      *types.Eth1InternalTransaction
		expected []byte
	}{
		{
			name: "create type transaction",
			itx: &types.Eth1InternalTransaction{
				Type: "create",
				From: alice,
				To:   contract,
			},
			expected: contract,
		},
		{
			name: "suicide type transaction",
			itx: &types.Eth1InternalTransaction{
				Type: "suicide",
				From: contract,
				To:   alice,
			},
			expected: contract,
		},
		{
			name: "invalid type transaction",
			itx: &types.Eth1InternalTransaction{
				Type: "invalid",
				From: alice,
				To:   contract,
			},
			expected: contract,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContractAddress(tt.itx)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetERC20TransferValue tests getERC20TransferValue function to verify
// if the function correctly extracts the ERC20 transfer value from the ERC20 transfer contract
func TestGetERC20TransferValue(t *testing.T) {
	tests := []struct {
		name     string
		transfer *contracts.ERC20Transfer
		expected []byte
	}{
		{
			name: "transfer with value",
			transfer: &contracts.ERC20Transfer{
				Value: big.NewInt(1),
			},
			expected: big.NewInt(1).Bytes(),
		},
		{
			name: "transfer with empty value",
			transfer: &contracts.ERC20Transfer{
				Value: nil,
			},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getERC20TransferValue(tt.transfer)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetERC1155TransferIDs tests getERC1155TransferIDs function to verify
// if the function correctly extracts the ERC1155 transfer IDs from the ID list
func TestGetERC1155TransferIDs(t *testing.T) {
	tests := []struct {
		name     string
		idList   []*big.Int
		expected [][]byte
	}{
		{
			name: "single ID",
			idList: []*big.Int{
				big.NewInt(1),
			},
			expected: [][]byte{
				big.NewInt(1).Bytes(),
			},
		},
		{
			name: "multiple IDs",
			idList: []*big.Int{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
			expected: [][]byte{
				big.NewInt(1).Bytes(),
				big.NewInt(2).Bytes(),
				big.NewInt(3).Bytes(),
			},
		},
		{
			name:     "empty ID list",
			idList:   []*big.Int{},
			expected: [][]byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getERC1155TransferIDs(tt.idList)
			if len(result) != len(tt.expected) {
				t.Errorf("got %v IDs, want %v IDs", len(result), len(tt.expected))
			}
			for i := range result {
				if !bytes.Equal(result[i], tt.expected[i]) {
					t.Errorf("got ID %v, want %v", result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestGetERC1155TransferValues tests getERC1155TransferValues function to verify
// if the function correctly extracts the ERC1155 transfer values from the value list
func TestGetERC1155TransferValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []*big.Int
		expected [][]byte
	}{
		{
			name: "single value",
			values: []*big.Int{
				big.NewInt(1),
			},
			expected: [][]byte{
				big.NewInt(1).Bytes(),
			},
		},
		{
			name: "multiple values",
			values: []*big.Int{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
			expected: [][]byte{
				big.NewInt(1).Bytes(),
				big.NewInt(2).Bytes(),
				big.NewInt(3).Bytes(),
			},
		},
		{
			name:     "empty values list",
			values:   []*big.Int{},
			expected: [][]byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getERC1155TransferValues(tt.values)
			if len(result) != len(tt.expected) {
				t.Errorf("got %v values, want %v values", len(result), len(tt.expected))
			}
			for i := range result {
				if !bytes.Equal(result[i], tt.expected[i]) {
					t.Errorf("got value %v, want %v", result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestUpdateITxStatus tests the updateITxStatus function to verify
// if the function correctly updates the status of the internal transaction
func TestUpdateITxStatus(t *testing.T) {
	tests := []struct {
		name             string
		internalTx       []*types.Eth1InternalTransaction
		indexedTx        *types.Eth1TransactionIndexed
		expectedStatus   types.StatusType
		expectedErrorMsg string
	}{
		{
			name:       "no internal transactions",
			internalTx: []*types.Eth1InternalTransaction{},
			indexedTx: &types.Eth1TransactionIndexed{
				Status: types.StatusType_SUCCESS,
			},
			expectedStatus:   types.StatusType_SUCCESS,
			expectedErrorMsg: "",
		},
		{
			name: "internal transaction with error",
			internalTx: []*types.Eth1InternalTransaction{
				{
					ErrorMsg: "fail",
				},
			},
			indexedTx: &types.Eth1TransactionIndexed{
				Status: types.StatusType_SUCCESS,
			},
			expectedStatus:   types.StatusType_PARTIAL,
			expectedErrorMsg: "fail",
		},
		{
			name: "internal transaction with error and status failed",
			internalTx: []*types.Eth1InternalTransaction{
				{
					ErrorMsg: "fail",
				},
			},
			indexedTx: &types.Eth1TransactionIndexed{
				Status: types.StatusType_FAILED,
			},
			expectedStatus:   types.StatusType_FAILED,
			expectedErrorMsg: "fail",
		},
		{
			name: "multiple internal transactions, one with error",
			internalTx: []*types.Eth1InternalTransaction{
				{
					ErrorMsg: "",
				},
				{
					ErrorMsg: "fail",
				},
			},
			indexedTx: &types.Eth1TransactionIndexed{
				Status: types.StatusType_SUCCESS,
			},
			expectedStatus:   types.StatusType_PARTIAL,
			expectedErrorMsg: "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateITxStatus(tt.indexedTx, tt.internalTx)
			if tt.indexedTx.Status != tt.expectedStatus {
				t.Errorf("got status %v, want %v", tt.indexedTx.Status, tt.expectedStatus)
			}
			if tt.indexedTx.ErrorMsg != tt.expectedErrorMsg {
				t.Errorf("got error message %v, want %v", tt.indexedTx.ErrorMsg, tt.expectedErrorMsg)
			}
		})
	}
}

// TestCalculateTxFee tests the calculateTxFee function to verify
// if the function correctly calculates the transaction fee
func TestCalculateTxFee(t *testing.T) {
	tests := []struct {
		name     string
		tx       *types.Eth1Transaction
		baseFee  []byte
		expected *big.Int
	}{
		{
			name: "transaction with no base fee",
			tx: &types.Eth1Transaction{
				GasPrice: big.NewInt(10).Bytes(),
				GasUsed:  1000,
			},
			baseFee:  []byte{},
			expected: big.NewInt(10 * 1000),
		},
		{
			name: "transaction with priority and base fee",
			tx: &types.Eth1Transaction{
				GasPrice:             big.NewInt(100).Bytes(),
				MaxPriorityFeePerGas: big.NewInt(10).Bytes(),
				MaxFeePerGas:         big.NewInt(200).Bytes(),
				GasUsed:              1000,
			},
			baseFee:  big.NewInt(50).Bytes(),
			expected: big.NewInt(10 * 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTxFee(tt.tx, tt.baseFee)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateMevFromBlock tests the calculateMevFromBlock function to verify
// if the function correctly calculates the MEV from the block
func TestCalculateMevFromBlock(t *testing.T) {
	tests := []struct {
		name     string
		block    *types.Eth1Block
		expected *big.Int
	}{
		{
			name: "no MEV",
			block: &types.Eth1Block{
				Coinbase: []byte("coinbase"),
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								From:  alice,
								To:    common.Address{}.Bytes(),
								Value: big.NewInt(100).Bytes(),
							},
						},
					},
				},
			},
			expected: big.NewInt(0),
		},
		{
			name: "MEV from one transaction",
			block: &types.Eth1Block{
				Coinbase: []byte("coinbase"),
				Transactions: []*types.Eth1Transaction{
					{
						Itx: []*types.Eth1InternalTransaction{
							{
								From:  alice,
								To:    []byte("coinbase"),
								Value: big.NewInt(100).Bytes(),
							},
						},
					},
				},
			},
			expected: big.NewInt(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMevFromBlock(tt.block)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateBlockUncleReward tests the calculateBlockUncleReward function to verify
// if the function correctly calculates the uncle reward from the block
func TestCalculateBlockUncleReward(t *testing.T) {
	tests := []struct {
		name     string
		block    *types.Eth1Block
		chainID  string
		expected *big.Int
	}{
		{
			name: "no uncles",
			block: &types.Eth1Block{
				Uncles: []*types.Eth1Block{},
			},
			chainID:  "1",
			expected: big.NewInt(0),
		},
		{
			name: "one uncle",
			block: &types.Eth1Block{
				Number:     10,
				Difficulty: big.NewInt(100).Bytes(),
				Uncles: []*types.Eth1Block{
					{
						Number: 1,
					},
				},
			},
			chainID:  "1",
			expected: new(big.Int).Div(eth1BlockReward("1", 10, big.NewInt(100).Bytes()), big.NewInt(32)),
		},
		{
			name: "two uncles",
			block: &types.Eth1Block{
				Number:     10,
				Difficulty: big.NewInt(100).Bytes(),
				Uncles: []*types.Eth1Block{
					{
						Number: 1,
					},
					{
						Number: 2,
					},
				},
			},
			chainID:  "1",
			expected: new(big.Int).Mul(big.NewInt(2), new(big.Int).Div(eth1BlockReward("1", 10, big.NewInt(100).Bytes()), big.NewInt(32))),
		},
		{
			name: "no uncle rewards",
			block: &types.Eth1Block{
				Number:     10,
				Difficulty: []byte{},
				Uncles: []*types.Eth1Block{
					{
						Number: 1,
					},
				},
			},
			chainID:  "1",
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBlockUncleReward(tt.block, tt.chainID)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateUncleReward tests the calculateUncleReward function to verify
// if the function correctly calculates the single uncle reward
func TestCalculateUncleReward(t *testing.T) {
	tests := []struct {
		name     string
		block    *types.Eth1Block
		uncle    *types.Eth1Block
		chainID  string
		expected *big.Int
	}{
		{
			name: "no uncles",
			block: &types.Eth1Block{
				Uncles: []*types.Eth1Block{},
			},
			chainID:  "1",
			expected: big.NewInt(0),
		},
		{
			name: "one uncle",
			block: &types.Eth1Block{
				Number:     10,
				Difficulty: big.NewInt(100).Bytes(),
			},
			uncle: &types.Eth1Block{
				Number: 1,
			},
			chainID:  "1",
			expected: new(big.Int).Div(eth1BlockReward("1", 10, big.NewInt(100).Bytes()), big.NewInt(32)),
		},
		{
			name: "no uncle rewards",
			block: &types.Eth1Block{
				Number:     10,
				Difficulty: []byte{},
			},
			uncle: &types.Eth1Block{
				Number: 1,
			},
			chainID:  "1",
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUncleReward(tt.block, tt.uncle, tt.chainID)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestVerifyName tests the verifyName function to verify
// if the function correctly validates the name based on the length
func TestVerifyName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid name",
			input:    "test",
			expected: nil,
		},
		{
			name:     "empty name",
			input:    "",
			expected: nil,
		},
		{
			name:     "maximum length name",
			input:    string(make([]byte, 2048)),
			expected: nil,
		},
		{
			name:     "name too long",
			input:    string(make([]byte, 2049)),
			expected: fmt.Errorf("name too long: %v", string(make([]byte, 2049))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyName(tt.input)

			if result != nil {
				if tt.expected == nil {
					t.Errorf("got %v, want %v", result, tt.expected)

				}
				if tt.expected != nil {
					if result.Error() != tt.expected.Error() {
						t.Errorf("got %v, want %v", result, tt.expected)
					}
				}
			}
			if result == nil {
				if tt.expected != nil {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}
