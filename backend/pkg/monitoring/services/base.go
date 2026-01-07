package services

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
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
}

func (s *ServiceBase) InitServices() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *ServiceBase) Stop() {
	if !s.running.CompareAndSwap(true, false) {
		return
	}
	s.cancel()
	s.wg.Wait()
}

func newStatusReport(id constants.Event, timeout time.Duration, check_interval time.Duration) func(status constants.StatusType, metadata map[string]string) {
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

			// report status to monitoring
			timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// wrap in clickhouse context so we can set the setting throw_if_deduplication_in_dependent_materialized_views_enabled_with_async_insert to 0
			// we have no materialized views on the status_reports table, but it triggers as we need the deduplication setting for the other tables
			ctx := clickhouse.Context(timeoutContext, clickhouse.WithSettings(
				clickhouse.Settings{
					"throw_if_deduplication_in_dependent_materialized_views_enabled_with_async_insert": 0,
					"deduplicate_blocks_in_dependent_materialized_views":                               0,
				},
			))

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
				"deployment_type": utils.Config.DeploymentType,
				"insert_id":       flake,
				"expires_at":      expires_at,
				"timeouts_at":     timeouts_at,
				"metadata":        metadata,
			}, "sending status report")
			var err error
			if db.ClickHouseNativeWriter != nil {
				err = db.ClickHouseNativeWriter.AsyncInsert(
					ctx,
					"INSERT INTO status_reports (emitter, event_id, deployment_type, insert_id, expires_at, timeouts_at, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)",
					false, // true means wait for settlement, but we want to shoot and forget. false does mean we cant log any errors that occur during settlement
					utils.GetUUID(),
					id,
					utils.Config.DeploymentType,
					flake,
					expires_at,
					timeouts_at,
					metadata,
				)
			} else if utils.Config.DeploymentType != "development" {
				log.Error(nil, "clickhouse native writer is nil", 0)
			}
			if err != nil && utils.Config.DeploymentType != "development" {
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

type statusReport interface {
	NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) func(status constants.StatusType, metadata map[string]string)
}

type stubStatusReporter struct{}

func (sr stubStatusReporter) NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) func(status constants.StatusType, metadata map[string]string) {
	return func(status constants.StatusType, metadata map[string]string) {
		// no-op implementation
		// only warn if utils.Config is initialized and we're not in development environment
		if utils.Config != nil && utils.Config.DeploymentType != "development" {
			log.Warnf("STUB STATUS REPORTER IN USE IN %s ENVIRONMENT! Event: %s, Status: %s, Metadata: %v",
				utils.Config.DeploymentType,
				id,
				status,
				metadata,
			)
		}
	}
}

type statusReporter struct{}

func (sr statusReporter) NewStatusReport(id constants.Event, timeout time.Duration, checkInterval time.Duration) func(status constants.StatusType, metadata map[string]string) {
	return newStatusReport(id, timeout, checkInterval)
}

var StatusReporter statusReport = stubStatusReporter{}

func InitStatusReport() {
	StatusReporter = statusReporter{}
}
