package th

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/Tangui-Bitfly/ethsimtracer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/internal/contracts"
)

type EOA struct {
	*bind.TransactOpts
	PrivateKey *ecdsa.PrivateKey
}

const simulatedChainID = 1337

var (
	OneEther            = big.NewInt(params.Ether)
	DefaultTokenBalance = big.NewInt(100)
)

func CreateEOA(t *testing.T) *EOA {
	t.Helper()

	priv, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(priv, big.NewInt(simulatedChainID))
	if err != nil {
		t.Fatal(err)
	}
	return &EOA{
		PrivateKey:   priv,
		TransactOpts: opts,
	}
}

type BlockchainBackend struct {
	*ethsimtracer.Backend
	BankAccount *EOA
	ChainID     int
	Endpoint    string

	rpc *rpc.Client
}

func NewBackend(t *testing.T, accounts ...common.Address) *BlockchainBackend {
	t.Helper()

	genesis := make(map[common.Address]types.Account)
	for _, account := range accounts {
		genesis[account] = types.Account{Balance: OneEther}
	}
	bankAccount := CreateEOA(t)
	genesis[bankAccount.From] = types.Account{Balance: new(big.Int).Mul(big.NewInt(1000), OneEther)}

	log.SetDefault(log.NewLogger(log.DiscardHandler()))
	// TODO use random port
	bk := ethsimtracer.NewBackend(genesis, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
		nodeConf.HTTPModules = append(nodeConf.HTTPModules, "debug", "eth")
		nodeConf.HTTPHost = "127.0.0.1"
		nodeConf.HTTPPort = 4242
	})
	endpoint := "http://127.0.0.1:4242"
	client, err := rpc.Dial(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	return &BlockchainBackend{
		Backend:     bk,
		BankAccount: bankAccount,
		ChainID:     simulatedChainID,
		Endpoint:    endpoint,
		rpc:         client,
	}
}

func (b *BlockchainBackend) Client() *ethclient.Client {
	return ethclient.NewClient(b.rpc)
}

func (b *BlockchainBackend) FundOneEther(t *testing.T, to common.Address) string {
	t.Helper()

	signedTx := b.MakeTx(t, b.BankAccount, &to, OneEther, nil)
	if err := b.Client().SendTransaction(context.Background(), signedTx); err != nil {
		t.Fatal(err)
	}

	b.Commit()
	return signedTx.Hash().Hex()
}

func (b *BlockchainBackend) Fund(t *testing.T, to common.Address, amount *big.Int) string {
	t.Helper()

	signedTx := b.MakeTx(t, b.BankAccount, &to, amount, nil)
	if err := b.Client().SendTransaction(context.Background(), signedTx); err != nil {
		t.Fatal(err)
	}
	b.Commit()
	return signedTx.Hash().Hex()
}

func (b *BlockchainBackend) MakeTx(t *testing.T, sender *EOA, to *common.Address, value *big.Int, data []byte) *types.Transaction {
	t.Helper()

	signedTx, err := types.SignTx(b.MakeUnsignedTx(t, sender.From, to, value, data), types.LatestSignerForChainID(big.NewInt(int64(b.ChainID))), sender.PrivateKey)
	if err != nil {
		t.Errorf("could not sign tx: %v", err)
	}
	return signedTx
}

func (b *BlockchainBackend) MakeUnsignedTx(t *testing.T, from common.Address, to *common.Address, value *big.Int, data []byte) *types.Transaction {
	t.Helper()

	head, _ := b.Backend.Client().HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := b.Backend.Client().ChainID(context.Background())
	nonce, err := b.Backend.Client().PendingNonceAt(context.Background(), from)
	if err != nil {
		t.Fatal(err)
	}
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainid,
		Nonce:     nonce,
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		Gas:       100000,
		To:        to,
		Value:     value,
		Data:      data,
	})
}

func (b *BlockchainBackend) DeployContract(t *testing.T, contractData []byte) common.Address {
	t.Helper()

	head, _ := b.Backend.Client().HeaderByNumber(context.Background(), nil)
	gasPrice := new(big.Int).Add(head.BaseFee, big.NewInt(params.GWei))
	chainid, _ := b.Backend.Client().ChainID(context.Background())
	nonce, err := b.Backend.Client().PendingNonceAt(context.Background(), b.BankAccount.From)
	if err != nil {
		t.Fatal(err)
	}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainid,
		Nonce:     nonce,
		GasTipCap: big.NewInt(params.GWei),
		GasFeeCap: gasPrice,
		Gas:       2000000,
		Data:      contractData,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(b.ChainID))), b.BankAccount.PrivateKey)
	if err != nil {
		t.Errorf("could not sign tx: %v", err)
	}

	if err := b.Client().SendTransaction(context.Background(), signedTx); err != nil {
		t.Fatal(err)
	}
	b.Commit()

	receipt, err := b.Client().TransactionReceipt(context.Background(), signedTx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if receipt == nil {
		t.Fatal("empty receipt")
	}
	if (receipt.ContractAddress == common.Address{}) {
		t.Fatal("expected actual deployed contracts address")
	}

	return receipt.ContractAddress
}

func (b *BlockchainBackend) DeployToken(t *testing.T, name string, symbol string, accounts ...common.Address) (common.Address, *contracts.Token) {
	t.Helper()
	address, _, token, err := contracts.DeployToken(b.BankAccount.TransactOpts, b.Client(), name, symbol)
	if err != nil {
		t.Fatal(err)
	}
	b.Commit()

	for _, account := range accounts {
		_, err := token.Mint(b.BankAccount.TransactOpts, account, DefaultTokenBalance)
		if err != nil {
			t.Fatal(err)
		}
		b.Commit()
	}
	return address, token
}

func (b *BlockchainBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return b.Client().CallContract(ctx, call, blockNumber)
}

func (b *BlockchainBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return b.Client().SendTransaction(ctx, tx)
}

func (b *BlockchainBackend) Commit() {
	b.Backend.Commit()
}
