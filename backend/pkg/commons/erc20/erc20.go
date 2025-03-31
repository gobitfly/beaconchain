package erc20

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

var ABI, _ = contracts.ERC20MetaData.GetAbi()

var TransferTopic = ABI.Events["Transfer"].ID

var tokenMap = make(map[string]*ERC20TokenDetail)

func InitTokenList(path string) {
	body, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err, "unable to retrieve erc20 token list", 0)
	}
	TokenList := &ERC20TokenList{}

	err = json.Unmarshal(body, TokenList)
	if err != nil {
		log.Fatal(err, "unable to parse erc20 token list", 0)
	}

	for _, token := range TokenList.Tokens {
		address := strings.Replace(token.Address, "0x", "", -1)
		address = strings.ToLower(address)
		tokenMap[address] = token
		// logger.Info(address)
	}
}

func GetTokenDetail(address string) *ERC20TokenDetail {
	return tokenMap[address]
}

type ERC20TokenList struct {
	Keywords  []string            `json:"keywords"`
	LogoURI   string              `json:"logoURI"`
	Name      string              `json:"name"`
	Timestamp string              `json:"timestamp"`
	Tokens    []*ERC20TokenDetail `json:"tokens"`
	Version   struct {
		Major int64 `json:"major"`
		Minor int64 `json:"minor"`
		Patch int64 `json:"patch"`
	} `json:"version"`
}

type ERC20TokenDetail struct {
	Address  string `json:"address"`
	Owner    string `json:"-"`
	Decimals int64  `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Divider  *big.Int
	Contract *contracts.IERC20
}

func (td *ERC20TokenDetail) FormatAmount(in *big.Int) string {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(td.Decimals)))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	return fmt.Sprintf("%v", result)
}

func (td *ERC20TokenDetail) FormatAmountFloat(in *big.Int) float64 {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	f, _ := result.Float64()
	return f
}

func (td *ERC20TokenDetail) ToScaled(in *big.Int) decimal.Decimal {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	num := decimal.NewFromBigInt(in, 0)
	result := num.Div(mul)

	return result
}

func (td *ERC20TokenDetail) RawAmount(in float64) *big.Int {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.New((td.Decimals), 1))
	res, _ := new(big.Int).SetString(decimal.NewFromFloat(in).Mul(mul).String(), 10)
	return res
}
