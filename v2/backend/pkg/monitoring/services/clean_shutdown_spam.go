package services

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
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
	id := "monitoring_clean_shutdown_spam"
	r := NewStatusReport(id, constants.Default, 30*time.Second)
	r(constants.Running, nil)
	if db.ClickHouseReader == nil {
		r(constants.Failure, map[string]string{"error": "clickhouse reader is nil"})
		// ignore
		return
	}
	log.Tracef("checking clean shutdown spam")

	query := `
		SELECT
			emitter
		FROM
			status_reports
		WHERE
			deployment_type = ?
			AND inserted_at >= now() - interval 5 minutes
			AND event_id = ?
			`
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	var emitters []string
	err := db.ClickHouseReader.SelectContext(ctx, &emitters, query, utils.Config.DeploymentType, constants.CleanShutdownEvent)
	if err != nil {
		r(constants.Failure, map[string]string{"error": err.Error()})
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
			r(constants.Failure, map[string]string{"error": err.Error()})
			return
		}
		md["emitters"] = string(payload)
		r(constants.Failure, md)
		return
	}
	r(constants.Success, md)
}
