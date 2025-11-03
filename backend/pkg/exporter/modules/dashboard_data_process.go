package modules

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/go-bitfield"
	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
)

func (d *dashboardData) processRunner(data *MultiEpochData, tar *[]types.VDBDataEpochColumns, epochs []edb.EpochMetadata) error {
	d.log.Info("starting processRunner")
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_overall").Observe(time.Since(start).Seconds())
	}()
	var err error
	seenInsertIds := []uuid.UUID{*epochs[0].InsertBatchID}
	currentTarIndex := 0
	bulkValiCount := 0
	bulkEpochsContained := make([]uint64, 0)
	// sanity check, epochs in data.epochsBaseData.epochs should be identical to []epochs.Epoch
	for i, epoch := range data.epochBasedData.epochs {
		if epoch != epochs[i].Epoch {
			return fmt.Errorf("epoch mismatch, expected %d, got %d", epochs[i].Epoch, epoch)
		}
	}

	// prepare tar
	for i, epoch := range epochs {
		insertId := *epoch.InsertBatchID
		// debug log
		d.log.Infof("epoch %d, insertId %s", epoch.Epoch, insertId)
		if insertId != seenInsertIds[len(seenInsertIds)-1] {
			d.log.Infof("allocating tar insert id %s with %d entries", seenInsertIds[len(seenInsertIds)-1], bulkValiCount)
			(*tar)[currentTarIndex], err = types.NewVDBDataEpochColumns(bulkValiCount)
			if err != nil {
				return errors.Wrap(err, "failed to allocate new VDBDataEpochColumns")
			}
			(*tar)[currentTarIndex].EpochsContained = bulkEpochsContained
			(*tar)[currentTarIndex].InsertBatchID = []uuid.UUID{seenInsertIds[len(seenInsertIds)-1]} // important to use insertID because pointers
			currentTarIndex++
			bulkEpochsContained = make([]uint64, 0)
			bulkValiCount = 0
			seenInsertIds = append(seenInsertIds, insertId)
		}
		d.log.Infof("adding epoch %d to tar insert id %s", epoch.Epoch, insertId)
		bulkEpochsContained = append(bulkEpochsContained, epoch.Epoch)
		data.epochBasedData.tarIndices[i] = currentTarIndex
		data.epochBasedData.tarOffsets[i] = bulkValiCount
		bulkValiCount += len(data.epochBasedData.validatorStates[int64(epoch.Epoch)].Data)
	}
	if bulkValiCount > 0 {
		d.log.Infof("leftover valis, allocating tar index %d with %d entries", currentTarIndex, bulkValiCount)
		(*tar)[currentTarIndex], err = types.NewVDBDataEpochColumns(bulkValiCount)
		if err != nil {
			return errors.Wrap(err, "failed to allocate new VDBDataEpochColumns")
		}
		(*tar)[currentTarIndex].EpochsContained = bulkEpochsContained
		(*tar)[currentTarIndex].InsertBatchID = []uuid.UUID{seenInsertIds[len(seenInsertIds)-1]} // important to use insertID because pointers
	}
	for i := range *tar {
		d.log.Warnf("tar %d, insertId %s, epochs %v, seendInsertIds %v", i, (*tar)[i].InsertBatchID, (*tar)[i].EpochsContained, seenInsertIds)
	}
	d.log.Infof("prepared tar with %d entries", len(*tar))
	// errgroup
	g := &errgroup.Group{}
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_validator_states_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processValidatorStates(data, tar)
		if err != nil {
			return fmt.Errorf("error in processValidatorStates: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_scheduled_attestations_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processScheduledAttestations(data, tar)
		if err != nil {
			return fmt.Errorf("error in processScheduledAttestations: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_blocks_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processBlocks(data, tar)
		if err != nil {
			return fmt.Errorf("error in processBlocks: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_deposits_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processDeposits(data, tar)
		if err != nil {
			return fmt.Errorf("error in processDeposits: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_deposits_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processElectraDeposits(data, tar)
		if err != nil {
			return fmt.Errorf("error in processElectraDeposits: %w", err)
		}
		return nil
	})
	// electra consolidations
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_consolidations_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processElectraConsolidations(data, tar)
		if err != nil {
			return fmt.Errorf("error in processElectraConsolidations: %w", err)
		}
		return nil
	})
	// RemovedExcessBalanceEvent
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_removed_excess_balance_events_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processElectraRemovedExcessBalanceEvents(data, tar)
		if err != nil {
			return fmt.Errorf("error in processElectraRemovedExcessBalanceEvents: %w", err)
		}
		return nil
	})

	// force sequential operation of attestation rewards and proposal rewards
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		d.phase0HotfixMutex.Lock()
	}
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_attestation_rewards_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processAttestationRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processAttestationRewards: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_proposal_rewards_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processWithdrawals(data, tar)
		if err != nil {
			return fmt.Errorf("error in processWithdrawals: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_attestations_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processAttestations(data, tar)
		if err != nil {
			return fmt.Errorf("error in processAttestations: %w", err)
		}
		return nil
	})
	// processExpectedSyncPeriods
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_expected_sync_periods_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processExpectedSyncPeriods(data, tar)
		if err != nil {
			return fmt.Errorf("error in processExpectedSyncPeriods: %w", err)
		}
		return nil
	})
	// processSyncVotes
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_sync_votes_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processSyncVotes(data, tar)
		if err != nil {
			return fmt.Errorf("error in processSyncVotes: %w", err)
		}
		d.log.Infof("processed sync votes in %v", time.Since(start))
		return nil
	})
	// processProposalRewards
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_proposal_rewards_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processProposalRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processProposalRewards: %w", err)
		}
		return nil
	})
	// processSyncCommitteeRewards
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_sync_committee_rewards_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processSyncCommitteeRewards(data, tar)
		if err != nil {
			return fmt.Errorf("error in processSyncCommitteeRewards: %w", err)
		}
		return nil
	})
	// processBlocksExpected
	g.Go(func() error {
		start := time.Now()
		defer func() {
			metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_blocks_expected_overall").Observe(time.Since(start).Seconds())
		}()
		err := d.processBlocksExpected(data, tar)
		if err != nil {
			return fmt.Errorf("error in processBlocksExpected: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error in processRunner: %w", err)
	}

	return nil
}

func (d *dashboardData) processValidatorStates(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		iEpoch := int64(epoch)
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		g.Go(func() error {
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_validator_states_single").Observe(time.Since(start).Seconds())
			}()
			ts := utils.EpochToTime(epoch)
			startSate := data.epochBasedData.validatorStates[iEpoch-1]
			endState := data.epochBasedData.validatorStates[iEpoch]
			nextState := data.epochBasedData.validatorStates[iEpoch+1]
			for j := range endState.Data {
				(*tar)[tI].ValidatorIndex[tO+j] = uint64(j)
				(*tar)[tI].Epoch[tO+j] = iEpoch
				(*tar)[tI].EpochTimestamp[tO+j] = &ts
				(*tar)[tI].BalanceEnd[tO+j] = int64(endState.Data[j].Balance)
				(*tar)[tI].BalanceEffectiveEnd[tO+j] = int64(endState.Data[j].EffectiveBalance)
				// balance effective attestations needs to look 2 epochs ahead because attestation processing happens in the transition from n+1 => n+2
				(*tar)[tI].BalanceEffectiveAttestations[tO+j] = int64(nextState.Data[j].EffectiveBalance)
				// do NOT set the slashed flag here. it is set by the block processing
			}
			if epoch == 0 {
				// we do not set the start state for epoch 0, this is because validators that join
				// after the genesis epoch will always have a start balance of 0 (they didn't exist
				// in the previous epoch yet) and we want to emulate the same behavior
				return nil
			}
			for j := range startSate.Data {
				(*tar)[tI].BalanceStart[tO+j] = int64(startSate.Data[j].Balance)
				(*tar)[tI].BalanceEffectiveStart[tO+j] = int64(startSate.Data[j].EffectiveBalance)
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processScheduledAttestations(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_scheduled_attestations_single").Observe(time.Since(now).Seconds())
			}()
			// pre-init should not be required, is pointer and defaults to null
			for slot := startSlot; slot < endSlot; slot++ {
				// fetch from attestati	on assignments
				if _, ok := data.slotBasedData.assignments.attestationAssignments[slot]; !ok {
					// error, should never happen, we should always have assignments
					return fmt.Errorf("no attestation assignments for slot %d", slot)
				}
				for committee, validatorIndices := range data.slotBasedData.assignments.attestationAssignments[slot] {
					for committeeIndex, validatorIndex := range validatorIndices {
						(*tar)[tI].AttestationsScheduled[tO+int(validatorIndex)]++
						(*tar)[tI].AttestationAssignmentsSlot[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsSlot[tO+int(validatorIndex)],
							int64(slot),
						)
						(*tar)[tI].AttestationAssignmentsCommittee[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsCommittee[tO+int(validatorIndex)],
							int64(committee),
						)
						(*tar)[tI].AttestationAssignmentsIndex[tO+int(validatorIndex)] = append(
							(*tar)[tI].AttestationAssignmentsIndex[tO+int(validatorIndex)],
							int64(committeeIndex),
						)
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processBlocks(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_blocks_single").Observe(time.Since(now).Seconds())
			}()
			for slot := startSlot; slot < endSlot; slot++ {
				// skip slot 0, as nobody proposed it
				if slot == 0 {
					continue
				}
				proposer := data.slotBasedData.assignments.blockAssignments[slot]
				(*tar)[tI].BlocksStatusSlot[tO+int(proposer)] = append((*tar)[tI].BlocksStatusSlot[tO+int(proposer)], int64(slot))

				if _, ok := data.slotBasedData.blocks[slot]; !ok {
					(*tar)[tI].BlocksStatusProposed[tO+int(proposer)] = append((*tar)[tI].BlocksStatusProposed[tO+int(proposer)], false)
					continue
				}

				(*tar)[tI].BlocksStatusProposed[tO+int(proposer)] = append((*tar)[tI].BlocksStatusProposed[tO+int(proposer)], true)

				for _, index := range data.slotBasedData.blocks[slot].SlashedIndices {
					(*tar)[tI].Slashed[tO+int(index)] = true
					(*tar)[tI].BlocksSlashingCount[uint64(tO)+data.slotBasedData.blocks[slot].ProposerIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processDeposits(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	if d.signingDomain == nil {
		domain, err := utils.GetSigningDomain()
		if err != nil {
			return err
		}
		d.signingDomain = domain
	}
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch >= utils.Config.ClConfig.ElectraForkEpoch {
			// skip deposits after the electra fork epoch because they will get converted to electra deposits instead
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_deposits_single").Observe(time.Since(now).Seconds())
			}()
			// genesis deposits
			if epoch == 0 {
				for j := range data.epochBasedData.validatorStates[0].Data {
					(*tar)[tI].DepositsCount[tO+j] = 1
					(*tar)[tI].DepositsAmount[tO+j] = int64(data.epochBasedData.validatorStates[0].Data[j].Balance)
				}
			}
			for jj := startSlot; jj < endSlot; jj++ {
				slot := jj
				if _, ok := data.slotBasedData.blocks[slot]; !ok {
					// nothing to do
					continue
				}

				for depositIndex, depositData := range data.slotBasedData.blocks[slot].Deposits {
					index, indexExists := data.validatorBasedData.validatorIndices[string(depositData.Data.Pubkey)]
					if !indexExists {
						// skip
						d.log.Infof("validator not found for deposit at index %d in slot %v", depositIndex, data.slotBasedData.blocks[slot].Slot)
						continue
					}
					err := utils.VerifyDepositSignature(&phase0.DepositData{
						PublicKey:             phase0.BLSPubKey(depositData.Data.Pubkey),
						WithdrawalCredentials: depositData.Data.WithdrawalCredentials,
						Amount:                phase0.Gwei(depositData.Data.Amount),
						Signature:             phase0.BLSSignature(depositData.Data.Signature),
					}, d.signingDomain)
					if err != nil {
						d.log.Error(fmt.Errorf("deposit at index %d in slot %v is invalid: %v (signature: %s)", depositIndex, data.slotBasedData.blocks[slot].Slot, err, depositData.Data.Signature), "", 0)
						// this fails if there was no valid deposits in our entire epoch range
						// so we can skip this deposit
						if !indexExists {
							d.log.Infof("validator did not have a valid deposit within epoch range - assuming discard of deposit")
							continue
						}
						// validator has been created. check if it already exists in the current epoch. if yes, all deposits can be assumed as valid
						// check if it existed in the previous state
						var existedPreviousEpoch bool
						if uint64(len(data.epochBasedData.validatorStates[int64(epoch)].Data)) > index {
							existedPreviousEpoch = true
						}

						if (*tar)[tI].DepositsCount[uint64(tO)+index] == 0 && !existedPreviousEpoch {
							d.log.Infof("validator did not have a valid deposit before current deposit within epoch and did not exist in previous epoch - assuming discard of deposit")
							continue
						}
					}
					(*tar)[tI].DepositsAmount[uint64(tO)+index] += int64(depositData.Data.Amount)

					(*tar)[tI].DepositsCount[uint64(tO)+index]++
					d.log.Infof("processed deposit at index %d in slot %v", depositIndex, data.slotBasedData.blocks[slot].Slot)
				}
			}
			return nil
		})
	}

	return g.Wait()
}

/*
deposit processing:
this is technically not needed with how the new deposit events work, but since the current hoodi events are already in this state we have to support it for now
the tldr is that the we get depositProcessedEvent with validsig = false for rejected deposits, but also for accepted deposits after a valid one
so the logic should be:
  - check if the validator already existed before the current epochs state transition (basically from where we get balance start)
  - if it did, accept all deposits as valid
  - if not, check if the validator exists in the next epochs state transition:
  - if it does not, reject all deposits as invalid
  - if it does, process all deposits sequentially, rejecting all the ones with validsig = false that are before the first valid one
*/
func (d *dashboardData) processElectraDeposits(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_deposits_single").Observe(time.Since(now).Seconds())
			}()
			if _, ok := data.epochBasedData.electraDeposits[epoch]; !ok {
				// nothing to do
				return nil
			}
			// group by pubkey, retain order
			deposits := make(map[string][]constypes.ElectraDeposit)
			for _, deposit := range data.epochBasedData.electraDeposits[epoch] {
				if _, ok := deposits[string(deposit.Pubkey)]; !ok {
					deposits[string(deposit.Pubkey)] = make([]constypes.ElectraDeposit, 0)
				}
				deposits[string(deposit.Pubkey)] = append(deposits[string(deposit.Pubkey)], deposit)
			}
			// now we have a map of pubkey => deposits
			// loop over the map and process the deposits
			epochStart := int64(epoch) - 1
			epochEnd := int64(epoch)
			for pubkey, deposits := range deposits {
				// try to resolve the pubkey => index, if it fails assume all deposits are invalid
				// the validatorIndices mapping gets populated with the validators of all epochs that are currently being processed.
				// so if the validator is not in the map, we can assume all deposits we will process are invalid
				index, indexExists := data.validatorBasedData.validatorIndices[pubkey]
				if !indexExists {
					d.log.Tracef("validator %s does not exist in validator indices map, assuming all deposits are invalid", pubkey)
					continue
				}
				// check if the validator exists at the end of the epoch
				if uint64(len(data.epochBasedData.validatorStates[epochEnd].Data)) <= index {
					d.log.Tracef("validator %s does not exist at the end of the epoch, assuming all deposits are invalid", pubkey)
					continue
				}
				// check if the validator already existed at the start of the epoch
				if uint64(len(data.epochBasedData.validatorStates[epochStart].Data)) > index {
					for _, deposit := range deposits {
						(*tar)[tI].DepositsAmount[uint64(tO)+index] += deposit.Amount
						(*tar)[tI].DepositsCount[uint64(tO)+index]++
						d.log.Tracef("processed electra deposit of %d GWEI for validator %d in epoch %d (assumed valid, validator did exist at start of epoch)", deposit.Amount, index, epoch)
					}
					continue
				}
				// validator exists at the end of the epoch (but not the start), so at least one of the deposits must be valid
				sawValid := false
				for _, deposit := range deposits {
					if deposit.SignatureValid {
						sawValid = true
					}
					if !sawValid {
						// no valid deposit
						d.log.Debugf("processed electra deposit of %d GWEI for validator %d in epoch %d (discarded, no valid deposit yet)", deposit.Amount, index, epoch)
						continue
					}
					// this deposit is valid, so we can process it
					(*tar)[tI].DepositsAmount[uint64(tO)+index] += deposit.Amount
					(*tar)[tI].DepositsCount[uint64(tO)+index]++
					d.log.Tracef("processed electra deposit of %d GWEI for validator %d in epoch %d (at or after first valid deposit)", deposit.Amount, index, epoch)
				}
				if !sawValid {
					return fmt.Errorf("no valid deposits for validator %s in epoch %d despite it transitioning from non-existent to existent", pubkey, epoch)
				}
			}
			return nil
		})
	}

	return g.Wait()
}

func (d *dashboardData) processElectraConsolidations(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_consolidations_single").Observe(time.Since(now).Seconds())
			}()
			if _, ok := data.epochBasedData.electraConsolidations[epoch]; !ok {
				// nothing to do
				return nil
			}
			for _, consolidation := range data.epochBasedData.electraConsolidations[epoch] {
				targetIndex, ok := data.validatorBasedData.validatorIndices[string(consolidation.TargetPubkey)]
				if !ok {
					return fmt.Errorf("target validator %s not found in validator indices map", string(consolidation.TargetPubkey))
				}
				sourceIndex, ok := data.validatorBasedData.validatorIndices[string(consolidation.SourcePubkey)]
				if !ok {
					return fmt.Errorf("source validator %s not found in validator indices map", string(consolidation.SourcePubkey))
				}
				(*tar)[tI].ConsolidationsIncomingAmount[uint64(tO)+targetIndex] += int64(consolidation.Amount)
				(*tar)[tI].ConsolidationsIncomingCount[uint64(tO)+targetIndex]++
				(*tar)[tI].ConsolidationsOutgoingAmount[uint64(tO)+sourceIndex] += int64(consolidation.Amount)
				(*tar)[tI].ConsolidationsOutgoingCount[uint64(tO)+sourceIndex]++
				// source => target
				d.log.Tracef("processed electra consolidation: %d => %d %d GWEI in epoch %d", consolidation.SourcePubkey, consolidation.TargetPubkey, consolidation.Amount, epoch)
			}
			return nil
		})
	}

	return g.Wait()
}

