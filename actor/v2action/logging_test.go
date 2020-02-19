package v2action_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeUAAClient             *v2actionfakes.FakeUAAClient
		fakeConfig                *v2actionfakes.FakeConfig

		fakeLogCacheClient *sharedactionfakes.FakeLogCacheClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, fakeUAAClient, fakeConfig = NewTestActor()
		fakeLogCacheClient = new(sharedactionfakes.FakeLogCacheClient)
		fakeConfig.AccessTokenReturns("AccessTokenForTest")
	})

	Describe("LogMessage", func() {
		Describe("Staging", func() {
			When("the log is a staging log", func() {
				It("returns true", func() {
					message := sharedaction.NewLogMessage("", "OUT", time.Now(), "STG", "")
					Expect(message.Staging()).To(BeTrue())
				})
			})

			When("the log is any other kind of log", func() {
				It("returns true", func() {
					message := sharedaction.NewLogMessage("", "OUT", time.Now(), "APP", "")
					Expect(message.Staging()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetStreamingLogs", func() {
		var (
			expectedAppGUID string

			messages      <-chan sharedaction.LogMessage
			errs          <-chan error
			stopStreaming context.CancelFunc

			message sharedaction.LogMessage
		)

		BeforeEach(func() {
			expectedAppGUID = "some-app-guid"
		})

		AfterEach(func() {
			Eventually(messages).Should(BeClosed())
			Eventually(errs).Should(BeClosed())
		})

		JustBeforeEach(func() {
			messages, errs, stopStreaming = actor.GetStreamingLogs(expectedAppGUID, fakeLogCacheClient)
		})

		When("receiving events", func() {
			BeforeEach(func() {
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
							// 3 seconds in the past to get past Walk delay
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
						}, {
							Timestamp:  time.Now().Add(-4 * time.Second).UnixNano(),
							SourceId:   expectedAppGUID,
							InstanceId: "some-source-instance",
							Message: &loggregator_v2.Envelope_Log{
								Log: &loggregator_v2.Log{
									Payload: []byte("message-3"),
									Type:    loggregator_v2.Log_OUT,
								},
							},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						}, {
							Timestamp:  time.Now().Add(-1 * time.Second).UnixNano(),
							SourceId:   expectedAppGUID,
							InstanceId: "some-source-instance",
							Message: &loggregator_v2.Envelope_Log{
								Log: &loggregator_v2.Log{
									Payload: []byte("message-4"),
									Type:    loggregator_v2.Log_OUT,
								},
							},
							Tags: map[string]string{
								"source_type": "some-source-type",
							},
						},
						},
						ctx.Err()
				}
			})

			It("converts them to log messages, does not sort them, and passes them through the messages channel", func() {
				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-1"))
				Expect(message.Type()).To(Equal("OUT"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-2"))
				Expect(message.Type()).To(Equal("OUT"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-3"))
				Expect(message.Type()).To(Equal("OUT"))

				Expect(message.SourceType()).To(Equal("some-source-type"))
				Expect(message.SourceInstance()).To(Equal("some-source-instance"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-4"))
				Expect(message.Type()).To(Equal("OUT"))
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

			When("LogCache returns logs", func() {
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
					fakeConfig.DialTimeoutReturns(60 * time.Minute)

					fakeLogCacheClient.ReadReturns([]*loggregator_v2.Envelope{{
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
					}},
						nil,
					)
				})

				It("returns all the recent logs and warnings", func() {
					messages, warnings, err := actor.GetRecentLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warnings"))
					Expect(messages[0].Message()).To(Equal("message-2"))
					Expect(messages[0].Type()).To(Equal("OUT"))
					Expect(messages[0].SourceType()).To(Equal("some-source-type"))
					Expect(messages[0].SourceInstance()).To(Equal("some-source-instance"))

					Expect(messages[1].Message()).To(Equal("message-1"))
					Expect(messages[1].Type()).To(Equal("OUT"))
					Expect(messages[1].SourceType()).To(Equal("some-source-type"))
					Expect(messages[1].SourceInstance()).To(Equal("some-source-instance"))
				})
			})

			When("LogCache errors", func() {
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

				Expect(fakeLogCacheClient.ReadCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetStreamingLogsForApplicationByNameAndSpace", func() {
		When("the application can be found", func() {
			var (
				expectedAppGUID string

				messages      <-chan sharedaction.LogMessage
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
								Type:    loggregator_v2.Log_ERR,
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
				var message sharedaction.LogMessage

				messages, logErrs, stopStreaming, warnings, err = actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-1"))
				Expect(message.Type()).To(Equal("OUT"))
				Expect(message.SourceType()).To(Equal("some-source-type"))
				Expect(message.SourceInstance()).To(Equal("some-source-instance"))

				Eventually(messages).Should(Receive(&message))
				Expect(message.Message()).To(Equal("message-2"))
				Expect(message.Type()).To(Equal("ERR"))
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
				_, _, _, warnings, err := actor.GetStreamingLogsForApplicationByNameAndSpace("some-app", "some-space-guid", fakeLogCacheClient)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warnings"))

				Expect(fakeLogCacheClient.ReadCallCount()).To(Equal(0))
			})
		})
	})

	Describe("ScheduleTokenRefresh", func() {
		var (
			stop                   chan struct{}
			stoppedRefreshingToken chan struct{}
			ticker                 chan time.Time
			after                  func(time.Duration) <-chan time.Time
			delay                  chan time.Duration
			errChannel             <-chan error
			err                    error
		)

		BeforeEach(func() {
			fakeConfig.RefreshTokenReturns(helpers.BuildTokenString(time.Now().Add(5 * time.Minute)))
			ticker = make(chan time.Time)
			delay = make(chan time.Duration, 100)
			after = func(t time.Duration) <-chan time.Time {
				delay <- t
				return ticker
			}
		})

		JustBeforeEach(func() {
			stop = make(chan struct{})
			stoppedRefreshingToken = make(chan struct{})
			errChannel, err = actor.ScheduleTokenRefresh(after, stop, stoppedRefreshingToken)
		})

		AfterEach(func() {
			if stop != nil {
				close(stop)
			}
		})

		When("the access token is not expiring soon", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(5 * time.Minute)))
			})

			It("does not refresh the access token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))
			})
		})

		When("the access token is expiring soon", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(30 * time.Second)))
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: helpers.BuildTokenString(time.Now().Add(5 * time.Minute)),
					Type:        "bearer",
				}, nil)
			})

			It("refreshes the access token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
			})
		})

		When("the access token has already expired", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(-30 * time.Second)))
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: helpers.BuildTokenString(time.Now().Add(5 * time.Minute)),
					Type:        "bearer",
				}, nil)
			})

			It("refreshes the access token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
			})

			When("and the attempt to refresh the access token fails", func() {
				BeforeEach(func() {
					fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{}, errors.New("some error"))
				})

				It("will not refresh the access token", func() {
					Expect(err).To(MatchError("some error"))
				})
				It("closes a channel to indicate it has finished", func() {
					close(stop)
					Eventually(stoppedRefreshingToken).Should(BeClosed())
				})
				AfterEach(func() {
					stop = nil
				})
			})
		})

		When("the access token is invalid", func() {
			BeforeEach(func() {
				fakeConfig.AccessTokenReturns("bogus-access-token")
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: helpers.BuildTokenString(time.Now().Add(5 * time.Minute)),
					Type:        "bearer",
				}, nil)
			})
			It("refreshes the access token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(1))
			})
		})

		When("the access token expires while we are streaming logs", func() {

			BeforeEach(func() {
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(2 * time.Minute)))
				fakeConfig.RefreshTokenReturns(helpers.BuildTokenString(time.Now().Add(2 * time.Minute)))
			})

			JustBeforeEach(func() {
				stop = make(chan struct{})
				stoppedRefreshingToken = make(chan struct{})
				errChannel, err = actor.ScheduleTokenRefresh(after, stop, stoppedRefreshingToken)
			})

			It("refreshes the access token", func() {
				Expect(err).NotTo(HaveOccurred())

				By("not initially refreshing the token when it is unnecessary")
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))

				By("refreshing the first access token when it is expiring soon")
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(30 * time.Second)))
				ticker <- time.Time{}
				Eventually(fakeUAAClient.RefreshAccessTokenCallCount).Should(Equal(1))

				By("not refreshing the second access token when it is not close to expiry")
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(90 * time.Second)))
				ticker <- time.Time{}
				Consistently(fakeUAAClient.RefreshAccessTokenCallCount).Should(Equal(1))

				By("refreshing the second access token when it is near to expiry")
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(30 * time.Second)))
				ticker <- time.Time{}
				Eventually(fakeUAAClient.RefreshAccessTokenCallCount).Should(Equal(2))
			})

			It("sleeps until the token is approaching expiry", func() {
				Expect(err).NotTo(HaveOccurred())

				By("not initially refreshing the token when it is unnecessary")
				Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))

				By("refreshing the first access token when it is expiring soon")
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(30 * time.Second)))
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: helpers.BuildTokenString(time.Now().Add(30 * time.Second)),
					Type:        "bearer",
				}, nil)

				ticker <- time.Time{}
				Eventually(delay).Should(Receive(BeNumerically("~", 27*time.Second, time.Second)))

				By("using the expiry time of the token to calculate the sleep delay")
				fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(10 * time.Second)))
				fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{
					AccessToken: helpers.BuildTokenString(time.Now().Add(10 * time.Second)),
					Type:        "bearer",
				}, nil)
				ticker <- time.Time{}
				Eventually(delay).Should(Receive(BeNumerically("~", 9*time.Second, time.Second)))
			})

			When("and the token refresh errors", func() {
				BeforeEach(func() {
					fakeUAAClient.RefreshAccessTokenReturns(uaa.RefreshedTokens{}, errors.New("some error"))
				})
				It("propagates the error to the caller", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUAAClient.RefreshAccessTokenCallCount()).To(Equal(0))
					fakeConfig.AccessTokenReturns(helpers.BuildTokenString(time.Now().Add(30 * time.Second)))
					go func() {
						defer GinkgoRecover()
						Eventually(errChannel).Should(Receive(MatchError("some error")))
					}()
					ticker <- time.Time{}
					Eventually(fakeUAAClient.RefreshAccessTokenCallCount).Should(Equal(1))
				})
			})

			When("token refreshing is stopped", func() {
				It("closes a channel to indicate it has finished", func() {
					close(stop)
					Eventually(stoppedRefreshingToken).Should(BeClosed())
				})
				AfterEach(func() {
					stop = nil
				})
			})

		})

	})
})
