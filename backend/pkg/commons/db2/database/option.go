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
}

func newOptions(opts []Option) options {
	options := options{
		OpenRange:       false,
		OpenCloseRange:  false,
		ClosedOpenRange: false,
		Limit:           defaultLimit,
		BatchSize:       defaultBatchSize,
		StatsReporter:   nil,
		RowKeyFilter:    "",
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

func WithOpenRange(r bool) Option {
	return func(opts *options) {
		opts.OpenRange = r
	}
}

func WithClosedOpenRangeOption(r bool) Option {
	return func(opts *options) {
		opts.ClosedOpenRange = r
	}
}

func WithOpenCloseRange(r bool) Option {
	return func(opts *options) {
		opts.OpenCloseRange = r
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

type StatsReporter func(msg string, args ...any)
