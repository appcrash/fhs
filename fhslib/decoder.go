package fhslib

import (
	"bytes"
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
	var n int
	wbuf := bytes.Buffer{}
	rbuf := make([]byte, max_segment)

	for {
		n, err = decoder.reader.Read(rbuf)
		if err != nil {
			written += int64(n)
			if err != io.EOF {
				Log.Errorf("decoder read with error %s", err)
			} else {
				Log.Debugf("decoder read eof")
			}

			return
		}

		n, err = writer.Write(wbuf.Bytes())
		written += int64(n)
	}
}
