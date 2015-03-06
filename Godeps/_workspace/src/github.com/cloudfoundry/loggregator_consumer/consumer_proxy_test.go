package loggregator_consumer_test

import (
	"bytes"
	"crypto/tls"
	"errors"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/loggregatorlib/server/handlers"
	"github.com/elazarl/goproxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

var _ = Describe("Loggregator Consumer behind a Proxy", func() {
	var (
		connection        consumer.LoggregatorConsumer
		endpoint          string
		testServer        *httptest.Server
		tlsSettings       *tls.Config
		consumerProxyFunc func(*http.Request) (*url.URL, error)

		appGuid         string
		authToken       string
		incomingChan    <-chan *logmessage.LogMessage
		messagesToSend  chan []byte
		testProxyServer *httptest.Server
		goProxyHandler  *goproxy.ProxyHttpServer

		err error
	)

	BeforeEach(func() {
		messagesToSend = make(chan []byte, 256)

		testServer = httptest.NewServer(handlers.NewWebsocketHandler(messagesToSend, 100*time.Millisecond, loggertesthelper.Logger()))
		endpoint = "ws://" + testServer.Listener.Addr().String()
		goProxyHandler = goproxy.NewProxyHttpServer()
		goProxyHandler.Logger = log.New(bytes.NewBufferString(""), "", 0)
		testProxyServer = httptest.NewServer(goProxyHandler)
		consumerProxyFunc = func(*http.Request) (*url.URL, error) {
			return url.Parse(testProxyServer.URL)
		}
	})

	AfterEach(func() {
		consumerProxyFunc = nil
		if testProxyServer != nil {
			testProxyServer.Close()
		}
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("Tail", func() {

		AfterEach(func() {
			close(messagesToSend)
		})

		perform := func() {
			connection = consumer.New(endpoint, tlsSettings, consumerProxyFunc)
			incomingChan, err = connection.Tail(appGuid, authToken)
		}

		It("connects using valid URL to running consumerProxyFunc server", func() {
			messagesToSend <- marshalMessage(createMessage("hello", 0))
			perform()

			message := <-incomingChan

			Expect(message.Message).To(Equal([]byte("hello")))
		})

		It("connects using valid URL to a stopped consumerProxyFunc server", func() {
			testProxyServer.Close()

			perform()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})

		It("connects using invalid URL", func() {
			errMsg := "Invalid consumerProxyFunc URL"
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return nil, errors.New(errMsg)
			}

			perform()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})

		It("connects to a consumerProxyFunc server rejecting CONNECT requests", func() {
			goProxyHandler.OnRequest().HandleConnect(goproxy.AlwaysReject)

			perform()

			Expect(err).To(HaveOccurred())
		})

		It("connects to a non-consumerProxyFunc server", func() {
			nonProxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Go away, I am not a consumerProxyFunc!", http.StatusBadRequest)
			}))
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return url.Parse(nonProxyServer.URL)
			}

			perform()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(http.StatusText(http.StatusBadRequest)))
		})
	})

	Describe("Recent", func() {
		var httpTestServer *httptest.Server
		var incomingMessages []*logmessage.LogMessage

		perform := func() {
			close(messagesToSend)
			connection = consumer.New(endpoint, tlsSettings, consumerProxyFunc)
			incomingMessages, err = connection.Recent(appGuid, authToken)
		}

		BeforeEach(func() {
			httpTestServer = httptest.NewServer(handlers.NewHttpHandler(messagesToSend, loggertesthelper.Logger()))
			endpoint = "ws://" + httpTestServer.Listener.Addr().String()
		})

		AfterEach(func() {
			httpTestServer.Close()
		})

		It("returns messages from the server", func() {
			messagesToSend <- marshalMessage(createMessage("test-message-0", 0))
			messagesToSend <- marshalMessage(createMessage("test-message-1", 0))

			perform()

			Expect(err).NotTo(HaveOccurred())
			Expect(incomingMessages).To(HaveLen(2))
			Expect(incomingMessages[0].Message).To(Equal([]byte("test-message-0")))
			Expect(incomingMessages[1].Message).To(Equal([]byte("test-message-1")))
		})

		It("connects using failing proxyFunc", func() {
			errMsg := "Invalid consumerProxyFunc URL"
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return nil, errors.New(errMsg)
			}

			perform()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})
	})
})
