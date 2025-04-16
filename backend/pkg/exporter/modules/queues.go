// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconchain Dashboard.
//
// Beaconchain Dashboard is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain Dashboard is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain Dashboard.  If not, see <https://www.gnu.org/licenses/>.

package modules

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"

	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
)

type QueueClient interface {
	GetPendingDeposits(stateID any) (*constypes.StandardBeaconPendingDepositsResponse, error)
	GetChainHead() (*types.ChainHead, error)
	GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error)
}

type pendingQueueExporter struct {
	lc    QueueClient
	db    db2.ConsensusRepository
	ctx   context.Context
	delay time.Duration
}

func newPendingQueueExporter(ctx context.Context, client rpc.Client, db db2.ConsensusRepository) *pendingQueueExporter {
	indexer := &pendingQueueExporter{
		lc:  client,
		db:  db,
		ctx: ctx,

		// interval MUST be longer than one epoch
		// Background: A freshly exported validator will have an eligible epoch of max uint64, by keeping the pending deposits
		// a bit longer in the db, we can rely on the pending deposits table to still get us an estimate for eligibility
		delay: time.Minute * 10,
	}
	return indexer
}

func (s *pendingQueueExporter) Export() {
	for {
		select {
		case <-s.ctx.Done():
			log.Info("pending deposit queue export loop cancelled")
			return
		default:
			startTime := time.Now()
			statusReport := services.StatusReporter.NewStatusReport(constants.Event_ExporterModulePendingDepositQueue, constants.Default, time.Second*12)
			statusReport(constants.Running, nil)

			err := s.Index()
			if err != nil {
				log.Error(err, "error exporting pending deposit queue", 0)
				statusReport(constants.Failure, map[string]string{"error": err.Error()})
			} else {
				statusReport(constants.Success, map[string]string{
					"took":     time.Since(startTime).String(),
					"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
				})
			}

			time.Sleep(s.delay)
		}
	}
}

