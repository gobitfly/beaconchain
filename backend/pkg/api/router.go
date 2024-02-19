package api

import (
	"net/http"

	apihandlers "github.com/gobitfly/beaconchain/pkg/api/handlers"
	"github.com/gorilla/mux"
)

type Endpoint struct {
	Method  string
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
}

func GetApiRouter() *mux.Router {
	router := mux.NewRouter().PathPrefix("/api").Subrouter()
	publicRouter := router.PathPrefix("/v2").Subrouter()
	addPublicRoutes(publicRouter)

	// TODO add internal routes

	return router
}

func addPublicRoutes(router *mux.Router) {
	endpoints := []Endpoint{
		{"GET", "/healthz", apihandlers.PublicGetHealthz},
		{"GET", "/healthz-loadbalancer", apihandlers.PublicGetHealthzLoadbalancer},

		{"POST", "/oauth/token", apihandlers.PublicPostOauthToken},

		{"GET", "/users/me/dashboards", apihandlers.PublicGetUserDashboards},

		{"POST", "/account-dashboards", apihandlers.PublicPostAccountDashboards},
		{"GET", "/account-dashboards/{dashboard_id}", apihandlers.PublicGetAccountDashboard},
		{"DELETE", "/account-dashboards/{dashboard_id}", apihandlers.PublicDeleteAccountDashboard},
		{"POST", "/account-dashboard/{dashboard_id}/groups", apihandlers.PublicPostAccountDashboardGroups},
		{"DELETE", "/account-dashboard/{dashboard_id}/groups/{group_id}", apihandlers.PublicDeleteAccountDashboardGroups},
		{"POST", "/account-dashboard/{dashboard_id}/accounts", apihandlers.PublicPostAccountDashboardAccounts},
		{"GET", "/account-dashboard/{dashboard_id}/accounts", apihandlers.PublicGetAccountDashboardAccounts},
		{"DELETE", "/account-dashboard/{dashboard_id}/accounts", apihandlers.PublicDeleteAccountDashboardAccounts},
		{"PUT", "/account-dashboard/{dashboard_id}/accounts/{address}", apihandlers.PublicPutAccountDashboardAccount},
		{"POST", "/account-dashboard/{dashboard_id}/public-ids", apihandlers.PublicPostAccountDashboardPublicIds},
		{"PUT", "/account-dashboard/{dashboard_id}/public-ids/{public_id}", apihandlers.PublicPutAccountDashboardPublicId},
		{"DELETE", "/account-dashboard/{dashboard_id}/public-ids/{public_id}", apihandlers.PublicDeleteAccountDashboardPublicId},
		{"GET", "/account-dashboard/{dashboard_id}/transactions", apihandlers.PublicGetAccountDashboardTransactions},
		{"PUT", "/account-dashboard/{dashboard_id}/transactions/settings", apihandlers.PublicPutAccountDashboardTransactionsSettings},

		{"POST", "/validator-dashboards", apihandlers.PublicPostValidatorDashboards},
		{"GET", "/validator-dashboards/{dashboard_id}", apihandlers.PublicGetValidatorDashboard},
		{"DELETE", "/validator-dashboards/{dashboard_id}", apihandlers.PublicDeleteValidatorDashboard},
		{"POST", "/validator-dashboard/{dashboard_id}/groups", apihandlers.PublicPostValidatorDashboardGroups},
		{"DELETE", "/validator-dashboard/{dashboard_id}/groups/{group_id}", apihandlers.PublicDeleteValidatorDashboardGroups},
		{"POST", "/validator-dashboard/{dashboard_id}/validators", apihandlers.PublicPostValidatorDashboardValidators},
		{"DELETE", "/validator-dashboard/{dashboard_id}/validators", apihandlers.PublicDeleteValidatorDashboardValidators},
		{"POST", "/validator-dashboard/{dashboard_id}/public-ids", apihandlers.PublicPostValidatorDashboardPublicIds},
		{"PUT", "/validator-dashboard/{dashboard_id}/public-ids/{public_id}", apihandlers.PublicPutValidatorDashboardPublicId},
		{"DELETE", "/validator-dashboard/{dashboard_id}/public-ids/{public_id}", apihandlers.PublicDeleteValidatorDashboardPublicId},
		{"GET", "/validator-dashboard/{dashboard_id}/slot-viz", apihandlers.PublicGetValidatorDashboardSlotViz},
		{"GET", "/validator-dashboard/{dashboard_id}/summary", apihandlers.PublicGetValidatorDashboardSummary},
		{"GET", "/validator-dashboard/{dashboard_id}/groups/{group_id}/summary", apihandlers.PublicGetValidatorDashboardGroupSummary},
		{"GET", "/validator-dashboard/{dashboard_id}/summary-chart", apihandlers.PublicGetValidatorDashboardSummaryChart},
		{"GET", "/validator-dashboard/{dashboard_id}/rewards", apihandlers.PublicGetValidatorDashboardRewards},
		{"GET", "/validator-dashboard/{dashboard_id}/groups/{group_id}/rewards", apihandlers.PublicGetValidatorDashboardGroupRewards},
		{"GET", "/validator-dashboard/{dashboard_id}/rewards-chart", apihandlers.PublicGetValidatorDashboardRewardsChart},
		{"GET", "/validator-dashboard/{dashboard_id}/duties/{epoch}", apihandlers.PublicGetValidatorDashboardDuties},
		{"GET", "/validator-dashboard/{dashboard_id}/blocks", apihandlers.PublicGetValidatorDashboardBlocks},
		{"GET", "/validator-dashboard/{dashboard_id}/heatmap", apihandlers.PublicGetValidatorDashboardHeatmap},
		{"GET", "/validator-dashboard/{dashboard_id}/groups/{group_id}/heatmap", apihandlers.PublicGetValidatorDashboardGroupHeatmap},
		{"GET", "/validator-dashboard/{dashboard_id}/execution-layer-deposits", apihandlers.PublicGetValidatorDashboardExecutionLayerDeposits},
		{"GET", "/validator-dashboard/{dashboard_id}/consensus-layer-deposits", apihandlers.PublicGetValidatorDashboardConsensusLayerDeposits},
		{"GET", "/validator-dashboard/{dashboard_id}/withdrawals", apihandlers.PublicGetValidatorDashboardWithdrawals},

		{"GET", "/networks/{network}/validators", apihandlers.PublicGetNetworkValidators},
		{"GET", "/networks/{network}/validators/{validator}", apihandlers.PublicGetNetworkValidator},
		{"GET", "/networks/{network}/validators/{validator}/duties", apihandlers.PublicGetNetworkValidatorDuties},
		{"GET", "/networks/{network}/addresses/{address}/validators", apihandlers.PublicGetNetworkAddressValidators},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/validators", apihandlers.PublicGetNetworkWithdrawalCredentialValidators},
		{"GET", "/networks/{network}/validator-statuses", apihandlers.PublicGetNetworkValidatorStatuses},
		{"GET", "/networks/{network}/validator-leaderboard", apihandlers.PublicGetNetworkValidatorLeaderboard},
		{"GET", "/networks/{network}/validator-queue", apihandlers.PublicGetNetworkValidatorQueue},

		{"GET", "/networks/{network}/epochs", apihandlers.PublicGetNetworkEpochs},
		{"GET", "/networks/{network}/epochs/{epoch}", apihandlers.PublicGetNetworkEpoch},

		{"GET", "/networks/{network}/blocks", apihandlers.PublicGetNetworkBlocks},
		{"GET", "/networks/{network}/blocks/{block}", apihandlers.PublicGetNetworkBlock},
		{"GET", "/networks/{network}/slots", apihandlers.PublicGetNetworkSlots},
		{"GET", "/networks/{network}/slots/{slot}", apihandlers.PublicGetNetworkSlot},
		{"GET", "/networks/{network}/validators/{validator}/blocks", apihandlers.PublicGetNetworkValidatorBlocks},
		{"GET", "/networks/{network}/addresses/{address}/priority-fee-blocks", apihandlers.PublicGetNetworkAddressPriorityFeeBlocks},
		{"GET", "/networks/{network}/addresses/{address}/proposer-reward-blocks", apihandlers.PublicGetNetworkAddressProposerRewardBlocks},
		{"GET", "/networks/{network}/forked-blocks", apihandlers.PublicGetNetworkForkedBlocks},
		{"GET", "/networks/{network}/forked-blocks/{block}", apihandlers.PublicGetNetworkForkedBlock},
		{"GET", "/networks/{network}/forked-slots/{slot}", apihandlers.PublicGetNetworkForkedSlot},
		{"GET", "/networks/{network}/block-sizes", apihandlers.PublicGetNetworkBlockSizes},

		{"GET", "/networks/{network}/validators/{validator}/attestations", apihandlers.PublicGetNetworkValidatorAttestations},
		{"GET", "/networks/{network}/epochs/{epoch}/attestations", apihandlers.PublicGetNetworkEpochAttestations},
		{"GET", "/networks/{network}/slots/{slot}/attestations", apihandlers.PublicGetNetworkSlotAttestations},
		{"GET", "/networks/{network}/blocks/{block}/attestations", apihandlers.PublicGetNetworkBlockAttestations},
		{"GET", "/networks/{network}/aggregated-attestations", apihandlers.PublicGetNetworkAggregatedAttestations},

		{"GET", "/networks/{network}/ethstore/{day}", apihandlers.PublicGetNetworkEthStore},
		{"GET", "/networks/{network}/validators/{validator}/reward-history", apihandlers.PublicGetNetworkValidatorRewardHistory},
		{"GET", "/networks/{network}/validators/{validator}/balance-history", apihandlers.PublicGetNetworkValidatorBalanceHistory},
		{"GET", "/networks/{network}/validators/{validator}/performance-history", apihandlers.PublicGetNetworkValidatorPerformanceHistory},

		{"GET", "/networks/{network}/slashings", apihandlers.PublicGetNetworkSlashings},
		{"GET", "/networks/{network}/validators/{validator}/slashings", apihandlers.PublicGetNetworkValidatorSlashings},

		{"GET", "/networks/{network}/deposits", apihandlers.PublicGetNetworkDeposits},
		{"GET", "/networks/{network}/validators/{validator}/deposits", apihandlers.PublicGetNetworkValidatorDeposits},
		{"GET", "/networks/{network}/transactions/{hash}/deposits", apihandlers.PublicGetNetworkTransactionDeposits},

		{"GET", "/networks/{network}/withdrawals", apihandlers.PublicGetNetworkWithdrawals},
		{"GET", "/networks/{network}/slots/{slot}/withdrawals", apihandlers.PublicGetNetworkSlotWithdrawals},
		{"GET", "/networks/{network}/blocks/{block}/withdrawals", apihandlers.PublicGetNetworkBlockWithdrawals},
		{"GET", "/networks/{network}/validators/{validator}/withdrawals", apihandlers.PublicGetNetworkValidatorWithdrawals},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/withdrawals", apihandlers.PublicGetNetworkWithdrawalCredentialWithdrawals},

		{"GET", "/networks/{network}/voluntary-exits", apihandlers.PublicGetNetworkVoluntaryExits},
		{"GET", "/networks/{network}/epochs/{epoch}/voluntary-exits", apihandlers.PublicGetNetworkEpochVoluntaryExits},
		{"GET", "/networks/{network}/slots/{slot}/voluntary-exits", apihandlers.PublicGetNetworkSlotVoluntaryExits},
		{"GET", "/networks/{network}/blocks/{block}/voluntary-exits", apihandlers.PublicGetNetworkBlockVoluntaryExits},

		{"GET", "/networks/{network}/addresses/{address}/balance-history", apihandlers.PublicGetNetworkAddressBalanceHistory},
		{"GET", "/networks/{network}/addresses/{address}/token-supply-history", apihandlers.PublicGetNetworkAddressTokenSupplyHistory},
		{"GET", "/networks/{network}/addresses/{address}/event-logs", apihandlers.PublicGetNetworkAddressEventLogs},

		{"GET", "/networks/{network}/transactions", apihandlers.PublicGetNetworkTransactions},
		{"GET", "/networks/{network}/transactions/{hash}", apihandlers.PublicGetNetworkTransaction},
		{"GET", "/networks/{network}/addresses/{address}/transactions", apihandlers.PublicGetNetworkAddressTransactions},
		{"GET", "/networks/{network}/slots/{slot}/transactions", apihandlers.PublicGetNetworkSlotTransactions},
		{"GET", "/networks/{network}/blocks/{block}/transactions", apihandlers.PublicGetNetworkBlockTransactions},
		{"GET", "/networks/{network}/blocks/{block}/blobs", apihandlers.PublicGetNetworkBlockBlobs},

		{"GET", "/networks/{network}/bls-changes", apihandlers.PublicGetNetworkBlsChanges},
		{"GET", "/networks/{network}/epochs/{epoch}/bls-changes", apihandlers.PublicGetNetworkEpochBlsChanges},
		{"GET", "/networks/{network}/slots/{slot}/bls-changes", apihandlers.PublicGetNetworkSlotBlsChanges},
		{"GET", "/networks/{network}/blocks/{block}/bls-changes", apihandlers.PublicGetNetworkBlockBlsChanges},
		{"GET", "/networks/{network}/validators/{validator}/bls-changes", apihandlers.PublicGetNetworkValidatorBlsChanges},

		{"GET", "/networks/ethereum/addresses/{address}/ens", apihandlers.PublicGetNetworkAddressEns},
		{"GET", "/networks/ethereum/ens/{ens_name}", apihandlers.PublicGetNetworkEns},

		{"GET", "/networks/{layer_2_network}/batches", apihandlers.PublicGetNetworkBatches},
		{"GET", "/networks/{layer_2_network}/layer1-to-layer2-transactions", apihandlers.PublicGetNetworkLayer1ToLayer2Transactions},
		{"GET", "/networks/{layer_2_network}/layer2-to-layer1-transactions", apihandlers.PublicGetNetworkLayer2ToLayer1Transactions},

		{"POST", "/networks/{network}/broadcasts", apihandlers.PublicPostNetworkBroadcasts},
		{"GET", "/eth-price-history", apihandlers.PublicGetEthPriceHistory},

		{"GET", "/networks/{network}/gasnow", apihandlers.PublicGetNetworkGasNow},
		{"GET", "/networks/{network}/average-gas-limit-history", apihandlers.PublicGetNetworkAverageGasLimitHistory},
		{"GET", "/networks/{network}/gas-used-history", apihandlers.PublicGetNetworkGasUsedHistory},

		{"GET", "/rocket-pool/nodes", apihandlers.PublicGetRocketPoolNodes},
		{"GET", "/rocket-pool/minipools", apihandlers.PublicGetRocketPoolMinipools},

		{"GET", "/networks/{network}/sync-committee/{period}", apihandlers.PublicGetNetworkSyncCommittee},

		{"GET", "/multisig-safes/{address}", apihandlers.PublicGetMultisigSafe},
		{"GET", "/multisig-safes/{address}/transactions", apihandlers.PublicGetMultisigSafeTransactions},
		{"GET", "/multisig-transactions/{hash}/confirmations", apihandlers.PublicGetMultisigTransactionConfirmations},
	}
	for _, endpoint := range endpoints {
		router.HandleFunc(endpoint.Path, endpoint.Handler).Methods(endpoint.Method)
	}
}
