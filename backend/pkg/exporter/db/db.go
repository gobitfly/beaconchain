package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func SaveBlock(block *types.Block, forceSlotUpdate bool, tx *sqlx.Tx) error {
	blocksMap := make(map[uint64]map[string]*types.Block)
	if blocksMap[block.Slot] == nil {
		blocksMap[block.Slot] = make(map[string]*types.Block)
	}
	blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

	err := saveBlocks(blocksMap, tx, forceSlotUpdate)
	if err != nil {
		log.Fatal(err, "error saving blocks to db", 0)
		return fmt.Errorf("error saving blocks to db: %w", err)
	}

	return nil
}

func saveBlocks(blocks map[uint64]map[string]*types.Block, tx *sqlx.Tx, forceSlotUpdate bool) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_blocks").Observe(time.Since(start).Seconds())
	}()

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return err
	}

	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, graffiti_text, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, syncaggregate_bits, syncaggregate_signature, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, withdrawalcount, voluntaryexitscount, syncaggregate_participation, proposer, status, exec_parent_hash, exec_fee_recipient, exec_state_root, exec_receipts_root, exec_logs_bloom, exec_random, exec_block_number, exec_gas_limit, exec_gas_used, exec_timestamp, exec_extra_data, exec_base_fee_per_gas, exec_block_hash, exec_transactions_count, exec_blob_gas_used, exec_excess_blob_gas, exec_blob_transactions_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

	stmtWithdrawals, err := tx.Prepare(`
		INSERT INTO blocks_withdrawals (block_slot, block_root, withdrawalindex, validatorindex, address, amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (block_slot, block_root, withdrawalindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtWithdrawals.Close()

	stmtBLSChange, err := tx.Prepare(`
		INSERT INTO blocks_bls_change (block_slot, block_root, validatorindex, signature, pubkey, address)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (block_slot, block_root, validatorindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBLSChange.Close()

	stmtProposerSlashing, err := tx.Prepare(`
		INSERT INTO blocks_proposerslashings (block_slot, block_index, block_root, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtProposerSlashing.Close()

	stmtAttesterSlashing, err := tx.Prepare(`
		INSERT INTO blocks_attesterslashings (block_slot, block_index, block_root, attestation1_indices, attestation1_signature, attestation1_slot, attestation1_index, attestation1_beaconblockroot, attestation1_source_epoch, attestation1_source_root, attestation1_target_epoch, attestation1_target_root, attestation2_indices, attestation2_signature, attestation2_slot, attestation2_index, attestation2_beaconblockroot, attestation2_source_epoch, attestation2_source_root, attestation2_target_epoch, attestation2_target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (block_slot, block_index) DO UPDATE SET attestation1_indices = excluded.attestation1_indices, attestation2_indices = excluded.attestation2_indices`)
	if err != nil {
		return err
	}
	defer stmtAttesterSlashing.Close()

	stmtAttestations, err := tx.Prepare(`
		INSERT INTO blocks_attestations (block_slot, block_index, block_root, aggregationbits, validators, signature, slot, committeeindex, beaconblockroot, source_epoch, source_root, target_epoch, target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestations.Close()

	stmtDeposits, err := tx.Prepare(`
		INSERT INTO blocks_deposits (block_slot, block_index, block_root, proof, publickey, withdrawalcredentials, amount, signature, valid_signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtDeposits.Close()

	stmtBlobs, err := tx.Prepare(`
		INSERT INTO blocks_blob_sidecars (block_slot, block_root, index, kzg_commitment, kzg_proof, blob_versioned_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (block_root, index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlobs.Close()

	stmtVoluntaryExits, err := tx.Prepare(`
		INSERT INTO blocks_voluntaryexits (block_slot, block_index, block_root, epoch, validatorindex, signature)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtVoluntaryExits.Close()

	stmtProposalAssignments, err := tx.Prepare(`
		INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex, proposerslot) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtProposalAssignments.Close()

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			if !forceSlotUpdate {
				var dbBlockRootHash []byte
				err := tx.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)
				if err == nil && bytes.Equal(dbBlockRootHash, b.BlockRoot) {
					log.InfoWithFields(log.Fields{"slot": b.Slot, "blockRoot": fmt.Sprintf("%x", b.BlockRoot)}, "skipping export of block as it is already present in the db")
					continue
				} else if err != nil && err != sql.ErrNoRows {
					return fmt.Errorf("error checking for block in db: %w", err)
				}
			}

			res, err := tx.Exec("DELETE FROM blocks WHERE slot = $1 AND length(blockroot) = 1", b.Slot) // Delete placeholder block
			if err != nil {
				return fmt.Errorf("error deleting placeholder block: %w", err)
			}
			ra, err := res.RowsAffected()
			if err != nil && ra > 0 {
				log.InfoWithFields(log.Fields{"slot": b.Slot, "blockRoot": fmt.Sprintf("%x", b.BlockRoot)}, "deleted placeholder block")
			}

			// Set proposer to MAX_SQL_INTEGER if it is the genesis-block (since we are using integers for validator-indices right now)
			if b.Slot == 0 {
				b.Proposer = db.MaxSqlInteger
			}
			syncAggBits := []byte{}
			syncAggSig := []byte{}
			syncAggParticipation := 0.0
			if b.SyncAggregate != nil {
				syncAggBits = b.SyncAggregate.SyncCommitteeBits
				syncAggSig = b.SyncAggregate.SyncCommitteeSignature
				syncAggParticipation = b.SyncAggregate.SyncAggregateParticipation
				// blockLog = blockLog.WithField("syncParticipation", b.SyncAggregate.SyncAggregateParticipation)
			}

			parentHash := []byte{}
			feeRecipient := []byte{}
			stateRoot := []byte{}
			receiptRoot := []byte{}
			logsBloom := []byte{}
			random := []byte{}
			blockNumber := uint64(0)
			gasLimit := uint64(0)
			gasUsed := uint64(0)
			timestamp := uint64(0)
			extraData := []byte{}
			baseFeePerGas := uint64(0)
			blockHash := []byte{}
			txCount := 0
			withdrawalCount := 0
			blobGasUsed := uint64(0)
			excessBlobGas := uint64(0)
			blobTxCount := 0
			if b.ExecutionPayload != nil {
				parentHash = b.ExecutionPayload.ParentHash
				feeRecipient = b.ExecutionPayload.FeeRecipient
				stateRoot = b.ExecutionPayload.StateRoot
				receiptRoot = b.ExecutionPayload.ReceiptsRoot
				logsBloom = b.ExecutionPayload.LogsBloom
				random = b.ExecutionPayload.Random
				blockNumber = b.ExecutionPayload.BlockNumber
				gasLimit = b.ExecutionPayload.GasLimit
				gasUsed = b.ExecutionPayload.GasUsed
				timestamp = b.ExecutionPayload.Timestamp
				extraData = b.ExecutionPayload.ExtraData
				baseFeePerGas = b.ExecutionPayload.BaseFeePerGas
				blockHash = b.ExecutionPayload.BlockHash
				txCount = len(b.ExecutionPayload.Transactions)
				withdrawalCount = len(b.ExecutionPayload.Withdrawals)
				blobGasUsed = b.ExecutionPayload.BlobGasUsed
				excessBlobGas = b.ExecutionPayload.ExcessBlobGas
				blobTxCount = len(b.BlobKZGCommitments)
			}
			_, err = stmtBlock.Exec(
				b.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch,
				b.Slot,
				b.BlockRoot,
				b.ParentRoot,
				b.StateRoot,
				b.Signature,
				b.RandaoReveal,
				b.Graffiti,
				utils.GraffitiToString(b.Graffiti),
				b.Eth1Data.DepositRoot,
				b.Eth1Data.DepositCount,
				b.Eth1Data.BlockHash,
				syncAggBits,
				syncAggSig,
				len(b.ProposerSlashings),
				len(b.AttesterSlashings),
				len(b.Attestations),
				len(b.Deposits),
				withdrawalCount,
				len(b.VoluntaryExits),
				syncAggParticipation,
				b.Proposer,
				strconv.FormatUint(b.Status, 10),
				parentHash,
				feeRecipient,
				stateRoot,
				receiptRoot,
				logsBloom,
				random,
				blockNumber,
				gasLimit,
				gasUsed,
				timestamp,
				extraData,
				baseFeePerGas,
				blockHash,
				txCount,
				blobGasUsed,
				excessBlobGas,
				blobTxCount,
			)
			if err != nil {
				return fmt.Errorf("error executing stmtBlocks for block %v: %w", b.Slot, err)
			}

			for i, c := range b.BlobKZGCommitments {
				_, err := stmtBlobs.Exec(b.Slot, b.BlockRoot, i, c, b.BlobKZGProofs[i], utils.VersionedBlobHash(c).Bytes())
				if err != nil {
					return fmt.Errorf("error executing stmtBlobs for block at slot %v index %v: %w", b.Slot, i, err)
				}
			}
			if payload := b.ExecutionPayload; payload != nil {
				for i, w := range payload.Withdrawals {
					_, err := stmtWithdrawals.Exec(b.Slot, b.BlockRoot, w.Index, w.ValidatorIndex, w.Address, w.Amount)
					if err != nil {
						return fmt.Errorf("error executing stmtWithdrawals for block at slot %v index %v: %w", b.Slot, i, err)
					}
				}
			}
			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, b.BlockRoot, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block at slot %v index %v: %w", b.Slot, i, err)
				}
			}
			for i, bls := range b.SignedBLSToExecutionChange {
				_, err := stmtBLSChange.Exec(b.Slot, b.BlockRoot, bls.Message.Validatorindex, bls.Signature, bls.Message.BlsPubkey, bls.Message.Address)
				if err != nil {
					return fmt.Errorf("error executing stmtBLSChange for block %v index %v: %w", b.Slot, i, err)
				}
			}

			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, b.BlockRoot, pq.Array(as.Attestation1.AttestingIndices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.AttestingIndices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v index %v: %w", b.Slot, i, err)
				}
			}
			for i, a := range b.Attestations {
				_, err = stmtAttestations.Exec(b.Slot, i, b.BlockRoot, a.AggregationBits, pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v index %v: %w", b.Slot, i, err)
				}
			}

			for i, d := range b.Deposits {
				err := utils.VerifyDepositSignature(&phase0.DepositData{
					PublicKey:             phase0.BLSPubKey(d.PublicKey),
					WithdrawalCredentials: d.WithdrawalCredentials,
					Amount:                phase0.Gwei(d.Amount),
					Signature:             phase0.BLSSignature(d.Signature),
				}, domain)

				signatureValid := err == nil

				_, err = stmtDeposits.Exec(b.Slot, i, b.BlockRoot, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, signatureValid)
				if err != nil {
					return fmt.Errorf("error executing stmtDeposits for block %v index %v: %w", b.Slot, i, err)
				}
			}

			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, b.BlockRoot, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v index %v: %w", b.Slot, i, err)
				}
			}

			_, err = stmtProposalAssignments.Exec(b.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %w", b.Slot, err)
			}

			// save the graffitiwall data of the block the db
			err = saveGraffitiwall(b, tx)
			if err != nil {
				return fmt.Errorf("error saving graffitiwall data to the db: %v", err)
			}
		}
	}

	return nil
}

