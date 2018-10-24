package fhslib

import (
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	formatter := &prefixed.TextFormatter{
		ForceFormatting: true,
		FullTimestamp:   true,
	}
	Log.Formatter = formatter
	hook := lfshook.NewHook(
		lfshook.PathMap{
			logrus.InfoLevel:  "info.log",
			logrus.DebugLevel: "info.log",
			logrus.FatalLevel: "error.log",
			logrus.ErrorLevel: "error.log",
			logrus.WarnLevel:  "error.log",
		},
		formatter,
	)
	Log.Hooks.Add(hook)
}

func SetLogLevel(level string) {
	switch level {
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "debug":
		fallthrough
	default:
		Log.SetLevel(logrus.DebugLevel)
	}
}
