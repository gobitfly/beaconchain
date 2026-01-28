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
	validatorList              searchTypeKey = "validator_list"
	validatorsByDepositAddress searchTypeKey = "validators_by_deposit_address"
	validatorsByDepositEnsName searchTypeKey = "validators_by_deposit_ens_name"
	//nolint:gosec
	validatorsByWithdrawalCredential searchTypeKey = "validators_by_withdrawal_credential"
	validatorsByWithdrawalAddress    searchTypeKey = "validators_by_withdrawal_address"
	validatorsByWithdrawalEns        searchTypeKey = "validators_by_withdrawal_ens_name"
	validatorsByGraffiti             searchTypeKey = "validators_by_graffiti"
	validatorsByGraffitiHex          searchTypeKey = "validators_by_graffiti_hex"

	addressKey          searchTypeKey = "address"
	addressByEnsNameKey searchTypeKey = "address_by_ens_name"
	ensNameKey          searchTypeKey = "ens_name"
	transactionKey      searchTypeKey = "transaction"
	slotKey             searchTypeKey = "slot"
	slotByBlockRootKey  searchTypeKey = "slot_by_block_root"
	slotByStateRootKey  searchTypeKey = "slot_by_state_root"
	blockKey            searchTypeKey = "block"
	epochKey            searchTypeKey = "epoch"
	tokenKey            searchTypeKey = "token"
)

type searchType struct {
	regex        *regexp.Regexp
	responseType string
	handlerFunc  func(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error)
}

var searchTypeMap map[searchTypeKey]searchType

// using init to avoid initialization cycles, since handler functions may reference the map
func init() {
	// source of truth for all possible search types and their regex
	searchTypeMap = map[searchTypeKey]searchType{
		validatorByIndex: {
			regex:        types.ReInteger,
			responseType: "validator",
			handlerFunc:  handleSearchValidatorByIndex,
		},
		validatorByPublicKey: {
			regex:        types.ReValidatorPublicKey,
			responseType: "validator",
			handlerFunc:  handleSearchValidatorByPublicKey,
		},
		validatorList: {
			regex:        types.ReValidatorList,
			responseType: string(validatorList),
			handlerFunc:  handleSearchValidatorList,
		},
		validatorsByDepositAddress: {
			regex:        types.ReEthereumAddress,
			responseType: string(validatorsByDepositAddress),
			handlerFunc:  handleSearchValidatorsByDepositAddress,
		},
		validatorsByDepositEnsName: {
			regex:        types.ReEnsName,
			responseType: string(validatorsByDepositAddress),
			handlerFunc:  handleSearchValidatorsByDepositEnsName,
		},
		validatorsByWithdrawalCredential: {
			regex:        types.ReWithdrawalCredential,
			responseType: string(validatorsByWithdrawalCredential),
			handlerFunc:  handleSearchValidatorsByWithdrawalCredential,
		},
		validatorsByWithdrawalAddress: {
			regex:        types.ReEthereumAddress,
			responseType: string(validatorsByWithdrawalCredential),
			handlerFunc:  handleSearchValidatorsByWithdrawalAddress,
		},
		validatorsByWithdrawalEns: {
			regex:        types.ReEnsName,
			responseType: string(validatorsByWithdrawalCredential),
			handlerFunc:  handleSearchValidatorsByWithdrawalEnsName,
		},
		validatorsByGraffiti: {
			regex:        types.ReGraffiti,
			responseType: string(validatorsByGraffiti),
			handlerFunc:  handleSearchValidatorsByGraffiti,
		},
		validatorsByGraffitiHex: {
			regex:        types.ReGraffitiHex,
			responseType: string(validatorsByGraffiti),
			handlerFunc:  handleSearchValidatorsByGraffitiHex,
		},
		addressKey: {
			regex:        types.ReEthereumAddress,
			responseType: "address",
			handlerFunc:  handleSearchAddress,
		},
		addressByEnsNameKey: {
			regex:        types.ReEnsName,
			responseType: "address",
			handlerFunc:  handleSearchAddressByEnsName,
		},
		ensNameKey: {
			regex:        types.ReEnsName,
			responseType: string(ensNameKey),
			handlerFunc:  handleSearchEnsName,
		},
		transactionKey: {
			regex:        types.ReTransactionHashOrLatest,
			responseType: string(transactionKey),
			handlerFunc:  handleSearchTransaction,
		},
		blockKey: {
			regex:        types.ReIntegerOrLatest,
			responseType: string(blockKey),
			handlerFunc:  handleSearchBlock,
		},
		slotKey: {
			regex:        types.ReIntegerOrLatest,
			responseType: "slot",
			handlerFunc:  handleSearchSlot,
		},
		slotByBlockRootKey: {
			regex:        types.Re64ByteHash,
			responseType: "slot",
			handlerFunc:  handleSearchSlotByBlockRoot,
		},
		slotByStateRootKey: {
			regex:        types.Re64ByteHash,
			responseType: "slot",
			handlerFunc:  handleSearchSlotByStateRoot,
		},
		epochKey: {
			regex:        types.ReIntegerOrLatest,
			responseType: string(epochKey),
			handlerFunc:  handleSearchEpoch,
		},
		tokenKey: {
			regex:        types.ReEthereumAddress,
			responseType: string(tokenKey),
			handlerFunc:  handleSearchToken,
		},
	}
}

