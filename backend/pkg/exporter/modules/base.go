package modules

import (
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/pkg/errors"
)

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
			log.Infof("beacon node is available with head slot: %v", head.HeadSlot)
			break
		}
		log.Error(err, "beacon-node seems to be unavailable", 0)
		time.Sleep(time.Second * 10)
	}

	firstRun := true

	slotExporter := NewSlotExporter(context)

	minWaitTimeBetweenRuns := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)
	for {
		start := time.Now()
		err := slotExporter.Start(nil)
		if err != nil {
			log.Error(err, "error during slot export run", 0)
		} else if err == nil && firstRun {
			firstRun = false
		}

		log.Infof("update run completed")
		elapsed := time.Since(start)
		if elapsed < minWaitTimeBetweenRuns {
			time.Sleep(minWaitTimeBetweenRuns - elapsed)
		}

		services.ReportStatus("slotExporter", "Running", nil)
	}
}

func GetModuleContext() (ModuleContext, error) {
	cl := consapi.NewNodeDataRetriever("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port)

	spec, err := cl.GetSpec()
	if err != nil {
		log.Fatal(err, "error getting spec", 0)
	}

	config.ClConfig = &spec.Data

	nodeImpl, ok := cl.RetrieverInt.(*consapi.NodeImplRetriever)
	if !ok {
		return ModuleContext{}, errors.New("lighthouse client can only be used with real node impl")
	}

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	clClient, err := rpc.NewLighthouseClient(nodeImpl, chainID)
	if err != nil {
		log.Fatal(err, "error creating lighthouse client", 0)
	}

	moduleContext := ModuleContext{
		CL:         cl,
		ConsClient: clClient,
	}

	return moduleContext, nil
}
