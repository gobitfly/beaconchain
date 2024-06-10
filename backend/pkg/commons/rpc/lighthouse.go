package rpc

import (
	"bytes"
	"net/http"

	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"golang.org/x/sync/errgroup"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prysmaticlabs/go-bitfield"
)

// TODO replace most lc.cl.Endpoint with lc.cl and use the interface

// LighthouseLatestHeadEpoch is used to cache the latest head epoch for participation requests
var LighthouseLatestHeadEpoch uint64 = 0

// LighthouseClient holds the Lighthouse client info
type LighthouseClient struct {
	cl                  *consapi.NodeClient
	assignmentsCache    *lru.Cache
	assignmentsCacheMux *sync.Mutex
	slotsCache          *lru.Cache
	slotsCacheMux       *sync.Mutex
	signer              gethtypes.Signer
}

// NewLighthouseClient is used to create a new Lighthouse client
func NewLighthouseClient(cl *consapi.NodeClient, chainID *big.Int) (*LighthouseClient, error) {
	signer := gethtypes.NewCancunSigner(chainID)
	client := &LighthouseClient{
		cl:                  cl,
		assignmentsCacheMux: &sync.Mutex{},
		slotsCacheMux:       &sync.Mutex{},
		signer:              signer,
	}
	client.assignmentsCache, _ = lru.New(10)
	client.slotsCache, _ = lru.New(128) // cache at most 128 slots

	return client, nil
}

func (lc *LighthouseClient) GetNewBlockChan() chan *types.Block {
	blkCh := make(chan *types.Block, 10)
	go func() {
		res := lc.cl.GetEvents([]constypes.EventTopic{constypes.EventHead})

		for event := range res {
			if event.Error != nil {
				log.Fatal(event.Error, "Lighthouse connection error (will automatically retry to connect)", 0)
			}

			if event.Event != constypes.EventHead {
				continue
			}

			head, err := event.Head()
			if err != nil {
				log.Error(err, "error parsing head event", 0)
				continue
			}

			block, err := lc.GetBlockBySlot(head.Slot)
			if err != nil {
				log.Warnf("failed to fetch block for slot %d: %v", head.Slot, err)
				continue
			}
			log.Infof("retrieved block for slot %v", head.Slot)
			// logger.Infof("pushing block %v", blk.Slot)
			blkCh <- block
		}
	}()
	return blkCh
}

// GetChainHead gets the chain head from Lighthouse
// Deprecated: Use retriever.GetChainHead() instead
func (lc *LighthouseClient) GetChainHead() (*types.ChainHead, error) {
	parsedHead, err := lc.cl.GetBlockHeader("head")
	if err != nil {
		return &types.ChainHead{}, err
	}

	id := parsedHead.Data.Header.Message.StateRoot.String()
	if parsedHead.Data.Header.Message.Slot == 0 {
		id = "genesis"
	}

	parsedFinality, err := lc.cl.GetFinalityCheckpoints(id)
	if err != nil {
		return &types.ChainHead{}, err
	}

	// The epoch in the Finalized Object is not the finalized epoch, but the epoch for the checkpoint - the 'real' finalized epoch is the one before
	var finalizedEpoch = parsedFinality.Data.Finalized.Epoch
	if finalizedEpoch > 0 {
		finalizedEpoch--
	}

	finalizedSlot := (finalizedEpoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch // The first Slot of the next epoch is finalized.
	if finalizedEpoch == 0 && utils.IsByteArrayAllZero(parsedFinality.Data.Finalized.Root) {
		finalizedSlot = 0
	}
	return &types.ChainHead{
		HeadSlot:                   parsedHead.Data.Header.Message.Slot,
		HeadEpoch:                  parsedHead.Data.Header.Message.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch,
		HeadBlockRoot:              parsedHead.Data.Root,
		FinalizedSlot:              finalizedSlot,
		FinalizedEpoch:             finalizedEpoch,
		FinalizedBlockRoot:         parsedFinality.Data.Finalized.Root,
		JustifiedSlot:              parsedFinality.Data.CurrentJustified.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch,
		JustifiedEpoch:             parsedFinality.Data.CurrentJustified.Epoch,
		JustifiedBlockRoot:         parsedFinality.Data.CurrentJustified.Root,
		PreviousJustifiedSlot:      parsedFinality.Data.PreviousJustified.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch,
		PreviousJustifiedEpoch:     parsedFinality.Data.PreviousJustified.Epoch,
		PreviousJustifiedBlockRoot: parsedFinality.Data.PreviousJustified.Root,
	}, nil
}

func (lc *LighthouseClient) GetValidatorQueue() (*types.ValidatorQueue, error) {
	// pre-filter the status, to return much less validators, thus much faster!
	parsedValidators, err := lc.cl.GetValidators("head", nil, []constypes.ValidatorStatus{constypes.PendingQueued, constypes.ActiveExiting, constypes.ActiveSlashed})
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator for head valiqdator queue check: %w", err)
	}

	// TODO: maybe track more status counts in the future?
	statusMap := make(map[string]uint64)

	for _, validator := range parsedValidators.Data {
		statusMap[string(validator.Status)] += 1
	}
	return &types.ValidatorQueue{
		Activating: statusMap["pending_queued"],
		Exiting:    statusMap["active_exiting"] + statusMap["active_slashed"],
	}, nil
}

