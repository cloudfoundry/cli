package noaa_test

import (
	"bytes"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/server/handlers"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/elazarl/goproxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Noaa behind a Proxy", func() {
	var (
		connection        *noaa.Consumer
		endpoint          string
		testServer        *httptest.Server
		tlsSettings       *tls.Config
		consumerProxyFunc func(*http.Request) (*url.URL, error)

		appGuid         string
		authToken       string
		incomingChan    chan *events.Envelope
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

	Describe("StreamWithoutReconnect", func() {
		var errorChan chan error
		var finishedChan chan struct{}

		BeforeEach(func() {
			errorChan = make(chan error, 10)
			finishedChan = make(chan struct{})
			incomingChan = make(chan *events.Envelope)
		})

		AfterEach(func() {
			close(messagesToSend)
			<-finishedChan
		})

		perform := func() {
			connection = noaa.NewConsumer(endpoint, tlsSettings, consumerProxyFunc)

			go func() {
				errorChan <- connection.StreamWithoutReconnect(appGuid, authToken, incomingChan)
				close(finishedChan)
			}()
		}

		It("connects using valid URL to running consumerProxyFunc server", func() {
			messagesToSend <- marshalMessage(createMessage("hello", 0))
			perform()

			message := <-incomingChan

			Expect(message.GetLogMessage().GetMessage()).To(Equal([]byte("hello")))
		})

		It("connects using valid URL to a stopped consumerProxyFunc server", func() {
			testProxyServer.Close()

			perform()

			for {
				select {
				case err := <-errorChan:
					if strings.Contains(err.Error(), "connection refused") {
						return
					}
				case <-time.After(time.Second):
					Fail("never received an error")
				}
			}
		})

		It("connects using invalid URL", func() {
			errMsg := "Invalid consumerProxyFunc URL"
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return nil, errors.New(errMsg)
			}

			perform()

			var err error
			Eventually(errorChan).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})

		It("connects to a consumerProxyFunc server rejecting CONNECT requests", func() {
			goProxyHandler.OnRequest().HandleConnect(goproxy.AlwaysReject)

			perform()

			var err error
			Eventually(errorChan).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error dialing traffic controller server"))
		})

		It("connects to a non-consumerProxyFunc server", func() {
			nonProxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Go away, I am not a consumerProxyFunc!", http.StatusBadRequest)
			}))
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return url.Parse(nonProxyServer.URL)
			}

			perform()

			var err error
			Eventually(errorChan).Should(Receive(&err))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(http.StatusText(http.StatusBadRequest)))
		})
	})

	Describe("RecentLogs", func() {
		var httpTestServer *httptest.Server
		var incomingMessages []*events.LogMessage

		perform := func() {
			close(messagesToSend)
			connection = noaa.NewConsumer(endpoint, tlsSettings, consumerProxyFunc)
			incomingMessages, err = connection.RecentLogs(appGuid, authToken)
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
			Expect(incomingMessages[0].GetMessage()).To(Equal([]byte("test-message-0")))
			Expect(incomingMessages[1].GetMessage()).To(Equal([]byte("test-message-1")))
		})

		It("connects using failing proxyFunc", func() {
			errMsg := "Invalid consumerProxyFunc URL"
			consumerProxyFunc = func(*http.Request) (*url.URL, error) {
				return nil, errors.New(errMsg)
			}

			perform()

			Expect(err).To(HaveOccurred(), "THIS WILL FAIL ON GOLANG 1.3 - 1.3.3 DUE TO BUG IN STANDARD LIBRARY (see https://code.google.com/p/go/issues/detail?id=8755)")
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})
	})
})
