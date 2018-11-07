package fhslib

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
)

type Header map[string]string
type Request struct {
	Action string
	Path   string
	Header *Header
	Data   *bytes.Buffer
}

type Response struct {
	State  string
	Header *Header
	Data   *bytes.Buffer
}

var config Config

func init() {
	config, _ = GetConfig()
}

type HttpServer struct {
	ListenAddr string
	Handler    ConnectionHandler
}

type ConnectionHandler interface {
	HandleConnection(net.Conn)
}

func NewHttpServer(addr string, handler ConnectionHandler) HttpServer {
	return HttpServer{addr, handler}
}

func (s *HttpServer) Listen() {
	l, err := net.Listen("tcp4", s.ListenAddr)
	if err != nil {
		Log.Errorf("listen on %s failed", s.ListenAddr)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			Log.Errorf("accept new connection failed")
			panic(err)
		}
		go s.handleConnection(conn)
	}
}

func (s *HttpServer) handleConnection(conn net.Conn) {
	s.Handler.HandleConnection(conn)
}

func ResolveName(name string, c chan *net.TCPAddr) {
	conn, err := net.Dial("tcp", config.Server.Ip)
	if err != nil {
		c <- nil
		return
	}

	c <- conn.LocalAddr().(*net.TCPAddr)
}

func parseHeader(line string) ([]string, error) {
	i := strings.Index(line, ":")
	if i == -1 {
		return nil, fmt.Errorf("invalid header %s", line)
	}
	name := strings.ToUpper(line[:i])
	value := strings.Trim(line[i+1:], " ")
	return []string{name, value}, nil
}

func AddAction(buf *bytes.Buffer, action string, location string) {
	action = strings.ToUpper(action)
	switch action {
	case "GET":
		fallthrough
	case "POST":
		fallthrough
	case "UPDATE":
		fallthrough
	case "DELETE":
		data := fmt.Sprintf("%s %s HTTP/1.1\r\n", action, location)
		buf.WriteString(data)
	default:
		Log.Errorf("invalid action: %s", action)
	}

}

func AddState(buf *bytes.Buffer, state int) {
	data := fmt.Sprintf("HTTP/1.1 %d StateText\r\n", state)
	buf.WriteString(data)
}

func AddHeader(buf *bytes.Buffer, header string, value string) {
	data := fmt.Sprintf("%s: %s\r\n", header, value)
	// Log.Debugf("add header string %s", data)
	buf.WriteString(data)
}

func AddDelimiter(buf *bytes.Buffer) {
	buf.WriteString("\r\n")
}

func AddData(buf *bytes.Buffer, data []byte) {
	buf.Write(data)
}

func ParseHeaders(buf string) (Header, error) {
	search := buf
	h := Header{}
	var i int = 0
	var err error = nil

	for i >= 0 {
		i = strings.Index(search, "\r\n")
		// Log.Debugf("ParseHeaders: index is %d, search is %s", i, search)
		if i > 0 {
			substr := search[:i]

			fields, e := parseHeader(substr)
			if e != nil {
				return nil, e
			}
			h[fields[0]] = fields[1]
		}

		if i+3 >= len(search) {
			break
		}

		search = search[i+2:] // truncate
	}

	return h, err
}

func GetRequests(reader io.Reader, c chan *Request) {
	const bufsize = 1024 * 1024 * 4
	rbuf := make([]byte, bufsize)
	var n int
	var err error
	action_regex, _ := regexp.Compile(`(?i)(get|post)\s+(/[/\w]*)\s+http/(1.1|2)`)
	delimiter_regex, _ := regexp.Compile(`\r\n\r\n`)

	for {
		n, err = reader.Read(rbuf)
		if err != nil {
			if err != io.EOF {
				Log.Errorf("GetRequests with error %s", err)
			}
			break
		}

		indexs := action_regex.FindSubmatchIndex(rbuf)
		if indexs == nil {
			Log.Error("GetRequests can not find action line")
			break
		}
		action := string(rbuf[indexs[2]:indexs[3]])
		path := string(rbuf[indexs[4]:indexs[5]])
		rest := rbuf[indexs[1]:n]
		// Log.Debugf("rest is %s", rest)

		delimiter_index := delimiter_regex.FindIndex(rest)
		if delimiter_index == nil {
			Log.Errorf("GetRequests can not find delimiter")
			break
		}

		header_end := delimiter_index[1] - 2
		header, err := ParseHeaders(string(rest[:header_end]))
		if err != nil {
			Log.Errorf("header error %s", err)
			Log.Error("GetRequests parse header error")
			break
		}

		data_buf := bytes.Buffer{}
		data_buf.Write(rest[delimiter_index[1]:])

		request := Request{action, path, &header, &data_buf}
		c <- &request
	}

	c <- nil
}

func GetResponses(reader io.Reader, c chan *Response) {
	const bufsize = 128 * 1024
	rbuf := make([]byte, bufsize)
	var n int
	var err error
	state_regex, _ := regexp.Compile(`(?i)http/(1.1|2)\s+(\d\d\d)\s+[^\r\n]+\r\n`)
	delimiter_regex, _ := regexp.Compile(`\r\n\r\n`)

	for {
		n, err = reader.Read(rbuf)
		if err != nil {
			if err != io.EOF {
				Log.Errorf("response reader error with error:%s", err)
			}
			break
		}

		indexs := state_regex.FindSubmatchIndex(rbuf)
		if indexs == nil {
			Log.Error("GetResponses can not find state line")
			break
		}

		state := string(rbuf[indexs[4]:indexs[5]])
		rest := rbuf[indexs[1]:n]

		delimiter_index := delimiter_regex.FindIndex(rest)
		if delimiter_index == nil {
			Log.Error("GetRersponses can not find delimiter")
			break
		}

		header_end := delimiter_index[1] - 2
		header, err := ParseHeaders(string(rest[:header_end]))
		if err != nil {
			Log.Errorf("parse response header error %s", err)
			break
		}

		data_buf := bytes.Buffer{}
		data_buf.Write(rest[delimiter_index[1]:])

		response := Response{state, &header, &data_buf}
		c <- &response

	}

	c <- nil
}
