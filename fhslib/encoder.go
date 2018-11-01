package fhslib

import (
	"bytes"
	// "encoding/binary"
	"io"
	"strconv"
)

// read from conn, encode data, write to conn
type RequestEncoder struct {
	key    string
	reader io.Reader
}

type ResponseEncoder struct {
	key    string
	reader io.Reader
}

const max_segment = 4096 // max request data

func NewRequestEncoder(key string, reader io.Reader) RequestEncoder {
	return RequestEncoder{key, reader}
}

func (encoder *RequestEncoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, encoder)
	Log.Debug("request encoder done")

	if c != nil {
		c <- "request encoder done"
	}
}

func (encoder *RequestEncoder) Read(p []byte) (n int, err error) {
	n, err = encoder.reader.Read(p)
	return
}

func (encoder *RequestEncoder) WriteTo(writer io.Writer) (written int64, err error) {
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
				Log.Debug("encoder read eof")
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

func (encoder *RequestEncoder) WriteResolveRequest(writer io.Writer, domain string) error {
	wbuf := bytes.Buffer{}

	AddAction(&wbuf, "post", "/dns")
	AddHeader(&wbuf, "host", "wwpp.com")
	AddHeader(&wbuf, "content-length", strconv.Itoa(len(domain)))
	AddDelimiter(&wbuf)
	AddData(&wbuf, []byte(domain))

	_, err := writer.Write(wbuf.Bytes())
	return err
}

//-----------------------response encoder--------------------------------
func NewResponseEncoder(key string, reader io.Reader) ResponseEncoder {
	return ResponseEncoder{key, reader}
}

func (encoder *ResponseEncoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, encoder)
	Log.Debug("response encoder done")

	if c != nil {
		c <- "response encoder done"
	}
}

func (encoder *ResponseEncoder) Read(p []byte) (n int, err error) {
	n, err = encoder.reader.Read(p)
	return
}

func (encoder *ResponseEncoder) WriteTo(writer io.Writer) (written int64, err error) {
	var n int
	wbuf := bytes.Buffer{}
	rbuf := make([]byte, max_segment)

	for {
		n, err = encoder.reader.Read(rbuf)
		if err != nil {
			written += int64(n)
			if err != io.EOF {
				Log.Errorf("response encoder read with error %s", err)
			} else {
				Log.Debug("response encoder read eof")
			}

			return
		}
		AddState(&wbuf, 200)
		AddHeader(&wbuf, "content-type", "text/plain")
		AddHeader(&wbuf, "host", "wwpp.com")
		AddHeader(&wbuf, "content-length", strconv.Itoa(n))
		AddDelimiter(&wbuf)
		AddData(&wbuf, rbuf)
		n, err = writer.Write(wbuf.Bytes())
		written += int64(n)
	}
}

func (encoder *ResponseEncoder) WriteResolveResponse(writer io.Writer, domain string) error {
	wbuf := bytes.Buffer{}

	AddState(&wbuf, 200)
	AddHeader(&wbuf, "server", "dns")
	AddHeader(&wbuf, "content-length", strconv.Itoa(len(domain)))
	AddDelimiter(&wbuf)
	AddData(&wbuf, []byte(domain))

	_, err := writer.Write(wbuf.Bytes())
	return err
}
