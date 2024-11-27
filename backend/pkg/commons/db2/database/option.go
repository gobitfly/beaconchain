package database

const (
	defaultBatchSize = 10000
	defaultLimit     = 100
)

type options struct {
	OpenRange      bool
	OpenCloseRange bool
	Limit          int64
	BatchSize      int64
	StatsReporter  func(msg string, args ...any)
}

func apply(opts []Option) options {
	options := options{
		OpenRange:      false,
		OpenCloseRange: false,
		BatchSize:      defaultBatchSize,
		Limit:          defaultLimit,
	}
	for _, o := range opts {
		o.apply(&options)
	}
	return options
}

type Option interface {
	apply(*options)
}

type openRangeOption bool

func (r openRangeOption) apply(opts *options) {
	opts.OpenRange = bool(r)
}

func WithOpenRange(r bool) Option {
	return openRangeOption(r)
}

type openCloseRangeOption bool

func (r openCloseRangeOption) apply(opts *options) {
	opts.OpenCloseRange = bool(r)
}

func WithOpenCloseRange(r bool) Option {
	return openCloseRangeOption(r)
}

type limitOption int64

func (l limitOption) apply(opts *options) {
	opts.Limit = int64(l)
}

func WithLimit(l int64) Option {
	return limitOption(l)
}

type withBatchSize int64

func (l withBatchSize) apply(opts *options) {
	opts.BatchSize = int64(l)
}

func WithBatchSize(l int64) Option {
	return withBatchSize(l)
}

type statsOption StatsReporter

func (l statsOption) apply(opts *options) {
	opts.StatsReporter = l
}

func WithStats(reporter StatsReporter) Option {
	return statsOption(reporter)
}

type StatsReporter func(msg string, args ...any)