func (d *dashboardData) processElectraRemovedExcessBalanceEvents(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_electra_removed_excess_balance_events_single").Observe(time.Since(now).Seconds())
			}()
			if _, ok := data.epochBasedData.electraRemovedExcessBalanceEvents[epoch]; !ok {
				// nothing to do
				return nil
			}
			for _, event := range data.epochBasedData.electraRemovedExcessBalanceEvents[epoch] {
				validatorIndex, ok := data.validatorBasedData.validatorIndices[string(event.ValidatorPubkey)]
				if !ok {
					return fmt.Errorf("validator %s not found in validator indices map", string(event.ValidatorPubkey))
				}
				(*tar)[tI].WithdrawalsAmount[uint64(tO)+validatorIndex] += int64(event.Amount)
				(*tar)[tI].WithdrawalsCount[uint64(tO)+validatorIndex]++
				d.log.Tracef("processed electra removed excess balance event: %d %d GWEI in epoch %d", event.ValidatorPubkey, event.Amount, epoch)
			}
			return nil
		})
	}

	return g.Wait()
}

func (d *dashboardData) processAttestationRewards(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		d.phase0HotfixMutex.Lock()
		defer d.phase0HotfixMutex.Unlock()
	}
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := data.epochBasedData.tarOffsets[i]
		if data.epochBasedData.rewards.attestationRewards[epoch] == nil {
			return fmt.Errorf("no rewards for epoch %d", epoch)
		}
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_attestation_rewards_single").Observe(time.Since(now).Seconds())
			}()
			nextState := data.epochBasedData.validatorStates[int64(epoch)+1]
			// calculate max per committee x attestation slot
			// array of slotsperepoch length, then array of committee per slot length. no maps because maps are slow
			hyperlocalizedMax := make([][]int64, int(utils.Config.Chain.ClConfig.SlotsPerEpoch))
			if utils.Config.Chain.ClConfig.TargetCommitteeSize == 0 {
				utils.Config.Chain.ClConfig.TargetCommitteeSize = 64
			}
			// pre-init committees
			for slot := 0; slot < int(utils.Config.Chain.ClConfig.SlotsPerEpoch); slot++ {
				hyperlocalizedMax[slot] = make([]int64, utils.Config.Chain.ClConfig.TargetCommitteeSize)
			}
			// validator_index => [slot, committee] mapping
			validatorSlotMap := make([]*struct {
				Slot      int64
				Committee int64
			}, len(data.epochBasedData.validatorStates[int64(epoch)].Data))
			startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
			endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
			for slot := startSlot; slot < endSlot; slot++ {
				if _, ok := data.slotBasedData.assignments.attestationAssignments[slot]; !ok {
					return fmt.Errorf("no attestation assignments for slot %d", slot)
				}
				for committee, validatorIndices := range data.slotBasedData.assignments.attestationAssignments[slot] {
					for _, validatorIndex := range validatorIndices {
						validatorSlotMap[validatorIndex] = &struct {
							Slot      int64
							Committee int64
						}{
							Slot:      int64(slot - startSlot),
							Committee: int64(committee),
						}
					}
				}
			}

			for _, ar := range data.epochBasedData.rewards.attestationRewards[epoch] {
				// there is a bug in lighthouse which causes the rewards api to return rewards for validators
				// that have been deposited but arent active yet for some reason
				// we can safely ignore these
				if ar.ValidatorIndex >= uint64(len(validatorSlotMap)) {
					d.log.Tracef("skipping reward for validator %d in epoch %d, outside of assignment range", ar.ValidatorIndex, epoch)
					continue
				}
				attData := validatorSlotMap[ar.ValidatorIndex]
				if attData == nil {
					d.log.Tracef("skipping reward for validator %d in epoch %d, no attestation assignment", ar.ValidatorIndex, epoch)
					continue
				}

				valiIndextO := uint64(tO) + ar.ValidatorIndex
				// ideal rewards
				// we need to use the effective balance of epoch n+1 because the attestation processing for epoch n happens in the transition from n+1 => n+2
				effectiveBalance := nextState.Data[ar.ValidatorIndex].EffectiveBalance
				idealReward, ok := data.epochBasedData.rewards.attestationIdealRewards[epoch][effectiveBalance]
				if !ok {
					return fmt.Errorf("no ideal reward for validator %d in epoch %d", valiIndextO, epoch)
				}
				(*tar)[tI].AttestationsIdealHeadReward[valiIndextO] = int64(idealReward.Head)
				(*tar)[tI].AttestationsIdealSourceReward[valiIndextO] = int64(idealReward.Source)
				(*tar)[tI].AttestationsIdealTargetReward[valiIndextO] = int64(idealReward.Target)
				(*tar)[tI].AttestationsIdealInclusionReward[valiIndextO] = int64(idealReward.InclusionDelay)
				(*tar)[tI].AttestationsIdealInactivityReward[valiIndextO] = int64(idealReward.Inactivity)

				(*tar)[tI].AttestationsHeadReward[valiIndextO] = int64(ar.Head)
				(*tar)[tI].AttestationsSourceReward[valiIndextO] = int64(ar.Source)
				(*tar)[tI].AttestationsTargetReward[valiIndextO] = int64(ar.Target)
				// phase0 hotfix - cap inclusion delay reward at ideal inclusion delay reward
				if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch && int64(ar.InclusionDelay) > int64(idealReward.InclusionDelay) {
					ar.InclusionDelay = idealReward.InclusionDelay
				}
				(*tar)[tI].AttestationsInclusionReward[valiIndextO] = int64(ar.InclusionDelay)
				(*tar)[tI].AttestationsInactivityReward[valiIndextO] = int64(ar.Inactivity)
				// total
				total := int64(0)
				for _, r := range []int64{int64(ar.Head), int64(ar.Source), int64(ar.Target), int64(ar.InclusionDelay), int64(ar.Inactivity)} {
					total += r
				}
				// generate hyper localized max
				if hyperlocalizedMax[attData.Slot][attData.Committee] < total {
					hyperlocalizedMax[attData.Slot][attData.Committee] = total
				}
			}
			// localize to slot
			localizedMax := make([]int64, int(utils.Config.Chain.ClConfig.SlotsPerEpoch))
			for slot := 0; slot < int(utils.Config.Chain.ClConfig.SlotsPerEpoch); slot++ {
				r := int64(0)
				for _, d := range hyperlocalizedMax[slot] {
					if d > r {
						r = d
					}
				}
				localizedMax[slot] = r
			}
			// write hyper localized max by iterating over all validators and looking up the max
			for _, ar := range data.epochBasedData.rewards.attestationRewards[epoch] {
				if ar.ValidatorIndex >= uint64(len(validatorSlotMap)) {
					continue
				}
				valiSlot := validatorSlotMap[ar.ValidatorIndex]

				// happens when the validator has been added to state but doesnt have a duty yet or has been slashed
				// fine to ignore
				if valiSlot == nil {
					continue
				}
				valiIndextO := uint64(tO) + ar.ValidatorIndex
				(*tar)[tI].AttestationsHyperLocalizedMaxReward[valiIndextO] = hyperlocalizedMax[valiSlot.Slot][valiSlot.Committee]
				(*tar)[tI].AttestationsLocalizedMaxReward[valiIndextO] = localizedMax[valiSlot.Slot]
			}
			return nil
		})
	}
	return g.Wait()
}

