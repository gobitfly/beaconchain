package monitoring

import (
	"sync"
	"sync/atomic"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

var monitoredServices []services.Service
var startedClickhouse atomic.Bool
var initMutex = sync.Mutex{}

func Init(full bool) {
	initMutex.Lock()
	defer initMutex.Unlock()
	metrics.UUID.WithLabelValues(utils.GetUUID()).Set(1) // so we can find out where the uuid is set
	metrics.DeploymentType.WithLabelValues(utils.Config.DeploymentType).Set(1)
	if db.ClickHouseNativeWriter == nil {
		log.Infof("initializing clickhouse writer")
		startedClickhouse.Store(true)
		db.ClickHouseNativeWriter = db.MustInitClickhouseNative(&types.DatabaseConfig{
			Username:     utils.Config.ClickHouse.WriterDatabase.Username,
			Password:     utils.Config.ClickHouse.WriterDatabase.Password,
			Name:         utils.Config.ClickHouse.WriterDatabase.Name,
			Host:         utils.Config.ClickHouse.WriterDatabase.Host,
			Port:         utils.Config.ClickHouse.WriterDatabase.Port,
			MaxOpenConns: utils.Config.ClickHouse.WriterDatabase.MaxOpenConns,
			SSL:          true,
			MaxIdleConns: utils.Config.ClickHouse.WriterDatabase.MaxIdleConns,
		})
	}
	monitoredServices = []services.Service{
		&services.ServerDbConnections{},
	}
	if full {
		monitoredServices = append(monitoredServices,
			&services.ServiceClickhouseRollings{},
			&services.ServiceClickhouseEpoch{},
			&services.ServiceTimeoutDetector{},
			&services.CleanShutdownSpamDetector{},
		)
	}

	for _, service := range monitoredServices {
		service.InitServices()
	}
}

func Start() {
	log.Infof("starting monitoring services")
	for _, service := range monitoredServices {
		service.Start()
	}
}

func Stop() {
	log.Infof("stopping monitoring services")
	for _, service := range monitoredServices {
		service.Stop()
	}
	// this prevents status reports that werent shut down cleanly from triggering alerts
	services.NewStatusReport(constants.CleanShutdownEvent, constants.Default, constants.Default)(constants.Success, nil)
	if startedClickhouse.Load() {
		db.ClickHouseNativeWriter.Close()
	}
}
