package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

func InitBigtableSchema() error {
	err := InitBigtableSchemaIndexed()
	if err != nil {
		return err
	}
	err = InitBigtableSchemaRaw()
	if err != nil {
		return err
	}
	return nil
}

func InitBigtableSchemaIndexed() error {
	tables := make(map[string]map[string]gcp_bigtable.GCPolicy)

	tables["beaconchain_validators"] = map[string]gcp_bigtable.GCPolicy{
		ATTESTATIONS_FAMILY: gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["beaconchain_validators_history"] = map[string]gcp_bigtable.GCPolicy{
		VALIDATOR_BALANCES_FAMILY:             nil,
		VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY: nil,
		ATTESTATIONS_FAMILY:                   gcp_bigtable.MaxAgeGCPolicy(utils.Day * 31),
		SYNC_COMMITTEES_FAMILY:                nil,
		INCOME_DETAILS_COLUMN_FAMILY:          gcp_bigtable.MaxAgeGCPolicy(utils.Day * 31),
		STATS_COLUMN_FAMILY:                   nil,
	}
	tables["blocks"] = map[string]gcp_bigtable.GCPolicy{
		DEFAULT_FAMILY_BLOCKS: gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["data"] = map[string]gcp_bigtable.GCPolicy{
		CONTRACT_METADATA_FAMILY: gcp_bigtable.MaxAgeGCPolicy(utils.Day),
		DEFAULT_FAMILY:           nil,
	}
	tables["machine_metrics"] = map[string]gcp_bigtable.GCPolicy{
		MACHINE_METRICS_COLUMN_FAMILY: gcp_bigtable.MaxAgeGCPolicy(utils.Day * 31),
	}
	tables["metadata"] = map[string]gcp_bigtable.GCPolicy{
		ACCOUNT_METADATA_FAMILY:  nil,
		CONTRACT_METADATA_FAMILY: nil,
		ERC20_METADATA_FAMILY:    nil,
		ERC721_METADATA_FAMILY:   nil,
		ERC1155_METADATA_FAMILY:  nil,
		SERIES_FAMILY:            gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["metadata_updates"] = map[string]gcp_bigtable.GCPolicy{
		METADATA_UPDATES_FAMILY_BLOCKS: gcp_bigtable.MaxAgeGCPolicy(utils.Day),
		DEFAULT_FAMILY:                 nil,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()

	if utils.Config.Bigtable.Emulator {
		if utils.Config.Bigtable.EmulatorHost == "" {
			utils.Config.Bigtable.EmulatorHost = "127.0.0.1"
		}
		log.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to %s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort))

		if err != nil {
			log.Fatal(err, "unable to set bigtable emulator environment variable", 0)
		}
	}

	admin, err := gcp_bigtable.NewAdminClient(ctx, utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance)
	if err != nil {
		return err
	}

	existingTables, err := admin.Tables(ctx)
	if err != nil {
		return err
	}

	if len(existingTables) > 0 {
		return fmt.Errorf("aborting bigtable schema init as tables are already present")
	}

	for name, definition := range tables {
		err := admin.CreateTable(ctx, name)
		if err != nil {
			return err
		}

		for columnFamily, gcPolicy := range definition {
			err := admin.CreateColumnFamily(ctx, name, columnFamily)
			if err != nil {
				return err
			}

			if gcPolicy != nil {
				err := admin.SetGCPolicy(ctx, name, columnFamily, gcPolicy)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func InitBigtableSchemaRaw() error {
	tables := make(map[string]map[string]gcp_bigtable.GCPolicy)

	// these are the families in current production-bigtable: b, v, r, a, s, t, u - not sure why some are missing in the raw-go-module
	tables["blocks-raw"] = map[string]gcp_bigtable.GCPolicy{
		raw.BT_COLUMNFAMILY_BLOCK:    nil,
		"v":                          nil,
		raw.BT_COLUMNFAMILY_RECEIPTS: nil,
		"a":                          nil,
		"s":                          nil,
		raw.BT_COLUMNFAMILY_TRACES:   nil,
		raw.BT_COLUMNFAMILY_UNCLES:   nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()

	if utils.Config.RawBigtable.Emulator {
		if utils.Config.RawBigtable.EmulatorHost == "" {
			utils.Config.RawBigtable.EmulatorHost = "127.0.0.1"
		}
		log.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to %s:%d", utils.Config.RawBigtable.EmulatorHost, utils.Config.RawBigtable.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", utils.Config.RawBigtable.EmulatorHost, utils.Config.RawBigtable.EmulatorPort))

		if err != nil {
			log.Fatal(err, "unable to set bigtable emulator environment variable", 0)
		}
	}

	admin, err := gcp_bigtable.NewAdminClient(ctx, utils.Config.RawBigtable.Project, utils.Config.RawBigtable.Instance)
	if err != nil {
		return err
	}

	existingTables, err := admin.Tables(ctx)
	if err != nil {
		return err
	}

	if len(existingTables) > 0 {
		return fmt.Errorf("aborting bigtable schema init as tables are already present")
	}

	for name, definition := range tables {
		err := admin.CreateTable(ctx, name)
		if err != nil {
			return err
		}

		for columnFamily, gcPolicy := range definition {
			err := admin.CreateColumnFamily(ctx, name, columnFamily)
			if err != nil {
				return err
			}

			if gcPolicy != nil {
				err := admin.SetGCPolicy(ctx, name, columnFamily, gcPolicy)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
