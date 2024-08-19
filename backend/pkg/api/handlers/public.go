package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

type StatusReports struct {
	StatusReports []StatusReport `json:"status_reports"`
	IsOk          bool           `json:"is_ok"`
}

type StatusReport struct {
	IsOk        bool     `json:"is_ok"`
	Tags        []string `json:"tags"`
	Id          string   `json:"id"`
	Description string   `json:"description"`
}

// All handler function names must include the HTTP method and the path they handle
// Public handlers may only be authenticated by an API key
// Public handlers must never call internal handlers

func (h *HandlerService) PublicGetHealthz(w http.ResponseWriter, r *http.Request) {
	s := StatusReports{
		IsOk: true,
	}
	// errgroup
	ch := make(chan StatusReport, 100)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		sr := StatusReport{
			Tags: []string{"clickhouse", "database"},
			Id:   "ch_connection",
		}
		var version string
		err := db.ClickHouseReader.GetContext(ctx, &version, "SELECT version()")
		if err != nil {
			sr.IsOk = false
			sr.Description = "clickhouse is not healthy"
			ch <- sr
			return nil
		}
		sr.IsOk = true
		sr.Description = "clickhouse is healthy, version: " + version
		ch <- sr
		return nil
	})

	g.Go(func() error {
		sr := StatusReport{
			Tags: []string{"clickhouse", "exporter", "validator_dashboard"},
			Id:   "ch_latest_epoch",
		}
		var t time.Time
		err := db.ClickHouseReader.GetContext(ctx, &t, "SELECT MAX(epoch_timestamp) FROM validator_dashboard_data_epoch")
		if err != nil {
			sr.IsOk = false
			sr.Description = "failed to get latest epoch from clickhouse: " + err.Error()
			ch <- sr
			return nil
		}
		if time.Since(t) > 1*time.Hour {
			sr.IsOk = false
			sr.Description = fmt.Sprintf("latest exported epoch is older than 1 hour: %s", time.Since(t))
			ch <- sr
			return nil
		}
		sr.IsOk = true
		sr.Description = fmt.Sprintf("latest exported epoch is %s old", time.Since(t))
		ch <- sr
		return nil
	})

	target_rollings := []string{"1h", "24h", "7d", "30d", "90d", "total"}
	for _, target_rolling := range target_rollings {
		target_rolling := target_rolling
		g.Go(func() error {
			sr := StatusReport{
				Tags: []string{"clickhouse", "exporter", "validator_dashboard", target_rolling},
				Id:   "ch_rolling_rolling_" + target_rolling,
			}
			var delta int
			err := db.ClickHouseReader.GetContext(ctx, &delta, fmt.Sprintf(`
				SELECT
				    coalesce((
				        SELECT
				            max(epoch)
				        FROM holesky.validator_dashboard_data_epoch
				        WHERE
				            epoch_timestamp = (
				                SELECT
				                    max(epoch_timestamp)
				                FROM holesky.validator_dashboard_data_epoch)) - MAX(epoch_end), 255) AS delta
				FROM
				    holesky.validator_dashboard_data_rolling_%s
				WHERE
				    validator_index = 0`, target_rolling))
			if err != nil {
				sr.IsOk = false
				sr.Description = fmt.Sprintf("failed to get epoch delta from clickhouse for rolling %s: %s", target_rolling, err.Error())
				ch <- sr
				return nil
			}
			threshold := 2
			if delta > threshold {
				sr.IsOk = false
				sr.Description = fmt.Sprintf("epoch delta for rolling %s is %d, threshold is %d", target_rolling, delta, threshold)
				ch <- sr
				return nil
			}
			sr.IsOk = true
			sr.Description = fmt.Sprintf("epoch delta for rolling %s is %d", target_rolling, delta)
			ch <- sr
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		returnInternalServerError(w, err)
	}
	close(ch)

	for report := range ch {
		s.StatusReports = append(s.StatusReports, report)
		if !report.IsOk && s.IsOk {
			s.IsOk = false
		}
	}
	if s.IsOk {
		returnOk(w, s)
	} else {
		writeResponse(w, http.StatusInternalServerError, s)
	}
}

func (h *HandlerService) PublicGetHealthzLoadbalancer(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetUserDashboards(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiDataResponse[types.UserDashboardsData]{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) PublicPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPutValidatorDashboardArchiving(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicPutValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicDeleteValidatorDashboardGroup(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		GroupId           uint64        `json:"group_id,omitempty"`
		Validators        []intOrString `json:"validators,omitempty"`
		DepositAddress    string        `json:"deposit_address,omitempty"`
		WithdrawalAddress string        `json:"withdrawal_address,omitempty"`
		Graffiti          string        `json:"graffiti,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// check if exactly one of validators, deposit_address, withdrawal_address, graffiti is set
	fields := []interface{}{req.Validators, req.DepositAddress, req.WithdrawalAddress, req.Graffiti}
	var count int
	for _, set := range fields {
		if !reflect.ValueOf(set).IsZero() {
			count++
		}
	}
	if count != 1 {
		v.add("request body", "exactly one of `validators`, `deposit_address`, `withdrawal_address`, `graffiti` must be set. please check the API documentation for more information")
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	ctx := r.Context()
	groupId := req.GroupId
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(ctx, dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(ctx, userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	limit := userInfo.PremiumPerks.ValidatorsPerDashboard
	if req.Validators == nil && !userInfo.PremiumPerks.BulkAdding && !isUserAdmin(userInfo) {
		returnConflict(w, errors.New("bulk adding not allowed with current subscription plan"))
		return
	}
	var data []types.VDBPostValidatorsData
	var dataErr error
	switch {
	case req.Validators != nil:
		indices, pubkeys := v.checkValidators(req.Validators, forbidEmpty)
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		validators, err := h.dai.GetValidatorsFromSlices(indices, pubkeys)
		if err != nil {
			handleErr(w, err)
			return
		}
		// check if adding more validators than allowed
		existingValidatorCount, err := h.dai.GetValidatorDashboardExistingValidatorCount(ctx, dashboardId, validators)
		if err != nil {
			handleErr(w, err)
			return
		}
		if uint64(len(validators)) > existingValidatorCount+limit {
			returnConflict(w, fmt.Errorf("adding more validators than allowed, limit is %v new validators", limit))
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidators(ctx, dashboardId, groupId, validators)

	case req.DepositAddress != "":
		depositAddress := v.checkRegex(reEthereumAddress, req.DepositAddress, "deposit_address")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByDepositAddress(ctx, dashboardId, groupId, depositAddress, limit)

	case req.WithdrawalAddress != "":
		withdrawalAddress := v.checkRegex(reWithdrawalCredential, req.WithdrawalAddress, "withdrawal_address")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByWithdrawalAddress(ctx, dashboardId, groupId, withdrawalAddress, limit)

	case req.Graffiti != "":
		graffiti := v.checkRegex(reNonEmpty, req.Graffiti, "graffiti")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByGraffiti(ctx, dashboardId, groupId, graffiti, limit)
	}

	if dataErr != nil {
		handleErr(w, dataErr)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) PublicGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	var indices []uint64
	var publicKeys []string
	if validatorsParam := r.URL.Query().Get("validators"); validatorsParam != "" {
		indices, publicKeys = v.checkValidatorList(validatorsParam, allowEmpty)
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, publicKeys)
	if err != nil {
		handleErr(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardValidators(r.Context(), dashboardId, validators)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) PublicPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) PublicGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardRocketPool(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardTotalRocketPool(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardNodeRocketPool(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetValidatorDashboardRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidator(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorDuties(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawalCredentialValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorStatuses(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorLeaderboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorQueue(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEpochs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEpoch(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlock(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlots(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlot(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressPriorityFeeBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressProposerRewardBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkForkedBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkForkedBlock(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkForkedSlot(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockSizes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEpochAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotVotes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockVotes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAggregatedAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEthStore(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorRewardHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorPerformanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlashings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorSlashings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkTransactionDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}
func (h *HandlerService) PublicGetNetworkBlockWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawalCredentialWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEpochVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressBalanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressTokenSupplyHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressEventLogs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkTransaction(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockBlobs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEpochBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSlotBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBlockBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAddressEns(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkEns(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkBatches(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkLayer2ToLayer1Transactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkLayer1ToLayer2Transactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicPostNetworkBroadcasts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) PublicGetEthPriceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkGasNow(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkAverageGasLimitHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkGasUsedHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetRocketPool(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetRocketPoolNodes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetNetworkSyncCommittee(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetMultisigSafe(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetMultisigSafeTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) PublicGetMultisigTransactionConfirmations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}
