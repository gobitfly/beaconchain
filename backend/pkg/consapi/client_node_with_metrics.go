package consapi

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics" //nolint:depguard
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type NodeClientWithMetrics struct {
	client  ClientInt
	metrics metrics.MetricsRepository
}

func NewNodeClientWithMetrics(client ClientInt, metrics metrics.MetricsRepository) Client {
	return Client{
		ClientInt: &NodeClientWithMetrics{
			client:  client,
			metrics: metrics,
		},
	}
}

func (c *NodeClientWithMetrics) GetValidatorBalances(stateID any) (*types.StandardValidatorBalancesResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_validator_balances", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetValidatorBalances(stateID)
	if err != nil {
		c.metrics.Error("node_get_validator_balances")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetFinalityCheckpoints(stateID any) (*types.StandardFinalityCheckpointsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_finality_checkpoints", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetFinalityCheckpoints(stateID)
	if err != nil {
		c.metrics.Error("node_get_finality_checkpoints")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetBlockHeader(blockID any) (*types.StandardBeaconHeaderResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_block_header", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetBlockHeader(blockID)
	if err != nil {
		c.metrics.Error("node_get_block_header")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetBlockHeaders(slot *uint64, parentRoot *any) (*types.StandardBeaconHeadersResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_block_headers", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetBlockHeaders(slot, parentRoot)
	if err != nil {
		c.metrics.Error("node_get_block_headers")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetSyncCommitteesAssignments(epoch *uint64, stateID any) (*types.StandardSyncCommitteesResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_sync_committees_assignments", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetSyncCommitteesAssignments(epoch, stateID)
	if err != nil {
		c.metrics.Error("node_get_sync_committees_assignments")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetSpec() (*types.StandardSpecResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_spec", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetSpec()
	if err != nil {
		c.metrics.Error("node_get_spec")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetSlot(blockID any) (*types.StandardBeaconSlotResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_slot", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetSlot(blockID)
	if err != nil {
		c.metrics.Error("node_get_slot")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetValidators(state any, ids []string, status []types.ValidatorStatus) (*types.StandardValidatorsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_validators", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetValidators(state, ids, status)
	if err != nil {
		c.metrics.Error("node_get_validators")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetValidator(validatorID, state any) (*types.StandardSingleValidatorsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_validator", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetValidator(validatorID, state)
	if err != nil {
		c.metrics.Error("node_get_validator")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetProposalAssignments(epoch uint64) (*types.StandardProposerAssignmentsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_proposal_assignments", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetProposalAssignments(epoch)
	if err != nil {
		c.metrics.Error("node_get_proposal_assignments")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetProposalRewards(blockID any) (*types.StandardBlockRewardsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_proposal_rewards", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetProposalRewards(blockID)
	if err != nil {
		c.metrics.Error("node_get_proposal_rewards")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetSyncRewards(blockID any) (*types.StandardSyncCommitteeRewardsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_sync_rewards", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetSyncRewards(blockID)
	if err != nil {
		c.metrics.Error("node_get_sync_rewards")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetAttestationRewards(epoch uint64) (*types.StandardAttestationRewardsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_attestation_rewards", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetAttestationRewards(epoch)
	if err != nil {
		c.metrics.Error("node_get_attestation_rewards")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetBlobSidecars(blockID any) (*types.StandardBlobSidecarsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_blob_sidecars", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetBlobSidecars(blockID)
	if err != nil {
		c.metrics.Error("node_get_blob_sidecars")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetCommittees(stateID any, epoch, index, slot *uint64) (*types.StandardCommitteesResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_committees", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetCommittees(stateID, epoch, index, slot)
	if err != nil {
		c.metrics.Error("node_get_committees")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetGenesis() (*types.StandardGenesisResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_genesis", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetGenesis()
	if err != nil {
		c.metrics.Error("node_get_genesis")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetEvents(topics []types.EventTopic) chan *types.EventResponse {
	timeStart := time.Now()
	originalCh := c.client.GetEvents(topics)
	wrappedCh := make(chan *types.EventResponse, len(originalCh))

	go func() {
		defer close(wrappedCh)
		firstEvent := true

		for event := range originalCh {
			if firstEvent {
				c.metrics.ObserveClientCallDuration("node", "get_events", time.Since(timeStart))
				firstEvent = false
			}

			if event.Error != nil {
				c.metrics.Error("node_get_events")
			}

			wrappedCh <- event
		}
	}()

	return wrappedCh
}

func (c *NodeClientWithMetrics) GetState(stateID any) (*types.StandardBeaconStateResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("node", "get_state", time.Since(timeStart))
	}(timeStart)

	resp, err := c.client.GetState(stateID)
	if err != nil {
		c.metrics.Error("node_get_state")
		return nil, err
	}
	return resp, nil
}

func (c *NodeClientWithMetrics) GetEndpoint() string {
	return c.client.GetEndpoint()
}
