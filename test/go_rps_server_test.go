package go_rps_test

import (
    . "github.com/onsi/ginkgo"
    // . "github.com/originate/go_rps/test/mocks"
    . "github.com/onsi/gomega"
    . "github.com/originate/go_rps/server"
    . "github.com/originate/go_rps/client"
    "net"
    "fmt"
    pb "github.com/Originate/go_rps/protobuf"
    "github.com/golang/protobuf/proto"
)

func startProtectedServer(listener *net.TCPListener) {
    for {
        conn, _ := listener.AcceptTCP()
        fmt.Println("PS: Connected")
        go handleConn(conn)
    }
}

// Simulate a server that reads data and returns something
func handleConn(conn *net.TCPConn) {
    for {
        // Read info from client
        bytes := make ([]byte, 4096)
        i, err := conn.Read(bytes)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        fmt.Printf("PS: received %s from client\n", bytes[0:i])

        conn.Write([]byte("HTTP 200 OK"))
    }
}

var _ = Describe("GoRps Server", func() {
    var server GoRpsServer
    var client GoRpsClient

    protectedServerAddr := &net.TCPAddr {
        IP: net.IPv4(127,0,0,1),
        Port: 3000,
    }

    psListener, _ := net.ListenTCP("tcp", protectedServerAddr)
    go startProtectedServer(psListener)

    BeforeEach(func() {
        serverTestChannel := make(chan string)
        clientTestChannel := make(chan string)
        server = GoRpsServer {
            TestChannel: serverTestChannel,
        }
        serverTCPAddr, err := server.Start()
        Expect(err).NotTo(HaveOccurred())

        client = GoRpsClient {
            ServerTCPAddr: serverTCPAddr,
            TestChannel: clientTestChannel,
        }
    })

    AfterEach(func() {
        
    })

    // It("should accept a connection from a client", func(done Done) {
    //     client.OpenTunnel(3000)
    //     Expect(<-server.TestChannel).To(Equal("1"))
    //     close(done)
    // }, 3)

    Describe("A user hitting the rps server", func() {
        Context("to access another protected server", func() {
            It("should forward user data to the client first, before the protected server", func(done Done) {
                _, exposedPort := client.OpenTunnel(3000)
                address := &net.TCPAddr {
                    IP: net.IPv4(127,0,0,1),
                    Port: exposedPort,
                }

                // Connect to Rps server
                userConn, err := net.DialTCP("tcp", nil, address)
                Expect(err).NotTo(HaveOccurred())
                userConn.Write([]byte("Hello world"))
                Expect(<-client.TestChannel).To(Equal("Hello world"))
                fmt.Printf("ADDRESS: %x\n", userConn)
                bytes := make ([]byte, 4096)
                i, err := userConn.Read(bytes)
                msg := &pb.TestMessage{}
                proto.Unmarshal(bytes[0:i], msg)

                fmt.Println("Successful read")
                Expect(err).NotTo(HaveOccurred())
                Expect(<-server.TestChannel).To(Equal("1"))
                Expect(<-server.TestChannel).To(Equal("HTTP 200 OK"))
                Expect(msg.Data).To(Equal("HTTP 200 OK"))
                close(done)
            }, 5)
        })
    })
})


