package modules

import (
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/exporter/rpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "exporter")

type ModuleInterface interface {
	Start(args []any) error
}

type ModuleContext struct {
	CL         consapi.Retriever
	ConsClient *rpc.LighthouseClient
}

var Client *rpc.Client

// Start will start the export of data from rpc into the database
func StartAll(context ModuleContext) {
	go networkLivenessUpdater(context.ConsClient)
	go eth1DepositsExporter()
	go genesisDepositsExporter(context.ConsClient)
	go checkSubscriptions()
	go syncCommitteesExporter(context.ConsClient)
	go syncCommitteesCountExporter()
	if utils.Config.SSVExporter.Enabled {
		go ssvExporter()
	}
	if utils.Config.RocketpoolExporter.Enabled {
		go rocketpoolExporter()
	}

	if utils.Config.Indexer.PubKeyTagsExporter.Enabled {
		go UpdatePubkeyTag()
	}

	if utils.Config.MevBoostRelayExporter.Enabled {
		go mevBoostRelaysExporter()
	}
	// wait until the beacon-node is available
	for {
		head, err := context.ConsClient.GetChainHead()
		if err == nil {
			logger.Infof("beacon node is available with head slot: %v", head.HeadSlot)
			break
		}
		logger.Errorf("beacon-node seems to be unavailable: %v", err)
		time.Sleep(time.Second * 10)
	}

	firstRun := true

	slotExporter := NewSlotExporter(context)

	minWaitTimeBetweenRuns := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)
	for {
		start := time.Now()
		err := slotExporter.Start(nil)
		if err != nil {
			logrus.Errorf("error during slot export run: %v", err)
		} else if err == nil && firstRun {
			firstRun = false
		}

		logrus.Info("update run completed")
		elapsed := time.Since(start)
		if elapsed < minWaitTimeBetweenRuns {
			time.Sleep(minWaitTimeBetweenRuns - elapsed)
		}

		services.ReportStatus("slotExporter", "Running", nil)
	}
}

func GetModuleContext() (ModuleContext, error) {
	cl := consapi.NewNodeDataRetriever(utils.Config.NodeJobsProcessor.ClEndpoint)

	spec, err := cl.GetSpec()
	if err != nil {
		utils.LogFatal(err, "error getting spec", 0)
	}

	config.ClConfig = &spec.Data

	nodeImpl, ok := cl.RetrieverInt.(*consapi.NodeImplRetriever)
	if !ok {
		return ModuleContext{}, errors.New("lighthouse client can only be used with real node impl")
	}

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	clClient, err := rpc.NewLighthouseClient(nodeImpl, chainID)
	if err != nil {
		utils.LogFatal(err, "error creating lighthouse client", 0)
	}

	moduleContext := ModuleContext{
		CL:         cl,
		ConsClient: clClient,
	}

	return moduleContext, nil
}
