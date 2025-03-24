package db

import (
	"sync"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
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
	balanceHistory, err := s.client.GetValidatorBalanceHistory(validators, startEpoch, endEpoch)
	if err != nil {
		metrics.Errors.WithLabelValues("slot_exporter_get_validator_balance_history").Inc()
		return nil, err
	}
	return balanceHistory, nil
}

func (s *SlotExporterBT) SaveAttestationDuties(attDuties map[types.Slot]map[types.ValidatorIndex][]types.Slot) error {
	err := s.client.SaveAttestationDuties(attDuties)
	if err != nil {
		metrics.Errors.WithLabelValues("slot_exporter_save_attestation_duties").Inc()
		return err
	}
	return nil
}

func (s *SlotExporterBT) SaveSyncCommitteeDuties(syncDuties map[types.Slot]map[types.ValidatorIndex]bool) error {
	err := s.client.SaveSyncCommitteeDuties(syncDuties)
	if err != nil {
		metrics.Errors.WithLabelValues("slot_exporter_save_committee_duties").Inc()
		return err
	}
	return nil
}

func (s *SlotExporterBT) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
	err := s.client.SaveValidatorBalances(epoch, validators)
	if err != nil {
		metrics.Errors.WithLabelValues("slot_exporter_save_validator_balances").Inc()
		return err
	}
	return nil
}

func (s *SlotExporterBT) GetLastAttestationCacheMux() *sync.Mutex {
	return s.client.LastAttestationCacheMux
}

func (s *SlotExporterBT) GetLastAttestationCache() map[uint64]uint64 {
	return s.client.LastAttestationCache
}
