package modules

import "github.com/gobitfly/beaconchain/exporter/clnode"

type ModuleInterfaceEpoch interface {
	Start(epoch int)
}

type ModuleInterfaceSlot interface {
	Start(slot int64)
}

type ModuleContext struct {
	CL clnode.Retriever
}
