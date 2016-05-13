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
func (c *GoRpsClient) OpenTunnel(portToConnect int) (err error) {
	c.portToConnect = portToConnect
	c.ConnToProtectedServer = make(map[int32]*net.TCPConn)

	// Connect to rps server
	clientTag()
	// fmt.Printf("Dialing %s...\n", c.ServerTCPAddr.String())
	c.ConnToRpsServer, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
	if err != nil {
		clientTag()
		// fmt.Printf("Error dialing rps server: %s\n", err.Error())
		return err
	}

	// Read which port to use from rps server
	clientTag()
	// fmt.Printf("Waiting for a response port from rps server...\n")

	msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
	if err != nil {
		clientTag()
		// fmt.Printf("Error receiving from rps server initially: %s\n", err.Error())
		return err
	}

	c.ExposedPort, err = strconv.Atoi(msg.Data)
	if err != nil {
		clientTag()
		// fmt.Printf("Error converting port: %s\n", err.Error())
		return err
	}
	clientTag()
	// fmt.Printf("Port to use is: %d\n", c.ExposedPort)

	go c.listenToServer()

	return nil
}

func (c *GoRpsClient) Stop() (err error) {
	// err = c.ConnToRpsServer.Close()
	// if err != nil {
	// 	return err
	// }
	err = c.ConnToRpsServer.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, connToPS := range c.ConnToProtectedServer {
		err = connToPS.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *GoRpsClient) listenToServer() {
	for {
		msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
		if err != nil {
			clientTag()
			// fmt.Printf("Error receiving from rps server: %s\n", err.Error())
			return
		}

		clientTag()
		// fmt.Printf("Received msg from server for user <%d>: Data: %s Type:%s\n", msg.Id, msg.Data, msg.Type.String())

		connToPS, ok := c.ConnToProtectedServer[msg.Id]
		switch msg.Type {
		// Start a new connection to protected server
		case pb.TestMessage_ConnectionOpen:
			{
				clientTag()
				// fmt.Printf("First connection for user <%d>\n", msg.Id)
				if connToPS != nil {
					clientTag()
					// fmt.Printf("Connection for user <%d> already exists.\n", msg.Id)
				} else {
					c.openConnection(msg.Id)
				}
				break
			}
		case pb.TestMessage_ConnectionClose:
			{
				if ok {
					clientTag()
					// fmt.Printf("Closing connection for user <%d>\n", msg.Id)
					err = connToPS.Close()
					if err != nil {
						// fmt.Printf("Error closing connection for user <%d>\n", msg.Id)
						break
					}
					delete(c.ConnToProtectedServer, msg.Id)
				} else {
					clientTag()
					// fmt.Printf("connection to PS for user <%d> is nil\n", msg.Id)
				}
				break
			}
		case pb.TestMessage_Data:
			{
				if !ok {
					clientTag()
					// fmt.Printf("No connection for user <%d>, trying to establish one\n", msg.Id)
					c.openConnection(msg.Id)
					connToPS = c.ConnToProtectedServer[msg.Id]
				}
				// Forward data to protected server
				_, err = connToPS.Write([]byte(msg.Data))
				if err != nil {
					clientTag()
					// fmt.Printf("Error forwarding data to PS: %s\n", err.Error())
				}
				break
			}
		default:
		}
	}
}

func (c *GoRpsClient) listenToProtectedServer(id int32) {
	for {
		// fmt.Printf("Listening to protected server for user <%d>\n", id)
		currentConn, ok := c.ConnToProtectedServer[id]
		if !ok {
			clientTag()
			// fmt.Printf("Connection for user <%d> has closed. Will stop listening.\n", id)
			return
		}
		msg, err := helper.GenerateProtobuf(currentConn, id)
		if err != nil {
			if err == io.EOF {
				clientTag()
				// fmt.Printf("Local server has disconnected.\n")
				currentConn.Close()

				// Tell server that it has closed so server can close all users connected
				msg := &pb.TestMessage{
					Type: pb.TestMessage_ConnectionClose,
					Data: pb.TestMessage_ConnectionClose.String(),
				}

				bytes, err2 := proto.Marshal(msg)
				if err2 != nil {
					// fmt.Printf("Error marshalling msg: %s\n", err2.Error())
					return
				}
				c.ConnToRpsServer.Write(bytes)
				return
			}
			clientTag()
			// fmt.Printf("Connection to PS closed: %s\n", err.Error())
			return
		}

		// Send back to server
		clientTag()
		// fmt.Printf("Forwarding to rps server: %s for user <%d>\n", msg.Data, msg.Id)
		c.Send(msg)
	}
}

func (c *GoRpsClient) openConnection(id int32) {
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: c.portToConnect,
	}

	clientTag()
	// fmt.Printf("Opening new connection for user <%d>\nDialing %s...\n", id, address.String())
	var err error
	c.ConnToProtectedServer[id], err = net.DialTCP("tcp", nil, address)
	if err != nil {
		clientTag()
		// fmt.Println("error: " + err.Error())
	}
	go c.listenToProtectedServer(id)
}

func (c *GoRpsClient) Send(msg *pb.TestMessage) {
	out, err := proto.Marshal(msg)
	if err != nil {
		clientTag()
		// fmt.Printf("Error marshalling: %s\n", err.Error())
		return
	}
	_, err = c.ConnToRpsServer.Write(out)
	if err != nil {
		clientTag()
		// fmt.Printf("Error writing to rps server: %s\n", err.Error())
	}
}

func clientTag() {
	// fmt.Print("Client: ")
	var _, _ = fmt.Print()
}
