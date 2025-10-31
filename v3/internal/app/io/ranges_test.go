package io

import (
	"testing"
	"time"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/stretchr/testify/assert"
)

func TestEpochToSlot(t *testing.T) {
	tests := []struct {
		name         string
		epoch        int
		config       config.ChainConfig
		expectedSlot int
	}{
		{
			name:  "epoch 0 should return slot 0",
			epoch: 0,
			config: config.ChainConfig{
				SlotsPerEpoch: 32,
			},
			expectedSlot: 0,
		},
		{
			name:  "epoch 1 with 32 slots per epoch",
			epoch: 1,
			config: config.ChainConfig{
				SlotsPerEpoch: 32,
			},
			expectedSlot: 32,
		},
		{
			name:  "epoch 10 with 32 slots per epoch",
			epoch: 10,
			config: config.ChainConfig{
				SlotsPerEpoch: 32,
			},
			expectedSlot: 320,
		},
		{
			name:  "epoch 5 with different slots per epoch (64)",
			epoch: 5,
			config: config.ChainConfig{
				SlotsPerEpoch: 64,
			},
			expectedSlot: 320,
		},
		{
			name:  "large epoch number",
			epoch: 1000,
			config: config.ChainConfig{
				SlotsPerEpoch: 32,
			},
			expectedSlot: 32000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EpochToStartSlot(tt.epoch, tt.config)
			assert.Equal(t, tt.expectedSlot, result)
		})
	}
}

func TestEpochToTimestamp(t *testing.T) {
	tests := []struct {
		name              string
		epoch             int
		config            config.ChainConfig
		expectedTimestamp time.Time
	}{
		{
			name:  "epoch 0 should return genesis timestamp",
			epoch: 0,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedTimestamp: time.Unix(1606824000, 0).UTC(),
		},
		{
			name:  "epoch 1 calculation",
			epoch: 1,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedTimestamp: time.Unix(1606824000+(32*12), 0).UTC(),
		},
		{
			name:  "epoch 10 with standard config",
			epoch: 10,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedTimestamp: time.Unix(1606824000+(320*12), 0).UTC(),
		},
		{
			name:  "epoch 5 with different seconds per slot",
			epoch: 5,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   6,
				GenesisTimestamp: 1606824000,
			},
			expectedTimestamp: time.Unix(1606824000+(160*6), 0).UTC(),
		},
		{
			name:  "zero genesis timestamp",
			epoch: 3,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 0,
			},
			expectedTimestamp: time.Unix(0+(96*12), 0).UTC(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EpochToStartTime(tt.epoch, tt.config)
			assert.Equal(t, tt.expectedTimestamp, result)
		})
	}
}

func TestEpochToResultRange(t *testing.T) {
	tests := []struct {
		name          string
		epoch         int
		config        config.ChainConfig
		expectedRange model.ResultRange
	}{
		{
			name:  "epoch 0 result range",
			epoch: 0,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedRange: model.ResultRange{
				Epoch: model.EpochRange{
					Start: 0,
					End:   0,
				},
				Slot: model.SlotRange{
					Start: 0,
					End:   31,
				},
				Timestamp: model.TimeRange{
					Start: 1606824000,
					End:   1606824000 + (32 * 12) - 1, // 1606824383
				},
			},
		},
		{
			name:  "epoch 1 result range",
			epoch: 1,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedRange: model.ResultRange{
				Epoch: model.EpochRange{
					Start: 1,
					End:   1,
				},
				Slot: model.SlotRange{
					Start: 32,
					End:   63,
				},
				Timestamp: model.TimeRange{
					Start: 1606824000 + (32 * 12),     // 1606824384
					End:   1606824000 + (64 * 12) - 1, // 1606824767
				},
			},
		},
		{
			name:  "epoch 5 with different config",
			epoch: 5,
			config: config.ChainConfig{
				SlotsPerEpoch:    64,
				SecondsPerSlot:   6,
				GenesisTimestamp: 1000000,
			},
			expectedRange: model.ResultRange{
				Epoch: model.EpochRange{
					Start: 5,
					End:   5,
				},
				Slot: model.SlotRange{
					Start: 320,
					End:   383,
				},
				Timestamp: model.TimeRange{
					Start: 1000000 + (320 * 6),
					End:   1000000 + (384 * 6) - 1,
				},
			},
		},
		{
			name:  "large epoch number",
			epoch: 100,
			config: config.ChainConfig{
				SlotsPerEpoch:    32,
				SecondsPerSlot:   12,
				GenesisTimestamp: 1606824000,
			},
			expectedRange: model.ResultRange{
				Epoch: model.EpochRange{
					Start: 100,
					End:   100,
				},
				Slot: model.SlotRange{
					Start: 3200,
					End:   3231,
				},
				Timestamp: model.TimeRange{
					Start: 1606824000 + (3200 * 12),
					End:   1606824000 + (3232 * 12) - 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EpochToResultRange(tt.epoch, tt.config)
			assert.Equal(t, tt.expectedRange, result)
		})
	}
}
