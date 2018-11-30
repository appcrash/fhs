package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	// "github.com/appcrash/fhs/fhslib"
	"io"
	"net"
	// "strconv"
	// "strings"
)

const (
	socks5Version    = uint8(5)
	connectCommand   = uint8(1)
	bindCommand      = uint8(2)
	associateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

const (
	successReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

type Socks5Server struct {
	listenAddr string
}

func (s *Socks5Server) listen() {
	l, err := net.Listen("tcp4", s.listenAddr)
	if err != nil {
		logger.Errorf("listen on %s failed", s.listenAddr)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Errorf("accept new connection failed")
			panic(err)
		}
		go handleConnection(conn)
	}
}

func readHeader(conn net.Conn) error {
	var bufConn *bufio.Reader = bufio.NewReader(conn)
	header := []byte{0, 0, 0}
	if _, err := bufConn.Read(header); err != nil {
		logger.Errorln("error reading socks version")
		return err
	}

	if header[0] != socks5Version {
		logger.Errorln("unsupported version when reading socks5 header")
		return fmt.Errorf("version not supported")
	}

	logger.Debugf("num of method is %d, method is %d", header[1], header[2])

	return nil
}

func sendNoAuthReply(writer io.Writer) error {
	_, err := writer.Write([]byte{socks5Version, 0})
	if err != nil {
		logger.Errorln("send no auth reply error")
	}
	return err
}

func sendRequestReply(writer io.Writer, resp uint8, addr net.IP, port uint16) error {
	ip_len := 4 // currently only ipv4 supported
	data := make([]byte, 6+ip_len)
	data[0] = socks5Version
	data[1] = resp
	data[3] = ipv4Address
	copy(data[4:], addr)
	binary.BigEndian.PutUint16(data[4+ip_len:], port)
	if _, err := writer.Write(data); err != nil {
		logger.Errorln("send request reply error")
		return err
	}
	return nil
}

func handleRequest(conn net.Conn) error {
	buf := []byte{0, 0, 0, 0}
	if num, err := io.ReadAtLeast(conn, buf, 4); err != nil {
		logger.Errorf("error when read request, only got %d bytes", num)
		return err
	}

	ver := buf[0]
	command := buf[1]
	addrType := buf[3]
	logger.Debugf("ver:%x cmd:%x, addrType: %x", ver, command, addrType)
	if ver != socks5Version {
		logger.Errorln("request version not supported")
		return fmt.Errorf("unsupported version")
	}

	var ip net.IP
	var ipLen int = 4
	var domain string

	// get destination's ip and port
	switch addrType {
	case fqdnAddress:
		addrLen := []byte{0}
		if _, err := conn.Read(addrLen); err != nil {
			logger.Errorln("read fqdn length error")
			return err
		}
		fqdn := make([]byte, uint(addrLen[0]))
		if _, err := conn.Read(fqdn); err != nil {
			logger.Errorln("read fqdn error")
			return err
		}
		domain = string(fqdn)
		// if addr, err := net.ResolveIPAddr("ip", domain); err != nil {
		// 	logger.Fatalf("resolving domain: %s error", domain)
		// 	return err
		// } else {
		// 	ip = addr.IP
		// }
		logger.Debugf("request domain: %s", domain)
	case ipv4Address:
		fallthrough
	case ipv6Address:
		ipLen = 16
		ip = make([]byte, ipLen)
		if _, err := conn.Read(ip); err != nil {
			logger.Errorln("read ip address error")
			return err
		}
	default:
		logger.Fatalf("unsupported request address type: %x", addrType)
		return fmt.Errorf("unsupported address type")
	}

	portBuf := []byte{0, 0}
	if _, err := conn.Read(portBuf); err != nil {
		logger.Errorln("read request port error")
		return err
	}
	// var port uint16 = binary.BigEndian.Uint16(portBuf)

	switch command {
	case connectCommand:
		// handleConnect(conn, domain, int(port))
	case bindCommand:
		fallthrough
	case associateCommand:
		fallthrough
	default:
		logger.Fatalf("unsupported command: %x", command)
		return fmt.Errorf("unsupported command: %x", command)
	}

	return nil
}

// func handleConnect(conn net.Conn, domain string, dest_port int) {
// 	config := fhslib.GetConfig()
// 	remote_addr := net.JoinHostPort(domain, strconv.Itoa(dest_port))
// 	fhs_server_addr := net.JoinHostPort(config.Server.Ip, strconv.Itoa(config.Server.Port))
// 	fhs_server_conn, conn_err := net.Dial("tcp", fhs_server_addr)

// 	if conn_err != nil {
// 		logger.Error("connect to remote server error")
// 		sendRequestReply(conn, hostUnreachable, []byte{0, 0, 0, 0}, 0)
// 		return
// 	}

// 	track_id := fhslib.GenerateId()
// 	encoder := fhslib.NewRequestEncoder(track_id, "key", conn)
// 	decoder := fhslib.NewResponseDecoder(track_id, "key", fhs_server_conn)
// 	encoder.WriteResolveRequest(fhs_server_conn, remote_addr)
// 	decoder.Prepare()
// 	resp := decoder.GetResponse()

// 	if resp == nil {
// 		logger.Error("decoder expect dns response but get nothing")
// 		return
// 	}
// 	addr := resp.Data.String()
// 	logger.Debugf("dns response is %s", addr)
// 	addr_array := strings.Split(addr, ":")
// 	ip_str, port_str := addr_array[0], addr_array[1]
// 	logger.Debugf("remote local bind ip:%s, port:%s", ip_str, port_str)
// 	ip := net.ParseIP(ip_str)
// 	port, _ := strconv.Atoi(port_str)
// 	logger.Debugf("send socks reply with success ip: %s, port: %d", ip, port)
// 	sendRequestReply(conn, successReply, ip, uint16(port))

// 	// pipe bidirection
// 	c := make(chan string, 2)
// 	go encoder.PipeTo(fhs_server_conn, c)
// 	go decoder.PipeTo(conn, c)

// 	for i := 0; i < 2; i++ {
// 		name := <-c
// 		logger.Debugf("%s done", name)
// 	}
// }

func handleConnection(conn net.Conn) {
	logger.Infof("new connection acceptted from %s", conn.RemoteAddr())
	defer conn.Close()

	if err := readHeader(conn); err != nil {
		return
	}

	if err := sendNoAuthReply(conn); err != nil {
		return
	}

	if err := handleRequest(conn); err != nil {
		return
	}
}
