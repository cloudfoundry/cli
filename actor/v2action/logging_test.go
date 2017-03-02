package v2action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	noaaErrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor          Actor
		fakeNOAAClient *v2actionfakes.FakeNOAAClient
		fakeConfig     *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeNOAAClient = new(v2actionfakes.FakeNOAAClient)
		fakeConfig = new(v2actionfakes.FakeConfig)
		actor = NewActor(nil, nil)
	})

	Describe("LogMessage", func() {
		Describe("Stagging", func() {
			Context("when the log is a staging log", func() {
				It("returns true", func() {
					message := NewLogMessage("", 0, time.Now(), "STG", "")
					Expect(message.Staging()).To(BeTrue())
				})
			})

			Context("when the log is any other kind of log", func() {
				It("returns true", func() {
					message := NewLogMessage("", 0, time.Now(), "APP", "")
					Expect(message.Staging()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetStreamingLogs", func() {
		var (
			expectedAppGUID string

			messages    <-chan *LogMessage
			errs        <-chan error
			eventStream chan *events.LogMessage
			errStream   chan error
		)

		BeforeEach(func() {
			expectedAppGUID = "some-app-guid"

			eventStream = make(chan *events.LogMessage)
			errStream = make(chan error)
		})

		// If tests panic due to this close, it is likely you have a failing
		// expectation and the channels are being closed because the test has
		// failed/short circuited and is going through teardown.
		AfterEach(func() {
			close(eventStream)
			close(errStream)

			Eventually(messages).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		JustBeforeEach(func() {
			messages, errs = actor.GetStreamingLogs(expectedAppGUID, fakeNOAAClient, fakeConfig)
		})

		Context("when receiving events", func() {
			BeforeEach(func() {
				fakeNOAAClient.TailingLogsStub = func(appGUID string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					Expect(appGUID).To(Equal(expectedAppGUID))
					Expect(authToken).To(BeEmpty())

					go func() {
						outMessage := events.LogMessage_OUT
						ts1 := int64(10)
						sourceType := "some-source-type"
						sourceInstance := "some-source-instance"

						eventStream <- &events.LogMessage{
							Message:        []byte("message-1"),
							MessageType:    &outMessage,
							Timestamp:      &ts1,
							SourceType:     &sourceType,
							SourceInstance: &sourceInstance,
						}

						errMessage := events.LogMessage_ERR
						ts2 := int64(20)

						eventStream <- &events.LogMessage{
							Message:        []byte("message-2"),
							MessageType:    &errMessage,
							Timestamp:      &ts2,
							SourceType:     &sourceType,
							SourceInstance: &sourceInstance,
						}
					}()

					return eventStream, errStream
				}
			})

			It("converts them to log messages and passes them through the messages channel", func() {
				message := <-messages
				Expect(message.Message()).To(Equal("message-1"))
				Expect(message.Type()).To(Equal("OUT"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 10)))
				Expect(message.SourceType()).To(Equal("some-source-type"))
				Expect(message.SourceInstance()).To(Equal("some-source-instance"))

				message = <-messages
				Expect(message.Message()).To(Equal("message-2"))
				Expect(message.Type()).To(Equal("ERR"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 20)))
				Expect(message.SourceType()).To(Equal("some-source-type"))
				Expect(message.SourceInstance()).To(Equal("some-source-instance"))
			})
		})

		Context("when receiving errors", func() {
			var (
				err1 error
				err2 error

				waiting chan bool
			)

			Describe("nil error", func() {
				BeforeEach(func() {
					waiting = make(chan bool)
					fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
						go func() {
							close(waiting)
							errStream <- nil
						}()

						return eventStream, errStream
					}
				})

				It("does not pass the nil along", func() {
					Eventually(waiting).Should(BeClosed())
					Consistently(errs).ShouldNot(Receive())
				})
			})

			Describe("unexpected error", func() {
				BeforeEach(func() {
					err1 = errors.New("ZOMG")
					err2 = errors.New("Fiddlesticks")

					fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
						go func() {
							errStream <- err1
							errStream <- err2
						}()

						return eventStream, errStream
					}
				})

				It("passes them through the errors channel", func() {
					Eventually(errs).Should(Receive(Equal(err1)))
					Eventually(errs).Should(Receive(Equal(err2)))
				})
			})

			Describe("NOAA's RetryError", func() {
				Context("when NOAA is able to recover", func() {
					BeforeEach(func() {
						fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
							go func() {
								errStream <- noaaErrors.NewRetryError(errors.New("error 1"))

								outMessage := events.LogMessage_OUT
								ts1 := int64(10)
								sourceType := "some-source-type"
								sourceInstance := "some-source-instance"

								eventStream <- &events.LogMessage{
									Message:        []byte("message-1"),
									MessageType:    &outMessage,
									Timestamp:      &ts1,
									SourceType:     &sourceType,
									SourceInstance: &sourceInstance,
								}
							}()

							return eventStream, errStream
						}
					})

					It("continues without issue", func() {
						Eventually(messages).Should(Receive())
						Consistently(errs).ShouldNot(Receive())
					})
				})

				Context("when NOAA is unable to recover", func() {
					BeforeEach(func() {
						Skip("waiting on NOAA #140000891")
						fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
							go func() {
								errStream <- noaaErrors.NewRetryError(errors.New("error 1"))
								errStream <- noaaErrors.NewRetryError(errors.New("error 2"))
							}()

							return eventStream, errStream
						}
					})

					It("returns a NOAATimeoutError", func() {
						Eventually(errs).Should(Receive(MatchError(NOAATimeoutError{})))
					})
				})
			})
		})
	})
})
