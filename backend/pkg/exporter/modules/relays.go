package modules

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	exportertypes "github.com/gobitfly/beaconchain/pkg/exporter/types"
)

func mevBoostRelaysExporter() {
	for {
		// we retrieve the relays from the db each loop to prevent having to restart the exporter for changes
		relays, err := db.GetRelays()
		if err != nil {
			log.Error(err, "failed to retrieve relays from db", 0)
			time.Sleep(time.Minute)
			continue
		}

		exportRelaysConcurrently(relays)
		time.Sleep(time.Minute)
	}
}

func exportRelaysConcurrently(relays []types.Relay) {
	var wg sync.WaitGroup
	var mux sync.Mutex

	for _, relay := range relays {
		if shouldTryToExportRelay(relay) {
			wg.Add(1)
			go func(r types.Relay) {
				// create relay logger
				defer wg.Done()
				singleRelayExport(r, &mux)
			}(relay)
		}
	}

	wg.Wait()
}

func singleRelayExport(r types.Relay, mux *sync.Mutex) {
	err := exportRelayBlocks(r)
	if err != nil {
		_ = handleRelayExportError(r, err, mux)
		log.Error(err, "error while updating relay export", 0, map[string]interface{}{"relay": r.ID})
		return
	}

	err = handleRelayExportSuccess(r, mux)
	if err != nil {
		log.Error(err, "error while updating relay export", 0, map[string]interface{}{"relay": r.ID})
		return
	}

	log.Infof("finished syncing payloads from relay %v", r.ID)
}

func handleRelayExportError(r types.Relay, err error, mux *sync.Mutex) error {
	errMsg := fmt.Errorf("failed to export blocks for relay %v: %w", r.ID, err)
	if shouldLogExportAsError(r) {
		log.Error(errMsg, "", 0, map[string]interface{}{"relay": r.ID})
	} else {
		log.WarnWithFields(log.Fields{"relay": r.ID}, errMsg.Error())
	}

	return updateRelayExportFailure(r, mux)
}

func handleRelayExportSuccess(r types.Relay, mux *sync.Mutex) error {
	mux.Lock()
	defer mux.Unlock()

	err := db.UpdateRelays(r.ID, r.Endpoint)
	if err != nil {
		log.Error(err, "could not update successful relay export", 0, map[string]interface{}{"relay": r.ID})
		return err
	}

	return nil
}

func updateRelayExportFailure(r types.Relay, mux *sync.Mutex) error {
	var err error
	mux.Lock()
	defer mux.Unlock()

	// only increase the export_failure_count if we haven't already reached the maximum wait time
	_, isMaxWaitTime := waitTimeToExportRelay(r)

	if isMaxWaitTime {
		err = db.UpdateRelayLastExportTry(r.ID, r.Endpoint)
	} else {
		err = db.UpdateRelayExportFailureCount(r.ExportFailureCount, r.ID, r.Endpoint)
	}

	if err != nil {
		log.Error(err, "could not update failed relay export", 0, map[string]interface{}{"relay": r.ID})
		return err
	}

	return nil
}

func exportRelayBlocks(r types.Relay) error {
	// retrieve the oldest tag usage so we know when to stop processing payloads from the head
	lastUsage, err := db.GetLastRelayBlock(r.ID)
	if err != nil {
		log.Error(err, "failed to retrieve last relay block from db, assuming none set", 0, map[string]interface{}{"relay": r.ID})
	}

	err = retrieveAndInsertPayloadsFromRelay(r, lastUsage.BlockSlot, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve and insert payloads: %w", err)
	}

	// to make sure we dont have an incomplete table, check if there are any payloads before our first tag usage
	firstUsage, err := db.GetFirstRelayBlock(r.ID)
	if err != nil {
		log.Error(err, "failed to retrieve first relay block from db, assuming none set", 0, map[string]interface{}{"relay": r.ID})
	}

	if firstUsage.BlockSlot != 0 {
		err = retrieveAndInsertPayloadsFromRelay(r, 0, firstUsage.BlockSlot)
		if err != nil {
			log.Error(err, "failed to retrieve and insert possibly missing payloads", 0, map[string]interface{}{"relay": r.ID})
			return err
		}
	}

	return nil
}

