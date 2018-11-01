package main

import (
	"github.com/appcrash/fhs/fhslib"
	"net"
)

type Server struct {
}

func (s *Server) HandleConnection(conn net.Conn) {
	req_decoder := fhslib.NewRequestDecoder("key", conn)
	req_decoder.Prepare()

	req := req_decoder.GetRequest()
	if req == nil {
		logger.Error("handle connection with first request nil")
		return
	}
	// action := req.Action
	path := req.Path

	if path != "/dns" {
		logger.Errorf("server handle connection with first request's path:%s", path)
		return
	}

	domain := req.Data.String()
	logger.Debugf("server get dns request: %s", domain)

}
