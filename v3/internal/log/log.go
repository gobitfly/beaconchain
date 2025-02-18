package log

import (
	"github.com/sirupsen/logrus"
)

// TODO: Lots of stuff removed here, for simplicity, from the original implementation of this pkg in github.com/gobitfly/beaconchain/log. Consider re-merging that stuff back in when we get to a more reliable state.

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func InfoWithFields(additionalInfos Fields, msg string) {
	logrus.WithFields(additionalInfos).Info(msg)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func WarnWithFields(additionalInfos Fields, msg string) {
	logrus.WithFields(additionalInfos).Warn(msg)
}

func Tracef(format string, args ...interface{}) {
	logrus.Tracef(format, args...)
}

func TraceWithFields(additionalInfos Fields, msg string) {
	logrus.WithFields(additionalInfos).Trace(msg)
}

func DebugWithFields(additionalInfos Fields, msg string) {
	logrus.WithFields(additionalInfos).Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

type Fields = logrus.Fields
