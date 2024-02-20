package api

import (
	"net/http"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	handlers "github.com/gobitfly/beaconchain/pkg/api/handlers"
	"github.com/gorilla/mux"
)

type endpoint struct {
	Method  string
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
}

func NewApiRouter(dai dataaccess.DataAccessInterface) *mux.Router {
	handlerService := handlers.NewHandlerService(dai)
	router := mux.NewRouter().PathPrefix("/api").Subrouter()
	publicRouter := router.PathPrefix("/v2").Subrouter()
	addPublicRoutes(publicRouter, handlerService)

	internalRouter := router.PathPrefix("/i").Subrouter()
	addInternalRoutes(internalRouter, handlerService)

	return router
}

func addPublicRoutes(router *mux.Router, hs handlers.HandlerService) {
	endpoints := []endpoint{
		{"GET", "/healthz", hs.PublicGetHealthz},
		{"GET", "/healthz-loadbalancer", hs.PublicGetHealthzLoadbalancer},

		{"POST", "/oauth/token", hs.PublicPostOauthToken},

		{"GET", "/users/me/dashboards", hs.PublicGetUserDashboards},

		{"POST", "/account-dashboards", hs.PublicPostAccountDashboards},
		{"GET", "/account-dashboards/{dashboard_id}", hs.PublicGetAccountDashboard},
		{"DELETE", "/account-dashboards/{dashboard_id}", hs.PublicDeleteAccountDashboard},
		{"POST", "/account-dashboards/{dashboard_id}/groups", hs.PublicPostAccountDashboardGroups},
		{"DELETE", "/account-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteAccountDashboardGroups},
		{"POST", "/account-dashboards/{dashboard_id}/accounts", hs.PublicPostAccountDashboardAccounts},
		{"GET", "/account-dashboards/{dashboard_id}/accounts", hs.PublicGetAccountDashboardAccounts},
		{"DELETE", "/account-dashboards/{dashboard_id}/accounts", hs.PublicDeleteAccountDashboardAccounts},
		{"PUT", "/account-dashboards/{dashboard_id}/accounts/{address}", hs.PublicPutAccountDashboardAccount},
		{"POST", "/account-dashboards/{dashboard_id}/public-ids", hs.PublicPostAccountDashboardPublicIds},
		{"PUT", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutAccountDashboardPublicId},
		{"DELETE", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteAccountDashboardPublicId},
		{"GET", "/account-dashboards/{dashboard_id}/transactions", hs.PublicGetAccountDashboardTransactions},
		{"PUT", "/account-dashboards/{dashboard_id}/transactions/settings", hs.PublicPutAccountDashboardTransactionsSettings},

		{"POST", "/validator-dashboards", hs.PublicPostValidatorDashboards},
		{"GET", "/validator-dashboards/{dashboard_id}", hs.PublicGetValidatorDashboard},
		{"DELETE", "/validator-dashboards/{dashboard_id}", hs.PublicDeleteValidatorDashboard},
		{"POST", "/validator-dashboards/{dashboard_id}/groups", hs.PublicPostValidatorDashboardGroups},
		{"DELETE", "/validator-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteValidatorDashboardGroups},
		{"POST", "/validator-dashboards/{dashboard_id}/validators", hs.PublicPostValidatorDashboardValidators},
		{"DELETE", "/validator-dashboards/{dashboard_id}/validators", hs.PublicDeleteValidatorDashboardValidators},
		{"POST", "/validator-dashboards/{dashboard_id}/public-ids", hs.PublicPostValidatorDashboardPublicIds},
		{"PUT", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutValidatorDashboardPublicId},
		{"DELETE", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteValidatorDashboardPublicId},
		{"GET", "/validator-dashboards/{dashboard_id}/slot-viz", hs.PublicGetValidatorDashboardSlotViz},
		{"GET", "/validator-dashboards/{dashboard_id}/summary", hs.PublicGetValidatorDashboardSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/summary", hs.PublicGetValidatorDashboardGroupSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/summary-chart", hs.PublicGetValidatorDashboardSummaryChart},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards", hs.PublicGetValidatorDashboardRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/rewards", hs.PublicGetValidatorDashboardGroupRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards-chart", hs.PublicGetValidatorDashboardRewardsChart},
		{"GET", "/validator-dashboards/{dashboard_id}/duties/{epoch}", hs.PublicGetValidatorDashboardDuties},
		{"GET", "/validator-dashboards/{dashboard_id}/blocks", hs.PublicGetValidatorDashboardBlocks},
		{"GET", "/validator-dashboards/{dashboard_id}/heatmap", hs.PublicGetValidatorDashboardHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/heatmap", hs.PublicGetValidatorDashboardGroupHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/execution-layer-deposits", hs.PublicGetValidatorDashboardExecutionLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/consensus-layer-deposits", hs.PublicGetValidatorDashboardConsensusLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/withdrawals", hs.PublicGetValidatorDashboardWithdrawals},

		{"GET", "/networks/{network}/validators", hs.PublicGetNetworkValidators},
		{"GET", "/networks/{network}/validators/{validator}", hs.PublicGetNetworkValidator},
		{"GET", "/networks/{network}/validators/{validator}/duties", hs.PublicGetNetworkValidatorDuties},
		{"GET", "/networks/{network}/addresses/{address}/validators", hs.PublicGetNetworkAddressValidators},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/validators", hs.PublicGetNetworkWithdrawalCredentialValidators},
		{"GET", "/networks/{network}/validator-statuses", hs.PublicGetNetworkValidatorStatuses},
		{"GET", "/networks/{network}/validator-leaderboard", hs.PublicGetNetworkValidatorLeaderboard},
		{"GET", "/networks/{network}/validator-queue", hs.PublicGetNetworkValidatorQueue},

		{"GET", "/networks/{network}/epochs", hs.PublicGetNetworkEpochs},
		{"GET", "/networks/{network}/epochs/{epoch}", hs.PublicGetNetworkEpoch},

		{"GET", "/networks/{network}/blocks", hs.PublicGetNetworkBlocks},
		{"GET", "/networks/{network}/blocks/{block}", hs.PublicGetNetworkBlock},
		{"GET", "/networks/{network}/slots", hs.PublicGetNetworkSlots},
		{"GET", "/networks/{network}/slots/{slot}", hs.PublicGetNetworkSlot},
		{"GET", "/networks/{network}/validators/{validator}/blocks", hs.PublicGetNetworkValidatorBlocks},
		{"GET", "/networks/{network}/addresses/{address}/priority-fee-blocks", hs.PublicGetNetworkAddressPriorityFeeBlocks},
		{"GET", "/networks/{network}/addresses/{address}/proposer-reward-blocks", hs.PublicGetNetworkAddressProposerRewardBlocks},
		{"GET", "/networks/{network}/forked-blocks", hs.PublicGetNetworkForkedBlocks},
		{"GET", "/networks/{network}/forked-blocks/{block}", hs.PublicGetNetworkForkedBlock},
		{"GET", "/networks/{network}/forked-slots/{slot}", hs.PublicGetNetworkForkedSlot},
		{"GET", "/networks/{network}/block-sizes", hs.PublicGetNetworkBlockSizes},

		{"GET", "/networks/{network}/validators/{validator}/attestations", hs.PublicGetNetworkValidatorAttestations},
		{"GET", "/networks/{network}/epochs/{epoch}/attestations", hs.PublicGetNetworkEpochAttestations},
		{"GET", "/networks/{network}/slots/{slot}/attestations", hs.PublicGetNetworkSlotAttestations},
		{"GET", "/networks/{network}/blocks/{block}/attestations", hs.PublicGetNetworkBlockAttestations},
		{"GET", "/networks/{network}/aggregated-attestations", hs.PublicGetNetworkAggregatedAttestations},

		{"GET", "/networks/{network}/ethstore/{day}", hs.PublicGetNetworkEthStore},
		{"GET", "/networks/{network}/validators/{validator}/reward-history", hs.PublicGetNetworkValidatorRewardHistory},
		{"GET", "/networks/{network}/validators/{validator}/balance-history", hs.PublicGetNetworkValidatorBalanceHistory},
		{"GET", "/networks/{network}/validators/{validator}/performance-history", hs.PublicGetNetworkValidatorPerformanceHistory},

		{"GET", "/networks/{network}/slashings", hs.PublicGetNetworkSlashings},
		{"GET", "/networks/{network}/validators/{validator}/slashings", hs.PublicGetNetworkValidatorSlashings},

		{"GET", "/networks/{network}/deposits", hs.PublicGetNetworkDeposits},
		{"GET", "/networks/{network}/validators/{validator}/deposits", hs.PublicGetNetworkValidatorDeposits},
		{"GET", "/networks/{network}/transactions/{hash}/deposits", hs.PublicGetNetworkTransactionDeposits},

		{"GET", "/networks/{network}/withdrawals", hs.PublicGetNetworkWithdrawals},
		{"GET", "/networks/{network}/slots/{slot}/withdrawals", hs.PublicGetNetworkSlotWithdrawals},
		{"GET", "/networks/{network}/blocks/{block}/withdrawals", hs.PublicGetNetworkBlockWithdrawals},
		{"GET", "/networks/{network}/validators/{validator}/withdrawals", hs.PublicGetNetworkValidatorWithdrawals},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/withdrawals", hs.PublicGetNetworkWithdrawalCredentialWithdrawals},

		{"GET", "/networks/{network}/voluntary-exits", hs.PublicGetNetworkVoluntaryExits},
		{"GET", "/networks/{network}/epochs/{epoch}/voluntary-exits", hs.PublicGetNetworkEpochVoluntaryExits},
		{"GET", "/networks/{network}/slots/{slot}/voluntary-exits", hs.PublicGetNetworkSlotVoluntaryExits},
		{"GET", "/networks/{network}/blocks/{block}/voluntary-exits", hs.PublicGetNetworkBlockVoluntaryExits},

		{"GET", "/networks/{network}/addresses/{address}/balance-history", hs.PublicGetNetworkAddressBalanceHistory},
		{"GET", "/networks/{network}/addresses/{address}/token-supply-history", hs.PublicGetNetworkAddressTokenSupplyHistory},
		{"GET", "/networks/{network}/addresses/{address}/event-logs", hs.PublicGetNetworkAddressEventLogs},

		{"GET", "/networks/{network}/transactions", hs.PublicGetNetworkTransactions},
		{"GET", "/networks/{network}/transactions/{hash}", hs.PublicGetNetworkTransaction},
		{"GET", "/networks/{network}/addresses/{address}/transactions", hs.PublicGetNetworkAddressTransactions},
		{"GET", "/networks/{network}/slots/{slot}/transactions", hs.PublicGetNetworkSlotTransactions},
		{"GET", "/networks/{network}/blocks/{block}/transactions", hs.PublicGetNetworkBlockTransactions},
		{"GET", "/networks/{network}/blocks/{block}/blobs", hs.PublicGetNetworkBlockBlobs},

		{"GET", "/networks/{network}/handlerService-changes", hs.PublicGetNetworkBlsChanges},
		{"GET", "/networks/{network}/epochs/{epoch}/handlerService-changes", hs.PublicGetNetworkEpochBlsChanges},
		{"GET", "/networks/{network}/slots/{slot}/handlerService-changes", hs.PublicGetNetworkSlotBlsChanges},
		{"GET", "/networks/{network}/blocks/{block}/handlerService-changes", hs.PublicGetNetworkBlockBlsChanges},
		{"GET", "/networks/{network}/validators/{validator}/handlerService-changes", hs.PublicGetNetworkValidatorBlsChanges},

		{"GET", "/networks/ethereum/addresses/{address}/ens", hs.PublicGetNetworkAddressEns},
		{"GET", "/networks/ethereum/ens/{ens_name}", hs.PublicGetNetworkEns},

		{"GET", "/networks/{layer_2_network}/batches", hs.PublicGetNetworkBatches},
		{"GET", "/networks/{layer_2_network}/layer1-to-layer2-transactions", hs.PublicGetNetworkLayer1ToLayer2Transactions},
		{"GET", "/networks/{layer_2_network}/layer2-to-layer1-transactions", hs.PublicGetNetworkLayer2ToLayer1Transactions},

		{"POST", "/networks/{network}/broadcasts", hs.PublicPostNetworkBroadcasts},
		{"GET", "/eth-price-history", hs.PublicGetEthPriceHistory},

		{"GET", "/networks/{network}/gasnow", hs.PublicGetNetworkGasNow},
		{"GET", "/networks/{network}/average-gas-limit-history", hs.PublicGetNetworkAverageGasLimitHistory},
		{"GET", "/networks/{network}/gas-used-history", hs.PublicGetNetworkGasUsedHistory},

		{"GET", "/rocket-pool/nodes", hs.PublicGetRocketPoolNodes},
		{"GET", "/rocket-pool/minipools", hs.PublicGetRocketPoolMinipools},

		{"GET", "/networks/{network}/sync-committee/{period}", hs.PublicGetNetworkSyncCommittee},

		{"GET", "/multisig-safes/{address}", hs.PublicGetMultisigSafe},
		{"GET", "/multisig-safes/{address}/transactions", hs.PublicGetMultisigSafeTransactions},
		{"GET", "/multisig-transactions/{hash}/confirmations", hs.PublicGetMultisigTransactionConfirmations},
	}
	addRoutesToRouter(endpoints, router)
}

func addInternalRoutes(router *mux.Router, handlerService handlers.HandlerService) {
	endpoints := []endpoint{

		{"GET", "/users/me/dashboards", handlerService.InternalGetUserDashboards},

		{"POST", "/account-dashboards", handlerService.InternalPostAccountDashboards},
		{"GET", "/account-dashboards/{dashboard_id}", handlerService.InternalGetAccountDashboard},
		{"DELETE", "/account-dashboards/{dashboard_id}", handlerService.InternalDeleteAccountDashboard},
		{"POST", "/account-dashboards/{dashboard_id}/groups", handlerService.InternalPostAccountDashboardGroups},
		{"DELETE", "/account-dashboards/{dashboard_id}/groups/{group_id}", handlerService.InternalDeleteAccountDashboardGroups},
		{"POST", "/account-dashboards/{dashboard_id}/accounts", handlerService.InternalPostAccountDashboardAccounts},
		{"GET", "/account-dashboards/{dashboard_id}/accounts", handlerService.InternalGetAccountDashboardAccounts},
		{"DELETE", "/account-dashboards/{dashboard_id}/accounts", handlerService.InternalDeleteAccountDashboardAccounts},
		{"PUT", "/account-dashboards/{dashboard_id}/accounts/{address}", handlerService.InternalPutAccountDashboardAccount},
		{"POST", "/account-dashboards/{dashboard_id}/public-ids", handlerService.InternalPostAccountDashboardPublicIds},
		{"PUT", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", handlerService.InternalPutAccountDashboardPublicId},
		{"DELETE", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", handlerService.InternalDeleteAccountDashboardPublicId},
		{"GET", "/account-dashboards/{dashboard_id}/transactions", handlerService.InternalGetAccountDashboardTransactions},
		{"PUT", "/account-dashboards/{dashboard_id}/transactions/settings", handlerService.InternalPutAccountDashboardTransactionsSettings},

		{"POST", "/validator-dashboards", handlerService.InternalPostValidatorDashboards},
		{"GET", "/validator-dashboards/{dashboard_id}", handlerService.InternalGetValidatorDashboard},
		{"DELETE", "/validator-dashboards/{dashboard_id}", handlerService.InternalDeleteValidatorDashboard},
		{"POST", "/validator-dashboards/{dashboard_id}/groups", handlerService.InternalPostValidatorDashboardGroups},
		{"DELETE", "/validator-dashboards/{dashboard_id}/groups/{group_id}", handlerService.InternalDeleteValidatorDashboardGroups},
		{"POST", "/validator-dashboards/{dashboard_id}/validators", handlerService.InternalPostValidatorDashboardValidators},
		{"DELETE", "/validator-dashboards/{dashboard_id}/validators", handlerService.InternalDeleteValidatorDashboardValidators},
		{"POST", "/validator-dashboards/{dashboard_id}/public-ids", handlerService.InternalPostValidatorDashboardPublicIds},
		{"PUT", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", handlerService.InternalPutValidatorDashboardPublicId},
		{"DELETE", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", handlerService.InternalDeleteValidatorDashboardPublicId},
		{"GET", "/validator-dashboards/{dashboard_id}/slot-viz", handlerService.InternalGetValidatorDashboardSlotViz},
		{"GET", "/validator-dashboards/{dashboard_id}/summary", handlerService.InternalGetValidatorDashboardSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/summary", handlerService.InternalGetValidatorDashboardGroupSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/summary-chart", handlerService.InternalGetValidatorDashboardSummaryChart},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards", handlerService.InternalGetValidatorDashboardRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/rewards", handlerService.InternalGetValidatorDashboardGroupRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards-chart", handlerService.InternalGetValidatorDashboardRewardsChart},
		{"GET", "/validator-dashboards/{dashboard_id}/duties/{epoch}", handlerService.InternalGetValidatorDashboardDuties},
		{"GET", "/validator-dashboards/{dashboard_id}/blocks", handlerService.InternalGetValidatorDashboardBlocks},
		{"GET", "/validator-dashboards/{dashboard_id}/heatmap", handlerService.InternalGetValidatorDashboardHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/heatmap", handlerService.InternalGetValidatorDashboardGroupHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/execution-layer-deposits", handlerService.InternalGetValidatorDashboardExecutionLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/consensus-layer-deposits", handlerService.InternalGetValidatorDashboardConsensusLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/withdrawals", handlerService.InternalGetValidatorDashboardWithdrawals},
	}
	addRoutesToRouter(endpoints, router)
}

func addRoutesToRouter(endpoints []endpoint, router *mux.Router) {
	for _, endpoint := range endpoints {
		router.HandleFunc(endpoint.Path, endpoint.Handler).Methods(endpoint.Method)
	}
}
