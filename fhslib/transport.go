package fhslib

import (
	//"github.com/golang-collections/go-datastructures/queue"
	// "github.com/satori/go.uuid"
	"net"
)

type flowdata chan []byte

// type Flow interface {
// 	NewSession(s *ClientSession) error
// 	Start(fromLeft flowdata, toRight flowdata, fromRight flowdata, toLeft flowdata)
// }

type ClientSession struct {
	id   string
	conn net.Conn

	RemoteAddr string
}

var all_session = map[string]ClientSession{}

func GetClientSession(id string) ClientSession {
	return all_session[id]
}

// func StartClientSession(conn net.Conn, remoteAddr string, c chan *net.TCPAddr) ClientSession {
// 	newid, _ := uuid.NewV4()
// 	idstr := newid.String()
// 	s := ClientSession{
// 		id:         idstr,
// 		conn:       conn,
// 		flows:      []*Flow{&HttpFlow{}},
// 		RemoteAddr: remoteAddr,
// 	}

// 	for _, flow := range s.flows {
// 		flow.NewSession(idstr)
// 	}

// 	var leftIn, leftOut, rightIn, rightOut flowdata
// 	for idx, flow := range s.flows {
// 		leftIn := make(flowdata)
// 		leftOut := make(flowdata)
// 		if idx == 0 {
// 			flowOutFromConnection(conn, leftIn)
// 			flowIntoConnection(conn, leftOut)
// 		}
// 		if idx != len(s.flows) {
// 			rightIn = make(flowdata)
// 			rightOut = make(flowdata)
// 		} else {
// 			rightIn = nil
// 			rightOut = nil
// 		}
// 		flow.Start(leftIn, rightOut, rightIn, leftOut)
// 	}

// 	all_session[idstr] = s
// 	return s
// }

func flowOutFromConnection(conn net.Conn, c flowdata) {
	go func() {
		buffer := make([]byte, 32*1024)
		for {
			nr, err := conn.Read(buffer)
			if nr > 0 {
				c <- buffer[0:nr]
			}
			if err != nil {
				Log.Debug("finish flow out connection")
				close(c)
			}
		}
	}()
}

func flowIntoConnection(conn net.Conn, c flowdata) {
	go func() {
		for data := range c {
			conn.Write(data)
		}

		Log.Debug("finish flow into connection")
	}()
}
