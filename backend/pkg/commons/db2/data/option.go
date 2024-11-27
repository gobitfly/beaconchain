package data

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
)

type Option interface {
	apply(*options)
}

type options struct {
	from, to      *timestamp.Timestamp
	method        *string
	chainID       *string
	onlySent      bool
	onlyReceived  bool
	asset         *common.Address
	with          *common.Address
	onlyTxs       bool
	onlyTransfers bool
	statsReporter func(msg string, args ...any)
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

type byMethodOption string

func ByMethod(method string) Option {
	return byMethodOption(method)
}

func (r byMethodOption) apply(opts *options) {
	opts.method = (*string)(&r)
}

type byChainIDOption string

func ByChainID(chainID string) Option {
	return byChainIDOption(chainID)
}

func (r byChainIDOption) apply(opts *options) {
	opts.chainID = (*string)(&r)
}

type onlyTransactionsOption bool

func OnlyTransactions() Option {
	return onlyTransactionsOption(true)
}

func (r onlyTransactionsOption) apply(opts *options) {
	opts.onlyTxs = bool(r)
}

type onlyTransfersOption bool

func OnlyTransfers() Option {
	return onlyTransfersOption(true)
}

func (r onlyTransfersOption) apply(opts *options) {
	opts.onlyTransfers = bool(r)
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

type withOption common.Address

func With(address common.Address) Option {
	return withOption(address)
}

func (r withOption) apply(opts *options) {
	opts.with = (*common.Address)(&r)
}

type withDatabaseStatsOption database.StatsReporter

func WithDatabaseStats(reporter database.StatsReporter) Option {
	return withDatabaseStatsOption(reporter)
}

func (r withDatabaseStatsOption) apply(opts *options) {
	opts.statsReporter = r
}
