package main

import (
	"errors"
	"flag"
	"fmt"

	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type options struct {
	configPath                string
	statisticsDayToExport     int64
	statisticsDaysToExport    string
	statisticsValidatorToggle bool
	statisticsChartToggle     bool
	statisticsGraffitiToggle  bool
	resetStatus               bool
}

var opt = &options{}

func main() {
	flag.StringVar(&opt.configPath, "config", "", "Path to the config file")
	flag.Int64Var(&opt.statisticsDayToExport, "statistics.day", -1, "Day to export statistics (will export the day independent if it has been already exported or not")
	flag.StringVar(&opt.statisticsDaysToExport, "statistics.days", "", "Days to export statistics (will export the day independent if it has been already exported or not")
	flag.BoolVar(&opt.statisticsValidatorToggle, "validators.enabled", false, "Toggle exporting validator statistics")
	flag.BoolVar(&opt.statisticsChartToggle, "charts.enabled", false, "Toggle exporting chart series")
	flag.BoolVar(&opt.statisticsGraffitiToggle, "graffiti.enabled", false, "Toggle exporting graffiti statistics")
	flag.BoolVar(&opt.resetStatus, "validators.reset", false, "Export stats independent if they have already been exported previously")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		log.Infof(version.Version)
		log.Infof(version.GoVersion)
		return
	}

	log.Infof("version: %v, config file path: %v", version.Version, opt.configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opt.configPath)

	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	if utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 || utils.Config.Chain.ClConfig.SecondsPerSlot == 0 {
		log.Fatal(fmt.Errorf("error ether SlotsPerEpoch [%v] or SecondsPerSlot [%v] are not set", utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SecondsPerSlot), "", 0)
		return
	} else {
		log.Infof("Writing statistic with: SlotsPerEpoch [%v] or SecondsPerSlot [%v]", utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SecondsPerSlot)
	}

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			log.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}(utils.Config.Metrics.Address)
	}

	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		SSL:          cfg.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		SSL:          cfg.ReaderDatabase.SSL,
	}, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	db.FrontendWriterDB, db.FrontendReaderDB = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.Frontend.WriterDatabase.Username,
		Password:     cfg.Frontend.WriterDatabase.Password,
		Name:         cfg.Frontend.WriterDatabase.Name,
		Host:         cfg.Frontend.WriterDatabase.Host,
		Port:         cfg.Frontend.WriterDatabase.Port,
		MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.Frontend.ReaderDatabase.Username,
		Password:     cfg.Frontend.ReaderDatabase.Password,
		Name:         cfg.Frontend.ReaderDatabase.Name,
		Host:         cfg.Frontend.ReaderDatabase.Host,
		Port:         cfg.Frontend.ReaderDatabase.Port,
		MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
	}, "pgx", "postgres")
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}

	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(nil, "No cache provider set. Please set TierdCacheProvider (example redis)", 0)
	}

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	}

	var rpcClient rpc.Client

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	if utils.Config.Indexer.Node.Type == "lighthouse" {
		cl := consapi.NewClient("http://" + cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
		nodeImpl, ok := cl.ClientInt.(*consapi.NodeClient)
		if !ok {
			log.Fatal(nil, "lighthouse client can only be used with real node impl", 0)
		}

		rpcClient, err = rpc.NewLighthouseClient(nodeImpl, chainID)
		if err != nil {
			log.Fatal(err, "new explorer lighthouse client error", 0)
		}
	} else {
		log.Fatal(errors.New("invalid node type"), fmt.Sprintf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type), 0)
	}

	if opt.statisticsDaysToExport != "" {
		s := strings.Split(opt.statisticsDaysToExport, "-")
		if len(s) < 2 {
			log.Fatal(errors.New("invalid arg"), "invalid arg", 0)
		}
		firstDay, err := strconv.ParseUint(s[0], 10, 64)
		if err != nil {
			log.Fatal(err, "error parsing first day of statisticsDaysToExport flag to uint", 0)
		}
		lastDay, err := strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			log.Fatal(err, "error parsing last day of statisticsDaysToExport flag to uint", 0)
		}

		if opt.statisticsValidatorToggle {
			log.Infof("exporting validator statistics for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {
				if opt.resetStatus {
					clearStatsStatusTable(d)
				}

				err = db.WriteValidatorStatisticsForDay(d, rpcClient)
				if err != nil {
					log.Error(err, fmt.Errorf("error exporting stats for day %v", d), 0)
					break
				}
			}
		}

		if opt.statisticsChartToggle {
			log.Infof("exporting chart series for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {
				_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", d)
				if err != nil {
					log.Fatal(err, "error resetting status for chart series status for day", 0, log.Fields{"day": d})
				}

				err = db.WriteChartSeriesForDay(int64(d))
				if err != nil {
					log.Error(err, "error exporting chart series from day", 0, log.Fields{"day": d})
					break
				}
			}
		}

		if opt.statisticsGraffitiToggle {
			for d := firstDay; d <= lastDay; d++ {
				err = db.WriteGraffitiStatisticsForDay(int64(d))
				if err != nil {
					log.Error(err, fmt.Errorf("error exporting graffiti-stats from day %v", opt.statisticsDayToExport), 0)
					break
				}
			}
		}

		return
	} else if opt.statisticsDayToExport >= 0 {
		if opt.statisticsValidatorToggle {
			if opt.resetStatus {
				clearStatsStatusTable(uint64(opt.statisticsDayToExport))
			}

			err = db.WriteValidatorStatisticsForDay(uint64(opt.statisticsDayToExport), rpcClient)
			if err != nil {
				log.Error(err, fmt.Errorf("error exporting stats for day %v", opt.statisticsDayToExport), 0)
			}
		}

		if opt.statisticsChartToggle {
			_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", opt.statisticsDayToExport)
			if err != nil {
				log.Fatal(err, "error resetting status for chart series status for day", 0, log.Fields{"day": opt.statisticsDayToExport})
			}

			err = db.WriteChartSeriesForDay(opt.statisticsDayToExport)
			if err != nil {
				log.Error(err, fmt.Errorf("error exporting chart series from day %v", opt.statisticsDayToExport), 0)
			}
		}

		if opt.statisticsGraffitiToggle {
			err = db.WriteGraffitiStatisticsForDay(opt.statisticsDayToExport)
			if err != nil {
				log.Error(err, fmt.Errorf("error exporting graffiti-stats from day %v", opt.statisticsDayToExport), 0)
			}
		}
		return
	}

	go statisticsLoop(rpcClient)

	utils.WaitForCtrlC()

	log.Infof("exiting...")
}

