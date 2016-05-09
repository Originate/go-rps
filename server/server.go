package server

import (
	"fmt"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
)

type GoRpsServer struct {
	TunnelPort int
	// InboundPortRange []int
	ExposedPortsToClients map[int]*net.TCPConn
	listener              *net.TCPListener
	ClientConnToUser      map[*net.TCPConn]*net.TCPConn

	// For testing only
	// TestChannel chan string
}

func (s GoRpsServer) Start() (*net.TCPAddr, error) {
	s.ExposedPortsToClients = make(map[int]*net.TCPConn)
	s.ClientConnToUser = make(map[*net.TCPConn]*net.TCPConn)
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 56325, // Constant port for now
	}

	var err error
	s.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
		return nil, err
	}

	go s.listenForClients()

	ret, err := net.ResolveTCPAddr("tcp", s.listener.Addr().String())
	if err != nil {
		serverTag()
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Printf("port: %d\n", ret.Port)
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

func (s GoRpsServer) handleClientConnection(clientConn *net.TCPConn) {
	// Read data from clients
	for {
		serverTag()
		fmt.Println("Waiting for data from client...")
		bytes := make([]byte, 4096)
		i, err := clientConn.Read(bytes)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
			return
		}
		serverTag()
		fmt.Printf("Received %s from client\n", bytes)
		msg := &pb.TestMessage{}
		if err := proto.Unmarshal(bytes[0:i], msg); err != nil {
			serverTag()
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("Sending: %s\n", msg.Data)
		// s.TestChannel <- msg.Data
		fmt.Println("---------")
		// Write to associated user connection
		serverTag()
		fmt.Printf("Trying %x as clientConn\n", clientConn)
		if s.ClientConnToUser[clientConn] != nil {
			serverTag()
			fmt.Printf("Writing to: %x\n", s.ClientConnToUser[clientConn])
			s.ClientConnToUser[clientConn].Write(bytes)
			// s.TestChannel <- msg.Data
		}
	}
}

func (s GoRpsServer) handleUserConnection(userConn *net.TCPConn, clientConn *net.TCPConn) {
	for {
		// Read info from user
		bytes := make([]byte, 4096)
		i, err := userConn.Read(bytes)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
			return
		}

		serverTag()
		fmt.Printf("Read %s from user @ %x\n", bytes, userConn)

		msg := &pb.TestMessage{
			Type: pb.TestMessage_Data,
			Data: string(bytes[0:i]),
		}

		// Forward data to associated client
		sendToClient(msg, clientConn)
	}
}

func (s GoRpsServer) listenForClients() {
	for {
		// Listen for a client to connect
		clientConn, err := s.listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		serverTag()
		fmt.Printf("Client connected: %x\n", clientConn)
		// Choose a random free port to expose to users
		address := &net.TCPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 0,
		}

		// Create a listener for that port, and extract the chosen port
		userListener, err := net.ListenTCP("tcp", address)
		addr, err := net.ResolveTCPAddr("tcp", userListener.Addr().String())
		exposedPort := addr.Port
		serverTag()
		fmt.Printf("Listening for users on port: %d\n", exposedPort)

		portStr := strconv.Itoa(exposedPort)

		msg := &pb.TestMessage{
			Type: pb.TestMessage_ConnectionOpen,
			Data: portStr,
		}
		bytes, err := proto.Marshal(msg)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
		}

		// Tell the client what port is exposed to users for their connection
		clientConn.Write(bytes)

		// Associate the exposed port with the current client
		s.ExposedPortsToClients[exposedPort] = clientConn

		// Start listening for users
		go s.listenForUsers(userListener, exposedPort, clientConn)

		go s.handleClientConnection(clientConn)

		serverTag()
		// s.TestChannel <- strconv.Itoa(len(s.ExposedPortsToClients))
	}
}

func (s GoRpsServer) listenForUsers(userListener *net.TCPListener, exposedPort int, clientConn *net.TCPConn) {
	for {
		serverTag()
		fmt.Printf("Listening for users on addr: %s...\n", userListener.Addr().String())
		// Listen for a user connection
		userConn, err := userListener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		serverTag()
		fmt.Printf("Saving %x: %x\n", clientConn, userConn)
		s.ClientConnToUser[clientConn] = userConn

		serverTag()
		fmt.Println("User connection established")

		go s.handleUserConnection(userConn, clientConn)
	}
}

func sendToClient(msg *pb.TestMessage, clientConn *net.TCPConn) {
	out, err := proto.Marshal(msg)
	if err != nil {
		serverTag()
		fmt.Println(err.Error())
		return
	}

	// Forward data to the associated client
	clientConn.Write(out)
}

func serverTag() {
	fmt.Print("Server: ")
}
