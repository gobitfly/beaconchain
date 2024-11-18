package database

type options struct {
	OpenRange      bool
	OpenCloseRange bool
	Limit          int64
}

func apply(opts []Option) options {
	options := options{}
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
