package services

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type CleanShutdownSpamDetector struct {
	ServiceBase
}

func (s *CleanShutdownSpamDetector) Start() {
	if !s.running.CompareAndSwap(false, true) {
		// already running, return error
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

func (s *CleanShutdownSpamDetector) internalProcess() {
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

func (s *CleanShutdownSpamDetector) runChecks() {
	statusReporter := NewStatusReporter(constants.Event_MonitoringCleanShutdownSpam, constants.Default, 30*time.Second)
	statusReporter.Report(constants.Running, nil)
	if db.ClickHouseReader == nil {
		statusReporter.Report(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Tracef("checking clean shutdown spam")

	emitters, err := s.db.GetEmitters()
	if err != nil {
		statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
		return
	}

	threshold := 10
	md := map[string]string{
		"count":     strconv.Itoa(len(emitters)),
		"threshold": strconv.Itoa(threshold),
	}
	if len(emitters) > threshold {
		payload, err := json.Marshal(emitters)
		if err != nil {
			statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
			return
		}
		md["emitters"] = string(payload)
		statusReporter.Report(constants.Failure, md)
		return
	}
	statusReporter.Report(constants.Success, md)
}
