package executionlayer

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
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
			updateITxStatus(tt.internalTx, tt.indexedTx)
			if tt.indexedTx.Status != tt.expectedStatus {
				t.Errorf("got status %v, want %v", tt.indexedTx.Status, tt.expectedStatus)
			}
			if tt.indexedTx.ErrorMsg != tt.expectedErrorMsg {
				t.Errorf("got error message %v, want %v", tt.indexedTx.ErrorMsg, tt.expectedErrorMsg)
			}
		})
	}
}

func addPrefix(b []byte, length int) []byte {
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
