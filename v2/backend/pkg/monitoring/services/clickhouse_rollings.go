package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
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
	rollings := []string{
		"1h",
		"24h",
		"7d",
		"30d",
		"90d",
		"total",
	}
	wg := sync.WaitGroup{}
	expiry := 5 * time.Minute
	for _, rolling := range rollings {
		rolling := rolling
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("ch_rolling_%s", rolling)
			if db.ClickHouseReader == nil {
				ReportStatus(s.ctx, id, fmt.Errorf("clickhouse reader is nil"), &expiry, nil)
				// ignore
				return
			}
			log.Debugf("checking clickhouse rolling %s", rolling)
			// context with deadline
			ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
			defer cancel()
			var delta uint64
			err := db.ClickHouseReader.GetContext(ctx, &delta, fmt.Sprintf(`
					SELECT
						coalesce((
							SELECT
								max(epoch)
							FROM holesky.validator_dashboard_data_epoch
							WHERE
								epoch_timestamp = (
									SELECT
										max(epoch_timestamp)
									FROM holesky.validator_dashboard_data_epoch)) - MAX(epoch_end), 255) AS delta
					FROM
						holesky.validator_dashboard_data_rolling_%s
					WHERE
						validator_index = 0`, rolling))
			if err != nil {
				ReportStatus(s.ctx, id, err, &expiry, nil)
				return
			}
			// check if delta is out of bounds
			threshold := 2
			md := map[string]string{"delta": fmt.Sprintf("%d", delta), "threshold": fmt.Sprintf("%d", threshold)}
			if delta > uint64(threshold) {
				ReportStatus(s.ctx, id, fmt.Errorf("delta is over threshold %d", threshold), &expiry, md)
				return
			}
			ReportStatus(s.ctx, id, nil, &expiry, md)
		}()
	}
	wg.Wait()
}
