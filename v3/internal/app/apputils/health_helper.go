package apputils

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
)

func InitHealthHandler(dataSources *data_sources.ApiDataSources) *HealthHandler {
	timeout := 10 * time.Second

	return NewHealthHandler(map[string]Pingable{
		"Redis":     WithTimeout(func(ctx context.Context) error { return dataSources.Redis.Ping(ctx).Err() }, timeout),
		"RoAdminDb": WithTimeout(dataSources.RoAdminDb.DB.PingContext, timeout),
		"RwAdminDb": WithTimeout(dataSources.RwAdminDb.DB.PingContext, timeout),
		"RoChDb":    WithTimeout(dataSources.RoChDb.DB.PingContext, timeout),
		"RwChDb":    WithTimeout(dataSources.RwChDb.DB.PingContext, timeout),
	})
}

func WithTimeout(p Pingable, d time.Duration) Pingable {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return p.PingContext(ctx)
	}
}
