package app

import (
	"context"
	"net/http"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/common/pagination"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

func (service *ApiService) GetValidatorOverview(ctx context.Context, request model.GetValidatorOverviewRequestObject) (model.GetValidatorOverviewResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorApyRoi(ctx context.Context, request model.GetValidatorApyRoiRequestObject) (model.GetValidatorApyRoiResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorAttestations(ctx context.Context, request model.GetValidatorAttestationsRequestObject) (model.GetValidatorAttestationsResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorBalances(ctx context.Context, in model.GetValidatorBalancesRequestObject) (model.GetValidatorBalancesResponseObject, error) {
	chain, err := io.AsChain(in.Body.Chain)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid chain parameter")
	}
	validatorsSelector, err := io.AsValidatorsSelector(in.Body.Validator)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid validator selector")
	}
	requestedEpoch, err := io.ResolveEpoch(ctx, service.ethereumNetworkRepo, chain, in.Body.Epoch)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid epoch parameter")
	}

	finalizedState, err := service.ethereumNetworkRepo.GetLatestState(ctx, chain, domain.ConsensusViewFinalized)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusInternalServerError, "Failed to get latest state")
	}

	fetch := func(cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorBalance, error) {
		return service.validatorRepository.GetBalances(ctx, chain, io.EpochToStartTimestamp(requestedEpoch, service.chainConfigs[chain]), validatorsSelector, cursor, pageSize)
	}
	balances, paging, err := pagination.Handle(
		in.Body.Cursor,
		in.Body.PageSize,
		transformValidatorBalancesToCursor,
		fetch,
	)
	if err != nil {
		return nil, err
	}
	data := make([]model.ValidatorBalancesData, len(balances))
	finality := io.FormatFinalized(requestedEpoch, finalizedState.Epoch)
	for i := range balances {
		data[i] = transformValidatorBalanceToModel(balances[i], finality)
	}
	resultRange := io.EpochToResultRange(requestedEpoch, service.chainConfigs[chain])
	response := model.GetValidatorBalances200JSONResponse{}
	response.Data = &data
	response.Paging = &paging
	response.Range = &resultRange
	return response, nil
}

func transformValidatorBalancesToCursor(balance domain.ValidatorBalance) domain.ValidatorIndexCursor {
	return domain.ValidatorIndexCursor{
		Index: balance.ValidatorIndex,
	}
}

func transformValidatorBalanceToModel(balance domain.ValidatorBalance, finality model.FinalityParams) model.ValidatorBalancesData {
	return model.ValidatorBalancesData{
		Validator: model.Validator{
			Index:     balance.ValidatorIndex,
			PublicKey: model.ValidatorPublicKey(balance.ValidatorPublicKey),
		},
		Balance: model.ValidatorBalance{
			Current:   balance.CurrentBalance.String(),
			Effective: balance.EffectiveBalance.String(),
		},
		Finality: finality,
	}
}
func (service *ApiService) GetValidatorDeposits(ctx context.Context, request model.GetValidatorDepositsRequestObject) (model.GetValidatorDepositsResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorPerformanceDetailed(ctx context.Context, request model.GetValidatorPerformanceDetailedRequestObject) (model.GetValidatorPerformanceDetailedResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorPerformanceSummary(ctx context.Context, request model.GetValidatorPerformanceSummaryRequestObject) (model.GetValidatorPerformanceSummaryResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorProposals(ctx context.Context, request model.GetValidatorProposalsRequestObject) (model.GetValidatorProposalsResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorQueueStats(ctx context.Context, request model.GetValidatorQueueStatsRequestObject) (model.GetValidatorQueueStatsResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorRewardsDetailed(ctx context.Context, request model.GetValidatorRewardsDetailedRequestObject) (model.GetValidatorRewardsDetailedResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorRewardsSummary(ctx context.Context, request model.GetValidatorRewardsSummaryRequestObject) (model.GetValidatorRewardsSummaryResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorSyncCommittees(ctx context.Context, request model.GetValidatorSyncCommitteesRequestObject) (model.GetValidatorSyncCommitteesResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorUpcomingDuties(ctx context.Context, request model.GetValidatorUpcomingDutiesRequestObject) (model.GetValidatorUpcomingDutiesResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorWithdrawals(ctx context.Context, request model.GetValidatorWithdrawalsRequestObject) (model.GetValidatorWithdrawalsResponseObject, error) {
	return nil, nil
}
