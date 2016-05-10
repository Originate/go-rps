package go_rps_test

import (
	pb "github.com/Originate/go_rps/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/originate/go_rps/client"
	. "github.com/originate/go_rps/server"
)

var _ = Describe("GoRps Client", func() {
	var server GoRpsServer
	var client GoRpsClient

	BeforeEach(func() {
		server = GoRpsServer{}
		serverTCPAddr, err := server.Start()
		Expect(err).NotTo(HaveOccurred())

		client = GoRpsClient{
			ServerTCPAddr: serverTCPAddr,
		}
	})

	AfterEach(func() {
		// err := server.Stop()
		// Expect(err).NotTo(HaveOccurred())
	})

	Describe("Sending data using protocol buffers", func() {
		Context("From client to server", func() {
			It("should send data through the tunnel, eventually", func(done Done) {
				message := &pb.TestMessage{
					Id:   1,
					Data: "hello world",
					Type: pb.TestMessage_Data,
				}
				client.ConnToRpsServer, _ = client.OpenTunnel(3000)
				client.Send(message)
				close(done)
			}, 3)
		})
	})
})
