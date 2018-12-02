package fhslib

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
)

const (
	dtTunnelInfo = iota
	dtTunnelData
)

type Packet struct {
	Cmd      int
	TunnelId string
	Data     *bytes.Buffer
}

type ErrorInfo struct {
	TunnelId string
	Error    error
}

type ClientHttpSocket struct {
	track_id string
	encoder  *RequestEncoder
	decoder  *ResponseDecoder
}

type ClientConnectionTracker struct {
	socks_map              map[string]*ClientHttpSocket // socket tracck id => http socks
	tunnel_map             map[string]net.Conn          // tunnel id => local client connection
	channel_sendbuf_packet chan *Packet
	channel_recvbuf_packet chan *Packet
	channel_error          chan *ErrorInfo
}

func newClientHttpSocket(track_id string) (*ClientHttpSocket, error) {
	config := GetConfig()
	fhs_server_addr := net.JoinHostPort(config.Server.Ip, strconv.Itoa(config.Server.Port))
	if fhs_server_conn, err := net.Dial("tcp", fhs_server_addr); err != nil {
		return nil, err
	} else {
		key := config.Common.Password
		encoder := NewRequestEncoder(track_id, key, fhs_server_conn)
		decoder := NewResponseDecoder(track_id, key, fhs_server_conn)
		return &ClientHttpSocket{track_id, &encoder, &decoder}, nil
	}
}

func NewClientConnecionTracker() ClientConnectionTracker {
	smap := make(map[string]*ClientHttpSocket)
	tmap := make(map[string]net.Conn)
	csend, crecv := make(chan *Packet, 10), make(chan *Packet, 10)
	cerror := make(chan *ErrorInfo)
	return ClientConnectionTracker{smap, tmap, csend, crecv, cerror}
}

func (tracker *ClientConnectionTracker) AddHttpSocket() {
	var track_id string
	if len(tracker.socks_map) <= GetConfig().Common.MaxConnection {
		for {
			track_id = GenerateId()
			if _, ok := tracker.socks_map[track_id]; !ok { // avoid name confliction
				break
			}
		}

		if s, err := newClientHttpSocket(track_id); err != nil {
			Log.Errorf("connection tracker: can not create more http socket: %s", err)
		} else {
			Log.Debugf("connection tracker: add http socket with track id:%s", track_id)
			tracker.socks_map[track_id] = s
		}
	}
}

func (tracker *ClientConnectionTracker) Send(tunnel_id string, data *bytes.Buffer) {
	p := Packet{dtTunnelData, tunnel_id, data}
	tracker.channel_sendbuf_packet <- &p
}

func (t *ClientConnectionTracker) Loop() {
	var sock *ClientHttpSocket
	for {
		select {
		case p := <-t.channel_sendbuf_packet:
			n := len(t.socks_map)
			n = rand.Intn(n)
			for _, sock = range t.socks_map {
				n--
				if n == 0 {
					break
				}
			}
			if err := sock.encoder.Send(p); err != nil {
				Log.Errorf("send packet error:%s", err)
			}
		case p := <-t.channel_recvbuf_packet:
			if p == nil {
				continue
			}
			tid := p.TunnelId
			if conn, ok := t.tunnel_map[tid]; ok {
				data := p.Data.Bytes()
				if _, err := conn.Write(data); err != nil {
					Log.Errorf("write data back to tunnel:%s error with:%s ", tid, err)
				}
			} else {
				Log.Debugf("discard tunnel:(%s) data as it closed", tid)
			}
		case e := <-t.channel_error:
			if e == nil {
				Log.Error("error channel get nil")
				continue
			}
			id := e.TunnelId
			Log.Debugf("(%s) tunnel closed due to %s", id, e.Error)
			delete(t.tunnel_map, id)
		}
	}

}
