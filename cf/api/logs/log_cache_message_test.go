package logs_test

import (
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/cf/api/logs"
	"code.cloudfoundry.org/cli/cf/api/logs/logsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("log messages", func() {

	var (
		fakeColorLogger *logsfakes.FakeColorLogger
		logCacheMessage *logs.LogCacheMessage
	)

	// TODO: review if that is the only way to initialize variables of these types

	message := *sharedaction.NewLogMessage(
		"some-message",
		"OUT",
		time.Unix(0, 0),
		"APP/PROC/WEB",
		"0",
	)

	BeforeEach(func() {
		fakeColorLogger = new(logsfakes.FakeColorLogger)
		message = *sharedaction.NewLogMessage(
			"some-message",
			"OUT",
			time.Unix(0, 0),
			"APP/PROC/WEB",
			"0",
		)
		logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)

		fakeColorLogger.LogStdoutColorCalls(func(data string) string {
			return data
		})

		fakeColorLogger.LogStderrColorCalls(func(data string) string {
			return data
		})

		fakeColorLogger.LogSysHeaderColorCalls(func(data string) string {
			return data
		})

	})

	Describe("ToSimpleLog", func() {
		It("returns the message", func() {
			Expect(logCacheMessage.ToSimpleLog()).To(Equal("some-message"))
		})
	})

	Describe("GetSourceName", func() {
		It("returns the source name", func() {
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
				logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)

				Expect(logCacheMessage.ToLog(time.UTC)).To(ContainSubstring("[APP/PROC/WEB]"))

			})
			It("includes sourceID to the header when sourceID is not empty", func() {
				Expect(logCacheMessage.ToLog(time.UTC)).To(ContainSubstring("[APP/PROC/WEB/0]"))
			})
		})
		Context("trimming of message text", func() {
			It("trims newlines from the end of the message", func() {
				message = *sharedaction.NewLogMessage(
					"some-message\r\n",
					"OUT",
					time.Unix(0, 0),
					"APP/PROC/WEB",
					"",
				)
				logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
				Expect(logCacheMessage.ToLog(time.UTC)).To(HaveSuffix("some-message"), "invisible characters present")
			})
		})
		Context("padding calculation", func() {
			Context("with a short source type empty source instance id", func() {
				It("prepending output with long padding", func() {
					message = *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"",
					)
					logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
					Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [STG]        OUT some-message"))
				})
			})

			Context("with a short source type and an instance id", func() {
				It("prepending output with medium padding", func() {
					message = *sharedaction.NewLogMessage(
						"some-message",
						"OUT",
						time.Unix(0, 0),
						"STG",
						"1",
					)
					logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
					Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [STG/1]      OUT some-message"))
				})
			})

			Context("with a long source type and a source instance id", func() {
				It("prepending output with no padding", func() {
					Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [APP/PROC/WEB/0]OUT some-message"))
				})
			})
		})

		Context("multi-line log messages", func() {
			It("splits the message into multiple lines", func() {
				message = *sharedaction.NewLogMessage(
					"some-message1\nsome-message2",
					"OUT",
					time.Unix(0, 0),
					"STG",
					"1",
				)
				logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
				Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [STG/1]      OUT some-message1\n" +
					"                                         some-message2"))
			})
		})

		Context("handles both STDERR and STDOUT", func() {
			It("correctly colors the STDOUT log message", func() {
				message = *sharedaction.NewLogMessage(
					"some-message",
					"OUT",
					time.Unix(0, 0),
					"STG",
					"1",
				)
				logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
				fakeColorLogger.LogSysHeaderColorReturns("colorized-header")
				fakeColorLogger.LogStdoutColorReturns("colorized-message")
				Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("colorized-header      colorized-message"))
				Expect(fakeColorLogger.LogSysHeaderColorCallCount()).To(Equal(1))
				Expect(fakeColorLogger.LogStdoutColorCallCount()).To(Equal(1))
				Expect(fakeColorLogger.LogStderrColorCallCount()).To(Equal(0))
			})
			It("correctly colors the STDERR log message", func() {
				message = *sharedaction.NewLogMessage(
					"some-message",
					"ERR",
					time.Unix(0, 0),
					"STG",
					"1",
				)
				logCacheMessage = logs.NewLogCacheMessage(fakeColorLogger, message)
				fakeColorLogger.LogSysHeaderColorReturns("colorized-header")
				fakeColorLogger.LogStderrColorReturns("colorized-message")
				Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("colorized-header      colorized-message"))
				Expect(fakeColorLogger.LogSysHeaderColorCallCount()).To(Equal(1))
				Expect(fakeColorLogger.LogStdoutColorCallCount()).To(Equal(0))
				Expect(fakeColorLogger.LogStderrColorCallCount()).To(Equal(1))
			})
		})

		It("correctly formats the log", func() {
			Expect(logCacheMessage.ToLog(time.UTC)).To(Equal("1970-01-01T00:00:00.00+0000 [APP/PROC/WEB/0]OUT some-message"))
		})
	})
})
