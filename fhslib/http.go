package fhslib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const (
	stateFirstLine = iota
	stateFirstLineMore
	stateFirstLineDone
	stateHeader
	stateHeaderMore
	stateHeaderDone
	stateBody
	stateBodyMore
)

var delimiter []byte = []byte("\r\n")

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
	header = strings.ToUpper(header)
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
			name := strings.ToUpper(fields[0])
			h[name] = fields[1]
		}

		if i+3 >= len(search) {
			break
		}

		search = search[i+2:] // truncate
	}

	return h, err
}

func GetRequests(reader io.Reader, c chan *Request) {
	var n, body_left, body_length int
	var err error
	var line, more_line, body_data []byte
	var has_more bool
	var action, path string
	var headers *Header
	action_regex, _ := regexp.Compile(`(?i)(get|post)\s+(/[/\w]*)\s+http/(1.1|2)`)
	header_regex, _ := regexp.Compile(`([\w\-]+):\s*([\w:\-_*/]+)`)
	rd := bufio.NewReader(reader)

	state := stateFirstLine
Loop:
	for {
		switch state {
		case stateFirstLine:
			if line, has_more, err = rd.ReadLine(); err != nil {
				if err != io.EOF {
					Log.Errorf("GetRequests when in state first line with error: %s", err)
				} else {
					Log.Debug("GetRequests when in state first line with EOF")
				}

				break Loop
			}
			if has_more {
				Log.Debug("GetRequests when first line has more(too big)")
				state = stateFirstLineMore
				continue
			}
			state = stateFirstLineDone
		case stateFirstLineMore:
			if more_line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetRequests when in state first more line wtih error: %s", err)
				break Loop
			}
			line = append(line, more_line...)
			if !has_more {
				state = stateFirstLineDone
			}
		case stateFirstLineDone:
			if indexs := action_regex.FindSubmatchIndex(line); indexs == nil {
				Log.Errorf("GetRequests can not find action line, line: %s", line)
				break Loop
			} else {
				action = strings.ToUpper(string(line[indexs[2]:indexs[3]]))
				path = string(line[indexs[4]:indexs[5]])
			}
			state = stateHeader
			headers = &Header{} // create new headers for this request
		case stateHeader:
			if line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetRequests when in state header line with error: %s", err)
				break Loop
			}
			if has_more {
				Log.Debug("GetRequests when header line has more(too big)")
				state = stateHeaderMore
				continue
			}
			Log.Debugf("GetRequests read header line %s", line)
			state = stateHeaderDone
		case stateHeaderMore:
			if more_line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetRequests when in state header more line wtih error: %s", err)
				break Loop
			}
			line = append(line, more_line...)
			if !has_more {
				Log.Debugf("GetRequests read header line %s", line)
				state = stateHeaderDone
			}
		case stateHeaderDone:
			if indexs := header_regex.FindSubmatchIndex(line); indexs == nil {
				Log.Errorf("GetRequests can not validate header line, line: %s", line)
				break Loop
			} else {
				header_name := string(line[indexs[2]:indexs[3]])
				header_value := string(line[indexs[4]:indexs[5]])
				(*headers)[strings.ToUpper(header_name)] = header_value
			}

			if next, next_err := rd.Peek(2); next_err != nil {
				Log.Errorf("GetRequests peek next 2 bytes at header end with error: %s", next_err)
				break Loop
			} else {
				if bytes.Compare(next, delimiter) == 0 { // we found the delimiter after reading one header line, skip it to body
					rd.Discard(2)
					state = stateBody
					continue
				}
			}
			state = stateHeader

		case stateBody:
			switch action {
			case "GET", "HEAD":
				body_data = nil
				body_left = 0
			case "POST":
				if length_str, exists := (*headers)["CONTENT-LENGTH"]; !exists {
					Log.Errorf("GetRequests headers has no content-length: %v", headers)
					break Loop
				} else {
					if body_length, err = strconv.Atoi(length_str); err != nil {
						Log.Errorf("GetRequests parse content length(%s) error:%s", length_str, err)
						break Loop
					}
				}
				body_data = make([]byte, body_length)
				body_left = body_length
			}

			state = stateBodyMore
		case stateBodyMore:
			for body_left > 0 {
				if n, err = rd.Read(body_data[(body_length - body_left):]); err != nil {
					break
				} else {
					body_left -= n
				}
			}

			if err != nil {
				if err != io.EOF {
					Log.Errorf("GetRequests read body with error %s when body_left:%d, body_length:%d",
						err, body_left, body_length)
				} else {
					Log.Debug("GetRequests got EOF")
				}
				// exit the read cycle
				break Loop
			}

			data_buf := bytes.Buffer{}
			if body_data != nil {
				data_buf.Write(body_data)
			}

			request := Request{action, path, headers, &data_buf}
			c <- &request

			state = stateFirstLine // read next request
		}
	}

	c <- nil
}

