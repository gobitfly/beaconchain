package types

import (
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
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
	Hash       Hash   `json:"hash"`
	IsContract bool   `json:"is_contract"`
	Ens        string `json:"ens,omitempty"`
	Label      string `json:"label,omitempty"`
}

type LuckItem struct {
	Percent                float64 `json:"percent"`
	ExpectedTimestamp      uint64  `json:"expected_timestamp"`
	AverageIntervalSeconds uint64  `json:"average_interval_seconds"`
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

type PercentageDetails[T any] struct {
	Percentage float64 `json:"percentage"`
	MinValue   T       `json:"min_value"`
	MaxValue   T       `json:"max_value"`
}

type ChartSeries[I int | string, D float64 | decimal.Decimal] struct {
	Id       I      `json:"id"`                 // id may be a string or an int
	Property string `json:"property,omitempty"` // for stacking bar charts
	Data     []D    `json:"data"`               // y-axis values
}

type ChartData[I int | string, D float64 | decimal.Decimal] struct {
	Categories []uint64            `json:"categories"` // x-axis
	Series     []ChartSeries[I, D] `json:"series"`
}

type ValidatorHistoryEvent struct {
	Status string          `json:"status" tstype:"'success' | 'partial' | 'failed'" faker:"oneof: success, partial, failed"`
	Income decimal.Decimal `json:"income"`
}

type ValidatorHistoryProposal struct {
	Status                       string          `json:"status" tstype:"'success' | 'partial' | 'failed'" faker:"oneof: success, partial, failed"`
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

type ChainConfig struct {
	ChainId uint64 `json:"chain_id"`
	Name    string `json:"name"`
	// TODO: add more fields, depending on what frontend needs
}

type VDBPublicId struct {
	PublicId      string `json:"public_id"`
	DashboardId   int    `json:"-"`
	Name          string `json:"name,omitempty"`
	ShareSettings struct {
		ShareGroups bool `json:"share_groups"`
	} `json:"share_settings"`
}

type ChartHistorySeconds struct {
	Epoch  uint64 `json:"epoch"`
	Hourly uint64 `json:"hourly"`
	Daily  uint64 `json:"daily"`
	Weekly uint64 `json:"weekly"`
}

type IndexEpoch struct {
	Index uint64 `json:"index"`
	Epoch uint64 `json:"epoch"`
}

type IndexBlocks struct {
	Index  uint64   `json:"index"`
	Blocks []uint64 `json:"blocks"`
}

type IndexSlots struct {
	Index uint64   `json:"index"`
	Slots []uint64 `json:"slots"`
}

type ValidatorStateCounts struct {
	Online  uint64 `json:"online"`
	Offline uint64 `json:"offline"`
	Pending uint64 `json:"pending"`
	Exited  uint64 `json:"exited"`
	Slashed uint64 `json:"slashed"`
}

type SearchType int

// all possible search types
const (
	SearchTypeInteger SearchType = iota
	SearchTypeName
	SearchTypeEthereumAddress
	SearchTypeWithdrawalCredential
	SearchTypeEnsName
	SearchTypeGraffiti
	SearchTypeCursor
	SearchTypeEmail
	SearchTypePassword
	SearchTypeEmailUserToken
	SearchTypeJsonContentType
	// Validator Dashboard
	SearchTypeValidatorDashboardPublicId
	SearchTypeValidatorPublicKeyWithPrefix
	SearchTypeValidatorPublicKey
)

type Searchable interface {
	GetSearches() []SearchType // optional: implement in embedding structs to limit regex pattern matches
	SetSearchValue(s string)   // optional: implement for custom behavior
	SetSearchType(st SearchType, b bool)
	IsEnabled() bool
	HasAnyMatches() bool
}

// not to be used directly, only for embedding
type basicSearch struct {
	types map[SearchType]bool
	value string
}

func (bs *basicSearch) SetSearchValue(s string) {
	if bs == nil {
		log.Warnf("BasicSearch is nil, can't apply search: %s", s)
		return
	}
	bs.value = s
}

func (bs *basicSearch) SetSearchType(st SearchType, b bool) {
	if bs == nil {
		bs = &basicSearch{}
	}
	bs.types[st] = b
}

func (bs *basicSearch) IsEnabled() bool {
	return bs != nil && bs.value != ""
}

func (bs *basicSearch) HasAnyMatches() bool {
	for _, v := range bs.types {
		if v {
			return true
		}
	}
	return false
}

func (bs *basicSearch) GetSearches() []SearchType {
	return []SearchType{
		SearchTypeName,
		SearchTypeInteger,
		SearchTypeEthereumAddress,
		SearchTypeWithdrawalCredential,
		SearchTypeEnsName,
		SearchTypeGraffiti,
		SearchTypeCursor,
		SearchTypeEmail,
		SearchTypePassword,
		SearchTypeEmailUserToken,
		SearchTypeJsonContentType,
		SearchTypeValidatorDashboardPublicId,
		SearchTypeValidatorPublicKeyWithPrefix,
		SearchTypeValidatorPublicKey,
	}
}

type baseSearchResult struct {
	Enabled bool
}

type SearchNumber struct {
	baseSearchResult
	Value uint64
}

type SearchString struct {
	baseSearchResult
	Value string
}

func (bs *basicSearch) AsNumber(st SearchType) SearchNumber {
	if !bs.IsEnabled() {
		log.Warn("tried accessing invalid search", 1)
		return SearchNumber{}
	}

	if !bs.types[st] {
		return SearchNumber{}
	}

	switch st {
	case SearchTypeInteger:
		number, err := strconv.ParseUint(bs.value, 10, 64)
		if err != nil {
			log.Error(err, "error converting search value, check regex parsing", 0)
			return SearchNumber{}
		}
		return SearchNumber{baseSearchResult{true}, number}
	}
	return SearchNumber{}
}

func (bs *basicSearch) AsString(st SearchType) SearchString {
	if !bs.IsEnabled() {
		log.Warn("tried accessing invalid search", 1)
		return SearchString{}
	}

	// apply custom conversion by type (e.g. prefix search term with 0x)
	switch st {
	case SearchTypeValidatorPublicKeyWithPrefix:
		return SearchString{baseSearchResult{true}, strings.ToLower(bs.value)}
	default:
		return SearchString{baseSearchResult{true}, bs.value}
	}
}

// commonly used table search options
type SearchTableByIndexPubkeyGroup struct {
	*basicSearch
	// conditionals
	DashboardId VDBId
}

func (s SearchTableByIndexPubkeyGroup) Index() SearchNumber {
	return s.AsNumber(SearchTypeInteger)
}

func (s SearchTableByIndexPubkeyGroup) Pubkey() SearchString {
	return s.AsString(SearchTypeValidatorPublicKeyWithPrefix)
}

func (s SearchTableByIndexPubkeyGroup) Group() SearchString {
	if s.DashboardId.AggregateGroups || s.DashboardId.Validators != nil {
		return SearchString{baseSearchResult{false}, ""}
	}
	return s.AsString(SearchTypeName)
}

func (s SearchTableByIndexPubkeyGroup) GetSearches() []SearchType {
	return []SearchType{
		SearchTypeInteger,
		SearchTypeName,
		SearchTypeValidatorPublicKeyWithPrefix,
	}
}

// custom to filter out certain group searches
func (s SearchTableByIndexPubkeyGroup) HasAnyMatches() bool {
	return s.Group().Enabled || s.basicSearch.HasAnyMatches()
}
