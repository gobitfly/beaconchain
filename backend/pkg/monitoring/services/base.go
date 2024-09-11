package services

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

func NewStatusReport(id string, timeout time.Duration, check_interval time.Duration) func(status constants.StatusType, metadata map[string]string) {
	runId := uuid.New().String()
	return func(status constants.StatusType, metadata map[string]string) {
		// acquire snowflake synchronously
		flake := utils.GetSnowflake()
		now := time.Now()
		go func() {
			if metadata == nil {
				metadata = make(map[string]string)
			}

			metadata["run_id"] = runId
			metadata["status"] = string(status)
			metadata["executable_version"] = fmt.Sprintf("%s (%s)", version.Version, version.GoVersion)

			// report status to monitoring
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			timeouts_at := now.Add(1 * time.Minute)
			if timeout != constants.Default {
				timeouts_at = now.Add(timeout)
			}
			expires_at := timeouts_at.Add(5 * time.Minute)
			if check_interval >= 5*time.Minute {
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
					true,
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
