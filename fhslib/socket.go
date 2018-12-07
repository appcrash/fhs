package fhslib

type HttpSocket struct {
	socket_id string
	encoder   Encoder
	decoder   Decoder
}

type SocketErrorInfo struct {
	socket_id string
	err       error
}
