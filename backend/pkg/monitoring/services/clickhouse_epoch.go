package services

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type ServiceClickhouseEpoch struct {
	ServiceBase
}

func (s *ServiceClickhouseEpoch) Start() {
	if !s.running.CompareAndSwap(false, true) {
		// already running, return error
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

func (s *ServiceClickhouseEpoch) internalProcess() {
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

func (s *ServiceClickhouseEpoch) runChecks() {
	statusReport := StatusReporter().NewStatusReport(constants.Event_ClickhouseDashboardEpoch, constants.Default, 30*time.Second)
	statusReport(constants.Running, nil)
	if db.ClickHouseReader == nil {
		statusReport(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Tracef("checking clickhouse epoch")

	ts, err := s.db.GetVDLatestEpochTs()
	if err != nil {
		statusReport(constants.Failure, map[string]string{"error": err.Error()})
		return
	}
	// check if delta is out of bounds
	threshold := 1 * time.Hour
	md := map[string]string{"delta": time.Since(ts).String(), "threshold": threshold.String()}
	if time.Since(ts) > threshold {
		md["error"] = "delta is over threshold"
		statusReport(constants.Failure, md)
		return
	}
	statusReport(constants.Success, md)
}