func GetResponses(reader io.Reader, c chan *Response) {
	var n, body_left, body_length int
	var err error
	var line, more_line, body_data []byte
	var has_more bool
	var state_code string
	var headers *Header
	state_line_regex, _ := regexp.Compile(`(?i)http/(1.1|2)\s+(\d\d\d)\s+[^\r\n]+`)
	header_regex, _ := regexp.Compile(`([\w\-]+):\s*([\w:\-_*/]+)`)
	rd := bufio.NewReader(reader)

	state := stateFirstLine
Loop:
	for {
		switch state {
		case stateFirstLine:
			if line, has_more, err = rd.ReadLine(); err != nil {
				if err != io.EOF {
					Log.Errorf("GetResponses when in state first line with error: %s", err)
				} else {
					Log.Debug("GetResponses when in state first line with EOF")
				}

				break Loop
			}
			if has_more {
				Log.Debug("GetResponses when first line has more(too big)")
				state = stateFirstLineMore
				continue
			}
			state = stateFirstLineDone
		case stateFirstLineMore:
			if more_line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetResponses when in state first more line wtih error: %s", err)
				break Loop
			}
			line = append(line, more_line...)
			if !has_more {
				state = stateFirstLineDone
			}
		case stateFirstLineDone:
			if indexs := state_line_regex.FindSubmatchIndex(line); indexs == nil {
				Log.Errorf("GetResponses can not find state code line, line: %s", line)
				break Loop
			} else {
				state_code = string(line[indexs[4]:indexs[5]])
			}
			state = stateHeader
			headers = &Header{} // create new headers for this response
		case stateHeader:
			if line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetResponses when in state header line with error: %s", err)
				break Loop
			}
			if has_more {
				Log.Debug("GetResponses when header line has more(too big)")
				state = stateHeaderMore
				continue
			}
			Log.Debugf("GetResponses read header line %s", line)
			state = stateHeaderDone
		case stateHeaderMore:
			if more_line, has_more, err = rd.ReadLine(); err != nil {
				Log.Errorf("GetResponses when in state header more line wtih error: %s", err)
				break Loop
			}
			line = append(line, more_line...)
			if !has_more {
				Log.Debugf("GetResponses read header line %s", line)
				state = stateHeaderDone
			}
		case stateHeaderDone:
			if indexs := header_regex.FindSubmatchIndex(line); indexs == nil {
				Log.Errorf("GetResponses can not validate header line, line: %s", line)
				break Loop
			} else {
				header_name := string(line[indexs[2]:indexs[3]])
				header_value := string(line[indexs[4]:indexs[5]])
				(*headers)[strings.ToUpper(header_name)] = header_value
			}

			if next, next_err := rd.Peek(2); next_err != nil {
				Log.Errorf("GetResponses peek next 2 bytes at header end with error: %s", next_err)
				break Loop
			} else {
				if bytes.Compare(next, delimiter) == 0 { // we found the delimiter after reading one header line, skip it to body
					rd.Discard(2)
					state = stateBody
					continue
				}
			}
			state = stateHeader

		case stateBody:
			// content length is a must, currently not support chunk
			if length_str, exists := (*headers)["CONTENT-LENGTH"]; !exists {
				Log.Errorf("GetResponses headers has no content-length: %v", headers)
				break Loop
			} else {
				if body_length, err = strconv.Atoi(length_str); err != nil {
					Log.Errorf("GetResponses parse content length(%s) error:%s", length_str, err)
					break Loop
				}
			}
			body_data = make([]byte, body_length)
			body_left = body_length

			state = stateBodyMore
		case stateBodyMore:
			for body_left > 0 {
				if n, err = rd.Read(body_data[(body_length - body_left):]); err != nil {
					break
				} else {
					body_left -= n
				}
			}

			if err != nil {
				if err != io.EOF {
					Log.Errorf("GetResponses read body with error %s when body_left:%d, body_length:%d",
						err, body_left, body_length)
				} else {
					Log.Debug("GetResponses got EOF")
				}
				// exit the read cycle
				break Loop
			}

			data_buf := bytes.Buffer{}
			if body_data != nil {
				data_buf.Write(body_data)
			}

			response := Response{state_code, headers, &data_buf}
			c <- &response

			state = stateFirstLine // read next response
		}
	}

	c <- nil
}
