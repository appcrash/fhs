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
	fhslib.SetLogLevel(config.Client.Loglevel)
}

func main() {
	config := fhslib.GetConfig()
	addr := fmt.Sprintf("%s:%d", config.Client.Ip, config.Client.Port)
	s := Socks5Server{addr}
	s.listen()
}
