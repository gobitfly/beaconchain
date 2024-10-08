package node_jobs_processor

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/nodejobs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := fs.Bool("version", false, "Print version and exit")
	_ = fs.Parse(os.Args[2:])

	if *versionFlag {
		log.Info(version.Version)
		log.Info(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg
	log.InfoWithFields(log.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	if utils.Config.Metrics.Enabled {
		go func() {
			log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
			if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof, utils.Config.Metrics.PprofExtra); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}()
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

	nrp := NewNodeJobsProcessor(utils.Config.NodeJobsProcessor.ClEndpoint, utils.Config.NodeJobsProcessor.ElEndpoint)
	go nrp.Run()

	utils.WaitForCtrlC()
	log.Infof("exiting …")
}

type NodeJobsProcessor struct {
	ELAddr string
	CLAddr string
}

func NewNodeJobsProcessor(clAddr, elAddr string) *NodeJobsProcessor {
	njp := &NodeJobsProcessor{
		CLAddr: clAddr,
		ELAddr: elAddr,
	}
	return njp
}

func (njp *NodeJobsProcessor) Run() {
	for {
		err := njp.Process()
		if err != nil {
			log.Error(err, "error processing node-jobs", 0)
		}
		time.Sleep(time.Second * 10)
	}
}

func (njp *NodeJobsProcessor) Process() error {
	err := nodejobs.UpdateNodeJobs()
	if err != nil {
		return fmt.Errorf("error updating job: %w", err)
	}
	err = nodejobs.SubmitNodeJobs()
	if err != nil {
		return fmt.Errorf("error submitting job: %w", err)
	}
	return nil
}
