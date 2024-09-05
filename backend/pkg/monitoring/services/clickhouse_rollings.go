package services

import (
	"context"
	"fmt"
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
	rollings := []string{
		"1h",
		"24h",
		"7d",
		"30d",
		"90d",
		"total",
	}
	wg := sync.WaitGroup{}
	for _, rolling := range rollings {
		rolling := rolling
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("ch_rolling_%s", rolling)
			r := NewStatusReport(id, constants.Default, 30*time.Second)
			r(constants.Running, nil)
			if db.ClickHouseReader == nil {
				r(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
				// ignore
				return
			}
			log.Tracef("checking clickhouse rolling %s", rolling)
			// context with deadline
			ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
			defer cancel()
			var tsEpochTable time.Time
			err := db.ClickHouseReader.GetContext(ctx, &tsEpochTable, `
					SELECT
						max(epoch_timestamp)
					FROM holesky.validator_dashboard_data_epoch`,
			)
			if err != nil {
				r(constants.Failure, map[string]string{"error": err.Error()})
				return
			}
			var epochRollingTable uint64
			err = db.ClickHouseReader.GetContext(ctx, &epochRollingTable, fmt.Sprintf(`
					SELECT
						max(epoch_end)
					FROM holesky.validator_dashboard_data_rolling_%s`,
				rolling,
			),
			)
			if err != nil {
				r(constants.Failure, map[string]string{"error": err.Error()})
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
				r(constants.Failure, md)
				return
			}
			r(constants.Success, md)
		}()
	}
	wg.Wait()
}
