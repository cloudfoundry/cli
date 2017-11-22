package consumer_test

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consumer connecting through a Proxy", func() {
	var (
		connection *consumer.Consumer

		messagesToSend chan []byte
		testServer     *httptest.Server
		endpoint       string
		proxy          func(*http.Request) (*url.URL, error)

		testProxyServer *httptest.Server
		goProxyHandler  *goproxy.ProxyHttpServer
	)

	BeforeEach(func() {
		messagesToSend = make(chan []byte, 256)
		testServer = httptest.NewServer(NewWebsocketHandler(messagesToSend, 100*time.Millisecond))
		endpoint = "ws://" + testServer.Listener.Addr().String()

		goProxyHandler = goproxy.NewProxyHttpServer()
		goProxyHandler.Logger = log.New(bytes.NewBufferString(""), "", 0)
		testProxyServer = httptest.NewServer(goProxyHandler)
		u, err := url.Parse(testProxyServer.URL)
		proxy = func(*http.Request) (*url.URL, error) {
			return u, err
		}
	})

	JustBeforeEach(func() {
		connection = consumer.New(endpoint, nil, proxy)
	})

	AfterEach(func() {
		testProxyServer.Close()
		testServer.Close()
	})

	Describe("StreamWithoutReconnect", func() {
		var (
			incoming <-chan *events.Envelope
			errs     <-chan error
		)

		JustBeforeEach(func() {
			incoming, errs = connection.StreamWithoutReconnect("fakeAppGuid", "authToken")
		})

		AfterEach(func() {
			close(messagesToSend)
		})

		Context("with a message in the trafficcontroller", func() {
			BeforeEach(func() {
				messagesToSend <- marshalMessage(createMessage("hello", 0))
			})

			It("connects using valid URL to running proxy server", func() {
				message := <-incoming
				Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
			})
		})

		Context("with an auth proxy server", func() {
			BeforeEach(func() {
				goProxyHandler.OnRequest().HandleConnect(auth.BasicConnect("my_realm", func(user, passwd string) bool {
					return user == "user" && passwd == "password"
				}))
				proxyURL, err := url.Parse(testProxyServer.URL)
				proxy = func(*http.Request) (*url.URL, error) {
					proxyURL.User = url.UserPassword("user", "password")
					if err != nil {
						return nil, err
					}

					return proxyURL, nil
				}
				messagesToSend <- marshalMessage(createMessage("hello", 0))
			})

			It("connects successfully", func() {
				message := <-incoming
				Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
			})
		})

		Context("with an auth proxy server with bad credential", func() {
			BeforeEach(func() {
				goProxyHandler.OnRequest().HandleConnect(auth.BasicConnect("my_realm", func(user, passwd string) bool {
					return user == "user" && passwd == "password"
				}))
				proxyURL, err := url.Parse(testProxyServer.URL)
				proxy = func(*http.Request) (*url.URL, error) {
					proxyURL.User = url.UserPassword("user", "passwrd")
					if err != nil {
						return nil, err
					}

					return proxyURL, nil
				}
			})

			It("connects successfully", func() {
				var err error
				Eventually(errs).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Proxy Authentication Required"))
			})
		})

		Context("with a closed proxy server", func() {
			BeforeEach(func() {
				testProxyServer.Close()
			})

			It("sends a connection refused error", func() {
				Eventually(func() string {
					err := <-errs
					return err.Error()
				}).Should(ContainSubstring("connection refused"))
			})
		})

		Context("with a proxy that returns errors", func() {
			const errMsg = "Invalid proxy URL"

			BeforeEach(func() {
				proxy = func(*http.Request) (*url.URL, error) {
					return nil, errors.New(errMsg)
				}
			})

			It("sends the errors", func() {
				var err error
				Eventually(errs).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
			})
		})

		Context("with a proxy that rejects connections", func() {
			BeforeEach(func() {
				goProxyHandler.OnRequest().HandleConnect(goproxy.AlwaysReject)
			})

			It("sends a dialing error", func() {
				var err error
				Eventually(errs).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error dialing trafficcontroller server"))
			})
		})

		Context("with a non-proxy server", func() {
			BeforeEach(func() {
				nonProxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Go away, I am not a proxy!", http.StatusBadRequest)
				}))
				proxy = func(*http.Request) (*url.URL, error) {
					return url.Parse(nonProxyServer.URL)
				}
			})

			It("sends a bad request error", func() {
				var err error
				Eventually(errs).Should(Receive(&err))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(http.StatusText(http.StatusBadRequest)))
			})
		})
	})

	Describe("RecentLogs", func() {
		var (
			httpTestServer   *httptest.Server
			incomingMessages []*events.LogMessage
			err              error
		)

		BeforeEach(func() {
			httpTestServer = httptest.NewServer(NewHttpHandler(messagesToSend))
			endpoint = "ws://" + httpTestServer.Listener.Addr().String()
		})

		JustBeforeEach(func() {
			close(messagesToSend)
			incomingMessages, err = connection.RecentLogs("fakeAppGuid", "authToken")
		})

		AfterEach(func() {
			httpTestServer.Close()
		})

		Context("with recent logs", func() {
			BeforeEach(func() {
				messagesToSend <- marshalMessage(createMessage("test-message-0", 0))
				messagesToSend <- marshalMessage(createMessage("test-message-1", 0))
			})

			It("returns those logs from the server", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(incomingMessages).To(HaveLen(2))
				Expect(incomingMessages[0].GetMessage()).To(Equal([]byte("test-message-0")))
				Expect(incomingMessages[1].GetMessage()).To(Equal([]byte("test-message-1")))
			})
		})

		Context("with a proxy that returns errors", func() {
			const errMsg = "Invalid proxy URL"

			BeforeEach(func() {
				proxy = func(*http.Request) (*url.URL, error) {
					return nil, errors.New(errMsg)
				}
			})

			It("connects using failing proxyFunc", func() {
				Expect(err).To(HaveOccurred(), "THIS WILL FAIL ON GOLANG 1.3 - 1.3.3 DUE TO BUG IN STANDARD LIBRARY (see https://code.google.com/p/go/issues/detail?id=8755)")
				Expect(err.Error()).To(ContainSubstring(errMsg))
			})
		})
	})
})
