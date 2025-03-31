package price

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/contracts/chainlink_feed"
	"github.com/gobitfly/beaconchain/pkg/commons/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

var availableCurrencies = []string{}

var runOnce sync.Once
var runOnceWg sync.WaitGroup
var prices = map[string]float64{}
var pricesMu = &sync.Mutex{}
var didInit = uint64(0)
var feeds = map[string]*chainlink_feed.Feed{}
var mainCurrency = "ETH" // currency in which all exchange rates are denominated, e.g. ETH or GNO
var clCurrency = "ETH"   // currency in which all CL values are denominated, e.g. ETH or mGNO
var elCurrency = "ETH"   // currency in which all EL values are denominated, e.g. ETH or xDai

var currencies = map[string]struct {
	Symbol string
	Label  string
}{
	"AUD":  {"A$", "Australian Dollar"},
	"CAD":  {"C$", "Canadian Dollar"},
	"CNY":  {"¥", "Chinese Yuan"},
	"xDAI": {"DAI", "xDAI stablecoin"},
	"ETH":  {"ETH", "Ether"},
	"EUR":  {"€", "Euro"},
	"GBP":  {"£", "Pound Sterling"},
	"GNO":  {"GNO", "Gnosis"},
	"mGNO": {"mGNO", "mGnosis"},
	"JPY":  {"¥", "Japanese Yen"},
	"USD":  {"$", "United States Dollar"},
}

func init() {
	runOnceWg.Add(1)
}

func Init(chainId uint64, eth1Endpoint, mainCurrencyParam, clCurrencyParam, elCurrencyParam string) {
	if atomic.AddUint64(&didInit, 1) > 1 {
		log.Warnf("price.Init called multiple times")
		return
	}

	switch chainId {
	case 1, 100:
	default:
		setPrice(mainCurrency, elCurrency, 1)
		setPrice(mainCurrency, clCurrency, 1)
		availableCurrencies = []string{clCurrency, elCurrency}
		log.Warnf("chainId not supported for fetching prices: %v", chainId)
		runOnce.Do(func() { runOnceWg.Done() })
		return
	}

	mainCurrency = mainCurrencyParam
	clCurrency = clCurrencyParam
	elCurrency = elCurrencyParam

	eClient, err := ethclient.Dial(eth1Endpoint)
	if err != nil {
		log.Error(err, "error dialing pricing eth1 endpoint", 0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	clientChainId, err := eClient.ChainID(ctx)
	if err != nil {
		log.Fatal(err, "failed getting chainID", 0)
	}
	if chainId != clientChainId.Uint64() {
		log.Fatal(err, "chainId does not match chainId from client", 0, map[string]interface{}{"chainId": chainId, "clientChainId": clientChainId})
	}

	feedAddrs := map[string]string{}
	switch chainId {
	case 1:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/
		feedAddrs["ETH/USD"] = "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"
		feedAddrs["EUR/USD"] = "0xb49f677943bc038e9857d61e7d053caa2c1734c1"
		feedAddrs["CAD/USD"] = "0xa34317db73e77d453b1b8d04550c44d10e981c8e"
		feedAddrs["CNY/USD"] = "0xef8a4af35cd47424672e3c590abd37fbb7a7759a"
		feedAddrs["JPY/USD"] = "0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"
		feedAddrs["GBP/USD"] = "0x5c0ab2d9b5a7ed9f470386e82bb36a3613cdd4b5"
		feedAddrs["AUD/USD"] = "0x77f9710e7d0a19669a13c055f62cd80d313df022"
		availableCurrencies = []string{"ETH", "USD", "EUR", "GBP", "CNY", "CAD", "AUD", "JPY"}
	case 100:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/?network=gnosis-chain
		feedAddrs["GNO/USD"] = "0x22441d81416430A54336aB28765abd31a792Ad37"
		feedAddrs["xDAI/USD"] = "0x678df3415fc31947dA4324eC63212874be5a82f8"
		feedAddrs["EUR/USD"] = "0xab70BCB260073d036d1660201e9d5405F5829b7a"
		feedAddrs["JPY/USD"] = "0x2AfB993C670C01e9dA1550c58e8039C1D8b8A317"
		feedAddrs["ETH/USD"] = "0xa767f745331D267c7751297D982b050c93985627"
		setPrice("GNO", "mGNO", 32)
		availableCurrencies = []string{"GNO", "mGNO", "xDAI", "ETH", "USD", "EUR", "JPY"}
	default:
		log.Fatal(fmt.Errorf("unsupported chainId %v", chainId), "", 0)
	}

	for pair, addrHex := range feedAddrs {
		feed, err := chainlink_feed.NewFeed(common.HexToAddress(addrHex), eClient)
		if err != nil {
			log.Error(err, "failed to initialized chainlink feed", 0, map[string]interface{}{"pair": pair, "addrHex": addrHex})
			return
		}
		feeds[pair] = feed
	}

	go func() {
		for {
			updatePrices()
			time.Sleep(time.Minute)
		}
	}()
}

func updatePrices() {
	g := &errgroup.Group{}
	for pair, feed := range feeds {
		pair := pair
		feed := feed
		g.Go(func() error {
			price, err := getPriceFromFeed(feed)
			if err != nil {
				return fmt.Errorf("error getting price from feed for %v: %w", pair, err)
			}
			pricesMu.Lock()
			defer pricesMu.Unlock()
			prices[pair] = price
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		log.Error(err, "error upating prices", 0)
		return
	}

	// add prices of main currency to all other currencies
	pricesMu.Lock()
	defer pricesMu.Unlock()
	currencyUsdPrice, exists := prices[mainCurrency+"/USD"]
	if !exists {
		log.Error(fmt.Errorf("failed updating prices: cant find %v pair %+v", mainCurrency+"/USD", prices), "", 0)
		return
	}
	for pair, price := range prices {
		s := strings.Split(pair, "/")
		if len(s) < 2 || s[1] != "USD" {
			continue
		}
		prices[mainCurrency+"/"+s[0]] = currencyUsdPrice / price
	}

	runOnce.Do(func() { runOnceWg.Done() })
}

func setPrice(a, b string, v float64) {
	pricesMu.Lock()
	defer pricesMu.Unlock()
	prices[a+"/"+b] = v
}

func GetPrice(a, b string) float64 {
	if didInit < 1 {
		log.Fatal(fmt.Errorf("using GetPrice without calling price.Init once"), "", 0)
	}
	runOnceWg.Wait()
	pricesMu.Lock()
	defer pricesMu.Unlock()
	price, exists := prices[a+"/"+b]
	if !exists {
		log.WarnWithFields(log.Fields{"pair": a + "/" + b}, "price pair not found")
		return 1
	}
	return price
}

func getPriceFromFeed(feed *chainlink_feed.Feed) (float64, error) {
	decimals := decimal.NewFromInt(1e8) // 8 decimal places for the Chainlink feeds
	res, err := feed.LatestRoundData(&bind.CallOpts{})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest chainlink eth/usd price feed data: %w", err)
	}
	return decimal.NewFromBigInt(res.Answer, 0).Div(decimals).InexactFloat64(), nil
}

func GetAvailableCurrencies() []string {
	return availableCurrencies
}

func IsAvailableCurrency(currency string) bool {
	for _, c := range availableCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

func GetCurrencyLabel(currency string) string {
	x, exists := currencies[currency]
	if !exists {
		return ""
	}
	return x.Label
}

func GetCurrencySymbol(currency string) string {
	x, exists := currencies[currency]
	if !exists {
		return ""
	}
	return x.Symbol
}
