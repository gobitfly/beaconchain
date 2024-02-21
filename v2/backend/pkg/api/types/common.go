package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// frontend can ignore this, it's just for the backend
type ApiResponse struct {
	Paging *Paging     `json:"paging,omitempty"`
	Data   interface{} `json:"data"`
}

type ApiErrorResponse struct {
	Error string `json:"error"`
}

type ApiDataResponse[T any] struct {
	Data T `json:"data"`
}
type ApiPagingResponse[T any] struct {
	Paging Paging `json:"paging"`
	Data   []T    `json:"data"`
}

type Paging struct {
	PrevCursor string `json:"prev_cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	TotalCount uint64 `json:"total_count,omitempty"`
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

type ClElUnion interface {
	float64 | decimal.Decimal
}

type ClElValue[T ClElUnion] struct {
	El T `json:"el"`
	Cl T `json:"cl"`
}

type PeriodicClElValues[T ClElUnion] struct {
	Total ClElValue[T] `json:"total"`
	Day   ClElValue[T] `json:"day"`
	Week  ClElValue[T] `json:"week"`
	Month ClElValue[T] `json:"month"`
	Year  ClElValue[T] `json:"year"`
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

type DashboardData struct {
	ValidatorDashboards []Dashboard `json:"validator_dashboards"`
	AccountDashboards   []Dashboard `json:"account_dashboards"`
}

type Dashboard struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
