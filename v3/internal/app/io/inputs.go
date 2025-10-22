package io

import (
	"context"
	"errors"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var (
	ErrInvalidParam = errors.New("invalid parameter")
)

// ResolveEpoch returns an epoch number either from a named view ("head"/"finalized")
// or directly from a numeric string.
func ResolveSlot(
	ctx context.Context,
	repo ethereumnetworkrepo.LatestStateRepository,
	chain domain.Chain,
	input model.SlotParam,
) (int, error) {
	intParam, err := input.AsSlot()
	if err == nil {
		return intParam, nil
	}

	stringParam, err := input.AsChainView()
	if err != nil {
		return 0, ErrInvalidParam
	}

	view, ok := parseView(stringParam)
	if !ok {
		return 0, ErrInvalidParam
	}

	state, err := repo.GetLatestState(ctx, chain, view)
	if err != nil {
		return 0, err
	}
	return state.Epoch, nil
}

// ResolveEpoch returns an epoch number either from a named view ("head"/"finalized")
// or directly from a numeric string.
func ResolveEpoch(
	ctx context.Context,
	repo ethereumnetworkrepo.LatestStateRepository,
	chain domain.Chain,
	input model.EpochParam,
) (int, error) {
	intParam, err := input.AsEpoch()
	if err == nil {
		return intParam, nil
	}

	stringParam, err := input.AsChainView()
	if err != nil {
		return 0, ErrInvalidParam
	}

	view, ok := parseView(stringParam)
	if !ok {
		return 0, ErrInvalidParam
	}

	state, err := repo.GetLatestState(ctx, chain, view)
	if err != nil {
		return 0, err
	}
	return state.Epoch, nil
}

func AsChain(chain model.Chain) (domain.Chain, error) {
	switch chain {
	case model.Mainnet, model.Chain(""):
		return domain.ChainMainnet, nil
	case model.Hoodi:
		return domain.ChainHoodi, nil
	default:
		return domain.ChainUnknown, fmt.Errorf("unsupported chain: %s", chain)
	}
}

func parseView(input model.ChainView) (domain.ConsensusView, bool) {
	switch input {
	case model.ChainViewLatest:
		return domain.ConsensusViewHead, true
	case model.ChainViewFinalized:
		return domain.ConsensusViewFinalized, true
	default:
		return domain.ConsensusViewHead, false
	}
}
