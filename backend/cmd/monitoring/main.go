package monitoring

import (
	"flag"
	"os"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/monitoring"
)

func Run() {
	fs := flag.NewFlagSet("monitoring", flag.ExitOnError)
	configPath := fs.String("config", "config.yml", "path to config")
	versionFlag := fs.Bool("version", false, "print version and exit")
	_ = fs.Parse(os.Args[2:])

	if *versionFlag {
		log.Infof("%s", version.Version)
		log.Infof("%s", version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	db.ClickHouseWriter, db.ClickHouseReader = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.ClickHouse.WriterDatabase.Username,
		Password:     cfg.ClickHouse.WriterDatabase.Password,
		Name:         cfg.ClickHouse.WriterDatabase.Name,
		Host:         cfg.ClickHouse.WriterDatabase.Host,
		Port:         cfg.ClickHouse.WriterDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.WriterDatabase.MaxOpenConns,
		SSL:          true,
		MaxIdleConns: cfg.ClickHouse.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ClickHouse.ReaderDatabase.Username,
		Password:     cfg.ClickHouse.ReaderDatabase.Password,
		Name:         cfg.ClickHouse.ReaderDatabase.Name,
		Host:         cfg.ClickHouse.ReaderDatabase.Host,
		Port:         cfg.ClickHouse.ReaderDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.ReaderDatabase.MaxOpenConns,
		SSL:          true,
		MaxIdleConns: cfg.ClickHouse.ReaderDatabase.MaxIdleConns,
	}, "clickhouse", "clickhouse")
	defer db.ClickHouseReader.Close()
	defer db.ClickHouseWriter.Close()

	monitoring.Init(true)
	monitoring.Start()
	defer monitoring.Stop()

	// gotta wait forever
	utils.WaitForCtrlC()
}
