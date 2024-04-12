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
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ModuleInterface interface {
	Init() error
	GetName() string // Used for logging

	// -- !Do not block in this functions for an extended period of time! --

	OnHead(*types.StandardEventHeadResponse) error

	// Note that "StandardFinalizedCheckpointResponse" event contains the current justified epoch, not the finalized one
	// An epoch becomes finalized once the next epoch gets justified
	// Do not assume event.Epoch -1 is finalized by default as it could be that it is not justified
	OnFinalizedCheckpoint(*types.StandardFinalizedCheckpointResponse) error

	OnChainReorg(*types.StandardEventChainReorg) error
}

var Client *rpc.Client

// Start will start the export of data from rpc into the database
func StartAll(context ModuleContext) {
	if !utils.Config.JustV2 {
		go networkLivenessUpdater(context.ConsClient)
		go eth1DepositsExporter()
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

	modules := []ModuleInterface{}

	if !utils.Config.JustV2 {
		modules = append(modules, NewSlotExporter(context))
	} else {
		modules = append(modules, NewDashboardDataModule(context))
	}

	startSubscriptionModules(&context, modules)
}

func startSubscriptionModules(context *ModuleContext, modules []ModuleInterface) {
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
	events := context.CL.GetEvents([]types.EventTopic{
		types.EventHead,
		types.EventFinalizedCheckpoint,
		types.EventChainReorg,
	})

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
			err := f(module)
			if err != nil {
				log.Error(err, fmt.Sprintf("error in module %s", module.GetName()), 0)
			}
			return nil
		})
	}
}

func GetModuleContext() (ModuleContext, error) {
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

	moduleContext := ModuleContext{
		CL:         cl,
		ConsClient: clClient,
	}

	return moduleContext, nil
}

type ModuleContext struct {
	CL         consapi.Client
	ConsClient *rpc.LighthouseClient
}

type ModuleLog struct {
	module ModuleInterface
}

func (m ModuleLog) Info(message string) {
	log.InfoWithFields(log.Fields{"module": m.module.GetName()}, message)
}

func (m ModuleLog) Infof(format string, args ...interface{}) {
	log.InfoWithFields(log.Fields{"module": m.module.GetName()}, fmt.Sprintf(format, args...))
}

func (m ModuleLog) Debug(message string) {
	log.DebugWithFields(log.Fields{"module": m.module.GetName()}, message)
}

func (m ModuleLog) Debugf(format string, args ...interface{}) {
	log.DebugWithFields(log.Fields{"module": m.module.GetName()}, fmt.Sprintf(format, args...))
}

func (m ModuleLog) InfoWithFields(additionalInfos log.Fields, msg string) {
	additionalInfos["module"] = m.module.GetName()
	log.InfoWithFields(additionalInfos, msg)
}

func (m ModuleLog) Error(err error, errorMsg interface{}, callerSkip int, additionalInfos ...log.Fields) {
	additionalInfos = append(additionalInfos, log.Fields{"module": m.module.GetName()})
	log.Error(err, errorMsg, callerSkip, additionalInfos...)
}

func (m ModuleLog) Warn(err error, errorMsg interface{}, callerSkip int, additionalInfos ...log.Fields) {
	additionalInfos = append(additionalInfos, log.Fields{"module": m.module.GetName()})
	log.WarnWithStackTrace(err, errorMsg, callerSkip, additionalInfos...)
}

func (m ModuleLog) Warnf(format string, args ...interface{}) {
	log.WarnWithFields(log.Fields{"module": m.module.GetName()}, fmt.Sprintf(format, args...))
}

func (m ModuleLog) Fatal(err error, errorMsg interface{}, callerSkip int, additionalInfos ...log.Fields) {
	additionalInfos = append(additionalInfos, log.Fields{"module": m.module.GetName()})
	log.Fatal(err, errorMsg, callerSkip, additionalInfos...)
}
