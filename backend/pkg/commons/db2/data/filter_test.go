package data

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestQueryFilter(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    string
		err     string
	}{
		{
			name: "all",
			want: "all:<address>",
		},
		{
			name: "err for ByMethod and ByAsset",
			options: []Option{
				ByMethod(""),
				ByAsset(common.Address{}),
			},
			err: "cannot filter by method and by asset together",
		},
		{
			name: "only sent",
			options: []Option{
				OnlySent(),
			},
			want: "out:<address>",
		},
		{
			name: "received",
			options: []Option{
				OnlyReceived(),
			},
			want: "in:<address>",
		},
		{
			name: "sent",
			options: []Option{
				OnlySent(),
			},
			want: "out:<address>",
		},
		{
			name: "sent to",
			options: []Option{
				OnlySent(),
				With(common.Address{}),
			},
			want: "out:with:<address>:0000000000000000000000000000000000000000",
		},
		{
			name: "on a chain ID",
			options: []Option{
				ByChainID("1234"),
			},
			want: "all:chainID:<address>:1234",
		},
		{
			name: "sent on a chain ID",
			options: []Option{
				OnlySent(),
				ByChainID("1234"),
			},
			want: "out:chainID:<address>:1234",
		},
		{
			name: "received on a chain ID",
			options: []Option{
				OnlyReceived(),
				ByChainID("1234"),
			},
			want: "in:chainID:<address>:1234",
		},
		{
			name: "received TX on a chain ID",
			options: []Option{
				OnlyTransactions(),
				OnlyReceived(),
				ByChainID("1234"),
			},
			want: "in:chainID:TX:<address>:1234",
		},
		{
			name: "received method on a chain ID",
			options: []Option{
				ByMethod("bar"),
				OnlyReceived(),
				ByChainID("1234"),
			},
			want: "in:chainID:TX:method:<address>:1234:bar",
		},
		{
			name: "received method on a chain ID",
			options: []Option{
				ByAsset(common.Address{}),
				OnlyReceived(),
				ByChainID("1234"),
			},
			want: "in:chainID:ERC20:asset:<address>:1234:0000000000000000000000000000000000000000",
		},
		{
			name: "with time range",
			options: []Option{
				WithTimeRange(timestamppb.New(t0), timestamppb.New(t1)),
			},
			want: "all:<address>:" + reversePaddedTimestamp(timestamppb.New(t1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := newQueryFilter(apply(tt.options))
			if err != nil {
				if got, want := err.Error(), tt.err; got != want {
					t.Errorf("got error %v, want %v", got, want)
				}
				return
			}
			addr := common.Address{}
			query := filter.get(addr)
			if got, want := query, strings.ReplaceAll(tt.want, "<address>", toHex(addr.Bytes())); got != want {
				t.Errorf("get() = %v, want %v", got, want)
			}
		})
	}
}