func retrieveAndInsertPayloadsFromRelay(r types.Relay, lowBound, highBound uint64) error {
	minSlot := calculateMinSlot(lowBound)
	offset := highBound

	for {
		payloads, err := fetchDeliveredPayloads(r, offset)
		if err != nil {
			return fmt.Errorf("error calling fetchDeliveredPayloads with offset: %v for relay: %v: %w", offset, r.ID, err)
		}

		if len(payloads) == 0 {
			log.Error(fmt.Errorf("got no payloads"), "", 0, map[string]interface{}{"relay": r.ID})
			break
		}

		for _, payload := range payloads {
			if err := insertPayloadIntoDB(r.ID, payload); err != nil {
				return fmt.Errorf("failed to insert payload into DB: %w", err)
			}
		}

		if payloads[len(payloads)-1].Slot < minSlot {
			// last payload we received is bellow than our calculated min_slot
			break
		}

		if len(payloads) < 100 {
			// if the response is less than 100 payloads, we assume that we have reached the end and break
			break
		}

		if payloads[len(payloads)-1].Slot == offset {
			return fmt.Errorf("relay doesn't follow spec, last returned slot matches offset (sort order ascending instead of descending)")
		}

		// sleep for a bit to not kill the relay
		offset = payloads[len(payloads)-1].Slot
		time.Sleep(time.Second)
	}

	return nil
}

func calculateMinSlot(lowBound uint64) uint64 {
	if lowBound > 10 {
		return lowBound - 10
	}
	return 0
}

func insertPayloadIntoDB(tagID string, payload exportertypes.BidTrace) error {
	// first insert the tag into the blocks_tags table
	err := db.SaveBlocksTags(tagID,
		payload.Slot,
		utils.MustParseHex(payload.BlockHash))
	if err != nil {
		return fmt.Errorf("failed to insert payload into blocks_tags table: %w", err)
	}

	err = db.SaveBlocksRelays(tagID,
		payload.Slot,
		payload.Value,
		utils.MustParseHex(payload.BlockHash),
		utils.MustParseHex(payload.BuilderPubkey),
		utils.MustParseHex(payload.ProposerPubkey),
		utils.MustParseHex(payload.ProposerFeeRecipient))
	if err != nil {
		return fmt.Errorf("failed to insert payload into relays_blocks table: %w", err)
	}

	return nil
}

func fetchDeliveredPayloads(r types.Relay, offset uint64) ([]exportertypes.BidTrace, error) {
	url := fmt.Sprintf("%s/relay/v1/data/bidtraces/proposer_payload_delivered?limit=100", r.Endpoint)
	if offset != 0 {
		url += fmt.Sprintf("&cursor=%v", offset)
	}

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch delivered payloads from relay %v: %w", r.ID, err)
	}
	defer resp.Body.Close()

	return decodePayloads(resp, r.ID)
}

func decodePayloads(resp *http.Response, tagID string) ([]exportertypes.BidTrace, error) {
	var payloads []exportertypes.BidTrace
	if err := json.NewDecoder(resp.Body).Decode(&payloads); err != nil {
		return nil, fmt.Errorf("failed to decode payloads from relay %v: %w", tagID, err)
	}

	return payloads, nil
}

func shouldTryToExportRelay(r types.Relay) bool {
	if r.ExportFailureCount == 0 {
		return true
	}

	waitTime, _ := waitTimeToExportRelay(r)
	return time.Since(r.LastExportTryTs) >= waitTime
}

func waitTimeToExportRelay(r types.Relay) (waitTime time.Duration, isMaxWaitTime bool) {
	maxWaitTimeForRelayExport := utils.Day
	waitTime = time.Duration(math.Exp2(float64(r.ExportFailureCount))) * time.Minute
	if waitTime >= maxWaitTimeForRelayExport {
		waitTime = maxWaitTimeForRelayExport
		isMaxWaitTime = true
	}
	return waitTime, isMaxWaitTime
}

func shouldLogExportAsError(r types.Relay) bool {
	maxWaitTimeForRelayExportError := utils.Month
	return time.Since(r.LastExportSuccessTs) >= maxWaitTimeForRelayExportError
}
