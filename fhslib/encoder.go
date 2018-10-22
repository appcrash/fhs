package fhslib

import (
	"bytes"
	// "encoding/binary"
	"io"
	"strconv"
)

// read from conn, encode data, write to conn
type Encoder struct {
	key    string
	reader io.Reader
}

const max_segment = 4096 // max request data

func (encoder *Encoder) Read(p []byte) (n int, err error) {
	n, err = encoder.reader.Read(p)
	return
}

func (encoder *Encoder) WriteTo(writer io.Writer) (written int64, err error) {
	var n int
	wbuf := bytes.Buffer{}
	rbuf := make([]byte, max_segment)

	for {
		n, err = encoder.reader.Read(rbuf)
		if err != nil {
			written += int64(n)
			if err != io.EOF {
				Log.Errorf("encoder read with error %s", err)
			} else {
				Log.Debugf("encoder read eof")
			}

			return
		}
		AddAction(&wbuf, "post", "/img")
		AddHeader(&wbuf, "content-type", "text/plain")
		AddHeader(&wbuf, "host", "wwpp.com")
		AddHeader(&wbuf, "content-length", strconv.Itoa(n))
		AddDelimiter(&wbuf)
		AddData(&wbuf, rbuf)
		n, err = writer.Write(wbuf.Bytes())
		written += int64(n)
	}
}
