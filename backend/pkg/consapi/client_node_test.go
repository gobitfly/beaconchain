package consapi_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/consapi/network"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

var client consapi.Client

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	client = consapi.NewNodeDataRetriever("http://localhost:32787")
}

func TestGetBlockHeader(t *testing.T) {
	res, err := client.GetBlockHeader("head")
	if err != nil {
		t.Errorf("Error getting block header: %v", err)
	}
	log.Printf("Block header: %v\n", res)
}

func TestGetSlot(t *testing.T) {
	res, err := client.GetSlot(0)
	if err != nil {
		t.Errorf("Error getting slot: %v", err)
	}
	log.Printf("Slot: %v\n", res)
}

func TestGetValidators(t *testing.T) {
	res, err := client.GetValidators("head", nil, nil)
	if err != nil {
		t.Errorf("Error getting validators: %v", err)
	}
	log.Printf("Validators: %v\n", res)
}

func TestGetValidatorsFilter(t *testing.T) {
	filter := types.ActiveSlashed
	res, err := client.GetValidators("head", nil, []types.ValidatorStatus{filter})
	if err != nil {
		t.Errorf("Error getting validators: %v", err)
	}
	for _, v := range res.Data {
		if v.Status != filter {
			t.Errorf("Invalid validator status: %v", v.Status)
		}
	}
}

func TestGetValidatorsFilterIndex(t *testing.T) {
	filter := []string{"4", "5", "6"}
	res, err := client.GetValidators("head", filter, nil)
	if err != nil {
		t.Errorf("Error getting validators: %v", err)
	}
	for _, v := range res.Data {
		if !contains(filter, fmt.Sprintf("%v", v.Index)) {
			t.Errorf("Error getting validators: %v", v.Index)
		}
	}
}

func TestGetValidatorsFilterBoth(t *testing.T) {
	filter := []string{"4", "5", "6"}
	filterStatus := types.ActiveOngoing
	res, err := client.GetValidators("head", filter, []types.ValidatorStatus{filterStatus})
	if err != nil {
		t.Errorf("Error getting validators: %v", err)
	}
	for _, v := range res.Data {
		if !contains(filter, fmt.Sprintf("%v", v.Index)) {
			t.Errorf("Error getting validators: %v", v.Index)
		}
	}
}

func TestGetPropoalAssignments(t *testing.T) {
	res, err := client.GetPropoalAssignments(0)
	if err != nil {
		t.Errorf("Error getting proposal assignments: %v", err)
	}
	log.Printf("Proposal assignments: %v\n", res)
}

func TestGetPropoalRewards(t *testing.T) {
	res, err := client.GetPropoalRewards("head")
	if err != nil {
		t.Errorf("Error getting proposal rewards: %v", err)
	}
	log.Printf("Proposal rewards: %v\n", res)
}

func TestGetSyncRewards(t *testing.T) {
	res, err := client.GetSyncRewards("head")
	if err != nil {
		t.Errorf("Error getting sync rewards: %v", err)
	}
	log.Printf("Sync rewards: %v\n", res)
}

func TestGetAttestationRewards(t *testing.T) {
	res, err := client.GetAttestationRewards(0)
	if err != nil {
		t.Errorf("Error getting attestation rewards: %v", err)
	}
	log.Printf("Attestation rewards: %v\n", res)
}

func TestGetSyncCommitteesAssignments(t *testing.T) {
	res, err := client.GetSyncCommitteesAssignments(0, "head")
	if err != nil {
		t.Errorf("Error getting sync committees assignments: %v", err)
	}
	log.Printf("Sync committees assignments: %v\n", res)
}

func TestGetSpec(t *testing.T) {
	res, err := client.GetSpec()
	if err != nil {
		httpErr, rpcErr := network.SpecificError(err)
		if httpErr != nil {
			t.Errorf("Error getting spec, http error: %v", err)
		} else if rpcErr != nil {
			t.Errorf("Error getting spec, rpc error: %v", err)
		}
		t.Errorf("Error getting spec: %v", err)
	}

	log.Printf("Spec: %v\n", res)
}

func TestGetBlockHeaders(t *testing.T) {
	res, err := client.GetBlockHeaders(nil, nil)
	if err != nil {
		t.Errorf("Error getting block headers: %v", err)
	}
	log.Printf("Block headers: %v\n", res)
}

func TestGetBlockHeadersSlot(t *testing.T) {
	slot := uint64(3)
	res, err := client.GetBlockHeaders(&slot, nil)
	if err != nil {
		t.Errorf("Error getting block headers: %v", err)
	}
	log.Printf("Block headers: %v\n", res)
}

func TestGetFinalityCheckpoints(t *testing.T) {
	res, err := client.GetFinalityCheckpoints("head")
	if err != nil {
		t.Errorf("Error getting finality checkpoints: %v", err)
	}
	log.Printf("Finality checkpoints: %v\n", res)
}

func TestGetValidatorBalances(t *testing.T) {
	res, err := client.GetValidatorBalances("head")
	if err != nil {
		t.Errorf("Error getting validator balances: %v", err)
	}
	log.Printf("Validator balances: %v\n", res)
}

func TestGetBlobSidecars(t *testing.T) {
	res, err := client.GetBlobSidecars("head")
	if err != nil {
		t.Errorf("Error getting blob sidecars: %v", err)
	}
	log.Printf("Blob sidecars: %v\n", res)
}

func TestGetCommittees(t *testing.T) {
	res, err := client.GetCommittees("head", nil, nil, nil)
	if err != nil {
		t.Errorf("Error getting committees: %v", err)
	}
	log.Printf("Committees: %v\n", res)
}

func TestGetGenesis(t *testing.T) {
	res, err := client.GetGenesis()
	if err != nil {
		t.Errorf("Error getting genesis: %v", err)
	}
	log.Printf("Genesis: %v\n", res)
}

func TestGetEvents(t *testing.T) {
	res := client.GetEvents([]types.EventTopic{types.EventHead, types.EventBlock, types.EventChainReorg, types.EventFinalizedCheckpoint})

	for event := range res {
		if event.Error != nil {
			t.Errorf("Error getting event: %v", event.Error)
		}

		if event.Event == types.EventHead {
			response, err := event.Head()
			if err != nil {
				t.Errorf("Error getting head event: %v", err)
			}
			log.Printf("Head: %v\n", response)
		}

		if event.Event == types.EventBlock {
			response, err := event.Block()
			if err != nil {
				t.Errorf("Error getting block event: %v", err)
			}
			log.Printf("Block: %v\n", response)
		}

		if event.Event == types.EventChainReorg {
			response, err := event.ChainReorg()
			if err != nil {
				t.Errorf("Error getting chain reorg event: %v", err)
			}
			log.Printf("Chain reorg: %v\n", response)
		}

		if event.Event == types.EventFinalizedCheckpoint {
			response, err := event.FinalizedCheckpoint()
			if err != nil {
				t.Errorf("Error getting finalized checkpoint event: %v", err)
			}
			log.Printf("Finalized checkpoint: %v\n", response)
		}
	}
}

func contains(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
