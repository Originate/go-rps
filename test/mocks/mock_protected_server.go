package mocks

import (
	"fmt"
	"net"
)

// Listen for new clients
func StartProtectedServer(port int) {
	protectedServerAddr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	}

	psListener, _ := net.ListenTCP("tcp", protectedServerAddr)
	go listenForConn(psListener)

}

func listenForConn(listener *net.TCPListener) {
	for {
		conn, _ := listener.AcceptTCP()
		go handleConn(conn)
	}
}

// Simulate a simple server that reads data and returns something
// This is the server that will be protected and require a proxy to access it
func handleConn(conn *net.TCPConn) {
	for {
		// Read info from client
		bytes := make([]byte, 4096)
		i, err := conn.Read(bytes)
		if err != nil {
			fmt.Printf("PS: %s\n", err.Error())
			return
		}
		// fmt.Printf("PS: Received from client:%s\n", bytes[0:i])

		// Write back some fake data
		conn.Write(append([]byte("Received: "), bytes[0:i]...))
	}
}