func saveGraffitiwall(block *types.Block, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_graffitiwall").Observe(time.Since(start).Seconds())
	}()

	stmtGraffitiwall, err := tx.Prepare(`
		INSERT INTO graffitiwall (
            x,
            y,
            color,
            slot,
            validator
        )
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (slot) DO UPDATE SET
            x = EXCLUDED.x,
            y = EXCLUDED.y,
            color = EXCLUDED.color,
            validator = EXCLUDED.validator;
		`)
	if err != nil {
		return err
	}
	defer stmtGraffitiwall.Close()

	regexes := [...]*regexp.Regexp{
		regexp.MustCompile("graffitiwall:([0-9]{1,3}):([0-9]{1,3}):#([0-9a-fA-F]{6})"),
		regexp.MustCompile("gw:([0-9]{3})([0-9]{3})([0-9a-fA-F]{6})"),
	}

	var matches []string
	for _, regex := range regexes {
		matches = regex.FindStringSubmatch(string(block.Graffiti))
		if len(matches) > 0 {
			break
		}
	}
	if len(matches) == 4 {
		x, err := strconv.Atoi(matches[1])
		if err != nil || x >= 1000 {
			return fmt.Errorf("error parsing x coordinate for graffiti %v of block %v", string(block.Graffiti), block.Slot)
		}

		y, err := strconv.Atoi(matches[2])
		if err != nil || y >= 1000 {
			return fmt.Errorf("error parsing y coordinate for graffiti %v of block %v", string(block.Graffiti), block.Slot)
		}
		color := matches[3]

		log.Infof("set graffiti at %v - %v with color %v for slot %v by validator %v", x, y, color, block.Slot, block.Proposer)
		_, err = stmtGraffitiwall.Exec(x, y, color, block.Slot, block.Proposer)

		if err != nil {
			return fmt.Errorf("error executing graffitiwall statement: %w", err)
		}
	}
	return nil
}

