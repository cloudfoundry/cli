package loggingaction_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/loggingaction"

	"code.cloudfoundry.org/cli/actor/loggingaction/loggingactionfakes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
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
