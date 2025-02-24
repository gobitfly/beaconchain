package modules

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
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

type ModuleContext struct {
	CL         consapi.Client
	ConsClient *rpc.LighthouseClient
}

var Client *rpc.Client

var EventPoolLimit = 16

// Start will start the export of data from rpc into the database
func StartAll(context ModuleContext, modules []ModuleInterface, justV2 bool) {
	if !justV2 {
		go networkLivenessUpdater(context.ConsClient)
		go genesisDepositsExporter(context.ConsClient)
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
	// start subscription modules
	err := startSubscriptionModules(&context, modules)
	if err != nil {
		log.Fatal(err, "error initializing modules", 0)
	}
}

func startSubscriptionModules(context *ModuleContext, modules []ModuleInterface) error {
	// Initialize modules
	if err := initializeModules(modules); err != nil {
		return err
	}

	log.Infof("subscribing to node events")

	// subscribe to node events and notify modules
	events := getEvents(context)

	handleEvents(events, modules)

	return nil
}

func initializeModules(modules []ModuleInterface) error {
	if len(modules) == 0 {
		return errors.New("no modules to initialize")
	}

	goPool := &errgroup.Group{}

	log.Infof("initialising exporter modules")

	notifyAllModules(goPool, modules, func(module ModuleInterface) error {
		return module.Init()
	})

	return goPool.Wait()
}

func getEvents(context *ModuleContext) chan *types.EventResponse {
	events := context.CL.GetEvents([]types.EventTopic{
		types.EventHead,
		types.EventFinalizedCheckpoint,
		types.EventChainReorg,
	})
	return events
}

func handleEvents(events chan *types.EventResponse, modules []ModuleInterface) {
	eventPool := &errgroup.Group{}
	eventPool.SetLimit(EventPoolLimit)

	for event := range events {
		err := handleEvent(event, eventPool, modules)
		if err != nil {
			log.Error(err, "error getting event", 0)
		}
	}
}

func handleEvent(event *types.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
	if event.Error != nil {
		return fmt.Errorf("error getting event: %v", event.Error)
	}

	switch event.Event {
	case types.EventHead:
		err := handleHeadEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting head event: %v", err)
		}
	case types.EventFinalizedCheckpoint:
		err := handleFinalizedCheckpointEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting finalized checkpoint event: %v", err)
		}
	case types.EventChainReorg:
		err := handleChainReorgEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting chain reorg event: %v", err)
		}
	}

	return nil
}

func handleHeadEvent(event *types.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
	res, err := event.Head()
	if err != nil {
		return err
	}
	log.InfoWithFields(
		log.Fields{"slot": res.Slot, "epoch-transition": res.EpochTransition},
		"notifying exporter modules about new head",
	)
	notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
		return module.OnHead(res)
	})

	return nil
}

func handleFinalizedCheckpointEvent(event *types.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
	res, err := event.FinalizedCheckpoint()
	if err != nil {
		return err
	}
	log.InfoWithFields(log.Fields{"epoch": res.Epoch}, "notifying exporter modules about new finalized checkpoint")
	notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
		return module.OnFinalizedCheckpoint(res)
	})

	return nil
}

func handleChainReorgEvent(event *types.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
	res, err := event.ChainReorg()
	if err != nil {
		return err
	}
	log.InfoWithFields(log.Fields{"slot": res.Slot, "depth": res.Depth}, "notifying exporter modules about chain reorg")
	notifyAllModules(eventPool, modules, func(module ModuleInterface) error {
		return module.OnChainReorg(res)
	})

	return nil
}

func notifyAllModules(goPool *errgroup.Group, modules []ModuleInterface, f func(ModuleInterface) error) {
	deploymentType := utils.Config.DeploymentType
	for _, module := range modules {
		module := module
		goPool.Go(func() error {
			start := time.Now()
			statusReport := services.NewStatusReport(module.GetMonitoringEventId(), 5*time.Minute, constants.Default, deploymentType)
			statusReport(constants.Running, nil)
			err := f(module)
			if err != nil {
				log.Error(err, fmt.Sprintf("error in module %s", module.GetName()), 0)
				statusReport(constants.Failure, map[string]string{"error": err.Error()})
				return nil // return never gets caught anywhere? lets not risk a memory leak and instead return nil
			}
			statusReport(constants.Success, map[string]string{"took_raw": fmt.Sprintf("%v", time.Since(start).Milliseconds())})
			return nil
		})
	}
}

func GetModuleContext() (ModuleContext, error) {
	var moduleContext ModuleContext
	clientCreator := &consapi.DefaultClientCreator{}

	cl, err := createClient(clientCreator)
	if err != nil {
		log.Fatal(err, "error getting spec", 0)
	}

	clClient, err := createLighthouseClient(cl)
	if err != nil {
		log.Fatal(err, "error creating lighthouse client", 0)
	}
	moduleContext.CL = cl
	moduleContext.ConsClient = clClient

	return moduleContext, nil
}

func createClient(clientCreator consapi.ClientCreator) (consapi.Client, error) {
	cl := clientCreator.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port)
	spec, err := cl.GetSpec()
	if err != nil {
		return nil, err
	}

	config.ClConfig = &spec.Data

	return cl, nil
}

func createLighthouseClient(cl consapi.Client) (*rpc.LighthouseClient, error) {
	nodeImpl, ok := cl.(*consapi.NodeClient)
	if !ok {
		return nil, errors.New("lighthouse client can only be used with real node impl")
	}
	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	return rpc.NewLighthouseClient(nodeImpl, chainID)
}
