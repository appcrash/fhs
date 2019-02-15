package fhslib

import (
	"net"
)

type HttpSocket struct {
	socket_id string
	encoder   Encoder
	decoder   Decoder
	conn      net.Conn
}

type SocketErrorInfo struct {
	socket_id string
	err       error
}

const (
	CmdTunnelInfo = iota
	CmdTunnelData
)

type Packet struct {
	Cmd      int
	TunnelId string
	Data     []byte
}

type PacketChannel chan *Packet
type TunnelErrorChannel chan *TunnelErrorInfo

type TunnelErrorInfo struct {
	TunnelId string
	Error    error
}

type TunnelInfo struct {
	TunnelId string
	DataIn   PacketChannel
	DataOut  PacketChannel
	ErrorIn  TunnelErrorChannel
	ErrorOut TunnelErrorChannel
}

func NewHttpSocket(sid string, encoder Encoder, decoder Decoder, conn net.Conn) *HttpSocket {
	return &HttpSocket{sid, encoder, decoder, conn}
}

func (s *HttpSocket) Send(p *Packet) {
	data := s.encoder.Encode(p)
	s.conn.Write(data)
}

func (s *HttpSocket) Receive(data interface{}) *Packet {
	var p *Packet
	switch d := data.(type) {
	case *Request:
		p = s.decoder.Decode(d)
	case *Response:
		p = s.decoder.Decode(d)
	default:
		Log.Error("httpsocket receive with unknonw data type")
	}

	return p
}

func (s *HttpSocket) GetConnection() net.Conn {
	return s.conn
}

func NewTunnelInfo(tid string) *TunnelInfo {
	return &TunnelInfo{
		tid,
		make(PacketChannel), make(PacketChannel),
		make(TunnelErrorChannel), make(TunnelErrorChannel),
	}
}
