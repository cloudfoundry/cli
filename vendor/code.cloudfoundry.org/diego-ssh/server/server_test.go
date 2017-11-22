package server_test

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"code.cloudfoundry.org/diego-ssh/server"
	"code.cloudfoundry.org/diego-ssh/server/fakes"
	"code.cloudfoundry.org/diego-ssh/test_helpers"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_net"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Server", func() {
	var (
		logger lager.Logger
		srv    *server.Server

		handler *fakes.FakeConnectionHandler

		address string
	)

	BeforeEach(func() {
		handler = &fakes.FakeConnectionHandler{}
		address = fmt.Sprintf("127.0.0.1:%d", 7001+GinkgoParallelNode())
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Run", func() {
		var process ifrit.Process

		BeforeEach(func() {
			srv = server.NewServer(logger, address, handler)
			process = ifrit.Invoke(srv)
		})

		AfterEach(func() {
			process.Signal(os.Interrupt)
			Eventually(process.Wait()).Should(Receive())
		})

		It("accepts connections on the specified address", func() {
			_, err := net.Dial("tcp", address)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a second client connects", func() {
			JustBeforeEach(func() {
				_, err := net.Dial("tcp", address)
				Expect(err).NotTo(HaveOccurred())
			})

			It("accepts the new connection", func() {
				_, err := net.Dial("tcp", address)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("SetListener", func() {
		var fakeListener *fake_net.FakeListener

		BeforeEach(func() {
			fakeListener = &fake_net.FakeListener{}

			srv = server.NewServer(logger, address, handler)
			srv.SetListener(fakeListener)
		})

		Context("when a listener has already been set", func() {
			It("returns an error", func() {
				listener := &fake_net.FakeListener{}
				err := srv.SetListener(listener)
				Expect(err).To(MatchError("Listener has already been set"))
			})
		})
	})

	Describe("Serve", func() {
		var fakeListener *fake_net.FakeListener
		var fakeConn *fake_net.FakeConn

		BeforeEach(func() {
			fakeListener = &fake_net.FakeListener{}
			fakeConn = &fake_net.FakeConn{}

			connectionCh := make(chan net.Conn, 1)
			connectionCh <- fakeConn

			fakeListener.AcceptStub = func() (net.Conn, error) {
				cx := connectionCh
				select {
				case conn := <-cx:
					return conn, nil
				default:
					return nil, errors.New("fail")
				}
			}
		})

		JustBeforeEach(func() {
			srv = server.NewServer(logger, address, handler)
			srv.SetListener(fakeListener)
			srv.Serve()
		})

		It("accepts inbound connections", func() {
			Expect(fakeListener.AcceptCallCount()).To(Equal(2))
		})

		It("passes the connection to the connection handler", func() {
			Eventually(handler.HandleConnectionCallCount).Should(Equal(1))
			Expect(handler.HandleConnectionArgsForCall(0)).To(Equal(fakeConn))
		})

		Context("when accept returns a permanent error", func() {
			BeforeEach(func() {
				fakeListener.AcceptReturns(nil, errors.New("oops"))
			})

			It("closes the listener", func() {
				Expect(fakeListener.CloseCallCount()).To(Equal(1))
			})
		})

		Context("when accept returns a temporary error", func() {
			var timeCh chan time.Time

			BeforeEach(func() {
				timeCh = make(chan time.Time, 3)

				fakeListener.AcceptStub = func() (net.Conn, error) {
					timeCh := timeCh
					select {
					case timeCh <- time.Now():
						return nil, test_helpers.NewTestNetError(false, true)
					default:
						close(timeCh)
						return nil, test_helpers.NewTestNetError(false, false)
					}
				}
			})

			It("retries the accept after a short delay", func() {
				Expect(timeCh).To(HaveLen(3))

				times := make([]time.Time, 0)
				for t := range timeCh {
					times = append(times, t)
				}

				Expect(times[1]).To(BeTemporally("~", times[0].Add(100*time.Millisecond), 20*time.Millisecond))
				Expect(times[2]).To(BeTemporally("~", times[1].Add(100*time.Millisecond), 20*time.Millisecond))
			})
		})
	})

	Describe("Shutdown", func() {
		var fakeListener *fake_net.FakeListener

		BeforeEach(func() {
			fakeListener = &fake_net.FakeListener{}

			srv = server.NewServer(logger, address, handler)
			srv.SetListener(fakeListener)
		})

		Context("when the server is shutdown", func() {
			BeforeEach(func() {
				srv.Shutdown()
			})

			It("closes the listener", func() {
				Expect(fakeListener.CloseCallCount()).To(Equal(1))
			})

			It("marks the server as stopping", func() {
				Expect(srv.IsStopping()).To(BeTrue())
			})

			It("does not log an accept failure", func() {
				Eventually(func() error {
					_, err := net.Dial("tcp", address)
					return err
				}).Should(HaveOccurred())
				Consistently(logger).ShouldNot(gbytes.Say("test.serve.accept-failed"))
			})
		})
	})

	Describe("ListenAddr", func() {
		var listener net.Listener
		BeforeEach(func() {
			srv = server.NewServer(logger, address, handler)
		})

		Context("when the server has a listener", func() {
			BeforeEach(func() {
				var err error
				listener, err = net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				srv = server.NewServer(logger, address, handler)
				err = srv.SetListener(listener)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the address reported by the listener", func() {
				Expect(srv.ListenAddr()).To(Equal(listener.Addr()))
			})
		})

		Context("when the server does not have a listener", func() {
			It("returns an error", func() {
				_, err := srv.ListenAddr()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
