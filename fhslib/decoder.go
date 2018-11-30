package fhslib

import (
	"bytes"
	// "encoding/binary"
	// "bufio"
	// "crypto/md5"
	// "fmt"
	"github.com/golang/snappy"
	"io"
	"regexp"
)

// read from conn, decode data, write to conn
type RequestDecoder struct {
	id     string
	key    string
	reader io.Reader
}

type ResponseDecoder struct {
	id     string
	key    string
	reader io.Reader
}

// -------------------------- Request Decoder ------------------------------------------
func NewRequestDecoder(id string, key string, reader io.Reader) RequestDecoder {
	return RequestDecoder{id, key, reader}
}

func (decoder *RequestDecoder) Receive(ch_packet chan<- *Packet) {
	c := make(chan *Request)
	path_regex, _ := regexp.Compile(`/([^/]+)/(.+)`)
	go GetRequests(decoder.reader, c)
	for {
		request := <-c
		if request == nil {
			ch_packet <- nil
			break
		}

		var cmd_type int
		var tunnel_id, cmd string
		path := request.Path
		if i := path_regex.FindSubmatchIndex([]byte(path)); i == nil {
			Log.Debugf("(%s)reqeust decoder: can not parse request path:%s", decoder.id, path)
			continue // ignore this request
		} else {
			cmd = path[i[2]:i[3]]
			tunnel_id = path[i[4]:i[5]]
		}

		switch cmd {
		case "img":
			cmd_type = dtTunnelData
		case "new":
			cmd_type = dtTunnelInfo
		default:
			Log.Debugf("(%s)request decoder: can not recognize path cmd:%s", decoder.id, cmd)
			continue // ignore this request
		}

		ch_packet <- &Packet{cmd_type, tunnel_id, request.Data}
	}
}

// func (decoder *RequestDecoder) WriteTo(writer io.Writer) (written int64, err error) {
// 	for {
// 		req := <-decoder.c

// 		if req == nil {
// 			Log.Debugf("(%s)request decoder has no more request", decoder.id)
// 			return
// 		}

// 		data := req.Data.Bytes()
// 		payload_len := req.Data.Len()
// 		current := 0
// 		if payload_len <= 0 {
// 			Log.Errorf("(%s)request decoder gets request with data len %d", decoder.id, payload_len)
// 			break
// 		}

// 		for current < payload_len {
// 			n, e := writer.Write(data[current:payload_len])
// 			written += int64(n)
// 			if e != nil {
// 				err = e
// 				Log.Errorf("(%s)request decoder writer error", decoder.id)
// 				return
// 			}
// 			current += n
// 		}

// 		Log.Debugf("(%s)request decoder write %d bytes with payload(%d)", decoder.id, current, payload_len)

// 	}

// 	return
// }

// -------------------------- Response Decoder ------------------------------------------
func NewResponseDecoder(id string, key string, reader io.Reader) ResponseDecoder {
	return ResponseDecoder{id, key, reader}
}

func (decoder *ResponseDecoder) Receive(ch_packet chan<- *Packet) {
	c := make(chan *Response)
	cookie_regex, _ := regexp.Compile(`([^=]+)=(.+)`)
	go GetResponses(decoder.reader, c)
	for {
		response := <-c
		if response == nil {
			ch_packet <- nil
			break
		}

		var cmd_type int
		var tunnel_id, cmd string
		if cookie, ok := (*response.Header)["COOKIE"]; !ok {
			Log.Debugf("(%s)response decoder: can not find cookies", decoder.id)
			continue // ignore this response
		} else {
			if i := cookie_regex.FindSubmatchIndex([]byte(cookie)); i == nil {
				Log.Debugf("(%s)response decoder: can not parse cookie string:%s", decoder.id, cookie)
				continue // ignore this response
			} else {
				cmd = cookie[i[2]:i[3]]
				tunnel_id = cookie[i[4]:i[5]]
			}
		}

		switch cmd {
		case "data":
			cmd_type = dtTunnelData
		case "domain":
			cmd_type = dtTunnelInfo
		default:
			Log.Debugf("(%s)request decoder: can not recognize path cmd:%s", decoder.id, cmd)
			continue // ignore this request
		}

		data, decode_err := snappy.Decode(nil, response.Data.Bytes())
		if decode_err != nil {
			Log.Errorf("(%s)response decoder: snappy decode data error %s", decoder.id, decode_err)
			continue
		}

		ch_packet <- &Packet{cmd_type, tunnel_id, bytes.NewBuffer(data)}
	}
}

// func (decoder *ResponseDecoder) WriteTo(writer io.Writer) (written int64, err error) {
// 	for {
// 		resp := <-decoder.c

// 		if resp == nil {
// 			Log.Debug("decoder has no more response")
// 			return
// 		}

// 		origin_data := resp.Data.Bytes()
// 		total_len := resp.Data.Len()
// 		if total_len <= 0 {
// 			Log.Errorf("(%s)response decoder gets response with data len %d", decoder.id, total_len)
// 			break
// 		}

// 		// Log.Infof("(%s)response decoder get content md5sum:%x", decoder.id, md5.Sum(data))
// 		data, decode_err := snappy.Decode(nil, origin_data)
// 		if decode_err != nil {
// 			Log.Errorf("(%s)response decoder snappy decoder error: %s", decode_err)
// 			break
// 		}
// 		current := 0
// 		total_len = len(data)

// 		for current < total_len {
// 			n, e := writer.Write(data[current:total_len])

// 			if e != nil {
// 				err = e
// 				if e != io.EOF {
// 					Log.Errorf("(%s)response decoder writer error: %s", decoder.id, e)
// 				}

// 				return
// 			}
// 			current += n
// 			written += int64(n)
// 		}

// 		Log.Infof("(%s)response decoder write %d bytes", decoder.id, current)

// 	}

// 	return
// }
