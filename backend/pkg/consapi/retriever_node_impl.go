package consapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/consapi/utils"
)

func NewNodeDataRetriever(endpoint string) Retriever {
	retriever := Retriever{
		RetrieverInt: &NodeImplRetriever{
			Endpoint: endpoint,
			httpClient: &http.Client{
				Timeout: 350 * time.Second,
			},
		},
	}
	return retriever
}

func (r *NodeImplRetriever) GetFinalityCheckpoints(stateID any) (types.StandardFinalityCheckpointsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", r.Endpoint, stateID)
	return get[types.StandardFinalityCheckpointsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetBlockHeader(blockID any) (types.StandardBeaconHeaderResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/headers/%v", r.Endpoint, blockID)
	return get[types.StandardBeaconHeaderResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetSyncCommitteesAssignments(epoch int, stateID any) (types.StandardSyncCommitteesResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%v/sync_committees?epoch=%d", r.Endpoint, stateID, epoch)
	return get[types.StandardSyncCommitteesResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetSpec() (types.StandardSpecResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/config/spec", r.Endpoint)
	return get[types.StandardSpecResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetSlot(blockID any) (types.StandardBeaconSlotResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v2/beacon/blocks/%v", r.Endpoint, blockID)
	return get[types.StandardBeaconSlotResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetValidators(state any) (types.StandardValidatorsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", r.Endpoint, state)
	return get[types.StandardValidatorsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetValidator(validatorID, state any) (types.StandardSingleValidatorsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%s/validators/%v", r.Endpoint, state, validatorID)
	return get[types.StandardSingleValidatorsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetPropoalAssignments(epoch int) (types.StandardProposerAssignmentsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", r.Endpoint, epoch)
	return get[types.StandardProposerAssignmentsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetPropoalRewards(blockID any) (types.StandardBlockRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%v", r.Endpoint, blockID)
	return get[types.StandardBlockRewardsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetSyncRewards(blockID any) (types.StandardSyncCommitteeRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%v", r.Endpoint, blockID)
	return post[types.StandardSyncCommitteeRewardsResponse](r, requestURL)
}

func (r *NodeImplRetriever) GetAttestationRewards(blockID any) (types.StandardAttestationRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%v", r.Endpoint, blockID)
	return post[types.StandardAttestationRewardsResponse](r, requestURL)
}

// Helper for get and unmarshal
func get[T any](r *NodeImplRetriever, url string) (T, error) {
	result, err := genericRequest("GET", url, r.httpClient)
	if err != nil || result == nil {
		var target T
		return target, err
	}
	return utils.Unmarshal[T](result, err)
}

// Helper for post and unmarshal
func post[T any](r *NodeImplRetriever, url string) (T, error) {
	result, err := genericRequest("POST", url, r.httpClient)
	if err != nil || result == nil {
		var target T
		return target, err
	}
	return utils.Unmarshal[T](result, err)
}

func genericRequest(method string, requestURL string, httpClient *http.Client) ([]byte, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		// rethink error handling in explorer?
		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		if res.StatusCode == http.StatusBadRequest {
			return nil, nil
		}
		if res.StatusCode == http.StatusInternalServerError {
			return nil, nil
		}
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return resString, nil
}

type NodeImplRetriever struct {
	Endpoint   string
	httpClient *http.Client
}
