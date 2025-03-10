package db

import (
	"sync"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type SlotExporterBTRepository interface {
	GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error)
	SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error
	SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error
	SaveValidatorBalances(epoch uint64, validators []*types.Validator) error
	GetLastAttestationCacheMux() *sync.Mutex
	GetLastAttestationCache() map[uint64]uint64
}

type SlotExporterBT struct {
	client *db.Bigtable
}

func NewSlotExporterBT(btClient *db.Bigtable) *SlotExporterBT {
	return &SlotExporterBT{
		client: btClient,
	}
}

func (s *SlotExporterBT) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	return s.client.GetValidatorBalanceHistory(validators, startEpoch, endEpoch)
}

func (s *SlotExporterBT) SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {
	return s.client.SaveAttestationDuties(attDuties)
}

func (s *SlotExporterBT) SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error {
	return s.client.SaveSyncCommitteeDuties(syncDuties)
}

func (s *SlotExporterBT) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
	return s.client.SaveValidatorBalances(epoch, validators)
}

func (s *SlotExporterBT) GetLastAttestationCacheMux() *sync.Mutex {
	return s.client.LastAttestationCacheMux
}

func (s *SlotExporterBT) GetLastAttestationCache() map[uint64]uint64 {
	return s.client.LastAttestationCache
}
