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

type statusConfig struct {
	initialized    bool
	deploymentType string
}

var (
	config     statusConfig
	configOnce sync.Once
)

func InitStatusReporter(deploymentType string) {
	if deploymentType == "" {
		log.Warn("deployment type is empty, defaulting to 'development'")
		deploymentType = "development"
	}

	configOnce.Do(func() {
		config = statusConfig{
			initialized:    true,
			deploymentType: deploymentType,
		}
	})
}

func NewStatusReporter(event constants.Event, defaultTimeout, defaultInterval time.Duration) StatusReporter {
	if !config.initialized {
		return &stubStatusReporter{}
	}

	return &statusReporter{
		id:            event,
		timeout:       defaultTimeout,
		checkInterval: defaultInterval,
		runID:         uuid.New().String(),
	}
}

type StatusReporter interface {
	Report(status constants.StatusType, metadata map[string]string)
}

type statusReporter struct {
	id            constants.Event
	timeout       time.Duration
	checkInterval time.Duration
	runID         string
}

func (r *statusReporter) Report(status constants.StatusType, metadata map[string]string) {
	go r.report(status, metadata)
}

func (r *statusReporter) report(status constants.StatusType, metadata map[string]string) {
	flake := utils.GetSnowflake()
	callerProgramCounter, callerFilePath, callerLine, callerOK := runtime.Caller(2)
	now := time.Now()

	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata["run_id"] = r.runID
	metadata["status"] = string(status)
	metadata["executable_version"] = fmt.Sprintf("%s (%s)", version.Version, version.GoVersion)

	if callerOK {
		callerFunction := runtime.FuncForPC(callerProgramCounter).Name()
		callerFile := filepath.Base(callerFilePath)
		metadata["caller"] = fmt.Sprintf("%s %s:%d", callerFunction, callerFile, callerLine)
	}

	timeoutsAt := now.Add(1 * time.Minute)
	if r.timeout != constants.Default {
		timeoutsAt = now.Add(r.timeout)
	}
	expiresAt := timeoutsAt.Add(1 * time.Minute)
	if r.checkInterval >= 1*time.Minute {
		expiresAt = timeoutsAt.Add(r.checkInterval)
	}
	log.TraceWithFields(log.Fields{
		"emitter":         r.id,
		"event_id":        utils.GetUUID(),
		"deployment_type": config.deploymentType,
		"insert_id":       flake,
		"expires_at":      expiresAt,
		"timeouts_at":     timeoutsAt,
		"metadata":        metadata,
	}, "sending status report")

	if db.ClickHouseNativeWriter != nil {
		monitoringCH := db2.NewMonitoringClickhouse(nil, db.ClickHouseNativeWriter)
		err := monitoringCH.SaveNewStatusReport(db2.StatusReport{
			ID:         r.id,
			Flake:      flake,
			ExpiresAt:  expiresAt,
			TimeoutsAt: timeoutsAt,
			Metadata:   metadata,
		})
		if err != nil && config.deploymentType != "development" {
			log.Error(err, "error inserting status report", 0)
		}
	} else if config.deploymentType != "development" {
		log.Error(nil, "clickhouse native writer is nil", 0)
	}
}

type stubStatusReporter struct{}

func (r *stubStatusReporter) Report(status constants.StatusType, metadata map[string]string) {
	if config.deploymentType != "development" && config.deploymentType != "" {
		log.Warnf("WARN: STUB STATUS REPORTER IN USE IN %s ENVIRONMENT! Status: %s, Metadata: %v",
			config.deploymentType,
			status,
			metadata,
		)
	}
}

func GetRequiredEvents(deploymentType string, rocketpoolEnabled, pubkeyTagsEnabled bool) []constants.Event {
	// create copy of slice to avoid modifying the original event slice
	requiredEvents := make([]constants.Event, len(constants.RequiredEvents))
	copy(requiredEvents, constants.RequiredEvents)

	if deploymentType != "production" {
		return requiredEvents
	}
	requiredEvents = append(requiredEvents, constants.ProductionRequiredEvents...)

	if rocketpoolEnabled {
		requiredEvents = append(requiredEvents, constants.Event_ExporterLegacyRocketPool)
	}
	if pubkeyTagsEnabled {
		requiredEvents = append(requiredEvents, constants.Event_ExporterLegacyPubkeyTags)
	}

	return requiredEvents
}
