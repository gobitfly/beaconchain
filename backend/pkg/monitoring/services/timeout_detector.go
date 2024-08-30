package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
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
	id := "monitoring_timeouts"
	r := NewStatusReport(id, constants.Default, 30*time.Second)
	r(constants.Running, nil)
	if db.ClickHouseReader == nil {
		r(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Debugf("checking services timeouts")

	query := `
		with active_reports as (
			SELECT
				event_id,
				emitter,
				run_id,
				inserted_at,
				insert_id,
				expires_at,
				timeouts_at,
				status,
				metadata
			FROM status_reports
			WHERE expires_at > now() and deployment_type = ?
			ORDER BY
				event_id ASC,
				emitter ASC,
				run_id ASC,
				insert_id DESC
		), latest_report_per_run as (
			SELECT
				event_id,
				emitter,
				any(inserted_at) as inserted_at, 
				any(insert_id) as insert_id, 
				any(expires_at) as expires_at,
				any(timeouts_at) as timeouts_at,
				any(status) AS status,
				any(metadata) AS metadata
			FROM
				active_reports
			GROUP BY
				event_id,
				emitter,
				run_id
			order by insert_id desc
		)
		SELECT
			event_id,
			status,
			inserted_at,
			expires_at,
			timeouts_at,
			metadata
		FROM
			latest_report_per_run
		where status = 'running' and timeouts_at < now()
		ORDER BY event_id ASC, inserted_at DESC`
	// context with deadline
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	var victims []struct {
		EventID    string            `db:"event_id"`
		Status     string            `db:"status"`
		InsertedAt time.Time         `db:"inserted_at"`
		ExpiresAt  time.Time         `db:"expires_at"`
		TimeoutsAt time.Time         `db:"timeouts_at"`
		Metadata   map[string]string `db:"metadata"`
	}
	err := db.ClickHouseReader.SelectContext(ctx, &victims, query, utils.Config.DeploymentType)
	if err != nil {
		r(constants.Failure, map[string]string{"error": err.Error()})
		return
	}
	if len(victims) == 0 {
		r(constants.Success, nil)
		return
	}
	payload, err := json.Marshal(victims)
	if err != nil {
		r(constants.Failure, map[string]string{"error": err.Error()})
		return
	}
	md := map[string]string{"failing_reports": string(payload), "error": "reports are running for too long"}
	r(constants.Failure, md)
}
