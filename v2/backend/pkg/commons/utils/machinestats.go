package utils

func GetMachineStatsGap(resultCount uint64) int {
	if resultCount > 20160 { // more than 14 (31)
		return 8
	}
	if resultCount > 10080 { // more than 7 (14)
		return 7
	}
	if resultCount > 2880 { // more than 2 (7)
		return 5
	}
	if resultCount > 1440 { // more than 1 (2)
		return 4
	}
	if resultCount > 770 { // more than 12h
		return 2
	}
	return 1
}
