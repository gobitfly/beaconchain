package modules

import (
	"github.com/gobitfly/beaconchain/pkg/consapi"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "exporter")

type ModuleInterfaceEpoch interface {
	Start(epoch int)
}

type ModuleInterfaceSlot interface {
	Start() error
}

type ModuleContext struct {
	CL consapi.Retriever
}
