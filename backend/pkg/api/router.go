package api

import (
	"fmt"
	"net/http"
	"regexp"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	handlers "github.com/gobitfly/beaconchain/pkg/api/handlers"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type endpoint struct {
	Method         string
	Path           string
	PublicHandler  func(w http.ResponseWriter, r *http.Request)
	InternalHander func(w http.ResponseWriter, r *http.Request)
}

func NewApiRouter(dataAccessor dataaccess.DataAccessor, cfg *types.Config) *mux.Router {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	publicRouter := apiRouter.PathPrefix("/v2").Subrouter()
	internalRouter := apiRouter.PathPrefix("/i").Subrouter()
	sessionManager := newSessionManager(cfg)
	internalRouter.Use(sessionManager.LoadAndSave)

	debug := cfg.Frontend.Debug
	if !debug {
		internalRouter.Use(getCsrfProtectionMiddleware(cfg), csrfInjecterMiddleware)
	}
	handlerService := handlers.NewHandlerService(dataAccessor, sessionManager)

	// TODO @patrick - remove this test route
	router.HandleFunc("/test/stripe", TestStripe).Methods(http.MethodGet)
	apiRouter.HandleFunc("/test/stripe", TestStripe).Methods(http.MethodGet)

	addRoutes(handlerService, publicRouter, internalRouter, cfg)

	return router
}

// TODO @patrick - remove this test route
func TestStripe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(fmt.Sprintf(`
<h2>manage</h2>
<div>
<button class="manage-billing">Manage Billing</button>
</div>
<div>
<button class="purchase" data-product-id="plankton">purchase plankton</button>
<button class="purchase" data-product-id="goldfish">purchase goldfish</button>
<button class="purchase" data-product-id="whale">purchase whale</button>
<button class="purchase" data-product-id="guppy">purchase guppy</button>
<button class="purchase" data-product-id="guppyYearly">purchase guppy.yearly</button>
<button class="purchase" data-product-id="dolphin">purchase dolphin</button>
<button class="purchase" data-product-id="dolphinYearly">purchase dolphin.yearly</button>
<button class="purchase" data-product-id="orca">purchase orca</button>
<button class="purchase" data-product-id="orcaYearly">purchase orca.yearly</button>
</div>
<div>
<input type="number" id="addon-quantity" value="1">
<button class="purchase" data-product-id="vdbAddon1k">purchase vdbAddon1k</button>
<button class="purchase" data-product-id="vdbAddon1kYearly">purchase vdbAddon1kYearly</button>
<button class="purchase" data-product-id="vdbAddon10k">purchase vdbAddon10k</button>
<button class="purchase" data-product-id="vdbAddon10kYearly">purchase vdbAddon10kYearly</button>
</div>

<h2>info</h2>
<pre id="userInfoRaw"></pre>

<script src="https://js.stripe.com/v3/"></script>
<script>

var config = {
	betaKey  	     : "%[1]s",
	publicKey        : "%[2]s",
	plankton         : "%[3]s",
	goldfish         : "%[4]s",
	whale            : "%[5]s",
	guppy            : "%[6]s",
	dolphin          : "%[7]s",
	orca             : "%[8]s",
	vdbAddon1k       : "%[9]s",
	vdbAddon10k      : "%[10]s",
	guppyYearly      : "%[11]s",
	dolphinYearly    : "%[12]s",
	orcaYearly       : "%[13]s",
	vdbAddon1kYearly : "%[14]s",
	vdbAddon10kYearly: "%[15]s",
}
console.log('config',config)

fetch('/api/i/users/me',{headers:{'Authorization':'Bearer '+config.betaKey}}).then((r)=>{
	config.csrfToken = r.headers.get('x-csrf-token')
	return r.json()
}).then((d)=>{
	console.log('userInfo',d)
	document.getElementById('userInfoRaw').innerText = JSON.stringify(d, null, 2)
}).catch(err => {
	console.error("error getting api user me", err)
})

function handleFetchResult(result) {
	if (!result.ok) {
		return result.json().then(function (json) {
			if (json.error && json.error.message) {
				throw new Error(result.url + ' ' + result.status + ' ' + json.error.message)
			}
		})
	}
	return result.json()
}

function createCheckoutSession(priceId) {
	let addonQuantity = parseInt(document.getElementById("addon-quantity").value)
	if (isNaN(addonQuantity)) addonQuantity = 1
	return fetch("/user/stripe/create-checkout-session", {
		method: "POST",
		headers: { 
			"Content-Type": "application/json",
			"X-CSRF-Token": config.csrfToken
		},
		credentials: 'include',
		body: JSON.stringify({ priceId: priceId, addonQuantity: addonQuantity })
	})
	.then(handleFetchResult)
	.catch(err => {
		console.error("error posting to create checkout session endpoint", err)
	})
}

function setupStripe() {
	try {
		var stripe = Stripe(config.publicKey)
		var purchaseButtons = document.querySelectorAll(".purchase")
		for (let i = 0; i < purchaseButtons.length; i++) {
			purchaseButtons[i].addEventListener('click', function(e) {
				let priceId = config[e.target.getAttribute('data-product-id')]
				createCheckoutSession(priceId).then((d) => {
					stripe.redirectToCheckout({ sessionId: d.sessionId }).then(handleResult).catch(err => {
						console.error("error redirecting to stripe checkout", err)
					})
				})
			})
		}
	} catch (err) {
		console.error("error creating stripe object", err)
	}
}

var manageBillingButtons = document.querySelectorAll(".manage-billing")
for (let i = 0; i < manageBillingButtons.length; i++) {
	manageBillingButtons[i].addEventListener("click", function (e) {
    fetch("/user/stripe/customer-portal", {
      method: "POST",
      headers: { 
		"Content-Type": "application/json",
		"X-CSRF-Token": config.csrfToken
	},
      credentials: "include",
      body: JSON.stringify({returnURL: window.location.href}),
    })
      .then((response) => response.json())
      .then((data) => {
        // window.location.href = data.url
		window.open(data.url, '_blank').focus();
      })
      .catch((error) => {
        console.error("Error:", error)
      })
  })
}

setupStripe()
</script>
`,
		utils.Config.ApiKeySecret,
		utils.Config.Frontend.Stripe.PublicKey,
		utils.Config.Frontend.Stripe.Plankton,
		utils.Config.Frontend.Stripe.Goldfish,
		utils.Config.Frontend.Stripe.Whale,
		utils.Config.Frontend.Stripe.Guppy,
		utils.Config.Frontend.Stripe.Dolphin,
		utils.Config.Frontend.Stripe.Orca,
		utils.Config.Frontend.Stripe.VdbAddon1k,
		utils.Config.Frontend.Stripe.VdbAddon10k,
		utils.Config.Frontend.Stripe.GuppyYearly,
		utils.Config.Frontend.Stripe.DolphinYearly,
		utils.Config.Frontend.Stripe.OrcaYearly,
		utils.Config.Frontend.Stripe.VdbAddon1kYearly,
		utils.Config.Frontend.Stripe.VdbAddon10kYearly,
	)))
	if err != nil {
		log.Error(err, "error writing response", 0)
		http.Error(w, "error writing response", http.StatusInternalServerError)
	}
}

