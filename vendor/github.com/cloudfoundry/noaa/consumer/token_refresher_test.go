package consumer_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry/noaa/consumer"

	. "github.com/apoydence/eachers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RefreshTokenFrom", func() {
	Context("Asynchronous", func() {
		var (
			cnsmr       *consumer.Consumer
			testHandler *errorRespondingHandler
			tcURL       string
			refresher   *mockTokenRefresher
		)

		BeforeEach(func() {
			testHandler = &errorRespondingHandler{
				subHandler:       NewWebsocketHandler(make(chan []byte), 100*time.Millisecond),
				responseStatuses: make(chan int, 10),
			}
			server := httptest.NewServer(testHandler)
			tcURL = "ws://" + server.Listener.Addr().String()

			refresher = newMockTokenRefresher()
			cnsmr = consumer.New(tcURL, nil, nil)

			cnsmr.RefreshTokenFrom(refresher)
		})

		Describe("TailingLogs", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.TailingLogs("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.TailingLogs("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("returns any error when fetching the token from the refresher", func() {
				errMsg := "Fetching authToken failed"
				refresher.RefreshAuthTokenOutput.Token <- ""
				refresher.RefreshAuthTokenOutput.AuthError <- errors.New(errMsg)

				_, errChan := cnsmr.TailingLogs("some-fake-app-guid", "")
				Eventually(errChan).Should(Receive(MatchError(errMsg)))
			})
		})

		Describe("TailingLogsWithoutReconnect", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.TailingLogsWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.TailingLogsWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})

		Describe("StreamWithoutReconnect", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.StreamWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.StreamWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})

		Describe("Stream", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.Stream("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.Stream("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})

		Describe("FirehoseWithoutReconnect", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.FirehoseWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.FirehoseWithoutReconnect("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})

		Describe("Firehose", func() {
			It("loads a token if the provided token is empty", func() {
				cnsmr.Firehose("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				testHandler.responseStatuses <- http.StatusUnauthorized
				cnsmr.Firehose("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})
	})

	Context("Synchronous", func() {
		var (
			cnsmr       *consumer.Consumer
			statuses    chan int
			testHandler http.Handler
			tcURL       string
			refresher   *mockTokenRefresher
		)

		BeforeEach(func() {
			statuses = make(chan int, 10)
			testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case status := <-statuses:
					w.WriteHeader(status)
				default:
					w.WriteHeader(http.StatusOK)
				}
			})
			server := httptest.NewServer(testHandler)
			tcURL = "ws://" + server.Listener.Addr().String()

			refresher = newMockTokenRefresher()
			refresher.RefreshAuthTokenOutput.Token <- "some-example-token"
			refresher.RefreshAuthTokenOutput.AuthError <- nil

			cnsmr = consumer.New(tcURL, nil, nil)

			cnsmr.RefreshTokenFrom(refresher)
		})

		Describe("RecentLogs", func() {
			It("uses the token refresher to obtain a new token", func() {
				cnsmr.RecentLogs("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				statuses <- http.StatusUnauthorized
				cnsmr.RecentLogs("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})

		Describe("ContainerMetrics", func() {
			It("uses the token refresher to obtain a new token", func() {
				cnsmr.ContainerMetrics("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})

			It("loads a token if the provided token fails with 401", func() {
				statuses <- http.StatusUnauthorized
				cnsmr.ContainerMetrics("some-fake-app-guid", "")
				Eventually(refresher.RefreshAuthTokenCalled).Should(BeCalled())
			})
		})
	})

	It("does not use the token refresher if an auth token is valid", func() {
		refresher := newMockTokenRefresher()

		cnsmr := consumer.New("fakeTrafficControllerURL", nil, nil)

		cnsmr.RefreshTokenFrom(refresher)

		cnsmr.TailingLogs("some-fake-app-guid", "someToken")
		Consistently(refresher.RefreshAuthTokenCalled).ShouldNot(BeCalled())
	})
})

type errorRespondingHandler struct {
	subHandler       http.Handler
	responseStatuses chan int
}

func (h *errorRespondingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case status := <-h.responseStatuses:
		w.WriteHeader(status)
	default:
		h.subHandler.ServeHTTP(w, r)
	}
}
