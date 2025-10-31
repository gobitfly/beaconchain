package io

import (
	"time"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
)

// EpochToStartSlot converts an epoch number to its starting slot number.
func EpochToStartSlot(epoch int, config config.ChainConfig) int {
	return epoch * config.SlotsPerEpoch
}

// SlotToStartTime converts a slot number to its corresponding timestamp.
func SlotToStartTime(slot int, config config.ChainConfig) time.Time {
	return time.Unix(int64(config.GenesisTimestamp+slot*config.SecondsPerSlot), 0).UTC()
}

// EpochToTimeRange converts an epoch number to starting timestamp.
func EpochToStartTime(epoch int, config config.ChainConfig) time.Time {
	return SlotToStartTime(EpochToStartSlot(epoch, config), config)
}

// EpochToResultRange converts an epoch number to its corresponding ResultRange.
func EpochToResultRange(epoch int, config config.ChainConfig) model.ResultRange {
	startSlot := EpochToStartSlot(epoch, config)
	endSlot := EpochToStartSlot(epoch+1, config) - 1

	startTime := EpochToStartTime(epoch, config)
	endTime := EpochToStartTime(epoch+1, config).Add(-time.Second)

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
			Start: model.Timestamp(startTime.Unix()),
			End:   model.Timestamp(endTime.Unix()),
		},
	}
}
