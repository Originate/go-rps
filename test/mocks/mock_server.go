 package mocks

import (
  "net"
  "fmt"
  "bufio"
)
  
type MockServer struct {
  Tunnels []*net.TCPConn
  listener *net.TCPListener
  TunnelChannel chan int
  ReceivedMessages chan string
}

func (m MockServer) Start() (string, error) {
	address := &net.TCPAddr {
	  IP: net.IPv4(127,0,0,1),
	  Port: 0,
	}
	
  var err error
	m.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
	  return "", err
	}
	
	go m.listen()
	
	return m.listener.Addr().String(), nil
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
	bufReader := bufio.NewReader(conn)
	for {
		bytes, err := bufReader.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s", bytes)
		m.ReceivedMessages <- string(bytes)
	}
}

func (m MockServer) listen() {
	for {
		conn, err := m.listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		m.Tunnels = append(m.Tunnels, conn)
		go m.handleConnection(conn)
		m.TunnelChannel <- len(m.Tunnels)
	}
}
