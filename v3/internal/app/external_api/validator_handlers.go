package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/common/islices"
	"github.com/gobitfly/beaconchain-backend/internal/common/pagination"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/shopspring/decimal"
)

// -----------------------------
// POST v2/ethereum/validators
func (service *ApiService) GetValidatorOverview(ctx context.Context, in model.GetValidatorOverviewRequestObject) (model.GetValidatorOverviewResponseObject, error) {
	chain, err := io.AsChain(in.Body.Chain)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid chain parameter")
	}
	validatorsSelector, err := io.AsValidatorsSelector(in.Body.Validator)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid validator selector")
	}
	overviews, pagination, err := pagination.Handle(
		in.Body.Cursor,
		in.Body.PageSize,
		transformOverviewToCursor,
		func(cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorOverview, error) {
			return service.validatorRepository.GetOverview(ctx, chain, validatorsSelector, cursor, pageSize)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator overviews: %w", err)
	}
	if len(overviews) == 0 {
		return nil, common.NewAPIUserFacingError(http.StatusNotFound, "no validators found")
	}

	// need to fetch balances separately with indexes and head epoch
	latestState, err := service.ethereumNetworkRepo.GetLatestState(ctx, chain, domain.ConsensusViewHead)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest state: %w", err)
	}
	indices := islices.Transform(overviews, func(o domain.ValidatorOverview) domain.ValidatorIndex { return o.ValidatorIndex }) // extract indices from overviews
	balances, err := service.validatorRepository.GetHeadBalances(ctx, chain, io.EpochToStartTime(latestState.Epoch, service.chainConfigs[chain]), indices)
	if err != nil {
		return nil, fmt.Errorf("failed to get head balances for overview: %w", err)
	}

	// transform to map for easy lookup
	balancesMap := islices.ToMap(balances, func(b domain.ValidatorBalance) domain.ValidatorIndex { return b.ValidatorIndex }) // key is validator index
	data := make([]model.ValidatorOverviewData, len(overviews))
	for i, overview := range overviews {
		balance, ok := balancesMap[overview.ValidatorIndex]
		if !ok {
			return nil, fmt.Errorf("missing balance for validator index %d", overview.ValidatorIndex)
		}
		data[i], err = transformOverviewToModel(overview, balance, latestState.Epoch)
		if err != nil {
			return nil, fmt.Errorf("failed to transform overview for validator %d to model: %w", overview.ValidatorIndex, err)
		}
	}
	response := model.GetValidatorOverview200JSONResponse{}
	response.Data = data
	response.Paging = pagination
	response.Range = io.EpochToResultRange(latestState.Epoch, service.chainConfigs[chain])
	return response, nil
}
func transformOverviewToCursor(overview domain.ValidatorOverview) domain.ValidatorIndexCursor {
	return domain.ValidatorIndexCursor{
		Index: overview.ValidatorIndex,
	}
}

func transformOverviewToModel(overview domain.ValidatorOverview, balance domain.ValidatorBalance, latestEpoch int) (model.ValidatorOverviewData, error) {
	withdrawalCredential, err := io.AsWithdrawalCredential(overview.WithdrawalCredential)
	if err != nil {
		return model.ValidatorOverviewData{}, err
	}
	return model.ValidatorOverviewData{
		Validator: io.AsModelValidator(overview.ValidatorIndex, overview.ValidatorPublicKey),
		Finality:  io.OnlyNotFinalized,
		Balances: model.ValidatorBalancesHead{
			Current:   balance.CurrentBalance.String(),
			Effective: balance.EffectiveBalance.String(),
		},
		LifeCycleEpochs: model.LifeCycleEpochs{
			ActivationEligibility: io.AsNullable(overview.ActivationEligibilityEpoch),
			Activation:            io.AsNullable(overview.ActivationEpoch),
			Exit:                  io.AsNullable(overview.ExitEpoch),
			Withdrawable:          io.AsNullable(overview.WithdrawableEpoch),
		},
		Online:                io.AsNullable(overview.Online),
		Slashed:               overview.Slashed,
		WithdrawalCredentials: withdrawalCredential,
		Status: validatorStatus(
			latestEpoch,
			overview.ActivationEligibilityEpoch,
			overview.ActivationEpoch,
			overview.ExitEpoch,
			overview.WithdrawableEpoch,
			overview.Slashed,
			balance.CurrentBalance,
		),
	}, nil
}

// validatorStatus determines the validator status based on lifecycle epochs, slashed status, and balance.
// taken from https://hackmd.io/ofFJ5gOmQpu1jjHilHbdQQ, FAR_FUTURE_EPOCH is represented as nil here
func validatorStatus(
	latestEpoch int,
	activationEligibilityEpoch *int,
	activationEpoch *int,
	exitEpoch *int,
	withdrawableEpoch *int,
	slashed bool,
	balance decimal.Decimal,
) model.DataStatus {

	// pending
	if activationEligibilityEpoch == nil {
		return model.PendingInitialized
	}
	if activationEpoch == nil || *activationEpoch > latestEpoch {
		return model.PendingQueued
	}

	// active
	if exitEpoch == nil {
		return model.ActiveOngoing
	}
	if latestEpoch < *exitEpoch {
		if slashed {
			return model.ActiveSlashed
		}
		return model.ActiveExiting
	}

	// withdrawal
	if withdrawableEpoch != nil && *withdrawableEpoch <= latestEpoch {
		if balance.IsZero() {
			return model.WithdrawalDone
		}
		return model.WithdrawalPossible
	}

	// exited
	if slashed {
		return model.ExitedSlashed
	}
	return model.ExitedUnslashed
}

// -----------------------------
func (service *ApiService) GetValidatorApyRoi(ctx context.Context, request model.GetValidatorApyRoiRequestObject) (model.GetValidatorApyRoiResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetValidatorAttestations(ctx context.Context, request model.GetValidatorAttestationsRequestObject) (model.GetValidatorAttestationsResponseObject, error) {
	return nil, nil
}

// -----------------------------
// POST v2/ethereum/validators/balances
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
		return service.validatorRepository.GetBalances(ctx, chain, io.EpochToStartTime(requestedEpoch, service.chainConfigs[chain]), validatorsSelector, cursor, pageSize)
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
	response.Data = data
	response.Paging = paging
	response.Range = resultRange
	return response, nil
}

func transformValidatorBalancesToCursor(balance domain.ValidatorBalance) domain.ValidatorIndexCursor {
	return domain.ValidatorIndexCursor{
		Index: balance.ValidatorIndex,
	}
}

func transformValidatorBalanceToModel(balance domain.ValidatorBalance, finality model.FinalityParams) model.ValidatorBalancesData {
	return model.ValidatorBalancesData{
		Validator: io.AsModelValidator(balance.ValidatorIndex, balance.ValidatorPublicKey),
		Balance: model.ValidatorBalance{
			Current:   balance.CurrentBalance.String(),
			Effective: balance.EffectiveBalance.String(),
		},
		Finality: finality,
	}
}

// -----------------------------
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
