package data_sources

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"google.golang.org/api/option"
)

type Bigtable struct {
	client *gcp_bigtable.Client

	tableBeaconchain       *gcp_bigtable.Table
	tableValidators        *gcp_bigtable.Table
	tableValidatorsHistory *gcp_bigtable.Table

	tableData            *gcp_bigtable.Table
	tableBlocks          *gcp_bigtable.Table
	tableMetadataUpdates *gcp_bigtable.Table
	tableMetadata        *gcp_bigtable.Table

	tableMachineMetrics *gcp_bigtable.Table

	LastAttestationCache    map[uint64]uint64
	LastAttestationCacheMux *sync.Mutex
}

func InitBigtable(ctx context.Context, bigtableConfig *config.BigtableConfig) (*Bigtable, error) {
	if bigtableConfig.Emulator {
		if bigtableConfig.EmulatorHost == "" {
			bigtableConfig.EmulatorHost = "127.0.0.1"
		}
		log.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to %s:%d", bigtableConfig.EmulatorHost, bigtableConfig.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", bigtableConfig.EmulatorHost, bigtableConfig.EmulatorPort))

		if err != nil {
			log.Fatal(err, "unable to set bigtable emulator environment variable", 0)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	poolSize := 50
	btClient, err := gcp_bigtable.NewClient(context.Background(), bigtableConfig.Project, bigtableConfig.Instance, option.WithGRPCConnectionPool(poolSize))

	if err != nil {
		return nil, err
	}

	bt := &Bigtable{
		client:                  btClient,
		tableData:               btClient.Open("data"),
		tableBlocks:             btClient.Open("blocks"),
		tableMetadataUpdates:    btClient.Open("metadata_updates"),
		tableMetadata:           btClient.Open("metadata"),
		tableBeaconchain:        btClient.Open("beaconchain"),
		tableMachineMetrics:     btClient.Open("machine_metrics"),
		tableValidators:         btClient.Open("beaconchain_validators"),
		tableValidatorsHistory:  btClient.Open("beaconchain_validators_history"),
		LastAttestationCacheMux: &sync.Mutex{},
	}

	return bt, nil
}
