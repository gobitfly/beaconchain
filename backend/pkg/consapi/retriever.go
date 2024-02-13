package consapi

import "github.com/gobitfly/beaconchain/pkg/consapi/types"

type Retriever struct {
	RetrieverInt
}
type RetrieverInt interface {
	// /eth/v2/beacon/blocks/{block_id}
	GetSlot(blockID any) (types.StandardBeaconSlotResponse, error)

	// eth/v1/beacon/states/{state_id}/validators
	GetValidators(state any) (types.StandardValidatorsResponse, error)

	// eth/v1/beacon/states/{state_id}/validators/{validator_id}
	GetValidator(validatorID, stateID any) (types.StandardSingleValidatorsResponse, error)

	// /eth/v1/validator/duties/proposer/{epoch}
	GetPropoalAssignments(epoch int) (types.StandardProposerAssignmentsResponse, error)

	// /eth/v1/beacon/rewards/blocks/{block_id}
	GetPropoalRewards(blockID any) (types.StandardBlockRewardsResponse, error)

	// /eth/v1/beacon/rewards/sync_committee/{block_id}
	GetSyncRewards(blockID any) (types.StandardSyncCommitteeRewardsResponse, error)

	// /eth/v1/beacon/rewards/attestations/{epoch}
	GetAttestationRewards(blockID any) (types.StandardAttestationRewardsResponse, error)

	// /eth/v1/beacon/states/{state_id}/sync_committees
	GetSyncCommitteesAssignments(epoch int, stateID any) (types.StandardSyncCommitteesResponse, error)

	// /eth/v1/config/spec
	GetSpec() (types.StandardSpecResponse, error)

	// /eth/v1/beacon/headers/{block_id}
	GetBlockHeader(blockID any) (types.StandardBeaconHeaderResponse, error)

	// /eth/v1/beacon/states/{state_id}/finality_checkpoints
	GetFinalityCheckpoints(stateID any) (types.StandardFinalityCheckpointsResponse, error)
}
