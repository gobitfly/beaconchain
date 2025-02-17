package executionlayer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

var testAddress = common.HexToAddress("0x000000000000000000000000000000000000beef")

func TestLlama_GetPrices(t *testing.T) {
	tests := []struct {
		name  string
		coins []common.Address
		resp  defiLlamaResponse
		want  []*types.ERC20TokenPrice
	}{
		{
			name:  "happy case",
			coins: []common.Address{testAddress},
			resp: defiLlamaResponse{
				Coins: map[string]defiLlamaCoin{
					"ethereum:" + testAddress.String(): {
						Price: decimal.NewFromFloat(0.1),
					},
				},
			},
			want: []*types.ERC20TokenPrice{
				{
					Token: testAddress.Bytes(),
					Price: []byte("0.1"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := json.Marshal(tt.resp)
				_, _ = w.Write(b)
			}))
			defer server.Close()
			l := DefiLlama{
				url: server.URL,
			}
			got, err := l.GetPrices(tt.coins)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type stubExternalPricer struct {
	prices map[common.Address]*types.ERC20TokenPrice
	err    error
}

func (s stubExternalPricer) GetPrices(tokens []common.Address) ([]*types.ERC20TokenPrice, error) {
	var prices []*types.ERC20TokenPrice
	for _, address := range tokens {
		prices = append(prices, s.prices[address])
	}
	return prices, s.err
}

func TestTokenPricer(t *testing.T) {
	btClient, btAdmin := databasetest.NewBigTable(t)
	metadataBigtable, err := database.NewBigTableWithClient(context.Background(), btClient, btAdmin, db2.Schema)
	store := db2.NewStoreV1(nil, database.Wrap(metadataBigtable, db2.MetadataTable), nil, db2.NoopCache{})
	if err != nil {
		t.Fatal(err)
	}
	backend := th.NewBackend(t)

	tokenAddress, token := backend.DeployERC20(t, "usdt", "usdt", backend.BankAccount.From)
	multicall := backend.DeployContract(t, common.FromHex(contracts.MulticallMetaData.Bin))
	supply, _ := token.TotalSupply(nil)

	tests := []struct {
		name   string
		pricer stubExternalPricer
		want   *types.ERC20TokenPrice
	}{
		{
			name: "full flow",
			pricer: stubExternalPricer{
				prices: map[common.Address]*types.ERC20TokenPrice{
					tokenAddress: {
						Token: tokenAddress.Bytes(),
						Price: []byte("1"),
					},
				},
			},
			want: &types.ERC20TokenPrice{
				Token:       tokenAddress.Bytes(),
				Price:       []byte("1"),
				TotalSupply: supply.Bytes(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() { _ = metadataBigtable.Clear() }()

			pricer := NewTokenPricer(
				store,
				fmt.Sprintf("%d", backend.ChainID),
				tt.pricer,
				erc20.ERC20TokenList{
					Tokens: []*erc20.ERC20TokenDetail{
						{Address: tokenAddress.String()},
					},
				},
				evm.NewMulticallBatcher(backend.Client(), multicall, 0),
			)
			if err := pricer.UpdateTokens(); err != nil {
				t.Fatal(err)
			}
			price, err := store.TokenPrice(fmt.Sprintf("%d", backend.ChainID), tokenAddress)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := price.Token, tt.want.Token; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := price.TotalSupply, tt.want.TotalSupply; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := price.Price, tt.want.Price; !bytes.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
