package fhslib

import (
	"bytes"
	"io"
	"testing"
)

func init() {
	SetLogLevel("debug")
}

func TestEncoder(t *testing.T) {
	buf := bytes.Buffer{}
	dest := bytes.Buffer{}

	test_str := "just test"
	buf.Write([]byte(test_str))

	encoder := Encoder{"key", &buf}
	n, err := io.Copy(&dest, &encoder)
	if err != io.EOF || n <= 0 {
		t.Errorf("encoder error with err:%s   n:%d", err, n)
	}

	t.Logf("request detail:\n%s", dest.String())
}
