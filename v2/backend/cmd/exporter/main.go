package main

import (
	"flag"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DBDSN  string
	CLNode string
}

var conf Config

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&conf.DBDSN, "db.dsn", "postgres://user:pass@host:port/dbnames", "data-source-name of db, if it starts with projects/ it will use gcp-secretmanager")
	flag.StringVar(&conf.CLNode, "cl.endpoint", "http://localhost:4000", "cl node endpoint")
	flag.Parse()

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	logrus.WithFields(logrus.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"commit":    version.GitCommit,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}).Printf("starting")

	// err := db.InitWithDSN(conf.DBDSN)

	go startModules()

	// Keep the program alive until Ctrl+C is pressed
	utils.WaitForCtrlC()
}

func startModules() {
	cl := consapi.NewNodeDataRetriever(conf.CLNode)

	spec, err := cl.GetSpec()
	if err != nil {
		utils.LogFatal(err, "error getting spec", 0)
	}

	config.ClConfig = &spec.Data
	//utils.Config.Chain.ClConfig = spec.Data

	moduleContext := modules.ModuleContext{
		CL: cl,
		// TODO: EL, DB
	}

	// slot, err := cl.GetSlot(128038)
	// if err != nil {
	// 	utils.LogFatal(err, "error getting slot", 0)
	// }

	// domain, err := utils.GetSigningDomain()
	// if err != nil {
	// 	utils.LogFatal(err, "can not get signing domain", 0)
	// 	return
	// }

	// depositData := slot.Data.Message.Body.Deposits[0]
	// err = utils.VerifyDepositSignature(&phase0.DepositData{
	// 	PublicKey:             phase0.BLSPubKey(utils.MustParseHex(depositData.Data.Pubkey)),
	// 	WithdrawalCredentials: utils.MustParseHex(depositData.Data.WithdrawalCredentials),
	// 	Amount:                phase0.Gwei(uint64(depositData.Data.Amount)),
	// 	Signature:             phase0.BLSSignature(utils.MustParseHex(depositData.Data.Signature)),
	// }, domain)

	// if err != nil {

	// 	utils.LogFatal(err, "can not verify deposit signature", 0)
	// }
	// return

	registeredModules := []modules.ModuleInterfaceEpoch{
		modules.NewDashboardDataModule(moduleContext),
		// todo: add more modules here
	}

	// result, err := cl.GetValidator(1, "head")
	// if err != nil {
	// 	utils.LogFatal(err, "error getting validator", 0)
	// }
	// fmt.Printf("result:%v", result)

	for _, module := range registeredModules {
		module.Start(28356)
	}
}
