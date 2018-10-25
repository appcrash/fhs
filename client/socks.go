package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/appcrash/fhs/fhslib"
	"io"
	"net"
	"strconv"
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
	data := make([]byte, 6+len(addr))
	data[0] = socks5Version
	data[1] = resp
	data[3] = ipv4Address
	copy(data[4:], addr)
	binary.BigEndian.PutUint16(data[4+len(addr):], port)
	if _, err := writer.Write(data); err != nil {
		logger.Errorln("send request reply error")
		return err
	}
	return nil
}

func pipe(reader io.Reader, writer io.Writer, name string, c chan string) {
	_, err := io.Copy(writer, reader)
	logger.Debugf("pipe %s done", name)
	if err != nil {
		logger.Errorf("pipe %s error: %v", name, err)
	}

	c <- name
}

func handleRequest(conn net.Conn) error {
	buf := []byte{0, 0, 0, 0}
	if num, err := io.ReadAtLeast(conn, buf, 4); err != nil {
		logger.Fatalf("error when read request, only got %d bytes", num)
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
		domainName := string(fqdn)
		if addr, err := net.ResolveIPAddr("ip", domainName); err != nil {
			logger.Fatalf("resolving domain: %s error", domainName)
			return err
		} else {
			ip = addr.IP
		}
		logger.Debugf("request domain: %s, ip: %s", domainName, ip)
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
	var port uint16 = binary.BigEndian.Uint16(portBuf)

	switch command {
	case connectCommand:
		handleConnect(conn, ip, int(port))
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

func handleConnect(conn net.Conn, destIp net.IP, destPort int) {
	remoteAddr := net.JoinHostPort(destIp.String(), strconv.Itoa(destPort))
	// remoteConn, connErr := net.Dial("tcp", remoteAddr)

	// if connErr != nil {
	// 	logger.Errorf("connect to remote host error: ip:%s port:%d", destIp, destPort)
	// 	sendRequestReply(conn, hostUnreachable, []byte{0, 0, 0, 0}, 0)
	// 	return
	// }

	c := make(chan *net.TCPAddr)
	go fhslib.ResolveName(remoteAddr, c)
	localBindAddr := <-c

	sendRequestReply(conn, successReply, localBindAddr.IP, uint16(localBindAddr.Port))

	// c := make(chan string, 2)
	// go pipe(remoteConn, conn, "remote2client", c)
	// go pipe(conn, remoteConn, "client2remote", c)

	// for i := 0; i < 2; i++ {
	// 	name := <-c
	// 	logger.Debugf("%s done", name)
	// }
}

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
