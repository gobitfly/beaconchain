package app

import (
	"context"
	"net/http"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

func (service *ApiService) GetSlot(ctx context.Context, in model.GetSlotRequestObject) (model.GetSlotResponseObject, error) {
	chain, err := io.AsChain(in.Body.Chain)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid chain parameter")
	}

	resolvedSlot, err := io.ResolveSlot(ctx, service.ethereumNetworkRepo, chain, in.Slot)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusBadRequest, "Invalid slot parameter")
	}

	slot, err := service.ethereumNetworkRepo.GetSlot(ctx, chain, resolvedSlot)
	if err != nil {
		return nil, common.NewAPIUserFacingError(http.StatusInternalServerError, "Failed to get slot data")
	}

	transformedSlot := transformSlotToModel(slot)
	response := model.GetSlot200JSONResponse{}
	response.Data = &transformedSlot
	return response, nil
}
func transformSlotToModel(slot domain.Slot) model.SlotOverviewData {
	return model.SlotOverviewData{
		AttestationCount: slot.AttestationCount,
		// TODO: Missing ProposerSlashingCount in spec
		BlockRoot: model.ConsensusLayerBlockRoot(slot.BlockRoot),
		// expected spec-change:
		// DataRecency: model.DataRecency{
		// 	IsEpochFinalized: slot.Finalized,
		// },
		Graffiti: model.Graffiti(slot.Graffiti),
		Proposer: model.Validator{
			Index:     slot.Proposer.Index,
			PublicKey: model.ValidatorPublicKey(slot.Proposer.Pubkey),
		},
		Status: model.DutyStatus(slot.Status),
		// TODO: convert slot to time for a given chain
		// TODO: events
	}
}

func (service *ApiService) GetNetworkConfig(ctx context.Context, in model.GetNetworkConfigRequestObject) (model.GetNetworkConfigResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetPerformanceSummary(ctx context.Context, in model.GetPerformanceSummaryRequestObject) (model.GetPerformanceSummaryResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetQueueStats(ctx context.Context, in model.GetQueueStatsRequestObject) (model.GetQueueStatsResponseObject, error) {
	return nil, nil
}

func (service *ApiService) GetSyncCommitteeDutiesForSlot(ctx context.Context, in model.GetSyncCommitteeDutiesForSlotRequestObject) (model.GetSyncCommitteeDutiesForSlotResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetChainState(ctx context.Context, in model.GetChainStateRequestObject) (model.GetChainStateResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSyncCommitteeForPeriod(ctx context.Context, in model.GetSyncCommitteeForPeriodRequestObject) (model.GetSyncCommitteeForPeriodResponseObject, error) {
	return nil, nil
}
func (service *ApiService) GetSyncCommitteeValidatorsForPeriod(ctx context.Context, in model.GetSyncCommitteeValidatorsForPeriodRequestObject) (model.GetSyncCommitteeValidatorsForPeriodResponseObject, error) {
	return nil, nil
}