func SaveValidators(epoch uint64, validators []*types.Validator, client rpc.Client, activationBalanceBatchSize int, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validators").Observe(time.Since(start).Seconds())
	}()

	if activationBalanceBatchSize <= 0 {
		activationBalanceBatchSize = 10000
	}

	var genesisBalances map[uint64][]*types.ValidatorBalance

	if epoch == 0 {
		var err error

		indices := make([]uint64, 0, len(validators))

		for _, validator := range validators {
			indices = append(indices, validator.Index)
		}
		genesisBalances, err = db.BigtableClient.GetValidatorBalanceHistory(indices, 0, 0)
		if err != nil {
			return fmt.Errorf("error retrieving genesis validator balances: %w", err)
		}
	}

	validatorsByIndex := make(map[uint64]*types.Validator, len(validators))
	for _, v := range validators {
		validatorsByIndex[v.Index] = v
	}

	var currentState []*types.Validator
	err := tx.Select(&currentState, "SELECT validatorindex, withdrawableepoch, withdrawalcredentials, slashed, activationeligibilityepoch, activationepoch, exitepoch, status FROM validators;")

	if err != nil {
		return fmt.Errorf("error retrieving current validator state set: %v", err)
	}

	for ; ; time.Sleep(time.Second) { // wait till the last attestation in memory cache has been populated by the exporter
		db.BigtableClient.LastAttestationCacheMux.Lock()
		if db.BigtableClient.LastAttestationCache != nil {
			db.BigtableClient.LastAttestationCacheMux.Unlock()
			break
		}
		db.BigtableClient.LastAttestationCacheMux.Unlock()
		log.Infof("waiting until LastAttestation in memory cache is available")
	}

	currentStateMap := make(map[uint64]*types.Validator, len(currentState))
	latestBlock := uint64(0)
	db.BigtableClient.LastAttestationCacheMux.Lock()
	for _, v := range currentState {
		if db.BigtableClient.LastAttestationCache[v.Index] > latestBlock {
			latestBlock = db.BigtableClient.LastAttestationCache[v.Index]
		}
		currentStateMap[v.Index] = v
	}
	db.BigtableClient.LastAttestationCacheMux.Unlock()

	thresholdSlot := uint64(0)
	if latestBlock >= 64 {
		thresholdSlot = latestBlock - 64
	}

	latestEpoch := latestBlock / utils.Config.Chain.ClConfig.SlotsPerEpoch

	var queries strings.Builder

	insertStmt, err := tx.Prepare(`INSERT INTO validators (
		validatorindex,
		pubkey,
		withdrawableepoch,
		withdrawalcredentials,
		balance,
		effectivebalance,
		slashed,
		activationeligibilityepoch,
		activationepoch,
		exitepoch,
		pubkeyhex,
		status
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`)
	if err != nil {
		return fmt.Errorf("error preparing insert validator statement: %w", err)
	}

	updates := 0
	for _, v := range validators {
		// exchange farFutureEpoch with the corresponding max sql value
		if v.WithdrawableEpoch == db.FarFutureEpoch {
			v.WithdrawableEpoch = db.MaxSqlNumber
		}
		if v.ExitEpoch == db.FarFutureEpoch {
			v.ExitEpoch = db.MaxSqlNumber
		}
		if v.ActivationEligibilityEpoch == db.FarFutureEpoch {
			v.ActivationEligibilityEpoch = db.MaxSqlNumber
		}
		if v.ActivationEpoch == db.FarFutureEpoch {
			v.ActivationEpoch = db.MaxSqlNumber
		}

		c := currentStateMap[v.Index]

		if c == nil {
			if v.Index%1000 == 0 {
				log.Infof("validator %v is new", v.Index)
			}

			_, err = insertStmt.Exec(
				v.Index,
				v.PublicKey,
				v.WithdrawableEpoch,
				v.WithdrawalCredentials,
				0,
				0,
				v.Slashed,
				v.ActivationEligibilityEpoch,
				v.ActivationEpoch,
				v.ExitEpoch,
				fmt.Sprintf("%x", v.PublicKey),
				v.Status,
			)

			if err != nil {
				log.Error(err, "error saving new validator", 0, map[string]interface{}{"index": v.Index})
			}
		} else {
			// status                     =
			// CASE
			// WHEN EXCLUDED.exitepoch <= %[1]d AND EXCLUDED.slashed THEN 'slashed'
			// WHEN EXCLUDED.exitepoch <= %[1]d THEN 'exited'
			// WHEN EXCLUDED.activationeligibilityepoch = 9223372036854775807 THEN 'deposited'
			// WHEN EXCLUDED.activationepoch > %[1]d THEN 'pending'
			// WHEN EXCLUDED.slashed AND EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'slashing_offline'
			// WHEN EXCLUDED.slashed THEN 'slashing_online'
			// WHEN EXCLUDED.exitepoch < 9223372036854775807 AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'exiting_offline'
			// WHEN EXCLUDED.exitepoch < 9223372036854775807 THEN 'exiting_online'
			// WHEN EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'active_offline'
			// ELSE 'active_online'
			// END
			db.BigtableClient.LastAttestationCacheMux.Lock()
			offline := db.BigtableClient.LastAttestationCache[v.Index] < thresholdSlot
			db.BigtableClient.LastAttestationCacheMux.Unlock()

			if v.ExitEpoch <= latestEpoch && v.Slashed {
				v.Status = "slashed"
			} else if v.ExitEpoch <= latestEpoch {
				v.Status = "exited"
			} else if v.ActivationEligibilityEpoch == 9223372036854775807 {
				v.Status = "deposited"
			} else if v.ActivationEpoch > latestEpoch {
				v.Status = "pending"
			} else if v.Slashed && v.ActivationEpoch < latestEpoch && offline {
				v.Status = "slashing_offline"
			} else if v.Slashed {
				v.Status = "slashing_online"
			} else if v.ExitEpoch < 9223372036854775807 && offline {
				v.Status = "exiting_offline"
			} else if v.ExitEpoch < 9223372036854775807 {
				v.Status = "exiting_online"
			} else if v.ActivationEpoch < latestEpoch && offline {
				v.Status = "active_offline"
			} else {
				v.Status = "active_online"
			}

			if c.Status != v.Status {
				log.Tracef("Status changed for validator %v from %v to %v", v.Index, c.Status, v.Status)
				// logger.Tracef("v.ActivationEpoch %v, latestEpoch %v, lastAttestationSlots[v.Index] %v, thresholdSlot %v", v.ActivationEpoch, latestEpoch, lastAttestationSlots[v.Index], thresholdSlot)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET status = '%s' WHERE validatorindex = %d;\n", v.Status, c.Index))
				updates++
			}
			// if c.Balance != v.Balance {
			// 	// log.LogInfo("Balance changed for validator %v from %v to %v", v.Index, c.Balance, v.Balance)
			// 	queries.WriteString(fmt.Sprintf("UPDATE validators SET balance = %d WHERE validatorindex = %d;\n", v.Balance, c.Index))
			// 	updates++
			// }
			// if c.EffectiveBalance != v.EffectiveBalance {
			// 	// log.LogInfo("EffectiveBalance changed for validator %v from %v to %v", v.Index, c.EffectiveBalance, v.EffectiveBalance)
			// 	queries.WriteString(fmt.Sprintf("UPDATE validators SET effectivebalance = %d WHERE validatorindex = %d;\n", v.EffectiveBalance, c.Index))
			// 	updates++
			// }
			if c.Slashed != v.Slashed {
				log.Infof("Slashed changed for validator %v from %v to %v", v.Index, c.Slashed, v.Slashed)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET slashed = %v WHERE validatorindex = %d;\n", v.Slashed, c.Index))
				updates++
			}
			if c.ActivationEligibilityEpoch != v.ActivationEligibilityEpoch {
				log.Infof("ActivationEligibilityEpoch changed for validator %v from %v to %v", v.Index, c.ActivationEligibilityEpoch, v.ActivationEligibilityEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET activationeligibilityepoch = %d WHERE validatorindex = %d;\n", v.ActivationEligibilityEpoch, c.Index))
				updates++
			}
			if c.ActivationEpoch != v.ActivationEpoch {
				log.Infof("ActivationEpoch changed for validator %v from %v to %v", v.Index, c.ActivationEpoch, v.ActivationEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET activationepoch = %d WHERE validatorindex = %d;\n", v.ActivationEpoch, c.Index))
				updates++
			}
			if c.ExitEpoch != v.ExitEpoch {
				log.Infof("ExitEpoch changed for validator %v from %v to %v", v.Index, c.ExitEpoch, v.ExitEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET exitepoch = %d WHERE validatorindex = %d;\n", v.ExitEpoch, c.Index))
				updates++
			}
			if c.WithdrawableEpoch != v.WithdrawableEpoch {
				log.Infof("WithdrawableEpoch changed for validator %v from %v to %v", v.Index, c.WithdrawableEpoch, v.WithdrawableEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET withdrawableepoch = %d WHERE validatorindex = %d;\n", v.WithdrawableEpoch, c.Index))
				updates++
			}
			if !bytes.Equal(c.WithdrawalCredentials, v.WithdrawalCredentials) {
				log.Infof("WithdrawalCredentials changed for validator %v from %x to %x", v.Index, c.WithdrawalCredentials, v.WithdrawalCredentials)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET withdrawalcredentials = '\\x%x' WHERE validatorindex = %d;\n", v.WithdrawalCredentials, c.Index))
				updates++
			}
		}
	}

	err = insertStmt.Close()
	if err != nil {
		return fmt.Errorf("error closing insert validator statement: %w", err)
	}

	if updates > 0 {
		updateStart := time.Now()
		log.Infof("applying %v validator table update queries", updates)
		_, err = tx.Exec(queries.String())
		if err != nil {
			log.Error(err, "error executing validator update query", 0)
			return err
		}
		log.Infof("validator table update completed, took %v", time.Since(updateStart))
	}

	s := time.Now()
	newValidators := []struct {
		Validatorindex  uint64
		ActivationEpoch uint64
	}{}

	err = tx.Select(&newValidators, "SELECT validatorindex, activationepoch FROM validators WHERE balanceactivation IS NULL ORDER BY activationepoch LIMIT $1", activationBalanceBatchSize)
	if err != nil {
		return fmt.Errorf("error retreiving activation epoch balances from db: %w", err)
	}

	balanceCache := make(map[uint64]map[uint64]uint64)
	currentActivationEpoch := uint64(0)

	// get genesis balances of all validators for performance

	for _, newValidator := range newValidators {
		if newValidator.ActivationEpoch > epoch {
			continue
		}

		if newValidator.ActivationEpoch != currentActivationEpoch {
			log.Infof("removing epoch %v from the activation epoch balance cache", currentActivationEpoch)
			delete(balanceCache, currentActivationEpoch) // remove old items from the map
			currentActivationEpoch = newValidator.ActivationEpoch
		}

		var balance map[uint64][]*types.ValidatorBalance
		if newValidator.ActivationEpoch == 0 {
			balance = genesisBalances
		} else {
			balance, err = db.BigtableClient.GetValidatorBalanceHistory([]uint64{newValidator.Validatorindex}, newValidator.ActivationEpoch, newValidator.ActivationEpoch)
			if err != nil {
				return fmt.Errorf("error retreiving validator balance history: %w", err)
			}
		}

		foundBalance := uint64(0)
		if balance[newValidator.Validatorindex] == nil || len(balance[newValidator.Validatorindex]) == 0 {
			log.Warnf("no activation epoch balance found for validator %v for epoch %v in bigtable, trying node", newValidator.Validatorindex, newValidator.ActivationEpoch)

			if balanceCache[newValidator.ActivationEpoch] == nil {
				balances, err := client.GetBalancesForEpoch(int64(newValidator.ActivationEpoch))
				if err != nil {
					return fmt.Errorf("error retrieving balances for epoch %d: %v", newValidator.ActivationEpoch, err)
				}
				balanceCache[newValidator.ActivationEpoch] = balances
			}
			foundBalance = balanceCache[newValidator.ActivationEpoch][newValidator.Validatorindex]
		} else {
			foundBalance = balance[newValidator.Validatorindex][0].Balance
		}

		log.Infof("retrieved activation epoch balance of %v for validator %v", foundBalance, newValidator.Validatorindex)

		_, err = tx.Exec("update validators set balanceactivation = $1 WHERE validatorindex = $2 AND balanceactivation IS NULL;", foundBalance, newValidator.Validatorindex)
		if err != nil {
			return fmt.Errorf("error updating activation epoch balance for validator %v: %w", newValidator.Validatorindex, err)
		}
	}

	log.Infof("updating validator activation epoch balance completed, took %v", time.Since(s))

	s = time.Now()
	_, err = tx.Exec("ANALYZE (SKIP_LOCKED) validators;")
	if err != nil {
		return fmt.Errorf("analyzing validators table: %w", err)
	}
	log.Infof("analyze of validators table completed, took %v", time.Since(s))

	return nil
}

// SaveEpoch will save the epoch data into the database
func SaveEpoch(epoch uint64, validators []*types.Validator, client rpc.Client, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_epoch").Observe(time.Since(start).Seconds())
		log.InfoWithFields(log.Fields{"epoch": epoch, "duration": time.Since(start)}, "completed saving epoch")
	}()

	log.InfoWithFields(log.Fields{"chainEpoch": utils.TimeToEpoch(time.Now()), "exportEpoch": epoch}, "starting export of epoch")

	log.Infof("exporting epoch statistics data")
	proposerSlashingsCount := 0
	attesterSlashingsCount := 0
	attestationsCount := 0
	depositCount := 0
	voluntaryExitCount := 0
	withdrawalCount := 0

	// for _, slot := range data.Blocks {
	// 	for _, b := range slot {
	// 		proposerSlashingsCount += len(b.ProposerSlashings)
	// 		attesterSlashingsCount += len(b.AttesterSlashings)
	// 		attestationsCount += len(b.Attestations)
	// 		depositCount += len(b.Deposits)
	// 		voluntaryExitCount += len(b.VoluntaryExits)
	// 		if b.ExecutionPayload != nil {
	// 			withdrawalCount += len(b.ExecutionPayload.Withdrawals)
	// 		}
	// 	}
	// }

	validatorBalanceSum := decimal.NewFromInt(0)
	validatorEffectiveBalanceSum := decimal.NewFromInt(0)
	validatorsCount := 0
	for _, v := range validators {
		if v.ExitEpoch > epoch && v.ActivationEpoch <= epoch {
			validatorsCount++
			validatorBalanceSum = validatorBalanceSum.Add(decimal.NewFromInt(int64(v.Balance)))
			validatorEffectiveBalanceSum = validatorEffectiveBalanceSum.Add(decimal.NewFromInt(int64(v.EffectiveBalance)))
		}
	}

	validatorBalanceAverage := validatorBalanceSum.Div(decimal.NewFromInt(int64(validatorsCount)))

	_, err := tx.Exec(`
		INSERT INTO epochs (
			epoch,
			blockscount,
			proposerslashingscount,
			attesterslashingscount,
			attestationscount,
			depositscount,
			withdrawalcount,
			voluntaryexitscount,
			validatorscount,
			averagevalidatorbalance,
			totalvalidatorbalance,
			eligibleether,
			globalparticipationrate,
			votedether,
			finalized
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (epoch) DO UPDATE SET
			blockscount             = excluded.blockscount,
			proposerslashingscount  = excluded.proposerslashingscount,
			attesterslashingscount  = excluded.attesterslashingscount,
			attestationscount       = excluded.attestationscount,
			depositscount           = excluded.depositscount,
			withdrawalcount         = excluded.withdrawalcount,
			voluntaryexitscount     = excluded.voluntaryexitscount,
			validatorscount         = excluded.validatorscount,
			averagevalidatorbalance = excluded.averagevalidatorbalance,
			totalvalidatorbalance   = excluded.totalvalidatorbalance,
			eligibleether           = excluded.eligibleether,
			globalparticipationrate = excluded.globalparticipationrate,
			votedether              = excluded.votedether,
			finalized               = excluded.finalized`,
		epoch,
		0,
		proposerSlashingsCount,
		attesterSlashingsCount,
		attestationsCount,
		depositCount,
		withdrawalCount,
		voluntaryExitCount,
		validatorsCount,
		validatorBalanceAverage.BigInt().String(),
		validatorBalanceSum.BigInt().String(),
		validatorEffectiveBalanceSum.BigInt().String(),
		0,
		0,
		false)

	if err != nil {
		return fmt.Errorf("error executing save epoch statement: %w", err)
	}

	lookback := uint64(0)
	if epoch > 3 {
		lookback = epoch - 3
	}
	// delete duplicate scheduled slots
	_, err = tx.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x0})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}

	// delete duplicate missed blocks
	_, err = tx.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x1})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}
	return nil
}

