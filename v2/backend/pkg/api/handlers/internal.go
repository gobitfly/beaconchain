package apihandlers

import (
	"net/http"
)

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

func InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalGetDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	// TODO read params

	// TODO validate params

	// TODO execute query

	// TODO respond
	ReturnCreated(w, r)
}

func InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalPostValidatorDashboardvalidators(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, r)
}

func InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w, r)
}

func InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}

func InternalValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, r)
}
