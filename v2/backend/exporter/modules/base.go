package modules

import "github.com/gobitfly/beaconchain/exporter/clnode"

type ModuleInterface interface {
	Start()
}

type ModuleContext struct {
	CL clnode.Retriever
}
