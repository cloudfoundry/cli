package logs_test

import (
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
			client := sharedactionfakes.FakeLogCacheClient{}
			repository := logs.NewLogCacheRepository(&client, recentLogsFunc)
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
				client := sharedactionfakes.FakeLogCacheClient{}
				repository := logs.NewLogCacheRepository(&client, recentLogsFunc)
				_, err := repository.RecentLogsFor("app-guid")
				Expect(err).To(MatchError("some error"))
			})
		})
	})

	XDescribe("TailLogsFor", func() {
	})

	XDescribe("Authentication Token Refresh", func() {
	})
})
