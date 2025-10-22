package app

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
)

/* func (service *ApiService) GetV2EthereumSlotsSlot(ctx context.Context, in model.GetV2EthereumSlotsSlotRequestObject) (model.GetV2EthereumSlotsSlotResponseObject, error) {
	chain, err := io.AsChain(in.Params.Chain)
	model.GetNetworkConfigRequestObject
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid chain parameter")
	}

	resolvedSlot, err := io.ResolveSlot(ctx, service.ethereumNetworkRepo, chain, in.Slot)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid slot parameter")
	}

	slotData, err := service.ethereumNetworkRepo.GetSlot(ctx, chain, resolvedSlot)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusInternalServerError, "Failed to get slot data")
	}

	return model.GetV2EthereumSlotsSlot200JSONResponse(model.GetSlotResponse{Data: transformSlotToModel(slotData)}), nil
}

func transformSlotToModel(slot *domain.Slot) model.GetSlotResponseData {
	return model.GetSlotResponseData{
		AttestationCount: slot.AttestationCount,
		// TODO: Missing ProposerSlashingCount in spec
		AttestationSlashingCount: slot.AttestationSlashingCount,
		BlockRoot:                slot.BlockRoot,
		// expected spec-change:
		// DataRecency: model.DataRecency{
		// 	IsEpochFinalized: slot.Finalized,
		// },
		Graffiti: slot.Graffiti,
		Proposer: model.Validator{
			Index:  slot.Proposer.Index,
			Pubkey: model.ValidatorPubkey(slot.Proposer.Pubkey),
		},
		Status: model.DutyStatus(slot.Status),
		// TODO: convert slot to time for a given chain
		// TODO: events
	}
} */

func (service *ApiService) GetNetworkConfig(ctx context.Context, request model.GetNetworkConfigRequestObject) (model.GetNetworkConfigResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetPerformanceSummary(ctx context.Context, request model.GetPerformanceSummaryRequestObject) (model.GetPerformanceSummaryResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetQueueStats(ctx context.Context, request model.GetQueueStatsRequestObject) (model.GetQueueStatsResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSlot(ctx context.Context, request model.GetSlotRequestObject) (model.GetSlotResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSyncCommitteeDutiesForSlot(ctx context.Context, request model.GetSyncCommitteeDutiesForSlotRequestObject) (model.GetSyncCommitteeDutiesForSlotResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetChainState(ctx context.Context, request model.GetChainStateRequestObject) (model.GetChainStateResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSyncCommitteeForPeriod(ctx context.Context, request model.GetSyncCommitteeForPeriodRequestObject) (model.GetSyncCommitteeForPeriodResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSyncCommitteeValidatorsForPeriod(ctx context.Context, request model.GetSyncCommitteeValidatorsForPeriodRequestObject) (model.GetSyncCommitteeValidatorsForPeriodResponseObject, error) {
	return nil, nil
}
