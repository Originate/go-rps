 package mocks

import (
  "net"
  "fmt"
  // "bufio"
)
  
type MockServer struct {
  Tunnels []*net.TCPConn
  listener *net.TCPListener

  // For testing only
  TunnelChannel chan int
  ReceivedMessages chan string
}

func (m MockServer) Start() (*net.TCPAddr, error) {
	address := &net.TCPAddr {
	  IP: net.IPv4(127,0,0,1),
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
	
	// bufReader := bufio.NewReader(conn)
	for {		
		fmt.Println("Here2")
		// bytes, err := bufReader.ReadBytes('\r')
		bytes := make ([]byte, 256)
		_, err := conn.Read(bytes)
		fmt.Println("Here3")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(bytes))
		fmt.Println(string(len(string(bytes))))
		if (len(string(bytes)) > 0) {
			m.ReceivedMessages <- string(bytes)	
		}
		
		fmt.Print(".")
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