func GetLatestDashboardEpoch() (uint64, error) {
	var lastEpoch uint64
	err := db.AlloyWriter.Get(&lastEpoch, fmt.Sprintf("SELECT COALESCE(max(epoch), 0) FROM %s", EpochWriterTableName))
	return lastEpoch, err
}

func GetOldestDashboardEpoch() (uint64, error) {
	var epoch uint64
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT COALESCE(min(epoch), 0) FROM %s", EpochWriterTableName))
	return epoch, err
}

func GetMinOldHourlyEpoch() (uint64, error) {
	var epoch uint64
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT min(epoch_start) as epoch_start FROM %s", HourWriterTableName))
	return epoch, err
}

type EpochBounds struct {
	EpochStart uint64 `db:"epoch_start"`
	EpochEnd   uint64 `db:"epoch_end"`
}

type DayBounds struct {
	Day        time.Time `db:"day"`
	EpochStart uint64    `db:"epoch_start"`
	EpochEnd   uint64    `db:"epoch_end"`
}

func GetLastExportedTotalEpoch() (*EpochBounds, error) {
	var epoch EpochBounds
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT COALESCE(max(epoch_start),0) as epoch_start, COALESCE(max(epoch_end),0) as epoch_end FROM %s", RollingTotalWriterTableName))
	return &epoch, err
}

