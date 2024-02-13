package utils

import "time"

const (
	Day      = 24 * time.Hour
	Week     = 7 * Day
	Month    = 30 * Day
	Year     = 12 * Month
	LongTime = 37 * Year
)

// SlotToTime returns a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.ClConfig.SecondsPerSlot), 0)
}

// TimeToSlot returns time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if Config.Chain.GenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.ClConfig.SecondsPerSlot
}

func TimeToFirstSlotOfEpoch(timestamp uint64) uint64 {
	slot := TimeToSlot(timestamp)
	lastEpochOffset := slot % Config.Chain.ClConfig.SlotsPerEpoch
	slot = slot - lastEpochOffset
	return slot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.ClConfig.SecondsPerSlot*Config.Chain.ClConfig.SlotsPerEpoch), 0)
}

// TimeToDay will return a days since genesis for an timestamp
func TimeToDay(timestamp uint64) uint64 {
	const hoursInADay = float64(Day / time.Hour)
	return uint64(time.Unix(int64(timestamp), 0).Sub(time.Unix(int64(Config.Chain.GenesisTimestamp), 0)).Hours() / hoursInADay)
}

func DayToTime(day int64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp), 0).Add(Day * time.Duration(day))
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(Config.Chain.GenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.ClConfig.SecondsPerSlot) / int64(Config.Chain.ClConfig.SlotsPerEpoch)
}

func EpochsPerDay() uint64 {
	return (uint64(Day.Seconds()) / Config.Chain.ClConfig.SlotsPerEpoch) / Config.Chain.ClConfig.SecondsPerSlot
}

func GetFirstAndLastEpochForDay(day uint64) (firstEpoch uint64, lastEpoch uint64) {
	firstEpoch = day * EpochsPerDay()
	lastEpoch = firstEpoch + EpochsPerDay() - 1
	return firstEpoch, lastEpoch
}

func GetLastBalanceInfoSlotForDay(day uint64) uint64 {
	return ((day+1)*EpochsPerDay() - 1) * Config.Chain.ClConfig.SlotsPerEpoch
}
