package logs_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs with log cache", func() {

	Describe("RecentLogsFor", func() {
		It("retrieves the logs", func() {
			recentLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error) {
				message := *sharedaction.NewLogMessage(
					"some-message",
					"OUT",
					time.Unix(0, 0),
					"STG",
					"some-source-instance",
				)
				message2 := *sharedaction.NewLogMessage(
					"some-message2",
					"OUT",
					time.Unix(0, 0),
					"STG",
					"some-source-instance2",
				)
				return []sharedaction.LogMessage{message, message2}, nil
			}
			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				inputLogChan := make(chan sharedaction.LogMessage)
				errors := make(chan error)
				cancelFunc := func() {}
				return inputLogChan, errors, cancelFunc
			}
			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			logs, err := repository.RecentLogsFor("app-guid")

			Expect(err).ToNot(HaveOccurred())
			Expect(len(logs)).To(Equal(2))
			Expect(logs[0].ToSimpleLog()).To(Equal("some-message"))
			Expect(logs[1].ToSimpleLog()).To(Equal("some-message2"))
		})

		When("theres an error retrieving the logs", func() {
			It("returns the error", func() {
				recentLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error) {
					return nil, errors.New("some error")
				}
				getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
					inputLogChan := make(chan sharedaction.LogMessage)
					errors := make(chan error)
					cancelFunc := func() {}
					return inputLogChan, errors, cancelFunc
				}
				client := sharedactionfakes.FakeLogCacheClient{}
				repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
				_, err := repository.RecentLogsFor("app-guid")
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("TailLogsFor", func() {
		var (
			inputLogChan    chan sharedaction.LogMessage
			inputErrorChan  chan error
			logCacheMessage *logs.LogCacheMessage
			internalError   error
		)

		cancelFunc := func() {}

		recentLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error) {
			return []sharedaction.LogMessage{}, nil
		}

		It("writes to output log channel until input log message channel is closed", func() {
			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				inputLogChan = make(chan sharedaction.LogMessage, 2)
				inputErrorChan = make(chan error, 2)

				go func() {
					defer close(inputLogChan)

					message := *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"some-source-instance",
					)
					message2 := *sharedaction.NewLogMessage(
						"some-message2",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"some-source-instance2",
					)
					inputLogChan <- message
					inputLogChan <- message2

					time.Sleep(1 * time.Second)
				}()

				return inputLogChan, inputErrorChan, cancelFunc
			}

			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			outputLogChan := make(chan logs.Loggable, 2)
			outputErrorChan := make(chan error, 2)

			repository.TailLogsFor("app-guid", func() {}, outputLogChan, outputErrorChan)

			Eventually(outputLogChan).Should(HaveLen(2))
			Expect(outputLogChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
			Expect(outputLogChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message2"))
			Expect(inputErrorChan).ToNot(Receive())
		})

		It("writes to output log channel until input error channel is closed", func() {
			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				inputLogChan = make(chan sharedaction.LogMessage, 2)
				inputErrorChan = make(chan error, 2)

				go func() {
					defer close(inputErrorChan)

					message := *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"some-source-instance",
					)
					message2 := *sharedaction.NewLogMessage(
						"some-message2",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"some-source-instance2",
					)
					inputLogChan <- message
					inputLogChan <- message2

					time.Sleep(1 * time.Second)
				}()

				return inputLogChan, inputErrorChan, cancelFunc
			}

			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			outputLogChan := make(chan logs.Loggable, 2)
			outputErrorChan := make(chan error, 2)

			repository.TailLogsFor("app-guid", func() {}, outputLogChan, outputErrorChan)

			Eventually(outputLogChan).Should(HaveLen(2))
			Expect(outputLogChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
			Expect(outputLogChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message2"))
			Expect(inputErrorChan).ToNot(Receive())
		})

		It("writes recoverable errors to output error channel until input error channel is closed", func() {
			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				inputLogChan = make(chan sharedaction.LogMessage)
				inputErrorChan = make(chan error, 1)

				go func() {
					defer close(inputErrorChan)
					inputErrorChan <- errors.New("internal error")
					time.Sleep(1 * time.Second)
				}()

				return inputLogChan, inputErrorChan, cancelFunc
			}

			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			outputLogChan := make(chan logs.Loggable)
			outputErrorChan := make(chan error, 1)
			repository.TailLogsFor("app-guid", func() {}, outputLogChan, outputErrorChan)

			Eventually(outputErrorChan).Should(HaveLen(1))
			Expect(outputErrorChan).To(Receive(&internalError))
			Expect(internalError).To(MatchError("internal error"))
			Expect(inputLogChan).ToNot(Receive())
		})

		It("writes recoverable errors to output error channel until input log message channel is closed", func() {
			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				inputLogChan = make(chan sharedaction.LogMessage)
				inputErrorChan = make(chan error, 1)

				go func() {
					defer close(inputLogChan)
					inputErrorChan <- errors.New("internal error")
					time.Sleep(1 * time.Second)
				}()

				return inputLogChan, inputErrorChan, cancelFunc
			}

			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			outputLogChan := make(chan logs.Loggable)
			outputErrorChan := make(chan error, 1)
			repository.TailLogsFor("app-guid", func() {}, outputLogChan, outputErrorChan)

			Eventually(outputErrorChan).Should(HaveLen(1))
			Expect(outputErrorChan).To(Receive(&internalError))
			Expect(internalError).To(MatchError("internal error"))
			Expect(inputLogChan).ToNot(Receive())
		})

		When("Close() called", func() {

			It("calls repository.cancelFunc", func() {
				cancelFuncCallCount := 0

				cancelFunc = func() {
					cancelFuncCallCount += 1
				}

				getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
					inputLogChan = make(chan sharedaction.LogMessage)
					inputErrorChan = make(chan error, 1)
					close(inputLogChan)
					close(inputErrorChan)
					return inputLogChan, inputErrorChan, cancelFunc
				}

				client := sharedactionfakes.FakeLogCacheClient{}
				repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
				outputLogChan := make(chan logs.Loggable)
				outputErrorChan := make(chan error, 1)

				repository.TailLogsFor("app-guid", func() {}, outputLogChan, outputErrorChan)
				repository.Close()
				Expect(cancelFuncCallCount).To(Equal(1))
			})
		})
	})

	XDescribe("Authentication Token Refresh", func() {
	})
})
