package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
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
	validatorByIndex:                 reNumber,
	validatorByPublicKey:             reValidatorPublicKey,
	validatorsByDepositAddress:       reEthereumAddress,
	validatorsByDepositEnsName:       reEnsName,
	validatorsByWithdrawalCredential: reWithdrawalCredential,
	validatorsByWithdrawalAddress:    reEthereumAddress,
	validatorsByWithdrawalEns:        reEnsName,
	validatorsByGraffiti:             reNonEmpty,
}

// --------------------------------------
//   Handler func

func (h *HandlerService) InternalPostSearch(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Input    string          `json:"input"`
		Networks []network       `json:"networks,omitempty"`
		Types    []searchTypeKey `json:"types,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	// if the input slices are empty, the sets will contain all possible values
	networkSet := v.checkNetworkSlice(req.Networks)
	searchTypeSet := v.checkSearchTypes(req.Types)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	g, ctx := errgroup.WithContext(r.Context())
	g.SetLimit(20)
	searchResultChan := make(chan types.SearchResult)

	// iterate over all combinations of search types and networks
	for searchType := range searchTypeSet {
		// check if input matches the regex for the search type
		if !searchTypeToRegex[searchType].MatchString(req.Input) {
			continue
		}
		for network := range networkSet {
			network := network
			searchType := searchType
			g.Go(func() error {
				searchResult, err := h.handleSearch(ctx, req.Input, searchType, uint64(network))
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
		handleErr(w, err)
		return
	}

	response := types.InternalPostSearchResponse{
		Data: data,
	}
	returnOk(w, response)
}

// --------------------------------------
//	 Search Helper Functions

func (h *HandlerService) handleSearch(ctx context.Context, input string, searchType searchTypeKey, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
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
}

func (h *HandlerService) handleSearchValidatorByIndex(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		index, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			// input should've been checked by the regex before, this should never happen
			return nil, err
		}
		result, err := h.dai.GetSearchValidatorByIndex(ctx, chainId, index)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorByIndex),
			ChainId:   chainId,
			HashValue: hex.EncodeToString(result.PublicKey),
			NumValue:  &result.Index,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorByPublicKey(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		publicKey, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
		if err != nil {
			// input should've been checked by the regex before, this should never happen
			return nil, err
		}
		result, err := h.dai.GetSearchValidatorByPublicKey(ctx, chainId, publicKey)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorByPublicKey),
			ChainId:   chainId,
			HashValue: hex.EncodeToString(result.PublicKey),
			NumValue:  &result.Index,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByDepositAddress(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		address, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
		if err != nil {
			return nil, err
		}
		result, err := h.dai.GetSearchValidatorsByDepositAddress(ctx, chainId, address)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorsByDepositAddress),
			ChainId:   chainId,
			HashValue: hex.EncodeToString(result.Address),
			NumValue:  &result.Count,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByDepositEnsName(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		result, err := h.dai.GetSearchValidatorsByDepositEnsName(ctx, chainId, input)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorsByDepositEnsName),
			ChainId:   chainId,
			StrValue:  result.EnsName,
			HashValue: hex.EncodeToString(result.Address),
			NumValue:  &result.Count,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalCredential(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		withdrawalCredential, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
		if err != nil {
			return nil, err
		}
		result, err := h.dai.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorsByWithdrawalCredential),
			ChainId:   chainId,
			HashValue: hex.EncodeToString(result.WithdrawalCredential),
			NumValue:  &result.Count,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalAddress(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		withdrawalString := "010000000000000000000000" + strings.TrimPrefix(input, "0x")
		withdrawalCredential, err := hex.DecodeString(withdrawalString)
		if err != nil {
			return nil, err
		}
		result, err := h.dai.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorsByWithdrawalAddress),
			ChainId:   chainId,
			HashValue: hex.EncodeToString(result.WithdrawalCredential),
			NumValue:  &result.Count,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByWithdrawalEnsName(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		result, err := h.dai.GetSearchValidatorsByWithdrawalEnsName(ctx, chainId, input)
		if err != nil {
			return nil, err
		}

		return &types.SearchResult{
			Type:      string(validatorsByWithdrawalEns),
			ChainId:   chainId,
			StrValue:  result.EnsName,
			HashValue: hex.EncodeToString(result.Address),
			NumValue:  &result.Count,
		}, nil
	}
}

func (h *HandlerService) handleSearchValidatorsByGraffiti(ctx context.Context, input string, chainId uint64) (*types.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		result, err := h.dai.GetSearchValidatorsByGraffiti(ctx, chainId, input)
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
}

// --------------------------------------
//   Input Validation

// if the passed slice is empty, return a set with all networks; otherwise check if the passed networks are valid
func (v *validationError) checkNetworkSlice(networksSlice []network) map[network]struct{} {
	networkSet := map[network]struct{}{}
	// if the list is empty, query all networks
	if len(networksSlice) == 0 {
		for _, n := range allNetworks {
			networkSet[network(n.ChainId)] = struct{}{}
		}
		return networkSet
	}
	// list not empty, check if networks are valid
	for _, n := range networksSlice {
		// chain id was already checked in the unmarshal step, if it's invalid it will be -1
		if n == -1 {
			v.add("networks", "list contains invalid network, please check the API documentation")
			break
		}
		networkSet[n] = struct{}{}
	}
	return networkSet
}

// if the passed slice is empty, return a set with all search types; otherwise check if the passed types are valid
func (v *validationError) checkSearchTypes(types []searchTypeKey) map[searchTypeKey]struct{} {
	typeSet := map[searchTypeKey]struct{}{}
	// if the list is empty, query all types
	if len(types) == 0 {
		for t := range searchTypeToRegex {
			typeSet[t] = struct{}{}
		}
		return typeSet
	}
	// list not empty, check if types are valid
	for _, t := range types {
		if _, typeExists := searchTypeToRegex[t]; !typeExists {
			v.add("types", "list contains invalid type, please check the API documentation")
			break
		}
		typeSet[t] = struct{}{}
	}
	return typeSet
}
