package services

import (
	"context"
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
	id := "ch_dashboard_epoch"
	r := NewStatusReport(id, constants.Default, 30*time.Second)
	r(constants.Running, nil)
	if db.ClickHouseReader == nil {
		r(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Tracef("checking clickhouse epoch")
	// context with deadline
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	var t time.Time
	err := db.ClickHouseReader.GetContext(ctx, &t, "SELECT MAX(epoch_timestamp) FROM validator_dashboard_data_epoch")
	if err != nil {
		r(constants.Failure, map[string]string{"error": err.Error()})
		return
	}
	// check if delta is out of bounds
	threshold := 1 * time.Hour
	md := map[string]string{"delta": time.Since(t).String(), "threshold": threshold.String()}
	if time.Since(t) > threshold {
		md["error"] = "delta is over threshold"
		r(constants.Failure, md)
		return
	}
	r(constants.Success, md)
}
