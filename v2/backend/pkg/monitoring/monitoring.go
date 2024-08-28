package monitoring

import (
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

var monitoredServices []services.Service

func Init(full bool) {
	metrics.UUID.WithLabelValues(utils.GetUUID()).Set(1) // so we can find out where the uuid is set
	metrics.DeploymentType.WithLabelValues(utils.Config.DeploymentType).Set(1)
	if db.ClickHouseNativeWriter == nil {
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
		)
	}

	for _, service := range monitoredServices {
		service.InitServices()
	}
}

func Start() {
	for _, service := range monitoredServices {
		service.Start()
	}
}

func Stop() {
	for _, service := range monitoredServices {
		service.Stop()
	}
}
