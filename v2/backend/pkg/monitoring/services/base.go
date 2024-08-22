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

func ReportStatus(ctx context.Context, id string, err error, expires_in *time.Duration, metadata map[string]string) {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	// if "status" is not set set it to "failure" if err is not nil or "heartbeat" if err is nil
	if _, ok := metadata["status"]; !ok {
		if err != nil {
			metadata["status"] = "failure"
		} else {
			metadata["status"] = "heartbeat"
		}
	}
	metadata["executable_version"] = fmt.Sprintf("%s (%s)", version.Version, version.GoVersion)

	if err != nil {
		metadata["error"] = err.Error()
	}

	// report status to monitoring
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	//log.Infof("new status report at %v", time.Now())
	if db.ClickHouseNativeWriter == nil {
		log.Error(nil, "clickhouse native writer is nil", 0)
		return
	}
	expires_at := time.Now().Add(1 * time.Minute)
	if expires_in != nil {
		expires_at = time.Now().Add(*expires_in)
	}
	err = db.ClickHouseNativeWriter.AsyncInsert(
		ctx,
		"INSERT INTO status_reports (emitter, event_id, inserted_at, expires_at, metadata) VALUES (?, ?, ?, ?, ?)",
		false,
		utils.GetUUID(),
		id,
		time.Now().UnixMilli(),
		expires_at,
		metadata,
	)
	if err != nil {
		log.Error(err, "error inserting status report", 0)
	}
}
