package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type DutyRepository interface {
	SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error
	SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error
}

type dutyRepository struct{}

func NewDutyRepository() DutyRepository {
	return &dutyRepository{}
}

func (r *dutyRepository) SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {
	return db.BigtableClient.SaveAttestationDuties(attDuties)
}

func (r *dutyRepository) SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error {
	return db.BigtableClient.SaveSyncCommitteeDuties(syncDuties)
}
