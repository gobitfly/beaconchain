package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"golang.org/x/sync/errgroup"
)

type searchTypeKey string

const (
	validatorByIndex           searchTypeKey = "validator_by_index"
	validatorByPublicKey       searchTypeKey = "validator_by_public_key"
	validatorsByDepositAddress searchTypeKey = "validators_by_deposit_address"
	validatorsByDepositEnsName searchTypeKey = "validators_by_deposit_ens_name"
	//nolint:gosec
	validatorsByWithdrawalCredential searchTypeKey = "validators_by_withdrawal_credential"
	validatorsByWithdrawalAddress    searchTypeKey = "validators_by_withdrawal_address"
	validatorsByWithdrawalEns        searchTypeKey = "validators_by_withdrawal_ens_name"
	validatorsByGraffiti             searchTypeKey = "validators_by_graffiti"
)

// source of truth for all possible search types and their regex
var searchTypeToRegex = map[searchTypeKey]*regexp.Regexp{
	validatorByIndex:                 reInteger,
	validatorByPublicKey:             reValidatorPublicKey,
	validatorsByDepositAddress:       reEthereumAddress,
	validatorsByDepositEnsName:       reEnsName,
	validatorsByWithdrawalCredential: reWithdrawalCredential,
	validatorsByWithdrawalAddress:    reEthereumAddress,
	validatorsByWithdrawalEns:        reEnsName,
	validatorsByGraffiti:             reGraffiti,
}

// --------------------------------------
//   Handler func

