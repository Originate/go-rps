package go_rps_test

import (
    . "github.com/onsi/ginkgo"
    // . "github.com/originate/go_rps/test/mocks"
    . "github.com/onsi/gomega"
    . "github.com/originate/go_rps/server"
    . "github.com/originate/go_rps/client"
    "net"
    "fmt"
    // pb "github.com/Originate/go_rps/protobuf"
)

var _ = Describe("GoRps Server", func() {
    var server GoRpsServer
    var client GoRpsClient

    BeforeEach(func() {
        testChannel := make(chan string)
        server = GoRpsServer {}
        serverTCPAddr, err := server.Start()
        Expect(err).NotTo(HaveOccurred())

        client = GoRpsClient {
            ServerTCPAddr: serverTCPAddr,
            TestChannel: testChannel,
        }
    })

    AfterEach(func() {
        
    })

    // It("should accept a connection from a client", func() {
    //     client.OpenTunnel(3000)
    //     Expect(<-server.TestChannel).To(Equal("1"))
    // })

    // Describe("Sending data using protocol buffers", func() {
    //     Context("from server to client", func() {
    //         It("should send data through the tunnel, eventually", func(done Done) {
    //             // message := &pb.TestMessage {
    //             //     Id: "1",
    //             //     Data: "hello world",
    //             //     Type: pb.TestMessage_Data,
    //             // }
    //             close(done)
    //         },10)
    //     })
    // })

    Describe("A user hitting the rps server", func() {
        Context("to access another protected server", func() {
            It("should forward user data to the client first, before the protected server", func() {
                _, exposedPort := client.OpenTunnel(3000)
                address := &net.TCPAddr {
                    IP: net.IPv4(127,0,0,1),
                    Port: exposedPort,
                }
                fmt.Printf("exposed addr is %s\n", address.String())
                userConn, err := net.DialTCP("tcp", nil, address)
                Expect(err).NotTo(HaveOccurred())
                userConn.Write([]byte("Hello world"))
                Expect(<-client.TestChannel).To(Equal("Hello world"))
            })
        })
    })
})


