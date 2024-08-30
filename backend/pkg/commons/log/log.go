package log

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/sirupsen/logrus"
)

// Fatal logs a fatal error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where Fatal is called.
func Fatal(err error, errorMsg interface{}, callerSkip int, additionalInfos ...Fields) {
	logErrorInfo(err, callerSkip, additionalInfos...).Fatal(errorMsg)
}

// Error logs an error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where Error is called.
func Error(err error, errorMsg interface{}, callerSkip int, additionalInfos ...Fields) {
	logErrorInfo(err, callerSkip, additionalInfos...).Error(errorMsg)
}

func WarnWithStackTrace(err error, errorMsg interface{}, callerSkip int, additionalInfos ...Fields) {
	logErrorInfo(err, callerSkip, additionalInfos...).Warn(errorMsg)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func InfoWithFields(additionalInfos Fields, msg string) {
	logFields := logrus.NewEntry(logrus.New())
	for name, info := range additionalInfos {
		logFields = logFields.WithField(name, info)
	}

	logFields.Info(msg)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func WarnWithFields(additionalInfos Fields, msg string) {
	logFields := logrus.NewEntry(logrus.New())
	for name, info := range additionalInfos {
		logFields = logFields.WithField(name, info)
	}

	logFields.Warn(msg)
}

func Tracef(format string, args ...interface{}) {
	logrus.Tracef(format, args...)
}

func TraceWithFields(additionalInfos Fields, msg string) {
	logFields := logrus.NewEntry(logrus.New())
	for name, info := range additionalInfos {
		logFields = logFields.WithField(name, info)
	}

	logFields.Trace(msg)
}

func DebugWithFields(additionalInfos Fields, msg string) {
	logFields := logrus.NewEntry(logrus.New())
	for name, info := range additionalInfos {
		logFields = logFields.WithField(name, info)
	}

	logFields.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func logErrorInfo(err error, callerSkip int, additionalInfos ...Fields) *logrus.Entry {
	logFields := logrus.NewEntry(logrus.New())

	metricName := "unknown"
	if err != nil {
		metricName = err.Error()
	}
	pc, fullFilePath, line, ok := runtime.Caller(callerSkip + 2)
	if ok {
		logFields = logFields.WithFields(logrus.Fields{
			"_file":     filepath.Base(fullFilePath),
			"_function": runtime.FuncForPC(pc).Name(),
			"_line":     line,
		})
		metricName = fmt.Sprintf("%s:%d", fullFilePath, line)
	} else {
		logFields = logFields.WithField("runtime", "Callstack cannot be read")
	}
	if len(metricName) > 30 {
		metricName = metricName[len(metricName)-30:]
	}
	metrics.Errors.WithLabelValues(metricName).Inc()

	errColl := []string{}
	for {
		errColl = append(errColl, fmt.Sprint(err))
		nextErr := errors.Unwrap(err)
		if nextErr != nil {
			err = nextErr
		} else {
			break
		}
	}

	errMarkSign := "~"
	for idx := 0; idx < (len(errColl) - 1); idx++ {
		errInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx, errMarkSign)
		nextErrInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx+1, errMarkSign)
		if idx == (len(errColl) - 2) {
			nextErrInfoText = fmt.Sprintf("%serror%s", errMarkSign, errMarkSign)
		}

		// Replace the last occurrence of the next error in the current error
		lastIdx := strings.LastIndex(errColl[idx], errColl[idx+1])
		if lastIdx != -1 {
			errColl[idx] = errColl[idx][:lastIdx] + nextErrInfoText + errColl[idx][lastIdx+len(errColl[idx+1]):]
		}

		errInfoText = strings.ReplaceAll(errInfoText, errMarkSign, "")
		logFields = logFields.WithField(errInfoText, errColl[idx])
	}

	if err != nil {
		logFields = logFields.WithField("errType", fmt.Sprintf("%T", err)).WithError(err)
	}

	for _, infoMap := range additionalInfos {
		for name, info := range infoMap {
			logFields = logFields.WithField(name, info)
		}
	}

	return logFields
}

type Fields map[string]interface{}
