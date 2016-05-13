package go_rps_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/originate/go_rps/test/mocks"
	"fmt"
	// pb "github.com/Originate/go_rps/protobuf"
	// "github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	. "github.com/originate/go_rps/client"
	. "github.com/originate/go_rps/server"
	"net"
	// "time"
)

// Listen for new clients
func startProtectedServer(listener *net.TCPListener) {
	for {
		conn, _ := listener.AcceptTCP()
		fmt.Println("PS: Connected")
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
		fmt.Printf("PS: Received from client:%s\n", bytes[0:i])

		// Write back some fake data
		conn.Write(append([]byte("Received: "), bytes[0:i]...))
	}
}

var _ = Describe("GoRps", func() {
	var server GoRpsServer
	var client GoRpsClient

	protectedServerAddr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3000,
	}

	psListener, _ := net.ListenTCP("tcp", protectedServerAddr)
	go startProtectedServer(psListener)

	BeforeEach(func() {
		server = GoRpsServer{}
		serverTCPAddr, err := server.Start()
		Expect(err).NotTo(HaveOccurred())

		client = GoRpsClient{
			ServerTCPAddr: serverTCPAddr,
		}
	})

	AfterEach(func() {

	})

	Describe("A user hitting the rps server", func() {
		Context("to access the protected server", func() {
			It("should forward user data to the client, then the protected server", func(done Done) {
				_, exposedPort := client.OpenTunnel(3000)
				address := &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: exposedPort,
				}

				// Connect to Rps server
				userConn, err := net.DialTCP("tcp", nil, address)
				Expect(err).NotTo(HaveOccurred())

				// Send some data
				userConn.Write([]byte("Hello world"))

				// Read the response
				bytes := make([]byte, 4096)
				i, err := userConn.Read(bytes)
				Expect(err).NotTo(HaveOccurred())

				// Should be the response from the simulated protected server
				Expect(bytes[0:i]).To(Equal([]byte("Received: Hello world")))
				userConn.Close()
				close(done)
			}, 10)
		})
	})

	Describe("A user hitting the rps server", func() {
		Context("sending two messages", func() {
			It("should successfully get both messages to the protected server", func(done Done) {
				_, exposedPort := client.OpenTunnel(3000)
				address := &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: exposedPort,
				}

				// Connect to Rps server
				userConn, err := net.DialTCP("tcp", nil, address)
				Expect(err).NotTo(HaveOccurred())

				// Send msg 1
				userConn.Write([]byte("Message 1"))

				// Read the response
				bytes := make([]byte, 4096)
				i, err := userConn.Read(bytes)
				Expect(err).NotTo(HaveOccurred())
				// Should be the response from the simulated protected server
				Expect(bytes[0:i]).To(Equal([]byte("Received: Message 1")))

				// Send msg 2
				userConn.Write([]byte("Message 2"))

				// Read the response
				bytes = make([]byte, 4096)
				i, err = userConn.Read(bytes)
				Expect(err).NotTo(HaveOccurred())
				// Should be the response from the simulated protected server
				Expect(bytes[0:i]).To(Equal([]byte("Received: Message 2")))
				userConn.Close()
				close(done)
			}, 10)
		})
	})

	Describe("Two users hitting the rps server", func() {
		Context("to access the same protected server", func() {
			It("should forward users' datum to the client, then the protected server", func(done Done) {
				_, exposedPort := client.OpenTunnel(3000)
				address := &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: exposedPort,
				}

				// First user connects to Rps server
				userConn0, err := net.DialTCP("tcp", nil, address)
				Expect(err).NotTo(HaveOccurred())

				// Second user connects to Rps server
				userConn1, err := net.DialTCP("tcp", nil, address)
				Expect(err).NotTo(HaveOccurred())

				// First user sends some data
				userConn0.Write([]byte("Hello from user0"))

				// First user reads the response
				bytes := make([]byte, 4096)
				fmt.Printf("Before read 1\n")
				i, err := userConn0.Read(bytes)
				fmt.Printf("After read 1\n")
				Expect(err).NotTo(HaveOccurred())
				// Should be the response from the simulated protected server
				Expect(bytes[0:i]).To(Equal([]byte("Received: Hello from user0")))

				// Second user sends some data
				userConn1.Write([]byte("Hello from user1"))

				// Second user reads the response
				bytes = make([]byte, 4096)
				fmt.Printf("Before read 2\n")
				i, err = userConn1.Read(bytes)
				fmt.Printf("After read 2\n")
				Expect(err).NotTo(HaveOccurred())
				// Should be the response from the simulated protected server
				Expect(bytes[0:i]).To(Equal([]byte("Received: Hello from user1")))

				userConn0.Close()
				userConn1.Close()

				close(done)
			}, 10)
		})
	})
})
