package sharedaction_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"

	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient
	)

	BeforeEach(func() {
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
	})

	Describe("LogMessage", func() {
		Describe("Staging", func() {
			When("the log is a staging log", func() {
				It("returns true", func() {
					message := *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"some-source-instance",
					)

					Expect(message.Staging()).To(BeTrue())
				})
			})

			When("the log is any other kind of log", func() {
				It("returns false", func() {
					message := *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"APP",
						"some-source-instance",
					)
					Expect(message.Staging()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetStreamingLogs", func() {
		var (
			expectedAppGUID string

			messages              <-chan sharedaction.LogMessage
			errs                  <-chan error
			stopStreaming         context.CancelFunc
			mostRecentTime        time.Time
			mostRecentEnvelope    loggregator_v2.Envelope
			slightlyOlderEnvelope loggregator_v2.Envelope
		)

		BeforeEach(func() {
			expectedAppGUID = "some-app-guid"
			// 2 seconds in the past to get past Walk delay
			// Walk delay context: https://github.com/cloudfoundry/cli/blob/b8324096a3d5a495bdcae9d1e7f6267ff135fe82/vendor/code.cloudfoundry.org/log-cache/pkg/client/walk.go#L74
			mostRecentTime = time.Now().Add(-2 * time.Second)
			mostRecentTimestamp := mostRecentTime.UnixNano()
			slightlyOlderTimestamp := mostRecentTime.Add(-500 * time.Millisecond).UnixNano()

			mostRecentEnvelope = loggregator_v2.Envelope{
				Timestamp:  mostRecentTimestamp,
				SourceId:   "some-app-guid",
				InstanceId: "some-source-instance",
				Message: &loggregator_v2.Envelope_Log{
					Log: &loggregator_v2.Log{
						Payload: []byte("message-2"),
						Type:    loggregator_v2.Log_OUT,
					},
				},
				Tags: map[string]string{
					"source_type": "some-source-type",
				},
			}

			slightlyOlderEnvelope = loggregator_v2.Envelope{
				Timestamp:  slightlyOlderTimestamp,
				SourceId:   "some-app-guid",
				InstanceId: "some-source-instance",
				Message: &loggregator_v2.Envelope_Log{
					Log: &loggregator_v2.Log{
						Payload: []byte("message-1"),
						Type:    loggregator_v2.Log_OUT,
					},
				},
				Tags: map[string]string{
					"source_type": "some-source-type",
				},
			}
		})

		AfterEach(func() {
			Eventually(messages).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		JustBeforeEach(func() {
			messages, errs, stopStreaming = sharedaction.GetStreamingLogs(expectedAppGUID, fakeLogCacheClient)
		})

		When("receiving logs", func() {
			var walkStartTime time.Time

			BeforeEach(func() {
				fakeLogCacheClient.ReadStub = func(
					ctx context.Context,
					sourceID string,
					start time.Time,
					opts ...logcache.ReadOption,
				) ([]*loggregator_v2.Envelope, error) {
					if fakeLogCacheClient.ReadCallCount() > 2 {
						stopStreaming()
						return []*loggregator_v2.Envelope{}, ctx.Err()
					}

					if (start == time.Time{}) {
						return []*loggregator_v2.Envelope{&mostRecentEnvelope}, ctx.Err()
					}

					walkStartTime = start
					return []*loggregator_v2.Envelope{&slightlyOlderEnvelope, &mostRecentEnvelope}, ctx.Err()
				}
			})

			It("it starts walking at 1 second previous to the mostRecentEnvelope's time", func() {
				Eventually(messages).Should(BeClosed())
				Expect(walkStartTime).To(BeTemporally("~", mostRecentTime.Add(-1*time.Second), time.Millisecond))
			})

			It("converts them to log messages and passes them through the messages channel", func() {
				Eventually(messages).Should(HaveLen(2))
				var message sharedaction.LogMessage
				Expect(messages).To(Receive(&message))
				Expect(message.Message()).To(Equal("message-1"))
				Expect(messages).To(Receive(&message))
				Expect(message.Message()).To(Equal("message-2"))

				Expect(errs).ToNot(Receive())
			})
		})

		When("cancelling log streaming", func() {
			BeforeEach(func() {
				fakeLogCacheClient.ReadStub = func(
					ctx context.Context,
					sourceID string,
					start time.Time,
					opts ...logcache.ReadOption,
				) ([]*loggregator_v2.Envelope, error) {
					return nil, ctx.Err()
				}
			})

			It("can be called multiple times", func() {
				Expect(stopStreaming).ToNot(Panic())
				Expect(stopStreaming).ToNot(Panic())
			})
		})

		Describe("log cache error", func() {
			BeforeEach(func() {
				fakeLogCacheClient.ReadStub = func(
					ctx context.Context,
					sourceID string,
					start time.Time,
					opts ...logcache.ReadOption,
				) ([]*loggregator_v2.Envelope, error) {
					return nil, fmt.Errorf("error number %d", fakeLogCacheClient.ReadCallCount())
				}
			})

			AfterEach(func() {
				stopStreaming()
			})

			// Error handling is now done by the initial client.Read() call
			// Error handling used to be done by the client.Walk() function, which had retry logic
			It("passes one error through the errors channel", func() {
				Eventually(errs, 2*time.Second).Should(HaveLen(1))
				Eventually(errs).Should(Receive(MatchError("error number 1")))
				Eventually(errs).ShouldNot(Receive(MatchError("error number 2")))
			})

			It("retries exactly 1 times", func() {
				Eventually(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(1))
				Consistently(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(1))
			})
		})
	})

	Describe("GetRecentLogs", func() {
		When("the application can be found", func() {
			When("Log Cache returns logs", func() {
				BeforeEach(func() {
					messages := []*loggregator_v2.Envelope{
						{
							Timestamp:  int64(20),
							SourceId:   "some-app-guid",
							InstanceId: "some-source-instance",
							Message: &loggregator_v2.Envelope_Log{
								Log: &loggregator_v2.Log{
									Payload: []byte("message-2"),
									Type:    loggregator_v2.Log_OUT,
								},
							},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						},
						{
							Timestamp:  int64(10),
							SourceId:   "some-app-guid",
							InstanceId: "some-source-instance",
							Message: &loggregator_v2.Envelope_Log{
								Log: &loggregator_v2.Log{
									Payload: []byte("message-1"),
									Type:    loggregator_v2.Log_OUT,
								},
							},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						},
					}

					fakeLogCacheClient.ReadReturns(messages, nil)
				})

				It("returns all the recent logs and warnings", func() {
					messages, err := sharedaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).ToNot(HaveOccurred())

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

			When("Log Cache returns non-log envelopes", func() {
				BeforeEach(func() {
					messages := []*loggregator_v2.Envelope{
						{
							Timestamp:  int64(10),
							SourceId:   "some-app-guid",
							InstanceId: "some-source-instance",
							Message:    &loggregator_v2.Envelope_Counter{},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						},
					}

					fakeLogCacheClient.ReadReturns(messages, nil)
				})

				It("ignores them", func() {
					messages, err := sharedaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(messages).To(BeEmpty())
				})
			})

			When("Log Cache errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("ZOMG")
					fakeLogCacheClient.ReadReturns(nil, expectedErr)
				})

				It("returns error and warnings", func() {
					_, err := sharedaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})
	})

})
