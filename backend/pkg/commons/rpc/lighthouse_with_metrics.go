package rpc

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type LighthouseWithMetrics struct {
	client  LighthouseClient
	metrics metrics.MetricsRepository
}

func NewLighthouseWithMetrics(client LighthouseClient, metrics metrics.MetricsRepository) LighthouseWithMetrics {
	return LighthouseWithMetrics{
		client:  client,
		metrics: metrics,
	}
}

func (l *LighthouseWithMetrics) GetNewBlockChan() chan *types.Block {
	timeStart := time.Now()
	originalCh := l.client.GetNewBlockChan()
	wrappedCh := make(chan *types.Block, len(originalCh))

	go func() {
		defer close(wrappedCh)
		firstBlock := true

		for block := range originalCh {
			if firstBlock {
				l.metrics.ObserveClientCallDuration("lighthouse", "get_new_block_chan", time.Since(timeStart))
				firstBlock = false
			}

			wrappedCh <- block
		}
	}()

	return wrappedCh
}

func (l *LighthouseWithMetrics) GetChainHead() (*types.ChainHead, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_chain_head", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetChainHead()
	if err != nil {
		l.metrics.Error("lighthouse_get_chain_head")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetValidatorQueue() (*types.ValidatorQueue, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_validator_queue", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetValidatorQueue()
	if err != nil {
		l.metrics.Error("lighthouse_get_validator_queue")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_epoch_assignments", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetEpochAssignments(epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_epoch_assignments")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetEpochProposerAssignments(epoch uint64) (*constypes.StandardProposerAssignmentsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_epoch_proposer_assignments", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetEpochProposerAssignments(epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_epoch_proposer_assignments")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_validator_state", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetValidatorState(epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_validator_state")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_epoch_data", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetEpochData(epoch, skipHistoricBalances)
	if err != nil {
		l.metrics.Error("lighthouse_get_epoch_data")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_balances_for_epoch", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetBalancesForEpoch(epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_balances_for_epoch")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetBlockByBlockroot(blockroot []byte) (*types.Block, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_block_by_blockroot", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetBlockByBlockroot(blockroot)
	if err != nil {
		l.metrics.Error("lighthouse_get_block_by_blockroot")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_block_header", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetBlockHeader(slot)
	if err != nil {
		l.metrics.Error("lighthouse_get_block_header")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetBlockBySlot(slot uint64) (*types.Block, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_block_by_slot", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetBlockBySlot(slot)
	if err != nil {
		l.metrics.Error("lighthouse_get_block_by_slot")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_validator_participation", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetValidatorParticipation(epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_validator_participation")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_sync_committee", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetSyncCommittee(stateID, epoch)
	if err != nil {
		l.metrics.Error("lighthouse_get_sync_committee")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetBlobSidecars(stateID string) (*constypes.StandardBlobSidecarsResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_blob_sidecars", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetBlobSidecars(stateID)
	if err != nil {
		l.metrics.Error("lighthouse_get_blob_sidecars")
		return nil, err
	}
	return resp, nil
}

func (l *LighthouseWithMetrics) GetStandardBeaconState(stateID any) (*constypes.StandardBeaconStateResponse, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		l.metrics.ObserveClientCallDuration("lighthouse", "get_standard_beacon_state", time.Since(timeStart))
	}(timeStart)

	resp, err := l.client.GetStandardBeaconState(stateID)
	if err != nil {
		l.metrics.Error("lighthouse_get_standard_beacon_state")
		return nil, err
	}
	return resp, nil
}
