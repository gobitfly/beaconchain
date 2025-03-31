package modules

import "github.com/gobitfly/beaconchain/pkg/commons/utils"

func GetBalanceChurnLimit(totalActiveBalance uint64) uint64 {
	balanceChurnLimit := totalActiveBalance / utils.Config.ClConfig.ChurnLimitQuotient
	if balanceChurnLimit < utils.Config.ClConfig.MinPerEpochChurnLimitElectra {
		balanceChurnLimit = utils.Config.ClConfig.MinPerEpochChurnLimitElectra
	}
	return balanceChurnLimit - (balanceChurnLimit % utils.Config.ClConfig.EffectiveBalanceIncrement)
}

func GetActivationExitChurnLimit(totalActiveBalance uint64) uint64 {
	balanceChurnLimit := GetBalanceChurnLimit(totalActiveBalance)
	if balanceChurnLimit > utils.Config.ClConfig.MaxPerEpochActivationExitChurnLimit {
		return utils.Config.ClConfig.MaxPerEpochActivationExitChurnLimit
	}
	return balanceChurnLimit
}
