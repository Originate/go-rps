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
    Conn *net.TCPConn
    portToConnect int
    ExposedPort int

    // For testing only
    TestChannel chan string
}


func (c GoRpsClient) OpenTunnel(portToConnect int) (*net.TCPConn, int) {
    c.portToConnect = portToConnect
    var err error
    c.Conn, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
    if err != nil {
        fmt.Println("error: " + err.Error())
        return nil, 0
    }
    bytes := make ([]byte, 4096)
    i, err := c.Conn.Read(bytes)
    if err != nil {
        fmt.Println(err.Error())
        return nil, 0
    }
    fmt.Printf("Initial Received %s from server\n", bytes)
    msg := &pb.TestMessage{}
    err = proto.Unmarshal(bytes[0:i], msg)
    if err != nil {
        fmt.Println(err.Error())
        return nil, 0
    }
    c.ExposedPort, err = strconv.Atoi(msg.Data)
    if err != nil {
        fmt.Println(err.Error())
        return nil, 0
    }
    fmt.Printf("Got response, port to use is: %d\n", c.ExposedPort)

    go c.listenToServer()

    return c.Conn, c.ExposedPort
}

func (c GoRpsClient) listenToServer() {
    for {
        fmt.Println("Waiting for data from server...")
        bytes := make ([]byte, 4096)
        i, err := c.Conn.Read(bytes)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        fmt.Printf("Received %s from server\n", bytes)

        msg := &pb.TestMessage{}
        err = proto.Unmarshal(bytes[0:i], msg)
        if err != nil {
            fmt.Println(err.Error())
            return
        }

        switch msg.Type {
            case pb.TestMessage_ConnectionOpen: break
            case pb.TestMessage_ConnectionClose: break
            case pb.TestMessage_Data: {
                c.TestChannel <- msg.Data 
                break
            }
            default: 
        }
    }
}

func (c GoRpsClient) Send(msg *pb.TestMessage) {
    out, err := proto.Marshal(msg)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    c.Conn.Write(out)
}