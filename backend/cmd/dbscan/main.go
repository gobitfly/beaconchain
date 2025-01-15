package dbscan

import (
	"flag"
	"fmt"
	"os"

	"google.golang.org/api/option"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)

	configPath := fs.String("config", "config/default.config.yml", "Path to the config file")
	startBlocks := fs.Int64("start", 0, "Block to start scanning")
	endBlocks := fs.Int64("end", 0, "Block to finish scanning")
	chainID := fs.Uint64("chainID", 0, "Chain ID of the chain")

	_ = fs.Parse(os.Args[2:])

	if *endBlocks == 0 {
		log.Fatal(nil, "end flag must be specified", 0)
	}
	if *chainID == 0 {
		log.Fatal(nil, "chainID flag must be specified", 0)
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	} else {
		log.Info("reading config completed")
	}
	utils.Config = cfg

	bt, err := database.NewBigTable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, nil, option.WithGRPCConnectionPool(1))
	if err != nil {
		log.Fatal(err, "creating new client for Bigtable", 0)
	}
	defer bt.Close()
	store := raw.NewStore(database.Wrap(bt, raw.Table))

	if err := Scan(*startBlocks, *endBlocks, *chainID, store); err != nil {
		log.Fatal(err, "cannot scan", 0)
	}
}

func Scan(start, end int64, chainID uint64, store raw.Store) error {
	batch := int64(50)
	for i := start; i <= end; i += batch + 1 {
		localEnd := i + batch
		if localEnd > end {
			localEnd = end
		}
		blocks, err := store.ReadBlocksByNumber(chainID, i, localEnd)
		if err != nil {
			return fmt.Errorf("error reading blocks %d-%d: %w", i, localEnd, err)
		}
		for _, block := range blocks {
			if err := raw.ValidateBlock(*block); err != nil {
				log.Error(fmt.Errorf("block %d: %w", block.BlockNumber, err), "invalid block found", 0)
				continue
			}
			log.Info(fmt.Sprintf("block %d valid", block.BlockNumber))
		}
	}
	return nil
}
