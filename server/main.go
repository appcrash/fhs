package main

import (
	"github.com/appcrash/fhs/fhslib"
	"github.com/sirupsen/logrus"
	"log"
)

var logger *logrus.Logger

func init() {
	config, err := fhslib.GetConfig()
	if err != nil {
		log.Fatalf("config error %v", err)
	}
	logger = fhslib.Log
	fhslib.SetLogLevel(config.Server.Loglevel)
}

func main() {
	s := fhslib.HttpServer{"127.0.0.1:8080"}
	s.Listen()
}
