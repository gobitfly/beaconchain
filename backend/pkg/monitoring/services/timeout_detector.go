package services

import (
	"encoding/json"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type ServiceTimeoutDetector struct {
	ServiceBase
}

func (s *ServiceTimeoutDetector) Start() {
	if !s.running.CompareAndSwap(false, true) {
		// already running, return error
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

func (s *ServiceTimeoutDetector) internalProcess() {
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

func (s *ServiceTimeoutDetector) runChecks() {
	statusReporter := NewStatusReporter(constants.Event_MonitoringTimeouts, constants.Default, 30*time.Second)
	statusReporter.Report(constants.Running, nil)
	if db.ClickHouseReader == nil {
		statusReporter.Report(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Tracef("checking services timeouts")

	victims, err := s.db.GetLatestStatusReport()
	if err != nil {
		statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
		return
	}
	if len(victims) == 0 {
		statusReporter.Report(constants.Success, nil)
		return
	}
	payload, err := json.Marshal(victims)
	if err != nil {
		statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
		return
	}

	md := map[string]string{"failing_reports": string(payload), "error": "reports are running for too long"}
	statusReporter.Report(constants.Failure, md)
}