// --------------------------------------
//   Handler func

func (h *HandlerService) InternalPostSearch(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Input    string          `json:"input"`
		Networks []intOrString   `json:"networks"`
		Types    []searchTypeKey `json:"types"`
	}{}
	if err := v.checkBody(&req, r.Body); err != nil {
		handleErr(w, r, err)
		return
	}
	// if the input slices are empty, the sets will contain all possible values
	chainIdSet := v.checkNetworkSlice(req.Networks)
	searchTypeSet := v.checkSearchTypes(req.Types)
	if err := v.AsError(); err != nil {
		handleErr(w, r, err)
		return
	}

	g, ctx := errgroup.WithContext(r.Context())
	g.SetLimit(20)
	searchResultChan := make(chan types.SearchResult)

	// iterate over all combinations of search types and networks
	for _, searchType := range searchTypeSet {
		// check if input matches the regex for the search type
		if !searchTypeMap[searchType].regex.MatchString(req.Input) {
			continue
		}
		for _, chainId := range chainIdSet {
			chainId := chainId
			searchType := searchType
			g.Go(func() error {
				searchResult, err := searchTypeMap[searchType].handlerFunc(ctx, h, req.Input, chainId)
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

func asSearchResult[In any](searchType searchTypeKey, chainId uint64, result *In, err error) (*types.SearchResult, error) {
	if err != nil || result == nil {
		return nil, err
	}
	return &types.SearchResult{
		Type:    searchTypeMap[searchType].responseType,
		ChainId: chainId,
		Value:   result,
	}, nil
}

func handleSearchValidatorByIndex(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	index, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		// input should've been checked by the regex before, this should never happen
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorByIndex(ctx, chainId, index)
	return asSearchResult(validatorByIndex, chainId, result, err)
}

func handleSearchValidatorByPublicKey(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	publicKey, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		// input should've been checked by the regex before, this should never happen
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorByPublicKey(ctx, chainId, publicKey)
	return asSearchResult(validatorByPublicKey, chainId, result, err)
}

func handleSearchValidatorList(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	var v validationError
	// split the input string into a slice of strings
	indices, pubkeys := v.checkValidatorList(input, forbidEmpty)
	if v.hasErrors() {
		return nil, nil // return no error as to not disturb the other search types
	}
	validators, err := h.daService.GetValidatorsFromSlices(ctx, indices, pubkeys)
	if err != nil || validators == nil || len(validators) == 0 {
		return nil, err
	}

	return &types.SearchResult{
		Type:    searchTypeMap[validatorList].responseType,
		ChainId: chainId,
		Value:   types.SearchValidatorList{Validators: validators},
	}, nil
}

func handleSearchValidatorsByDepositAddress(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	address, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByDepositAddress(ctx, chainId, address)
	return asSearchResult(validatorsByDepositAddress, chainId, result, err)
}

func handleSearchValidatorsByDepositEnsName(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchValidatorsByDepositEnsName(ctx, chainId, input)
	return asSearchResult(validatorsByDepositEnsName, chainId, result, err)
}

func handleSearchValidatorsByWithdrawalCredential(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	withdrawalCredential, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
	return asSearchResult(validatorsByWithdrawalCredential, chainId, result, err)
}

func handleSearchValidatorsByWithdrawalAddress(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	withdrawalString := "010000000000000000000000" + strings.TrimPrefix(input, "0x")
	withdrawalCredential, err := hex.DecodeString(withdrawalString)
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, withdrawalCredential)
	return asSearchResult(validatorsByWithdrawalAddress, chainId, result, err)
}

func handleSearchValidatorsByWithdrawalEnsName(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchValidatorsByWithdrawalEnsName(ctx, chainId, input)
	return asSearchResult(validatorsByWithdrawalEns, chainId, result, err)
}

func handleSearchValidatorsByGraffiti(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	// regex could only verify max character length, validate max byte length here
	if len(input) > 32 {
		return nil, nil // return no error as to not disturb the other search types
	}
	result, err := h.daService.GetSearchValidatorsByGraffiti(ctx, chainId, input)
	return asSearchResult(validatorsByGraffiti, chainId, result, err)
}

func handleSearchValidatorsByGraffitiHex(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	graffitiHex, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	// exclude the empty hex graffiti
	var graffitiArray [32]byte
	copy(graffitiArray[:], graffitiHex)
	if graffitiArray == [32]byte{} {
		return nil, nil // return no error as to not disturb the other search types
	}
	result, err := h.daService.GetSearchValidatorsByGraffitiHex(ctx, chainId, graffitiHex)
	return asSearchResult(validatorsByGraffitiHex, chainId, result, err)
}

func handleSearchAddress(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	address, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchAddress(ctx, chainId, address)
	return asSearchResult(addressKey, chainId, result, err)
}

func handleSearchAddressByEnsName(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchAddressByEnsName(ctx, chainId, input)
	return asSearchResult(addressByEnsNameKey, chainId, result, err)
}

func handleSearchEnsName(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	result, err := h.daService.GetSearchEnsName(ctx, chainId, input)
	return asSearchResult(ensNameKey, chainId, result, err)
}

const latestSearch = "/latest"

func handleSearchTransaction(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	if input == latestSearch {
		result, err := h.daService.GetLatestTransaction(ctx)
		if err != nil {
			return nil, err
		}
		return asSearchResult(transactionKey, chainId, &types.SearchTransaction{TransactionHash: result}, nil)
	}
	transactionHash, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchTransaction(ctx, chainId, transactionHash)
	return asSearchResult(transactionKey, chainId, result, err)
}

func handleSearchBlock(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	if input == latestSearch {
		result, err := h.daService.GetLatestBlock(ctx)
		if err != nil {
			return nil, err
		}
		return asSearchResult(blockKey, chainId, &types.SearchBlock{BlockNumber: result}, nil)
	}
	blockNumber, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchBlock(ctx, chainId, blockNumber)
	return asSearchResult(blockKey, chainId, result, err)
}

func handleSearchSlot(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	if input == latestSearch {
		result, err := h.daService.GetLatestSlot(ctx)
		if err != nil {
			return nil, err
		}
		return asSearchResult(slotKey, chainId, &types.SearchSlot{Slot: result}, nil)
	}
	slot, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchSlot(ctx, chainId, slot)
	return asSearchResult(slotKey, chainId, result, err)
}

func handleSearchSlotByBlockRoot(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	blockRoot, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchSlotByBlockRoot(ctx, chainId, blockRoot)
	return asSearchResult(slotByBlockRootKey, chainId, result, err)
}

func handleSearchSlotByStateRoot(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	stateRoot, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchSlotByStateRoot(ctx, chainId, stateRoot)
	return asSearchResult(slotByStateRootKey, chainId, result, err)
}

func handleSearchEpoch(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	if input == latestSearch {
		result, err := h.daService.GetLatestSlot(ctx)
		if err != nil {
			return nil, err
		}
		epoch := result / h.cfg.ClConfig.SlotsPerEpoch
		return asSearchResult(epochKey, chainId, &types.SearchEpoch{Epoch: epoch}, nil)
	}
	epoch, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchEpoch(ctx, chainId, epoch)
	return asSearchResult(epochKey, chainId, result, err)
}

func handleSearchToken(ctx context.Context, h *HandlerService, input string, chainId uint64) (*types.SearchResult, error) {
	tokenAddress, err := hex.DecodeString(strings.TrimPrefix(input, "0x"))
	if err != nil {
		return nil, err
	}
	result, err := h.daService.GetSearchToken(ctx, chainId, tokenAddress)
	return asSearchResult(tokenKey, chainId, result, err)
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
		for t := range searchTypeMap {
			typeSet[t] = struct{}{}
		}
		return slices.Collect(maps.Keys(typeSet))
	}
	// list not empty, check if types are valid
	for _, t := range types {
		if _, typeExists := searchTypeMap[t]; !typeExists {
			v.add("types", fmt.Sprintf("invalid search type '%s'", t))
			continue
		}
		typeSet[t] = struct{}{}
	}
	return slices.Collect(maps.Keys(typeSet))
}
