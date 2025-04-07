package services

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/google/uuid"
)

// go interface for basic service

type Service interface {
	InitServices()
	Start()
	Stop()
}

type ServiceBase struct {
	ctx     context.Context
	cancel  context.CancelFunc
	running atomic.Bool
	wg      sync.WaitGroup
	db      db2.Monitoring
}

func (s *ServiceBase) InitServices() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.db = db2.NewMonitoringClickhouse(db.ClickHouseReader, db.ClickHouseNativeWriter)
}

func (s *ServiceBase) Stop() {
	if !s.running.CompareAndSwap(true, false) {
		return
	}
	s.cancel()
	s.wg.Wait()
}

func newStatusReport(id constants.Event, timeout time.Duration, check_interval time.Duration, deploymentType string) func(status constants.StatusType, metadata map[string]string) {
	runId := uuid.New().String()
	return func(status constants.StatusType, metadata map[string]string) {
		// acquire snowflake synchronously
		flake := utils.GetSnowflake()
		callerProgramCounter, callerFullFilePath, callerLine, callerOK := runtime.Caller(1)
		now := time.Now()
		go func() {
			if metadata == nil {
				metadata = make(map[string]string)
			}

			metadata["run_id"] = runId
			metadata["status"] = string(status)
			metadata["executable_version"] = fmt.Sprintf("%s (%s)", version.Version, version.GoVersion)
			if callerOK {
				callerFunction := runtime.FuncForPC(callerProgramCounter).Name()
				callerFile := filepath.Base(callerFullFilePath)
				metadata["caller"] = fmt.Sprintf("%s %s:%d", callerFunction, callerFile, callerLine)
			}

			timeouts_at := now.Add(1 * time.Minute)
			if timeout != constants.Default {
				timeouts_at = now.Add(timeout)
			}
			expires_at := timeouts_at.Add(1 * time.Minute)
			if check_interval >= 1*time.Minute {
				expires_at = timeouts_at.Add(check_interval)
			}
			log.TraceWithFields(log.Fields{
				"emitter":         id,
				"event_id":        utils.GetUUID(),
				"deployment_type": deploymentType,
				"insert_id":       flake,
				"expires_at":      expires_at,
				"timeouts_at":     timeouts_at,
				"metadata":        metadata,
			}, "sending status report")
			var err error
			if db.ClickHouseNativeWriter != nil {
				monitoringCH := db2.NewMonitoringClickhouse(nil, db.ClickHouseNativeWriter)
				err = monitoringCH.SaveNewStatusReport(db2.StatusReport{
					ID:         id,
					Flake:      flake,
					ExpiresAt:  expires_at,
					TimeoutsAt: timeouts_at,
					Metadata:   metadata,
				})
			} else if deploymentType != "development" {
				log.Error(nil, "clickhouse native writer is nil", 0)
			}
			if err != nil && deploymentType != "development" {
				log.Error(err, "error inserting status report", 0)
			}
		}()
	}
}

func GetRequiredEvents() []constants.Event {
	// i would hope this simple of a function doesnt need caching
	requiredEvents := constants.RequiredEvents
	if utils.Config.DeploymentType != "production" {
		return requiredEvents
	}
	requiredEvents = append(requiredEvents, constants.ProductionRequiredEvents...)
	if utils.Config.RocketpoolExporter.Enabled {
		requiredEvents = append(requiredEvents, constants.Event_ExporterLegacyRocketPool)
	}
	if utils.Config.Indexer.PubKeyTagsExporter.Enabled {
		requiredEvents = append(requiredEvents, constants.Event_ExporterLegacyPubkeyTags)
	}
	return requiredEvents
}

type StatusReporterFunc func(status constants.StatusType, metadata map[string]string)

type StatusReport interface {
	NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) StatusReporterFunc
}

type stubStatusReporter struct {
	initialized    bool
	deploymentType string
}

func (sr stubStatusReporter) NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) StatusReporterFunc {
	return func(status constants.StatusType, metadata map[string]string) {
		// no-op implementation
		// only warn if we're not in development environment
		if sr.initialized && sr.deploymentType != "development" {
			log.Warnf("STUB STATUS REPORTER IN USE IN %s ENVIRONMENT! Event: %s, Status: %s, Metadata: %v",
				sr.deploymentType,
				id,
				status,
				metadata,
			)
		}
	}
}

type statusReporter struct {
	initialized    bool
	deploymentType string
}

func (sr statusReporter) NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) StatusReporterFunc {
	if !sr.initialized {
		log.Warn("status reporter not initialized, using stub implementation")
		return stubStatusReporter{initialized: false}.NewStatusReport(id, timeout, checkInterval)
	}
	return newStatusReport(id, timeout, checkInterval, sr.deploymentType)
}

// default to stub implementation
var globalStatusReporter StatusReport = stubStatusReporter{initialized: false}

func InitStatusReporter(deploymentType string) {
	globalStatusReporter = statusReporter{
		initialized:    true,
		deploymentType: deploymentType,
	}
}

func StatusReporter() StatusReport {
	return globalStatusReporter
}
