package database

const (
	defaultBatchSize = 10_000
	defaultLimit     = 100
)

type options struct {
	OpenRange       bool
	OpenCloseRange  bool
	ClosedOpenRange bool
	Limit           int64
	BatchSize       int64
	StatsReporter   func(msg string, args ...any)
	RowKeyFilter    string

	FamilyFilter         string
	ColumnFilter         string
	TimestampRangeFilter []int64

	WithoutValue bool
}

func newOptions(opts []Option) options {
	options := options{
		OpenRange:            false,
		OpenCloseRange:       false,
		ClosedOpenRange:      false,
		Limit:                defaultLimit,
		BatchSize:            defaultBatchSize,
		StatsReporter:        nil,
		RowKeyFilter:         "",
		FamilyFilter:         "",
		ColumnFilter:         "",
		TimestampRangeFilter: nil,
		WithoutValue:         false,
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

type Option func(opts *options)

func WithRowKeyFilter(regex string) Option {
	return func(opts *options) {
		opts.RowKeyFilter = regex
	}
}

func WithFamilyFilter(regex string) Option {
	return func(opts *options) {
		opts.FamilyFilter = regex
	}
}

func WithColumnFilter(regex string) Option {
	return func(opts *options) {
		opts.ColumnFilter = regex
	}
}

func WithTimestampRangeFilter(start, end int64) Option {
	return func(opts *options) {
		opts.TimestampRangeFilter = []int64{start, end}
	}
}

// WithOpenRange will contain all keys greater than the
// start and less than the end: (start, end).
func WithOpenRange() Option {
	return func(opts *options) {
		opts.OpenRange = true
	}
}

// WithClosedOpenRangeOption will contain all keys greater than or
// equal to the start and less than the end: [start, end).
func WithClosedOpenRangeOption() Option {
	return func(opts *options) {
		opts.ClosedOpenRange = true
	}
}

// WithOpenCloseRange will contain all keys greater than
// the start and less than or equal to the end: (start, end].
func WithOpenCloseRange() Option {
	return func(opts *options) {
		opts.OpenCloseRange = true
	}
}

func WithLimit(limit int64) Option {
	return func(opts *options) {
		opts.Limit = limit
	}
}

func WithBatchSize(size int64) Option {
	return func(opts *options) {
		opts.BatchSize = size
	}
}

func WithStats(reporter StatsReporter) Option {
	return func(opts *options) {
		opts.StatsReporter = reporter
	}
}

func WithoutValue() Option {
	return func(opts *options) {
		opts.WithoutValue = true
	}
}

type StatsReporter func(msg string, args ...any)
