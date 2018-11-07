package fhslib

import (
	// "bytes"
	// "encoding/binary"
	// "bufio"
	"io"
)

// read from conn, decode data, write to conn
type RequestDecoder struct {
	id     string
	key    string
	reader io.Reader
	c      chan *Request
}

type ResponseDecoder struct {
	id     string
	key    string
	reader io.Reader
	c      chan *Response
}

// -------------------------- Request Decoder ------------------------------------------
func NewRequestDecoder(id string, key string, reader io.Reader) RequestDecoder {
	return RequestDecoder{id, key, reader, make(chan *Request)}
}

func (decoder *RequestDecoder) Read(p []byte) (n int, err error) {
	n, err = decoder.reader.Read(p)
	return
}

func (decoder *RequestDecoder) Prepare() {
	go GetRequests(decoder.reader, decoder.c)
}

func (decoder *RequestDecoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, decoder)

	if c != nil {
		c <- decoder.id + " request decoder"
	}
}

func (decoder *RequestDecoder) GetRequest() *Request {
	return <-decoder.c
}

func (decoder *RequestDecoder) WriteTo(writer io.Writer) (written int64, err error) {
	for {
		req := <-decoder.c

		if req == nil {
			Log.Debugf("(%s)request decoder has no more request", decoder.id)
			return
		}

		payload_len := req.Data.Len()
		if payload_len <= 0 {
			Log.Errorf("(%s)request decoder gets request with data len %d", decoder.id, payload_len)
			break
		}

		n, e := writer.Write(req.Data.Bytes())
		written += int64(n)
		if e != nil {
			err = e
			Log.Errorf("(%s)request decoder writer error", decoder.id)
			return
		}
		Log.Debugf("(%s)request decoder write %d bytes with payload(%d)", decoder.id, n, payload_len)

	}

	return
}

// -------------------------- Response Decoder ------------------------------------------
func NewResponseDecoder(id string, key string, reader io.Reader) ResponseDecoder {
	return ResponseDecoder{id, key, reader, make(chan *Response)}
}

func (decoder *ResponseDecoder) Read(p []byte) (n int, err error) {
	n, err = decoder.reader.Read(p)
	return
}

func (decoder *ResponseDecoder) Prepare() {
	go GetResponses(decoder.reader, decoder.c)
}

func (decoder *ResponseDecoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, decoder)
	if c != nil {
		c <- decoder.id + " response decoder"
	}
}

func (decoder *ResponseDecoder) GetResponse() *Response {
	return <-decoder.c
}

func (decoder *ResponseDecoder) WriteTo(writer io.Writer) (written int64, err error) {
	for {
		resp := <-decoder.c

		if resp == nil {
			Log.Debug("decoder has no more response")
			return
		}

		n := resp.Data.Len()
		if n <= 0 {
			Log.Errorf("response decoder gets response with data len %d", n)
			break
		}

		n, e := writer.Write(resp.Data.Bytes())
		written += int64(n)
		if e != nil {
			err = e
			Log.Error("response decoder writer error")
			return
		}

	}

	return
}
