package clnode

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/commons/utils"
)

/**
* This implementation retrieves data directly from node
 */

func NewNodeDataRetriever(endpoint string, chainConfig *ClChainConfig) (Retriever, error) {
	retriever := Retriever{
		RetrieverInt: &nodeImplRetriever{
			endpoint: endpoint,
			httpClient: &http.Client{
				Timeout: 120 * time.Second,
			},
		},
	}

	if chainConfig != nil {
		retriever.ChainConfig = *chainConfig
	} else {
		config, err := retriever.GetSpec()
		if err != nil {
			return retriever, fmt.Errorf("error retrieving chain config: %v", err)
		}
		retriever.ChainConfig = config.Data
	}

	return retriever, nil
}

func (r *nodeImplRetriever) GetFinalityCheckpoints(state_id string) (StandardFinalityCheckpointsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", r.endpoint, state_id)
	return get[StandardFinalityCheckpointsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetBlockHeader(block_id string) (StandardBeaconHeaderResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/headers/head", r.endpoint)
	return get[StandardBeaconHeaderResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetSyncCommitteesAssignments(epoch int, slot int64) (GetSyncCommitteeAssignmentsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees?epoch=%d", r.endpoint, slot, epoch)
	return get[GetSyncCommitteeAssignmentsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetSpec() (GetSpecResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/config/spec", r.endpoint)
	return get[GetSpecResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetSlot(slot int) (GetBeaconSlotResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", r.endpoint, slot)
	return get[GetBeaconSlotResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetValidators(slot int) (GetValidatorsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators", r.endpoint, slot)
	return get[GetValidatorsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetPropoalAssignments(epoch int) (GetProposerAssignmentsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", r.endpoint, epoch)
	return get[GetProposerAssignmentsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetPropoalRewards(slot int) (GetBlockRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%d", r.endpoint, slot)
	return get[GetBlockRewardsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetSyncRewards(slot int) (GetSyncCommitteeRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%d", r.endpoint, slot)
	return post[GetSyncCommitteeRewardsResponse](r, requestUrl)
}

func (r *nodeImplRetriever) GetAttestationRewards(epoch int) (GetAttestationRewardsResponse, error) {
	requestUrl := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%d", r.endpoint, epoch)
	return post[GetAttestationRewardsResponse](r, requestUrl)
}

// Helper for get and unmarshal
func get[T any](r *nodeImplRetriever, url string) (T, error) {
	return utils.Unmarshal[T](genericRequest("GET", url, r.httpClient))
}

// Helper for post and unmarshal
func post[T any](r *nodeImplRetriever, url string) (T, error) {
	return utils.Unmarshal[T](genericRequest("POST", url, r.httpClient))
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

type nodeImplRetriever struct {
	endpoint   string
	httpClient *http.Client
}
