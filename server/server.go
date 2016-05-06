package server

import (
  "net"
  "fmt"
  pb "github.com/Originate/go_rps/protobuf"
  "github.com/golang/protobuf/proto"
  "strconv"
)

type GoRpsServer struct {
	TunnelPort       int
	// InboundPortRange []int
  	ExposedPortsToClients map[int]*net.TCPConn
  	listener *net.TCPListener

  	// For testing only
  	TestChannel chan string
}

func (s GoRpsServer) Start() (*net.TCPAddr, error) {
	s.ExposedPortsToClients = make(map[int]*net.TCPConn)
	address := &net.TCPAddr {
	  IP: net.IPv4(127,0,0,1),
	  Port: 0,
	}
	
  	var err error
	s.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
	  return nil, err
	}
	
	go s.listenForClients()

	ret, err := net.ResolveTCPAddr("tcp", s.listener.Addr().String())
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	s.TunnelPort = ret.Port

	return ret, nil
}


func (s GoRpsServer) Stop() error {
  for _, conn := range s.ExposedPortsToClients {
    err := conn.Close()
    if err != nil {
      return err
    }
  }
  return s.listener.Close()
}

func (s GoRpsServer) handleConnection(conn *net.TCPConn) {
	// Read data from clients
	for {		
		fmt.Println("Waiting for data from client...")
		bytes := make ([]byte, 4096)
		i, err := conn.Read(bytes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("Received %s from client\n", bytes)
		msg := &pb.TestMessage{}
		if err := proto.Unmarshal(bytes[0:i], msg); err != nil {
			fmt.Println(err.Error())
			return
		}

		s.TestChannel <- msg.Data
	}
}

func (s GoRpsServer) listenForClients() {
	for {
		// Listen for a client to connect
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		
		// Choose a random free port to expose to users
		address := &net.TCPAddr {
			IP: net.IPv4(127,0,0,1),
			Port: 0,
		}

		// Create a listener for that port, and extract the chosen port
		userListener, err := net.ListenTCP("tcp", address)
		addr, err := net.ResolveTCPAddr("tcp", userListener.Addr().String())
		exposedPort := addr.Port
		fmt.Printf("Listening for users on port: %d\n", exposedPort)

		portStr := strconv.Itoa(exposedPort)
		// fmt.Println(portStr)

		msg := &pb.TestMessage {
			Type: pb.TestMessage_ConnectionOpen,
			Data: portStr,
		}
		bytes, err := proto.Marshal(msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("Sending %s\n", bytes)
		// Tell the client what port is exposed to users for their connection
		conn.Write(bytes)

		// Associate the exposed port with the current client
		s.ExposedPortsToClients[exposedPort] = conn

		// Start listening for users 
		go s.listenForUsers(userListener, exposedPort, conn)

		go s.handleConnection(conn)

		s.TestChannel <- strconv.Itoa(len(s.ExposedPortsToClients))
	}
}

func (s GoRpsServer) listenForUsers(userListener *net.TCPListener, exposedPort int, clientConn *net.TCPConn) {
	for {
		fmt.Printf("Listening for users on addr: %s\n", userListener.Addr().String())
		// Listen for a user connection
		userConn, err := userListener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		fmt.Println("User connection established")
		// Read info from user
		bytes := make ([]byte, 4096)
		i, err := userConn.Read(bytes)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Read %s from user\n", bytes)

		msg := &pb.TestMessage {
			Type: pb.TestMessage_Data,
			Data: string(bytes[0:i]),
		}

		out, err := proto.Marshal(msg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Forward data to the associated client
		clientConn.Write(out)

	}
}
