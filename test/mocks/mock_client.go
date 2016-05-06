package mocks

import (
	"net"
	"fmt"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
)

type MockClient struct {
  ServerTCPAddr *net.TCPAddr
  Conn *net.TCPConn
  port int

  // For testing only
  TestChannel chan string
}


func (c MockClient) OpenTunnel(tunnelPort int) *net.TCPConn{
  c.port = tunnelPort
  var err error
  c.Conn, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
  if err != nil {
    fmt.Println("error: " + err.Error())
    return nil
  }
  
  return c.Conn
}

func (c MockClient) Send(msg *pb.TestMessage) {
  out, err := proto.Marshal(msg)
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  c.Conn.Write(out)
}