func GetCorsMiddleware(allowedHosts []string) func(http.Handler) http.Handler {
	if len(allowedHosts) == 0 {
		log.Warn("CORS allowed hosts not set, allowing all origins")
		return gorillaHandlers.CORS(
			gorillaHandlers.AllowedOrigins([]string{"*"}),
			gorillaHandlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodHead}),
			gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-CSRF-Token"}),
			gorillaHandlers.ExposedHeaders([]string{"X-CSRF-Token"}),
		)
	}

	allowedHostsRegex := make([]*regexp.Regexp, len(allowedHosts))
	var err error
	for i, host := range allowedHosts {
		allowedHostsRegex[i], err = regexp.Compile(host)

		if err != nil {
			log.Fatal(err, "error compiling allowed host regex", 0)
		}
	}

	return gorillaHandlers.CORS(
		gorillaHandlers.AllowedOriginValidator(func(s string) bool {
			for _, host := range allowedHostsRegex {
				if host.MatchString(s) {
					return true
				}
			}
			return false
		}),
		gorillaHandlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodHead}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-CSRF-Token"}),
		gorillaHandlers.ExposedHeaders([]string{"X-CSRF-Token"}),
		gorillaHandlers.AllowCredentials(),
	)
}

func addRoutes(hs *handlers.HandlerService, publicRouter, internalRouter *mux.Router, cfg *types.Config) {
	addValidatorDashboardRoutes(hs, publicRouter, internalRouter, cfg)
	endpoints := []endpoint{
		{http.MethodGet, "/healthz", hs.PublicGetHealthz, nil},
		{http.MethodGet, "/healthz-loadbalancer", hs.PublicGetHealthzLoadbalancer, nil},

		{http.MethodPost, "/login", nil, hs.InternalPostLogin},
		{http.MethodPost, "/logout", nil, hs.InternalPostLogout},

		{http.MethodPost, "/oauth/token", hs.PublicPostOauthToken, nil},

		{http.MethodGet, "/latest-state", nil, hs.InternalGetLatestState},

		{http.MethodGet, "/product-summary", nil, hs.InternalGetProductSummary},
		{http.MethodGet, "/users/me", nil, hs.InternalGetUserInfo},

		{http.MethodGet, "/users/me/dashboards", hs.PublicGetUserDashboards, hs.InternalGetUserDashboards},

		{http.MethodPost, "/search", nil, hs.InternalPostSearch},

		{http.MethodPost, "/account-dashboards", hs.PublicPostAccountDashboards, hs.InternalPostAccountDashboards},
		{http.MethodGet, "/account-dashboards/{dashboard_id}", hs.PublicGetAccountDashboard, hs.InternalGetAccountDashboard},
		{http.MethodDelete, "/account-dashboards/{dashboard_id}", hs.PublicDeleteAccountDashboard, hs.InternalDeleteAccountDashboard},
		{http.MethodPost, "/account-dashboards/{dashboard_id}/groups", hs.PublicPostAccountDashboardGroups, hs.InternalPostAccountDashboardGroups},
		{http.MethodDelete, "/account-dashboards/{dashboard_id}/groups/{group_id}", hs.PublicDeleteAccountDashboardGroups, hs.InternalDeleteAccountDashboardGroups},
		{http.MethodPost, "/account-dashboards/{dashboard_id}/accounts", hs.PublicPostAccountDashboardAccounts, hs.InternalPostAccountDashboardAccounts},
		{http.MethodGet, "/account-dashboards/{dashboard_id}/accounts", hs.PublicGetAccountDashboardAccounts, hs.InternalGetAccountDashboardAccounts},
		{http.MethodDelete, "/account-dashboards/{dashboard_id}/accounts", hs.PublicDeleteAccountDashboardAccounts, hs.InternalDeleteAccountDashboardAccounts},
		{http.MethodPut, "/account-dashboards/{dashboard_id}/accounts/{address}", hs.PublicPutAccountDashboardAccount, hs.InternalPutAccountDashboardAccount},
		{http.MethodPost, "/account-dashboards/{dashboard_id}/public-ids", hs.PublicPostAccountDashboardPublicIds, hs.InternalPostAccountDashboardPublicIds},
		{http.MethodPut, "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicPutAccountDashboardPublicId, hs.InternalPutAccountDashboardPublicId},
		{http.MethodDelete, "/account-dashboards/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteAccountDashboardPublicId, hs.InternalDeleteAccountDashboardPublicId},
		{http.MethodGet, "/account-dashboards/{dashboard_id}/transactions", hs.PublicGetAccountDashboardTransactions, hs.InternalGetAccountDashboardTransactions},
		{http.MethodPut, "/account-dashboards/{dashboard_id}/transactions/settings", hs.PublicPutAccountDashboardTransactionsSettings, hs.InternalPutAccountDashboardTransactionsSettings},

		{http.MethodGet, "/networks/{network}/validators", hs.PublicGetNetworkValidators, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}", hs.PublicGetNetworkValidator, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/duties", hs.PublicGetNetworkValidatorDuties, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/validators", hs.PublicGetNetworkAddressValidators, nil},
		{http.MethodGet, "/networks/{network}/withdrawal-credentials/{credential}/validators", hs.PublicGetNetworkWithdrawalCredentialValidators, nil},
		{http.MethodGet, "/networks/{network}/validator-statuses", hs.PublicGetNetworkValidatorStatuses, nil},
		{http.MethodGet, "/networks/{network}/validator-leaderboard", hs.PublicGetNetworkValidatorLeaderboard, nil},
		{http.MethodGet, "/networks/{network}/validator-queue", hs.PublicGetNetworkValidatorQueue, nil},

		{http.MethodGet, "/networks/{network}/epochs", hs.PublicGetNetworkEpochs, nil},
		{http.MethodGet, "/networks/{network}/epochs/{epoch}", hs.PublicGetNetworkEpoch, nil},

		{http.MethodGet, "/networks/{network}/blocks", hs.PublicGetNetworkBlocks, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}", hs.PublicGetNetworkBlock, nil},
		{http.MethodGet, "/networks/{network}/slots", hs.PublicGetNetworkSlots, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}", hs.PublicGetNetworkSlot, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/blocks", hs.PublicGetNetworkValidatorBlocks, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/priority-fee-blocks", hs.PublicGetNetworkAddressPriorityFeeBlocks, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/proposer-reward-blocks", hs.PublicGetNetworkAddressProposerRewardBlocks, nil},
		{http.MethodGet, "/networks/{network}/forked-blocks", hs.PublicGetNetworkForkedBlocks, nil},
		{http.MethodGet, "/networks/{network}/forked-blocks/{block}", hs.PublicGetNetworkForkedBlock, nil},
		{http.MethodGet, "/networks/{network}/forked-slots/{slot}", hs.PublicGetNetworkForkedSlot, nil},
		{http.MethodGet, "/networks/{network}/block-sizes", hs.PublicGetNetworkBlockSizes, nil},

		{http.MethodGet, "/networks/{network}/validators/{validator}/attestations", hs.PublicGetNetworkValidatorAttestations, nil},
		{http.MethodGet, "/networks/{network}/epochs/{epoch}/attestations", hs.PublicGetNetworkEpochAttestations, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}/attestations", hs.PublicGetNetworkSlotAttestations, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/attestations", hs.PublicGetNetworkBlockAttestations, nil},
		{http.MethodGet, "/networks/{network}/aggregated-attestations", hs.PublicGetNetworkAggregatedAttestations, nil},

		{http.MethodGet, "/networks/{network}/ethstore/{day}", hs.PublicGetNetworkEthStore, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/reward-history", hs.PublicGetNetworkValidatorRewardHistory, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/balance-history", hs.PublicGetNetworkValidatorBalanceHistory, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/performance-history", hs.PublicGetNetworkValidatorPerformanceHistory, nil},

		{http.MethodGet, "/networks/{network}/slashings", hs.PublicGetNetworkSlashings, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/slashings", hs.PublicGetNetworkValidatorSlashings, nil},

		{http.MethodGet, "/networks/{network}/deposits", hs.PublicGetNetworkDeposits, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/deposits", hs.PublicGetNetworkValidatorDeposits, nil},
		{http.MethodGet, "/networks/{network}/transactions/{hash}/deposits", hs.PublicGetNetworkTransactionDeposits, nil},

		{http.MethodGet, "/networks/{network}/withdrawals", hs.PublicGetNetworkWithdrawals, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}/withdrawals", hs.PublicGetNetworkSlotWithdrawals, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/withdrawals", hs.PublicGetNetworkBlockWithdrawals, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/withdrawals", hs.PublicGetNetworkValidatorWithdrawals, nil},
		{http.MethodGet, "/networks/{network}/withdrawal-credentials/{credential}/withdrawals", hs.PublicGetNetworkWithdrawalCredentialWithdrawals, nil},

		{http.MethodGet, "/networks/{network}/voluntary-exits", hs.PublicGetNetworkVoluntaryExits, nil},
		{http.MethodGet, "/networks/{network}/epochs/{epoch}/voluntary-exits", hs.PublicGetNetworkEpochVoluntaryExits, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}/voluntary-exits", hs.PublicGetNetworkSlotVoluntaryExits, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/voluntary-exits", hs.PublicGetNetworkBlockVoluntaryExits, nil},

		{http.MethodGet, "/networks/{network}/addresses/{address}/balance-history", hs.PublicGetNetworkAddressBalanceHistory, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/token-supply-history", hs.PublicGetNetworkAddressTokenSupplyHistory, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/event-logs", hs.PublicGetNetworkAddressEventLogs, nil},

		{http.MethodGet, "/networks/{network}/transactions", hs.PublicGetNetworkTransactions, nil},
		{http.MethodGet, "/networks/{network}/transactions/{hash}", hs.PublicGetNetworkTransaction, nil},
		{http.MethodGet, "/networks/{network}/addresses/{address}/transactions", hs.PublicGetNetworkAddressTransactions, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}/transactions", hs.PublicGetNetworkSlotTransactions, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/transactions", hs.PublicGetNetworkBlockTransactions, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/blobs", hs.PublicGetNetworkBlockBlobs, nil},

		{http.MethodGet, "/networks/{network}/handlerService-changes", hs.PublicGetNetworkBlsChanges, nil},
		{http.MethodGet, "/networks/{network}/epochs/{epoch}/handlerService-changes", hs.PublicGetNetworkEpochBlsChanges, nil},
		{http.MethodGet, "/networks/{network}/slots/{slot}/handlerService-changes", hs.PublicGetNetworkSlotBlsChanges, nil},
		{http.MethodGet, "/networks/{network}/blocks/{block}/handlerService-changes", hs.PublicGetNetworkBlockBlsChanges, nil},
		{http.MethodGet, "/networks/{network}/validators/{validator}/handlerService-changes", hs.PublicGetNetworkValidatorBlsChanges, nil},

		{http.MethodGet, "/networks/ethereum/addresses/{address}/ens", hs.PublicGetNetworkAddressEns, nil},
		{http.MethodGet, "/networks/ethereum/ens/{ens_name}", hs.PublicGetNetworkEns, nil},

		{http.MethodGet, "/networks/{layer_2_network}/batches", hs.PublicGetNetworkBatches, nil},
		{http.MethodGet, "/networks/{layer_2_network}/layer1-to-layer2-transactions", hs.PublicGetNetworkLayer1ToLayer2Transactions, nil},
		{http.MethodGet, "/networks/{layer_2_network}/layer2-to-layer1-transactions", hs.PublicGetNetworkLayer2ToLayer1Transactions, nil},

		{http.MethodPost, "/networks/{network}/broadcasts", hs.PublicPostNetworkBroadcasts, nil},
		{http.MethodGet, "/eth-price-history", hs.PublicGetEthPriceHistory, nil},

		{http.MethodGet, "/networks/{network}/gasnow", hs.PublicGetNetworkGasNow, nil},
		{http.MethodGet, "/networks/{network}/average-gas-limit-history", hs.PublicGetNetworkAverageGasLimitHistory, nil},
		{http.MethodGet, "/networks/{network}/gas-used-history", hs.PublicGetNetworkGasUsedHistory, nil},

		{http.MethodGet, "/rocket-pool/nodes", hs.PublicGetRocketPoolNodes, nil},
		{http.MethodGet, "/rocket-pool/minipools", hs.PublicGetRocketPoolMinipools, nil},

		{http.MethodGet, "/networks/{network}/sync-committee/{period}", hs.PublicGetNetworkSyncCommittee, nil},

		{http.MethodGet, "/multisig-safes/{address}", hs.PublicGetMultisigSafe, nil},
		{http.MethodGet, "/multisig-safes/{address}/transactions", hs.PublicGetMultisigSafeTransactions, nil},
		{http.MethodGet, "/multisig-transactions/{hash}/confirmations", hs.PublicGetMultisigTransactionConfirmations, nil},
	}
	addEndpointsToRouters(endpoints, publicRouter, internalRouter)
}

func addValidatorDashboardRoutes(hs *handlers.HandlerService, publicRouter, internalRouter *mux.Router, cfg *types.Config) {
	vdbPath := "/validator-dashboards"
	publicRouter.HandleFunc(vdbPath, hs.PublicPostValidatorDashboards).Methods(http.MethodPost, http.MethodOptions)
	internalRouter.HandleFunc(vdbPath, hs.InternalPostValidatorDashboards).Methods(http.MethodPost, http.MethodOptions)

	publicDashboardRouter := publicRouter.PathPrefix(vdbPath).Subrouter()
	internalDashboardRouter := internalRouter.PathPrefix(vdbPath).Subrouter()
	// add middleware to check if user has access to dashboard
	if !cfg.Frontend.Debug {
		publicDashboardRouter.Use(hs.GetVDBAuthMiddleware(hs.GetUserIdByApiKey))
		internalDashboardRouter.Use(hs.GetVDBAuthMiddleware(hs.GetUserIdBySession), GetAuthMiddleware(cfg.ApiKeySecret))
	}

	endpoints := []endpoint{
		{http.MethodGet, "/{dashboard_id}", hs.PublicGetValidatorDashboard, hs.InternalGetValidatorDashboard},
		{http.MethodDelete, "/{dashboard_id}", hs.PublicDeleteValidatorDashboard, hs.InternalDeleteValidatorDashboard},
		{http.MethodPut, "/{dashboard_id}/name", nil, hs.InternalPutValidatorDashboardName},
		{http.MethodPost, "/{dashboard_id}/groups", hs.PublicPostValidatorDashboardGroups, hs.InternalPostValidatorDashboardGroups},
		{http.MethodPut, "/{dashboard_id}/groups/{group_id}", hs.PublicPutValidatorDashboardGroups, hs.InternalPutValidatorDashboardGroups},
		{http.MethodDelete, "/{dashboard_id}/groups/{group_id}", hs.PublicDeleteValidatorDashboardGroups, hs.InternalDeleteValidatorDashboardGroups},
		{http.MethodPost, "/{dashboard_id}/validators", hs.PublicPostValidatorDashboardValidators, hs.InternalPostValidatorDashboardValidators},
		{http.MethodGet, "/{dashboard_id}/validators", hs.PublicGetValidatorDashboardValidators, hs.InternalGetValidatorDashboardValidators},
		{http.MethodDelete, "/{dashboard_id}/validators", hs.PublicDeleteValidatorDashboardValidators, hs.InternalDeleteValidatorDashboardValidators},
		{http.MethodPost, "/{dashboard_id}/public-ids", hs.PublicPostValidatorDashboardPublicIds, hs.InternalPostValidatorDashboardPublicIds},
		{http.MethodPut, "/{dashboard_id}/public-ids/{public_id}", hs.PublicPutValidatorDashboardPublicId, hs.InternalPutValidatorDashboardPublicId},
		{http.MethodDelete, "/{dashboard_id}/public-ids/{public_id}", hs.PublicDeleteValidatorDashboardPublicId, hs.InternalDeleteValidatorDashboardPublicId},
		{http.MethodGet, "/{dashboard_id}/slot-viz", hs.PublicGetValidatorDashboardSlotViz, hs.InternalGetValidatorDashboardSlotViz},
		{http.MethodGet, "/{dashboard_id}/summary", hs.PublicGetValidatorDashboardSummary, hs.InternalGetValidatorDashboardSummary},
		{http.MethodGet, "/{dashboard_id}/validator-indices", nil, hs.InternalGetValidatorDashboardValidatorIndices},
		{http.MethodGet, "/{dashboard_id}/groups/{group_id}/summary", hs.PublicGetValidatorDashboardGroupSummary, hs.InternalGetValidatorDashboardGroupSummary},
		{http.MethodGet, "/{dashboard_id}/summary-chart", hs.PublicGetValidatorDashboardSummaryChart, hs.InternalGetValidatorDashboardSummaryChart},
		{http.MethodGet, "/{dashboard_id}/rewards", hs.PublicGetValidatorDashboardRewards, hs.InternalGetValidatorDashboardRewards},
		{http.MethodGet, "/{dashboard_id}/groups/{group_id}/rewards/{epoch}", hs.PublicGetValidatorDashboardGroupRewards, hs.InternalGetValidatorDashboardGroupRewards},
		{http.MethodGet, "/{dashboard_id}/rewards-chart", hs.PublicGetValidatorDashboardRewardsChart, hs.InternalGetValidatorDashboardRewardsChart},
		{http.MethodGet, "/{dashboard_id}/duties/{epoch}", hs.PublicGetValidatorDashboardDuties, hs.InternalGetValidatorDashboardDuties},
		{http.MethodGet, "/{dashboard_id}/blocks", hs.PublicGetValidatorDashboardBlocks, hs.InternalGetValidatorDashboardBlocks},
		{http.MethodGet, "/{dashboard_id}/epoch-heatmap", hs.PublicGetValidatorDashboardEpochHeatmap, hs.InternalGetValidatorDashboardEpochHeatmap},
		{http.MethodGet, "/{dashboard_id}/daily-heatmap", hs.PublicGetValidatorDashboardDailyHeatmap, hs.InternalGetValidatorDashboardDailyHeatmap},
		{http.MethodGet, "/{dashboard_id}/groups/{group_id}/epoch-heatmap/{epoch}", hs.PublicGetValidatorDashboardGroupEpochHeatmap, hs.InternalGetValidatorDashboardGroupEpochHeatmap},
		{http.MethodGet, "/{dashboard_id}/groups/{group_id}/daily-heatmap/{date}", hs.PublicGetValidatorDashboardGroupDailyHeatmap, hs.InternalGetValidatorDashboardGroupDailyHeatmap},
		{http.MethodGet, "/{dashboard_id}/execution-layer-deposits", hs.PublicGetValidatorDashboardExecutionLayerDeposits, hs.InternalGetValidatorDashboardExecutionLayerDeposits},
		{http.MethodGet, "/{dashboard_id}/consensus-layer-deposits", hs.PublicGetValidatorDashboardConsensusLayerDeposits, hs.InternalGetValidatorDashboardConsensusLayerDeposits},
		{http.MethodGet, "/{dashboard_id}/total-execution-layer-deposits", nil, hs.InternalGetValidatorDashboardTotalExecutionLayerDeposits},
		{http.MethodGet, "/{dashboard_id}/total-consensus-layer-deposits", nil, hs.InternalGetValidatorDashboardTotalConsensusLayerDeposits},
		{http.MethodGet, "/{dashboard_id}/withdrawals", hs.PublicGetValidatorDashboardWithdrawals, hs.InternalGetValidatorDashboardWithdrawals},
		{http.MethodGet, "/{dashboard_id}/total-withdrawals", nil, hs.InternalGetValidatorDashboardTotalWithdrawals},
	}
	addEndpointsToRouters(endpoints, publicDashboardRouter, internalDashboardRouter)
}

func addEndpointsToRouters(endpoints []endpoint, publicRouter *mux.Router, internalRouter *mux.Router) {
	for _, endpoint := range endpoints {
		if endpoint.PublicHandler != nil {
			publicRouter.HandleFunc(endpoint.Path, endpoint.PublicHandler).Methods(endpoint.Method, http.MethodOptions)
		}
		if endpoint.InternalHander != nil {
			internalRouter.HandleFunc(endpoint.Path, endpoint.InternalHander).Methods(endpoint.Method, http.MethodOptions)
		}
	}
}
