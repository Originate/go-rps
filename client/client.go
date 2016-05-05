package client

import (
  "net"
  "fmt"
  // "bufio"
  pb "github.com/Originate/go_rps/protobuf"
  "github.com/golang/protobuf/proto"
)

type GoRpsClient struct {
  ServerTCPAddr *net.TCPAddr
  conn *net.TCPConn
  port int
}


func (c GoRpsClient) OpenTunnel(tunnelPort int) {
  c.port = tunnelPort
  var err error
  c.conn, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
  if err != nil {
    // error
    fmt.Println("error: " + err.Error())
    return
  }
  fmt.Println("Opening tunnel to " + c.ServerTCPAddr.String())
  if c.conn == nil {
    fmt.Println("0")
  }

  // fmt.Fprintf(c.conn, "tunnel open msg")//c.conn.Write([]byte("tunnel open msg"))
  
}

func (c GoRpsClient) Send(msg *pb.TestMessage) {
  out, err := proto.Marshal(msg)
  if err != nil {
    fmt.Println(err.Error())
    return
  }

  fmt.Println("Sending data:")
  fmt.Println(string(out))
  
  if c.conn == nil {
    fmt.Println("2")
    c.conn, err = net.DialTCP("tcp", nil, c.ServerTCPAddr)
    if err != nil {
      // error
      fmt.Println("error: " + err.Error())
      return
    }
    fmt.Println("Opening tunnel to " + c.ServerTCPAddr.String())
    if c.conn == nil {
      fmt.Println("3")
    }
    
  }
  fmt.Fprintf(c.conn, string(out))
  // fmt.Fprintf(c.conn, string(out))
  // status, err := bufio.NewReader(c.conn).ReadString('\n')
  // if err != nil {
  //   fmt.Println(err.Error())
  // }
  // fmt.Println(string(status))
}