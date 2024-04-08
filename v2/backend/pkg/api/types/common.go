package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// frontend can ignore ApiResponse type, it's just for the backend

type Paging struct {
	PrevCursor string `json:"prev_cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	TotalCount uint64 `json:"total_count,omitempty"`
}
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

type PubKey string
type Hash string // blocks, txs etc.

type Address struct {
	Hash Hash   `json:"hash"`
	Ens  string `json:"ens,omitempty"`
}
type LuckItem struct {
	Percent  float64       `json:"percent"`
	Expected time.Time     `json:"expected"`
	Average  time.Duration `json:"average"`
}

type Luck struct {
	Proposal LuckItem `json:"proposal"`
	Sync     LuckItem `json:"sync"`
}

type StatusCount struct {
	Success uint64 `json:"success"`
	Failed  uint64 `json:"failed"`
}

type ClElValue[T any] struct {
	El T `json:"el"`
	Cl T `json:"cl"`
}

type PeriodicValues[T any] struct {
	AllTime T `json:"all_time"`
	Last24h T `json:"last_24h"`
	Last7d  T `json:"last_7d"`
	Last30d T `json:"last_30d"`
}

type ChartSeries[T int | string] struct {
	Id    T         `json:"id"`              // id may be a string or an int
	Stack string    `json:"stack,omitempty"` // for stacking bar charts
	Data  []float64 `json:"data"`            // y-axis values
}

type ChartData[T int | string] struct {
	Categories []uint64         `json:"categories"` // x-axis
	Series     []ChartSeries[T] `json:"series"`
}

type SearchResult struct {
	Type      string `json:"type"`
	ChainId   uint64 `json:"chain_id"`
	HashValue string `json:"hash_value,omitempty"`
	NumValue  uint64 `json:"num_value,omitempty"`
	StrValue  string `json:"str_value,omitempty"`
}

type SearchResponse struct {
	Data []SearchResult `json:"data"`
}

type ValidatorHistoryEvent struct {
	Status string          `json:"status" tstype:"'success' | 'partial' | 'failed'"`
	Income decimal.Decimal `json:"income"`
}

type ValidatorHistoryProposal struct {
	Status                       string          `json:"status" tstype:"'success' | 'partial' | 'failed' | 'orphaned'"`
	ElIncome                     decimal.Decimal `json:"el_income"`
	ClAttestationInclusionIncome decimal.Decimal `json:"cl_attestation_inclusion_income"`
	ClSyncInclusionIncome        decimal.Decimal `json:"cl_sync_inclusion_income"`
	ClSlashingInclusionIncome    decimal.Decimal `json:"cl_slashing_inclusion_income"`
}

type ValidatorHistoryDuties struct {
	AttestationSource *ValidatorHistoryEvent    `json:"attestation_source,omitempty"`
	AttestationTarget *ValidatorHistoryEvent    `json:"attestation_target,omitempty"`
	AttestationHead   *ValidatorHistoryEvent    `json:"attestation_head,omitempty"`
	Sync              *ValidatorHistoryEvent    `json:"sync,omitempty"`
	Slashing          *ValidatorHistoryEvent    `json:"slashing,omitempty"`
	Proposal          *ValidatorHistoryProposal `json:"proposal,omitempty"`

	SyncCount uint64 `json:"sync_count,omitempty"` // count of successful sync duties for the epoch
}
