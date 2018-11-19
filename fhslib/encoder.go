package fhslib

import (
	"bytes"
	// "encoding/binary"
	// "crypto/md5"
	"github.com/golang/snappy"
	"io"
	"strconv"
)

// read from conn, encode data, write to conn
type RequestEncoder struct {
	id     string
	key    string
	reader io.Reader
}

type ResponseEncoder struct {
	id     string
	key    string
	reader io.Reader
}

const max_segment = 1024 * 32 // max request data

func NewRequestEncoder(id string, key string, reader io.Reader) RequestEncoder {
	return RequestEncoder{id, key, reader}
}

func (encoder *RequestEncoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, encoder)

	if c != nil {
		c <- "(" + encoder.id + ") request encoder"
	}
}

func (encoder *RequestEncoder) Read(p []byte) (n int, err error) {
	n, err = encoder.reader.Read(p)
	return
}

func (encoder *RequestEncoder) WriteTo(writer io.Writer) (written int64, err error) {
	var n, payload_len int
	rbuf := make([]byte, max_segment)
	wbuf := bytes.Buffer{}

	for {
		payload_len, err = encoder.reader.Read(rbuf)
		if err != nil {
			written += int64(payload_len)
			if err != io.EOF {
				Log.Errorf("(%s)encoder read with error %s", encoder.id, err)
			} else {
				Log.Debugf("(%s)encoder read eof", encoder.id)
			}

			return
		}

		wbuf.Reset()
		AddAction(&wbuf, "post", "/img")
		AddHeader(&wbuf, "content-type", "text/plain")
		AddHeader(&wbuf, "host", "from-req-encoder.com")
		AddHeader(&wbuf, "content-length", strconv.Itoa(payload_len))
		AddDelimiter(&wbuf)
		AddData(&wbuf, rbuf[:payload_len])

		n, err = writer.Write(wbuf.Bytes())
		if err != nil {
			Log.Errorf("(%s)request encoder write error with :%s", encoder.id, err)
			return
		}
		Log.Debugf("(%s)request encoder write %d bytes payload", encoder.id, payload_len)
		written += int64(n)

	}
}

func (encoder *RequestEncoder) WriteResolveRequest(writer io.Writer, domain string) error {
	wbuf := bytes.Buffer{}

	AddAction(&wbuf, "post", "/dns")
	AddHeader(&wbuf, "host", "dns.com")
	AddHeader(&wbuf, "content-length", strconv.Itoa(len(domain)))
	AddDelimiter(&wbuf)
	AddData(&wbuf, []byte(domain))

	_, err := writer.Write(wbuf.Bytes())
	return err
}

//-----------------------response encoder--------------------------------
func NewResponseEncoder(id string, key string, reader io.Reader) ResponseEncoder {

	return ResponseEncoder{id, key, reader}
}

func (encoder *ResponseEncoder) PipeTo(writer io.Writer, c chan string) {
	io.Copy(writer, encoder)

	if c != nil {
		c <- "(" + encoder.id + ") response encoder"
	}
}

func (encoder *ResponseEncoder) Read(p []byte) (n int, err error) {
	n, err = encoder.reader.Read(p)
	return
}

func (encoder *ResponseEncoder) WriteTo(writer io.Writer) (written int64, err error) {
	var n int
	rbuf := make([]byte, max_segment)
	wbuf := bytes.Buffer{}

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

		// if n > 0 {
		// 	Log.Infof("(%s)response encoder read content md5sum:%x", encoder.id, md5.Sum(rbuf[:n]))
		// }

		Log.Debugf("(%s)response encoder read %d bytes from reader", encoder.id, n)
		encoded_buf := snappy.Encode(nil, rbuf[:n])
		n = len(encoded_buf)
		wbuf.Reset()
		AddState(&wbuf, 200)
		AddHeader(&wbuf, "content-type", "text/plain")
		AddHeader(&wbuf, "host", "from-resp-encoder.com")
		AddHeader(&wbuf, "content-length", strconv.Itoa(n))
		AddDelimiter(&wbuf)
		//AddData(&wbuf, rbuf[:n])
		AddData(&wbuf, encoded_buf[:n])
		payload_len := n
		n, err = writer.Write(wbuf.Bytes())
		if err != nil {
			Log.Errorf("(%s)response encoder write error with %s", encoder.id, err)
		}
		Log.Debugf("(%s)response encoder write %d bytes, with payload:%d", encoder.id, n, payload_len)
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
