package consapi

import (
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type Client struct {
	ClientInt
}
type ClientInt interface {
	// /eth/v2/beacon/blocks/{block_id}
	GetSlot(blockID any) (*types.StandardBeaconSlotResponse, error)

	// Optional params ids and status to filter the response.
	// eth/v1/beacon/states/{state_id}/validators
	GetValidators(state any, ids []string, status []types.ValidatorStatus) (*types.StandardValidatorsResponse, error)

	// eth/v1/beacon/states/{state_id}/validators/{validator_id}
	GetValidator(validatorID, stateID any) (*types.StandardSingleValidatorsResponse, error)

	// /eth/v1/validator/duties/proposer/{epoch}
	GetPropoalAssignments(epoch uint64) (*types.StandardProposerAssignmentsResponse, error)

	// /eth/v1/beacon/rewards/blocks/{block_id}
	GetPropoalRewards(blockID any) (*types.StandardBlockRewardsResponse, error)

	// /eth/v1/beacon/rewards/sync_committee/{block_id}
	GetSyncRewards(blockID any) (*types.StandardSyncCommitteeRewardsResponse, error)

	// /eth/v1/beacon/rewards/attestations/{epoch}
	GetAttestationRewards(epoch uint64) (*types.StandardAttestationRewardsResponse, error)

	// /eth/v1/beacon/states/{state_id}/sync_committees
	GetSyncCommitteesAssignments(epoch *uint64, stateID any) (*types.StandardSyncCommitteesResponse, error)

	// /eth/v1/config/spec
	GetSpec() (*types.StandardSpecResponse, error)

	// /eth/v1/beacon/headers/{block_id}
	GetBlockHeader(blockID any) (*types.StandardBeaconHeaderResponse, error)

	// /eth/v1/beacon/headers
	GetBlockHeaders(slot *uint64, parentRoot *any) (*types.StandardBeaconHeadersResponse, error)

	// /eth/v1/beacon/states/{state_id}/finality_checkpoints
	GetFinalityCheckpoints(stateID any) (*types.StandardFinalityCheckpointsResponse, error)

	// /eth/v1/beacon/states/{state_id}/validator_balances
	GetValidatorBalances(stateID any) (*types.StandardValidatorBalancesResponse, error)

	// /eth/v1/beacon/blob_sidecars/{block_id}
	GetBlobSidecars(blockID any) (*types.StandardBlobSidecarsResponse, error)

	// Optional params epoch, index and slot
	// /eth/v1/beacon/states/%v/committees
	GetCommittees(stateID any, epoch, index, slot *uint64) (*types.StandardCommitteesResponse, error)

	// /eth/v1/beacon/genesis
	GetGenesis() (*types.StandardGenesisResponse, error)

	// /eth/v1/events
	GetEvents(topics []types.EventTopic) chan *types.EventResponse
}
type NodeClient struct {
	Endpoint   string
	httpClient *http.Client
}
