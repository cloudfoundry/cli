package v2action_test

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"code.cloudfoundry.org/cli/actor/loggingaction/loggingactionfakes"
	"context"
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
		fakeLogCacheClient        *loggingactionfakes.FakeLogCacheClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, fakeConfig = NewTestActor()
		fakeLogCacheClient = new(loggingactionfakes.FakeLogCacheClient)
		fakeConfig.AccessTokenReturns("AccessTokenForTest")
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
					messages, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warnings"))

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

			When("Log Cache errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("ZOMG")
					fakeLogCacheClient.ReadReturns(nil, expectedErr)
				})

				It("returns error and warnings", func() {
					_, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
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
				_, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warnings"))
			})
		})
	})

	Describe("GetStreamingLogsForApplicationByNameAndSpace", func() {
		When("the application can be found", func() {
			var (
				expectedAppGUID string

				messages      <-chan loggingaction.LogMessage
				logErrs       <-chan error
				stopStreaming context.CancelFunc
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
						SourceId:   expectedAppGUID,
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
						SourceId:   expectedAppGUID,
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

			It("converts them to log messages and passes them through the messages channel", func() {
				var err error
				var warnings Warnings
				messages, logErrs, warnings, err, stopStreaming = actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
				defer stopStreaming()

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				var message loggingaction.LogMessage
				Eventually(messages).Should(Receive(&message))
				Expect(message.Message).To(Equal("message-1"))
				Eventually(messages).Should(Receive(&message))
				Expect(message.Message).To(Equal("message-2"))
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
				_, _, warnings, err, _ := actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				Expect(fakeLogCacheClient.ReadCallCount()).To(Equal(0))
			})
		})
	})
})
