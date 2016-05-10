package client

import (
	"fmt"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
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
	fmt.Printf("Dialing %s...\n", c.ServerTCPAddr.String())
	var err error
	c.ConnToRpsServer, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
	if err != nil {
		clientTag()
		fmt.Println("error: " + err.Error())
		return nil, 0
	}

	// Read which port to use from rps server
	fmt.Printf("Waiting for a response from rps server...\n")
	bytes := make([]byte, 4096)
	i, err := c.ConnToRpsServer.Read(bytes)
	if err != nil {
		clientTag()
		fmt.Println(err.Error())
		return nil, 0
	}

	msg := &pb.TestMessage{}
	err = proto.Unmarshal(bytes[0:i], msg)
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
		bytes := make([]byte, 4096)
		i, err := c.ConnToRpsServer.Read(bytes)
		if err != nil {
			clientTag()
			fmt.Println(err.Error())
			return
		}
		clientTag()
		fmt.Printf("Received from server:\n%s\n", bytes)

		msg := &pb.TestMessage{}
		err = proto.Unmarshal(bytes[0:i], msg)
		if err != nil {
			clientTag()
			fmt.Println(err.Error())
			return
		}

		clientTag()
		fmt.Printf("Type: %s\n", msg.Type.String())

		switch msg.Type {
		// Start a new connection to protected server
		case pb.TestMessage_ConnectionOpen:
			if c.ConnToProtectedServer[msg.Id] != nil {
				break
			}
			c.openConnection(msg.Id)
			break
		case pb.TestMessage_ConnectionClose:
			break
		case pb.TestMessage_Data:
			{
				if c.ConnToProtectedServer[msg.Id] == nil {
					c.openConnection(msg.Id)
				}
				currentConn := c.ConnToProtectedServer[msg.Id]
				// Forward data to protected server
				currentConn.Write(bytes[0:i])

				// Read response and write back to server using protobuf
				bytes := make([]byte, 4096)
				i, err := currentConn.Read(bytes)
				if err != nil {
					clientTag()
					fmt.Println(err.Error())
					return
				}
				msg := &pb.TestMessage{
					Type: pb.TestMessage_Data,
					Id:   msg.Id,
					Data: string(bytes[0:i]),
				}

				// Send back to server
				c.Send(msg)
				break
			}
		default:
		}
	}
}

func (c GoRpsClient) openConnection(id int32) {
	address := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: c.portToConnect,
	}

	fmt.Printf("Opening new connection for user <%d>\n", id)
	fmt.Printf("Dialing %s...\n", address.String())
	var err error
	c.ConnToProtectedServer[id], err = net.DialTCP("tcp", nil, address)
	if err != nil {
		clientTag()
		fmt.Println("error: " + err.Error())
	}
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
