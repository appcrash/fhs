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
	c := make(chan string)

	test_str := "just test"
	buf.Write([]byte(test_str))

	encoder := NewRequestEncoder("test", "key", &buf)
	go encoder.PipeTo(&dest, c)
	result := <-c

	t.Logf("encoder pipeto result: %s", result)
	t.Logf("request detail:\n%s", dest.String())
}
