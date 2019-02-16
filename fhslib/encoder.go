package fhslib

import (
	"bytes"
	// "encoding/binary"
	// "crypto/md5"
	"fmt"
	"github.com/golang/snappy"
	// "io"
	"strconv"
)

// read from conn, encode data, write to conn
type RequestEncoder struct {
	id  string
	key string
}

type ResponseEncoder struct {
	id  string
	key string
}

type Encoder interface {
	Encode(*Packet) []byte
}

const max_segment = 1024 * 32 // max request data

func NewRequestEncoder(id string, key string) *RequestEncoder {
	return &RequestEncoder{id, key}
}

func (encoder *RequestEncoder) encodeData(tunnel_id string, data []byte) []byte {
	wbuf := bytes.Buffer{}
	payload_len := len(data)
	path := fmt.Sprintf("/img/%s", tunnel_id)
	AddAction(&wbuf, "post", path)
	AddHeader(&wbuf, "content-type", "text/plain")
	AddHeader(&wbuf, "host", "from-req-encoder.com")
	AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
	AddDelimiter(&wbuf)
	AddData(&wbuf, data)

	return wbuf.Bytes()
}

func (encoder *RequestEncoder) encodeNewTunnel(tunnel_id string, domain string) []byte {
	wbuf := bytes.Buffer{}
	payload_len := len(domain)
	path := fmt.Sprintf("/new/%s", tunnel_id)
	AddAction(&wbuf, "post", path)
	AddHeader(&wbuf, "content-type", "text/plain")
	AddHeader(&wbuf, "host", "from-req-encoder.com")
	AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
	AddDelimiter(&wbuf)
	AddData(&wbuf, []byte(domain))

	return wbuf.Bytes()
}

func (encoder *RequestEncoder) Encode(packet *Packet) (data []byte) {
	tid := packet.TunnelId
	switch packet.Cmd {
	case CmdTunnelInfo:
		data = encoder.encodeNewTunnel(tid, string(packet.Data))
	case CmdTunnelData:
		data = encoder.encodeData(tid, packet.Data)
	}
	return
}

// func (encoder *RequestEncoder) WriteTo(writer io.Writer) (written int64, err error) {
// 	var n, payload_len int
// 	rbuf := make([]byte, max_segment)
// 	wbuf := bytes.Buffer{}

// 	for {
// 		payload_len, err = encoder.reader.Read(rbuf)
// 		if err != nil {
// 			written += int64(payload_len)
// 			if err != io.EOF {
// 				Log.Errorf("(%s)encoder read with error %s", encoder.id, err)
// 			} else {
// 				Log.Debugf("(%s)encoder read eof", encoder.id)
// 			}

// 			return
// 		}

// 		wbuf.Reset()
// 		AddAction(&wbuf, "post", "/img")
// 		AddHeader(&wbuf, "content-type", "text/plain")
// 		AddHeader(&wbuf, "host", "from-req-encoder.com")
// 		AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
// 		AddDelimiter(&wbuf)
// 		AddData(&wbuf, rbuf[:payload_len])

// 		n, err = writer.Write(wbuf.Bytes())
// 		if err != nil {
// 			Log.Errorf("(%s)request encoder write error with :%s", encoder.id, err)
// 			return
// 		}
// 		Log.Debugf("(%s)request encoder write %d bytes payload", encoder.id, payload_len)
// 		written += int64(n)

// 	}
// }

//-----------------------response encoder--------------------------------
func NewResponseEncoder(id string, key string) *ResponseEncoder {

	return &ResponseEncoder{id, key}
}

func (encoder *ResponseEncoder) encodeData(tunnel_id string, data []byte) []byte {
	wbuf := bytes.Buffer{}
	data = snappy.Encode(nil, data)
	payload_len := len(data)
	cookie_str := fmt.Sprintf("data=%s", tunnel_id)
	AddState(&wbuf, 200)
	AddHeader(&wbuf, "content-type", "text/plain")
	AddHeader(&wbuf, "host", "from-resp-encoder.com")
	AddHeader(&wbuf, "cookie", cookie_str)
	AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
	AddDelimiter(&wbuf)
	AddData(&wbuf, data)

	return wbuf.Bytes()
}

func (encoder *ResponseEncoder) encodeNewTunnel(tunnel_id string, domain string) []byte {
	wbuf := bytes.Buffer{}
	data := snappy.Encode(nil, []byte(domain))
	payload_len := len(data)
	cookie_str := fmt.Sprintf("domain=%s", tunnel_id)
	AddState(&wbuf, 200)
	AddHeader(&wbuf, "content-type", "text/plain")
	AddHeader(&wbuf, "host", "from-req-encoder.com")
	AddHeader(&wbuf, "cookie", cookie_str)
	AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
	AddDelimiter(&wbuf)
	AddData(&wbuf, data)

	return wbuf.Bytes()
}

func (encoder *ResponseEncoder) Encode(packet *Packet) (data []byte) {
	tid := packet.TunnelId
	switch packet.Cmd {
	case CmdTunnelInfo:
		data = encoder.encodeNewTunnel(tid, string(packet.Data))
	case CmdTunnelData:
		data = encoder.encodeData(tid, packet.Data)
	}
	return
}

// func (encoder *ResponseEncoder) WriteTo(writer io.Writer) (written int64, err error) {
// 	var n int
// 	rbuf := make([]byte, max_segment)
// 	wbuf := bytes.Buffer{}

// 	for {
// 		n, err = encoder.reader.Read(rbuf)
// 		if err != nil {
// 			written += int64(n)
// 			if err != io.EOF {
// 				Log.Errorf("response encoder read with error %s", err)
// 			} else {
// 				Log.Debug("response encoder read eof")
// 			}

// 			return
// 		}

// 		// if n > 0 {
// 		// 	Log.Infof("(%s)response encoder read content md5sum:%x", encoder.id, md5.Sum(rbuf[:n]))
// 		// }

// 		Log.Debugf("(%s)response encoder read %d bytes from reader", encoder.id, n)
// 		encoded_buf := snappy.Encode(nil, rbuf[:n])
// 		n = len(encoded_buf)
// 		wbuf.Reset()
// 		AddState(&wbuf, 200)
// 		AddHeader(&wbuf, "content-type", "text/plain")
// 		AddHeader(&wbuf, "host", "from-resp-encoder.com")
// 		AddHeader(&wbuf, "content-length", strconv.Itoa(n))
// 		AddDelimiter(&wbuf)
// 		//AddData(&wbuf, rbuf[:n])
// 		AddData(&wbuf, encoded_buf[:n])
// 		payload_len := n
// 		n, err = writer.Write(wbuf.Bytes())
// 		if err != nil {
// 			Log.Errorf("(%s)response encoder write error with %s", encoder.id, err)
// 		}
// 		Log.Debugf("(%s)response encoder write %d bytes, with payload:%d", encoder.id, n, payload_len)
// 		written += int64(n)
// 	}
// }
