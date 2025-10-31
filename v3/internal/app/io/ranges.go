package io

import (
	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
)

// EpochToStartSlot converts an epoch number to its starting slot number.
func EpochToStartSlot(epoch int, config config.ChainConfig) int {
	return epoch * config.SlotsPerEpoch
}

// SlotToStartTimestamp converts a slot number to its corresponding timestamp.
func SlotToStartTimestamp(slot int, config config.ChainConfig) int {
	return config.GenesisTimestamp + slot*config.SecondsPerSlot
}

// EpochToTimeRange converts an epoch number to starting timestamp.
func EpochToStartTimestamp(epoch int, config config.ChainConfig) int {
	return SlotToStartTimestamp(EpochToStartSlot(epoch, config), config)
}

// EpochToResultRange converts an epoch number to its corresponding ResultRange.
func EpochToResultRange(epoch int, config config.ChainConfig) model.ResultRange {
	startSlot := EpochToStartSlot(epoch, config)
	endSlot := EpochToStartSlot(epoch+1, config) - 1

	startTime := EpochToStartTimestamp(epoch, config)
	endTime := EpochToStartTimestamp(epoch+1, config) - 1

	return model.ResultRange{
		Epoch: model.EpochRange{
			Start: epoch,
			End:   epoch,
		},
		Slot: model.SlotRange{
			Start: startSlot,
			End:   endSlot,
		},
		Timestamp: model.TimeRange{
			Start: startTime,
			End:   endTime,
		},
	}
}
