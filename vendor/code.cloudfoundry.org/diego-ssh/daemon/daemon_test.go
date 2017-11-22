package daemon_test

import (
	"errors"
	"net"

	"code.cloudfoundry.org/diego-ssh/daemon"
	"code.cloudfoundry.org/diego-ssh/handlers"
	"code.cloudfoundry.org/diego-ssh/handlers/fake_handlers"
	"code.cloudfoundry.org/diego-ssh/test_helpers"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_net"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Daemon", func() {
	var (
		logger lager.Logger
		sshd   *daemon.Daemon

		serverSSHConfig *ssh.ServerConfig
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		serverSSHConfig = &ssh.ServerConfig{
			NoClientAuth: true,
		}
		serverSSHConfig.AddHostKey(TestHostKey)
	})

	Describe("HandleConnection", func() {
		var fakeConn *fake_net.FakeConn

		Context("when the function returns", func() {
			BeforeEach(func() {
				fakeConn = &fake_net.FakeConn{}
				fakeConn.ReadReturns(0, errors.New("oops"))

				sshd = daemon.New(logger, serverSSHConfig, nil, nil)
			})

			It("closes the connection", func() {
				sshd.HandleConnection(fakeConn)
				Expect(fakeConn.CloseCallCount()).To(BeNumerically(">=", 1))
			})
		})

		Context("when an ssh client connects", func() {
			var (
				serverNetConn net.Conn
				clientNetConn net.Conn

				clientConn     ssh.Conn
				clientChannels <-chan ssh.NewChannel
				clientRequests <-chan *ssh.Request
				clientConnErr  error

				client *ssh.Client
			)

			BeforeEach(func() {
				serverSSHConfig.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
					return nil, nil
				}

				serverNetConn, clientNetConn = test_helpers.Pipe()

				clientConfig := &ssh.ClientConfig{
					User: "username",
					Auth: []ssh.AuthMethod{
						ssh.Password("secret"),
					},
				}

				sshd = daemon.New(logger, serverSSHConfig, nil, nil)
				go sshd.HandleConnection(serverNetConn)

				clientConn, clientChannels, clientRequests, clientConnErr = ssh.NewClientConn(clientNetConn, "0.0.0.0", clientConfig)
				Expect(clientConnErr).NotTo(HaveOccurred())

				client = ssh.NewClient(clientConn, clientChannels, clientRequests)
			})

			AfterEach(func() {
				if client != nil {
					client.Close()
				}
			})

			It("performs a handshake", func() {
				Expect(clientConnErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("handleGlobalRequests", func() {
		var (
			globalRequestHandlers map[string]handlers.GlobalRequestHandler

			fakeHandler *fake_handlers.FakeGlobalRequestHandler
			client      *ssh.Client
		)

		BeforeEach(func() {
			fakeHandler = &fake_handlers.FakeGlobalRequestHandler{}
			globalRequestHandlers = map[string]handlers.GlobalRequestHandler{
				"known-handler": fakeHandler,
			}

			serverNetConn, clientNetConn := test_helpers.Pipe()

			sshd = daemon.New(logger, serverSSHConfig, globalRequestHandlers, nil)
			go sshd.HandleConnection(serverNetConn)

			client = test_helpers.NewClient(clientNetConn, nil)
		})

		AfterEach(func() {
			client.Close()
		})

		Context("when a global request is recevied", func() {
			var (
				accepted   bool
				requestErr error

				name      string
				wantReply bool
			)

			JustBeforeEach(func() {
				accepted, _, requestErr = client.SendRequest(name, wantReply, []byte("payload"))
			})

			Context("and there is an associated handler", func() {
				BeforeEach(func() {
					name = "known-handler"
					wantReply = true

					fakeHandler.HandleRequestStub = func(logger lager.Logger, request *ssh.Request) {
						request.Reply(true, []byte("response"))
					}
				})

				It("calls the handler to handle the request", func() {
					Eventually(fakeHandler.HandleRequestCallCount).Should(Equal(1))
				})

				It("does not reject the request as unknown", func() {
					Expect(requestErr).NotTo(HaveOccurred())
					Expect(accepted).To(BeTrue())
				})
			})

			Context("and there is not an associated handler", func() {
				Context("when WantReply is true", func() {
					BeforeEach(func() {
						name = "unknown-handler"
						wantReply = true
					})

					It("rejects the request", func() {
						Expect(requestErr).NotTo(HaveOccurred())
						Expect(accepted).To(BeFalse())
					})
				})
			})
		})
	})

	Describe("handleNewChannels", func() {
		var newChannelHandlers map[string]handlers.NewChannelHandler
		var fakeHandler *fake_handlers.FakeNewChannelHandler
		var client *ssh.Client

		BeforeEach(func() {
			fakeHandler = &fake_handlers.FakeNewChannelHandler{}
			newChannelHandlers = map[string]handlers.NewChannelHandler{
				"known-channel-type": fakeHandler,
			}

			serverNetConn, clientNetConn := test_helpers.Pipe()

			sshd = daemon.New(logger, serverSSHConfig, nil, newChannelHandlers)
			go sshd.HandleConnection(serverNetConn)

			client = test_helpers.NewClient(clientNetConn, nil)
		})

		AfterEach(func() {
			client.Close()
		})

		Context("when a new channel request is received", func() {
			var (
				channelType string

				openError error
			)

			JustBeforeEach(func() {
				_, _, openError = client.OpenChannel(channelType, []byte("extra-data"))
			})

			Context("and there is an associated handler", func() {
				BeforeEach(func() {
					channelType = "known-channel-type"

					fakeHandler.HandleNewChannelStub = func(logger lager.Logger, newChannel ssh.NewChannel) {
						ch, _, err := newChannel.Accept()
						Expect(err).NotTo(HaveOccurred())
						ch.Close()
					}
				})

				It("calls the handler to process the new channel request", func() {
					Expect(fakeHandler.HandleNewChannelCallCount()).To(Equal(1))

					logger, actualChannel := fakeHandler.HandleNewChannelArgsForCall(0)
					Expect(logger).NotTo(BeNil())

					Expect(actualChannel.ChannelType()).To(Equal("known-channel-type"))
					Expect(actualChannel.ExtraData()).To(Equal([]byte("extra-data")))
				})
			})

			Context("and there is not an associated handler", func() {
				BeforeEach(func() {
					channelType = "unknown-channel-type"
				})

				It("rejects the new channel request", func() {
					Expect(openError).To(HaveOccurred())

					channelError, ok := openError.(*ssh.OpenChannelError)
					Expect(ok).To(BeTrue())

					Expect(channelError.Reason).To(Equal(ssh.UnknownChannelType))
					Expect(channelError.Message).To(Equal("unknown-channel-type"))
				})
			})
		})
	})
})
