package logs_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/sharedaction/sharedactionfakes"
	"code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
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
				logMessages := make(chan sharedaction.LogMessage)
				errors := make(chan error)
				cancelFunc := func() {}
				return logMessages, errors, cancelFunc
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
					logMessages := make(chan sharedaction.LogMessage)
					errors := make(chan error)
					cancelFunc := func() {}
					return logMessages, errors, cancelFunc
				}
				client := sharedactionfakes.FakeLogCacheClient{}
				repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
				_, err := repository.RecentLogsFor("app-guid")
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	Describe("TailLogsFor", func() {
		It("writes logs to the log channel", func() {
			var (
				logMessages chan sharedaction.LogMessage
				errorChan chan error
				logCacheMessage *logs.LogCacheMessage
			)

			recentLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) ([]sharedaction.LogMessage, error) {
				return []sharedaction.LogMessage{}, nil
			}

			getStreamingLogsFunc := func(appGUID string, client sharedaction.LogCacheClient) (<-chan sharedaction.LogMessage, <-chan error, context.CancelFunc) {
				logMessages = make(chan sharedaction.LogMessage, 2)
				errorChan = make(chan error, 2)
				cancelFunc := func() {}

				go func() {
					defer close(logMessages)
					defer close(errorChan)

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
					logMessages <- message
					logMessages <- message2

					time.Sleep(1 * time.Second)
				}()

				return logMessages, errorChan, cancelFunc
			}

			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc, getStreamingLogsFunc)
			logChan := make(chan logs.Loggable, 2)
			errChan := make(chan error, 2)
			repository.TailLogsFor("app-guid", func() {}, logChan, errChan)

			Eventually(logChan).Should(HaveLen(2))

			Expect(logChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
			Expect(logChan).To(Receive(&logCacheMessage))
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message2"))

			Expect(errorChan).ToNot(Receive())
		})

		It("writes errors to the error channel", func() {

		})
	})

	XDescribe("Authentication Token Refresh", func() {
	}) // }}
})
