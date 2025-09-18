package app

import (
	"context"
	"net/http"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

func (service *ApiService) GetV2EthereumSlotsSlot(ctx context.Context, in model.GetV2EthereumSlotsSlotRequestObject) (model.GetV2EthereumSlotsSlotResponseObject, error) {
	chain, err := io.AsChain(in.Params.Chain)
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
}