// GetEpochAssignments will get the epoch assignments from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error) {
	var err error

	lc.assignmentsCacheMux.Lock()
	cachedValue, found := lc.assignmentsCache.Get(epoch)
	if found {
		lc.assignmentsCacheMux.Unlock()
		return cachedValue.(*types.EpochAssignments), nil
	}
	lc.assignmentsCacheMux.Unlock()

	parsedProposerResponse, err := lc.cl.GetPropoalAssignments(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving proposer duties for epoch %v: %w", epoch, err)
	}

	// fetch the block root that the proposer data is dependent on
	parsedHeader, err := lc.cl.GetBlockHeader(parsedProposerResponse.DependentRoot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving proposer duties dependent header for epoch %v: %w", epoch, err)
	}
	depStateRoot := parsedHeader.Data.Header.Message.StateRoot.String()

	parsedCommittees, err := lc.cl.GetCommittees(depStateRoot, &epoch, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving committees data: %w", err)
	}

	assignments := &types.EpochAssignments{
		ProposerAssignments: make(map[uint64]uint64),
		AttestorAssignments: make(map[string]uint64),
	}

	// propose
	for _, duty := range parsedProposerResponse.Data {
		assignments.ProposerAssignments[uint64(duty.Slot)] = duty.ValidatorIndex
	}

	// attest
	for _, committee := range parsedCommittees.Data {
		for i, valIndex := range committee.Validators {
			k := utils.FormatAttestorAssignmentKey(committee.Slot, committee.Index, uint64(i))
			assignments.AttestorAssignments[k] = uint64(valIndex)
		}
	}

	if epoch >= utils.Config.Chain.ClConfig.AltairForkEpoch {
		syncCommitteeState := depStateRoot
		if epoch == utils.Config.Chain.ClConfig.AltairForkEpoch {
			syncCommitteeState = fmt.Sprintf("%d", utils.Config.Chain.ClConfig.AltairForkEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
		parsedSyncCommittees, err := lc.GetSyncCommittee(syncCommitteeState, epoch)
		if err != nil {
			return nil, err
		}
		assignments.SyncAssignments = make([]uint64, len(parsedSyncCommittees.Validators))

		// sync
		for i, valIndex := range parsedSyncCommittees.Validators {
			assignments.SyncAssignments[i] = uint64(valIndex)
		}
	}

	if len(assignments.AttestorAssignments) > 0 && len(assignments.ProposerAssignments) > 0 {
		lc.assignmentsCacheMux.Lock()
		lc.assignmentsCache.Add(epoch, assignments)
		lc.assignmentsCacheMux.Unlock()
	}

	return assignments, nil
}

// GetEpochProposerAssignments will get the epoch proposer assignments from Lighthouse RPC api
// Deprecated: use cl retriever GetPropoalAssignments
func (lc *LighthouseClient) GetEpochProposerAssignments(epoch uint64) (*constypes.StandardProposerAssignmentsResponse, error) {
	return lc.cl.GetPropoalAssignments(epoch)
}

func (lc *LighthouseClient) GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error) {
	parsedValidators, err := lc.cl.GetValidators(epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, nil, nil)
	if err != nil && epoch == 0 {
		parsedValidators, err = lc.cl.GetValidators("genesis", nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving validators for genesis: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
	}

	return parsedValidators, nil
}

// GetEpochData will get the epoch data from Lighthouse RPC api
func (lc *LighthouseClient) GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error) {
	wg := &errgroup.Group{}
	mux := &sync.Mutex{}

	head, err := lc.GetChainHead()
	if err != nil {
		return nil, fmt.Errorf("error retrieving chain head: %w", err)
	}
	data := &types.EpochData{
		SyncDuties:        make(map[types.Slot]map[types.ValidatorIndex]bool),
		AttestationDuties: make(map[types.Slot]map[types.ValidatorIndex][]types.Slot),
	}
	data.Epoch = epoch
	if head.FinalizedEpoch >= epoch {
		data.Finalized = true
	}

	if head.FinalizedEpoch == 0 && epoch == 0 {
		data.Finalized = false
	}

	parsedValidators, err := lc.GetValidatorState(epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch validators: %w", err)
	}

	for _, validator := range parsedValidators.Data {
		data.Validators = append(data.Validators, &types.Validator{
			Index:                      validator.Index,
			PublicKey:                  validator.Validator.Pubkey,
			WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
			Balance:                    validator.Balance,
			EffectiveBalance:           validator.Validator.EffectiveBalance,
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.Validator.ActivationEpoch,
			ExitEpoch:                  validator.Validator.ExitEpoch,
			WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
			Status:                     string(validator.Status),
		})
	}

	log.Infof("retrieved data for %v validators for epoch %v", len(data.Validators), epoch)

	wg.Go(func() error {
		var err error
		data.ValidatorAssignmentes, err = lc.GetEpochAssignments(epoch)
		if err != nil {
			return fmt.Errorf("error retrieving assignments for epoch %v: %w", epoch, err)
		}

		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if data.SyncDuties[types.Slot(slot)] == nil {
				data.SyncDuties[types.Slot(slot)] = make(map[types.ValidatorIndex]bool)
			}
			for _, validatorIndex := range data.ValidatorAssignmentes.SyncAssignments {
				data.SyncDuties[types.Slot(slot)][types.ValidatorIndex(validatorIndex)] = false
			}
		}

		for key, validatorIndex := range data.ValidatorAssignmentes.AttestorAssignments {
			keySplit := strings.Split(key, "-")
			attestedSlot, err := strconv.ParseUint(keySplit[0], 10, 64)

			if err != nil {
				return fmt.Errorf("error parsing attested slot from attestation key: %w", err)
			}

			if data.AttestationDuties[types.Slot(attestedSlot)] == nil {
				data.AttestationDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}

			data.AttestationDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validatorIndex)] = []types.Slot{}
		}
		log.Infof("retrieved validator assignment data for epoch %v", epoch)
		return nil
	})

	if epoch < head.HeadEpoch {
		wg.Go(func() error {
			var err error
			data.EpochParticipationStats, err = lc.GetValidatorParticipation(epoch)
			if err != nil {
				if strings.HasSuffix(err.Error(), "can't be retrieved as it hasn't finished yet") { // should no longer happen
					log.Warnf("error retrieving epoch participation statistics for epoch %v: %v", epoch, err)
				} else {
					return fmt.Errorf("error retrieving epoch participation statistics for epoch %v: %w", epoch, err)
				}
				data.EpochParticipationStats = &types.ValidatorParticipation{
					Epoch:                   epoch,
					GlobalParticipationRate: 0.0,
					VotedEther:              0,
					EligibleEther:           0,
				}
			}
			return nil
		})
	} else {
		data.EpochParticipationStats = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: 0.0,
			VotedEther:              0,
			EligibleEther:           0,
		}
	}

	err = wg.Wait()
	if err != nil {
		return nil, err
	}
	wg = &errgroup.Group{}
	// Retrieve all blocks for the epoch
	data.Blocks = make(map[uint64]map[string]*types.Block)

	wg.Go(func() error {
		for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if slot != 0 && slot > head.HeadSlot { // don't export slots that have not occurred yet
				continue
			}
			start := time.Now()
			block, err := lc.GetBlockBySlot(slot)

			if err != nil {
				return fmt.Errorf("error retrieving block for slot %v: %w", slot, err)
			}

			mux.Lock()
			if data.Blocks[block.Slot] == nil {
				data.Blocks[block.Slot] = make(map[string]*types.Block)
			}
			data.Blocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

			for validator, duty := range block.SyncDuties {
				data.SyncDuties[types.Slot(block.Slot)][validator] = duty
			}
			for validator, attestedSlots := range block.AttestationDuties {
				for _, attestedSlot := range attestedSlots {
					if data.AttestationDuties[attestedSlot] == nil {
						data.AttestationDuties[attestedSlot] = make(map[types.ValidatorIndex][]types.Slot)
					}
					if data.AttestationDuties[attestedSlot][validator] == nil {
						data.AttestationDuties[attestedSlot][validator] = make([]types.Slot, 0, 10)
					}
					data.AttestationDuties[attestedSlot][validator] = append(data.AttestationDuties[attestedSlot][validator], types.Slot(block.Slot))
				}
			}
			mux.Unlock()
			log.Infof("processed data for current epoch slot %v in %v", slot, time.Since(start))
		}
		return nil
	})

	// we need future blocks to properly tracke fulfilled attestation duties
	data.FutureBlocks = make(map[uint64]map[string]*types.Block)
	wg.Go(func() error {
		for slot := (epoch + 1) * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= (epoch+2)*utils.Config.Chain.ClConfig.SlotsPerEpoch-1; slot++ {
			if slot != 0 && slot > head.HeadSlot { // don't export slots that have not occurred yet
				continue
			}
			start := time.Now()
			block, err := lc.GetBlockBySlot(slot)

			if err != nil {
				return fmt.Errorf("error retrieving block for slot %v: %w", slot, err)
			}

			mux.Lock()
			if data.FutureBlocks[block.Slot] == nil {
				data.FutureBlocks[block.Slot] = make(map[string]*types.Block)
			}
			data.FutureBlocks[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

			// fill out performed attestation duties
			for validator, attestedSlots := range block.AttestationDuties {
				for _, attestedSlot := range attestedSlots {
					if attestedSlot < types.Slot((epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch) {
						data.AttestationDuties[attestedSlot][validator] = append(data.AttestationDuties[attestedSlot][validator], types.Slot(block.Slot))
					}
				}
			}
			mux.Unlock()
			log.Infof("processed data for next epoch slot %v in %v", slot, time.Since(start))
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	log.Infof("retrieved %v blocks for epoch %v", len(data.Blocks), epoch)

	if data.ValidatorAssignmentes == nil {
		return data, fmt.Errorf("no assignments for epoch %v", epoch)
	}

	// Fill up missed and scheduled blocks
	for slot, proposer := range data.ValidatorAssignmentes.ProposerAssignments {
		_, found := data.Blocks[slot]
		if !found {
			// Proposer was assigned but did not yet propose a block
			data.Blocks[slot] = make(map[string]*types.Block)
			data.Blocks[slot]["0x0"] = &types.Block{
				Status:            0,
				Proposer:          proposer,
				BlockRoot:         []byte{0x0},
				Slot:              slot,
				ParentRoot:        []byte{},
				StateRoot:         []byte{},
				Signature:         []byte{},
				RandaoReveal:      []byte{},
				Graffiti:          []byte{},
				BodyRoot:          []byte{},
				Eth1Data:          &types.Eth1Data{},
				ProposerSlashings: make([]*types.ProposerSlashing, 0),
				AttesterSlashings: make([]*types.AttesterSlashing, 0),
				Attestations:      make([]*types.Attestation, 0),
				Deposits:          make([]*types.Deposit, 0),
				VoluntaryExits:    make([]*types.VoluntaryExit, 0),
				SyncAggregate:     nil,
			}

			if utils.SlotToTime(slot).After(time.Now().Add(time.Second * -4)) {
				// Block is in the future, set status to scheduled
				data.Blocks[slot]["0x0"].Status = 0
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x0}
			} else {
				// Block is in the past, set status to missed
				data.Blocks[slot]["0x0"].Status = 2
				data.Blocks[slot]["0x0"].BlockRoot = []byte{0x1}
			}
		}
	}

	// for validator, duty := range data.AttestationDuties {
	// 	for slot, inclusions := range duty {
	// 		log.LogInfo("validator %v has attested for slot %v in slots %v", validator, slot, inclusions)
	// 	}
	// }

	return data, nil
}

func uint64List(li []constypes.Uint64Str) []uint64 {
	out := make([]uint64, len(li))
	for i, v := range li {
		out[i] = uint64(v)
	}
	return out
}

func (lc *LighthouseClient) GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error) {
	if epoch < 0 {
		epoch = 0
	}

	var err error

	validatorBalances := make(map[uint64]uint64)

	parsedResponse, err := lc.cl.GetValidatorBalances(epoch * int64(utils.Config.Chain.ClConfig.SlotsPerEpoch))
	if err != nil && epoch == 0 {
		parsedResponse, err = lc.cl.GetValidatorBalances("genesis")
		if err != nil {
			return validatorBalances, err
		}
	} else if err != nil {
		return validatorBalances, err
	}

	for _, b := range parsedResponse.Data {
		validatorBalances[b.Index] = b.Balance
	}

	return validatorBalances, nil
}

func (lc *LighthouseClient) GetBlockByBlockroot(blockroot []byte) (*types.Block, error) {
	parsedHeaders, err := lc.cl.GetBlockHeader(fmt.Sprintf("0x%x", blockroot))
	if err != nil {
		httpErr := network.SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			// no block found
			return &types.Block{}, nil
		}
		return nil, fmt.Errorf("error retrieving headers for blockroot 0x%x: %w", blockroot, err)
	}

	slot := parsedHeaders.Data.Header.Message.Slot

	parsedResponse, err := lc.cl.GetSlot(parsedHeaders.Data.Root.String())
	if err != nil {
		log.Error(err, "error parsing block data for slot", 0, map[string]interface{}{"slot": parsedHeaders.Data.Header.Message.Slot})
		return nil, fmt.Errorf("error retrieving block data at slot %v: %w", slot, err)
	}

	return lc.blockFromResponse(parsedHeaders, parsedResponse)
}

// GetBlockHeader will get the block header by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error) {
	parsedHeaders, err := lc.cl.GetBlockHeader(slot)

	if err != nil && slot == 0 {
		parsedHeader, err := lc.cl.GetBlockHeaders(nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving chain head for slot %v: %w", slot, err)
		}

		if len(parsedHeader.Data) == 0 {
			return nil, fmt.Errorf("error no headers available for slot %v", slot)
		}

		parsedHeaders = &constypes.StandardBeaconHeaderResponse{
			Data: parsedHeader.Data[len(parsedHeader.Data)-1],
		}
	} else if err != nil {
		httpErr := network.SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			// no block found
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving headers at slot %v: %w", slot, err)
	}

	return parsedHeaders, nil
}