func (qi *pendingQueueExporter) Index() error {
	defer log.Infof("pending deposit queue export finished")

	head, err := qi.lc.GetChainHead()
	if err != nil {
		return errors.Wrap(err, "failed to get chain head")
	}
	epoch := head.HeadEpoch

	if !utils.ElectraHasHappened(epoch) {
		log.Infof("pending deposit queue export skipped, electra has not happened yet")
		return nil
	}

	deposits, err := qi.lc.GetPendingDeposits("head")
	if err != nil {
		return errors.Wrap(err, "failed to get pending deposits")
	}

	validators, err := qi.lc.GetValidatorState(epoch)
	if err != nil {
		return errors.Wrap(err, "failed to get validator state")
	}

	type MiniState struct {
		Index             uint64
		ExitEpoch         uint64
		WithdrawableEpoch uint64
	}

	totalActiveEffectiveBalance := uint64(0)
	pubkeyToIndexMap := make(map[string]*MiniState)

	for _, v := range validators.Data {
		pubkeyToIndexMap[v.Validator.Pubkey.String()] = &MiniState{
			Index:             v.Index,
			ExitEpoch:         v.Validator.ExitEpoch,
			WithdrawableEpoch: v.Validator.WithdrawableEpoch,
		}
		if epoch >= v.Validator.ActivationEpoch && epoch < v.Validator.ExitEpoch {
			totalActiveEffectiveBalance += v.Validator.EffectiveBalance
		}
	}

	etherChurnByEpoch := utils.GetActivationExitChurnLimit(totalActiveEffectiveBalance)
	count := 0
	balanceAhead := uint64(0)
	clearEpoch := head.HeadEpoch + 1

	// transition period
	// pre electra system will keep going for follow distance until every deposit of the last system is converted to the new system
	// before the new system starts
	electraQueueDelay := utils.Config.ClConfig.Eth1FollowDistance/utils.Config.ClConfig.SlotsPerEpoch + utils.Config.ClConfig.EpochsPerEth1VotingPeriod
	if clearEpoch < utils.Config.ClConfig.ElectraForkEpoch+electraQueueDelay {
		clearEpoch = utils.Config.ClConfig.ElectraForkEpoch + electraQueueDelay
	}

	depositsList := make([]types.PendingDeposit, 0)

	// spec vars (in snake_case)
	next_deposit_index := uint64(0)
	max_pending_deposits_per_epoch := utils.Config.ClConfig.MaxPendingDepositsPerEpoch
	if max_pending_deposits_per_epoch == 0 { // eth mainnet spec default
		max_pending_deposits_per_epoch = uint64(16)
	}
	processed_amount := uint64(0)
	state_deposit_balance_to_consume := uint64(0)

	pending_deposits := deposits.Data
	depositsToPostpone := []types.PendingDeposit{} // est differently than the spec as we just set these to the same clearEpoch as the "normal" last entry. Not snake case to highlight the different handling to spec

	// emulate spec based on current view in time (approx estimation)
	// https://github.com/ethereum/consensus-specs/blob/dev/specs/electra/beacon-chain.md#new-process_pending_deposits
	for {
		next_epoch := clearEpoch + 1
		available_for_processing := state_deposit_balance_to_consume + etherChurnByEpoch
		processed_amount = 0
		next_deposit_index = 0

		is_churn_limit_reached := false
		finalized_slot := next_epoch * utils.Config.ClConfig.SlotsPerEpoch // first slot of next epoch is finalized
		// potential improvement: utils.GetActivationExitChurnLimit(totalActiveEffectiveBalance + balanceAhead - withdrawalsAhead)

		for _, deposit := range pending_deposits {
			if deposit.Slot > finalized_slot {
				break
			}

			if next_deposit_index >= max_pending_deposits_per_epoch {
				break
			}

			miniState, found := pubkeyToIndexMap[deposit.Pubkey.String()]
			var is_validator_exited bool
			var is_validator_withdrawn bool

			if found {
				is_validator_exited = miniState.ExitEpoch < 100_000_000_000
				is_validator_withdrawn = miniState.WithdrawableEpoch < next_epoch
			}

			getPendingDeposit := func() types.PendingDeposit {
				pendingDeposit := types.PendingDeposit{
					ID:                    count,
					Pubkey:                deposit.Pubkey,
					WithdrawalCredentials: deposit.WithdrawalCredentials,
					Amount:                deposit.Amount,
					Signature:             deposit.Signature,
					Slot:                  deposit.Slot,
					ValidatorIndex:        sql.NullInt64{},
					QueuedBalanceAhead:    balanceAhead,
					EstClearEpoch:         clearEpoch,
				}

				if found {
					pendingDeposit.ValidatorIndex = sql.NullInt64{
						Int64: int64(miniState.Index),
						Valid: true,
					}
				}
				return pendingDeposit
			}

			if is_validator_withdrawn { // do not consume churn
				depositsList = append(depositsList, getPendingDeposit())
			} else if is_validator_exited { // do not consume churn
				depositsToPostpone = append(depositsToPostpone, getPendingDeposit())
			} else {
				is_churn_limit_reached = processed_amount+deposit.Amount > available_for_processing
				if is_churn_limit_reached {
					break
				}
				processed_amount += deposit.Amount
				depositsList = append(depositsList, getPendingDeposit())
			}

			next_deposit_index++

			// out of spec
			balanceAhead += deposit.Amount
			count++
		}

		pending_deposits = pending_deposits[next_deposit_index:]

		if len(pending_deposits) == 0 {
			break
		}

		if is_churn_limit_reached {
			state_deposit_balance_to_consume = available_for_processing - processed_amount
		} else {
			state_deposit_balance_to_consume = 0
		}

		clearEpoch++
	}

	// treat postpones deposits differently, set to last epoch of "normal" deposits
	// since we can't accurately predict them anyway if they are that far out where there are no "normal" deposits with current state
	if len(depositsList) > 0 {
		lastEntry := depositsList[len(depositsList)-1]
		for i := range depositsToPostpone {
			depositsToPostpone[i].EstClearEpoch = lastEntry.EstClearEpoch
			depositsToPostpone[i].QueuedBalanceAhead = lastEntry.QueuedBalanceAhead
		}
		depositsList = append(depositsList, depositsToPostpone...)
	}

	return qi.db.SavePendingDepositsQueue(depositsList)
}
