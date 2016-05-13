package server

import (
	"fmt"
	"github.com/Originate/go_rps/helper"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"strconv"
)

type GoRpsServer struct {
	HostIP net.IP

	UserConn map[int32]*net.TCPConn
	UserId   map[*net.TCPConn]int32

	clientToUsers  map[*net.TCPConn][]*net.TCPConn
	users          int32
	clientListener *net.TCPListener
}

func (s GoRpsServer) Start() (*net.TCPAddr, error) {
	s.users = 0
	if s.HostIP == nil {
		s.HostIP = net.IPv4(0, 0, 0, 0)
	}
	s.UserConn = make(map[int32]*net.TCPConn)
	s.UserId = make(map[*net.TCPConn]int32)
	s.clientToUsers = make(map[*net.TCPConn][]*net.TCPConn)

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

	// Listen for clients
	go s.listenForClients()

	ret, err := net.ResolveTCPAddr("tcp", s.clientListener.Addr().String())
	if err != nil {
		serverTag()
		// fmt.Printf("Error resolving address: %s\n", err.Error())
		return nil, err
	}
	return ret, nil
}

func (s GoRpsServer) listenForClients() {
	for {
		// Listen for a client to connect
		clientConn, err := s.clientListener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go s.listenToClients(clientConn)

		serverTag()
		// fmt.Printf("Client connected!\n")
		// Choose a random free port to expose to users
		address := &net.TCPAddr{
			IP:   s.HostIP,
			Port: 0,
		}

		// Create a listener for that port, and extract the chosen port
		userListener, err := net.ListenTCP("tcp", address)

		addr, err := net.ResolveTCPAddr("tcp", userListener.Addr().String())
		exposedPort := addr.Port
		// Start listening for users
		go s.listenForUsers(userListener, exposedPort, clientConn)

		portStr := strconv.Itoa(exposedPort)

		msg := &pb.TestMessage{
			Type: pb.TestMessage_ConnectionOpen,
			Data: portStr,
			Id:   -1,
		}
		bytes, err := proto.Marshal(msg)
		if err != nil {
			serverTag()
			// fmt.Println(err.Error())
		}

		// Tell the client what port is exposed to users for their connection
		clientConn.Write(bytes)
	}
}

func (s GoRpsServer) listenForUsers(userListener *net.TCPListener, exposedPort int, clientConn *net.TCPConn) {
	serverTag()
	// fmt.Printf("Listening for users on addr: %s...\n", userListener.Addr().String())
	for {
		// Listen for a user connection
		userConn, err := userListener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		serverTag()
		// fmt.Println("User connection established")

		serverTag()
		// fmt.Printf("Saving User ID <%d> as connection <%x>\n", s.users, userConn)
		addrStr := fmt.Sprintf("%x", userConn)
		id, err := strconv.ParseInt(addrStr[3:len(addrStr)-2], 16, 64)
		if err != nil {
			// fmt.Printf("Error converting addr to id: %s\n", err.Error())
			continue
		}
		id32 := int32(id)

		s.UserConn[id32] = userConn
		s.UserId[userConn] = id32
		s.clientToUsers[clientConn] = append(s.clientToUsers[clientConn], userConn)

		serverTag()
		// fmt.Printf("Sending ConnectionOpen for user <%d> to client\n", id32)

		// Send ConnectionOpen msg to client with id
		msg := &pb.TestMessage{
			Type: pb.TestMessage_ConnectionOpen,
			Id:   id32,
			Data: pb.TestMessage_ConnectionOpen.String(),
		}
		s.users++

		sendToClient(msg, clientConn)

		go s.listenToUser(userConn, clientConn)
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

func (s GoRpsServer) listenToClients(clientConn *net.TCPConn) {
	// Read data from clients
	for {
		// Blocks until we receive some data from client
		msg, err := helper.ReceiveProtobuf(clientConn)
		if err != nil {
			// if err == io.EOF {
			serverTag()
			// fmt.Printf("Client has disconnected.\n")
			s.clientDisconnected(clientConn)
			return
			// }
			serverTag()
			// fmt.Println(err.Error())
			continue
		}

		serverTag()
		// fmt.Printf("Received from client: %s, Writing to user<%d>\n", msg.Data, msg.Id)

		switch msg.Type {
		// Client told us that protected server has disconnected
		// We need to disconnect all users associated with this client
		case pb.TestMessage_ConnectionClose:
			{
				for _, userConn := range s.clientToUsers[clientConn] {
					err = userConn.Close()
					if err != nil {
						serverTag()
						// fmt.Printf("Error closing connection for user <%d>: %s\n", s.UserId[userConn], err.Error())
					}
				}
				err = clientConn.Close()
				if err != nil {
					serverTag()
					// fmt.Printf("Error closing connection for client: %s\n", err.Error())
				}
				break
			}
		case pb.TestMessage_Data:
			{
				s.UserConn[msg.Id].Write([]byte(msg.Data))
				break
			}
		}

	}
}

func (s GoRpsServer) listenToUser(userConn *net.TCPConn, clientConn *net.TCPConn) {
	userId := s.UserId[userConn]
	for {
		// Read info from user

		// This is inside Generate Protobuf

		// bytes := make([]byte, 4096)
		// i, err := userConn.Read(bytes)
		// if err != nil {
		// 	serverTag()
		// fmt.Println(err.Error())
		// 	return
		// }

		// serverTag()
		// fmt.Printf("Read from user <%d>:%s\n", s.UserId[userConn], bytes)

		// msg := &pb.TestMessage{
		// 	Type: pb.TestMessage_Data,
		// 	Data: string(bytes[0:i]),
		// 	Id:   s.UserId[userConn],
		// }

		msg, err := helper.GenerateProtobuf(userConn, userId)
		if err != nil {
			if err == io.EOF {
				serverTag()
				// fmt.Printf("User <%d> has disconnected.\n", userId)
				err = userConn.Close()
				if err != nil {
					// fmt.Printf("Error closing connection for user <%d>: %s\n", userId, err.Error())
					continue
				}
				s.userDisconnected(userId, clientConn)
				return
			}
			serverTag()
			// fmt.Printf("Error reading from userConn <%d>: %s\n", userId, err.Error())
			return
		}
		serverTag()
		// fmt.Printf("Read from user <%d>:%s\n", userId, msg.Data)

		// Forward data to associated client
		sendToClient(msg, clientConn)
	}
}

func (s GoRpsServer) userDisconnected(userId int32, clientConn *net.TCPConn) {
	msg := &pb.TestMessage{
		Type: pb.TestMessage_ConnectionClose,
		Data: pb.TestMessage_ConnectionClose.String(),
		Id:   userId,
	}
	sendToClient(msg, clientConn)
}

func (s GoRpsServer) clientDisconnected(clientConn *net.TCPConn) {
	// Disconnect all users associated with clientConn
	for _, userConn := range s.clientToUsers[clientConn] {
		err := userConn.Close()
		if err != nil {
			// fmt.Printf("Error closing connection for user <%d>\n", s.UserId[userConn])
		}
	}
}

func sendToClient(msg *pb.TestMessage, clientConn *net.TCPConn) {
	out, err := proto.Marshal(msg)
	if err != nil {
		serverTag()
		// fmt.Println(err.Error())
		return
	}
	serverTag()
	fmt.Printf("Forwarding to client: Id: %d, Data: %s\n", msg.Id, msg.Data)
	// Forward data to the associated client
	clientConn.Write(out)
}

func serverTag() {
	var _, _ = fmt.Print()
	// fmt.Print("Server: ")
}
