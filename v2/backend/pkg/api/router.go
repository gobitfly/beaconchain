package api

import (
	"net/http"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	handlers "github.com/gobitfly/beaconchain/pkg/api/handlers"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gorilla/mux"
)

type endpoint struct {
	Method         string
	Path           string
	PublicHandler  func(w http.ResponseWriter, r *http.Request)
	InternalHander func(w http.ResponseWriter, r *http.Request)
}

const (
	get     = http.MethodGet
	post    = http.MethodPost
	put     = http.MethodPut
	delete  = http.MethodDelete
	options = http.MethodOptions
)

func NewApiRouter(dataAccessor dataaccess.DataAccessor, cfg *types.Config) *mux.Router {

	router := mux.NewRouter().PathPrefix("/api").Subrouter()
	publicRouter := router.PathPrefix("/v2").Subrouter()
	internalRouter := router.PathPrefix("/i").Subrouter()

	sessionManager := newSessionManager(cfg.RedisCacheEndpoint)
	internalRouter.Use(sessionManager.LoadAndSave)
	handlerService := handlers.NewHandlerService(dataAccessor, sessionManager)

	addRoutes(handlerService, publicRouter, internalRouter)

	return router
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == options {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func addRoutes(hs *handlers.HandlerService, publicRouter, internalRouter *mux.Router) {
	addValidatorDashboardRoutes(hs, publicRouter, internalRouter)
	endpoints := []endpoint{
		{get, "/healthz", hs.PublicGetHealthz, nil},
		{get, "/healthz-loadbalancer", hs.PublicGetHealthzLoadbalancer, nil},

		{post, "/oauth/token", hs.PublicPostOauthToken, nil},

		{get, "/users/me/dashboards", hs.PublicGetUserDashboards, hs.InternalGetUserDashboards},

		{post, "/account-dashboards", hs.PublicPostAccountDashboards, hs.InternalPostAccountDashboards},
		{get, "/account-dashboards/{dashboard_id}", hs.PublicGetAccountDashboard, hs.InternalGetAccountDashboard},
		{delete, "/account-dashboards/{dashboard_id}", hs.PublicDeleteAccountDashboard, hs.InternalDeleteAccountDashboard},
		{post, "/account-dashboards/{dashboard_id}/groups", hs.PublicPostAccountDashboardGroups, hs.InternalPostAccountDashboardGroups},
		{delete, "/account-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteAccountDashboardGroups, hs.InternalDeleteAccountDashboardGroups},
		{post, "/account-dashboards/{dashboard_id}/accounts", hs.PublicPostAccountDashboardAccounts, hs.InternalPostAccountDashboardAccounts},
		{get, "/account-dashboards/{dashboard_id}/accounts", hs.PublicGetAccountDashboardAccounts, hs.InternalGetAccountDashboardAccounts},
		{delete, "/account-dashboards/{dashboard_id}/accounts", hs.PublicDeleteAccountDashboardAccounts, hs.InternalDeleteAccountDashboardAccounts},
		{put, "/account-dashboards/{dashboard_id}/accounts/{address}", hs.PublicPutAccountDashboardAccount, hs.InternalPutAccountDashboardAccount},
		{post, "/account-dashboards/{dashboard_id}/public-ids", hs.PublicPostAccountDashboardPublicIds, hs.InternalPostAccountDashboardPublicIds},
		{put, "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutAccountDashboardPublicId, hs.InternalPutAccountDashboardPublicId},
		{delete, "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteAccountDashboardPublicId, hs.InternalDeleteAccountDashboardPublicId},
		{get, "/account-dashboards/{dashboard_id}/transactions", hs.PublicGetAccountDashboardTransactions, hs.InternalGetAccountDashboardTransactions},
		{put, "/account-dashboards/{dashboard_id}/transactions/settings", hs.PublicPutAccountDashboardTransactionsSettings, hs.InternalPutAccountDashboardTransactionsSettings},

		{get, "/networks/{network}/validators", hs.PublicGetNetworkValidators, nil},
		{get, "/networks/{network}/validators/{validator}", hs.PublicGetNetworkValidator, nil},
		{get, "/networks/{network}/validators/{validator}/duties", hs.PublicGetNetworkValidatorDuties, nil},
		{get, "/networks/{network}/addresses/{address}/validators", hs.PublicGetNetworkAddressValidators, nil},
		{get, "/networks/{network}/withdrawal-credentials/{credential}/validators", hs.PublicGetNetworkWithdrawalCredentialValidators, nil},
		{get, "/networks/{network}/validator-statuses", hs.PublicGetNetworkValidatorStatuses, nil},
		{get, "/networks/{network}/validator-leaderboard", hs.PublicGetNetworkValidatorLeaderboard, nil},
		{get, "/networks/{network}/validator-queue", hs.PublicGetNetworkValidatorQueue, nil},

		{get, "/networks/{network}/epochs", hs.PublicGetNetworkEpochs, nil},
		{get, "/networks/{network}/epochs/{epoch}", hs.PublicGetNetworkEpoch, nil},

		{get, "/networks/{network}/blocks", hs.PublicGetNetworkBlocks, nil},
		{get, "/networks/{network}/blocks/{block}", hs.PublicGetNetworkBlock, nil},
		{get, "/networks/{network}/slots", hs.PublicGetNetworkSlots, nil},
		{get, "/networks/{network}/slots/{slot}", hs.PublicGetNetworkSlot, nil},
		{get, "/networks/{network}/validators/{validator}/blocks", hs.PublicGetNetworkValidatorBlocks, nil},
		{get, "/networks/{network}/addresses/{address}/priority-fee-blocks", hs.PublicGetNetworkAddressPriorityFeeBlocks, nil},
		{get, "/networks/{network}/addresses/{address}/proposer-reward-blocks", hs.PublicGetNetworkAddressProposerRewardBlocks, nil},
		{get, "/networks/{network}/forked-blocks", hs.PublicGetNetworkForkedBlocks, nil},
		{get, "/networks/{network}/forked-blocks/{block}", hs.PublicGetNetworkForkedBlock, nil},
		{get, "/networks/{network}/forked-slots/{slot}", hs.PublicGetNetworkForkedSlot, nil},
		{get, "/networks/{network}/block-sizes", hs.PublicGetNetworkBlockSizes, nil},

		{get, "/networks/{network}/validators/{validator}/attestations", hs.PublicGetNetworkValidatorAttestations, nil},
		{get, "/networks/{network}/epochs/{epoch}/attestations", hs.PublicGetNetworkEpochAttestations, nil},
		{get, "/networks/{network}/slots/{slot}/attestations", hs.PublicGetNetworkSlotAttestations, nil},
		{get, "/networks/{network}/blocks/{block}/attestations", hs.PublicGetNetworkBlockAttestations, nil},
		{get, "/networks/{network}/aggregated-attestations", hs.PublicGetNetworkAggregatedAttestations, nil},

		{get, "/networks/{network}/ethstore/{day}", hs.PublicGetNetworkEthStore, nil},
		{get, "/networks/{network}/validators/{validator}/reward-history", hs.PublicGetNetworkValidatorRewardHistory, nil},
		{get, "/networks/{network}/validators/{validator}/balance-history", hs.PublicGetNetworkValidatorBalanceHistory, nil},
		{get, "/networks/{network}/validators/{validator}/performance-history", hs.PublicGetNetworkValidatorPerformanceHistory, nil},

		{get, "/networks/{network}/slashings", hs.PublicGetNetworkSlashings, nil},
		{get, "/networks/{network}/validators/{validator}/slashings", hs.PublicGetNetworkValidatorSlashings, nil},

		{get, "/networks/{network}/deposits", hs.PublicGetNetworkDeposits, nil},
		{get, "/networks/{network}/validators/{validator}/deposits", hs.PublicGetNetworkValidatorDeposits, nil},
		{get, "/networks/{network}/transactions/{hash}/deposits", hs.PublicGetNetworkTransactionDeposits, nil},

		{get, "/networks/{network}/withdrawals", hs.PublicGetNetworkWithdrawals, nil},
		{get, "/networks/{network}/slots/{slot}/withdrawals", hs.PublicGetNetworkSlotWithdrawals, nil},
		{get, "/networks/{network}/blocks/{block}/withdrawals", hs.PublicGetNetworkBlockWithdrawals, nil},
		{get, "/networks/{network}/validators/{validator}/withdrawals", hs.PublicGetNetworkValidatorWithdrawals, nil},
		{get, "/networks/{network}/withdrawal-credentials/{credential}/withdrawals", hs.PublicGetNetworkWithdrawalCredentialWithdrawals, nil},

		{get, "/networks/{network}/voluntary-exits", hs.PublicGetNetworkVoluntaryExits, nil},
		{get, "/networks/{network}/epochs/{epoch}/voluntary-exits", hs.PublicGetNetworkEpochVoluntaryExits, nil},
		{get, "/networks/{network}/slots/{slot}/voluntary-exits", hs.PublicGetNetworkSlotVoluntaryExits, nil},
		{get, "/networks/{network}/blocks/{block}/voluntary-exits", hs.PublicGetNetworkBlockVoluntaryExits, nil},

		{get, "/networks/{network}/addresses/{address}/balance-history", hs.PublicGetNetworkAddressBalanceHistory, nil},
		{get, "/networks/{network}/addresses/{address}/token-supply-history", hs.PublicGetNetworkAddressTokenSupplyHistory, nil},
		{get, "/networks/{network}/addresses/{address}/event-logs", hs.PublicGetNetworkAddressEventLogs, nil},

		{get, "/networks/{network}/transactions", hs.PublicGetNetworkTransactions, nil},
		{get, "/networks/{network}/transactions/{hash}", hs.PublicGetNetworkTransaction, nil},
		{get, "/networks/{network}/addresses/{address}/transactions", hs.PublicGetNetworkAddressTransactions, nil},
		{get, "/networks/{network}/slots/{slot}/transactions", hs.PublicGetNetworkSlotTransactions, nil},
		{get, "/networks/{network}/blocks/{block}/transactions", hs.PublicGetNetworkBlockTransactions, nil},
		{get, "/networks/{network}/blocks/{block}/blobs", hs.PublicGetNetworkBlockBlobs, nil},

		{get, "/networks/{network}/handlerService-changes", hs.PublicGetNetworkBlsChanges, nil},
		{get, "/networks/{network}/epochs/{epoch}/handlerService-changes", hs.PublicGetNetworkEpochBlsChanges, nil},
		{get, "/networks/{network}/slots/{slot}/handlerService-changes", hs.PublicGetNetworkSlotBlsChanges, nil},
		{get, "/networks/{network}/blocks/{block}/handlerService-changes", hs.PublicGetNetworkBlockBlsChanges, nil},
		{get, "/networks/{network}/validators/{validator}/handlerService-changes", hs.PublicGetNetworkValidatorBlsChanges, nil},

		{get, "/networks/ethereum/addresses/{address}/ens", hs.PublicGetNetworkAddressEns, nil},
		{get, "/networks/ethereum/ens/{ens_name}", hs.PublicGetNetworkEns, nil},

		{get, "/networks/{layer_2_network}/batches", hs.PublicGetNetworkBatches, nil},
		{get, "/networks/{layer_2_network}/layer1-to-layer2-transactions", hs.PublicGetNetworkLayer1ToLayer2Transactions, nil},
		{get, "/networks/{layer_2_network}/layer2-to-layer1-transactions", hs.PublicGetNetworkLayer2ToLayer1Transactions, nil},

		{post, "/networks/{network}/broadcasts", hs.PublicPostNetworkBroadcasts, nil},
		{get, "/eth-price-history", hs.PublicGetEthPriceHistory, nil},

		{get, "/networks/{network}/gasnow", hs.PublicGetNetworkGasNow, nil},
		{get, "/networks/{network}/average-gas-limit-history", hs.PublicGetNetworkAverageGasLimitHistory, nil},
		{get, "/networks/{network}/gas-used-history", hs.PublicGetNetworkGasUsedHistory, nil},

		{get, "/rocket-pool/nodes", hs.PublicGetRocketPoolNodes, nil},
		{get, "/rocket-pool/minipools", hs.PublicGetRocketPoolMinipools, nil},

		{get, "/networks/{network}/sync-committee/{period}", hs.PublicGetNetworkSyncCommittee, nil},

		{get, "/multisig-safes/{address}", hs.PublicGetMultisigSafe, nil},
		{get, "/multisig-safes/{address}/transactions", hs.PublicGetMultisigSafeTransactions, nil},
		{get, "/multisig-transactions/{hash}/confirmations", hs.PublicGetMultisigTransactionConfirmations, nil},
	}
	for _, endpoint := range endpoints {
		if endpoint.PublicHandler != nil {
			publicRouter.HandleFunc(endpoint.Path, endpoint.PublicHandler).Methods(endpoint.Method, options)
		}
		if endpoint.InternalHander != nil {
			internalRouter.HandleFunc(endpoint.Path, endpoint.InternalHander).Methods(endpoint.Method, options)
		}
	}
}

func addValidatorDashboardRoutes(hs *handlers.HandlerService, publicRouter, internalRouter *mux.Router) {
	vdbPath := "/validator-dashboards"
	publicRouter.HandleFunc(vdbPath, hs.PublicPostValidatorDashboards).Methods(post, options)
	internalRouter.HandleFunc(vdbPath, hs.InternalPostValidatorDashboards).Methods(post, options)

	publicDashboardRouter := publicRouter.PathPrefix(vdbPath).Subrouter()
	internalDashboardRouter := internalRouter.PathPrefix(vdbPath).Subrouter()
	// add middleware to check if user has access to dashboard
	publicDashboardRouter.Use(hs.VDBAuthMiddleware, hs.VDBValidationMiddleware)
	internalDashboardRouter.Use(hs.VDBAuthMiddleware, hs.VDBValidationMiddleware)

	endpoints := []endpoint{
		{get, "/{dashboard_id}", hs.PublicGetValidatorDashboard, hs.InternalGetValidatorDashboard},
		{delete, "/{dashboard_id}", hs.PublicDeleteValidatorDashboard, hs.InternalDeleteValidatorDashboard},
		{post, "/{dashboard_id}/groups", hs.PublicPostValidatorDashboardGroups, hs.InternalPostValidatorDashboardGroups},
		{put, "/{dashboard_id}/groups/{group_id}", hs.PublicPutValidatorDashboardGroups, hs.InternalPutValidatorDashboardGroups},
		{delete, "/{dashboard_id}/groups/{group_id}", hs.PublicDeleteValidatorDashboardGroups, hs.InternalDeleteValidatorDashboardGroups},
		{post, "/{dashboard_id}/validators", hs.PublicPostValidatorDashboardValidators, hs.InternalPostValidatorDashboardValidators},
		{get, "/{dashboard_id}/validators", hs.PublicGetValidatorDashboardValidators, hs.InternalGetValidatorDashboardValidators},
		{delete, "/{dashboard_id}/validators", hs.PublicDeleteValidatorDashboardValidators, hs.InternalDeleteValidatorDashboardValidators},
		{post, "/{dashboard_id}/public-ids", hs.PublicPostValidatorDashboardPublicIds, hs.InternalPostValidatorDashboardPublicIds},
		{put, "/{dashboard_id}/public-ids/{public_id}", hs.PublicPutValidatorDashboardPublicId, hs.InternalPutValidatorDashboardPublicId},
		{delete, "/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteValidatorDashboardPublicId, hs.InternalDeleteValidatorDashboardPublicId},
		{get, "/{dashboard_id}/slot-viz", hs.PublicGetValidatorDashboardSlotViz, hs.InternalGetValidatorDashboardSlotViz},
		{get, "/{dashboard_id}/summary", hs.PublicGetValidatorDashboardSummary, hs.InternalGetValidatorDashboardSummary},
		{get, "/{dashboard_id}/validator-indices", nil, hs.InternalGetValidatorDashboardValidatorIndices},
		{get, "/{dashboard_id}/groups/{group_id}/summary", hs.PublicGetValidatorDashboardGroupSummary, hs.InternalGetValidatorDashboardGroupSummary},
		{get, "/{dashboard_id}/summary-chart", hs.PublicGetValidatorDashboardSummaryChart, hs.InternalGetValidatorDashboardSummaryChart},
		{get, "/{dashboard_id}/rewards", hs.PublicGetValidatorDashboardRewards, hs.InternalGetValidatorDashboardRewards},
		{get, "/{dashboard_id}/groups/{group_id}/rewards/{epoch}", hs.PublicGetValidatorDashboardGroupRewards, hs.InternalGetValidatorDashboardGroupRewards},
		{get, "/{dashboard_id}/rewards-chart", hs.PublicGetValidatorDashboardRewardsChart, hs.InternalGetValidatorDashboardRewardsChart},
		{get, "/{dashboard_id}/duties/{epoch}", hs.PublicGetValidatorDashboardDuties, hs.InternalGetValidatorDashboardDuties},
		{get, "/{dashboard_id}/blocks", hs.PublicGetValidatorDashboardBlocks, hs.InternalGetValidatorDashboardBlocks},
		{get, "/{dashboard_id}/heatmap", hs.PublicGetValidatorDashboardHeatmap, hs.InternalGetValidatorDashboardHeatmap},
		{get, "/{dashboard_id}/groups/{group_id}/heatmap", hs.PublicGetValidatorDashboardGroupHeatmap, hs.InternalGetValidatorDashboardGroupHeatmap},
		{get, "/{dashboard_id}/execution-layer-deposits", hs.PublicGetValidatorDashboardExecutionLayerDeposits, hs.InternalGetValidatorDashboardExecutionLayerDeposits},
		{get, "/{dashboard_id}/consensus-layer-deposits", hs.PublicGetValidatorDashboardConsensusLayerDeposits, hs.InternalGetValidatorDashboardConsensusLayerDeposits},
		{get, "/{dashboard_id}/withdrawals", hs.PublicGetValidatorDashboardWithdrawals, hs.InternalGetValidatorDashboardWithdrawals},
	}
	for _, endpoint := range endpoints {
		if endpoint.PublicHandler != nil {
			publicDashboardRouter.HandleFunc(endpoint.Path, endpoint.PublicHandler).Methods(endpoint.Method, options)
		}
		if endpoint.InternalHander != nil {
			internalDashboardRouter.HandleFunc(endpoint.Path, endpoint.InternalHander).Methods(endpoint.Method, options)
		}
	}
}
