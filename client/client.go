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
	ConnToProtectedServer map[int32]*net.TCPConn // UserID -> connection to PS
	ExposedPort           int
	protectedServerPort   int
}

// Returns the port to hit on the server to reach the protected server
func (c *GoRpsClient) OpenTunnel(protectedServerPort int) (err error) {
	c.protectedServerPort = protectedServerPort
	fmt.Printf("protected Server Port is: %d\n", c.protectedServerPort)
	c.ConnToProtectedServer = make(map[int32]*net.TCPConn)

	// Connect to rps server
	fmt.Printf("Dialing rps server @: %s\n", c.ServerTCPAddr.String())
	c.ConnToRpsServer, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
	if err != nil {
		fmt.Printf("Error dialing rps server: %s\n", err.Error())
		return err
	}

	// Wait for rps server to tell us which port is exposed
	msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
	if err != nil {
		fmt.Printf("Error receiving exposed port from rps server: %s\n", err.Error())
		return err
	}

	c.ExposedPort, err = strconv.Atoi(string(msg.Data))
	if err != nil {
		return err
	}
	go c.handleServerConn()
	return nil
}

func (c *GoRpsClient) Stop() (err error) {
	// Tell server that client has stopped so server can close all users connected
	msg := &pb.TestMessage{
		Type: pb.TestMessage_ConnectionClose,
		Data: []byte(pb.TestMessage_ConnectionClose.String()),
		Id:   -1,
	}

	bytes, err2 := proto.Marshal(msg)
	if err2 != nil {
		fmt.Printf("Error marshalling msg: %s\n", err2.Error())
		return
	}
	c.ConnToRpsServer.Write(bytes)
	for _, connToPS := range c.ConnToProtectedServer {
		err = connToPS.Close()
		if err != nil {
			fmt.Printf("Error closing conn to ps: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func (c *GoRpsClient) handleServerConn() {
	for {
		// Blocks until we receive a message from the server
		msg, err := helper.ReceiveProtobuf(c.ConnToRpsServer)
		if err != nil {
			fmt.Printf("Error receiving from rps server: %s\n", err.Error())
			return
		}

		connToPS, ok := c.ConnToProtectedServer[msg.Id]
		switch msg.Type {
		// Start a new connection to protected server
		case pb.TestMessage_ConnectionOpen:
			{
				if connToPS == nil {
					c.openConnection(msg.Id)
				} else {
					fmt.Printf("Connection for user <%d> already exists.\n", msg.Id)
				}
				break
			}
		case pb.TestMessage_ConnectionClose:
			{
				if ok {
					fmt.Printf("Closing connection to PS for user <%d>\n", msg.Id)
					err = connToPS.Close()
					if err != nil {
						fmt.Printf("Error closing connection to PS for user <%d>\n", msg.Id)
						break
					}
					delete(c.ConnToProtectedServer, msg.Id)
				} else {
					fmt.Printf("connection to PS for user <%d> is nil\n", msg.Id)
				}
				break
			}
		case pb.TestMessage_Data:
			{
				if !ok {
					c.openConnection(msg.Id)
					connToPS = c.ConnToProtectedServer[msg.Id]
				}
				// Forward data to protected server
				_, err = connToPS.Write([]byte(msg.Data))
				if err != nil {
					fmt.Printf("Error forwarding data to PS: %s\n", err.Error())
				}
				break
			}
		default:
		}
	}
}

func (c *GoRpsClient) listenToProtectedServer(id int32) {
	for {
		currentConn, ok := c.ConnToProtectedServer[id]
		if !ok {
			fmt.Printf("Connection for user <%d> has closed.\n", id)
			return
		}
		msg, err := helper.GenerateProtobuf(currentConn, id)
		if err != nil {
			if err == io.EOF {
				currentConn.Close()

				// Tell server that it has closed so server can close all users connected
				msg := &pb.TestMessage{
					Type: pb.TestMessage_ConnectionClose,
					Data: []byte(pb.TestMessage_ConnectionClose.String()),
				}

				bytes, err2 := proto.Marshal(msg)
				if err2 != nil {
					fmt.Printf("Error marshalling msg: %s\n", err2.Error())
					return
				}
				c.ConnToRpsServer.Write(bytes)
				return
			}
			fmt.Printf("Connection to PS closed: %s\n", err.Error())
			return
		}

		// Send back to server
		c.Send(msg)
	}
}

func (c *GoRpsClient) openConnection(id int32) {
	fmt.Printf("in open connection protected Server Port is: %d\n", c.protectedServerPort)
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: c.protectedServerPort,
	}
	var err error
	fmt.Printf("Dialing protected server @: %s\n", address.String())
	c.ConnToProtectedServer[id], err = net.DialTCP("tcp", nil, address)
	if err != nil {
		fmt.Printf("Error open: " + err.Error())
		return
	}
	go c.listenToProtectedServer(id)
}

func (c *GoRpsClient) Send(msg *pb.TestMessage) {
	out, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("Error marshalling: %s\n", err.Error())
		return
	}
	_, err = c.ConnToRpsServer.Write(out)
	if err != nil {
		fmt.Printf("Error writing to rps server: %s\n", err.Error())
	}
}