// GetBlocksBySlot will get the blocks by slot from Lighthouse RPC api
func (lc *LighthouseClient) GetBlockBySlot(slot uint64) (*types.Block, error) {
	epoch := slot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	isFirstSlotOfEpoch := slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0

	parsedHeaders, err := lc.GetBlockHeader(slot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving headers at slot %v: %w", slot, err)
	}

	if parsedHeaders == nil { // not found
		proposerAssignments, err := lc.GetEpochProposerAssignments(epoch)
		if err != nil {
			return nil, err
		}

		proposer := uint64(math.MaxUint64)
		for _, pa := range proposerAssignments.Data {
			if uint64(pa.Slot) == slot {
				proposer = pa.ValidatorIndex
			}
		}

		block := &types.Block{
			Status:            0,
			Proposer:          proposer,
			BlockRoot:         []byte{0x0},
			Slot:              slot,
			ParentRoot:        []byte{},
			StateRoot:         []byte{},
			Signature:         []byte{},
			RandaoReveal:      []byte{},
			Graffiti:          []byte{},
			BodyRoot:          []byte{},
			Eth1Data:          &types.Eth1Data{},
			ProposerSlashings: make([]*types.ProposerSlashing, 0),
			AttesterSlashings: make([]*types.AttesterSlashing, 0),
			Attestations:      make([]*types.Attestation, 0),
			Deposits:          make([]*types.Deposit, 0),
			VoluntaryExits:    make([]*types.VoluntaryExit, 0),
			SyncAggregate:     nil,
		}

		if isFirstSlotOfEpoch {
			assignments, err := lc.GetEpochAssignments(epoch)
			if err != nil {
				return nil, err
			}

			block.EpochAssignments = assignments

			parsedValidators, err := lc.GetValidatorState(epoch)
			if err != nil {
				return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
			}
			block.Validators = make([]*types.Validator, 0, len(parsedValidators.Data))
			for _, validator := range parsedValidators.Data {
				block.Validators = append(block.Validators, &types.Validator{
					Index:                      validator.Index,
					PublicKey:                  validator.Validator.Pubkey,
					WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
					Balance:                    validator.Balance,
					EffectiveBalance:           validator.Validator.EffectiveBalance,
					Slashed:                    validator.Validator.Slashed,
					ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
					ActivationEpoch:            validator.Validator.ActivationEpoch,
					ExitEpoch:                  validator.Validator.ExitEpoch,
					WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
					Status:                     string(validator.Status),
				})
			}
		}

		return block, nil
	}

	lc.slotsCacheMux.Lock()
	cachedBlock, ok := lc.slotsCache.Get(parsedHeaders.Data.Root.String())
	if ok {
		lc.slotsCacheMux.Unlock()
		block, ok := cachedBlock.(*types.Block)

		if ok {
			log.Infof("retrieved slot %v (0x%x) from in memory cache", block.Slot, block.BlockRoot)
			return block, nil
		} else {
			log.Error(fmt.Errorf("unable to convert cached block to block type"), "", 0)
		}
	}
	lc.slotsCacheMux.Unlock()

	parsedResponse, err := lc.cl.GetSlot(parsedHeaders.Data.Root.String())
	if err != nil && slot == 0 {
		log.Error(err, "error parsing block data for slot", 0, map[string]interface{}{"slot": parsedHeaders.Data.Header.Message.Slot})

		return nil, fmt.Errorf("error retrieving block data at slot %v: %w", slot, err)
	}

	block, err := lc.blockFromResponse(parsedHeaders, parsedResponse)
	if err != nil {
		return nil, err
	}

	// for the first slot of an epoch, also retrieve the epoch assignments
	if block.Slot%utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 {
		var err error
		block.EpochAssignments, err = lc.GetEpochAssignments(block.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
		if err != nil {
			return nil, err
		}

		parsedValidators, err := lc.GetValidatorState(epoch)
		if err != nil {
			return nil, fmt.Errorf("error retrieving validators for epoch %v: %w", epoch, err)
		}

		block.Validators = make([]*types.Validator, 0, len(parsedValidators.Data))
		for _, validator := range parsedValidators.Data {
			block.Validators = append(block.Validators, &types.Validator{
				Index:                      validator.Index,
				PublicKey:                  validator.Validator.Pubkey,
				WithdrawalCredentials:      validator.Validator.WithdrawalCredentials,
				Balance:                    validator.Balance,
				EffectiveBalance:           validator.Validator.EffectiveBalance,
				Slashed:                    validator.Validator.Slashed,
				ActivationEligibilityEpoch: validator.Validator.ActivationEligibilityEpoch,
				ActivationEpoch:            validator.Validator.ActivationEpoch,
				ExitEpoch:                  validator.Validator.ExitEpoch,
				WithdrawableEpoch:          validator.Validator.WithdrawableEpoch,
				Status:                     string(validator.Status),
			})
		}
	}

	lc.slotsCacheMux.Lock()
	lc.slotsCache.Add(parsedHeaders.Data.Root.String(), block)
	lc.slotsCacheMux.Unlock()

	return block, nil
}

func (lc *LighthouseClient) blockFromResponse(parsedHeaders *constypes.StandardBeaconHeaderResponse, parsedResponse *constypes.StandardBeaconSlotResponse) (*types.Block, error) {
	parsedBlock := parsedResponse.Data
	slot := parsedHeaders.Data.Header.Message.Slot
	block := &types.Block{
		Status:       1,
		Finalized:    parsedHeaders.Finalized,
		Proposer:     parsedBlock.Message.ProposerIndex,
		BlockRoot:    parsedHeaders.Data.Root,
		Slot:         slot,
		ParentRoot:   parsedBlock.Message.ParentRoot,
		StateRoot:    parsedBlock.Message.StateRoot,
		Signature:    parsedBlock.Signature,
		RandaoReveal: parsedBlock.Message.Body.RandaoReveal,
		Graffiti:     parsedBlock.Message.Body.Graffiti,
		Eth1Data: &types.Eth1Data{
			DepositRoot:  parsedBlock.Message.Body.Eth1Data.DepositRoot,
			DepositCount: parsedBlock.Message.Body.Eth1Data.DepositCount,
			BlockHash:    parsedBlock.Message.Body.Eth1Data.BlockHash,
		},
		ProposerSlashings:          make([]*types.ProposerSlashing, len(parsedBlock.Message.Body.ProposerSlashings)),
		AttesterSlashings:          make([]*types.AttesterSlashing, len(parsedBlock.Message.Body.AttesterSlashings)),
		Attestations:               make([]*types.Attestation, len(parsedBlock.Message.Body.Attestations)),
		Deposits:                   make([]*types.Deposit, len(parsedBlock.Message.Body.Deposits)),
		VoluntaryExits:             make([]*types.VoluntaryExit, len(parsedBlock.Message.Body.VoluntaryExits)),
		SignedBLSToExecutionChange: make([]*types.SignedBLSToExecutionChange, len(parsedBlock.Message.Body.SignedBLSToExecutionChange)),
		BlobKZGCommitments:         make([][]byte, len(parsedBlock.Message.Body.BlobKZGCommitments)),
		BlobKZGProofs:              make([][]byte, len(parsedBlock.Message.Body.BlobKZGCommitments)),
		AttestationDuties:          make(map[types.ValidatorIndex][]types.Slot),
		SyncDuties:                 make(map[types.ValidatorIndex]bool),
	}

	for i, c := range parsedBlock.Message.Body.BlobKZGCommitments {
		block.BlobKZGCommitments[i] = c
	}

	if len(parsedBlock.Message.Body.BlobKZGCommitments) > 0 {
		res, err := lc.GetBlobSidecars(fmt.Sprintf("%#x", block.BlockRoot))
		if err != nil {
			return nil, err
		}
		if len(res.Data) != len(parsedBlock.Message.Body.BlobKZGCommitments) {
			return nil, fmt.Errorf("error constructing block at slot %v: len(blob_sidecars) != len(block.blob_kzg_commitments): %v != %v", block.Slot, len(res.Data), len(parsedBlock.Message.Body.BlobKZGCommitments))
		}
		for i, d := range res.Data {
			if !bytes.Equal(d.KzgCommitment, block.BlobKZGCommitments[i]) {
				return nil, fmt.Errorf("error constructing block at slot %v: unequal kzg_commitments at index %v: %#x != %#x", block.Slot, i, d.KzgCommitment, block.BlobKZGCommitments[i])
			}
			block.BlobKZGProofs[i] = d.KzgProof
		}
	}

	epochAssignments, err := lc.GetEpochAssignments(slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return nil, err
	}

	if agg := parsedBlock.Message.Body.SyncAggregate; agg != nil {
		bits := agg.SyncCommitteeBits

		if utils.Config.Chain.ClConfig.SyncCommitteeSize != uint64(len(bits)*8) {
			return nil, fmt.Errorf("sync-aggregate-bits-size does not match sync-committee-size: %v != %v", len(bits)*8, utils.Config.Chain.ClConfig.SyncCommitteeSize)
		}

		block.SyncAggregate = &types.SyncAggregate{
			SyncCommitteeValidators:    epochAssignments.SyncAssignments,
			SyncCommitteeBits:          bits,
			SyncAggregateParticipation: syncCommitteeParticipation(bits),
			SyncCommitteeSignature:     agg.SyncCommitteeSignature,
		}

		// fill out performed sync duties
		bitLen := len(block.SyncAggregate.SyncCommitteeBits) * 8
		valLen := len(block.SyncAggregate.SyncCommitteeValidators)
		if bitLen < valLen {
			return nil, fmt.Errorf("error getting sync_committee participants: bitLen != valLen: %v != %v", bitLen, valLen)
		}
		for i, valIndex := range block.SyncAggregate.SyncCommitteeValidators {
			block.SyncDuties[types.ValidatorIndex(valIndex)] = utils.BitAtVector(block.SyncAggregate.SyncCommitteeBits, i)
		}
	}

	if payload := parsedBlock.Message.Body.ExecutionPayload; payload != nil && !bytes.Equal(payload.ParentHash, make([]byte, 32)) {
		txs := make([]*types.Transaction, 0, len(payload.Transactions))
		for i, rawTx := range payload.Transactions {
			tx := &types.Transaction{Raw: rawTx}
			var decTx gethtypes.Transaction
			if err := decTx.UnmarshalBinary(rawTx); err != nil {
				return nil, fmt.Errorf("error parsing tx %d block %x: %w", i, payload.BlockHash, err)
			} else {
				h := decTx.Hash()
				tx.TxHash = h[:]
				tx.AccountNonce = decTx.Nonce()
				// big endian
				tx.Price = decTx.GasPrice().Bytes()
				tx.GasLimit = decTx.Gas()
				sender, err := lc.signer.Sender(&decTx)
				if err != nil {
					return nil, fmt.Errorf("transaction with invalid sender (slot: %v, tx-hash: %x): %w", slot, h, err)
				}
				tx.Sender = sender.Bytes()
				if v := decTx.To(); v != nil {
					tx.Recipient = v.Bytes()
				} else {
					tx.Recipient = []byte{}
				}
				tx.Amount = decTx.Value().Bytes()
				tx.Payload = decTx.Data()
				tx.MaxPriorityFeePerGas = decTx.GasTipCap().Uint64()
				tx.MaxFeePerGas = decTx.GasFeeCap().Uint64()

				if decTx.BlobGasFeeCap() != nil {
					tx.MaxFeePerBlobGas = decTx.BlobGasFeeCap().Uint64()
				}
				for _, h := range decTx.BlobHashes() {
					tx.BlobVersionedHashes = append(tx.BlobVersionedHashes, h.Bytes())
				}
			}
			txs = append(txs, tx)
		}
		withdrawals := make([]*types.Withdrawals, 0, len(payload.Withdrawals))
		for _, w := range payload.Withdrawals {
			withdrawals = append(withdrawals, &types.Withdrawals{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}

		block.ExecutionPayload = &types.ExecutionPayload{
			ParentHash:    payload.ParentHash,
			FeeRecipient:  payload.FeeRecipient,
			StateRoot:     payload.StateRoot,
			ReceiptsRoot:  payload.ReceiptsRoot,
			LogsBloom:     payload.LogsBloom,
			Random:        payload.PrevRandao,
			BlockNumber:   payload.BlockNumber,
			GasLimit:      payload.GasLimit,
			GasUsed:       payload.GasUsed,
			Timestamp:     payload.Timestamp,
			ExtraData:     payload.ExtraData,
			BaseFeePerGas: payload.BaseFeePerGas,
			BlockHash:     payload.BlockHash,
			Transactions:  txs,
			Withdrawals:   withdrawals,
			BlobGasUsed:   payload.BlobGasUsed,
			ExcessBlobGas: payload.ExcessBlobGas,
		}
	}

	// TODO: this is legacy from old lighthouse API. Does it even still apply?
	if block.Eth1Data.DepositCount > 2147483647 { // Sometimes the lighthouse node does return bogus data for the DepositCount value
		block.Eth1Data.DepositCount = 0
	}

	for i, proposerSlashing := range parsedBlock.Message.Body.ProposerSlashings {
		block.ProposerSlashings[i] = &types.ProposerSlashing{
			ProposerIndex: proposerSlashing.SignedHeader1.Message.ProposerIndex,
			Header1: &types.Block{
				Slot:       proposerSlashing.SignedHeader1.Message.Slot,
				ParentRoot: proposerSlashing.SignedHeader1.Message.ParentRoot,
				StateRoot:  proposerSlashing.SignedHeader1.Message.StateRoot,
				Signature:  proposerSlashing.SignedHeader1.Signature,
				BodyRoot:   proposerSlashing.SignedHeader1.Message.BodyRoot,
			},
			Header2: &types.Block{
				Slot:       proposerSlashing.SignedHeader2.Message.Slot,
				ParentRoot: proposerSlashing.SignedHeader2.Message.ParentRoot,
				StateRoot:  proposerSlashing.SignedHeader2.Message.StateRoot,
				Signature:  proposerSlashing.SignedHeader2.Signature,
				BodyRoot:   proposerSlashing.SignedHeader2.Message.BodyRoot,
			},
		}
	}

	for i, attesterSlashing := range parsedBlock.Message.Body.AttesterSlashings {
		block.AttesterSlashings[i] = &types.AttesterSlashing{
			Attestation1: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            attesterSlashing.Attestation1.Data.Slot,
					CommitteeIndex:  attesterSlashing.Attestation1.Data.Index,
					BeaconBlockRoot: attesterSlashing.Attestation1.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: attesterSlashing.Attestation1.Data.Source.Epoch,
						Root:  attesterSlashing.Attestation1.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: attesterSlashing.Attestation1.Data.Target.Epoch,
						Root:  attesterSlashing.Attestation1.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation1.Signature,
				AttestingIndices: uint64List(attesterSlashing.Attestation1.AttestingIndices),
			},
			Attestation2: &types.IndexedAttestation{
				Data: &types.AttestationData{
					Slot:            attesterSlashing.Attestation2.Data.Slot,
					CommitteeIndex:  attesterSlashing.Attestation2.Data.Index,
					BeaconBlockRoot: attesterSlashing.Attestation2.Data.BeaconBlockRoot,
					Source: &types.Checkpoint{
						Epoch: attesterSlashing.Attestation2.Data.Source.Epoch,
						Root:  attesterSlashing.Attestation2.Data.Source.Root,
					},
					Target: &types.Checkpoint{
						Epoch: attesterSlashing.Attestation2.Data.Target.Epoch,
						Root:  attesterSlashing.Attestation2.Data.Target.Root,
					},
				},
				Signature:        attesterSlashing.Attestation2.Signature,
				AttestingIndices: uint64List(attesterSlashing.Attestation2.AttestingIndices),
			},
		}
	}

	for i, attestation := range parsedBlock.Message.Body.Attestations {
		a := &types.Attestation{
			AggregationBits: attestation.AggregationBits,
			Attesters:       []uint64{},
			Data: &types.AttestationData{
				Slot:            attestation.Data.Slot,
				CommitteeIndex:  attestation.Data.Index,
				BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
				Source: &types.Checkpoint{
					Epoch: attestation.Data.Source.Epoch,
					Root:  attestation.Data.Source.Root,
				},
				Target: &types.Checkpoint{
					Epoch: attestation.Data.Target.Epoch,
					Root:  attestation.Data.Target.Root,
				},
			},
			Signature: attestation.Signature,
		}

		aggregationBits := bitfield.Bitlist(a.AggregationBits)
		assignments, err := lc.GetEpochAssignments(a.Data.Slot / utils.Config.Chain.ClConfig.SlotsPerEpoch)
		if err != nil {
			return nil, fmt.Errorf("error receiving epoch assignment for epoch %v: %w", a.Data.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch, err)
		}

		for i := uint64(0); i < aggregationBits.Len(); i++ {
			if aggregationBits.BitAt(i) {
				validator, found := assignments.AttestorAssignments[utils.FormatAttestorAssignmentKey(a.Data.Slot, uint64(a.Data.CommitteeIndex), i)]
				if !found { // This should never happen!
					validator = 0
					log.Fatal(fmt.Errorf("error retrieving assigned validator for attestation %v of block %v for slot %v committee index %v member index %v", i, block.Slot, a.Data.Slot, a.Data.CommitteeIndex, i), "", 0)
				}
				a.Attesters = append(a.Attesters, validator)

				if block.AttestationDuties[types.ValidatorIndex(validator)] == nil {
					block.AttestationDuties[types.ValidatorIndex(validator)] = []types.Slot{types.Slot(a.Data.Slot)}
				} else {
					block.AttestationDuties[types.ValidatorIndex(validator)] = append(block.AttestationDuties[types.ValidatorIndex(validator)], types.Slot(a.Data.Slot))
				}
			}
		}

		block.Attestations[i] = a
	}

	for i, deposit := range parsedBlock.Message.Body.Deposits {
		d := &types.Deposit{
			Proof:                 nil,
			PublicKey:             deposit.Data.Pubkey,
			WithdrawalCredentials: deposit.Data.WithdrawalCredentials,
			Amount:                deposit.Data.Amount,
			Signature:             deposit.Data.Signature,
		}

		block.Deposits[i] = d
	}

	for i, voluntaryExit := range parsedBlock.Message.Body.VoluntaryExits {
		block.VoluntaryExits[i] = &types.VoluntaryExit{
			Epoch:          voluntaryExit.Message.Epoch,
			ValidatorIndex: voluntaryExit.Message.ValidatorIndex,
			Signature:      voluntaryExit.Signature,
		}
	}

	for i, blsChange := range parsedBlock.Message.Body.SignedBLSToExecutionChange {
		block.SignedBLSToExecutionChange[i] = &types.SignedBLSToExecutionChange{
			Message: types.BLSToExecutionChange{
				Validatorindex: blsChange.Message.ValidatorIndex,
				BlsPubkey:      blsChange.Message.FromBlsPubkey,
				Address:        blsChange.Message.ToExecutionAddress,
			},
			Signature: blsChange.Signature,
		}
	}

	return block, nil
}

func syncCommitteeParticipation(bits []byte) float64 {
	participating := 0
	for i := 0; i < int(utils.Config.Chain.ClConfig.SyncCommitteeSize); i++ {
		if utils.BitAtVector(bits, i) {
			participating++
		}
	}
	return float64(participating) / float64(utils.Config.Chain.ClConfig.SyncCommitteeSize)
}

// GetValidatorParticipation will get the validator participation from the Lighthouse RPC api
func (lc *LighthouseClient) GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error) {
	head, err := lc.GetChainHead()
	if err != nil {
		return nil, err
	}

	if epoch > head.HeadEpoch {
		return nil, fmt.Errorf("epoch %v is newer than the latest head %v", epoch, LighthouseLatestHeadEpoch)
	}
	if epoch == head.HeadEpoch {
		// participation stats are calculated at the end of an epoch,
		// making it impossible to retrieve stats of an currently ongoing epoch
		return nil, fmt.Errorf("epoch %v can't be retrieved as it hasn't finished yet", epoch)
	}

	request_epoch := epoch

	if epoch+1 < head.HeadEpoch {
		request_epoch += 1
	}

	log.Infof("requesting validator inclusion data for epoch %v", request_epoch)

	parsedResponse, err := network.Get[LighthouseValidatorParticipationResponse](nil, fmt.Sprintf("%s/lighthouse/validator_inclusion/%d/global", lc.cl.Endpoint, request_epoch))
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator participation data for epoch %v: %w", request_epoch, err)
	}

	var res *types.ValidatorParticipation
	if epoch < request_epoch {
		// we requested the next epoch, so we have to use the previous value for everything here

		prevEpochActiveGwei := parsedResponse.Data.PreviousEpochActiveGwei
		if prevEpochActiveGwei == 0 {
			// lh@5.2.0+ has no previous_epoch_active_gwei field anymore, see https://github.com/sigp/lighthouse/pull/5279
			parsedPrevResponse, err := network.Get[LighthouseValidatorParticipationResponse](nil, fmt.Sprintf("%s/lighthouse/validator_inclusion/%d/global", lc.cl.Endpoint, request_epoch-1))
			if err != nil {
				return nil, fmt.Errorf("error retrieving validator participation data for prevEpoch %v: %w", request_epoch-1, err)
			}
			prevEpochActiveGwei = parsedPrevResponse.Data.CurrentEpochActiveGwei
		}

		res = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: float32(parsedResponse.Data.PreviousEpochTargetAttestingGwei) / float32(prevEpochActiveGwei),
			VotedEther:              uint64(parsedResponse.Data.PreviousEpochTargetAttestingGwei),
			EligibleEther:           uint64(prevEpochActiveGwei),
			Finalized:               epoch <= head.FinalizedEpoch && head.JustifiedEpoch > 0,
		}
	} else {
		res = &types.ValidatorParticipation{
			Epoch:                   epoch,
			GlobalParticipationRate: float32(parsedResponse.Data.CurrentEpochTargetAttestingGwei) / float32(parsedResponse.Data.CurrentEpochActiveGwei),
			VotedEther:              uint64(parsedResponse.Data.CurrentEpochTargetAttestingGwei),
			EligibleEther:           uint64(parsedResponse.Data.CurrentEpochActiveGwei),
			Finalized:               epoch <= head.FinalizedEpoch && head.JustifiedEpoch > 0,
		}
	}
	return res, nil
}

