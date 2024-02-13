package consapi

import "github.com/gobitfly/beaconchain/pkg/consapi/types"

type Retriever struct {
	RetrieverInt
}
type RetrieverInt interface {
	// /eth/v2/beacon/blocks/{block_id}
	GetSlot(block_id any) (types.StandardBeaconSlotResponse, error)

	// eth/v1/beacon/states/{state_id}/validators
	GetValidators(state any) (types.StandardValidatorsResponse, error)

	// eth/v1/beacon/states/{state_id}/validators/{validator_id}
	GetValidator(validator_id, state any) (types.StandardSingleValidatorsResponse, error)

	// /eth/v1/validator/duties/proposer/{epoch}
	GetPropoalAssignments(epoch int) (types.StandardProposerAssignmentsResponse, error)

	// /eth/v1/beacon/rewards/blocks/{block_id}
	GetPropoalRewards(block_id any) (types.StandardBlockRewardsResponse, error)

	// /eth/v1/beacon/rewards/sync_committee/{block_id}
	GetSyncRewards(block_id any) (types.StandardSyncCommitteeRewardsResponse, error)

	// /eth/v1/beacon/rewards/attestations/{epoch}
	GetAttestationRewards(block_id any) (types.StandardAttestationRewardsResponse, error)

	// /eth/v1/beacon/states/{state_id}/sync_committees
	GetSyncCommitteesAssignments(epoch int, state_id any) (types.StandardSyncCommitteesResponse, error)

	// /eth/v1/config/spec
	GetSpec() (types.StandardSpecResponse, error)

	// /eth/v1/beacon/headers/{block_id}
	GetBlockHeader(block_id any) (types.StandardBeaconHeaderResponse, error)

	// /eth/v1/beacon/states/{state_id}/finality_checkpoints
	GetFinalityCheckpoints(state_id any) (types.StandardFinalityCheckpointsResponse, error)
}