// withdrawals
func (d *dashboardData) processWithdrawals(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.CapellaForkEpoch {
			// d.log.Infof("skipping withdrawals for epoch %d (before capella)", epoch)
			// no withdrawals before cappella
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_withdrawals_single").Observe(time.Since(now).Seconds())
			}()
			for j := startSlot; j < endSlot; j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				for _, withdrawal := range data.slotBasedData.blocks[j].Withdrawals {
					(*tar)[tI].WithdrawalsAmount[tO+withdrawal.ValidatorIndex] = int64(withdrawal.Amount)
					(*tar)[tI].WithdrawalsCount[tO+withdrawal.ValidatorIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

// Function to filter an array of integers using a bit mask with reduced allocations
func filterArrayUsingBitMask(arr []uint64, bitmask []byte) []uint64 {
	result := make([]uint64, 0, len(arr))

	for i := 0; i < len(arr); i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		if byteIndex < len(bitmask) {
			if (bitmask[byteIndex] & (1 << bitIndex)) != 0 {
				result = append(result, arr[i])
			}
		} else {
			panic("bitmask out of bounds")
		}
	}

	return result
}

func IntegerSquareRoot(n uint64) uint64 {
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

// attestations
func (d *dashboardData) processAttestations(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	blockRoots := make(map[uint64]hexutil.Bytes)
	blockValidityMap := make(map[uint64]int64, len(data.slotBasedData.blocks))
	// loop through slots. if a slot is missing reuse the previous slot. skip if the first slot is missing just skip
	var lastBlockHash *hexutil.Bytes
	var previousValidHash hexutil.Bytes
	var toBeFilled []uint64
	for _, e := range data.epochBasedData.epochs {
		epoch := e
		start := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		end := start + utils.Config.Chain.ClConfig.SlotsPerEpoch
		for j := start; j < end; j++ {
			if _, ok := data.slotBasedData.blocks[j]; !ok {
				blockValidityMap[j] = 0
				// check if we have a previous block
				if lastBlockHash == nil {
					toBeFilled = append(toBeFilled, j)
					continue
				}
				blockRoots[j] = *lastBlockHash
				continue
			}
			blockValidityMap[j] = 1
			if lastBlockHash == nil {
				d.log.Tracef("slot %d, setting previousValidHash to %s", j, data.slotBasedData.blocks[j].ParentRoot)
				previousValidHash = data.slotBasedData.blocks[j].ParentRoot
			}
			a := data.slotBasedData.blocks[j].BlockRoot
			lastBlockHash = &a
			blockRoots[j] = *lastBlockHash
		}
	}
	// do blockValidityMap for n+1 epoch data
	lookaheadEpoch := data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] + 1
	start := lookaheadEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	end := start + utils.Config.Chain.ClConfig.SlotsPerEpoch
	for j := start; j < end; j++ {
		if _, ok := data.slotBasedData.blocks[j]; !ok {
			blockValidityMap[j] = 0
			continue
		}
		blockValidityMap[j] = 1
		if lastBlockHash == nil {
			d.log.Tracef("slot %d, setting previousValidHash to %s (lookahead)", j, data.slotBasedData.blocks[j].ParentRoot)
			previousValidHash = data.slotBasedData.blocks[j].ParentRoot
		}
		a := data.slotBasedData.blocks[j].BlockRoot
		lastBlockHash = &a
		blockRoots[j] = *lastBlockHash
	}
	// if there were no blocks between the start epoch and the lookahead, there were also no attestations
	// … because attestations are included in blocks
	// do not error, just warn and return
	if lastBlockHash == nil {
		d.log.Warnf("no blocks found between epoch %d and epoch %d, skipping attestation processing (there can be no attestations without blocks)", data.epochBasedData.epochs[0], lookaheadEpoch)
		return nil
	}
	// fill toBeFilled
	for _, slot := range toBeFilled {
		blockRoots[slot] = previousValidHash
	}

	for i, e := range data.epochBasedData.epochs {
		epoch := e
		//debugCounters := make(map[string]int)
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		squareSlotsPerEpoch := IntegerSquareRoot(utils.Config.Chain.ClConfig.SlotsPerEpoch)
		// cache array of committee indexes (0,1,…)
		committeeIndexes := make([]uint64, utils.Config.Chain.ClConfig.MaxCommitteesPerSlot)
		for i := range committeeIndexes {
			committeeIndexes[i] = uint64(i)
		}
		g.Go(func() error {
			start := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_attestations_single").Observe(time.Since(start).Seconds())
			}()
			for j := startSlot; j < startSlot+(utils.Config.Chain.ClConfig.SlotsPerEpoch*2); j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				// attestations
				for _, att := range data.slotBasedData.blocks[j].Attestations {
					// ignore if slot is not within our epoch of interest
					if att.Data.Slot < startSlot || att.Data.Slot >= endSlot {
						d.log.Tracef("skipping attestation in slot %d, not within epoch bounds %d-%d", att.Data.Slot, startSlot, endSlot)
						continue
					}
					// hardfork handling
					var v []uint64
					aggregateBits := bitfield.Bitlist(att.AggregationBits).BytesNoTrim()

					if epoch >= utils.Config.Chain.ClConfig.ElectraForkEpoch {
						// use committee_bits map to generate
						//var newAssignmentArray []uint64
						committees := filterArrayUsingBitMask(committeeIndexes, att.CommitteeBits)
						newAssignmentArray := make([]uint64, 0, len(committees)*int(utils.Config.Chain.ClConfig.MaxValidatorsPerCommittee))
						for _, committee := range committees {
							if len(data.slotBasedData.assignments.attestationAssignments[att.Data.Slot][committee]) == 0 {
								// bad
								return fmt.Errorf("no validators found in attestation assignment for slot %d and committee %d", att.Data.Slot, committee)
							}
							newAssignmentArray = append(newAssignmentArray, data.slotBasedData.assignments.attestationAssignments[att.Data.Slot][committee]...)
						}
						v = filterArrayUsingBitMask(newAssignmentArray, aggregateBits)
					} else {
						v = filterArrayUsingBitMask(data.slotBasedData.assignments.attestationAssignments[att.Data.Slot][att.Data.Index], aggregateBits)
					}
					// debug print the bitmask
					//d.log.Debugf("aggregation bits for slot %d contained in slot %d (epoch %d):\n%s", att.Data.Slot, j, epoch, utils.RenderByteArrayAsQuadrants(aggregateBits, 8, true))
					if len(v) == 0 {
						d.log.Tracef("skipping attestation in slot %d for epoch %d: no validators found", j, epoch)
						continue
					}
					inclusion_delay := int64(j - att.Data.Slot)
					if inclusion_delay < 1 {
						return fmt.Errorf("inclusion delay is less than 1 in slot %d for attestation of slot %d", j, att.Data.Slot)
					}
					// https://eips.ethereum.org/EIPS/eip-7045
					if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch && inclusion_delay > 32 {
						d.log.Tracef("skipping attestation in slot %d during epoch %d: inclusion delay %d - pre altair fork", j, epoch, inclusion_delay)
						continue
					}
					optimalInclusionDelay := int64(0)
					for k := att.Data.Slot + 1; k < j; k++ {
						c, ok := blockValidityMap[k]
						if !ok {
							return fmt.Errorf("no block validity found for slot %d", k)
						}
						optimalInclusionDelay += c
					}
					is_matching_source := true // enforced by the spec
					br, ok := blockRoots[att.Data.Target.Epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch]
					if !ok {
						return fmt.Errorf("no block root found for epoch %d", att.Data.Target.Epoch)
					}
					is_matching_target := bytes.Equal(att.Data.Target.Root, br)
					br, ok = blockRoots[att.Data.Slot]
					if !ok {
						return fmt.Errorf("no block root found for slot %d", att.Data.Slot)
					}
					is_matching_head := is_matching_target && bytes.Equal(att.Data.BeaconBlockRoot, br)
					sourceValue := int64(0)
					targetValue := int64(0)
					headValue := int64(0)
					switch utils.ForkVersionAtEpoch(epoch).CurrentVersion {
					case utils.Config.Chain.ClConfig.GenesisForkVersion:
						// genesis, head source and target only care about having accurate hashes
						if is_matching_source {
							sourceValue = 1
						}
						if is_matching_target {
							targetValue = 1
						}
						if is_matching_head {
							headValue = 1
						}
					default:
						if is_matching_source && inclusion_delay <= int64(squareSlotsPerEpoch) {
							sourceValue = 1
						}
						if is_matching_target { // 32 slot filter on target for pre altair forks is done above
							targetValue = 1
						}
						if is_matching_head && inclusion_delay == 1 {
							headValue = 1
						}
					}
					// log validator list that attested
					d.log.Tracef("attestation in slot %d during epoch %d: inclusion delay %d (optimal: %d), source %d (m: %t), target %d (m: %t), head %d (m: %t) - validators: %v",
						j, epoch, inclusion_delay, optimalInclusionDelay, sourceValue, is_matching_source, targetValue, is_matching_target, headValue, is_matching_head, v)
					for _, valiIndex := range v {
						valiIndex += tO
						// check if we already processed an attestations for this validator during this epoch
						if (*tar)[tI].AttestationsObserved[valiIndex] > 0 {
							//d.log.Tracef("skipping attestation in slot %d for validator %d during epoch %d - already processed",
							//	j, valiIndex, epoch)
							continue
						}
						// executed
						(*tar)[tI].AttestationsObserved[valiIndex]++
						// inclusion delay
						(*tar)[tI].InclusionDelaySum[valiIndex] += inclusion_delay - 1
						(*tar)[tI].OptimalInclusionDelaySum[valiIndex] += optimalInclusionDelay
						// metrics
						(*tar)[tI].AttestationsSourceExecuted[valiIndex] += sourceValue
						if is_matching_source {
							(*tar)[tI].AttestationsSourceMatched[valiIndex]++
						}
						(*tar)[tI].AttestationsTargetExecuted[valiIndex] += targetValue
						if is_matching_target {
							(*tar)[tI].AttestationsTargetMatched[valiIndex]++
						}
						(*tar)[tI].AttestationsHeadExecuted[valiIndex] += headValue
						if is_matching_head {
							(*tar)[tI].AttestationsHeadMatched[valiIndex]++
						}
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}

// sync odds
func (d *dashboardData) processExpectedSyncPeriods(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if utils.FirstEpochOfSyncPeriod(utils.SyncPeriodOfEpoch(epoch)) != epoch {
			// d.log.Infof("skipping epoch %d as it is not the first epoch of a sync period", epoch)
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_expected_sync_periods_single").Observe(time.Since(now).Seconds())
			}()
			iEpoch := int64(epoch)

			totalEffective := int64(0)
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					totalEffective += int64(valData.EffectiveBalance / 1e9)
				}
			}
			fTotalEffective := float64(totalEffective)
			fCommitteeSize := float64(utils.Config.Chain.ClConfig.SyncCommitteeSize)

			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					// See https://github.com/ethereum/annotated-spec/blob/master/altair/beacon-chain.md#get_sync_committee_indices
					// Note that this formula is not 100% the chance as defined in the spec, but after running simulations we found
					// it being precise enough for our purposes with an error margin of less than 0.003%
					syncChance := float64(valData.EffectiveBalance/1e9) / fTotalEffective
					(*tar)[tI].SyncCommitteesExpected[tO+valData.Index] = syncChance * fCommitteeSize
				}
			}

			return nil
		})
	}
	return g.Wait()
}

