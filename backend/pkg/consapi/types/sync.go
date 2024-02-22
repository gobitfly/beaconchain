package types

type StandardSyncCommittee struct {
	Validators          []string   `json:"validators"`
	ValidatorAggregates [][]string `json:"validator_aggregates"`
}

// /eth/v1/beacon/states/{state_id}/sync_committees
type StandardSyncCommitteesResponse struct {
	Data                StandardSyncCommittee `json:"data"`
	ExecutionOptimistic bool                  `json:"execution_optimistic"`
	Finalized           bool                  `json:"finalized"`
}

// /eth/v1/beacon/states/%v/committees
type StandardCommitteesResponse struct {
	Data []StandardCommitteeEntry `json:"data"`
}

type StandardCommitteeEntry struct {
	Index      uint64   `json:"index,string"`
	Slot       uint64   `json:"slot,string"`
	Validators []string `json:"validators"`
}