func (lc *LighthouseClient) GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error) {
	parsedSyncCommittees, err := lc.cl.GetSyncCommitteesAssignments(&epoch, stateID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sync_committees for epoch %v (state: %v): %w", epoch, stateID, err)
	}

	return &parsedSyncCommittees.Data, nil
}

func (lc *LighthouseClient) GetBlobSidecars(stateID string) (*constypes.StandardBlobSidecarsResponse, error) {
	return lc.cl.GetBlobSidecars(stateID)
}

type LighthouseValidatorParticipationResponse struct {
	Data struct {
		CurrentEpochActiveGwei           constypes.Uint64Str `json:"current_epoch_active_gwei"`
		PreviousEpochActiveGwei          constypes.Uint64Str `json:"previous_epoch_active_gwei"`
		CurrentEpochTargetAttestingGwei  constypes.Uint64Str `json:"current_epoch_target_attesting_gwei"`
		PreviousEpochTargetAttestingGwei constypes.Uint64Str `json:"previous_epoch_target_attesting_gwei"`
		PreviousEpochHeadAttestingGwei   constypes.Uint64Str `json:"previous_epoch_head_attesting_gwei"`
	} `json:"data"`
}

// https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
// https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
type ExecutionPayload struct {
	ParentHash    hexutil.Bytes       `json:"parent_hash"`
	FeeRecipient  hexutil.Bytes       `json:"fee_recipient"`
	StateRoot     hexutil.Bytes       `json:"state_root"`
	ReceiptsRoot  hexutil.Bytes       `json:"receipts_root"`
	LogsBloom     hexutil.Bytes       `json:"logs_bloom"`
	PrevRandao    hexutil.Bytes       `json:"prev_randao"`
	BlockNumber   constypes.Uint64Str `json:"block_number"`
	GasLimit      constypes.Uint64Str `json:"gas_limit"`
	GasUsed       constypes.Uint64Str `json:"gas_used"`
	Timestamp     constypes.Uint64Str `json:"timestamp"`
	ExtraData     hexutil.Bytes       `json:"extra_data"`
	BaseFeePerGas constypes.Uint64Str `json:"base_fee_per_gas"`
	BlockHash     hexutil.Bytes       `json:"block_hash"`
	Transactions  []hexutil.Bytes     `json:"transactions"`
	// present only after capella
	Withdrawals []constypes.WithdrawalPayload `json:"withdrawals"`
	// present only after deneb
	BlobGasUsed   constypes.Uint64Str `json:"blob_gas_used"`
	ExcessBlobGas constypes.Uint64Str `json:"excess_blob_gas"`
}
