package clnode

type Retriever interface {
	GetSlot(slot int) (GetBeaconSlotResponse, error)

	GetValidators(slot int) (GetValidatorsResponse, error)

	GetPropoalAssignments(epoch int) (GetProposerAssignmentsResponse, error)

	GetPropoalRewards(slot int) (GetBlockRewardsResponse, error)

	GetSyncRewards(slot int) (GetSyncCommitteeRewardsResponse, error)

	GetAttestationRewards(epoch int) (GetAttestationRewardsResponse, error)

	GetSyncCommitteesAssignments(epoch int, slot int64) (GetSyncCommitteeAssignmentsResponse, error)

	GetSpec() (GetSpecResponse, error)
}
