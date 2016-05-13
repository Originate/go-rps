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

	clientToUsers        map[*net.TCPConn][]*net.TCPConn
	clientToUserListener map[*net.TCPConn]*net.TCPListener
	users                int32
	clientListener       *net.TCPListener
}

func (s *GoRpsServer) Start() (*net.TCPAddr, error) {
	s.users = 0
	if s.HostIP == nil {
		s.HostIP = net.IPv4(0, 0, 0, 0)
	}
	s.UserConn = make(map[int32]*net.TCPConn)
	s.UserId = make(map[*net.TCPConn]int32)
	s.clientToUsers = make(map[*net.TCPConn][]*net.TCPConn)
	s.clientToUserListener = make(map[*net.TCPConn]*net.TCPListener)

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
		// fmt.Printf("Error resolving address: %s\n", err.Error())
		return nil, err
	}
	return ret, nil
}

func (s *GoRpsServer) listenForClients() {
	fmt.Printf("Server listening for clients on: %s\n", s.clientListener.Addr().String())
	for {
		// Listen for a client to connect
		clientConn, err := s.clientListener.AcceptTCP()
		if err != nil {
			return
		}
		go s.listenToClients(clientConn)

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
			// fmt.Println(err.Error())
		}

		// Tell the client what port is exposed to users for their connection
		clientConn.Write(bytes)
	}
}

func (s *GoRpsServer) listenForUsers(userListener *net.TCPListener, exposedPort int, clientConn *net.TCPConn) {
	fmt.Printf("Server listening for users on: %s\n", userListener.Addr().String())
	s.clientToUserListener[clientConn] = userListener
	for {
		// Listen for a user connection
		userConn, err := userListener.AcceptTCP()

		if err != nil {
			return
		}

		// fmt.Printf("Accepted user connection (listener:<%x>) on: %s\n", userListener, userConn.LocalAddr().String())

		// fmt.Println("User connection established")

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
		fmt.Printf("Appending to clientToUsers[clientConn]\n")
		s.clientToUsers[clientConn] = append(s.clientToUsers[clientConn], userConn)
		fmt.Printf("Size of clientToUsers[clientConn]: %d\n", len(s.clientToUsers[clientConn]))
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

func (s *GoRpsServer) Stop() (err error) {
	for _, userListener := range s.clientToUserListener {
		err = userListener.Close()
		if err != nil {
			return err
		}
	}
	for clientConn, userConns := range s.clientToUsers {
		for _, userConn := range userConns {
			err = userConn.Close()
			if err != nil {
				fmt.Printf("Error closing user conn: %s\n", err.Error())
				return err
			}
		}
		err = clientConn.Close()
		if err != nil {
			fmt.Printf("Error closing client conn: %s\n", err.Error())
			return err
		}
	}

	s.clientToUsers = make(map[*net.TCPConn][]*net.TCPConn)
	s.clientToUserListener = make(map[*net.TCPConn]*net.TCPListener)
	return s.clientListener.Close()
}

func (s *GoRpsServer) listenToClients(clientConn *net.TCPConn) {
	// Read data from clients
	for {
		// Blocks until we receive some data from client
		msg, err := helper.ReceiveProtobuf(clientConn)
		if err != nil {
			if err == io.EOF {
				// fmt.Printf("Client has disconnected.\n")
				s.clientDisconnected(clientConn)
				return
			}
			fmt.Printf("Error receiving from client: %s\n", err.Error())
			return
		}

		// fmt.Printf("Received from client: %s, Writing to user<%d>\n", msg.Data, msg.Id)

		switch msg.Type {
		// Client told us that protected server has disconnected
		// We need to disconnect all users associated with this client
		case pb.TestMessage_ConnectionClose:
			{
				err = s.clientToUserListener[clientConn].Close()
				if err != nil {
					fmt.Printf("Error closing user listening: %s\n", err.Error())
				}
				delete(s.clientToUserListener, clientConn)

				fmt.Printf("Size of associated users: %d\n", len(s.clientToUsers[clientConn]))
				for _, userConn := range s.clientToUsers[clientConn] {
					fmt.Printf("Closing user <%d> connection\n", s.UserId[userConn])
					err = userConn.Close()
					if err != nil {
						fmt.Printf("Error closing connection for user <%d>: %s\n", s.UserId[userConn], err.Error())
					}
				}
				delete(s.clientToUsers, clientConn)

				fmt.Printf("Closing clientConn...\n")
				err = clientConn.Close()
				if err != nil {
					fmt.Printf("Error closing connection for client: %s\n", err.Error())
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

func (s *GoRpsServer) listenToUser(userConn *net.TCPConn, clientConn *net.TCPConn) {
	userId := s.UserId[userConn]
	for {
		msg, err := helper.GenerateProtobuf(userConn, userId)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("User <%d> has disconnected.\n", userId)
				err = userConn.Close()
				if err != nil {
					fmt.Printf("Error closing connection for user <%d>: %s\n", userId, err.Error())
					continue
				}
				fmt.Printf("User <%d> connection successfully closed.\n", userId)
				s.userDisconnected(userId, clientConn)
				return
			}
			fmt.Printf("Error receving from user: %s\n", err.Error())
			return
		}

		// Forward data to associated client
		err = sendToClient(msg, clientConn)
		if err != nil {
			fmt.Printf("Error forwarding data to client: %s\n", err.Error())
		}
	}
}

func (s *GoRpsServer) userDisconnected(userId int32, clientConn *net.TCPConn) {
	msg := &pb.TestMessage{
		Type: pb.TestMessage_ConnectionClose,
		Data: pb.TestMessage_ConnectionClose.String(),
		Id:   userId,
	}
	err := sendToClient(msg, clientConn)
	if err != nil {
		fmt.Printf("Error forwarding data to client: %s\n", err.Error())
	}
}

func (s *GoRpsServer) clientDisconnected(clientConn *net.TCPConn) {
	// Disconnect all users associated with clientConn
	for _, userConn := range s.clientToUsers[clientConn] {
		err := userConn.Close()
		if err != nil {
			fmt.Printf("Error closing connection for user <%d>\n", s.UserId[userConn])
		}
	}
}

func sendToClient(msg *pb.TestMessage, clientConn *net.TCPConn) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	// fmt.Printf("Forwarding to client: Id: %d, Data: %s\n", msg.Id, msg.Data)
	// Forward data to the associated client
	_, err = clientConn.Write(out)
	return err
}
