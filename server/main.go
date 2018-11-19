package main

import (
	"fmt"
	"github.com/appcrash/fhs/fhslib"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	config := fhslib.GetConfig()
	logger = fhslib.Log
	fhslib.SetLogLevel(config.Server.Loglevel)
}

func main() {
	config := fhslib.GetConfig()
	handler := Server{}
	addr := fmt.Sprintf("%s:%d", config.Server.Ip, config.Server.Port)
	s := fhslib.HttpServer{addr, &handler}
	s.Listen()
}
