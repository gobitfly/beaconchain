package log

import (
	"github.com/sirupsen/logrus"
)

// Fatal logs a fatal error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where Fatal is called.
func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

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
