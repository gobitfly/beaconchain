package consapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/consapi/utils"
)

func NewClient(endpoint string) Client {
	return NewClientWithConfig(endpoint, nil)
}

func NewClientWithConfig(endpoint string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 500 * time.Second,
		}
	}

	retriever := Client{
		ClientInt: &NodeClient{
			Endpoint:   endpoint,
			httpClient: httpClient,
		},
	}
	return retriever
}

func (r *NodeClient) GetValidatorBalances(stateID any) (*types.StandardValidatorBalancesResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%v/validator_balances", r.Endpoint, stateID)
	return network.Get[types.StandardValidatorBalancesResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetFinalityCheckpoints(stateID any) (*types.StandardFinalityCheckpointsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%s/finality_checkpoints", r.Endpoint, stateID)
	return network.Get[types.StandardFinalityCheckpointsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetBlockHeader(blockID any) (*types.StandardBeaconHeaderResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/headers/%v", r.Endpoint, blockID)
	return network.Get[types.StandardBeaconHeaderResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetBlockHeaders(slot *uint64, parentRoot *any) (*types.StandardBeaconHeadersResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/headers", r.Endpoint)
	if slot != nil {
		requestURL += fmt.Sprintf("?slot=%d", *slot)
	} else if parentRoot != nil {
		requestURL += fmt.Sprintf("?parent_root=%v", *parentRoot)
	}
	return network.Get[types.StandardBeaconHeadersResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetSyncCommitteesAssignments(epoch *uint64, stateID any) (*types.StandardSyncCommitteesResponse, error) {
	var requestURL string
	if epoch == nil {
		requestURL = fmt.Sprintf("%s/eth/v1/beacon/states/%v/sync_committees", r.Endpoint, stateID)
	} else {
		requestURL = fmt.Sprintf("%s/eth/v1/beacon/states/%v/sync_committees?epoch=%d", r.Endpoint, stateID, epoch)
	}
	return network.Get[types.StandardSyncCommitteesResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetSpec() (*types.StandardSpecResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/config/spec", r.Endpoint)
	return network.Get[types.StandardSpecResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetSlot(blockID any) (*types.StandardBeaconSlotResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v2/beacon/blocks/%v", r.Endpoint, blockID)
	return network.Get[types.StandardBeaconSlotResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetValidators(state any, ids []string, status []types.ValidatorStatus) (*types.StandardValidatorsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%v/validators", r.Endpoint, state)
	if len(ids) > 0 {
		idStr := strings.Join(ids, ",")
		requestURL += fmt.Sprintf("?id=%s", idStr)
	}

	if len(status) > 0 {
		if len(ids) > 0 {
			requestURL += "&"
		} else {
			requestURL += "?"
		}
		statusStr := strings.Join(types.ConvertToStringSlice(status), ",")
		requestURL += fmt.Sprintf("status=%s", statusStr)
	}

	return network.Get[types.StandardValidatorsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetValidator(validatorID, state any) (*types.StandardSingleValidatorsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%s/validators/%v", r.Endpoint, state, validatorID)
	return network.Get[types.StandardSingleValidatorsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetPropoalAssignments(epoch uint64) (*types.StandardProposerAssignmentsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/validator/duties/proposer/%d", r.Endpoint, epoch)
	return network.Get[types.StandardProposerAssignmentsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetPropoalRewards(blockID any) (*types.StandardBlockRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%v", r.Endpoint, blockID)
	return network.Get[types.StandardBlockRewardsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetSyncRewards(blockID any) (*types.StandardSyncCommitteeRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/sync_committee/%v", r.Endpoint, blockID)
	return network.Post[types.StandardSyncCommitteeRewardsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetAttestationRewards(epoch uint64) (*types.StandardAttestationRewardsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/rewards/attestations/%v", r.Endpoint, epoch)
	return network.Post[types.StandardAttestationRewardsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetBlobSidecars(blockID any) (*types.StandardBlobSidecarsResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/blob_sidecars/%v", r.Endpoint, blockID)
	return network.Get[types.StandardBlobSidecarsResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetCommittees(stateID any, epoch, index, slot *uint64) (*types.StandardCommitteesResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/states/%v/committees", r.Endpoint, stateID)
	if epoch != nil {
		requestURL += fmt.Sprintf("?epoch=%d", *epoch)
	} else if index != nil {
		requestURL += fmt.Sprintf("?index=%d", *index)
	} else if slot != nil {
		requestURL += fmt.Sprintf("?slot=%d", *slot)
	}
	return network.Get[types.StandardCommitteesResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetGenesis() (*types.StandardGenesisResponse, error) {
	requestURL := fmt.Sprintf("%s/eth/v1/beacon/genesis", r.Endpoint)
	return network.Get[types.StandardGenesisResponse](r.httpClient, requestURL)
}

func (r *NodeClient) GetEvents(topics []types.EventTopic) chan *types.EventResponse {
	joinedTopics := strings.Join(utils.ConvertToStringSlice(topics), ",")
	requestURL := fmt.Sprintf("%s/eth/v1/events?topics=%v", r.Endpoint, joinedTopics)
	responseCh := make(chan *types.EventResponse, 32)

	go func() {
		stream, err := eventsource.Subscribe(requestURL, "")

		if err != nil {
			responseCh <- &types.EventResponse{Error: err}
			return
		}
		defer stream.Close()

		for {
			select {
			// It is important to register to Errors, otherwise the stream does not reconnect if the connection was lost
			case err := <-stream.Errors:
				responseCh <- &types.EventResponse{Error: err}
			case e := <-stream.Events:
				var response types.EventResponse
				response.Data = []byte(e.Data())
				response.Event = types.EventTopic(e.Event())

				responseCh <- &response
			}
		}
	}()
	return responseCh
}
