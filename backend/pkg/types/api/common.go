package apitypes

import (
	"time"

	"github.com/shopspring/decimal"
)

type Paging struct {
	PrevCursor string `json:"prev_cursor"`
	NextCursor string `json:"next_cursor"`
}

type ApiResponse struct {
	Paging Paging      `json:"paging,omitempty"`
	Data   interface{} `json:"data"`
}

type ApiErrorResponse struct {
	Error string `json:"error"`
}

type PubKey string
type Hash string // blocks, txs etc.

type Address struct {
	Hash Hash   `json:"hash"`
	Ens  string `json:"ens,omitempty"`
}

type Luck struct {
	Proposal LuckItem `json:"proposal"`
	Sync     LuckItem `json:"sync"`
}
type LuckItem struct {
	Percent  float64       `json:"percent"`
	Expected time.Time     `json:"expected"`
	Average  time.Duration `json:"average"`
}

type StatusCount struct {
	Success uint64 `json:"success"`
	Failed  uint64 `json:"failed"`
}

type ClElValue struct {
	El decimal.Decimal `json:"el"`
	Cl decimal.Decimal `json:"cl"`
}

type ClElValueFloat struct {
	El float64 `json:"el"`
	Cl float64 `json:"cl"`
}

type PeriodicClElValues struct {
	Total ClElValue `json:"total"`
	Day   ClElValue `json:"day"`
	Week  ClElValue `json:"week"`
	Month ClElValue `json:"month"`
	Year  ClElValue `json:"year"`
}

type PeriodicClElValuesFloat struct {
	Total ClElValueFloat `json:"total"`
	Day   ClElValueFloat `json:"day"`
	Week  ClElValueFloat `json:"week"`
	Month ClElValueFloat `json:"month"`
	Year  ClElValueFloat `json:"year"`
}

type HighchartsSeries struct {
	Name string                `json:"name"`
	Data []HighchartsDataPoint `json:"data"`
}

type HighchartsDataPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type SearchResponse struct {
	Data []SearchResult `json:"data"`
}

type SearchResult struct {
	Type      string `json:"type"`
	ChainId   uint64 `json:"chain_id"`
	HashValue string `json:"hash_value,omitempty"`
	NumValue  uint64 `json:"num_value,omitempty"`
	StrValue  string `json:"str_value,omitempty"`
}
