package v2action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	noaaErrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
		fakeNOAAClient            *v2actionfakes.FakeNOAAClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, fakeConfig = NewTestActor()
		fakeNOAAClient = new(v2actionfakes.FakeNOAAClient)
	})

	Describe("LogMessage", func() {
		Describe("Staging", func() {
			When("the log is a staging log", func() {
				It("returns true", func() {
					message := NewLogMessage("", 0, time.Now(), "STG", "")
					Expect(message.Staging()).To(BeTrue())
				})
			})

			When("the log is any other kind of log", func() {
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

			messages <-chan *LogMessage
			errs     <-chan error

			message *LogMessage
		)

		BeforeEach(func() {
			expectedAppGUID = "some-app-guid"
		})

		AfterEach(func() {
			Eventually(messages).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		JustBeforeEach(func() {
			messages, errs = actor.GetStreamingLogs(expectedAppGUID, fakeNOAAClient)
		})

		When("receiving events", func() {
			BeforeEach(func() {
				fakeConfig.DialTimeoutReturns(60 * time.Minute)

				fakeNOAAClient.TailingLogsStub = func(appGUID string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					Expect(appGUID).To(Equal(expectedAppGUID))
					Expect(authToken).To(BeEmpty())

					Expect(fakeNOAAClient.SetOnConnectCallbackCallCount()).To(Equal(1))
					onConnectOrOnRetry := fakeNOAAClient.SetOnConnectCallbackArgsForCall(0)

					eventStream := make(chan *events.LogMessage)
					errStream := make(chan error, 1)

					go func() {
						defer close(eventStream)
						defer close(errStream)
						onConnectOrOnRetry()

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

						ts3 := int64(0)
						eventStream <- &events.LogMessage{
							Message:        []byte("message-3"),
							MessageType:    &outMessage,
							Timestamp:      &ts3,
							SourceType:     &sourceType,
							SourceInstance: &sourceInstance,
						}

						ts4 := int64(15)
						eventStream <- &events.LogMessage{
							Message:        []byte("message-4"),
							MessageType:    &errMessage,
							Timestamp:      &ts4,
							SourceType:     &sourceType,
							SourceInstance: &sourceInstance,
						}
					}()

					return eventStream, errStream
				}
			})

			It("converts them to log messages, sorts them, and passes them through the messages channel", func() {
				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-3"))
				Expect(message.Type()).To(Equal("OUT"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 0)))
				Expect(message.SourceType()).To(Equal("some-source-type"))
				Expect(message.SourceInstance()).To(Equal("some-source-instance"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-1"))
				Expect(message.Type()).To(Equal("OUT"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 10)))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-4"))
				Expect(message.Type()).To(Equal("ERR"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 15)))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-2"))
				Expect(message.Type()).To(Equal("ERR"))
				Expect(message.Timestamp()).To(Equal(time.Unix(0, 20)))
			})
		})

		When("receiving errors", func() {
			var (
				err1 error
				err2 error

				waiting chan bool
			)

			Describe("nil error", func() {
				BeforeEach(func() {
					fakeConfig.DialTimeoutReturns(time.Minute)

					waiting = make(chan bool)
					fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
						eventStream := make(chan *events.LogMessage)
						errStream := make(chan error, 1)

						Expect(fakeNOAAClient.SetOnConnectCallbackCallCount()).To(Equal(1))
						onConnectOrOnRetry := fakeNOAAClient.SetOnConnectCallbackArgsForCall(0)

						go func() {
							defer close(eventStream)
							defer close(errStream)
							onConnectOrOnRetry()

							errStream <- nil
							close(waiting)
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
					fakeConfig.DialTimeoutReturns(time.Microsecond) // tests don't care about this timeout, ignore it

					err1 = errors.New("ZOMG")
					err2 = errors.New("Fiddlesticks")

					fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
						eventStream := make(chan *events.LogMessage)
						errStream := make(chan error, 1)

						go func() {
							defer close(eventStream)
							defer close(errStream)

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
				When("NOAA is able to recover", func() {
					BeforeEach(func() {
						fakeConfig.DialTimeoutReturns(60 * time.Minute)

						fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
							eventStream := make(chan *events.LogMessage)
							errStream := make(chan error, 1)

							Expect(fakeNOAAClient.SetOnConnectCallbackCallCount()).To(Equal(1))
							onConnectOrOnRetry := fakeNOAAClient.SetOnConnectCallbackArgsForCall(0)

							go func() {
								defer close(eventStream)
								defer close(errStream)

								// can be called multiple times. Should be resilient to that
								onConnectOrOnRetry()
								errStream <- noaaErrors.NewRetryError(errors.New("error 1"))
								onConnectOrOnRetry()

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

				When("NOAA has trouble connecting", func() {
					BeforeEach(func() {
						fakeConfig.DialTimeoutReturns(time.Microsecond)
						fakeNOAAClient.TailingLogsStub = func(_ string, _ string) (<-chan *events.LogMessage, <-chan error) {
							eventStream := make(chan *events.LogMessage)
							errStream := make(chan error, 1)

							go func() {
								defer close(eventStream)
								defer close(errStream)

								// explicitly skip the on call to simulate ready never being triggered

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

					It("returns a NOAATimeoutError and continues", func() {
						Eventually(errs).Should(Receive(MatchError(actionerror.NOAATimeoutError{})))
						Eventually(messages).Should(Receive())

						Expect(fakeConfig.DialTimeoutCallCount()).To(Equal(1))
					})
				})
			})
		})
	})

	Describe("GetRecentLogsForApplicationByNameAndSpace", func() {
		When("the application can be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							Name: "some-app",
							GUID: "some-app-guid",
						},
					},
					ccv2.Warnings{"some-app-warnings"},
					nil,
				)
			})

			When("NOAA returns logs", func() {
				BeforeEach(func() {
					outMessage := events.LogMessage_OUT
					ts1 := int64(10)
					ts2 := int64(20)
					sourceType := "some-source-type"
					sourceInstance := "some-source-instance"

					var messages []*events.LogMessage
					messages = append(messages, &events.LogMessage{
						Message:        []byte("message-2"),
						MessageType:    &outMessage,
						Timestamp:      &ts2,
						SourceType:     &sourceType,
						SourceInstance: &sourceInstance,
					})
					messages = append(messages, &events.LogMessage{
						Message:        []byte("message-1"),
						MessageType:    &outMessage,
						Timestamp:      &ts1,
						SourceType:     &sourceType,
						SourceInstance: &sourceInstance,
					})

					fakeNOAAClient.RecentLogsReturns(messages, nil)
				})

				It("returns all the recent logs and warnings", func() {
					messages, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeNOAAClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warnings"))
					Expect(messages[0].Message()).To(Equal("message-1"))
					Expect(messages[0].Type()).To(Equal("OUT"))
					Expect(messages[0].Timestamp()).To(Equal(time.Unix(0, 10)))
					Expect(messages[0].SourceType()).To(Equal("some-source-type"))
					Expect(messages[0].SourceInstance()).To(Equal("some-source-instance"))

					Expect(messages[1].Message()).To(Equal("message-2"))
					Expect(messages[1].Type()).To(Equal("OUT"))
					Expect(messages[1].Timestamp()).To(Equal(time.Unix(0, 20)))
					Expect(messages[1].SourceType()).To(Equal("some-source-type"))
					Expect(messages[1].SourceInstance()).To(Equal("some-source-instance"))
				})
			})

			When("NOAA errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("ZOMG")
					fakeNOAAClient.RecentLogsReturns(nil, expectedErr)
				})

				It("returns error and warnings", func() {
					_, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeNOAAClient)
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warnings"))
				})
			})
		})

		When("finding the application errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("ZOMG")
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv2.Warnings{"some-app-warnings"},
					expectedErr,
				)
			})

			It("returns error and warnings", func() {
				_, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeNOAAClient)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				Expect(fakeNOAAClient.RecentLogsCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetStreamingLogsForApplicationByNameAndSpace", func() {
		When("the application can be found", func() {
			var (
				expectedAppGUID string

				messages <-chan *LogMessage
				logErrs  <-chan error
			)

			AfterEach(func() {
				Eventually(messages).Should(BeClosed())
				Eventually(logErrs).Should(BeClosed())
			})

			BeforeEach(func() {
				expectedAppGUID = "some-app-guid"

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							Name: "some-app",
							GUID: expectedAppGUID,
						},
					},
					ccv2.Warnings{"some-app-warnings"},
					nil,
				)

				fakeConfig.DialTimeoutReturns(60 * time.Minute)

				fakeNOAAClient.TailingLogsStub = func(appGUID string, authToken string) (<-chan *events.LogMessage, <-chan error) {
					Expect(appGUID).To(Equal(expectedAppGUID))
					Expect(authToken).To(BeEmpty())

					Expect(fakeNOAAClient.SetOnConnectCallbackCallCount()).To(Equal(1))
					onConnectOrOnRetry := fakeNOAAClient.SetOnConnectCallbackArgsForCall(0)

					eventStream := make(chan *events.LogMessage)
					errStream := make(chan error, 1)

					go func() {
						defer close(eventStream)
						defer close(errStream)

						onConnectOrOnRetry()

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
				var err error
				var warnings Warnings
				messages, logErrs, warnings, err = actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeNOAAClient)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warnings"))

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

		When("finding the application errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("ZOMG")
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv2.Warnings{"some-app-warnings"},
					expectedErr,
				)
			})

			It("returns error and warnings", func() {
				_, _, warnings, err := actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeNOAAClient)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				Expect(fakeNOAAClient.TailingLogsCallCount()).To(Equal(0))
			})
		})
	})
})
