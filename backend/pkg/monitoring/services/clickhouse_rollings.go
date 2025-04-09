package services

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

// create db connection service that checks for the status of db connections

type ServiceClickhouseRollings struct {
	ServiceBase
}

func (s *ServiceClickhouseRollings) Start() {
	if !s.running.CompareAndSwap(false, true) {
		// already running, return error
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

func (s *ServiceClickhouseRollings) internalProcess() {
	defer s.wg.Done()
	s.runChecks()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			s.runChecks()
		}
	}
}

func (s *ServiceClickhouseRollings) runChecks() {
	rollings := map[string]constants.Event{
		"1h":    constants.Event_ClickhouseRolling_1h,
		"24h":   constants.Event_ClickhouseRolling_24h,
		"7d":    constants.Event_ClickhouseRolling_7d,
		"30d":   constants.Event_ClickhouseRolling_30d,
		"90d":   constants.Event_ClickhouseRolling_90d,
		"total": constants.Event_ClickhouseRolling_total,
	}
	wg := sync.WaitGroup{}
	for rolling := range maps.Keys(rollings) {
		rolling := rolling
		wg.Add(1)
		go func() {
			defer wg.Done()
			statusReporter := NewStatusReporter(rollings[rolling], constants.Default, 30*time.Second)
			statusReporter.Report(constants.Running, nil)
			if db.ClickHouseReader == nil {
				statusReporter.Report(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
				// ignore
				return
			}
			log.Tracef("checking clickhouse rolling %s", rolling)
			// context with deadline
			tsEpochTable, err := s.db.GetLatestEpoch()
			if err != nil {
				statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
				return
			}
			epochRollingTable, err := s.db.GetEpochEnd(rolling)
			if err != nil {
				statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
				return
			}
			// convert to timestamp
			tsRollingTable := utils.EpochToTime(epochRollingTable)
			threshold := 30 * time.Minute
			delta := tsEpochTable.Sub(tsRollingTable)
			// check if delta is out of bounds
			md := map[string]string{"delta": delta.String(), "threshold": threshold.String()}
			if delta > threshold {
				md["error"] = fmt.Sprintf("delta is over threshold %d", threshold)
				statusReporter.Report(constants.Failure, md)
				return
			}
			statusReporter.Report(constants.Success, md)
		}()
	}
	wg.Wait()
}