// blocks expected
func (d *dashboardData) processBlocksExpected(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		g.Go(func() error {
			defe := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_blocks_expected_single").Observe(time.Since(defe).Seconds())
			}()
			iEpoch := int64(epoch)
			totalEffective := int64(0)
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					totalEffective += int64(valData.EffectiveBalance / 1e9)
				}
			}
			fTotalEffective := float64(totalEffective)
			fSlotsPerEpoch := float64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
			if epoch == 0 {
				// cant propose slot 0
				fSlotsPerEpoch--
			}
			for _, valData := range data.epochBasedData.validatorStates[iEpoch].Data {
				if valData.Status.IsActive() {
					// See https://github.com/ethereum/annotated-spec/blob/master/phase0/beacon-chain.md#compute_proposer_index
					proposalChance := float64(valData.EffectiveBalance/1e9) / fTotalEffective
					(*tar)[tI].BlocksExpected[tO+valData.Index] = proposalChance * fSlotsPerEpoch
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processSyncVotes(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			//d.log.Infof("skipping sync votes for epoch %d (before altair)", epoch)
			// no sync votes before altair
			continue
		}
		syncPeriod := utils.SyncPeriodOfEpoch(epoch)
		iSyncPeriod := int64(syncPeriod)
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_sync_votes_single").Observe(time.Since(now).Seconds())
			}()
			// safety check, check if we have the sync period data
			if _, ok := data.syncPeriodBasedData.SyncAssignments[syncPeriod]; !ok {
				return fmt.Errorf("no sync assignments for sync period %d", syncPeriod)
			}
			for i, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
				(*tar)[tI].SyncCommitteeAssignmentsPeriod[tO+valiIndex] = append((*tar)[tI].SyncCommitteeAssignmentsPeriod[tO+valiIndex], iSyncPeriod)
				(*tar)[tI].SyncCommitteeAssignmentsIndex[tO+valiIndex] = append((*tar)[tI].SyncCommitteeAssignmentsIndex[tO+valiIndex], int64(i))
			}
			for j := startSlot; j < endSlot; j++ {
				for _, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
					(*tar)[tI].SyncStatusSlot[tO+valiIndex] = append((*tar)[tI].SyncStatusSlot[tO+valiIndex], int64(j))
					(*tar)[tI].SyncStatusExecuted[tO+valiIndex] = append((*tar)[tI].SyncStatusExecuted[tO+valiIndex], false)
				}
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// nothing to do
					continue
				}
				// you cant vote for slot 0
				if j == 0 {
					continue
				}
				// sync votes
				v := filterArrayUsingBitMask(
					data.syncPeriodBasedData.SyncAssignments[syncPeriod],
					data.slotBasedData.blocks[j].SyncAggregate.SyncCommitteeBits)
				for _, valiIndex := range v {
					(*tar)[tI].SyncStatusExecuted[tO+valiIndex][len((*tar)[tI].SyncStatusExecuted[tO+valiIndex])-1] = true
				}
				for _, valiIndex := range data.syncPeriodBasedData.SyncAssignments[syncPeriod] {
					(*tar)[tI].SyncScheduled[tO+valiIndex]++
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func (d *dashboardData) processProposalRewards(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	if data.epochBasedData.epochs[len(data.epochBasedData.epochs)-1] < utils.Config.Chain.ClConfig.AltairForkEpoch {
		defer d.phase0HotfixMutex.Unlock()
	}
	g := &errgroup.Group{}
	buffer := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_proposal_rewards_single").Observe(time.Since(now).Seconds())
			}()
			for j := startSlot; j < endSlot; j++ {
				// skip slot 0, as nobody proposed it
				if j == 0 {
					continue
				}
				// calculate median reward
				medianStartSlot := uint64(0)
				if j >= buffer {
					medianStartSlot = j - buffer
				}
				medianEndSlot := j + buffer
				// actually, lets do median instead
				medianArray := make([]int64, 0)
				for k := medianStartSlot; k < medianEndSlot; k++ {
					if r, ok := data.slotBasedData.rewards.blockRewards[k]; ok {
						rewards := r.Data
						reward := rewards.Attestations + rewards.AttesterSlashings + rewards.ProposerSlashings + rewards.SyncAggregate
						medianArray = append(medianArray, reward)
					}
				}
				if len(medianArray) == 0 {
					// no rewards in buffer
					// lets fall back to zero missed rewards. this gets triggered on gnosis during slot 11737794 (and more)
					medianArray = append(medianArray, 0)
				}
				// calculate median
				slices.Sort(medianArray)
				median := int64(0)
				if len(medianArray)%2 == 0 {
					// even
					median = (medianArray[len(medianArray)/2-1] + medianArray[len(medianArray)/2]) / 2
				} else {
					// odd
					median = medianArray[len(medianArray)/2]
				}

				if _, ok := data.slotBasedData.rewards.blockRewards[j]; !ok {
					// assign to validator that missed the slot
					proposer := data.slotBasedData.assignments.blockAssignments[j]
					(*tar)[tI].BlocksClMissedMedianReward[tO+proposer] += median
					continue
				}
				rewards := data.slotBasedData.rewards.blockRewards[j].Data
				(*tar)[tI].BlockRewardsSlot[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSlot[tO+rewards.ProposerIndex], int64(j))
				(*tar)[tI].BlockRewardsAttestationsReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsAttestationsReward[tO+rewards.ProposerIndex], rewards.Attestations)
				(*tar)[tI].BlockRewardsSyncAggregateReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSyncAggregateReward[tO+rewards.ProposerIndex], rewards.SyncAggregate)
				(*tar)[tI].BlockRewardsSlasherReward[tO+rewards.ProposerIndex] = append((*tar)[tI].BlockRewardsSlasherReward[tO+rewards.ProposerIndex], rewards.AttesterSlashings+rewards.ProposerSlashings)
				reward := rewards.Attestations + rewards.AttesterSlashings + rewards.ProposerSlashings + rewards.SyncAggregate
				if reward < median {
					(*tar)[tI].BlocksClMissedMedianReward[tO+rewards.ProposerIndex] += median - reward
				}
				if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
					// hotfix for phase0 blocks
					if (*data).epochBasedData.rewards.attestationRewards[epoch][rewards.ProposerIndex].InclusionDelay > int32(reward) {
						(*data).epochBasedData.rewards.attestationRewards[epoch][rewards.ProposerIndex].InclusionDelay -= int32(reward)
					}
				}
			}
			return nil
		})
	}
	return g.Wait()
}

