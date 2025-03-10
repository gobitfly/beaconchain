package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type SlotExporterBTRepository interface {
	GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error)
	SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error
	SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error
	SaveValidatorBalances(epoch uint64, validators []*types.Validator) error
}

type SlotExporterBT struct {
	btClient *db.Bigtable
}

func NewSlotExporterBT(btClient *db.Bigtable) *SlotExporterBT {
	return &SlotExporterBT{
		btClient: btClient,
	}
}

func (r *SlotExporterBT) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	return db.BigtableClient.GetValidatorBalanceHistory(validators, startEpoch, endEpoch)
}

func (r *SlotExporterBT) SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {
	return db.BigtableClient.SaveAttestationDuties(attDuties)
}

func (r *SlotExporterBT) SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error {
	return db.BigtableClient.SaveSyncCommitteeDuties(syncDuties)
}

func (r *SlotExporterBT) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
	return db.BigtableClient.SaveValidatorBalances(epoch, validators)
}