func statisticsLoop(client rpc.Client) {
	for {
		var loopError error
		latestEpoch := cache.LatestFinalizedEpoch.Get()
		if latestEpoch == 0 {
			log.Error(nil, "error retreiving latest finalized epoch from cache", 0)
			time.Sleep(time.Minute)
			continue
		}

		epochsPerDay := utils.EpochsPerDay()
		if latestEpoch < epochsPerDay {
			log.Infof("skipping exporting stats, first day has not been indexed yet")
			time.Sleep(time.Minute)
			continue
		}
		currentDay := latestEpoch / epochsPerDay
		previousDay := currentDay - 1

		log.Infof("Performing statisticsLoop with currentDay %v and previousDay %v", currentDay, previousDay)
		if previousDay > currentDay {
			previousDay = currentDay
		}

		if opt.statisticsValidatorToggle {
			lastExportedDayValidator, err := db.GetLastExportedStatisticDay()
			if err != nil {
				log.Error(err, "error retreiving latest exported day from the db", 0)
			}

			log.Infof("Validator Statistics: Latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayValidator)
			if lastExportedDayValidator != 0 {
				lastExportedDayValidator++
			}
			if lastExportedDayValidator <= previousDay || lastExportedDayValidator == 0 {
				for day := lastExportedDayValidator; day <= previousDay; day++ {
					err := db.WriteValidatorStatisticsForDay(day, client)
					if err != nil {
						log.Error(err, fmt.Errorf("error exporting stats for day %v", day), 0)
						loopError = err
						break
					}
				}
			}
		}

		if opt.statisticsChartToggle {
			var lastExportedDayChart uint64
			err := db.WriterDb.Get(&lastExportedDayChart, "select COALESCE(max(day), 0) from chart_series_status where status")
			if err != nil {
				log.Error(err, "error retreiving latest exported day from the db", 0)
			}

			log.Infof("Chart statistics: latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayChart)
			if lastExportedDayChart != 0 {
				lastExportedDayChart++
			}
			if lastExportedDayChart <= previousDay || lastExportedDayChart == 0 {
				for day := lastExportedDayChart; day <= previousDay; day++ {
					err = db.WriteChartSeriesForDay(int64(day))
					if err != nil {
						log.Error(err, fmt.Errorf("error exporting chart series from day %v", day), 0)
						loopError = err
						break
					}
				}
			}
		}

		if opt.statisticsGraffitiToggle {
			graffitiStatsStatus := []struct {
				Day    uint64
				Status bool
			}{}
			err := db.WriterDb.Select(&graffitiStatsStatus, "select day, status from graffiti_stats_status")
			if err != nil {
				log.Error(err, "error retrieving graffitiStatsStatus", 0)
			} else {
				graffitiStatsStatusMap := map[uint64]bool{}
				for _, s := range graffitiStatsStatus {
					graffitiStatsStatusMap[s.Day] = s.Status
				}
				for day := uint64(0); day <= currentDay; day++ {
					if !graffitiStatsStatusMap[day] {
						log.Infof("exporting graffiti-stats for day %v", day)
						err = db.WriteGraffitiStatisticsForDay(int64(day))
						if err != nil {
							log.Error(err, fmt.Errorf("error exporting graffiti-stats for day %v", day), 0)
							loopError = err
							break
						}
					}
				}
			}
		}

		if loopError == nil {
			services.ReportStatus("statistics", "Running", nil)
		} else {
			services.ReportStatus("statistics", loopError.Error(), nil)
		}
		time.Sleep(time.Minute)
	}
}

func clearStatsStatusTable(day uint64) {
	log.Infof("deleting validator_stats_status for day %v", day)
	_, err := db.WriterDb.Exec("DELETE FROM validator_stats_status WHERE day = $1", day)
	if err != nil {
		log.Fatal(err, "error resetting status for day", 0, log.Fields{"day": day})
	}
}
