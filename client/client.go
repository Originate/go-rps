package client

import (
  "net"
  "fmt"
  pb "github.com/Originate/go_rps/protobuf"
  "github.com/golang/protobuf/proto"
)

type GoRpsClient struct {
  ServerHost string
  conn net.Conn
}


func (c GoRpsClient) OpenTunnel(tunnelPort int) {
  var err error
  c.conn, err = net.Dial("tcp", c.ServerHost)
  if err != nil {
    // error
    fmt.Println("error: " + err.Error())
    return
  }
  fmt.Fprintf(c.conn, "data...") 
}

func (c GoRpsClient) Send(msg *pb.TestMessage) {
  out, err := proto.Marshal(msg)
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  c.conn.Write(out)
}