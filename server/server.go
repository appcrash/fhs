package main

import (
	"fmt"
	"github.com/appcrash/fhs/fhslib"
	"net"
)

type Server struct {
}

func (s *Server) HandleConnection(conn net.Conn) {
	track_id := fhslib.GenerateId()
	req_decoder := fhslib.NewRequestDecoder(track_id, "key", conn)
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
	remote_conn, err := net.Dial("tcp", domain)
	if err != nil {
		logger.Errorf("server connect to domain:  %s  ,error", domain)
	}

	addr, ok := remote_conn.LocalAddr().(*net.TCPAddr)
	if ok == false {
		logger.Error("server remote conn local address convert error")
		return
	}
	logger.Debugf("server local bind addr is %s", addr.String())
	ip, port := addr.IP, addr.Port
	resp_encoder := fhslib.NewResponseEncoder(track_id, "key", remote_conn)
	resp_encoder.WriteResolveResponse(conn, fmt.Sprintf("%s:%d", ip, port))

	// pipe bidirection
	c := make(chan string, 2)
	go resp_encoder.PipeTo(conn, c)
	go req_decoder.PipeTo(remote_conn, c)

	for i := 0; i < 2; i++ {
		name := <-c
		logger.Debugf("%s done", name)
	}

}
