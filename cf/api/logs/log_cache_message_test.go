package logs_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("log messages", func() {
	message := *sharedaction.NewLogMessage(
		"some-message",
		"OUT",
		time.Unix(0, 0),
		"STG",
		"some-source-instance",
	)
	logCacheMessage := logs.NewLogCacheMessage(message)

	Describe("ToSimpleLog", func() {
		It("returns the message", func() {
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
		})
	})

	Describe("GetSourceName", func() {
		It("returns the message", func() {
			// TODO: reconsider whether this should return SourceType or SourceInstance
			Expect(logCacheMessage.GetSourceName()).To(Equal("STG"))
		})
	})

	Describe("ToLog", func() {
		// timezone handling
		// sourceID sets the logHeader format
		// coloring of log header on logType value
		// trimming of message text
		// padding calculation
		// multi-line log messages
		// handles both STDERR and STDOUT differently

		It("correctly formats the log", func() {
			Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [STG/0] OUT some-message"))
		})
	})
})
