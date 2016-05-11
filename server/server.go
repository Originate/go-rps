package server

import (
	"fmt"
	"github.com/Originate/go_rps/helper"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
)

type GoRpsServer struct {
	TunnelPort int
	HostIP     net.IP

	UserConn map[int32]*net.TCPConn
	UserId   map[*net.TCPConn]int32

	users          int32
	clientListener *net.TCPListener
}

func (s GoRpsServer) Start() (*net.TCPAddr, error) {
	s.users = 0
	if s.HostIP == nil {
		serverTag()
		fmt.Printf("Setting default HostIP: localhost\n")
		s.HostIP = net.IPv4(0, 0, 0, 0)
	}
	s.UserConn = make(map[int32]*net.TCPConn)
	s.UserId = make(map[*net.TCPConn]int32)

	port := 0
	address := &net.TCPAddr{
		IP:   s.HostIP,
		Port: port,
	}

	var err error
	s.clientListener, err = net.ListenTCP("tcp", address)
	if err != nil {
		return nil, err
	}

	go s.listenForClients()

	ret, err := net.ResolveTCPAddr("tcp", s.clientListener.Addr().String())
	if err != nil {
		serverTag()
		fmt.Println(err.Error())
		return nil, err
	}
	serverTag()
	fmt.Printf("port: %d\n", ret.Port)
	s.TunnelPort = ret.Port

	return ret, nil
}

func (s GoRpsServer) listenForClients() {
	for {
		// Listen for a client to connect
		serverTag()
		fmt.Printf("Waiting for client connections...\n")
		clientConn, err := s.clientListener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		serverTag()
		fmt.Printf("Client connected: %x\n", clientConn)
		// Choose a random free port to expose to users
		address := &net.TCPAddr{
			IP:   s.HostIP,
			Port: 0,
		}

		// Create a listener for that port, and extract the chosen port
		userListener, err := net.ListenTCP("tcp", address)
		addr, err := net.ResolveTCPAddr("tcp", userListener.Addr().String())
		exposedPort := addr.Port
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
		// s.ExposedPortsToClients[exposedPort] = clientConn

		// Start listening for users
		go s.listenForUsers(userListener, exposedPort, clientConn)
		go s.handleClientConnection(clientConn)
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
		fmt.Println("User connection established")
		// s.ClientConnToUser[clientConn] = userConn

		serverTag()
		fmt.Printf("Saving User ID <%d> as connection <%x>\n", s.users, userConn)
		s.UserConn[s.users] = userConn
		s.UserId[userConn] = s.users

		serverTag()
		fmt.Printf("Sending ConnectionOpen for user <%d> to client\n", s.users)
		// Send ConnectionOpen msg to client with id
		msg := &pb.TestMessage{
			Type: pb.TestMessage_ConnectionOpen,
			Id:   s.users,
			Data: pb.TestMessage_ConnectionOpen.String(),
		}
		s.users++

		bytes, err := proto.Marshal(msg)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
		}

		serverTag()
		fmt.Printf("Sending: %s\n", bytes)
		_, err = clientConn.Write(bytes)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
		}

		go s.handleUserConnection(userConn, clientConn)
	}
}

func (s GoRpsServer) Stop() error {
	// for _, conn := range s.ExposedPortsToClients {
	// 	err := conn.Close()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return s.clientListener.Close()
}

func (s GoRpsServer) handleClientConnection(clientConn *net.TCPConn) {
	// Read data from clients
	for {
		serverTag()
		fmt.Println("Waiting for data from client...")

		// Blocks until we receive some data from client
		msg, err := helper.ReceiveProtobuf(clientConn)
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
			continue
		}

		serverTag()
		fmt.Printf("Received from client: %s, Writing to user<%d>\n", msg.Data, msg.Id)
		s.UserConn[msg.Id].Write([]byte(msg.Data))
	}
}

func (s GoRpsServer) handleUserConnection(userConn *net.TCPConn, clientConn *net.TCPConn) {
	for {
		// Read info from user

		// This is inside Generate Protobuf

		// bytes := make([]byte, 4096)
		// i, err := userConn.Read(bytes)
		// if err != nil {
		// 	serverTag()
		// 	fmt.Println(err.Error())
		// 	return
		// }

		// serverTag()
		// fmt.Printf("Read from user <%d>:%s\n", s.UserId[userConn], bytes)

		// msg := &pb.TestMessage{
		// 	Type: pb.TestMessage_Data,
		// 	Data: string(bytes[0:i]),
		// 	Id:   s.UserId[userConn],
		// }

		msg, err := helper.GenerateProtobuf(userConn, s.UserId[userConn])
		if err != nil {
			serverTag()
			fmt.Println(err.Error())
			continue
		}
		serverTag()
		fmt.Printf("Read from user <%d>:%s\n", s.UserId[userConn], msg.Data)

		// Forward data to associated client
		sendToClient(msg, clientConn)
	}
}

func sendToClient(msg *pb.TestMessage, clientConn *net.TCPConn) {
	out, err := proto.Marshal(msg)
	if err != nil {
		serverTag()
		fmt.Println(err.Error())
		return
	}
	serverTag()
	fmt.Printf("Forwarding to client\n")
	// Forward data to the associated client
	clientConn.Write(out)
}

func serverTag() {
	fmt.Print("Server: ")
}
