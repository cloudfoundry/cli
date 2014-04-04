package application_test

import (
	. "cf/commands/application"
	"cf/terminal"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

func createMessage(sourceId string, sourceName *string, msgType *logmessage.LogMessage_MessageType) (msg *logmessage.LogMessage) {
	timestamp := time.Now().UnixNano()
	msg = &logmessage.LogMessage{
		Message:     []byte("Hello World!\n\r\n\r"),
		AppId:       proto.String("my-app-guid"),
		MessageType: msgType,
		SourceId:    &sourceId,
		Timestamp:   &timestamp,
		SourceName:  sourceName,
	}

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestTimestampFormat", func() {
		Expect(TIMESTAMP_FORMAT).To(Equal("2006-01-02T15:04:05.00-0700"))
	})

	It("TestLogMessageOutput", func() {
		cloud_controller := "API"
		router := "RTR"
		uaa := "UAA"
		dea := "DEA"
		wardenContainer := "App"

		stdout := logmessage.LogMessage_OUT
		stderr := logmessage.LogMessage_ERR

		date := time.Now()
		msg := createMessage("0", &cloud_controller, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))

		msg = createMessage("0", &cloud_controller, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		msg = createMessage("1", &router, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))

		msg = createMessage("1", &router, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		msg = createMessage("2", &uaa, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage("2", &uaa, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		msg = createMessage("3", &dea, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage("3", &dea, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		msg = createMessage("4", &wardenContainer, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage("4", &wardenContainer, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))
	})
})