func (h *HandlerService) InternalPostSearch(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Input    string          `json:"input"`
		Networks []intOrString   `json:"networks,omitempty"`
		Types    []searchTypeKey `json:"types,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	// if the input slices are empty, the sets will contain all possible values
	chainIdSet := v.checkNetworkSlice(req.Networks)
	searchTypeSet := v.checkSearchTypes(req.Types)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	g, ctx := errgroup.WithContext(r.Context())
	g.SetLimit(20)
	searchResultChan := make(chan types.SearchResult)

	// iterate over all combinations of search types and networks
	for _, searchType := range searchTypeSet {
		// check if input matches the regex for the search type
		if !searchTypeToRegex[searchType].MatchString(req.Input) {
			continue
		}
		for _, chainId := range chainIdSet {
			chainId := chainId
			searchType := searchType
			g.Go(func() error {
				searchResult, err := h.handleSearch(ctx, req.Input, searchType, chainId)
				if err != nil {
					if errors.Is(err, dataaccess.ErrNotFound) {
						return nil
					}
					return err
				}
				if searchResult != nil { // if the search result is nil, the input didn't match the search type
					searchResultChan <- *searchResult
				}
				return nil
			})
		}
	}

	var err error
	go func() {
		err = g.Wait()
		close(searchResultChan)
	}()

	data := make([]types.SearchResult, 0)
	for result := range searchResultChan {
		data = append(data, result)
	}

	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalPostSearchResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

// --------------------------------------
//	 Search Helper Functions

func (h *HandlerService) handleSearch(ctx context.Context, input string, searchType searchTypeKey, chainId uint64) (*types.SearchResult, error) {
	switch searchType {
	case validatorByIndex:
		return h.handleSearchValidatorByIndex(ctx, input, chainId)
	case validatorByPublicKey:
		return h.handleSearchValidatorByPublicKey(ctx, input, chainId)
	case validatorsByDepositAddress:
		return h.handleSearchValidatorsByDepositAddress(ctx, input, chainId)
	case validatorsByDepositEnsName:
		return h.handleSearchValidatorsByDepositEnsName(ctx, input, chainId)
	case validatorsByWithdrawalCredential:
		return h.handleSearchValidatorsByWithdrawalCredential(ctx, input, chainId)
	case validatorsByWithdrawalAddress:
		return h.handleSearchValidatorsByWithdrawalAddress(ctx, input, chainId)
	case validatorsByWithdrawalEns:
		return h.handleSearchValidatorsByWithdrawalEnsName(ctx, input, chainId)
	case validatorsByGraffiti:
		return h.handleSearchValidatorsByGraffiti(ctx, input, chainId)
	default:
		return nil, errors.New("invalid search type")
	}
}

func (h *HandlerService) handleSearchValidatorByIndex(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	index, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		// input should've been checked by the regex before, this should never happen
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorByIndex(ctx, chainId, index)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorByIndex),
		ChainId:   chainId,
		HashValue: "0x" + hex.EncodeToString(result.PublicKey),
		NumValue:  &result.Index,
	}, nil
}

func (h *HandlerService) handleSearchValidatorByPublicKey(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	publicKey, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		// input should've been checked by the regex before, this should never happen
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorByPublicKey(ctx, chainId, publicKey)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorByPublicKey),
		ChainId:   chainId,
		HashValue: "0x" + hex.EncodeToString(result.PublicKey),
		NumValue:  &result.Index,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByDepositAddress(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	address, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByDepositAddress(ctx, chainId, address)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorsByDepositAddress),
		ChainId:   chainId,
		HashValue: "0x" + hex.EncodeToString(result.Address),
		NumValue:  &result.Count,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByDepositEnsName(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchValidatorsByDepositEnsName(ctx, chainId, input)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorsByDepositEnsName),
		ChainId:   chainId,
		StrValue:  result.EnsName,
		HashValue: "0x" + hex.EncodeToString(result.Address),
		NumValue:  &result.Count,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalCredential(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	withdrawalCredential, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorsByWithdrawalCredential),
		ChainId:   chainId,
		HashValue: "0x" + hex.EncodeToString(result.WithdrawalCredential),
		NumValue:  &result.Count,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalAddress(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	withdrawalString := "010000000000000000000000" + strings.TrimPrefix(input, "0x")
	withdrawalCredential, err := hex.DecodeString(withdrawalString)
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorsByWithdrawalAddress),
		ChainId:   chainId,
		HashValue: "0x" + hex.EncodeToString(result.WithdrawalCredential),
		NumValue:  &result.Count,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalEnsName(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchValidatorsByWithdrawalEnsName(ctx, chainId, input)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:      string(validatorsByWithdrawalEns),
		ChainId:   chainId,
		StrValue:  result.EnsName,
		HashValue: "0x" + hex.EncodeToString(result.Address),
		NumValue:  &result.Count,
	}, nil
}

func (h *HandlerService) handleSearchValidatorsByGraffiti(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchValidatorsByGraffiti(ctx, chainId, input)
	if err != nil {
		return nil, err
	}

	return &types.SearchResult{
		Type:     string(validatorsByGraffiti),
		ChainId:  chainId,
		StrValue: result.Graffiti,
		NumValue: &result.Count,
	}, nil
}

// --------------------------------------
//   Input Validation

// if the passed slice is empty, return a set with all chain IDs; otherwise check if the passed networks are valid
func (v *validationError) checkNetworkSlice(networks []intOrString) []uint64 {
	networkSet := map[uint64]struct{}{}
	// if the list is empty, query all networks
	if len(networks) == 0 {
		for _, n := range allNetworks {
			networkSet[n.ChainId] = struct{}{}
		}
		return slices.Collect(maps.Keys(networkSet))
	}
	// list not empty, check if networks are valid
	for _, network := range networks {
		chainId, ok := isValidNetwork(network)
		if !ok {
			v.add("networks", fmt.Sprintf("invalid network '%s'", network))
			break
		}
		networkSet[chainId] = struct{}{}
	}
	return slices.Collect(maps.Keys(networkSet))
}

// if the passed slice is empty, return a set with all search types; otherwise check if the passed types are valid
func (v *validationError) checkSearchTypes(types []searchTypeKey) []searchTypeKey {
	typeSet := map[searchTypeKey]struct{}{}
	// if the list is empty, query all types
	if len(types) == 0 {
		for t := range searchTypeToRegex {
			typeSet[t] = struct{}{}
		}
		return slices.Collect(maps.Keys(typeSet))
	}
	// list not empty, check if types are valid
	for _, t := range types {
		if _, typeExists := searchTypeToRegex[t]; !typeExists {
			v.add("types", fmt.Sprintf("invalid search type '%s'", t))
			continue
		}
		typeSet[t] = struct{}{}
	}
	return slices.Collect(maps.Keys(typeSet))
}
