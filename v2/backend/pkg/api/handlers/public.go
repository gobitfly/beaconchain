package apihandlers

import "net/http"

// All handler function names must include the HTTP method and the path they handle
// Public handlers may only be authenticated by an API key
// Public handlers must never call internal handlers

func PublicGetHealthz(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetHealthzLoadbalancer(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostOauthToken(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func PublicGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidators(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidator(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorDuties(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressValidators(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkWithdrawalCredentialValidators(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorStatuses(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorLeaderboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorQueue(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEpochs(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEpoch(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlock(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlots(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlot(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressPriorityFeeBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressProposerRewardBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkForkedBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkForkedBlock(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkForkedSlot(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockSizes(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorAttestations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEpochAttestations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlotAttestations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockAttestations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAggregatedAttestations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEthStore(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorRewardHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorPerformanceHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlashings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorSlashings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkTransactionDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}
func PublicGetNetworkBlockWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkWithdrawalCredentialWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEpochVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressBalanceHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressTokenSupplyHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressEventLogs(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkTransaction(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlotTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockBlobs(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlsChanges(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEpochBlsChanges(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSlotBlsChanges(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBlockBlsChanges(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkValidatorBlsChanges(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAddressEns(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkEns(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkBatches(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkLayer2ToLayer1Transactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkLayer1ToLayer2Transactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicPostNetworkBroadcasts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func PublicGetEthPriceHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkGasNow(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkAverageGasLimitHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkGasUsedHistory(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetRocketPoolNodes(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetNetworkSyncCommittee(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetMultisigSafe(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetMultisigSafeTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func PublicGetMultisigTransactionConfirmations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}
