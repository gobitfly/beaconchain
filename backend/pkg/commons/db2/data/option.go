package data

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type Option interface {
	apply(*options)
}

type options struct {
	from, to        *timestamp.Timestamp
	method          *string
	ignoreTxs       bool
	ignoreTransfers bool
	onlySent        bool
	onlyReceived    bool
	asset           *common.Address
}

func apply(opts []Option) options {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	return options
}

type byTimeRangeOption struct {
	from *timestamp.Timestamp
	to   *timestamp.Timestamp
}

func WithTimeRange(from *timestamp.Timestamp, to *timestamp.Timestamp) Option {
	return byTimeRangeOption{from: from, to: to}
}

func (r byTimeRangeOption) apply(opts *options) {
	opts.from = r.from
	opts.to = r.to
}

type byMethodOption struct {
	method string
}

func ByMethod(method string) Option {
	return byMethodOption{method: method}
}

func (r byMethodOption) apply(opts *options) {
	opts.method = &r.method
}

type ignoreTransactionsOption bool

func IgnoreTransactions() Option {
	return ignoreTransactionsOption(true)
}

func (r ignoreTransactionsOption) apply(opts *options) {
	opts.ignoreTxs = bool(r)
}

type ignoreTransfersOption bool

func IgnoreTransfers() Option {
	return ignoreTransfersOption(true)
}

func (r ignoreTransfersOption) apply(opts *options) {
	opts.ignoreTransfers = bool(r)
}

type onlySentOption bool

func OnlySent() Option {
	return onlySentOption(true)
}

func (r onlySentOption) apply(opts *options) {
	opts.onlySent = bool(r)
}

type onlyReceivedOption bool

func OnlyReceived() Option {
	return onlyReceivedOption(true)
}

func (r onlyReceivedOption) apply(opts *options) {
	opts.onlyReceived = bool(r)
}

type byAssetOption struct {
	asset common.Address
}

func ByAsset(asset common.Address) Option {
	return byAssetOption{asset: asset}
}

func (r byAssetOption) apply(opts *options) {
	opts.asset = &r.asset
}

func makeFilters(options options, typeFilter formatType) (filter, error) {
	var f chainFilter
	switch typeFilter {
	case typeTx:
		f = newChainFilterTx()
	case typeTransfer:
		f = newChainFilterTransfer()
	default:
		return nil, fmt.Errorf("unknown filter type: %s", typeFilter)
	}
	if options.onlyReceived && options.onlySent {
		options.onlyReceived = false
		options.onlySent = false
	}
	if options.onlySent {
		if err := f.addBySent(); err != nil {
			return nil, err
		}
	}
	if options.onlyReceived {
		if err := f.addByReceived(); err != nil {
			return nil, err
		}
	}
	if options.method != nil {
		if err := f.addByMethod(*options.method); err != nil {
			return nil, err
		}
	}
	if options.asset != nil {
		if err := f.addByAsset(*options.asset); err != nil {
			return nil, err
		}
	}
	if options.from != nil && options.to != nil {
		if err := f.addTimeRange(options.from, options.to); err != nil {
			return nil, err
		}
	}
	if err := f.valid(); err != nil {
		return nil, err
	}

	return f, nil
}
