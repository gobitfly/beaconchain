package enums

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

type Enum interface {
	Int() int
}

// Factory interface for creating enum values from strings
type EnumFactory[T Enum] interface {
	Enum
	NewFromString(string) T
}

func IsInvalidEnum(e Enum) bool {
	return e.Int() == -1
}

type SearchType int

// all possible search types
const (
	Name SearchType = iota // default
	Integer
	EthereumAddress
	WithdrawalCredential
	EnsName
	NonEmpty
	Graffiti
	Cursor
	Email
	Password
	EmailUserToken
	JsonContentType
	// Validator Dashboard
	ValidatorDashboardPublicId
	ValidatorPublicKeyWithPrefix
	ValidatorPublicKey
)

type Searchable interface {
	GetSearches() []SearchType
	SetSearchType(st SearchType, b bool) Searchable
	SetSearchValue(s string)
}

type BasicSearch struct {
	Value string // searched string
}

func (bs *BasicSearch) SetSearchValue(s string) {
	if bs == nil {
		log.Warnf("BasicSearch is nil, can't apply search: %s", s)
		return
	}
	bs.Value = s
}

// optional: implement in embedding structs to limit regex pattern matches
func (bs *BasicSearch) GetSearches() []SearchType {
	return []SearchType{
		Name,
		Integer,
		EthereumAddress,
		WithdrawalCredential,
		EnsName,
		NonEmpty,
		Graffiti,
		Cursor,
		Email,
		Password,
		EmailUserToken,
		JsonContentType,
		ValidatorDashboardPublicId,
		ValidatorPublicKeyWithPrefix,
		ValidatorPublicKey,
	}
}

type AdInsertMode int

var _ EnumFactory[AdInsertMode] = AdInsertMode(0)

const (
	AdInsertBefore  AdInsertMode = iota
	AdInsertAfter   AdInsertMode = iota
	AdInsertReplace AdInsertMode = iota
	AdInsertInsert  AdInsertMode = iota
)

func (c AdInsertMode) Int() int {
	return int(c)
}

func (AdInsertMode) NewFromString(s string) AdInsertMode {
	switch s {
	case "before":
		return AdInsertBefore
	case "after":
		return AdInsertAfter
	case "replace":
		return AdInsertReplace
	case "insert":
		return AdInsertInsert
	default:
		return AdInsertMode(-1)
	}
}

var AdInsertModes = struct {
	Before  AdInsertMode
	After   AdInsertMode
	Replace AdInsertMode
	Insert  AdInsertMode
}{
	AdInsertBefore,
	AdInsertAfter,
	AdInsertReplace,
	AdInsertInsert,
}

// ----------------
// Postgres sort direction enum
// SortOrder represents the sorting order, either ascending or descending.
type SortOrder int

// Constants for the sorting order.
const (
	ASC SortOrder = iota
	DESC
)

// String method converts SortOrder to string representation.
func (s SortOrder) String() string {
	if s == ASC {
		return "ASC"
	}
	return "DESC"
}

// Invert method inverts the sorting order.
func (s SortOrder) Invert() SortOrder {
	if s == ASC {
		return DESC
	}
	return ASC
}

var SortOrderColumns = struct {
	Asc  SortOrder
	Desc SortOrder
}{
	ASC,
	DESC,
}

// ----------------
// Time Periods

type TimePeriod int

const (
	AllTime TimePeriod = iota
	Last1h
	Last24h
	Last7d
	Last30d
)

func (t TimePeriod) Int() int {
	return int(t)
}

func (TimePeriod) NewFromString(s string) TimePeriod {
	switch s {
	case "all_time":
		return AllTime
	case "last_1h":
		return Last1h
	case "last_24h":
		return Last24h
	case "last_7d":
		return Last7d
	case "last_30d":
		return Last30d
	default:
		return TimePeriod(-1)
	}
}

var TimePeriods = struct {
	AllTime TimePeriod
	Last1h  TimePeriod
	Last24h TimePeriod
	Last7d  TimePeriod
	Last30d TimePeriod
}{
	AllTime,
	Last1h,
	Last24h,
	Last7d,
	Last30d,
}

func (t TimePeriod) Duration() time.Duration {
	day := 24 * time.Hour
	switch t {
	case Last1h:
		return time.Hour
	case Last24h:
		return day
	case Last7d:
		return 7 * day
	case Last30d:
		return 30 * day
	default:
		return 0
	}
}

// ----------------
// Validator Duties

type ValidatorDuty int

var _ EnumFactory[ValidatorDuty] = ValidatorDuty(0)

const (
	DutyNone ValidatorDuty = iota
	DutySync
	DutyProposal
	DutySlashed
)

func (d ValidatorDuty) Int() int {
	return int(d)
}

func (ValidatorDuty) NewFromString(s string) ValidatorDuty {
	switch s {
	case "":
		return DutyNone
	case "sync":
		return DutySync
	case "proposal":
		return DutyProposal
	case "slashed":
		return DutySlashed
	default:
		return ValidatorDuty(-1)
	}
}

var ValidatorDuties = struct {
	None     ValidatorDuty
	Sync     ValidatorDuty
	Proposal ValidatorDuty
	Slashed  ValidatorDuty
}{
	DutyNone,
	DutySync,
	DutyProposal,
	DutySlashed,
}

// ----------------
// Chart Aggregation Interval

type ChartAggregation int

var _ EnumFactory[ChartAggregation] = ChartAggregation(0)

const (
	IntervalEpoch ChartAggregation = iota
	IntervalHourly
	IntervalDaily
	IntervalWeekly
)

func (c ChartAggregation) Int() int {
	return int(c)
}

func (ChartAggregation) NewFromString(s string) ChartAggregation {
	switch s {
	case "epoch":
		return IntervalEpoch
	case "", "hourly":
		return IntervalHourly
	case "daily":
		return IntervalDaily
	case "weekly":
		return IntervalWeekly
	default:
		return ChartAggregation(-1)
	}
}

var ChartAggregations = struct {
	Epoch  ChartAggregation
	Hourly ChartAggregation
	Daily  ChartAggregation
	Weekly ChartAggregation
}{
	IntervalEpoch,
	IntervalHourly,
	IntervalDaily,
	IntervalWeekly,
}

func (c ChartAggregation) Duration(secondsPerEpoch uint64) time.Duration {
	switch c {
	case IntervalEpoch:
		return time.Second * time.Duration(secondsPerEpoch)
	case IntervalHourly:
		return time.Hour
	case IntervalDaily:
		return 24 * time.Hour
	case IntervalWeekly:
		return 7 * 24 * time.Hour
	default:
		return 0
	}
}
