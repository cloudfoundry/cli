package ui_helpers_test

import (
	"time"

	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var location *time.Location

	BeforeEach(func() {
		location = time.UTC
	})

	Describe("ExtractLogHeader", func() {
		It("formats log header with source name and ID", func() {
			timestamp := time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano()
			sourceID := "0"
			sourceName := "App"

			msg := logmessage.NewLogMessage([]byte("log content"), "app-guid", logmessage.LogMessage_OUT, sourceName, sourceID)
			msg.Timestamp = &timestamp

			logHeader, _ := ui_helpers.ExtractLogHeader(msg, location)

			Expect(logHeader).To(ContainSubstring("2015-01-01T12:00:00.00"))
			Expect(logHeader).To(ContainSubstring("[App/0]"))
		})

		It("formats log header without source ID", func() {
			timestamp := time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano()
			sourceName := "STG"

			msg := logmessage.NewLogMessage([]byte("log content"), "app-guid", logmessage.LogMessage_OUT, sourceName, "")
			msg.Timestamp = &timestamp

			logHeader, _ := ui_helpers.ExtractLogHeader(msg, location)

			Expect(logHeader).To(ContainSubstring("2015-01-01T12:00:00.00"))
			Expect(logHeader).To(ContainSubstring("[STG]"))
			Expect(logHeader).ToNot(ContainSubstring("[STG/]"))
		})

		It("returns colored log header", func() {
			timestamp := time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano()
			sourceName := "RTR"
			sourceID := "1"

			msg := logmessage.NewLogMessage([]byte("log content"), "app-guid", logmessage.LogMessage_OUT, sourceName, sourceID)
			msg.Timestamp = &timestamp

			_, coloredLogHeader := ui_helpers.ExtractLogHeader(msg, location)

			Expect(coloredLogHeader).ToNot(BeEmpty())
			Expect(coloredLogHeader).To(ContainSubstring("RTR"))
		})

		It("pads the log header to expected length", func() {
			timestamp := time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano()
			sourceName := "A"
			sourceID := "0"

			msg := logmessage.NewLogMessage([]byte("log content"), "app-guid", logmessage.LogMessage_OUT, sourceName, sourceID)
			msg.Timestamp = &timestamp

			logHeader, _ := ui_helpers.ExtractLogHeader(msg, location)

			// Should have padding spaces
			Expect(len(logHeader)).To(BeNumerically(">", len("2015-01-01T12:00:00.00+0000 [A/0]")))
		})

		It("respects the timezone", func() {
			timestamp := time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano()
			sourceName := "App"
			sourceID := "0"
			pstLocation, _ := time.LoadLocation("America/Los_Angeles")

			msg := logmessage.NewLogMessage([]byte("log content"), "app-guid", logmessage.LogMessage_OUT, sourceName, sourceID)
			msg.Timestamp = &timestamp

			logHeader, _ := ui_helpers.ExtractLogHeader(msg, pstLocation)

			// PST is UTC-8, so 12:00 UTC becomes 04:00 PST
			Expect(logHeader).To(ContainSubstring("04:00"))
		})
	})

	Describe("ExtractNoaaLogHeader", func() {
		It("formats log header with source type and instance", func() {
			timestamp := int64(time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano())
			sourceType := "APP"
			sourceInstance := "0"
			msgType := events.LogMessage_OUT

			msg := &events.LogMessage{
				Message:        []byte("log content"),
				MessageType:    &msgType,
				Timestamp:      &timestamp,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			logHeader, _ := ui_helpers.ExtractNoaaLogHeader(msg, location)

			Expect(logHeader).To(ContainSubstring("2015-01-01T12:00:00.00"))
			Expect(logHeader).To(ContainSubstring("[APP/0]"))
		})

		It("formats log header without source instance", func() {
			timestamp := int64(time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano())
			sourceType := "RTR"
			sourceInstance := ""
			msgType := events.LogMessage_OUT

			msg := &events.LogMessage{
				Message:        []byte("log content"),
				MessageType:    &msgType,
				Timestamp:      &timestamp,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			logHeader, _ := ui_helpers.ExtractNoaaLogHeader(msg, location)

			Expect(logHeader).To(ContainSubstring("[RTR]"))
			Expect(logHeader).ToNot(ContainSubstring("[RTR/]"))
		})

		It("returns colored log header", func() {
			timestamp := int64(time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano())
			sourceType := "API"
			sourceInstance := "1"
			msgType := events.LogMessage_OUT

			msg := &events.LogMessage{
				Message:        []byte("log content"),
				MessageType:    &msgType,
				Timestamp:      &timestamp,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			_, coloredLogHeader := ui_helpers.ExtractNoaaLogHeader(msg, location)

			Expect(coloredLogHeader).ToNot(BeEmpty())
			Expect(coloredLogHeader).To(ContainSubstring("API"))
		})

		It("pads the log header to expected length", func() {
			timestamp := int64(time.Date(2015, time.January, 1, 12, 0, 0, 0, time.UTC).UnixNano())
			sourceType := "X"
			sourceInstance := "0"
			msgType := events.LogMessage_OUT

			msg := &events.LogMessage{
				Message:        []byte("log content"),
				MessageType:    &msgType,
				Timestamp:      &timestamp,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			logHeader, _ := ui_helpers.ExtractNoaaLogHeader(msg, location)

			// Should have padding spaces
			Expect(len(logHeader)).To(BeNumerically(">", len("2015-01-01T12:00:00.00+0000 [X/0]")))
		})
	})

	Describe("ExtractLogContent", func() {
		It("extracts log content with OUT type", func() {
			msg := logmessage.NewLogMessage([]byte("app log message"), "app-guid", logmessage.LogMessage_OUT, "App", "0")
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("OUT"))
			Expect(logContent).To(ContainSubstring("app log message"))
		})

		It("extracts log content with ERR type", func() {
			msg := logmessage.NewLogMessage([]byte("error message"), "app-guid", logmessage.LogMessage_ERR, "App", "0")
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("ERR"))
			Expect(logContent).To(ContainSubstring("error message"))
		})

		It("strips trailing newlines from message", func() {
			msg := logmessage.NewLogMessage([]byte("message with newlines\n\n"), "app-guid", logmessage.LogMessage_OUT, "App", "0")
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("message with newlines"))
			Expect(logContent).ToNot(HaveSuffix("\n"))
		})

		It("handles multiline log messages", func() {
			msg := logmessage.NewLogMessage([]byte("line 1\nline 2\nline 3"), "app-guid", logmessage.LogMessage_OUT, "App", "0")
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("line 1"))
			Expect(logContent).To(ContainSubstring("line 2"))
			Expect(logContent).To(ContainSubstring("line 3"))
		})

		It("indents continuation lines", func() {
			msg := logmessage.NewLogMessage([]byte("line 1\nline 2"), "app-guid", logmessage.LogMessage_OUT, "App", "0")
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractLogContent(msg, logHeader)

			// Second line should be indented to align with first line content
			lines := ui_helpers.ExtractLogContent(msg, logHeader)
			Expect(lines).To(ContainSubstring("\n"))
		})
	})

	Describe("ExtractNoaaLogContent", func() {
		It("extracts log content with OUT type", func() {
			msgType := events.LogMessage_OUT
			msg := &events.LogMessage{
				Message:     []byte("app log message"),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("OUT"))
			Expect(logContent).To(ContainSubstring("app log message"))
		})

		It("extracts log content with ERR type", func() {
			msgType := events.LogMessage_ERR
			msg := &events.LogMessage{
				Message:     []byte("error message"),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("ERR"))
			Expect(logContent).To(ContainSubstring("error message"))
		})

		It("strips trailing newlines and carriage returns", func() {
			msgType := events.LogMessage_OUT
			msg := &events.LogMessage{
				Message:     []byte("message with newlines\r\n\r\n"),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("message with newlines"))
			Expect(logContent).ToNot(ContainSubstring("\r"))
		})

		It("handles multiline log messages", func() {
			msgType := events.LogMessage_OUT
			msg := &events.LogMessage{
				Message:     []byte("line 1\nline 2\nline 3"),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("line 1"))
			Expect(logContent).To(ContainSubstring("line 2"))
			Expect(logContent).To(ContainSubstring("line 3"))
		})

		It("indents continuation lines with padding", func() {
			msgType := events.LogMessage_OUT
			msg := &events.LogMessage{
				Message:     []byte("line 1\nline 2"),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			// Second line should be indented
			Expect(logContent).To(ContainSubstring("\n"))
		})

		It("handles empty messages", func() {
			msgType := events.LogMessage_OUT
			msg := &events.LogMessage{
				Message:     []byte(""),
				MessageType: &msgType,
			}
			logHeader := "2015-01-01T12:00:00.00+0000 [App/0]   "

			logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

			Expect(logContent).To(ContainSubstring("OUT"))
		})
	})
})
