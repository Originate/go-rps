package client

import (
    "net"
    "fmt"
    pb "github.com/Originate/go_rps/protobuf"
    "github.com/golang/protobuf/proto"
    "strconv"
)

type GoRpsClient struct {
    ServerTCPAddr *net.TCPAddr
    ConnToRpsServer *net.TCPConn
    portToConnect int
    ExposedPort int
    ConnToProtectedServer *net.TCPConn

    // For testing only
    TestChannel chan string
}

// Returns address of Rps server + port to hit on that server
func (c GoRpsClient) OpenTunnel(portToConnect int) (*net.TCPConn, int) {
    // Connect to protected server
    c.portToConnect = portToConnect
    address := &net.TCPAddr {
        IP: net.IPv4(127,0,0,1),
        Port: c.portToConnect,
    }
    var err error
    c.ConnToProtectedServer, err = net.DialTCP("tcp", nil, address)
    if err != nil {
        clientTag()
        fmt.Println("error: " + err.Error())
        return nil, 0
    }

    // Connect to rps server
    c.ConnToRpsServer, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
    if err != nil {
        clientTag()
        fmt.Println("error: " + err.Error())
        return nil, 0
    }
    bytes := make ([]byte, 4096)
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
        bytes := make ([]byte, 4096)
        i, err := c.ConnToRpsServer.Read(bytes)
        if err != nil {
            clientTag()
            fmt.Println(err.Error())
            return
        }
        clientTag()
        fmt.Printf("Received %s from server\n", bytes)

        msg := &pb.TestMessage{}
        err = proto.Unmarshal(bytes[0:i], msg)
        if err != nil {
            clientTag()
            fmt.Println(err.Error())
            return
        }

        switch msg.Type {
            case pb.TestMessage_ConnectionOpen: break
            case pb.TestMessage_ConnectionClose: break
            case pb.TestMessage_Data: {
                c.TestChannel <- msg.Data 
                // Forward data to protected server
                c.ConnToProtectedServer.Write(bytes[0:i])

                // Read response and write back to server using protobuf
                bytes := make([]byte, 4096)
                i, err := c.ConnToProtectedServer.Read(bytes)
                if err != nil {
                    clientTag()
                    fmt.Println(err.Error())
                    return
                }
                msg := &pb.TestMessage{
                    Type: pb.TestMessage_Data,
                    Data: string(bytes[0:i]),
                }
                
                c.Send(msg)
                break
            }
            default: 
        }
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