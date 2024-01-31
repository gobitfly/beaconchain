package clnode

import "github.com/gobitfly/beaconchain/commons/utils"

type Retriever struct {
	RetrieverInt
	ChainConfig ClChainConfig
}
type RetrieverInt interface {
	GetSlot(slot int) (GetBeaconSlotResponse, error)

	GetValidators(slot int) (GetValidatorsResponse, error)

	GetPropoalAssignments(epoch int) (GetProposerAssignmentsResponse, error)

	GetPropoalRewards(slot int) (GetBlockRewardsResponse, error)

	GetSyncRewards(slot int) (GetSyncCommitteeRewardsResponse, error)

	GetAttestationRewards(epoch int) (GetAttestationRewardsResponse, error)

	GetSyncCommitteesAssignments(epoch int, slot int64) (GetSyncCommitteeAssignmentsResponse, error)

	GetSpec() (GetSpecResponse, error)

	GetBlockHeader(block_id string) (StandardBeaconHeaderResponse, error)

	GetFinalityCheckpoints(state_id string) (StandardFinalityCheckpointsResponse, error)
}

func (r *Retriever) GetChainHead() (*ChainHead, error) {
	parsedHead, err := r.GetBlockHeader("head")
	if err != nil {
		return &ChainHead{}, err
	}

	id := parsedHead.Data.Header.Message.StateRoot
	if parsedHead.Data.Header.Message.Slot == 0 {
		id = "genesis"
	}

	parsedFinality, err := r.GetFinalityCheckpoints(id)
	if err != nil {
		return &ChainHead{}, err
	}

	// The epoch in the Finalized Object is not the finalized epoch, but the epoch for the checkpoint - the 'real' finalized epoch is the one before
	var finalizedEpoch = uint64(parsedFinality.Data.Finalized.Epoch)
	if finalizedEpoch > 0 {
		finalizedEpoch--
	}

	finalizedSlot := (finalizedEpoch + 1) * r.ChainConfig.SlotsPerEpoch // The first Slot of the next epoch is finalized.
	if finalizedEpoch == 0 && parsedFinality.Data.Finalized.Root == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		finalizedSlot = 0
	}
	return &ChainHead{
		HeadSlot:                   uint64(parsedHead.Data.Header.Message.Slot),
		HeadEpoch:                  uint64(parsedHead.Data.Header.Message.Slot) / r.ChainConfig.SlotsPerEpoch,
		HeadBlockRoot:              utils.MustParseHex(parsedHead.Data.Root),
		FinalizedSlot:              finalizedSlot,
		FinalizedEpoch:             finalizedEpoch,
		FinalizedBlockRoot:         utils.MustParseHex(parsedFinality.Data.Finalized.Root),
		JustifiedSlot:              uint64(parsedFinality.Data.CurrentJustified.Epoch) * r.ChainConfig.SlotsPerEpoch,
		JustifiedEpoch:             uint64(parsedFinality.Data.CurrentJustified.Epoch),
		JustifiedBlockRoot:         utils.MustParseHex(parsedFinality.Data.CurrentJustified.Root),
		PreviousJustifiedSlot:      uint64(parsedFinality.Data.PreviousJustified.Epoch) * r.ChainConfig.SlotsPerEpoch,
		PreviousJustifiedEpoch:     uint64(parsedFinality.Data.PreviousJustified.Epoch),
		PreviousJustifiedBlockRoot: utils.MustParseHex(parsedFinality.Data.PreviousJustified.Root),
	}, nil
}
