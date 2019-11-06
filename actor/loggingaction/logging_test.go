package loggingaction_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/loggingaction"

	"code.cloudfoundry.org/cli/actor/loggingaction/loggingactionfakes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		fakeLogCacheClient *loggingactionfakes.FakeLogCacheClient
	)

	BeforeEach(func() {
		fakeLogCacheClient = new(loggingactionfakes.FakeLogCacheClient)
	})

	Describe("LogMessage", func() {
		Describe("Staging", func() {
			When("the log is a staging log", func() {
				It("returns true", func() {
					message := loggingaction.LogMessage{SourceType: "STG"}
					Expect(message.Staging()).To(BeTrue())
				})
			})

			When("the log is any other kind of log", func() {
				It("returns false", func() {
					message := loggingaction.LogMessage{SourceType: "APP"}
					Expect(message.Staging()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetStreamingLogs", func() {
		var (
			expectedAppGUID string

			messages      <-chan loggingaction.LogMessage
			errs          <-chan error
			stopStreaming context.CancelFunc
		)

		BeforeEach(func() {
			expectedAppGUID = "some-app-guid"
		})

		AfterEach(func() {
			Eventually(messages).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		JustBeforeEach(func() {
			messages, errs, stopStreaming = loggingaction.GetStreamingLogs(expectedAppGUID, fakeLogCacheClient)
		})

		When("receiving logs", func() {
			BeforeEach(func() {
				fakeLogCacheClient.ReadStub = func(
					ctx context.Context,
					sourceID string,
					start time.Time,
					opts ...logcache.ReadOption,
				) ([]*loggregator_v2.Envelope, error) {
					if fakeLogCacheClient.ReadCallCount() > 2 {
						stopStreaming()
					}

					return []*loggregator_v2.Envelope{{
						// 2 seconds in the past to get past Walk delay
						Timestamp:  time.Now().Add(-3 * time.Second).UnixNano(),
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
					}, {
						// 2 seconds in the past to get past Walk delay
						Timestamp:  time.Now().Add(-2 * time.Second).UnixNano(),
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
					}}, ctx.Err()
				}
			})

			It("converts them to log messages, sorts them, and passes them through the messages channel", func() {
				Eventually(messages).Should(HaveLen(4))
				var message loggingaction.LogMessage
				Expect(messages).To(Receive(&message))
				Expect(message.Message).To(Equal("message-1"))
				Expect(messages).To(Receive(&message))
				Expect(message.Message).To(Equal("message-2"))

				Expect(errs).ToNot(Receive())
			})
		})

		When("logs are older than 5 seconds", func() {
			var readStart chan time.Time

			BeforeEach(func() {
				readStart = make(chan time.Time, 100)
				fakeLogCacheClient.ReadStub = func(
					ctx context.Context,
					sourceID string,
					start time.Time,
					opts ...logcache.ReadOption,
				) ([]*loggregator_v2.Envelope, error) {
					if fakeLogCacheClient.ReadCallCount() > 1 {
						stopStreaming()
					}

					readStart <- start

					return []*loggregator_v2.Envelope{{
						// 2 seconds in the past to get past Walk delay
						Timestamp:  time.Now().Add(-6 * time.Second).UnixNano(),
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
					}, {
						// 2 seconds in the past to get past Walk delay
						Timestamp:  time.Now().Add(-2 * time.Second).UnixNano(),
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
					}}, ctx.Err()
				}
			})

			It("ignores them", func() {
				Eventually(readStart).Should(Receive(BeTemporally("~", time.Now().Add(-5*time.Second), time.Second)))
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
					if fakeLogCacheClient.ReadCallCount() > 2 {
						stopStreaming()
					}

					return nil, fmt.Errorf("error number %d", fakeLogCacheClient.ReadCallCount())
				}
			})

			It("passes them through the errors channel", func() {
				Eventually(errs).Should(HaveLen(2))
				Eventually(errs).Should(Receive(MatchError("error number 1")))
				Eventually(errs).Should(Receive(MatchError("error number 2")))
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
					messages, err := loggingaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).ToNot(HaveOccurred())

					Expect(messages[0].Message).To(Equal("message-1"))
					Expect(messages[0].MessageType).To(Equal("OUT"))
					Expect(messages[0].Timestamp).To(Equal(time.Unix(0, 10)))
					Expect(messages[0].SourceType).To(Equal("some-source-type"))
					Expect(messages[0].SourceInstance).To(Equal("some-source-instance"))

					Expect(messages[1].Message).To(Equal("message-2"))
					Expect(messages[1].MessageType).To(Equal("OUT"))
					Expect(messages[1].Timestamp).To(Equal(time.Unix(0, 20)))
					Expect(messages[1].SourceType).To(Equal("some-source-type"))
					Expect(messages[1].SourceInstance).To(Equal("some-source-instance"))
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
					messages, err := loggingaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
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
					_, err := loggingaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})
	})
})