// sync rewards
func (d *dashboardData) processSyncCommitteeRewards(data *MultiEpochData, tar *[]types.VDBDataEpochColumns) error {
	g := &errgroup.Group{}
	for i, e := range data.epochBasedData.epochs {
		epoch := e
		if epoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
			//d.log.Infof("skipping sync rewards for epoch %d (before altair)", epoch)
			// no sync rewards before altair
			continue
		}
		tI := data.epochBasedData.tarIndices[i]
		tO := uint64(data.epochBasedData.tarOffsets[i])
		startSlot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		endSlot := startSlot + utils.Config.Chain.ClConfig.SlotsPerEpoch
		g.Go(func() error {
			now := time.Now()
			defer func() {
				metrics.TaskDuration.WithLabelValues("dashboard_data_exporter_process_sync_rewards_single").Observe(time.Since(now).Seconds())
			}()
			maxRewards := int64(0)
			for j := startSlot; j < endSlot; j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// no data is expected for missed blocks
					continue
				}
				for _, reward := range data.slotBasedData.rewards.syncCommitteeRewards[j].Data {
					(*tar)[tI].SyncRewardsSlot[tO+reward.ValidatorIndex] = append((*tar)[tI].SyncRewardsSlot[tO+reward.ValidatorIndex], int64(j))
					(*tar)[tI].SyncRewardsReward[tO+reward.ValidatorIndex] = append((*tar)[tI].SyncRewardsReward[tO+reward.ValidatorIndex], reward.Reward)
					if reward.Reward > maxRewards {
						maxRewards = reward.Reward
					}
				}
			}
			/*
				// removed safety check as it was triggered in a real scenario on gnosis
					if maxRewards <= 0 {
						// lets just bork ourselves to be safe
						return fmt.Errorf("no rewards in epoch %d", epoch)
					}
			*/
			for j := startSlot; j < endSlot; j++ {
				if _, ok := data.slotBasedData.blocks[j]; !ok {
					// since sync committee votes can only be included in the block they were meant for, we should not penalize the voter if the block was missed
					continue
				}
				for _, reward := range data.slotBasedData.rewards.syncCommitteeRewards[j].Data {
					(*tar)[tI].SyncLocalizedMaxReward[tO+reward.ValidatorIndex] += maxRewards
				}
			}
			return nil
		})
	}
	return g.Wait()
}
