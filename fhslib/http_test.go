package fhslib

import (
	"bytes"
	"strconv"
	"testing"
)

func init() {
	SetLogLevel("debug")
}

func TestParseRequest(t *testing.T) {
	buf := bytes.Buffer{}
	content1 := "post sample content"
	content2 := "post other data "

	AddAction(&buf, "post", "/some/loc")
	AddHeader(&buf, "content-type", "text/html")
	AddHeader(&buf, "content-length", strconv.Itoa(len(content1)))
	AddDelimiter(&buf)
	AddData(&buf, []byte(content1))

	AddAction(&buf, "post", "/sdoiew")
	AddHeader(&buf, "content-length", strconv.Itoa(len(content2)))
	AddDelimiter(&buf)
	AddData(&buf, []byte(content2))

	AddAction(&buf, "get", "/")
	AddHeader(&buf, "accept", "*/*")
	AddDelimiter(&buf)

	c := make(chan *Request)
	go GetRequests(&buf, c)
	print_req := func() {
		req := <-c
		t.Logf("request action is %s, path is %s", req.Action, req.Path)
		t.Logf("headers are %v", req.Header)
		t.Logf(req.Data.String())
	}

	for i := 0; i < 3; i++ {
		print_req()
	}

}

func TestParseResponse(t *testing.T) {
	buf := bytes.Buffer{}
	content1 := "some response sample content"
	content2 := "another response sample content"

	AddState(&buf, 200)
	AddHeader(&buf, "content-type", "image/jpg")
	AddHeader(&buf, "content-length", strconv.Itoa(len(content1)))
	AddDelimiter(&buf)
	AddData(&buf, []byte(content1))

	AddState(&buf, 404)
	AddHeader(&buf, "content-type", "text/html")
	AddHeader(&buf, "content-length", strconv.Itoa(len(content2)))
	AddDelimiter(&buf)
	AddData(&buf, []byte(content2))

	c := make(chan *Response)
	go GetResponses(&buf, c)

	print_resp := func() {
		resp := <-c

		if resp == nil {
			t.Error("parse response error")
			return
		}
		t.Logf("response state code is %s", resp.State)
		t.Logf("headers are %v", resp.Header)
		t.Logf(resp.Data.String())
	}

	for i := 0; i < 2; i++ {
		print_resp()
	}

}
