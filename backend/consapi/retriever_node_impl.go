package consapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/consapi/types"
	"github.com/gobitfly/beaconchain/consapi/utils"
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

func (r *NodeImplRetriever) GetFinalityCheckpoints(state_id any) (types.StandardFinalityCheckpointsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", r.Endpoint, state_id)
	return get[types.StandardFinalityCheckpointsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetBlockHeader(block_id any) (types.StandardBeaconHeaderResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/headers/%v", r.Endpoint, block_id)
	return get[types.StandardBeaconHeaderResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetSyncCommitteesAssignments(epoch int, state_id any) (types.StandardSyncCommitteesResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%v/sync_committees?epoch=%d", r.Endpoint, state_id, epoch)
	return get[types.StandardSyncCommitteesResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetSpec() (types.StandardSpecResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/config/spec", r.Endpoint)
	return get[types.StandardSpecResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetSlot(block_id any) (types.StandardBeaconSlotResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v2/beacon/blocks/%v", r.Endpoint, block_id)
	return get[types.StandardBeaconSlotResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetValidators(state any) (types.StandardValidatorsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", r.Endpoint, state)
	return get[types.StandardValidatorsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetValidator(validator_id, state any) (types.StandardSingleValidatorsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%s/validators/%v", r.Endpoint, state, validator_id)
	return get[types.StandardSingleValidatorsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetPropoalAssignments(epoch int) (types.StandardProposerAssignmentsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", r.Endpoint, epoch)
	return get[types.StandardProposerAssignmentsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetPropoalRewards(block_id any) (types.StandardBlockRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%v", r.Endpoint, block_id)
	return get[types.StandardBlockRewardsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetSyncRewards(block_id any) (types.StandardSyncCommitteeRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%v", r.Endpoint, block_id)
	return post[types.StandardSyncCommitteeRewardsResponse](r, requestUrl)
}

func (r *NodeImplRetriever) GetAttestationRewards(block_id any) (types.StandardAttestationRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%v", r.Endpoint, block_id)
	return post[types.StandardAttestationRewardsResponse](r, requestUrl)
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

func genericRequest(method string, requestUrl string, httpClient *http.Client) ([]byte, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestUrl, bytes.NewBuffer(data))
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
