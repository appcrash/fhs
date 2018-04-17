package fhslib

import (
	"bytes"
	"testing"
)

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
