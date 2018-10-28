package fhslib

import (
	// "bytes"
	// "encoding/binary"
	// "bufio"
	"io"
)

// read from conn, decode data, write to conn
type RequestDecoder struct {
	key    string
	reader io.Reader
	writer io.Writer
	c      chan *Request
}

type ResponseDecoder struct {
	RequestDecoder
}

func (decoder *RequestDecoder) NewRequestEncoder(key string, reader io.Reader, writer io.Writer) RequestDecoder {
	return RequestDecoder{key, reader, writer, make(chan *Request)}
}

func (decoder *RequestDecoder) Read(p []byte) (n int, err error) {
	n, err = decoder.reader.Read(p)
	return
}

func (decoder *RequestDecoder) Prepare() {
	go GetRequests(decoder.reader, decoder.c)
}

func (decoder *RequestDecoder) Start() {
	io.Copy(decoder.writer, decoder)
}

func (decoder *RequestDecoder) GetRequest() *Request {
	return <-decoder.c
}

func (decoder *RequestDecoder) WriteTo(writer io.Writer) (written int64, err error) {
	for {
		req := <-decoder.c

		if req == nil {
			Log.Debug("decoder has no more request")
			return
		}

		n := req.data.Len()
		if n <= 0 {
			Log.Errorf("decoder gets request with data len %d", n)
			break
		}

		n, e := writer.Write(req.data.Bytes())
		written += int64(n)
		if e != nil {
			err = e
			Log.Errorf("decoder writer error")
			return
		}

	}

	return
}
