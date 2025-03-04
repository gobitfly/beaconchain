package modules

import (
	"context"
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
)

type relaysExporter struct {
	db          db.ConsensusRepository
	relayClient relayClient
	delay       time.Duration
	ctx         context.Context
}

func newRelaysExporter(ctx context.Context, db db.ConsensusRepository) relaysExporter {
	return relaysExporter{
		db:          db,
		relayClient: nodeClient{},
		delay:       time.Minute,
		ctx:         ctx,
	}
}

func (rs *relaysExporter) MEVBoostRelaysExporter() {
	for {
		select {
		case <-rs.ctx.Done():
			log.Info("MEV boost relays export loop cancelled")
			return
		default:
			// we retrieve the relays from the db each loop to prevent having to restart the exporter for changes
			relays, err := rs.db.GetRelays()
			if err != nil {
				log.Error(err, "failed to retrieve relays from db", 0)
				time.Sleep(rs.delay)
				continue
			}
			var wg sync.WaitGroup
			for _, relay := range relays {
				if !shouldTryToExportRelay(relay) {
					continue
				}
				wg.Add(1)
				go func(r types.Relay) {
					defer wg.Done()
					rs.singleRelayExport(r)
				}(relay)
			}
			wg.Wait()
			time.Sleep(rs.delay)
		}
	}
}

func (rs *relaysExporter) singleRelayExport(r types.Relay) {
	err := rs.exportRelayBlocks(r)
	if err != nil {
		errMsg := fmt.Errorf("failed to export blocks for relay: %v", err)
		if shouldLogExportAsError(r) {
			log.Error(err, "", 0, map[string]interface{}{"relay": r.ID})
		} else {
			log.WarnWithFields(log.Fields{"relay": r.ID}, errMsg.Error())
		}

		// only increase the export_failure_count if we haven't already reached the maximum wait time
		_, isMaxWaitTime := waitTimeToExportRelay(r)

		if isMaxWaitTime {
			err = rs.db.UpdateRelayLastExportTry(r.ID, r.Endpoint)
		} else {
			err = rs.db.UpdateRelayExportFailureCount(r.ExportFailureCount, r.ID, r.Endpoint)
		}

		if err != nil {
			log.Error(err, "could not update failed relay export", 0, map[string]interface{}{"relay": r.ID})
		}

		return
	}

	err = rs.db.UpdateRelay(r.ID, r.Endpoint)
	if err != nil {
		log.Error(err, "could not update successful relay export", 0, map[string]interface{}{"relay": r.ID})
	}

	log.Infof("finished syncing payloads from relay %v", r.ID)
}

func (rs *relaysExporter) exportRelayBlocks(r types.Relay) error {
	// retrieve the oldest tag usage so we know when to stop processing payloads from the head
	lastUsage, err := rs.db.GetLastRelayBlock(r.ID)
	if err != nil {
		log.Error(err, "failed to retrieve last relay block from db, assuming none set", 0, map[string]interface{}{"relay": r.ID})
	}

	err = rs.retrieveAndInsertPayloadsFromRelay(r, lastUsage.BlockSlot, 0)
	if err != nil {
		return err
	}

	// to make sure we dont have an incomplete table, check if there are any payloads before our first tag usage
	firstUsage, err := rs.db.GetFirstRelayBlock(r.ID)
	if err != nil {
		log.Error(err, "failed to retrieve first relay block from db, assuming none set", 0, map[string]interface{}{"relay": r.ID})
	}

	if firstUsage.BlockSlot == 0 {
		return nil
	}

	err = rs.retrieveAndInsertPayloadsFromRelay(r, 0, firstUsage.BlockSlot)
	if err != nil {
		log.Error(err, "failed to retrieve and insert possibly missing payloads", 0, map[string]interface{}{"relay": r.ID})
		return err
	}

	return nil
}

func (rs *relaysExporter) retrieveAndInsertPayloadsFromRelay(r types.Relay, lowBound, highBound uint64) error {
	minSlot := calculateMinSlot(lowBound)
	offset := highBound

	for {
		select {
		case <-rs.ctx.Done():
			log.Info("fetch delivered payloads loop cancelled")
			return nil
		default:
			payloads, err := rs.relayClient.fetchDeliveredPayloads(r.Endpoint, r.ID, offset)
			if err != nil {
				return fmt.Errorf("error calling fetchDeliveredPayloads with offset: %v for relay: %v: %w", offset, r.ID, err)
			}

			if len(payloads) == 0 {
				log.Error(fmt.Errorf("got no payloads"), "", 0, map[string]interface{}{"relay": r.ID})
				break
			}

			for _, payload := range payloads {

				err := rs.db.SaveBlockTagsAndRelays(r.ID, payload)
				if err != nil {
					return err
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
	}
}

func calculateMinSlot(lowBound uint64) uint64 {
	if lowBound > 10 {
		return lowBound - 10
	}
	return 0
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

type relayClient interface {
	fetchDeliveredPayloads(endpoint string, id string, offset uint64) ([]types.BidTrace, error)
}

type nodeClient struct{}

func (nodeClient) fetchDeliveredPayloads(endpoint string, id string, offset uint64) ([]types.BidTrace, error) {
	url := fmt.Sprintf("%s/relay/v1/data/bidtraces/proposer_payload_delivered?limit=100", endpoint)
	if offset != 0 {
		url += fmt.Sprintf("&cursor=%v", offset)
	}

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Error(err, "error retrieving delivered payloads", 0, map[string]interface{}{"relay": id, "offset": offset, "url": url})
		return nil, fmt.Errorf("error retrieving delivered payloads for relay: %v, offset: %v, url: %v: %w", id, offset, url, err)
	}

	defer resp.Body.Close()

	var payloads []types.BidTrace
	err = json.NewDecoder(resp.Body).Decode(&payloads)
	if err != nil {
		return nil, fmt.Errorf("error decoding json for delivered payloads for relay: %v, offset: %v, url: %v: %w", id, offset, url, err)
	}

	return payloads, nil
}
