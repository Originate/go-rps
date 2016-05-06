 package mocks

import (
  "net"
  "fmt"
  pb "github.com/Originate/go_rps/protobuf"
  "github.com/golang/protobuf/proto"
  "log"
  // "io"
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
		// bytes, err := bufReader.ReadBytes('\r')
		bytes := make ([]byte, 256)
		i, err := conn.Read(bytes)
		// i, err := io.ReadFull(conn, bytes)
		if err != nil {
			fmt.Println("Got an error")
			fmt.Printf("Read %d bytes\n", i)
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("Received: %s---\n",bytes)
		msg := &pb.TestMessage{}
		if err := proto.Unmarshal(bytes, msg); err != nil {
			log.Fatal(err)//fmt.Println(err.Error())
			m.ReceivedMessages <- "error"
			return
		}
		m.ReceivedMessages <- string(bytes)
		
	}
	fmt.Println("end")
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
