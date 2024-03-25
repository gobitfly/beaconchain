package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	eth_rewards "github.com/gobitfly/eth-rewards"
	"github.com/gobitfly/eth-rewards/beacon"
	"golang.org/x/sync/errgroup"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	bnAddress := flag.String("beacon-node-address", "", "Url of the beacon node api")
	enAddress := flag.String("execution-node-address", "", "Url of the execution node api")
	epoch := flag.Int64("epoch", -1, "epoch to export (use -1 to export latest finalized epoch)")
	batchConcurrency := flag.Int("batch-concurrency", 5, "epoch to export at the same time (only for historic)")

	epochStart := flag.Uint64("epoch-start", 0, "start epoch to export")
	epochEnd := flag.Uint64("epoch-end", 0, "end epoch to export")
	sleepDuration := flag.Duration("sleep", time.Minute, "duration to sleep between export runs")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		log.Infof(version.Version)
		log.Infof(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if bnAddress == nil || *bnAddress == "" {
		if utils.Config.Indexer.Node.Host == "" {
			log.Fatal(nil, "no beacon node url provided", 0)
		} else {
			log.Infof("applying becon node endpoint from config")
			*bnAddress = fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Indexer.Node.Host, cfg.Indexer.Node.Port))
		}
	}

	if enAddress == nil || *enAddress == "" {
		if utils.Config.Eth1ErigonEndpoint == "" {
			log.Fatal(nil, "no execution node url provided", 0)
		} else {
			log.Infof("applying execution node endpoint from config")
			*enAddress = utils.Config.Eth1ErigonEndpoint
		}
	}

	client := beacon.NewClient(*bnAddress, time.Minute*5)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}
	defer bt.Close()

	cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())

	// Initialize the persistent redis client
	rdc := redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisSessionStoreEndpoint,
		ReadTimeout: time.Second * 20,
	})

	if err := rdc.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err, "error connecting to persistent redis store", 0)
	}

	db.PersistentRedisDbClient = rdc

	if *epochEnd != 0 {
		latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
		if *epochEnd > latestFinalizedEpoch {
			log.Error(fmt.Errorf("error epochEnd [%v] is greater then latestFinalizedEpoch [%v]", epochEnd, latestFinalizedEpoch), "", 0)
			return
		}
		g := errgroup.Group{}
		g.SetLimit(*batchConcurrency)

		start := time.Now()
		epochsCompleted := int64(0)
		notExportedEpochs := []uint64{}
		err = db.WriterDb.Select(&notExportedEpochs, "SELECT epoch FROM epochs WHERE NOT rewards_exported AND epoch >= $1 AND epoch <= $2 ORDER BY epoch DESC", *epochStart, *epochEnd)
		if err != nil {
			log.Fatal(err, "error retrieving not exported epochs from db", 0)
		}
		epochsToExport := int64(len(notExportedEpochs))

		go func() {
			for {
				c := atomic.LoadInt64(&epochsCompleted)

				if c == 0 {
					time.Sleep(time.Second)
					continue
				}

				epochsRemaining := epochsToExport - c

				elapsed := time.Since(start)
				remaining := time.Duration(epochsRemaining * time.Since(start).Nanoseconds() / c)
				epochDuration := time.Duration(elapsed.Nanoseconds() / c)

				log.Infof("exported %v of %v epochs in %v (%v/epoch), estimated time remaining: %vs", c, epochsToExport, elapsed, epochDuration, remaining)
				time.Sleep(time.Second * 10)
			}
		}()

		for _, e := range notExportedEpochs {
			e := e
			g.Go(func() error {
				var err error
				for i := 0; i < 10; i++ {
					err = export(e, bt, client, enAddress)

					if err != nil {
						log.Error(err, "error exporting rewards for epoch, retrying", 0, map[string]interface{}{"epoch": e})
					} else {
						break
					}
				}
				if err != nil {
					log.Error(err, "error exporting rewards for epoch", 0, map[string]interface{}{"epoch": e})
					return nil
				}

				_, err = db.WriterDb.Exec("UPDATE epochs SET rewards_exported = true WHERE epoch = $1", e)

				if err != nil {
					log.Error(err, "error rewards_exported as true for epoch", 0, map[string]interface{}{"epoch": e})
				}

				atomic.AddInt64(&epochsCompleted, 1)
				return nil
			})
		}
		err = g.Wait()
		if err != nil {
			log.Error(err, "error during epoch rewards export", 0)
		}
		return
	}

	if *epoch == -1 {
		lastExportedEpoch := uint64(0)
		for {
			latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
			notExportedEpochs := []uint64{}
			err = db.WriterDb.Select(&notExportedEpochs, "SELECT epoch FROM epochs WHERE NOT rewards_exported AND epoch > $1 AND epoch <= $2 ORDER BY epoch desc LIMIT 10", lastExportedEpoch, latestFinalizedEpoch)
			if err != nil {
				log.Fatal(err, "getting chain head from lighthouse error", 0)
			}
			for _, e := range notExportedEpochs {
				err := export(e, bt, client, enAddress)

				if err != nil {
					log.Error(err, "error exporting rewards for epoch, retrying", 0, map[string]interface{}{"epoch": e})
					continue
				}

				_, err = db.WriterDb.Exec("UPDATE epochs SET rewards_exported = true WHERE epoch = $1", e)

				if err != nil {
					log.Error(err, "error rewards_exported as true for epoch", 0, map[string]interface{}{"epoch": e})
				}
				services.ReportStatus("rewardsExporter", "Running", nil)

				if e > lastExportedEpoch {
					lastExportedEpoch = e
				}
			}

			services.ReportStatus("rewardsExporter", "Running", nil)
			time.Sleep(*sleepDuration)
		}
	}

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	if *epoch > int64(latestFinalizedEpoch) {
		log.Error(fmt.Errorf("error epoch [%v] is greater then latestFinalizedEpoch [%v]", epoch, latestFinalizedEpoch), "", 0)
		return
	}
	err = export(uint64(*epoch), bt, client, enAddress)
	if err != nil {
		log.Fatal(err, "error during epoch export", 0, map[string]interface{}{"epoch": *epoch})
	}
}

