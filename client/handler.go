package main

import (
	"github.com/appcrash/fhs/fhslib"
	"net"
	"strconv"
	"strings"
)

type tunnelData struct {
	Conn       net.Conn
	Domain     string
	Port       int
	Sockserver *Socks5Server
}

type routerHandler struct {
}

func (h *routerHandler) OnNewTunnel(r fhslib.Router, ti *fhslib.TunnelInfo) {
	tid := ti.TunnelId
	info := ti.CustomData.(*tunnelData)
	domain, port, sockserver, conn := info.Domain, info.Port, info.Sockserver, info.Conn

	config := fhslib.GetConfig()
	remote_addr := net.JoinHostPort(domain, strconv.Itoa(port))
	fhs_server_addr := net.JoinHostPort(config.Server.Ip, strconv.Itoa(config.Server.Port))
	fhs_server_conn, conn_err := net.Dial("tcp", fhs_server_addr)

	if conn_err != nil {
		logger.Error("connect to remote server error")
		sockserver.SendRequestReply(conn, hostUnreachable, []byte{0, 0, 0, 0}, 0)
		// TODO: delete tunnel from router
		return
	}

	http_socket := fhslib.NewHttpSocket(tid,
		fhslib.NewRequestEncoder(tid, "key"),
		fhslib.NewResponseDecoder(tid, "key"),
		fhs_server_conn)
	r.NewSocket(http_socket)

	go func() {
		for p := range ti.DataOut {
			conn.Write(p.Data)
		}
		logger.Infof("tunnel(%s) in channel closed", tid)
	}()
	go func() {
		rbuf := make([]byte, 1024*1024)
		for {
			payload_len, err := conn.Read(rbuf)
			if err != nil {
				tmp := make([]byte, payload_len)
				copy(tmp, rbuf[:payload_len])
				p := &fhslib.Packet{fhslib.CmdTunnelData, tid, tmp}
				r.ForwardTunnelPacket(p)
			} else {
				logger.Errorf("remote connection of tunnel(%s) error", tid)
			}
		}
	}()
}

func (h *routerHandler) OnNewSocket(r fhslib.Router, s *fhslib.HttpSocket) {
	conn := s.GetConnection()
	c := make(chan *fhslib.Response)
	go fhslib.GetResponses(conn, c)
	go func() {
		for resp := range c {
			if resp == nil {
				logger.Error("handle connection with first response nil")
				// TODO: notify router delete this socket
				return
			}
			if p := s.Receive(resp); p != nil {
				r.ForwardSocketPacket(p)
			} else {
				logger.Error("http socket receive requests error")
				return
			}
		}

	}()
}

func (h *routerHandler) OnTunnelData(r fhslib.Router, p *fhslib.Packet) {
	sock := r.GetSocket("") // random socket
	sock.Send(p)
}

func (h *routerHandler) OnSocketData(r fhslib.Router, p *fhslib.Packet) {
	switch p.Cmd {
	case fhslib.CmdTunnelInfo:
		go notifySocksConnection(r, p.TunnelId, string(p.Data))
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

func notifySocksConnection(r fhslib.Router, tunnel_id string, addr string) {
	logger.Debugf("client get tunnel bind address: %s", addr)
	tunnel := r.GetTunnel(tunnel_id)
	if tunnel == nil {
		logger.Errorf("no such tunnel(%s) in router", tunnel_id)
		return
	}
	addr_array := strings.Split(addr, ":")
	ip_str, port_str := addr_array[0], addr_array[1]
	logger.Debugf("remote local bind ip:%s, port:%s", ip_str, port_str)
	ip := net.ParseIP(ip_str)
	port, _ := strconv.Atoi(port_str)
	logger.Debugf("send socks reply with success ip: %s, port: %d", ip, port)

	custom_data := tunnel.CustomData.(*tunnelData)
	custom_data.Sockserver.SendRequestReply(custom_data.Conn, successReply, ip, uint16(port))
}
