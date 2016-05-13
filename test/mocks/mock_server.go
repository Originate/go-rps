package mocks

import (
	"fmt"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
)

type MockServer struct {
	Tunnels  []*net.TCPConn
	listener *net.TCPListener

	// For testing only
	TestChannel chan string
}

func (m MockServer) Start() (*net.TCPAddr, error) {
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0,
	}

	var err error
	m.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
		return nil, err
	}

	go m.listen()

	ret, err := net.ResolveTCPAddr("tcp", m.listener.Addr().String())
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return ret, nil
}

func (m MockServer) Stop() error {
	for _, conn := range m.Tunnels {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return m.listener.Close()
}

func (m MockServer) handleConnection(conn *net.TCPConn) {
	for {
		bytes := make([]byte, 4096)
		i, err := conn.Read(bytes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		msg := &pb.TestMessage{}
		if err := proto.Unmarshal(bytes[0:i], msg); err != nil {
			fmt.Println(err.Error())
			return
		}
		m.TestChannel <- msg.Data
	}
}

func (m MockServer) listen() {
	for {
		conn, err := m.listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		m.Tunnels = append(m.Tunnels, conn)

		address := &net.TCPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 0,
		}

		listener, err := net.ListenTCP("tcp", address)
		addr, err := net.ResolveTCPAddr("tcp", listener.Addr().String())
		fmt.Printf("Listening for users on port: %d\n", addr.Port)
		conn.Write([]byte(strconv.Itoa(addr.Port)))

		go m.handleConnection(conn)
		m.TestChannel <- strconv.Itoa(len(m.Tunnels))
	}
}
