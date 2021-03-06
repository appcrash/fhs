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

	router := fhslib.NewRouter(&routerHandler{})
	go func() {
		router.Loop()
	}()
	s := Socks5Server{addr, router}
	s.listen()
}
