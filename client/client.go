package client

import (
  "net"
  "fmt"
  "bytes"
  // "bufio"
  pb "github.com/Originate/go_rps/protobuf"
  "github.com/golang/protobuf/proto"
)

type GoRpsClient struct {
  ServerTCPAddr *net.TCPAddr
  Conn *net.TCPConn
  port int
}


func (c GoRpsClient) OpenTunnel(tunnelPort int) *net.TCPConn{
  c.port = tunnelPort
  var err error
  c.Conn, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
  if err != nil {
    // error
    fmt.Println("error: " + err.Error())
    return nil
  }
  fmt.Println("Opening tunnel to " + c.ServerTCPAddr.String())
  if c.Conn == nil {
    fmt.Println("conn is nil inside")
  }
  return c.Conn
}

func (c GoRpsClient) Send(msg *pb.TestMessage) {
  out, err := proto.Marshal(msg)
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  fmt.Printf("Sent: %s---\n",out)
  buf := &bytes.Buffer{}
  buf.Write(out)
  if _, err := c.Conn.Write(buf.Bytes()); err != nil {
    fmt.Println(err.Error())
  }
}