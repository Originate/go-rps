package client

import (
  "net"
  "fmt"
)

type GoRpsClient struct {
  ServerHost string  
}


func (c GoRpsClient) OpenTunnel(tunnelPort int) {
  conn, err := net.Dial("tcp", c.ServerHost)
  if err != nil {
    // error
    fmt.Println("error: " + err.Error())
    return
  }
  fmt.Fprintf(conn, "data...")
  
}