func GetLastExportedHour() (*EpochBounds, error) {
	var epoch EpochBounds
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT COALESCE(max(epoch_start),0) as epoch_start, COALESCE(max(epoch_end),0) as epoch_end FROM %s", HourWriterTableName))
	return &epoch, err
}

func GetLastExportedDay() (*DayBounds, error) {
	var epoch DayBounds
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT day, epoch_start, epoch_end FROM %s ORDER BY day DESC LIMIT 1", DayWriterTableName))
	return &epoch, err
}

func HasDashboardDataForEpoch(targetEpoch uint64) (bool, error) {
	var epoch uint64
	err := db.AlloyWriter.Get(&epoch, fmt.Sprintf("SELECT epoch FROM %s WHERE epoch = $1 LIMIT 1", EpochWriterTableName), targetEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// returns epochs between start and end that are missing in the database, start is inclusive end is exclusive
func GetMissingEpochsBetween(start, end int64) ([]uint64, error) {
	if start < 0 {
		start = 0
	}
	if end <= start {
		return nil, nil
	}

	if end-start > 100 {
		// for large ranges we use a different approach to avoid making tons of selects
		// this performs better for large ranges but is slow for short ranges
		var epochs []uint64
		err := db.AlloyWriter.Select(&epochs, fmt.Sprintf(`
			WITH
			epoch_range AS (
				SELECT generate_series($1::bigint, $2::bigint) AS epoch
			),
			distinct_present_epochs AS (
				SELECT DISTINCT epoch
				FROM %s
				WHERE epoch >= $1 AND epoch <= $2
			)
			SELECT epoch_range.epoch
			FROM epoch_range
			LEFT JOIN distinct_present_epochs ON epoch_range.epoch = distinct_present_epochs.epoch
			WHERE distinct_present_epochs.epoch IS NULL
			ORDER BY epoch_range.epoch
		`, EpochWriterTableName), start, end-1)
		return epochs, err
	}

	query := `SELECT TO_JSON(ARRAY_AGG(epoch)) AS result_array FROM (`

	for epoch := start; epoch < end; epoch++ {
		if epoch != start {
			query += " UNION "
		}
		query += fmt.Sprintf(`SELECT %[1]d AS epoch WHERE NOT EXISTS (SELECT 1 FROM %[2]s WHERE epoch = %[1]d LIMIT 1)`, epoch, EpochWriterTableName)
	}

	query += `) AS result_array;`

	var jsonArray sql.NullString

	err := db.AlloyReader.Get(&jsonArray, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
	}

	if !jsonArray.Valid {
		return nil, nil
	}

	missingEpochs := make([]uint64, 0)
	err = json.Unmarshal([]byte(jsonArray.String), &missingEpochs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	// sort asc
	sort.Slice(missingEpochs, func(i, j int) bool {
		return missingEpochs[i] < missingEpochs[j]
	})

	return missingEpochs, nil
}

func GetPartitionNamesOfTable(tableName string) ([]string, error) {
	var partitions []string
	err := db.AlloyWriter.Select(&partitions, fmt.Sprintf(`
		SELECT inhrelid::regclass AS partition_name
		FROM pg_inherits
		WHERE inhparent = 'public.%s'::regclass;`, tableName),
	)
	return partitions, err
}

func AddToColumnEngine(table, columns string) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		SELECT google_columnar_engine_add(
			relation => '%s',
			columns => '%s'
		);
		`, table, columns))
	return err
}

func AddToColumnEngineAllColumns(table string) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		SELECT google_columnar_engine_add(
			relation => '%s'
		);
		`, table))
	return err
}

const EpochWriterTableName = "validator_dashboard_data_epoch"
const DayWriterTableName = "validator_dashboard_data_daily"
const HourWriterTableName = "validator_dashboard_data_hourly"

const RollingTotalWriterTableName = "validator_dashboard_data_rolling_total"
const RollingDailyWriterTable = "validator_dashboard_data_rolling_daily"
const RollingWeeklyWriterTable = "validator_dashboard_data_rolling_weekly"
const RollingMonthlyWriterTable = "validator_dashboard_data_rolling_monthly"
const RollingNinetyDaysWriterTable = "validator_dashboard_data_rolling_90d"
