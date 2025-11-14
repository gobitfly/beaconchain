package modules

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

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

func (m ModuleLog) Trace(message string) {
	log.TraceWithFields(log.Fields{"module": m.module.GetName()}, message)
}

func (m ModuleLog) Tracef(format string, args ...interface{}) {
	log.TraceWithFields(log.Fields{"module": m.module.GetName()}, fmt.Sprintf(format, args...))
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
