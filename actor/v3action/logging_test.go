package v3action_test

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"code.cloudfoundry.org/cli/actor/loggingaction/loggingactionfakes"
	"context"
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
		fakeLogCacheClient        *loggingactionfakes.FakeLogCacheClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, fakeConfig, _, _ = NewTestActor()
		fakeLogCacheClient = new(loggingactionfakes.FakeLogCacheClient)
		fakeConfig.AccessTokenReturns("AccessTokenForTest")
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
					[]ccv3.Application{
						{
							Name: "some-app",
							GUID: expectedAppGUID,
						},
					},
					ccv3.Warnings{"some-app-warnings"},
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

					return []*loggregator_v2.Envelope{
						{
							Timestamp:  int64(20),
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
						{
							Timestamp:  int64(10),
							SourceId:   "some-app-guid",
							InstanceId: "some-source-instance",
							Message: &loggregator_v2.Envelope_Log{
								Log: &loggregator_v2.Log{
									Payload: []byte("message-2"),
									Type:    loggregator_v2.Log_ERR,
								},
							},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						},
					}, ctx.Err()
				}
			})

			It("converts them to log messages and passes them through the messages channel", func() {
				var err error
				var warnings Warnings
				messages, logErrs, warnings, err, stopStreaming = actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				var message loggingaction.LogMessage
				Eventually(messages).Should(Receive(&message))
				Expect(message.Message).To(Equal("message-1"))
				Expect(message.MessageType).To(Equal("OUT"))
				Expect(message.Timestamp).To(Equal(time.Unix(0, 20)))
				Expect(message.SourceType).To(Equal("some-source-type"))
				Expect(message.SourceInstance).To(Equal("some-source-instance"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message).To(Equal("message-2"))
				Expect(message.MessageType).To(Equal("ERR"))
				Expect(message.Timestamp).To(Equal(time.Unix(0, 10)))
				Expect(message.SourceType).To(Equal("some-source-type"))
				Expect(message.SourceInstance).To(Equal("some-source-instance"))
			})
		})

		When("finding the application errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("ZOMG")
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv3.Warnings{"some-app-warnings"},
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
