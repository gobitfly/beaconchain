package data

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name       string
		filter     chainFilter
		add        func(chainFilter) error
		expectErr  bool
		expectType filterType
	}{
		{
			name:   "tx by asset should err",
			filter: newChainFilterTx(),
			add: func(c chainFilter) error {
				return c.addByAsset(common.Address{})
			},
			expectErr: true,
		},
		{
			name:   "tx invalid time range",
			filter: newChainFilterTx(),
			add: func(c chainFilter) error {
				return c.addTimeRange(nil, nil)
			},
			expectErr: true,
		},
		{
			name:   "transfer by method should err",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				return c.addByMethod("")
			},
			expectErr: true,
		},
		{
			name:   "transfer by asset sent",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addByAsset(common.Address{}); err != nil {
					return err
				}
				return c.addBySent()
			},
			expectType: byAssetSent,
		},
		{
			name:   "transfer by asset received",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addByAsset(common.Address{}); err != nil {
					return err
				}
				return c.addByReceived()
			},
			expectType: byAssetReceived,
		},
		{
			name:   "transfer by sent asset",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addBySent(); err != nil {
					return err
				}
				return c.addByAsset(common.Address{})
			},
			expectType: byAssetSent,
		},
		{
			name:   "transfer by received asset",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addByReceived(); err != nil {
					return err
				}
				return c.addByAsset(common.Address{})
			},
			expectType: byAssetReceived,
		},
		{
			name:   "transfer invalid time range",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				return c.addTimeRange(nil, nil)
			},
			expectErr: true,
		},
		{
			name:   "transfer time range over bySent should err",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addTimeRange(timestamppb.New(t0), timestamppb.New(t0)); err != nil {
					return err
				}
				if err := c.addBySent(); err != nil {
					return err
				}
				return c.valid()
			},
			expectErr: true,
		},
		{
			name:   "transfer time range over byReceived should err",
			filter: newChainFilterTransfer(),
			add: func(c chainFilter) error {
				if err := c.addTimeRange(timestamppb.New(t0), timestamppb.New(t0)); err != nil {
					return err
				}
				if err := c.addByReceived(); err != nil {
					return err
				}
				return c.valid()
			},
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.add(tt.filter)
			if err != nil {
				if !tt.expectErr {
					t.Errorf("unexpected err: %s", err)
				}
				return
			}
			if tt.expectErr {
				t.Error("expected err but got nil")
			}
			if got, want := tt.filter.filterType(), tt.expectType; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
