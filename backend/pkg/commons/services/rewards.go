package services

import (
	"bytes"

	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jung-kurt/gofpdf"
	"github.com/lib/pq"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type rewardHistory struct {
	History       [][]string `json:"history"`
	TotalETH      string     `json:"total_eth"`
	TotalCurrency string     `json:"total_currency"`
	Validators    []uint64   `json:"validators"`
}

func GetValidatorHist(validatorArr []uint64, currency string, start uint64, end uint64) rewardHistory {
	var err error

	var pricesDb []types.Price
	// we get prices with a 1 day buffer to so we have no problems in different time zones
	var oneDay = uint64(24 * 60 * 60)

	if start == end { // no date range was provided, use the current day as ending boundary
		end = uint64(time.Now().Unix())
	}
	err = db.WriterDb.Select(&pricesDb,
		`select ts, eur, usd, gbp, cad, jpy, cny, aud from price where ts >= TO_TIMESTAMP($1) and ts <= TO_TIMESTAMP($2) order by ts desc`, start-oneDay, end+oneDay)
	if err != nil {
		log.Error(err, "error getting prices", 0, map[string]interface{}{"start": start, "end": end})
	}

	lowerBound := utils.TimeToDay(start)
	upperBound := utils.TimeToDay(end)

	// As the genesis timestamp is in the middle of the day and we get timestamps from the ui from the start of the day we add one to get the correct day,
	// except for the beaconchain day where we get a timestamp lower then the genesis day. The TimeToDay function still would transform it to 0 (and not -1) so we don't need to add one.
	if start > utils.Config.Chain.GenesisTimestamp {
		lowerBound++
	}

	income, err := db.GetValidatorIncomeHistory(validatorArr, lowerBound, upperBound, cache.LatestFinalizedEpoch.Get())
	if err != nil {
		log.Error(err, "error getting income history for validator hist", 0, map[string]interface{}{"validators": validatorArr, "lowerBound": lowerBound, "upperBound": upperBound})
	}

	prices := map[string]float64{}
	for _, item := range pricesDb {
		date := fmt.Sprintf("%v", item.TS)
		date = strings.Split(date, " ")[0]
		switch currency {
		case "eur":
			prices[date] = item.EUR
		case "usd":
			prices[date] = item.USD
		case "gbp":
			prices[date] = item.GBP
		case "cad":
			prices[date] = item.CAD
		case "cny":
			prices[date] = item.CNY
		case "jpy":
			prices[date] = item.JPY
		case "aud":
			prices[date] = item.AUD
		default:
			prices[date] = item.USD
			currency = "usd"
		}
	}

	data := make([][]string, len(income))
	tETH := 0.0
	tCur := 0.0

	for i, item := range income {
		key := fmt.Sprintf("%v", utils.DayToTime(item.Day))
		key = strings.Split(key, " ")[0]
		iETH := float64(item.ClRewards) / 1e9
		tETH += iETH
		iCur := iETH * prices[key]
		tCur += iCur
		data[i] = []string{
			key,
			addCommas(float64(item.EndBalance.Int64)/1e9, "%.5f"),                           // end of day balance
			addCommas(iETH, "%.5f"),                                                         // income of day ETH
			fmt.Sprintf("%s %s", strings.ToUpper(currency), addCommas(prices[key], "%.2f")), //price will default to 0 if key does not exist
			fmt.Sprintf("%s %s", strings.ToUpper(currency), addCommas(iCur, "%.2f")),        // income of day Currency
		}
	}

	return rewardHistory{
		History:       data,
		TotalETH:      addCommas(tETH, "%.5f"),
		TotalCurrency: fmt.Sprintf("%s %s", strings.ToUpper(currency), addCommas(tCur, "%.2f")),
		Validators:    validatorArr,
	}
}

func addCommas(balance float64, decimals string) string {
	p := message.NewPrinter(language.English)
	rb := []rune(p.Sprintf(decimals, balance))
	// remove trailing zeros
	if rb[len(rb)-2] == '.' || rb[len(rb)-3] == '.' {
		for rb[len(rb)-1] == '0' {
			rb = rb[:len(rb)-1]
		}
		if rb[len(rb)-1] == '.' {
			rb = rb[:len(rb)-1]
		}
	}

	return string(rb)
}

func GeneratePdfReport(hist rewardHistory, currency string) []byte {
	data := hist.History

	if len(data) <= 0 {
		log.Warnf("Can't generate PDF for Empty Slice")
		return []byte{}
	}

	sort.Slice(data, func(p, q int) bool {
		i, err := time.Parse("2006-01-02", data[p][0])
		if err != nil {
			return false
		}

		i2, err := time.Parse("2006-01-02", data[q][0])
		if err != nil {
			return false
		}
		return i2.Before(i)
	})

	validators := hist.Validators

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTopMargin(15)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(80, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("Beaconcha.in Income History (%s - %s)", data[len(data)-1][0], data[0][0]), "", 0, "C", false, 0, "")
		// pdf.Ln(-1)
	}, true)

	pdf.AddPage()
	pdf.SetFont("Times", "", 9)

	// generating the table
	const (
		colCount = 5
		colWd    = 40.0
		marginH  = 5.0
		lineHt   = 5.5
		maxHt    = 5
	)

	pdf.SetTextColor(24, 24, 24)
	pdf.SetFillColor(255, 255, 255)
	// pdf.Ln(-1)
	pdf.CellFormat(0, maxHt, fmt.Sprintf("Income For Timeframe %s | %s", hist.TotalETH, hist.TotalCurrency), "", 0, "CM", true, 0, "")

	header := [colCount]string{"Date", "Balance", "Income", "ETH Value", fmt.Sprintf("Income (%v)", currency)}

	// pdf.SetMargins(marginH, marginH, marginH)
	pdf.Ln(10)
	pdf.SetTextColor(224, 224, 224)
	pdf.SetFillColor(64, 64, 64)
	pdf.Cell(-5, 0, "")
	for col := range colCount {
		pdf.CellFormat(colWd, maxHt, header[col], "1", 0, "CM", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(24, 24, 24)
	pdf.SetFillColor(255, 255, 255)

	// Rows
	y := pdf.GetY()

	for i, row := range data {
		pdf.SetTextColor(24, 24, 24)
		pdf.SetFillColor(255, 255, 255)
		x := marginH
		if i%47 == 0 && i != 0 {
			pdf.AddPage()
			y = pdf.GetY()
		}
		for col := range colCount {
			if i%2 != 0 {
				pdf.SetFillColor(191, 191, 191)
			}
			pdf.Rect(x, y, colWd, maxHt, "D")
			cellY := y
			pdf.SetXY(x, cellY)
			pdf.CellFormat(colWd, maxHt, row[col], "", 0,
				"LM", true, 0, "")
			x += colWd
		}
		y += maxHt
	}

	// adding a footer
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetTextColor(24, 24, 24)
	pdf.SetFillColor(255, 255, 255)
	// pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, maxHt, "Validators", "", 0, "CM", true, 0, "")
	pdf.Ln(10)
	pdf.SetFont("Times", "", 9)

	const (
		vColCount = 4
		vColWd    = 50.0
	)
	vHeader := [vColCount]string{"Index", "Activation Balance", "Balance", "Last Attestation"}

	// pdf.SetMargins(marginH, marginH, marginH)
	// pdf.Ln(10)
	pdf.SetTextColor(224, 224, 224)
	pdf.SetFillColor(64, 64, 64)
	pdf.Cell(-5, 0, "")
	for col := range vColCount {
		pdf.CellFormat(vColWd, maxHt, vHeader[col], "1", 0, "CM", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(24, 24, 24)
	pdf.SetFillColor(255, 255, 255)

	y = pdf.GetY()

	for i, row := range getValidatorDetails(validators) {
		pdf.SetTextColor(24, 24, 24)
		pdf.SetFillColor(255, 255, 255)
		x := marginH

		if i%47 == 0 && i != 0 {
			pdf.AddPage()
			y = pdf.GetY()
		}

		for col := range vColCount {
			if i%2 != 0 {
				pdf.SetFillColor(191, 191, 191)
			}
			pdf.Rect(x, y, vColWd, maxHt, "D")
			cellY := y
			pdf.SetXY(x, cellY)
			pdf.CellFormat(vColWd, maxHt, row[col], "", 0,
				"LM", true, 0, "")
			x += vColWd
		}
		y += maxHt
	}

	// adding a footer
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	buf := new(bytes.Buffer)
	err := pdf.Output(buf)
	if err != nil {
		log.Error(err, "error generating pdf", 0, map[string]interface{}{"hist": hist})
		return []byte{}
	}

	return buf.Bytes()
}

func GetPdfReport(validatorArr []uint64, currency string, start uint64, end uint64) []byte {
	hist := GetValidatorHist(validatorArr, currency, start, end)
	return GeneratePdfReport(hist, currency)
}

func getValidatorDetails(validators []uint64) [][]string {
	validatorFilter := pq.Array(validators)
	var data []types.ValidatorPageData
	err := db.WriterDb.Select(&data,
		`SELECT validatorindex, balanceactivation
		 FROM validators
		 WHERE validatorindex = ANY($1)
		 ORDER BY validatorindex ASC`, validatorFilter)
	if err != nil {
		log.Error(err, "error getting validators data", 0, map[string]interface{}{"validators": validators})
		return [][]string{}
	}

	latestEpoch := cache.LatestEpoch.Get()
	balances, err := db.BigtableClient.GetValidatorBalanceHistory(validators, latestEpoch, latestEpoch)
	if err != nil {
		log.Error(err, "error getting validator balance history", 0, log.Fields{
			"validators":  validators,
			"latestEpoch": latestEpoch,
		})
		return [][]string{}
	}

	lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots(validators)
	if err != nil {
		log.Error(err, "error getting validator balance history", 0, log.Fields{
			"validators":  validators,
			"latestEpoch": latestEpoch,
		})
		return [][]string{}
	}

	for i, validator := range data {
		validator.LastAttestationSlot = lastAttestationSlots[validator.ValidatorIndex]
		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}
			if validator.ValidatorIndex == balanceIndex {
				validator.CurrentBalance = balance[0].Balance
				validator.EffectiveBalance = balance[0].EffectiveBalance
			}
		}
		data[i] = validator
	}

	result := [][]string{}
	for _, item := range data {
		la_date := "N/a"
		if item.LastAttestationSlot > 0 {
			la_time := utils.SlotToTime(item.LastAttestationSlot)
			la_date = la_time.Format(time.RFC822)
		}
		result = append(result, []string{
			fmt.Sprintf("%d", item.ValidatorIndex),
			addCommas(float64(item.BalanceActivation)/float64(1e9), "%.5f"),
			addCommas(float64(item.CurrentBalance)/float64(1e9), "%.5f"),
			la_date,
		})
	}

	return result
}
