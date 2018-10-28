package fhslib

import (
	"bytes"
	// "io"
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

	encoder := RequestEncoder{"key", &buf, &dest}
	encoder.Start()

	t.Logf("request detail:\n%s", dest.String())
}
