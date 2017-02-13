package v2action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor          Actor
		fakeNOAAClient *v2actionfakes.FakeNOAAClient
	)

	BeforeEach(func() {
		fakeNOAAClient = new(v2actionfakes.FakeNOAAClient)
		actor = NewActor(nil, nil)
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
				messages, errs = actor.GetStreamingLogs(expectedAppGUID, fakeNOAAClient)

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
			var err1, err2 error

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
				messages, errs = actor.GetStreamingLogs(expectedAppGUID, fakeNOAAClient)

				err := <-errs
				Expect(err).To(MatchError(err1))

				err = <-errs
				Expect(err).To(MatchError(err2))
			})
		})
	})
})
