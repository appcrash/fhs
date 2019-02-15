package main

import (
	// "fmt"
	"github.com/appcrash/fhs/fhslib"
	"net"
)

type Server struct {
	router fhslib.Router
}

type routerHandler struct {
}

func (h *routerHandler) OnNewTunnel(r fhslib.Router, ti *fhslib.TunnelInfo) {

}

func (h *routerHandler) OnNewSocket(r Router, s *HttpSocket) {
	conn := s.GetConnection()
	c := make(*fhslib.Request)
	go fhslib.GetRequests(conn, c)
	go func() {
		for req := range c {
			if req == nil {
				logger.Error("handle connection with first request nil")
				// TODO
				return
			}
			if p := s.Receive(req); p != nil {
				r.ForwardSocketPacket(p)
			} else {
				logger.Error("http socket receive requests error")
				return
			}
		}

	}()
}

func (h *routerHandler) OnTunnelData(r fhslib.Router, p *fhslib.Packet) {
	sock := r.GetSocket() // random socket
	sock.Send(p)
}

func (h *routerHandler) OnSocketData(r fhslib.Router, p *fhslib.Packet) {
	switch p.Cmd {
	case fhslib.CmdTunnelInfo:
		go connectToRemote(r, p.TunnelId, string(p.Data))
	case fhslib.CmdTunnelData:
		if ti := r.GetTunnel(p.TunnelId); ti != nil {
			// send to tunnel channel is safe as channel's close/removal happens only in the same coroutine
			// so if ti != nil, the channel is alive
			ti.DataIn <- p
		} else {
			logger.Errorf("server on socket data can not find channel(%s)", p.TunnelId)
		}
	}
}

func connectToRemote(r fhslib.Router, tunnel_id string, domain string) {
	logger.Debugf("server get dns request: %s", domain)
	remote_conn, err := net.Dial("tcp", domain)
	if err != nil {
		logger.Errorf("server connect to domain:  %s  ,error", domain)
		return
	}
	addr, ok := remote_conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		logger.Error("server remote conn local address convert error")
		return
	}

	tunnel := NewTunnelInfo(tunnel_id)
	go func() {
		for p := range tunnel.DataIn {
			// received from client, forward it to remote connection
			remote_conn.Write(p.Data)
		}
		logger.Infof("tunnel(%s) in channel closed", tunnel_id)
	}()
	go func() {
		rbuf := make([]byte, 1024*1024)
		for {
			payload_len, err := remote_conn.Read(rbuf)
			if err != nil {
				tmp := make([]byte, payload_len)
				copy(tmp, rbuf[:payload_len])
				p := &Packet(fhslib.CmdTunnelData, tunnel_id, tmp)
				r.ForwardTunnelPacket(p)
			} else {
				logger.Errorf("remote connection of tunnel(%s) error", tunnel_id)
			}
		}

	}()

	// when everything ready, notify router new tunnel created
	r.NewTunnel(tunnel)

	// TODO: this is a bug, move it to other place
	p := &Packet{fhslib.CmdTunnelInfo, tunnel_id, []byte(addr)} // this packet is used as first packet to client, with tunnel info
	r.ForwardTunnelPacket(p)
}

func (s *Server) HandleConnection(conn net.Conn) {
	socket_id := fhslib.GenerateId()
	http_socket := NewHttpSocket(socket_id,
		NewResponseEncoder(socket_id, "key"),
		NewRequestDecoder(socket_id, "key"),
		conn)
	s.router.NewSocket(http_socket)

	// req_decoder := fhslib.NewRequestDecoder(track_id, "key", conn)
	// req_decoder.Prepare()

	// req := req_decoder.GetRequest()
	// if req == nil {
	// 	logger.Error("handle connection with first request nil")
	// 	return
	// }
	// // action := req.Action
	// path := req.Path

	// if path != "/dns" {
	// 	logger.Errorf("server handle connection with first request's path:%s", path)
	// 	return
	// }

	// domain := req.Data.String()
	// logger.Debugf("server get dns request: %s", domain)
	// remote_conn, err := net.Dial("tcp", domain)
	// if err != nil {
	// 	logger.Errorf("server connect to domain:  %s  ,error", domain)
	// 	return
	// }

	// addr, ok := remote_conn.LocalAddr().(*net.TCPAddr)
	// if ok == false {
	// 	logger.Error("server remote conn local address convert error")
	// 	return
	// }
	// logger.Debugf("server local bind addr is %s", addr.String())
	// ip, port := addr.IP, addr.Port
	// resp_encoder := fhslib.NewResponseEncoder(track_id, "key", remote_conn)
	// resp_encoder.WriteResolveResponse(conn, fmt.Sprintf("%s:%d", ip, port))

	// // pipe bidirection
	// c := make(chan string, 2)
	// go resp_encoder.PipeTo(conn, c)
	// go req_decoder.PipeTo(remote_conn, c)

	// for i := 0; i < 2; i++ {
	// 	name := <-c
	// 	logger.Debugf("%s done", name)
	// }

}
