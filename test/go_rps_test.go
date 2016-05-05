package go_rps_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/originate/go_rps/test/mocks"
	. "github.com/originate/go_rps/client"
 	. "github.com/onsi/gomega"
  pb "github.com/Originate/go_rps/protobuf"
)

var _ = Describe("GoRps Client", func() {
  var server MockServer
  var client GoRpsClient
  
  tc := make(chan int, 1)
  rm := make(chan string, 1)
  BeforeEach(func() {
    server = MockServer {
      TunnelChannel: tc,
      ReceivedMessages: rm,
    }
    serverTCPAddr, err := server.Start()
    Expect(err).NotTo(HaveOccurred())
    
    client = GoRpsClient {
      ServerTCPAddr: serverTCPAddr,
    }
    client.OpenTunnel(3000)
  })
  
  AfterEach(func() {
    // err := server.Stop()
    // Expect(err).NotTo(HaveOccurred())
  })
  
  It("opens a tunnel to the mock server", func() {
    Expect(<-server.TunnelChannel).To(Equal(1))
  })
  
  Describe("Sending some stuff", func() {
    Context("Client to server", func() {
      It("should correctly send", func() {
        event := pb.TestMessage_Data
        message := pb.TestMessage {
          Id: 1234,
          Type: event,
          Data: "hello world",
        }
        client.Send(&message)
        Expect(<-server.ReceivedMessages).To(Equal("hello world"))
        })
      })
    })
})


