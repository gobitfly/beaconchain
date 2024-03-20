package api

import (
	"net/http"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	handlers "github.com/gobitfly/beaconchain/pkg/api/handlers"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gorilla/mux"
)

type endpoint struct {
	Method         string
	Path           string
	PublicHandler  func(w http.ResponseWriter, r *http.Request)
	InternalHander func(w http.ResponseWriter, r *http.Request)
}

func NewApiRouter(dai dataaccess.DataAccessor, cfg *types.Config) *mux.Router {
	handlerService := handlers.NewHandlerService(dai)
	router := mux.NewRouter().PathPrefix("/api").Subrouter()
	publicRouter := router.PathPrefix("/v2").Subrouter()
	internalRouter := router.PathPrefix("/i").Subrouter()

	addRoutes(handlerService, publicRouter, internalRouter)

	return router
}

// TODO replace with proper auth
func GetAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		query := r.URL.Query().Get("api_key")

		if header != "Bearer "+utils.Config.ApiKeySecret && query != utils.Config.ApiKeySecret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func addRoutes(hs handlers.HandlerService, publicRouter, internalRouter *mux.Router) {
	endpoints := []endpoint{
		{"GET", "/healthz", hs.PublicGetHealthz, nil},
		{"GET", "/healthz-loadbalancer", hs.PublicGetHealthzLoadbalancer, nil},

		{"POST", "/oauth/token", hs.PublicPostOauthToken, nil},

		{"GET", "/users/me/dashboards", hs.PublicGetUserDashboards, hs.InternalGetUserDashboards},

		{"POST", "/account-dashboards", hs.PublicPostAccountDashboards, hs.InternalPostAccountDashboards},
		{"GET", "/account-dashboards/{dashboard_id}", hs.PublicGetAccountDashboard, hs.InternalGetAccountDashboard},
		{"DELETE", "/account-dashboards/{dashboard_id}", hs.PublicDeleteAccountDashboard, hs.InternalDeleteAccountDashboard},
		{"POST", "/account-dashboards/{dashboard_id}/groups", hs.PublicPostAccountDashboardGroups, hs.InternalPostAccountDashboardGroups},
		{"DELETE", "/account-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteAccountDashboardGroups, hs.InternalDeleteAccountDashboardGroups},
		{"POST", "/account-dashboards/{dashboard_id}/accounts", hs.PublicPostAccountDashboardAccounts, hs.InternalPostAccountDashboardAccounts},
		{"GET", "/account-dashboards/{dashboard_id}/accounts", hs.PublicGetAccountDashboardAccounts, hs.InternalGetAccountDashboardAccounts},
		{"DELETE", "/account-dashboards/{dashboard_id}/accounts", hs.PublicDeleteAccountDashboardAccounts, hs.InternalDeleteAccountDashboardAccounts},
		{"PUT", "/account-dashboards/{dashboard_id}/accounts/{address}", hs.PublicPutAccountDashboardAccount, hs.InternalPutAccountDashboardAccount},
		{"POST", "/account-dashboards/{dashboard_id}/public-ids", hs.PublicPostAccountDashboardPublicIds, hs.InternalPostAccountDashboardPublicIds},
		{"PUT", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutAccountDashboardPublicId, hs.InternalPutAccountDashboardPublicId},
		{"DELETE", "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteAccountDashboardPublicId, hs.InternalDeleteAccountDashboardPublicId},
		{"GET", "/account-dashboards/{dashboard_id}/transactions", hs.PublicGetAccountDashboardTransactions, hs.InternalGetAccountDashboardTransactions},
		{"PUT", "/account-dashboards/{dashboard_id}/transactions/settings", hs.PublicPutAccountDashboardTransactionsSettings, hs.InternalPutAccountDashboardTransactionsSettings},

		{"POST", "/validator-dashboards", hs.PublicPostValidatorDashboards, hs.InternalPostValidatorDashboards},
		{"GET", "/validator-dashboards/{dashboard_id}", hs.PublicGetValidatorDashboard, hs.InternalGetValidatorDashboard},
		{"DELETE", "/validator-dashboards/{dashboard_id}", hs.PublicDeleteValidatorDashboard, hs.InternalDeleteValidatorDashboard},
		{"POST", "/validator-dashboards/{dashboard_id}/groups", hs.PublicPostValidatorDashboardGroups, hs.InternalPostValidatorDashboardGroups},
		{"DELETE", "/validator-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteValidatorDashboardGroups, hs.InternalDeleteValidatorDashboardGroups},
		{"POST", "/validator-dashboards/{dashboard_id}/validators", hs.PublicPostValidatorDashboardValidators, hs.InternalPostValidatorDashboardValidators},
		{"GET", "/validator-dashboards/{dashboard_id}/validators", hs.PublicGetValidatorDashboardValidators, hs.InternalGetValidatorDashboardValidators},
		{"DELETE", "/validator-dashboards/{dashboard_id}/validators", hs.PublicDeleteValidatorDashboardValidators, hs.InternalDeleteValidatorDashboardValidators},
		{"POST", "/validator-dashboards/{dashboard_id}/public-ids", hs.PublicPostValidatorDashboardPublicIds, hs.InternalPostValidatorDashboardPublicIds},
		{"PUT", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutValidatorDashboardPublicId, hs.InternalPutValidatorDashboardPublicId},
		{"DELETE", "/validator-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteValidatorDashboardPublicId, hs.InternalDeleteValidatorDashboardPublicId},
		{"GET", "/validator-dashboards/{dashboard_id}/slot-viz", hs.PublicGetValidatorDashboardSlotViz, hs.InternalGetValidatorDashboardSlotViz},
		{"GET", "/validator-dashboards/{dashboard_id}/summary", hs.PublicGetValidatorDashboardSummary, hs.InternalGetValidatorDashboardSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/summary", hs.PublicGetValidatorDashboardGroupSummary, hs.InternalGetValidatorDashboardGroupSummary},
		{"GET", "/validator-dashboards/{dashboard_id}/summary-chart", hs.PublicGetValidatorDashboardSummaryChart, hs.InternalGetValidatorDashboardSummaryChart},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards", hs.PublicGetValidatorDashboardRewards, hs.InternalGetValidatorDashboardRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/rewards", hs.PublicGetValidatorDashboardGroupRewards, hs.InternalGetValidatorDashboardGroupRewards},
		{"GET", "/validator-dashboards/{dashboard_id}/rewards-chart", hs.PublicGetValidatorDashboardRewardsChart, hs.InternalGetValidatorDashboardRewardsChart},
		{"GET", "/validator-dashboards/{dashboard_id}/duties/{epoch}", hs.PublicGetValidatorDashboardDuties, hs.InternalGetValidatorDashboardDuties},
		{"GET", "/validator-dashboards/{dashboard_id}/blocks", hs.PublicGetValidatorDashboardBlocks, hs.InternalGetValidatorDashboardBlocks},
		{"GET", "/validator-dashboards/{dashboard_id}/heatmap", hs.PublicGetValidatorDashboardHeatmap, hs.InternalGetValidatorDashboardHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/groups/{group_id}/heatmap", hs.PublicGetValidatorDashboardGroupHeatmap, hs.InternalGetValidatorDashboardGroupHeatmap},
		{"GET", "/validator-dashboards/{dashboard_id}/execution-layer-deposits", hs.PublicGetValidatorDashboardExecutionLayerDeposits, hs.InternalGetValidatorDashboardExecutionLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/consensus-layer-deposits", hs.PublicGetValidatorDashboardConsensusLayerDeposits, hs.InternalGetValidatorDashboardConsensusLayerDeposits},
		{"GET", "/validator-dashboards/{dashboard_id}/withdrawals", hs.PublicGetValidatorDashboardWithdrawals, hs.InternalGetValidatorDashboardWithdrawals},

		{"GET", "/networks/{network}/validators", hs.PublicGetNetworkValidators, nil},
		{"GET", "/networks/{network}/validators/{validator}", hs.PublicGetNetworkValidator, nil},
		{"GET", "/networks/{network}/validators/{validator}/duties", hs.PublicGetNetworkValidatorDuties, nil},
		{"GET", "/networks/{network}/addresses/{address}/validators", hs.PublicGetNetworkAddressValidators, nil},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/validators", hs.PublicGetNetworkWithdrawalCredentialValidators, nil},
		{"GET", "/networks/{network}/validator-statuses", hs.PublicGetNetworkValidatorStatuses, nil},
		{"GET", "/networks/{network}/validator-leaderboard", hs.PublicGetNetworkValidatorLeaderboard, nil},
		{"GET", "/networks/{network}/validator-queue", hs.PublicGetNetworkValidatorQueue, nil},

		{"GET", "/networks/{network}/epochs", hs.PublicGetNetworkEpochs, nil},
		{"GET", "/networks/{network}/epochs/{epoch}", hs.PublicGetNetworkEpoch, nil},

		{"GET", "/networks/{network}/blocks", hs.PublicGetNetworkBlocks, nil},
		{"GET", "/networks/{network}/blocks/{block}", hs.PublicGetNetworkBlock, nil},
		{"GET", "/networks/{network}/slots", hs.PublicGetNetworkSlots, nil},
		{"GET", "/networks/{network}/slots/{slot}", hs.PublicGetNetworkSlot, nil},
		{"GET", "/networks/{network}/validators/{validator}/blocks", hs.PublicGetNetworkValidatorBlocks, nil},
		{"GET", "/networks/{network}/addresses/{address}/priority-fee-blocks", hs.PublicGetNetworkAddressPriorityFeeBlocks, nil},
		{"GET", "/networks/{network}/addresses/{address}/proposer-reward-blocks", hs.PublicGetNetworkAddressProposerRewardBlocks, nil},
		{"GET", "/networks/{network}/forked-blocks", hs.PublicGetNetworkForkedBlocks, nil},
		{"GET", "/networks/{network}/forked-blocks/{block}", hs.PublicGetNetworkForkedBlock, nil},
		{"GET", "/networks/{network}/forked-slots/{slot}", hs.PublicGetNetworkForkedSlot, nil},
		{"GET", "/networks/{network}/block-sizes", hs.PublicGetNetworkBlockSizes, nil},

		{"GET", "/networks/{network}/validators/{validator}/attestations", hs.PublicGetNetworkValidatorAttestations, nil},
		{"GET", "/networks/{network}/epochs/{epoch}/attestations", hs.PublicGetNetworkEpochAttestations, nil},
		{"GET", "/networks/{network}/slots/{slot}/attestations", hs.PublicGetNetworkSlotAttestations, nil},
		{"GET", "/networks/{network}/blocks/{block}/attestations", hs.PublicGetNetworkBlockAttestations, nil},
		{"GET", "/networks/{network}/aggregated-attestations", hs.PublicGetNetworkAggregatedAttestations, nil},

		{"GET", "/networks/{network}/ethstore/{day}", hs.PublicGetNetworkEthStore, nil},
		{"GET", "/networks/{network}/validators/{validator}/reward-history", hs.PublicGetNetworkValidatorRewardHistory, nil},
		{"GET", "/networks/{network}/validators/{validator}/balance-history", hs.PublicGetNetworkValidatorBalanceHistory, nil},
		{"GET", "/networks/{network}/validators/{validator}/performance-history", hs.PublicGetNetworkValidatorPerformanceHistory, nil},

		{"GET", "/networks/{network}/slashings", hs.PublicGetNetworkSlashings, nil},
		{"GET", "/networks/{network}/validators/{validator}/slashings", hs.PublicGetNetworkValidatorSlashings, nil},

		{"GET", "/networks/{network}/deposits", hs.PublicGetNetworkDeposits, nil},
		{"GET", "/networks/{network}/validators/{validator}/deposits", hs.PublicGetNetworkValidatorDeposits, nil},
		{"GET", "/networks/{network}/transactions/{hash}/deposits", hs.PublicGetNetworkTransactionDeposits, nil},

		{"GET", "/networks/{network}/withdrawals", hs.PublicGetNetworkWithdrawals, nil},
		{"GET", "/networks/{network}/slots/{slot}/withdrawals", hs.PublicGetNetworkSlotWithdrawals, nil},
		{"GET", "/networks/{network}/blocks/{block}/withdrawals", hs.PublicGetNetworkBlockWithdrawals, nil},
		{"GET", "/networks/{network}/validators/{validator}/withdrawals", hs.PublicGetNetworkValidatorWithdrawals, nil},
		{"GET", "/networks/{network}/withdrawal-credentials/{credential}/withdrawals", hs.PublicGetNetworkWithdrawalCredentialWithdrawals, nil},

		{"GET", "/networks/{network}/voluntary-exits", hs.PublicGetNetworkVoluntaryExits, nil},
		{"GET", "/networks/{network}/epochs/{epoch}/voluntary-exits", hs.PublicGetNetworkEpochVoluntaryExits, nil},
		{"GET", "/networks/{network}/slots/{slot}/voluntary-exits", hs.PublicGetNetworkSlotVoluntaryExits, nil},
		{"GET", "/networks/{network}/blocks/{block}/voluntary-exits", hs.PublicGetNetworkBlockVoluntaryExits, nil},

		{"GET", "/networks/{network}/addresses/{address}/balance-history", hs.PublicGetNetworkAddressBalanceHistory, nil},
		{"GET", "/networks/{network}/addresses/{address}/token-supply-history", hs.PublicGetNetworkAddressTokenSupplyHistory, nil},
		{"GET", "/networks/{network}/addresses/{address}/event-logs", hs.PublicGetNetworkAddressEventLogs, nil},

		{"GET", "/networks/{network}/transactions", hs.PublicGetNetworkTransactions, nil},
		{"GET", "/networks/{network}/transactions/{hash}", hs.PublicGetNetworkTransaction, nil},
		{"GET", "/networks/{network}/addresses/{address}/transactions", hs.PublicGetNetworkAddressTransactions, nil},
		{"GET", "/networks/{network}/slots/{slot}/transactions", hs.PublicGetNetworkSlotTransactions, nil},
		{"GET", "/networks/{network}/blocks/{block}/transactions", hs.PublicGetNetworkBlockTransactions, nil},
		{"GET", "/networks/{network}/blocks/{block}/blobs", hs.PublicGetNetworkBlockBlobs, nil},

		{"GET", "/networks/{network}/handlerService-changes", hs.PublicGetNetworkBlsChanges, nil},
		{"GET", "/networks/{network}/epochs/{epoch}/handlerService-changes", hs.PublicGetNetworkEpochBlsChanges, nil},
		{"GET", "/networks/{network}/slots/{slot}/handlerService-changes", hs.PublicGetNetworkSlotBlsChanges, nil},
		{"GET", "/networks/{network}/blocks/{block}/handlerService-changes", hs.PublicGetNetworkBlockBlsChanges, nil},
		{"GET", "/networks/{network}/validators/{validator}/handlerService-changes", hs.PublicGetNetworkValidatorBlsChanges, nil},

		{"GET", "/networks/ethereum/addresses/{address}/ens", hs.PublicGetNetworkAddressEns, nil},
		{"GET", "/networks/ethereum/ens/{ens_name}", hs.PublicGetNetworkEns, nil},

		{"GET", "/networks/{layer_2_network}/batches", hs.PublicGetNetworkBatches, nil},
		{"GET", "/networks/{layer_2_network}/layer1-to-layer2-transactions", hs.PublicGetNetworkLayer1ToLayer2Transactions, nil},
		{"GET", "/networks/{layer_2_network}/layer2-to-layer1-transactions", hs.PublicGetNetworkLayer2ToLayer1Transactions, nil},

		{"POST", "/networks/{network}/broadcasts", hs.PublicPostNetworkBroadcasts, nil},
		{"GET", "/eth-price-history", hs.PublicGetEthPriceHistory, nil},

		{"GET", "/networks/{network}/gasnow", hs.PublicGetNetworkGasNow, nil},
		{"GET", "/networks/{network}/average-gas-limit-history", hs.PublicGetNetworkAverageGasLimitHistory, nil},
		{"GET", "/networks/{network}/gas-used-history", hs.PublicGetNetworkGasUsedHistory, nil},

		{"GET", "/rocket-pool/nodes", hs.PublicGetRocketPoolNodes, nil},
		{"GET", "/rocket-pool/minipools", hs.PublicGetRocketPoolMinipools, nil},

		{"GET", "/networks/{network}/sync-committee/{period}", hs.PublicGetNetworkSyncCommittee, nil},

		{"GET", "/multisig-safes/{address}", hs.PublicGetMultisigSafe, nil},
		{"GET", "/multisig-safes/{address}/transactions", hs.PublicGetMultisigSafeTransactions, nil},
		{"GET", "/multisig-transactions/{hash}/confirmations", hs.PublicGetMultisigTransactionConfirmations, nil},
	}
	for _, endpoint := range endpoints {
		if endpoint.PublicHandler != nil {
			publicRouter.HandleFunc(endpoint.Path, endpoint.PublicHandler).Methods(endpoint.Method, http.MethodOptions)
		}
		if endpoint.InternalHander != nil {
			internalRouter.HandleFunc(endpoint.Path, endpoint.InternalHander).Methods(endpoint.Method, http.MethodOptions)
		}
	}
}
