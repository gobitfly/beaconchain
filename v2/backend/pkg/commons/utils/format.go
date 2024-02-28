package utils

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/shopspring/decimal"
)

func FormatAttestorAssignmentKey(AttesterSlot, CommitteeIndex, MemberIndex uint64) string {
	return fmt.Sprintf("%v-%v-%v", AttesterSlot, CommitteeIndex, MemberIndex)
}

func ClToCurrency(valIf interface{}, currency string) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ClCurrencyDivisor), 18)
	if currency == Config.Frontend.ClCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ClCurrency, currency)))
}

func ElToCurrency(valIf interface{}, currency string) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ElCurrencyDivisor), 18)
	if currency == Config.Frontend.ElCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ElCurrency, currency)))
}

func FormatElCurrencyString(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	return formatCurrencyString(ElToCurrency(value, Config.Frontend.ElCurrency), Config.Frontend.ElCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, truncateAndAddTooltip)
}

func FormatClCurrencyString(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	return formatCurrencyString(ClToCurrency(value, Config.Frontend.ClCurrency), Config.Frontend.ClCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, truncateAndAddTooltip)
}

func formatCurrencyString(valIf interface{}, valueCurrency, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	val := IfToDec(valIf)

	valPriced := val
	if valueCurrency != targetCurrency {
		valPriced = val.Mul(decimal.NewFromFloat(price.GetPrice(valueCurrency, targetCurrency)))
	}

	currencyStr := ""
	if showCurrencySymbol {
		currencyStr = " " + price.GetCurrencySymbol(targetCurrency)
	}

	amountStr := ""
	tooltipStartStr := ""
	tooltipEndStr := ""
	if truncateAndAddTooltip {
		amountStr = valPriced.Truncate(int32(digitsAfterComma)).String()

		// only add tooltip if the value is actually truncated
		valStr := valPriced.String()
		if valStr != amountStr {
			tooltipStartStr = fmt.Sprintf(`<span data-toggle="tooltip" data-placement="top" title="%s%s">`, valPriced, currencyStr)
			tooltipEndStr = `</span>`
		}

		// add trailing zeros to always have the same amount of digits after the comma
		dotIndex := strings.Index(valStr, ".")
		if dotIndex >= 0 {
			if !strings.Contains(amountStr, ".") {
				amountStr += "."
			}
			missingZeros := digitsAfterComma - (len(amountStr) - dotIndex - 1)
			if missingZeros > 0 {
				amountStr += strings.Repeat("0", missingZeros)
			}
		}
	} else {
		amountStr = valPriced.StringFixed(int32(digitsAfterComma))
	}

	plusSignStr := ""
	if showPlusSign && valPriced.Cmp(decimal.NewFromInt(0)) >= 0 {
		plusSignStr = "+"
	}

	return fmt.Sprintf(`%s%s%s%s%s`, tooltipStartStr, plusSignStr, amountStr, currencyStr, tooltipEndStr)
}

// IfToDec trys to parse given parameter to decimal.Decimal, it only logs on error
func IfToDec(valIf interface{}) decimal.Decimal {
	var err error
	var val decimal.Decimal
	switch v := valIf.(type) {
	case *float64:
		val = decimal.NewFromFloat(*v)
	case *int64:
		val = decimal.NewFromInt(*v)
	case *uint64:
		val, err = decimal.NewFromString(fmt.Sprintf("%v", *v))
	case int, int64, float64, uint64, *big.Float:
		val, err = decimal.NewFromString(fmt.Sprintf("%v", valIf))
	case []uint8:
		val = decimal.NewFromBigInt(new(big.Int).SetBytes(v), 0)
	case *big.Int:
		val = decimal.NewFromBigInt(v, 0)
	case decimal.Decimal:
		val = v
	default:
		log.Error(nil, "invalid value passed to IfToDec", 0, log.Fields{"type": reflect.TypeOf(valIf), "val": valIf})
	}
	if err != nil {
		log.Error(err, "invalid value passed to IfToDec", 0, log.Fields{"type": reflect.TypeOf(valIf), "val": valIf, "error": err})
	}
	return val
}

/*
  - FormatHashRaw will return a hash formated
    hash is required, trunc is optional.
    Only the first value in trunc_opt will be used.
    ATTENTION: IT TRUNCATES BY DEFAULT, PASS FALSE TO trunc_opt TO DISABLE
*/
func FormatHashRaw(hash []byte, trunc_opt ...bool) string {
	s := fmt.Sprintf("%#x", hash)
	if len(s) == 42 { // if it's an address, we checksum it (0x + 40)
		s = common.BytesToAddress(hash).Hex()
	}
	if len(s) >= 10 && (len(trunc_opt) < 1 || trunc_opt[0]) {
		return fmt.Sprintf("%s…%s", s[:6], s[len(s)-4:])
	}
	return s
}
