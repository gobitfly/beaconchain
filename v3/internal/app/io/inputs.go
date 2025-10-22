package io

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var (
	ErrInvalidParam = errors.New("invalid parameter")
)

const (
	InputSlotHead      = "head"
	InputSlotFinalized = "finalized"
)

// ResolveSlot returns a slot number either from a named view ("head"/"finalized")
// or directly from a numeric string.
func ResolveSlot(
	ctx context.Context,
	repo ethereumnetworkrepo.LatestStateRepository,
	chain domain.Chain,
	input string,
) (int, error) {
	return resolveParam(ctx, repo, chain, input, func(state domain.LatestState) int {
		return state.Slot
	})
}

// ResolveEpoch returns an epoch number either from a named view ("head"/"finalized")
// or directly from a numeric string.
func ResolveEpoch(
	ctx context.Context,
	repo ethereumnetworkrepo.LatestStateRepository,
	chain domain.Chain,
	input string,
) (int, error) {
	return resolveParam(ctx, repo, chain, input, func(state domain.LatestState) int {
		return state.Epoch
	})
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

// --- helpers ---

func resolveParam(
	ctx context.Context,
	repo ethereumnetworkrepo.LatestStateRepository,
	chain domain.Chain,
	input string,
	extract func(domain.LatestState) int,
) (int, error) {
	if view, ok := parseView(input); ok {
		state, err := repo.GetLatestState(ctx, chain, view)
		if err != nil {
			return 0, err
		}
		return extract(state), nil
	}

	val, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return 0, errors.Join(ErrInvalidParam, err)
	}
	return int(val), nil
}

func parseView(input string) (domain.ConsensusView, bool) {
	switch input {
	case InputSlotHead:
		return domain.ConsensusViewHead, true
	case InputSlotFinalized:
		return domain.ConsensusViewFinalized, true
	default:
		return domain.ConsensusViewHead, false
	}
}
