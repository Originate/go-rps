package go_rps_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/originate/go_rps/test/mocks"
	. "github.com/originate/go_rps/client"
 	. "github.com/onsi/gomega"
)

var _ = Describe("GoRps Client", func() {
  var server MockServer
  var client GoRpsClient
  
  BeforeEach(func() {
    tc := make(chan int)
    server = MockServer {
      TunnelChannel: tc,
    }
    serverHost, err := server.Start()
    Expect(err).NotTo(HaveOccurred())
    
    client = GoRpsClient {
      ServerHost: serverHost,
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
  
})
