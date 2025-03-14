package executionlayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

type ExternalPricer interface {
	GetPrices(tokens []common.Address) ([]*types.ERC20TokenPrice, error)
}

type TokenStore interface {
	UpdateToken(chainID string, tokens []*types.ERC20TokenPrice) error
}

type TokenPricer struct {
	store    TokenStore
	external ExternalPricer
	batcher  evm.Batcher
	chainID  string
}

func NewTokenPricer(store TokenStore, chainID string, external ExternalPricer, batcher evm.Batcher) *TokenPricer {
	return &TokenPricer{
		store:    store,
		external: external,
		batcher:  batcher,
		chainID:  chainID,
	}
}

// UpdateTokens retrieve the prices from an external source and the total supply from the chain
func (t *TokenPricer) UpdateTokens(list erc20.ERC20TokenList) error {
	var tokens []common.Address
	for _, token := range list.Tokens {
		tokens = append(tokens, common.HexToAddress(token.Address))
	}
	prices, err := t.external.GetPrices(tokens)
	if err != nil {
		return fmt.Errorf("cannot get token prices: %w", err)
	}

	supplies, err := evm.ERC20Supply(t.batcher, tokens)
	if err != nil {
		return fmt.Errorf("cannot get token supplies: %w", err)
	}

	// for each token set the supply to the onchain response
	for i := 0; i < len(prices); i++ {
		prices[i].TotalSupply = supplies[i].Bytes()
	}
	if err := t.store.UpdateToken(t.chainID, prices); err != nil {
		return fmt.Errorf("cannot update token: %w", err)
	}
	return nil
}

type DefiLlama struct {
	url string
}

func NewLlamaClient() DefiLlama {
	return DefiLlama{
		url: "https://coins.llama.fi",
	}
}

type defiLlamaResponse struct {
	Coins map[string]defiLlamaCoin `json:"coins"`
}

type defiLlamaCoin struct {
	Decimals  int64           `json:"decimals"`
	Price     decimal.Decimal `json:"price"`
	Symbol    string          `json:"symbol"`
	Timestamp int64           `json:"timestamp"`
}

type defiLlamaPriceRequest struct {
	Coins []string `json:"coins"`
}

func (l DefiLlama) GetPrices(tokens []common.Address) ([]*types.ERC20TokenPrice, error) {
	var coins []string
	for _, token := range tokens {
		coins = append(coins, "ethereum:"+token.String())
	}
	reqEncoded, err := json.Marshal(&defiLlamaPriceRequest{
		Coins: coins,
	})
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	resp, err := httpClient.Post(fmt.Sprintf("%s/prices", l.url), "application/json", bytes.NewReader(reqEncoded))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error querying defillama api: %v", resp.Status)
	}

	var llamaResp defiLlamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&llamaResp); err != nil {
		return nil, err
	}

	var tokenPrices []*types.ERC20TokenPrice
	for address, data := range llamaResp.Coins {
		tokenPrices = append(tokenPrices, &types.ERC20TokenPrice{
			Token: common.FromHex(strings.TrimPrefix(address, "ethereum:0x")),
			Price: []byte(data.Price.String()),
		})
	}
	return tokenPrices, nil
}
