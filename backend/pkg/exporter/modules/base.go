package modules

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ModuleInterface interface {
	Init() error
	GetName() string // Used for logging
	GetMonitoringEventId() constants.Event

	OnHead(*types.StandardEventHeadResponse) error // !Do not block in this functions for an extended period of time!

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	OnFinalizedCheckpoint(*types.StandardFinalizedCheckpointResponse) error // !Do not block in this functions for an extended period of time!

	OnChainReorg(*types.StandardEventChainReorg) error // !Do not block in this functions for an extended period of time!
}

var Client *rpc.Client

// Start will start the export of data from rpc into the database
func StartAll(moduleCtx ModuleContext, modules []ModuleInterface, justV2 bool) {
	services.InitStatusReport()
	if !justV2 {
		ctx := context.Background()
		consDB := db2.NewConsensusRepository(db.ReaderDb, db.WriterDb)

		networkLivenessUpdater := newNetworkLivenessUpdater(ctx, moduleCtx.ConsClient, consDB)
		go networkLivenessUpdater.Export()

		genesisExporter := newGenesisDepositsExporter(ctx, moduleCtx.ConsClient, consDB)
		go genesisExporter.Export()

		syncCommitteesExporter := NewSyncCommitteesExporter(ctx, moduleCtx.ConsClient, consDB)
		go syncCommitteesExporter.Export()

		syncCommitteesCountExporter := newSyncCommitteesCountExporter(ctx, consDB)
		go syncCommitteesCountExporter.Export()

		if utils.Config.SSVExporter.Enabled {
			go ssvExporter()
		}
		if utils.Config.RocketpoolExporter.Enabled {
			go rocketpoolExporter()
		}

		if utils.Config.Indexer.PubKeyTagsExporter.Enabled {
			pubkeyTagsUpdater := newPubkeyTagsUpdater(ctx, consDB)
			go pubkeyTagsUpdater.Update()
		}

		if utils.Config.MevBoostRelayExporter.Enabled {
			relaysExporter := newRelaysExporter(ctx, consDB)
			go relaysExporter.MEVBoostRelaysExporter()
		}
	}
	// wait until the beacon-node is available
	for {
		head, err := moduleCtx.ConsClient.GetChainHead()
		if err == nil {
			log.Infof("beacon node is available with head slot: %v", head.HeadSlot)
			break
		}
		log.Error(err, "beacon-node seems to be unavailable", 0)
		time.Sleep(time.Second * 10)
	}
	// start subscription modules
	startSubscriptionModules(&moduleCtx, modules)
}

func startSubscriptionModules(moduleCtx *ModuleContext, modules []ModuleInterface) {
	goPool := &errgroup.Group{}
	log.Infof("initialising exporter modules")

	// Initialize modules
	notifyAllModules(goPool, modules, func(module ModuleInterface) error {
		return module.Init()
	})

	err := goPool.Wait()
	if err != nil {
		log.Fatal(err, "error initializing modules", 0)
		return
	}

	eventPool := &errgroup.Group{}
	eventPool.SetLimit(16)

	log.Infof("subscribing to node events")

	// subscribe to node events and notify modules
	events := moduleCtx.CL.GetEvents([]types.EventTopic{
		types.EventHead,
		types.EventFinalizedCheckpoint,
		types.EventChainReorg,
	})
	log.Infof("subscribed to node events")

	for event := range events {
		if event.Error != nil {
			log.Error(event.Error, "error getting event", 0)
			continue
		}

		switch event.Event {
		case types.EventHead:
			res, err := event.Head()
			if err != nil {
				log.Error(err, "error getting head event", 0)
				continue
			}
			log.InfoWithFields(
				log.Fields{"slot": res.Slot, "epoch-transition": res.EpochTransition},
				"notifying exporter modules about new head",
			)
			notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
				return module.OnHead(res)
			})

		case types.EventFinalizedCheckpoint:
			res, err := event.FinalizedCheckpoint()
			if err != nil {
				log.Error(err, "error getting finalized checkpoint event", 0)
				continue
			}
			log.InfoWithFields(log.Fields{"epoch": res.Epoch}, "notifying exporter modules about new finalized checkpoint")
			notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
				return module.OnFinalizedCheckpoint(res)
			})

		case types.EventChainReorg:
			res, err := event.ChainReorg()
			if err != nil {
				log.Error(err, "error getting chain reorg event", 0)
				continue
			}
			log.InfoWithFields(log.Fields{"slot": res.Slot, "depth": res.Depth}, "notifying exporter modules about chain reorg")
			notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
				return module.OnChainReorg(res)
			})
		}
	}
}

func notifyAllModules(goPool *errgroup.Group, modules []ModuleInterface, f func(ModuleInterface) error) {
	for _, module := range modules {
		module := module
		goPool.Go(func() error {
			start := time.Now()
			r := services.StatusReporter.NewStatusReport(module.GetMonitoringEventId(), 5*time.Minute, constants.Default)
			r(constants.Running, nil)
			err := f(module)
			if err != nil {
				log.Error(err, fmt.Sprintf("error in module %s", module.GetName()), 0)
				r(constants.Failure, map[string]string{"error": err.Error()})
				return nil // return never gets caught anywhere? lets not risk a memory leak and instead return nil
			}
			r(constants.Success, map[string]string{"took_raw": fmt.Sprintf("%v", time.Since(start).Milliseconds())})
			return nil
		})
	}
}
func GetModuleContext() (ModuleContext, error) {
	var moduleContext ModuleContext

	cl := consapi.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port)

	spec, err := cl.GetSpec()
	if err != nil {
		log.Fatal(err, "error getting spec", 0)
	}

	config.ClConfig = &spec.Data

	nodeImpl, ok := cl.ClientInt.(*consapi.NodeClient)
	if !ok {
		return ModuleContext{}, errors.New("lighthouse client can only be used with real node impl")
	}

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	clClient, err := rpc.NewLighthouseClient(nodeImpl, chainID)
	if err != nil {
		log.Fatal(err, "error creating lighthouse client", 0)
	}
	moduleContext.CL = cl
	moduleContext.ConsClient = clClient

	return moduleContext, nil
}

type ModuleContext struct {
	CL         consapi.Client
	ConsClient *rpc.LighthouseClient
}
