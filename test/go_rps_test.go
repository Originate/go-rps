package go_rps_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/originate/go_rps/test/mocks"
	"fmt"
	pb "github.com/Originate/go_rps/protobuf"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	. "github.com/originate/go_rps/client"
	. "github.com/originate/go_rps/server"
	"net"
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
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("PS: Received from client:\n%s\n", bytes[0:i])

		// Write back some fake data
		conn.Write([]byte("HTTP 200 OK"))
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
		Context("to access another protected server", func() {
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
				Expect(bytes[0:i]).To(Equal([]byte("HTTP 200 OK")))
				close(done)
			}, 5)
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

				// Send some data
				userConn.Write([]byte("Hello world"))

				userConn.Write([]byte("Hello world again"))

				// Read the response
				bytes := make([]byte, 4096)
				i, err := userConn.Read(bytes)
				msg := &pb.TestMessage{}
				proto.Unmarshal(bytes[0:i], msg)
				Expect(err).NotTo(HaveOccurred())

				// // Should be the response from the simulated protected server
				// Expect(msg.Data).To(Equal("HTTP 200 OK"))
				close(done)
			}, 5)
		})
	})

	// Describe("Two users hitting the rps server", func() {
	// 	Context("to access the same protected server", func() {
	// 		It("should forward user datum to the client, then the protected server", func(done Done) {
	// 			_, exposedPort := client.OpenTunnel(3000)
	// 			address := &net.TCPAddr{
	// 				IP:   net.IPv4(127, 0, 0, 1),
	// 				Port: exposedPort,
	// 			}

	// 			// First user connects to Rps server
	// 			userConn1, err := net.DialTCP("tcp", nil, address)
	// 			Expect(err).NotTo(HaveOccurred())

	// 			// Second user connects to Rps server
	// 			userConn2, err := net.DialTCP("tcp", nil, address)
	// 			Expect(err).NotTo(HaveOccurred())

	// 			// First user sends some data
	// 			userConn1.Write([]byte("Hello world from user1"))
	// 			userConn1.Write([]byte("Hello world again from user1"))

	// 			// Second user sends some data
	// 			userConn2.Write([]byte("Hello world from user2"))

	// 			// First user reads the response
	// 			bytes := make([]byte, 4096)
	// 			i, err := userConn1.Read(bytes)
	// 			msg := &pb.TestMessage{}
	// 			proto.Unmarshal(bytes[0:i], msg)
	// 			Expect(err).NotTo(HaveOccurred())

	// 			// Should be the response from the simulated protected server
	// 			// Expect(msg.Data).To(Equal("HTTP 200 OK"))

	// 			// Second user reads the response
	// 			bytes = make([]byte, 4096)
	// 			i, err = userConn2.Read(bytes)
	// 			proto.Unmarshal(bytes[0:i], msg)
	// 			Expect(err).NotTo(HaveOccurred())

	// 			// Should be the response from the simulated protected server
	// 			// Expect(msg.Data).To(Equal("HTTP 200 OK"))
	// 			close(done)
	// 		}, 5)
	//	})
})
