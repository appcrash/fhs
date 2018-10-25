package fhslib

import (
	// "bytes"
	// "encoding/binary"
	// "bufio"
	"io"
)

// read from conn, decode data, write to conn
type Decoder struct {
	key    string
	reader io.Reader
}

func (decoder *Decoder) Read(p []byte) (n int, err error) {
	n, err = decoder.reader.Read(p)
	return
}

func (decoder *Decoder) WriteTo(writer io.Writer) (written int64, err error) {
	c := make(chan *Request)

	go GetRequests(decoder.reader, c)
	for {
		req := <-c

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
