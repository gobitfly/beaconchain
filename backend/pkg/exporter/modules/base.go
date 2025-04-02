package modules

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/config"
	"github.com/gobitfly/beaconchain/pkg/commons/types"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ModuleInterface interface {
	Init() error
	GetName() string // Used for logging
	GetMonitoringEventId() constants.Event

	OnHead(*constypes.StandardEventHeadResponse) error // !Do not block in this functions for an extended period of time!

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	OnFinalizedCheckpoint(*constypes.StandardFinalizedCheckpointResponse) error // !Do not block in this functions for an extended period of time!

	OnChainReorg(*constypes.StandardEventChainReorg) error // !Do not block in this functions for an extended period of time!
}

type ConsClient interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
	GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error)
	GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error)
	GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error)
	GetBlockBySlot(slot uint64) (*types.Block, error)
	GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error)
	GetValidatorQueue() (*types.ValidatorQueue, error)
}

type ModuleContext struct {
	CL         consapi.ClientInt
	ConsClient ConsClient
}

var EventPoolLimit = 16

// Start will start the export of data from rpc into the database
func StartAll(moduleCtx ModuleContext, modules []ModuleInterface, justV2 bool) error {
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
	err := startSubscriptionModules(&moduleCtx, modules)
	if err != nil {
		log.Error(err, "error initializing modules: %v", 0)
		return err
	}

	return nil
}

func startSubscriptionModules(context *ModuleContext, modules []ModuleInterface) error {
	// Initialize modules
	if err := initializeModules(modules); err != nil {
		return err
	}

	log.Infof("subscribing to node events")

	// subscribe to node events and notify modules
	events := getEvents(context)

	log.Infof("subscribed to node events")

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

func getEvents(context *ModuleContext) chan *constypes.EventResponse {
	events := context.CL.GetEvents([]constypes.EventTopic{
		constypes.EventHead,
		constypes.EventFinalizedCheckpoint,
		constypes.EventChainReorg,
	})
	return events
}

func handleEvents(events chan *constypes.EventResponse, modules []ModuleInterface) {
	eventPool := &errgroup.Group{}
	eventPool.SetLimit(EventPoolLimit)

	for event := range events {
		err := handleEvent(event, eventPool, modules)
		if err != nil {
			log.Error(err, "error getting event", 0)
		}
	}
}

func handleEvent(event *constypes.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
	if event.Error != nil {
		return fmt.Errorf("error getting event: %v", event.Error)
	}

	switch event.Event {
	case constypes.EventHead:
		err := handleHeadEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting head event: %v", err)
		}
	case constypes.EventFinalizedCheckpoint:
		err := handleFinalizedCheckpointEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting finalized checkpoint event: %v", err)
		}
	case constypes.EventChainReorg:
		err := handleChainReorgEvent(event, eventPool, modules)
		if err != nil {
			return fmt.Errorf("error getting chain reorg event: %v", err)
		}
	}

	return nil
}

func handleHeadEvent(event *constypes.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
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

func handleFinalizedCheckpointEvent(event *constypes.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
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

func handleChainReorgEvent(event *constypes.EventResponse, eventPool *errgroup.Group, modules []ModuleInterface) error {
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
	for _, module := range modules {
		module := module
		goPool.Go(func() error {
			start := time.Now()
			statusReport := services.StatusReporter.NewStatusReport(module.GetMonitoringEventId(), 5*time.Minute, constants.Default)
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

	cl := consapi.NewClient("http://" + utils.Config.Indexer.Node.Host + ":" + utils.Config.Indexer.Node.Port)
	err := getClientSpec(cl)
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

func getClientSpec(client consapi.ClientInt) error {
	spec, err := client.GetSpec()
	if err != nil {
		return err
	}
	config.ClConfig = &spec.Data

	return nil
}

func createLighthouseClient(cl consapi.ClientInt) (*rpc.LighthouseClient, error) {
	nodeImpl, ok := cl.(*consapi.NodeClient)
	if !ok {
		return nil, errors.New("lighthouse client can only be used with real node impl")
	}
	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)

	return rpc.NewLighthouseClient(nodeImpl, chainID)
}
