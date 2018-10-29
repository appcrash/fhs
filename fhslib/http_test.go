package fhslib

import (
	"bytes"
	"strconv"
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

	t.Logf(req.Action)
	t.Logf("headers are %v", req.Header)
	t.Logf(req.Data.String())
}

func TestParseResponse(t *testing.T) {
	buf := bytes.Buffer{}
	content := "some response sample content"

	AddState(&buf, 200)
	AddHeader(&buf, "content-type", "text/html")
	AddHeader(&buf, "content-length", strconv.Itoa(len(content)))
	AddDelimiter(&buf)
	AddData(&buf, []byte(content))

	c := make(chan *Response)
	go GetResponses(&buf, c)
	resp := <-c

	if resp == nil {
		t.Error("parse response error")
		return
	}

	t.Logf("response state code is %s", resp.State)
	t.Logf("headers are %v", resp.Header)
	t.Logf(resp.Data.String())
}