func export(epoch uint64, bt *db.Bigtable, client *beacon.Client, elClient *string) error {
	start := time.Now()
	log.Infof("retrieving rewards details for epoch %v", epoch)

	rewards, err := eth_rewards.GetRewardsForEpoch(epoch, client, *elClient)

	if err != nil {
		return fmt.Errorf("error retrieving reward details for epoch %v: %v", epoch, err)
	} else {
		log.Infof("retrieved %v reward details for epoch %v in %v", len(rewards), epoch, time.Since(start))
	}

	redisCachedEpochRewards := &types.RedisCachedEpochRewards{
		Epoch:   types.Epoch(epoch),
		Rewards: rewards,
	}

	var serializedRewardsData bytes.Buffer
	enc := gob.NewEncoder(&serializedRewardsData)
	err = enc.Encode(redisCachedEpochRewards)
	if err != nil {
		return fmt.Errorf("error serializing rewards data for epoch %v: %w", epoch, err)
	}

	key := fmt.Sprintf("%d:%s:%d", utils.Config.Chain.ClConfig.DepositChainID, "er", epoch)

	expirationTime := utils.EpochToTime(epoch + 7) // keep it for at least 7 epochs in the cache
	expirationDuration := time.Until(expirationTime)
	log.Infof("writing rewards data to redis with a TTL of %v", expirationDuration)
	err = db.PersistentRedisDbClient.Set(context.Background(), key, serializedRewardsData.Bytes(), expirationDuration).Err()
	if err != nil {
		return fmt.Errorf("error writing rewards data to redis for epoch %v: %w", epoch, err)
	}
	log.Infof("writing epoch rewards to redis completed")

	log.Infof("exporting duties & balances for epoch %v", epoch)

	err = bt.SaveValidatorIncomeDetails(epoch, rewards)
	if err != nil {
		return fmt.Errorf("error saving reward details to bigtable: %v", err)
	}
	return nil
}
