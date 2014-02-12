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

func createMessage(protoMsg *logmessage.LogMessage, sourceName *string, msgType *logmessage.LogMessage_MessageType) (msg *logmessage.Message) {
	protoMsg.SourceName = sourceName
	protoMsg.MessageType = msgType

	data, err := proto.Marshal(protoMsg)
	Expect(err).NotTo(HaveOccurred())

	msg, err = logmessage.ParseMessage(data)
	Expect(err).NotTo(HaveOccurred())

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
		timestamp := date.UnixNano()

		sourceId := "0"

		protoMessage := &logmessage.LogMessage{
			Message:     []byte("Hello World!\n\r\n\r"),
			AppId:       proto.String("my-app-guid"),
			MessageType: &stdout,
			SourceId:    &sourceId,
			Timestamp:   &timestamp,
		}

		msg := createMessage(protoMessage, &cloud_controller, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))

		msg = createMessage(protoMessage, &cloud_controller, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		sourceId = "1"
		msg = createMessage(protoMessage, &router, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage(protoMessage, &router, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		sourceId = "2"
		msg = createMessage(protoMessage, &uaa, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage(protoMessage, &uaa, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		sourceId = "3"
		msg = createMessage(protoMessage, &dea, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage(protoMessage, &dea, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))

		sourceId = "4"
		msg = createMessage(protoMessage, &wardenContainer, &stdout)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStdoutColor("OUT Hello World!")))
		msg = createMessage(protoMessage, &wardenContainer, &stderr)
		Expect(LogMessageOutput(msg)).To(ContainSubstring(fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT))))
		Expect(LogMessageOutput(msg)).To(ContainSubstring(terminal.LogStderrColor("ERR Hello World!")))
	})
})
