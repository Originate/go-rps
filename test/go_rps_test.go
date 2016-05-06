package go_rps_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/originate/go_rps/test/mocks"
	. "github.com/originate/go_rps/client"
 	. "github.com/onsi/gomega"
  pb "github.com/Originate/go_rps/protobuf"
  "fmt"
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
    
  })
  
  AfterEach(func() {
    // err := server.Stop()
    // Expect(err).NotTo(HaveOccurred())
  })
  
  It("opens a tunnel to the mock server", func() {
    client.OpenTunnel(3000)
    Expect(<-server.TunnelChannel).To(Equal(1))
  })
  
  // Describe("Sending some stuff", func() {
  //   Context("Client to server", func() {
      It("should correctly send", func() {
        event := pb.TestMessage_Data
        message := &pb.TestMessage {
          Id: "1",
          Data: "hello world",
          Type: event,
        }
        client.Conn = client.OpenTunnel(3000)
        if client.Conn == nil {
          fmt.Println("conn is nil")
        }
        client.Send(message)
        Expect(<-server.ReceivedMessages).To(Equal("hello world"))
        })
    //   })
    // })
})


