package fhslib

import (
	// "bytes"
	"math/rand"
	// "net"
	// "strconv"
)

type DefaultRouter struct {
	socks_map                  map[string]*HttpSocket // socket id => out http sock
	tunnel_map                 map[string]*TunnelInfo // tunnel id => out tunnel
	channel_tunnel2sock_packet PacketChannel
	channel_sock2tunnel_packet PacketChannel
	channel_tunnel_error       chan *TunnelErrorInfo
	channel_socket_error       chan *SocketErrorInfo
	channel_new_tunnel         chan *TunnelInfo
	channel_new_socket         chan *HttpSocket
	handler                    RouterHandler
}

type Router interface {
	NewSocket(*HttpSocket)
	NewTunnel(*TunnelInfo)
	GetTunnel(string) *TunnelInfo
	GetSocket(string) *HttpSocket
	ForwardTunnelPacket(*Packet)
	ForwardSocketPacket(*Packet)
	Loop()
}

type RouterHandler interface {
	OnNewTunnel(Router, *TunnelInfo)
	OnNewSocket(Router, *HttpSocket)
	OnTunnelData(Router, *Packet)
	OnSocketData(Router, *Packet)
}

// func newClientHttpSocket(socket_id string) (*HttpSocket, error) {
// 	config := GetConfig()
// 	fhs_server_addr := net.JoinHostPort(config.Server.Ip, strconv.Itoa(config.Server.Port))
// 	if fhs_server_conn, err := net.Dial("tcp", fhs_server_addr); err != nil {
// 		return nil, err
// 	} else {
// 		key := config.Common.Password
// 		encoder := NewRequestEncoder(socket_id, key, fhs_server_conn)
// 		decoder := NewResponseDecoder(socket_id, key, fhs_server_conn)
// 		return &HttpSocket{socket_id, &encoder, &decoder}, nil
// 	}
// }

func NewRouter(h RouterHandler) Router {
	smap := make(map[string]*HttpSocket)
	tmap := make(map[string]*TunnelInfo)
	csend, crecv := make(PacketChannel, 10), make(PacketChannel, 10)
	terror, serror := make(chan *TunnelErrorInfo), make(chan *SocketErrorInfo)
	nt, ns := make(chan *TunnelInfo), make(chan *HttpSocket)
	return &DefaultRouter{
		smap, tmap,
		csend, crecv,
		terror, serror,
		nt, ns,
		h,
	}
}

func (r *DefaultRouter) setupTunnel(t *TunnelInfo) {
	tid, data_in, error_in := t.TunnelId, t.DataIn, t.ErrorIn
	if tid == "" {
		for {
			tid = GenerateId()
			if _, ok := r.tunnel_map[tid]; !ok {
				break
			}
		}
	}
	r.tunnel_map[tid] = t
	r.handler.OnNewTunnel(r, t)

	go func() {
		for p := range data_in { // library user send data packet to router, combine them to bus
			if p == nil {
				Log.Infof("router: tunnel(%s) data in tunnel closed", tid)
				break
			}
			r.channel_tunnel2sock_packet <- p
		}
	}()
	go func() {
		for p := range error_in { // library user send tunnel error info to router, combine them to bus
			if p == nil {
				Log.Infof("router: tunnel(%s) error tunnel closed", p.TunnelId)
				break
			}
			r.channel_tunnel_error <- p
		}
	}()
}

func (r *DefaultRouter) setupSocket(s *HttpSocket) {
	sid := s.socket_id
	if sid == "" {
		for {
			socket_id := GenerateId()
			if _, ok := r.socks_map[socket_id]; !ok { // avoid name confliction
				break
			}
		}
	}
	Log.Debugf("router: add http socket with socket id:%s", sid)
	r.socks_map[sid] = s
	r.handler.OnNewSocket(r, s)
}

func (r *DefaultRouter) NewSocket(s *HttpSocket) {
	r.channel_new_socket <- s
}

func (r *DefaultRouter) NewTunnel(info *TunnelInfo) {
	r.channel_new_tunnel <- info
}

func (r *DefaultRouter) GetSocket(id string) *HttpSocket {
	if s, ok := r.socks_map[id]; ok {
		return s
	} else {
		return nil
	}
}

func (r *DefaultRouter) GetTunnel(id string) *TunnelInfo {
	if t, ok := r.tunnel_map[id]; ok {
		return t
	} else {
		return nil
	}
}

func (r *DefaultRouter) ForwardTunnelPacket(p *Packet) {
	r.channel_tunnel2sock_packet <- p
}

func (r *DefaultRouter) ForwardSocketPacket(p *Packet) {
	r.channel_sock2tunnel_packet <- p
}

func (r *DefaultRouter) onTunnelData(p *Packet) {
	var sock *HttpSocket
	n := len(r.socks_map)
	n = rand.Intn(n)
	for _, sock = range r.socks_map {
		n--
		if n == 0 {
			break
		}
	}
	sock.Send(p)
}

func (r *DefaultRouter) onTunnelError(e *TunnelErrorInfo) {
	tid := e.TunnelId
	if tunnel, ok := r.tunnel_map[tid]; !ok {
		Log.Debugf("(%s) tunnel error but it has been deleted, %s", tid, e)
	} else {
		Log.Debugf("(%s) tunnel closed due to %s", tid, e.Error)
		close(tunnel.DataOut)
		close(tunnel.ErrorOut)
		delete(r.tunnel_map, tid)
	}
}

func (r *DefaultRouter) onSocketData(p *Packet) {
	tid := p.TunnelId
	if tunnel, ok := r.tunnel_map[tid]; ok {
		tunnel.DataOut <- p
	} else {
		Log.Debugf("discard tunnel:(%s) data as it closed", tid)
	}
}

func (r *DefaultRouter) onSocketError(e *SocketErrorInfo) {
	sid := e.socket_id
	if _, ok := r.socks_map[sid]; ok {
		Log.Debugf("(%s) http socket error with: %s", sid, e.err)
		delete(r.socks_map, sid)
	}
}

func (r *DefaultRouter) Loop() {
	for {
		select {
		case p := <-r.channel_tunnel2sock_packet:
			// r.onTunnelData(p)
			r.handler.OnTunnelData(r, p)
		case e := <-r.channel_tunnel_error:
			r.onTunnelError(e)
		case p := <-r.channel_sock2tunnel_packet:
			// r.onSocketData(p)
			r.handler.OnSocketData(r, p)
		case e := <-r.channel_socket_error:
			r.onSocketError(e)
		case i := <-r.channel_new_tunnel:
			r.setupTunnel(i)
		case i := <-r.channel_new_socket:
			r.setupSocket(i)
		}
	}

}
