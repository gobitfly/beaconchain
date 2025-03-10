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

func (s *SlotExporterBT) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	return s.btClient.GetValidatorBalanceHistory(validators, startEpoch, endEpoch)
}

func (s *SlotExporterBT) SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {
	return s.btClient.SaveAttestationDuties(attDuties)
}

func (s *SlotExporterBT) SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error {
	return s.btClient.SaveSyncCommitteeDuties(syncDuties)
}

func (s *SlotExporterBT) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
	return s.btClient.SaveValidatorBalances(epoch, validators)
}
