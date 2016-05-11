package client

import (
	"fmt"
	"github.com/Originate/go_rps/helper"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"strconv"
)

type GoRpsClient struct {
	ServerTCPAddr         *net.TCPAddr
	ConnToRpsServer       *net.TCPConn
	portToConnect         int
	ExposedPort           int
	ConnToProtectedServer map[int32]*net.TCPConn
}

// Returns address of Rps server + port to hit on that server
func (c GoRpsClient) OpenTunnel(portToConnect int) (*net.TCPConn, int) {
	c.portToConnect = portToConnect
	c.ConnToProtectedServer = make(map[int32]*net.TCPConn)

	// Connect to rps server
	clientTag()
	fmt.Printf("Dialing %s...\n", c.ServerTCPAddr.String())
	var err error
	c.ConnToRpsServer, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
	if err != nil {
		clientTag()
		fmt.Println("error: " + err.Error())
		return nil, 0
	}

	// Read which port to use from rps server
	clientTag()
	fmt.Printf("Waiting for a response from rps server...\n")

	msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
	if err != nil {
		clientTag()
		fmt.Println(err.Error())
		return nil, 0
	}

	c.ExposedPort, err = strconv.Atoi(msg.Data)
	if err != nil {
		clientTag()
		fmt.Println(err.Error())
		return nil, 0
	}
	clientTag()
	fmt.Printf("Port to use is: %d\n", c.ExposedPort)

	go c.listenToServer()

	return c.ConnToRpsServer, c.ExposedPort
}

func (c GoRpsClient) listenToServer() {
	for {
		clientTag()
		fmt.Println("Waiting for data from server...")
		msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
		if err != nil {
			clientTag()
			fmt.Println(err.Error())
			return
		}

		clientTag()
		fmt.Printf("Received msg from server:\nData:\n%s\nType:%s\n", msg.Data, msg.Type.String())

		currentConn := c.ConnToProtectedServer[msg.Id]
		switch msg.Type {
		// Start a new connection to protected server
		case pb.TestMessage_ConnectionOpen:
			{
				clientTag()
				fmt.Printf("First connection for user <%d>\n", msg.Id)
				if currentConn != nil {
					clientTag()
					fmt.Printf("Connection for user <%d> already exists.\n", msg.Id)
				} else {
					c.openConnection(msg.Id)
				}
				break
			}
		case pb.TestMessage_ConnectionClose:
			{
				if currentConn != nil {
					clientTag()
					fmt.Printf("Closing connection for user <%d>\n", msg.Id)
					err = currentConn.Close()
					if err != nil {
						fmt.Printf("Error closing connection for user <%d>\n", msg.Id)
						break
					}
					delete(c.ConnToProtectedServer, msg.Id)
				}
				break
			}
		case pb.TestMessage_Data:
			{
				if currentConn == nil {
					clientTag()
					fmt.Printf("No connection for user <%d>, trying to establish one\n", msg.Id)
					c.openConnection(msg.Id)
					currentConn = c.ConnToProtectedServer[msg.Id]
				}
				// Forward data to protected server
				currentConn.Write([]byte(msg.Data))
				break
			}
		default:
		}
	}
}

func (c GoRpsClient) listenToProtectedServer(id int32) {
	for {
		clientTag()
		fmt.Printf("Listening to protected server...\n")
		currentConn := c.ConnToProtectedServer[id]
		if currentConn == nil {
			clientTag()
			fmt.Printf("Connection for user <%d> has closed. Will stop listening.\n", id)
			return
		}
		msg, err := helper.GenerateProtobuf(currentConn, id)
		if err != nil {
			if err == io.EOF {
				clientTag()
				fmt.Printf("Local server has disconnected.\n")
				currentConn.Close()

				// Tell server that it has closed so server can close all users connected
				msg := &pb.TestMessage{
					Type: pb.TestMessage_ConnectionClose,
					Data: pb.TestMessage_ConnectionClose.String(),
				}

				bytes, err := proto.Marshal(msg)
				if err != nil {
					fmt.Printf(err.Error())
					return
				}
				c.ConnToRpsServer.Write(bytes)
				return
			}
			clientTag()
			fmt.Println(err.Error())
			return
		}

		// Send back to server
		c.Send(msg)
	}
}

func (c GoRpsClient) openConnection(id int32) {
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: c.portToConnect,
	}

	clientTag()
	fmt.Printf("Opening new connection for user <%d>\nDialing %s...\n", id, address.String())
	var err error
	c.ConnToProtectedServer[id], err = net.DialTCP("tcp", nil, address)
	if err != nil {
		clientTag()
		fmt.Println("error: " + err.Error())
	}
	go c.listenToProtectedServer(id)
}

func (c GoRpsClient) Send(msg *pb.TestMessage) {
	out, err := proto.Marshal(msg)
	if err != nil {
		clientTag()
		fmt.Println(err.Error())
		return
	}
	c.ConnToRpsServer.Write(out)
}

func clientTag() {
	fmt.Print("Client: ")
}
