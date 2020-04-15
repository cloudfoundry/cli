package sharedaction_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	logcache "code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
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
			// Walk delay context: https://github.com/cloudfoundry/cli/blob/283d5fcdefa1806b24f4242adea1fb85871b4c6b/vendor/code.cloudfoundry.org/go-log-cache/walk.go#L74
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

					if start.IsZero() {
						return []*loggregator_v2.Envelope{&mostRecentEnvelope}, ctx.Err()
					}

					walkStartTime = start
					return []*loggregator_v2.Envelope{&slightlyOlderEnvelope, &mostRecentEnvelope}, ctx.Err()
				}
			})

			It("starts walking at the mostRecentEnvelope's time", func() {
				Eventually(messages).Should(BeClosed())
				Expect(walkStartTime).To(BeTemporally("~", mostRecentTime, time.Millisecond))
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
					return []*loggregator_v2.Envelope{}, ctx.Err()
				}
			})

			It("can be called multiple times", func() {
				Expect(stopStreaming).ToNot(Panic())
				Expect(stopStreaming).ToNot(Panic())
			})
		})

		Describe("error handling", func() {
			When("there is an error 'peeking' at log-cache to determine the latest log", func() {
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

				It("passes 5 errors through the errors channel", func() {
					Eventually(errs, 2*time.Second).Should(HaveLen(5))
					Eventually(errs).Should(Receive(MatchError("error number 1")))
					Eventually(errs).Should(Receive(MatchError("error number 2")))
					Eventually(errs).Should(Receive(MatchError("error number 3")))
					Eventually(errs).Should(Receive(MatchError("error number 4")))
					Eventually(errs).Should(Receive(MatchError("error number 5")))
					Consistently(errs).ShouldNot(Receive())
				})

				It("tries exactly 5 times", func() {
					Eventually(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(5))
					Consistently(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(5))
				})
			})

			When("there is an error walking log-cache to retrieve logs", func() {
				BeforeEach(func() {
					fakeLogCacheClient.ReadStub = func(
						ctx context.Context,
						sourceID string,
						start time.Time,
						opts ...logcache.ReadOption,
					) ([]*loggregator_v2.Envelope, error) {
						if start.IsZero() {
							return []*loggregator_v2.Envelope{&mostRecentEnvelope}, ctx.Err()
						}
						return nil, fmt.Errorf("error number %d", fakeLogCacheClient.ReadCallCount()-1)
					}
				})

				AfterEach(func() {
					stopStreaming()
				})

				It("passes 5 errors through the errors channel", func() {
					Eventually(errs, 2*time.Second).Should(HaveLen(5))
					Eventually(errs).Should(Receive(MatchError("error number 1")))
					Eventually(errs).Should(Receive(MatchError("error number 2")))
					Eventually(errs).Should(Receive(MatchError("error number 3")))
					Eventually(errs).Should(Receive(MatchError("error number 4")))
					Eventually(errs).Should(Receive(MatchError("error number 5")))
					Consistently(errs).ShouldNot(Receive())
				})

				It("tries exactly 5 times", func() {
					initialPeekingRead := 1
					walkRetries := 5
					expectedReadCallCount := initialPeekingRead + walkRetries

					Eventually(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(expectedReadCallCount))
					Consistently(fakeLogCacheClient.ReadCallCount, 2*time.Second).Should(Equal(expectedReadCallCount))
				})
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
				BeforeEach(func() {
					fakeLogCacheClient.ReadReturns(nil, errors.New("some-recent-logs-error"))
				})

				It("returns error and warnings", func() {
					_, err := sharedaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)
					Expect(err).To(MatchError("Failed to retrieve logs from Log Cache: some-recent-logs-error"))
				})
			})

			When("Log Cache returns a resource-exhausted error from grpc", func() {
				resourceExhaustedErr := errors.New("unexpected status code 429")
				u := new(url.URL)
				v := make(url.Values)

				BeforeEach(func() {
					fakeLogCacheClient.ReadReturns([]*loggregator_v2.Envelope{}, resourceExhaustedErr)
				})

				It("attempts to halve numbber of requested logs, and eventually returns error and warnings", func() {
					_, err := sharedaction.GetRecentLogs("some-app-guid", fakeLogCacheClient)

					Expect(err).To(MatchError("Failed to retrieve logs from Log Cache: unexpected status code 429"))
					Expect(fakeLogCacheClient.ReadCallCount()).To(Equal(10))

					_, _, _, readOptions := fakeLogCacheClient.ReadArgsForCall(0)
					readOptions[1](u, v)
					Expect(v.Get("limit")).To(Equal("1000"))

					_, _, _, readOptions = fakeLogCacheClient.ReadArgsForCall(1)
					readOptions[1](u, v)
					Expect(v.Get("limit")).To(Equal("500"))
				})
			})
		})
	})

})
