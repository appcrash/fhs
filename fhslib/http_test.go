package fhslib

import (
	"bytes"
	"testing"
)

func init() {
	SetLogLevel("debug")
}

func TestHeader(t *testing.T) {
	buf := bytes.Buffer{}
	AddHeader(&buf, "Content-Length", "80")
	AddHeader(&buf, "Content-Type", "text/html")

	headers, err := ParseHeaders(buf.String())
	t.Logf("headers are %v", headers)
	if err != nil {
		t.Error(err)
	}

	v, ok := headers["CONTENT-LENGTH"]
	if !ok {
		t.Error("header name not found")
	}
	if v != "80" {
		t.Error("header value not correct")
	}
}

func TestParseRequest(t *testing.T) {
	buf := bytes.Buffer{}
	content := "post sample content"

	AddAction(&buf, "post", "/some/loc")
	AddHeader(&buf, "content-type", "text/html")
	AddDelimiter(&buf)
	AddData(&buf, []byte(content))

	c := make(chan *Request)
	go GetRequests(&buf, c)
	req := <-c

	t.Logf(req.action)
	t.Logf("headers are %v", req.header)
	t.Logf(req.data.String())
}
