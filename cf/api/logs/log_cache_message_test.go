package logs_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/cf/api/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("log messages", func() {
	// TODO: review if that is the only way to initialize variables of these types
	message := *sharedaction.NewLogMessage(
		"some-message",
		"OUT",
		time.Unix(0, 0),
		"APP/PROC/WEB",
		"0",
	)
	logCacheMessage := logs.NewLogCacheMessage(message)

	BeforeEach(func() {
		message = *sharedaction.NewLogMessage(
			"some-message",
			"OUT",
			time.Unix(0, 0),
			"APP/PROC/WEB",
			"0",
		)
		logCacheMessage = logs.NewLogCacheMessage(message)
	})

	Describe("ToSimpleLog", func() {
		It("returns the message", func() {
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
		})
	})

	Describe("GetSourceName", func() {
		It("returns the message", func() {
			// TODO: reconsider whether this should return SourceType or SourceInstance
			Expect(logCacheMessage.GetSourceName()).To(Equal("APP/PROC/WEB"))
		})
	})

	Describe("ToLog", func() {
		Context("timezone handling", func() {
			It("reflects timezone in log output", func() {
				Expect(logCacheMessage.ToLog(time.FixedZone("TST", 1*60*60))).To(ContainSubstring("1970-01-01T01:00:00.00+0100"))
			})
		})
		Context("sourceID sets the logHeader format", func() {
			It("omits sourceID from the header when sourceID is empty", func() {
				message = *sharedaction.NewLogMessage(
					"some-message",
					"OUT",
					time.Unix(0, 0),
					"APP/PROC/WEB",
					"",
				)
				logCacheMessage = logs.NewLogCacheMessage(message)

				Expect(logCacheMessage.ToLog(time.UTC)).To(ContainSubstring("[APP/PROC/WEB]"))

			})
			It("includes sourceID to the header when sourceID is not empty", func() {
				Expect(logCacheMessage.ToLog(time.UTC)).To(ContainSubstring("[APP/PROC/WEB/0]"))
			})
		})
		// coloring of log header on logType value
		Context("trimming of message text", func() {
			It("trims newlines from the end of the message", func() {
				message = *sharedaction.NewLogMessage(
					"some-message\r\n",
					"OUT",
					time.Unix(0, 0),
					"APP/PROC/WEB",
					"",
				)
				logCacheMessage = logs.NewLogCacheMessage(message)
				Expect(logCacheMessage.ToLog(time.UTC)).To(HaveSuffix("some-message"), "invisible characters present")
			})
		})
		Context("padding calculation", func() {
			It("prepending output with padding", func() {
				Expect(logCacheMessage.ToLog(time.UTC)).To(HavePrefix("    1970-01-01"))
			})
		})
		// multi-line log messages
		// handles both STDERR and STDOUT differently

		It("correctly formats the log", func() {
			Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [APP/PROC/WEB/0] OUT some-message"))
		})
	})
})
