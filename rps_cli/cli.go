package main

import (
	"fmt"
	. "github.com/Originate/go_rps/client"
	"github.com/codegangsta/cli"
	"net"
	"os"
	"strconv"
)

func main() {
	app := cli.NewApp()
	app.Name = "rps_cli"
	app.Usage = "Expose a local server hidden behind a firewall"
	app.Action = func(c *cli.Context) error {
		portStr := c.Args()[0]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Printf("Invalid port: %s\n", portStr)
			return nil
		}
		fmt.Printf("Exposing whatever is currently running on port: %d\n", port)

		serverTCPAddrStr := c.Args()[1]
		serverTCPAddr, err := net.ResolveTCPAddr("tcp", serverTCPAddrStr)
		fmt.Printf("Connecting to rps server @: %s\n", serverTCPAddrStr)
		if err != nil {
			fmt.Printf("Invalid server address: %s\n", serverTCPAddrStr)
			return nil
		}

		client := GoRpsClient{
			ServerTCPAddr: serverTCPAddr,
		}

		conn, _ := client.OpenTunnel(port)
		if conn == nil {
			fmt.Printf("Unable to open tunnel.\n")
			return nil
		}
		fmt.Printf("Tunnel opened\n")
		select {}
	}
	app.Run(os.Args)
}
