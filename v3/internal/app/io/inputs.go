package io

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var (
	ErrInvalidParam = errors.New("invalid parameter")
)

// ResolveSlot returns a slot number either from a named view ("head"/"finalized")
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
		return "", false
	}
}

// AsValidatorsSelector converts a model.ValidatorsSelector to a domain.ValidatorsSelector.
func AsValidatorsSelector(input model.ValidatorsSelector) (domain.ValidatorsSelector, error) {
	selector, err := asIdentifiers(input)
	if err == nil {
		return selector, nil
	}

	selector, err = asDepositAddress(input)
	if err == nil {
		return selector, nil
	}

	return asWithdrawalInput(input)
}

var errEmptySelector = errors.New("empty validator selector")

// asDepositAddress tries to parse the input as a deposit address selector.
func asDepositAddress(input model.ValidatorsSelector) (domain.ValidatorsSelector, error) {
	selector := domain.ValidatorsSelector{}
	byDeposit, err := input.AsValidatorsByDeposit()
	if err != nil {
		return selector, fmt.Errorf("invalid deposit address")
	}
	if byDeposit.DepositAddress == "" {
		return selector, errEmptySelector
	}

	address, err := ParseEthereumAddress(byDeposit.DepositAddress)
	if err != nil {
		return selector, fmt.Errorf("parsing deposit address: %w", err)
	}

	selector.DepositAddress = &address
	return selector, nil
}

// asWithdrawalInput tries to parse the input as a withdrawal address or credential selector.
func asWithdrawalInput(input model.ValidatorsSelector) (domain.ValidatorsSelector, error) {
	selector := domain.ValidatorsSelector{}
	byWithdrawal, err := input.AsValidatorsByWithdrawal()
	if err != nil {
		return selector, fmt.Errorf("invalid withdrawal selector: %w", errEmptySelector)
	}
	if byWithdrawal.Withdrawal == "" {
		return selector, errEmptySelector
	}

	credsBytes, err := decodeHexString(byWithdrawal.Withdrawal)
	if err != nil {
		return selector, fmt.Errorf("decoding withdrawal hex: %w", err)
	}

	if len(credsBytes) == 20 {
		address := domain.EthereumAddress(credsBytes)
		selector.WithdrawalAddress = &address
		return selector, nil
	}

	if len(credsBytes) != 32 {
		return selector, fmt.Errorf("invalid withdrawal credential: %d", len(credsBytes))
	}
	credential := domain.WithdrawalCredential(credsBytes)
	selector.WithdrawalCredential = &credential
	return selector, nil
}

// asIdentifiers tries to parse the input as a validator identifiers selector.
func asIdentifiers(input model.ValidatorsSelector) (domain.ValidatorsSelector, error) {
	selector := domain.ValidatorsSelector{}
	byIdentifiers, err := input.AsValidatorsByIdentifiers()
	if err != nil {
		return selector, errEmptySelector
	}

	var indices []int
	var publicKeys [][]byte

	for _, val := range byIdentifiers.ValidatorIdentifiers {
		if index, err := val.AsValidatorIndex(); err == nil {
			indices = append(indices, index)
			continue
		}

		publicKey, err := val.AsValidatorPublicKey()
		if err != nil {
			return selector, fmt.Errorf("invalid validator identifier: %w", err)
		}
		if publicKey == "" {
			return selector, errEmptySelector
		}

		pubkeyBytes, err := decodeHexString(publicKey)
		if err != nil {
			return selector, fmt.Errorf("invalid validator pubkey: %w", err)
		}

		publicKeys = append(publicKeys, pubkeyBytes)
	}
	if len(indices) == 0 && len(publicKeys) == 0 {
		return selector, errEmptySelector
	}

	selector.Identifiers = &domain.ValidatorsByIdentifier{
		Indices:    indices,
		PublicKeys: publicKeys,
	}
	return selector, nil
}

// ParseEthereumAddress parses a hex-encoded Ethereum address (with or without "0x" prefix).
func ParseEthereumAddress(address string) (domain.EthereumAddress, error) {
	addressBytes, err := decodeHexString(address)
	if err != nil {
		return domain.EthereumAddress{}, err
	}
	return domain.EthereumAddress(addressBytes), nil
}

// decodeHexString decodes a hex string, allowing an optional "0x" prefix.
func decodeHexString(s string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(s, "0x"))
}
