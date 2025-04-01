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

func NewStatusReport(id constants.Event, deploymentType string, timeout time.Duration, check_interval time.Duration) func(status constants.StatusType, metadata map[string]string) {
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
				"deployment_type": deploymentType,
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
					deploymentType,
					flake,
					expires_at,
					timeouts_at,
					metadata,
				)
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
