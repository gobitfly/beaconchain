package data_sources

import (
	"context"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"google.golang.org/api/option"
)

type Bigtable struct {
	client *gcp_bigtable.Client
}

func InitBigtable(ctx context.Context, bigtableConfig *config.BigtableConfig) (*Bigtable, error) {
	poolSize := 50
	btClient, err := gcp_bigtable.NewClient(ctx, bigtableConfig.Project, bigtableConfig.Instance, option.WithGRPCConnectionPool(poolSize))

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client: btClient,
	}

	return bt, nil
}
