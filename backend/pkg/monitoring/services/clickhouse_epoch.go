package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
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
	if db.ClickHouseReader == nil {
		ReportStatus(s.ctx, id, fmt.Errorf("clickhouse reader is nil"), nil, nil)
		// ignore
		return
	}
	log.Debugf("checking clickhouse epoch")
	// context with deadline
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	var t time.Time
	expiry := 5 * time.Minute
	err := db.ClickHouseReader.GetContext(ctx, &t, "SELECT MAX(epoch_timestamp) FROM validator_dashboard_data_epoch")
	if err != nil {
		ReportStatus(s.ctx, id, err, &expiry, nil)
		return
	}
	// check if delta is out of bounds
	threshold := 1 * time.Hour
	md := map[string]string{"delta": time.Since(t).String(), "threshold": threshold.String()}
	if time.Since(t) > threshold {
		ReportStatus(s.ctx, id, fmt.Errorf("delta is over threshold %d", threshold), &expiry, md)
		return
	}
	ReportStatus(s.ctx, id, nil, &expiry, md)
}
