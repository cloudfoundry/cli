package proxy_test

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"

	"code.cloudfoundry.org/diego-ssh/authenticators/fake_authenticators"
	"code.cloudfoundry.org/diego-ssh/daemon"
	"code.cloudfoundry.org/diego-ssh/handlers"
	"code.cloudfoundry.org/diego-ssh/handlers/fake_handlers"
	"code.cloudfoundry.org/diego-ssh/helpers"
	"code.cloudfoundry.org/diego-ssh/proxy"
	"code.cloudfoundry.org/diego-ssh/server"
	server_fakes "code.cloudfoundry.org/diego-ssh/server/fakes"
	"code.cloudfoundry.org/diego-ssh/test_helpers"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_net"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	fake_logs "github.com/cloudfoundry/dropsonde/log_sender/fake"
	"github.com/cloudfoundry/dropsonde/logs"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	"github.com/cloudfoundry/dropsonde/metrics"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Proxy", func() {
	var (
		logger        lager.Logger
		fakeLogSender *fake_logs.FakeLogSender
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("HandleConnection", func() {
		var (
			proxyAuthenticator *fake_authenticators.FakePasswordAuthenticator
			proxySSHConfig     *ssh.ServerConfig
			sshProxy           *proxy.Proxy

			daemonTargetConfig          proxy.TargetConfig
			daemonAuthenticator         *fake_authenticators.FakePasswordAuthenticator
			daemonSSHConfig             *ssh.ServerConfig
			daemonGlobalRequestHandlers map[string]handlers.GlobalRequestHandler
			daemonNewChannelHandlers    map[string]handlers.NewChannelHandler
			sshDaemon                   *daemon.Daemon

			proxyListener net.Listener
			sshdListener  net.Listener

			proxyAddress  string
			daemonAddress string

			proxyServer *server.Server
			sshdServer  *server.Server

			proxyDone  chan struct{}
			daemonDone chan struct{}
		)

		BeforeEach(func() {
			proxyDone = make(chan struct{})
			daemonDone = make(chan struct{})

			fakeLogSender = fake_logs.NewFakeLogSender()
			logs.Initialize(fakeLogSender)

			proxyAuthenticator = &fake_authenticators.FakePasswordAuthenticator{}

			proxySSHConfig = &ssh.ServerConfig{}
			proxySSHConfig.PasswordCallback = proxyAuthenticator.Authenticate
			proxySSHConfig.AddHostKey(TestHostKey)

			daemonAuthenticator = &fake_authenticators.FakePasswordAuthenticator{}
			daemonAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)

			daemonSSHConfig = &ssh.ServerConfig{}
			daemonSSHConfig.PasswordCallback = daemonAuthenticator.Authenticate
			daemonSSHConfig.AddHostKey(TestHostKey)
			daemonGlobalRequestHandlers = map[string]handlers.GlobalRequestHandler{}
			daemonNewChannelHandlers = map[string]handlers.NewChannelHandler{}

			var err error
			proxyListener, err = net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())
			proxyAddress = proxyListener.Addr().String()

			sshdListener, err = net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())
			daemonAddress = sshdListener.Addr().String()

			daemonTargetConfig = proxy.TargetConfig{
				Address:         daemonAddress,
				HostFingerprint: helpers.MD5Fingerprint(TestHostKey.PublicKey()),
				User:            "some-user",
				Password:        "fake-some-password",
			}

			targetConfigJson, err := json.Marshal(daemonTargetConfig)
			Expect(err).NotTo(HaveOccurred())

			logMessageJson, err := json.Marshal(proxy.LogMessage{
				Guid:    "a-guid",
				Message: "a-message",
				Index:   1,
			})
			Expect(err).NotTo(HaveOccurred())

			permissions := &ssh.Permissions{
				CriticalOptions: map[string]string{
					"proxy-target-config": string(targetConfigJson),
					"log-message":         string(logMessageJson),
				},
			}
			proxyAuthenticator.AuthenticateReturns(permissions, nil)
		})

		JustBeforeEach(func() {
			sshProxy = proxy.New(logger.Session("proxy"), proxySSHConfig)
			proxyServer = server.NewServer(logger.Session("proxy-server"), "", sshProxy)
			proxyServer.SetListener(proxyListener)
			go func() {
				proxyServer.Serve()
				close(proxyDone)
			}()

			sshDaemon = daemon.New(logger.Session("sshd"), daemonSSHConfig, daemonGlobalRequestHandlers, daemonNewChannelHandlers)
			sshdServer = server.NewServer(logger.Session("sshd-server"), "", sshDaemon)
			sshdServer.SetListener(sshdListener)
			go func() {
				sshdServer.Serve()
				close(daemonDone)
			}()
		})

		AfterEach(func() {
			proxyServer.Shutdown()
			sshdServer.Shutdown()

			Eventually(proxyDone).Should(BeClosed())
			Eventually(daemonDone).Should(BeClosed())
		})

		Context("when a new connection arrives", func() {
			var clientConfig *ssh.ClientConfig

			BeforeEach(func() {
				clientConfig = &ssh.ClientConfig{
					User: "diego:some-instance-guid",
					Auth: []ssh.AuthMethod{
						ssh.Password("diego-user:diego-password"),
					},
				}
			})

			It("performs a handshake with the client using the proxy server config", func() {
				_, err := ssh.Dial("tcp", proxyAddress, clientConfig)
				Expect(err).NotTo(HaveOccurred())

				Expect(proxyAuthenticator.AuthenticateCallCount()).To(Equal(1))

				metadata, password := proxyAuthenticator.AuthenticateArgsForCall(0)
				Expect(metadata.User()).To(Equal("diego:some-instance-guid"))
				Expect(string(password)).To(Equal("diego-user:diego-password"))
			})

			Context("when the handshake fails", func() {
				BeforeEach(func() {
					proxyAuthenticator.AuthenticateReturns(nil, errors.New("go away"))
				})

				JustBeforeEach(func() {
					_, err := ssh.Dial("tcp", proxyAddress, clientConfig)
					Expect(err).To(MatchError(ContainSubstring("ssh: handshake failed: ssh: unable to authenticate")))
				})

				It("does not attempt to authenticate with the daemon", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("when the client handshake is successful", func() {
				var client *ssh.Client

				JustBeforeEach(func() {
					var err error
					client, err = ssh.Dial("tcp", proxyAddress, clientConfig)
					Expect(err).NotTo(HaveOccurred())
				})

				It("handshakes with the target using the provided configuration", func() {
					Eventually(daemonAuthenticator.AuthenticateCallCount).Should(Equal(1))

					metadata, password := daemonAuthenticator.AuthenticateArgsForCall(0)
					Expect(metadata.User()).To(Equal("some-user"))
					Expect(string(password)).To(Equal("fake-some-password"))
				})

				It("emits a successful log message on behalf of the lrp", func() {
					Eventually(fakeLogSender.GetLogs).Should(HaveLen(1))
					logMessage := fakeLogSender.GetLogs()[0]
					Expect(logMessage.AppId).To(Equal("a-guid"))
					Expect(logMessage.SourceType).To(Equal("SSH"))
					Expect(logMessage.SourceInstance).To(Equal("1"))
					Expect(logMessage.Message).To(Equal("a-message"))
				})

				Context("when the target contains a host fingerprint", func() {
					Context("when the fingerprint is an md5 hash", func() {
						BeforeEach(func() {
							targetConfigJson, err := json.Marshal(proxy.TargetConfig{
								Address:         sshdListener.Addr().String(),
								HostFingerprint: helpers.MD5Fingerprint(TestHostKey.PublicKey()),
								User:            "some-user",
								Password:        "fake-some-password",
							})
							Expect(err).NotTo(HaveOccurred())

							permissions := &ssh.Permissions{
								CriticalOptions: map[string]string{
									"proxy-target-config": string(targetConfigJson),
								},
							}
							proxyAuthenticator.AuthenticateReturns(permissions, nil)
						})

						It("handshakes with the target using the provided configuration", func() {
							Eventually(daemonAuthenticator.AuthenticateCallCount).Should(Equal(1))
						})
					})

					Context("when the host fingerprint is a sha1 hash", func() {
						BeforeEach(func() {
							targetConfigJson, err := json.Marshal(proxy.TargetConfig{
								Address:         sshdListener.Addr().String(),
								HostFingerprint: helpers.SHA1Fingerprint(TestHostKey.PublicKey()),
								User:            "some-user",
								Password:        "fake-some-password",
							})
							Expect(err).NotTo(HaveOccurred())

							permissions := &ssh.Permissions{
								CriticalOptions: map[string]string{
									"proxy-target-config": string(targetConfigJson),
								},
							}
							proxyAuthenticator.AuthenticateReturns(permissions, nil)
						})

						It("handshakes with the target using the provided configuration", func() {
							Eventually(daemonAuthenticator.AuthenticateCallCount).Should(Equal(1))
						})
					})

					Context("when the actual host fingerpreint does not match the expected fingerprint", func() {
						BeforeEach(func() {
							targetConfigJson, err := json.Marshal(proxy.TargetConfig{
								Address:         sshdListener.Addr().String(),
								HostFingerprint: "bogus-fingerprint",
								User:            "some-user",
								Password:        "fake-some-password",
							})
							Expect(err).NotTo(HaveOccurred())

							permissions := &ssh.Permissions{
								CriticalOptions: map[string]string{
									"proxy-target-config": string(targetConfigJson),
								},
							}
							proxyAuthenticator.AuthenticateReturns(permissions, nil)
						})

						It("does not attempt authentication with the target", func() {
							Consistently(daemonAuthenticator.AuthenticateCallCount).Should(Equal(0))
						})

						It("closes the connection", func() {
							Eventually(client.Wait).Should(Equal(io.EOF))
						})

						It("logs the failure", func() {
							Eventually(logger).Should(gbytes.Say(`host-key-fingerprint-mismatch`))
						})
					})
				})

				Context("when the target address is unreachable", func() {
					BeforeEach(func() {
						permissions := &ssh.Permissions{
							CriticalOptions: map[string]string{
								"proxy-target-config": `{"address": "0.0.0.0:0"}`,
							},
						}
						proxyAuthenticator.AuthenticateReturns(permissions, nil)
					})

					It("closes the connection", func() {
						Eventually(client.Wait).Should(Equal(io.EOF))
					})

					It("logs the failure", func() {
						Eventually(logger).Should(gbytes.Say(`new-client-conn.dial-failed.*0\.0\.0\.0:0`))
					})
				})

				Context("when the handshake fails", func() {
					BeforeEach(func() {
						daemonAuthenticator.AuthenticateReturns(nil, errors.New("go away"))
					})

					It("closes the connection", func() {
						Eventually(client.Wait).Should(Equal(io.EOF))
					})

					It("logs the failure", func() {
						Eventually(logger).Should(gbytes.Say(`new-client-conn.handshake-failed`))
					})
				})
			})

			Context("when HandleConnection returns", func() {
				var fakeServerConnection *fake_net.FakeConn

				BeforeEach(func() {
					proxySSHConfig.NoClientAuth = true
					daemonSSHConfig.NoClientAuth = true
				})

				JustBeforeEach(func() {
					clientNetConn, serverNetConn := test_helpers.Pipe()

					fakeServerConnection = &fake_net.FakeConn{}
					fakeServerConnection.ReadStub = serverNetConn.Read
					fakeServerConnection.WriteStub = serverNetConn.Write
					fakeServerConnection.CloseStub = serverNetConn.Close

					go sshProxy.HandleConnection(fakeServerConnection)

					clientConn, clientChannels, clientRequests, err := ssh.NewClientConn(clientNetConn, "0.0.0.0", clientConfig)
					Expect(err).NotTo(HaveOccurred())

					client := ssh.NewClient(clientConn, clientChannels, clientRequests)
					client.Close()
				})

				It("ensures the network connection is closed", func() {
					Eventually(fakeServerConnection.CloseCallCount).Should(BeNumerically(">=", 1))
				})
			})
		})

		Context("after both handshakes have been performed", func() {
			var clientConfig *ssh.ClientConfig

			BeforeEach(func() {
				clientConfig = &ssh.ClientConfig{
					User: "diego:some-instance-guid",
					Auth: []ssh.AuthMethod{
						ssh.Password("diego-user:diego-password"),
					},
				}
				daemonSSHConfig.NoClientAuth = true
			})

			Describe("client requests to target", func() {
				var client *ssh.Client

				JustBeforeEach(func() {
					var err error
					client, err = ssh.Dial("tcp", proxyAddress, clientConfig)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					client.Close()
				})

				Context("when the client sends a global request", func() {
					var globalRequestHandler *fake_handlers.FakeGlobalRequestHandler

					BeforeEach(func() {
						globalRequestHandler = &fake_handlers.FakeGlobalRequestHandler{}
						globalRequestHandler.HandleRequestStub = func(logger lager.Logger, request *ssh.Request) {
							request.Reply(true, []byte("response-payload"))
						}
						daemonGlobalRequestHandlers["test-global-request"] = globalRequestHandler
					})

					It("gets forwarded to the daemon and the response comes back", func() {
						accepted, response, err := client.SendRequest("test-global-request", true, []byte("request-payload"))
						Expect(err).NotTo(HaveOccurred())
						Expect(accepted).To(BeTrue())
						Expect(response).To(Equal([]byte("response-payload")))

						Expect(globalRequestHandler.HandleRequestCallCount()).To(Equal(1))

						_, request := globalRequestHandler.HandleRequestArgsForCall(0)
						Expect(request.Type).To(Equal("test-global-request"))
						Expect(request.WantReply).To(BeTrue())
						Expect(request.Payload).To(Equal([]byte("request-payload")))
					})
				})

				Context("when the client requests a new channel", func() {
					var newChannelHandler *fake_handlers.FakeNewChannelHandler

					BeforeEach(func() {
						newChannelHandler = &fake_handlers.FakeNewChannelHandler{}
						newChannelHandler.HandleNewChannelStub = func(logger lager.Logger, newChannel ssh.NewChannel) {
							newChannel.Reject(ssh.Prohibited, "not now")
						}
						daemonNewChannelHandlers["test"] = newChannelHandler
					})

					It("gets forwarded to the daemon", func() {
						_, _, err := client.OpenChannel("test", nil)
						Expect(err).To(Equal(&ssh.OpenChannelError{Reason: ssh.Prohibited, Message: "not now"}))
					})
				})
			})

			Describe("target requests to client", func() {
				var (
					connectionHandler *server_fakes.FakeConnectionHandler

					target        *server.Server
					targetDone    chan struct{}
					listener      net.Listener
					targetAddress string

					clientChannels <-chan ssh.NewChannel
					clientRequests <-chan *ssh.Request
				)

				BeforeEach(func() {
					var err error
					listener, err = net.Listen("tcp", "127.0.0.1:0")
					Expect(err).NotTo(HaveOccurred())
					targetAddress = listener.Addr().String()

					connectionHandler = &server_fakes.FakeConnectionHandler{}
					targetDone = make(chan struct{})
				})

				JustBeforeEach(func() {
					target = server.NewServer(logger.Session("target"), "", connectionHandler)
					target.SetListener(listener)
					go func() {
						target.Serve()
						close(targetDone)
					}()

					clientNetConn, err := net.Dial("tcp", targetAddress)
					_, clientChannels, clientRequests, err = ssh.NewClientConn(clientNetConn, "0.0.0.0", &ssh.ClientConfig{})
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					target.Shutdown()
					Eventually(targetDone).Should(BeClosed())
				})

				Context("when the target sends a global request", func() {
					var handleConnDone chan struct{}

					BeforeEach(func() {
						handleConnDone = make(chan struct{})
						connectionHandler.HandleConnectionStub = func(conn net.Conn) {
							defer GinkgoRecover()
							defer func() {
								handleConnDone <- struct{}{}
							}()

							serverConn, _, _, err := ssh.NewServerConn(conn, daemonSSHConfig)
							Expect(err).NotTo(HaveOccurred())

							accepted, response, err := serverConn.SendRequest("test", true, []byte("test-data"))
							Expect(err).NotTo(HaveOccurred())
							Expect(accepted).To(BeTrue())
							Expect(response).To(Equal([]byte("response-data")))

							serverConn.Close()
						}
					})

					AfterEach(func() {
						close(handleConnDone)
					})

					It("gets forwarded to the client", func() {
						var req *ssh.Request
						Eventually(clientRequests).Should(Receive(&req))

						req.Reply(true, []byte("response-data"))

						Eventually(handleConnDone).Should(Receive())
					})
				})

				Context("when the target requests a new channel", func() {
					var done chan struct{}

					BeforeEach(func() {
						done = make(chan struct{})

						connectionHandler.HandleConnectionStub = func(conn net.Conn) {
							defer GinkgoRecover()

							serverConn, _, _, err := ssh.NewServerConn(conn, daemonSSHConfig)
							Expect(err).NotTo(HaveOccurred())

							channel, requests, err := serverConn.OpenChannel("test-channel", []byte("extra-data"))
							Expect(err).NotTo(HaveOccurred())
							Expect(channel).NotTo(BeNil())
							Expect(requests).NotTo(BeClosed())

							channel.Write([]byte("hello"))

							channelResponse := make([]byte, 7)
							channel.Read(channelResponse)
							Expect(string(channelResponse)).To(Equal("goodbye"))

							channel.Close()
							serverConn.Close()

							close(done)
						}
					})

					AfterEach(func() {
						Eventually(done).Should(BeClosed())
					})

					It("gets forwarded to the client", func() {
						var newChannel ssh.NewChannel
						Eventually(clientChannels).Should(Receive(&newChannel))

						Expect(newChannel.ChannelType()).To(Equal("test-channel"))
						Expect(newChannel.ExtraData()).To(Equal([]byte("extra-data")))

						channel, requests, err := newChannel.Accept()
						Expect(err).NotTo(HaveOccurred())
						Expect(channel).NotTo(BeNil())
						Expect(requests).NotTo(BeClosed())

						channelRequest := make([]byte, 5)
						channel.Read(channelRequest)
						Expect(string(channelRequest)).To(Equal("hello"))

						channel.Write([]byte("goodbye"))
						channel.Close()
					})
				})
			})

			Describe("connection metrics", func() {
				var (
					sender *fake.FakeMetricSender
				)

				BeforeEach(func() {
					sender = fake.NewFakeMetricSender()
					metrics.Initialize(sender, nil)
				})

				Context("when a connection is received", func() {
					It("emit a metric for the total number of connections", func() {
						_, err := ssh.Dial("tcp", proxyAddress, clientConfig)
						Expect(err).NotTo(HaveOccurred())

						Eventually(
							func() float64 {
								return sender.GetValue("ssh-connections").Value
							},
						).Should(Equal(float64(1)))

						_, err = ssh.Dial("tcp", proxyAddress, clientConfig)
						Expect(err).NotTo(HaveOccurred())

						Eventually(
							func() float64 {
								return sender.GetValue("ssh-connections").Value
							},
						).Should(Equal(float64(2)))
					})
				})

				Context("when a connection is closed", func() {
					It("emit a metric for the total number of connections", func() {
						conn, err := ssh.Dial("tcp", proxyAddress, clientConfig)
						Expect(err).NotTo(HaveOccurred())
						Eventually(
							func() float64 {
								return sender.GetValue("ssh-connections").Value
							},
						).Should(Equal(float64(1)))

						conn.Close()
						Eventually(
							func() float64 {
								return sender.GetValue("ssh-connections").Value
							},
						).Should(Equal(float64(0)))
					})
				})
			})

			Describe("app logs", func() {
				Context("when a connection is closed", func() {
					It("logs that the connection has been closed", func() {
						conn, err := ssh.Dial("tcp", proxyAddress, clientConfig)
						Expect(err).NotTo(HaveOccurred())

						conn.Close()

						Eventually(
							func() string {
								lastIdx := len(fakeLogSender.GetLogs()) - 1
								if lastIdx == -1 {
									return ""
								}
								return fakeLogSender.GetLogs()[lastIdx].Message
							},
						).Should(ContainSubstring("Remote access ended for"))
					})
				})
			})
		})
	})

	Describe("ProxyGlobalRequests", func() {
		var (
			sshConn *fake_ssh.FakeConn
			reqChan chan *ssh.Request

			done chan struct{}
		)

		BeforeEach(func() {
			sshConn = &fake_ssh.FakeConn{}
			reqChan = make(chan *ssh.Request, 2)
			done = make(chan struct{}, 1)
		})

		JustBeforeEach(func() {
			go func(done chan<- struct{}) {
				proxy.ProxyGlobalRequests(logger, sshConn, reqChan)
				done <- struct{}{}
			}(done)
		})

		Context("when a request is received", func() {
			BeforeEach(func() {
				request := &ssh.Request{Type: "test", WantReply: false, Payload: []byte("test-data")}
				reqChan <- request
				reqChan <- request
			})

			AfterEach(func() {
				close(reqChan)
			})

			It("forwards requests from the channel to the connection", func() {
				Eventually(sshConn.SendRequestCallCount).Should(Equal(2))
				Consistently(sshConn.SendRequestCallCount).Should(Equal(2))

				reqType, wantReply, payload := sshConn.SendRequestArgsForCall(0)
				Expect(reqType).To(Equal("test"))
				Expect(wantReply).To(BeFalse())
				Expect(payload).To(Equal([]byte("test-data")))

				reqType, wantReply, payload = sshConn.SendRequestArgsForCall(1)
				Expect(reqType).To(Equal("test"))
				Expect(wantReply).To(BeFalse())
				Expect(payload).To(Equal([]byte("test-data")))
			})
		})

		Context("when SendRequest fails", func() {
			BeforeEach(func() {
				callCount := 0
				sshConn.SendRequestStub = func(rt string, wr bool, p []byte) (bool, []byte, error) {
					callCount++
					if callCount == 1 {
						return false, nil, errors.New("woops")
					}
					return true, nil, nil
				}

				reqChan <- &ssh.Request{}
				reqChan <- &ssh.Request{}
			})

			AfterEach(func() {
				close(reqChan)
			})

			It("continues processing requests", func() {
				Eventually(sshConn.SendRequestCallCount).Should(Equal(2))
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say(`send-request-failed.*woops`))
			})
		})

		Context("when the request channel closes", func() {
			JustBeforeEach(func() {
				Consistently(reqChan).ShouldNot(BeClosed())
				close(reqChan)
			})

			It("returns gracefully", func() {
				Eventually(done).Should(Receive())
			})
		})
	})

	Describe("ProxyChannels", func() {
		var (
			targetConn  *fake_ssh.FakeConn
			newChanChan chan ssh.NewChannel

			newChan       *fake_ssh.FakeNewChannel
			sourceChannel *fake_ssh.FakeChannel
			sourceReqChan chan *ssh.Request
			sourceStderr  *fake_ssh.FakeChannel

			targetChannel *fake_ssh.FakeChannel
			targetReqChan chan *ssh.Request
			targetStderr  *fake_ssh.FakeChannel

			done chan struct{}
		)

		BeforeEach(func() {
			targetConn = &fake_ssh.FakeConn{}
			newChanChan = make(chan ssh.NewChannel, 1)

			newChan = &fake_ssh.FakeNewChannel{}
			sourceChannel = &fake_ssh.FakeChannel{}
			sourceReqChan = make(chan *ssh.Request, 2)
			sourceStderr = &fake_ssh.FakeChannel{}
			sourceStderr.ReadReturns(0, io.EOF)
			sourceChannel.StderrReturns(sourceStderr)

			targetChannel = &fake_ssh.FakeChannel{}
			targetReqChan = make(chan *ssh.Request, 2)
			targetStderr = &fake_ssh.FakeChannel{}
			targetStderr.ReadReturns(0, io.EOF)
			targetChannel.StderrReturns(targetStderr)

			done = make(chan struct{}, 1)
		})

		JustBeforeEach(func() {
			go func(done chan<- struct{}) {
				proxy.ProxyChannels(logger, targetConn, newChanChan)
				done <- struct{}{}
			}(done)
		})

		Context("when a new channel is opened by the client", func() {
			BeforeEach(func() {
				sourceChannel.ReadReturns(0, io.EOF)
				targetChannel.ReadReturns(0, io.EOF)

				newChan.ChannelTypeReturns("test")
				newChan.ExtraDataReturns([]byte("extra-data"))
				newChan.AcceptReturns(sourceChannel, sourceReqChan, nil)

				targetConn.OpenChannelReturns(targetChannel, targetReqChan, nil)

				newChanChan <- newChan
			})

			AfterEach(func() {
				close(newChanChan)
			})

			It("forwards the NewChannel request to the target", func() {
				Eventually(targetConn.OpenChannelCallCount).Should(Equal(1))
				Consistently(targetConn.OpenChannelCallCount).Should(Equal(1))

				channelType, extraData := targetConn.OpenChannelArgsForCall(0)
				Expect(channelType).To(Equal("test"))
				Expect(extraData).To(Equal([]byte("extra-data")))
			})

			Context("when the target accepts the connection", func() {
				It("accepts the source request", func() {
					Eventually(newChan.AcceptCallCount).Should(Equal(1))
				})

				Context("when the source channel has data available", func() {
					BeforeEach(func() {
						sourceChannel.ReadStub = func(dest []byte) (int, error) {
							if cap(dest) >= 3 {
								copy(dest, []byte("abc"))
								return 3, io.EOF
							}
							return 0, io.EOF
						}
						sourceStderr.ReadStub = func(dest []byte) (int, error) {
							if cap(dest) >= 3 {
								copy(dest, []byte("xyz"))
								return 3, io.EOF
							}
							return 0, io.EOF
						}
					})

					It("copies the source channel to the target channel", func() {
						Eventually(targetChannel.WriteCallCount).ShouldNot(Equal(0))

						data := targetChannel.WriteArgsForCall(0)
						Expect(data).To(Equal([]byte("abc")))

					})

					It("copies the source stderr to the target stderr", func() {
						Eventually(targetStderr.WriteCallCount).ShouldNot(Equal(0))

						data := targetStderr.WriteArgsForCall(0)
						Expect(data).To(Equal([]byte("xyz")))
					})
				})

				Context("when the target channel has data available", func() {
					BeforeEach(func() {
						targetChannel.ReadStub = func(dest []byte) (int, error) {
							if cap(dest) >= 3 {
								copy(dest, []byte("xyz"))
								return 3, io.EOF
							}
							return 0, io.EOF
						}
						targetStderr.ReadStub = func(dest []byte) (int, error) {
							if cap(dest) >= 3 {
								copy(dest, []byte("abc"))
								return 3, io.EOF
							}
							return 0, io.EOF
						}
					})

					It("copies the target channel to the source channel", func() {
						Eventually(sourceChannel.WriteCallCount).ShouldNot(Equal(0))

						data := sourceChannel.WriteArgsForCall(0)
						Expect(data).To(Equal([]byte("xyz")))

					})

					It("copies the target stderr to the source stderr", func() {
						Eventually(sourceStderr.WriteCallCount).ShouldNot(Equal(0))

						data := sourceStderr.WriteArgsForCall(0)
						Expect(data).To(Equal([]byte("abc")))
					})
				})

				Context("when the source channel closes", func() {
					BeforeEach(func() {
						sourceChannel.ReadReturns(0, io.EOF)
					})

					It("closes the target channel", func() {
						Eventually(sourceChannel.ReadCallCount).Should(Equal(1))
						Eventually(targetChannel.CloseWriteCallCount).Should(Equal(1))
					})
				})

				Context("when the target channel closes", func() {
					BeforeEach(func() {
						targetChannel.ReadReturns(0, io.EOF)
					})

					It("closes the source channel", func() {
						Eventually(sourceChannel.ReadCallCount).Should(Equal(1))
						Eventually(targetChannel.CloseWriteCallCount).Should(Equal(1))
					})
				})

				Context("when out of band requests are received on the source channel", func() {
					BeforeEach(func() {
						request := &ssh.Request{Type: "test", WantReply: false, Payload: []byte("test-data")}
						sourceReqChan <- request
					})

					It("forwards the request to the target channel", func() {
						Eventually(targetChannel.SendRequestCallCount).Should(Equal(1))

						reqType, wantReply, payload := targetChannel.SendRequestArgsForCall(0)
						Expect(reqType).To(Equal("test"))
						Expect(wantReply).To(BeFalse())
						Expect(payload).To(Equal([]byte("test-data")))
					})
				})

				Context("when out of band requests are received from the target channel", func() {
					BeforeEach(func() {
						request := &ssh.Request{Type: "test", WantReply: false, Payload: []byte("test-data")}
						targetReqChan <- request
					})

					It("forwards the request to the target channel", func() {
						Eventually(sourceChannel.SendRequestCallCount).Should(Equal(1))

						reqType, wantReply, payload := sourceChannel.SendRequestArgsForCall(0)
						Expect(reqType).To(Equal("test"))
						Expect(wantReply).To(BeFalse())
						Expect(payload).To(Equal([]byte("test-data")))
					})
				})
			})

			Context("when the target rejects the connection", func() {
				BeforeEach(func() {
					openError := &ssh.OpenChannelError{
						Reason:  ssh.Prohibited,
						Message: "go away",
					}
					targetConn.OpenChannelReturns(nil, nil, openError)
				})

				It("rejects the source request with the upstream error", func() {
					Eventually(newChan.RejectCallCount).Should(Equal(1))

					reason, message := newChan.RejectArgsForCall(0)
					Expect(reason).To(Equal(ssh.Prohibited))
					Expect(message).To(Equal("go away"))
				})

				It("continues processing new channel requests", func() {
					newChanChan <- newChan
					Eventually(newChan.RejectCallCount).Should(Equal(2))
				})
			})

			Context("when openning a channel failsfails", func() {
				BeforeEach(func() {
					targetConn.OpenChannelReturns(nil, nil, errors.New("woops"))
				})

				It("rejects the source request with a connection failed reason", func() {
					Eventually(newChan.RejectCallCount).Should(Equal(1))

					reason, message := newChan.RejectArgsForCall(0)
					Expect(reason).To(Equal(ssh.ConnectionFailed))
					Expect(message).To(Equal("woops"))
				})

				It("continues processing new channel requests", func() {
					newChanChan <- newChan
					Eventually(newChan.RejectCallCount).Should(Equal(2))
				})
			})
		})

		Context("when the new channel channel closes", func() {
			JustBeforeEach(func() {
				Consistently(newChanChan).ShouldNot(BeClosed())
				close(newChanChan)
			})

			It("returns gracefully", func() {
				Eventually(done).Should(Receive())
			})
		})
	})

	Describe("ProxyRequests", func() {
		var (
			channel *fake_ssh.FakeChannel
			reqChan chan *ssh.Request

			wg   *sync.WaitGroup
			done chan struct{}
		)

		BeforeEach(func() {
			wg = &sync.WaitGroup{}
			channel = &fake_ssh.FakeChannel{}
			reqChan = make(chan *ssh.Request, 2)
			done = make(chan struct{}, 1)
		})

		JustBeforeEach(func() {
			go func(done chan<- struct{}) {
				proxy.ProxyRequests(logger, "test", reqChan, channel, wg)
				done <- struct{}{}
			}(done)
		})

		Context("when a request is received", func() {
			BeforeEach(func() {
				request := &ssh.Request{Type: "test", WantReply: false, Payload: []byte("test-data")}
				reqChan <- request
				reqChan <- request
			})

			AfterEach(func() {
				close(reqChan)
			})

			It("forwards requests from the channel to the connection", func() {
				Eventually(channel.SendRequestCallCount).Should(Equal(2))
				Consistently(channel.SendRequestCallCount).Should(Equal(2))

				reqType, wantReply, payload := channel.SendRequestArgsForCall(0)
				Expect(reqType).To(Equal("test"))
				Expect(wantReply).To(BeFalse())
				Expect(payload).To(Equal([]byte("test-data")))

				reqType, wantReply, payload = channel.SendRequestArgsForCall(1)
				Expect(reqType).To(Equal("test"))
				Expect(wantReply).To(BeFalse())
				Expect(payload).To(Equal([]byte("test-data")))
			})
		})

		Context("when SendRequest fails", func() {
			BeforeEach(func() {
				callCount := 0
				channel.SendRequestStub = func(rt string, wr bool, p []byte) (bool, error) {
					callCount++
					if callCount == 1 {
						return false, errors.New("woops")
					}
					return true, nil
				}

				reqChan <- &ssh.Request{}
				reqChan <- &ssh.Request{}
			})

			AfterEach(func() {
				close(reqChan)
			})

			It("continues processing requests", func() {
				Eventually(channel.SendRequestCallCount).Should(Equal(2))
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say(`send-request-failed.*woops`))
			})
		})

		Context("when the request channel closes", func() {
			JustBeforeEach(func() {
				Consistently(reqChan).ShouldNot(BeClosed())
				close(reqChan)
			})

			It("returns gracefully", func() {
				Eventually(done).Should(Receive())
			})
		})

		Context("when an exit-status request is received", func() {
			BeforeEach(func() {
				request := &ssh.Request{Type: "exit-status", WantReply: false, Payload: []byte("test-data")}
				reqChan <- request
				reqChan <- request
			})

			AfterEach(func() {
				close(reqChan)
			})

			It("does not handle extra requests", func() {
				Eventually(channel.SendRequestCallCount).Should(Equal(1))
				Consistently(channel.SendRequestCallCount).Should(Equal(1))

				Eventually(channel.CloseCallCount).Should(Equal(1))
				reqType, wantReply, payload := channel.SendRequestArgsForCall(0)
				Expect(reqType).To(Equal("exit-status"))
				Expect(wantReply).To(BeFalse())
				Expect(payload).To(Equal([]byte("test-data")))
			})

			Context("when there is a wait group", func() {
				BeforeEach(func() {
					wg.Add(1)
				})

				It("exits when the waitgroup is done", func() {
					Eventually(channel.SendRequestCallCount).Should(Equal(1))
					Consistently(channel.SendRequestCallCount).Should(Equal(1))

					Consistently(channel.CloseCallCount).Should(Equal(0))
					wg.Done()
					Eventually(channel.CloseCallCount).Should(Equal(1))

					reqType, wantReply, payload := channel.SendRequestArgsForCall(0)
					Expect(reqType).To(Equal("exit-status"))
					Expect(wantReply).To(BeFalse())
					Expect(payload).To(Equal([]byte("test-data")))
				})
			})
		})
	})

	Describe("NewClientConn", func() {
		var (
			permissions *ssh.Permissions

			daemonSSHConfig *ssh.ServerConfig
			sshDaemon       *daemon.Daemon
			sshdListener    net.Listener
			sshdServer      *server.Server

			newClientConnErr error
		)

		BeforeEach(func() {
			permissions = &ssh.Permissions{
				CriticalOptions: map[string]string{},
			}

			daemonSSHConfig = &ssh.ServerConfig{}
			daemonSSHConfig.AddHostKey(TestHostKey)

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).NotTo(HaveOccurred())

			sshdListener = listener
		})

		JustBeforeEach(func() {
			sshDaemon = daemon.New(logger.Session("sshd"), daemonSSHConfig, nil, nil)
			sshdServer = server.NewServer(logger, "127.0.0.1:0", sshDaemon)
			sshdServer.SetListener(sshdListener)
			go sshdServer.Serve()

			_, _, _, newClientConnErr = proxy.NewClientConn(logger, permissions)
		})

		AfterEach(func() {
			sshdServer.Shutdown()
		})

		Context("when permissions is nil", func() {
			BeforeEach(func() {
				permissions = nil
			})

			It("returns an error", func() {
				Expect(newClientConnErr).To(HaveOccurred())
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say("permissions-and-critical-options-required"))
			})
		})

		Context("when permissions.CriticalOptions is nil", func() {
			BeforeEach(func() {
				permissions.CriticalOptions = nil
			})

			It("returns an error", func() {
				Expect(newClientConnErr).To(HaveOccurred())
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say("permissions-and-critical-options-required"))
			})
		})

		Context("when the config is missing", func() {
			BeforeEach(func() {
				delete(permissions.CriticalOptions, "proxy-target-config")
			})

			It("returns an error", func() {
				Expect(newClientConnErr).To(HaveOccurred())
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say("unmarshal-failed"))
			})
		})

		Context("when the config fails to unmarshal", func() {
			BeforeEach(func() {
				permissions.CriticalOptions["proxy-target-config"] = "{ this_is: invalid json"
			})

			It("returns an error", func() {
				Expect(newClientConnErr).To(HaveOccurred())
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say("unmarshal-failed"))
			})
		})

		Context("when the address in the config is bad", func() {
			BeforeEach(func() {
				permissions.CriticalOptions["proxy-target-config"] = `{ "address": "0.0.0.0:0" }`
			})

			It("returns an error", func() {
				Expect(newClientConnErr).To(HaveOccurred())
			})

			It("logs the failure", func() {
				Eventually(logger).Should(gbytes.Say("dial-failed"))
			})
		})

		Context("when the config contains a user and password", func() {
			var passwordAuthenticator *fake_authenticators.FakePasswordAuthenticator

			BeforeEach(func() {
				targetConfigJson, err := json.Marshal(proxy.TargetConfig{
					Address:  sshdListener.Addr().String(),
					User:     "my-user",
					Password: "my-password",
				})
				Expect(err).NotTo(HaveOccurred())

				permissions = &ssh.Permissions{
					CriticalOptions: map[string]string{
						"proxy-target-config": string(targetConfigJson),
					},
				}

				passwordAuthenticator = &fake_authenticators.FakePasswordAuthenticator{}
				daemonSSHConfig.PasswordCallback = passwordAuthenticator.Authenticate
			})

			It("uses the user and password for authentication", func() {
				Expect(passwordAuthenticator.AuthenticateCallCount()).To(Equal(1))

				metadata, password := passwordAuthenticator.AuthenticateArgsForCall(0)
				Expect(metadata.User()).To(Equal("my-user"))
				Expect(string(password)).To(Equal("my-password"))
			})
		})

		Context("when the config contains a public key", func() {
			var publicKeyAuthenticator *fake_authenticators.FakePublicKeyAuthenticator

			BeforeEach(func() {
				targetConfigJson, err := json.Marshal(proxy.TargetConfig{
					Address:    sshdListener.Addr().String(),
					PrivateKey: TestPrivatePem,
				})
				Expect(err).NotTo(HaveOccurred())

				permissions = &ssh.Permissions{
					CriticalOptions: map[string]string{
						"proxy-target-config": string(targetConfigJson),
					},
				}

				publicKeyAuthenticator = &fake_authenticators.FakePublicKeyAuthenticator{}
				publicKeyAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)
				daemonSSHConfig.PublicKeyCallback = publicKeyAuthenticator.Authenticate
			})

			It("will use the public key for authentication", func() {
				expectedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(TestPublicAuthorizedKey))
				Expect(err).NotTo(HaveOccurred())

				Expect(publicKeyAuthenticator.AuthenticateCallCount()).To(Equal(1))

				_, actualKey := publicKeyAuthenticator.AuthenticateArgsForCall(0)
				Expect(actualKey.Marshal()).To(Equal(expectedKey.Marshal()))
			})
		})

		Context("when the config contains a user and a public key", func() {
			var publicKeyAuthenticator *fake_authenticators.FakePublicKeyAuthenticator

			BeforeEach(func() {
				targetConfigJson, err := json.Marshal(proxy.TargetConfig{
					Address:    sshdListener.Addr().String(),
					User:       "my-user",
					PrivateKey: TestPrivatePem,
				})
				Expect(err).NotTo(HaveOccurred())

				permissions = &ssh.Permissions{
					CriticalOptions: map[string]string{
						"proxy-target-config": string(targetConfigJson),
					},
				}

				publicKeyAuthenticator = &fake_authenticators.FakePublicKeyAuthenticator{}
				publicKeyAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)
				daemonSSHConfig.PublicKeyCallback = publicKeyAuthenticator.Authenticate
			})

			It("will use the user and public key for authentication", func() {
				expectedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(TestPublicAuthorizedKey))
				Expect(err).NotTo(HaveOccurred())

				Expect(publicKeyAuthenticator.AuthenticateCallCount()).To(Equal(1))

				metadata, actualKey := publicKeyAuthenticator.AuthenticateArgsForCall(0)
				Expect(metadata.User()).To(Equal("my-user"))
				Expect(actualKey.Marshal()).To(Equal(expectedKey.Marshal()))
			})
		})

		Context("when the config contains a user, password, a public key", func() {
			var publicKeyAuthenticator *fake_authenticators.FakePublicKeyAuthenticator
			var passwordAuthenticator *fake_authenticators.FakePasswordAuthenticator

			BeforeEach(func() {
				targetConfigJson, err := json.Marshal(proxy.TargetConfig{
					Address:    sshdListener.Addr().String(),
					User:       "my-user",
					Password:   "my-password",
					PrivateKey: TestPrivatePem,
				})
				Expect(err).NotTo(HaveOccurred())

				permissions = &ssh.Permissions{
					CriticalOptions: map[string]string{
						"proxy-target-config": string(targetConfigJson),
					},
				}

				passwordAuthenticator = &fake_authenticators.FakePasswordAuthenticator{}
				daemonSSHConfig.PasswordCallback = passwordAuthenticator.Authenticate

				publicKeyAuthenticator = &fake_authenticators.FakePublicKeyAuthenticator{}
				publicKeyAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)
				daemonSSHConfig.PublicKeyCallback = publicKeyAuthenticator.Authenticate
			})

			It("will attempt to use the public key for authentication before the password", func() {
				Expect(publicKeyAuthenticator.AuthenticateCallCount()).To(Equal(1))
				Expect(passwordAuthenticator.AuthenticateCallCount()).To(Equal(0))
			})

			Context("when public key authentication fails", func() {
				BeforeEach(func() {
					passwordAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)
					publicKeyAuthenticator.AuthenticateReturns(nil, errors.New("go away"))
				})

				It("will fall back to password authentication", func() {
					Expect(publicKeyAuthenticator.AuthenticateCallCount()).To(Equal(1))
					Expect(passwordAuthenticator.AuthenticateCallCount()).To(Equal(1))
				})
			})
		})
	})

	Describe("Wait", func() {
		var (
			waitChans []chan struct{}
			waiters   []proxy.Waiter

			done chan struct{}
		)

		BeforeEach(func() {
			for i := 0; i < 3; i++ {
				idx := i
				waitChans = append(waitChans, make(chan struct{}))

				conn := &fake_ssh.FakeConn{}
				conn.WaitStub = func() error {
					<-waitChans[idx]
					return nil
				}
				waiters = append(waiters, conn)
			}

			done = make(chan struct{}, 1)
		})

		JustBeforeEach(func() {
			go func(done chan<- struct{}) {
				proxy.Wait(logger, waiters...)
				done <- struct{}{}
			}(done)
		})

		It("waits for all Waiters to finish", func() {
			Consistently(done).ShouldNot(Receive())
			close(waitChans[0])

			Consistently(done).ShouldNot(Receive())
			close(waitChans[1])

			Consistently(done).ShouldNot(Receive())
			close(waitChans[2])

			Eventually(done).Should(Receive())
		})
	})
})
