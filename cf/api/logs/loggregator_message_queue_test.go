package logs_test

import (
	. "code.cloudfoundry.org/cli/cf/api/logs"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoggregatorMessageQueue", func() {
	It("sorts messages based on their timestamp, clearing after it's enumerated over", func() {
		pq := NewLoggregatorMessageQueue()

		msg3 := logLoggregatorMessageWithTime("message 3", 130)
		msg2 := logLoggregatorMessageWithTime("message 2", 120)
		msg4 := logLoggregatorMessageWithTime("message 4", 140)
		msg1 := logLoggregatorMessageWithTime("message 1", 110)

		pq.PushMessage(msg3)
		pq.PushMessage(msg2)
		pq.PushMessage(msg4)
		pq.PushMessage(msg1)

		var messages []*logmessage.LogMessage

		pq.EnumerateAndClear(func(m *logmessage.LogMessage) {
			messages = append(messages, m)
		})

		Expect(messages).To(Equal([]*logmessage.LogMessage{
			msg1,
			msg2,
			msg3,
			msg4,
		}))

		var messagesAfter []*logmessage.LogMessage

		pq.EnumerateAndClear(func(m *logmessage.LogMessage) {
			messagesAfter = append(messagesAfter, m)
		})

		Expect(messagesAfter).To(BeEmpty())
	})
})

func logLoggregatorMessageWithTime(messageString string, timestamp int) *logmessage.LogMessage {
	return generateLoggregatorMessage(messageString, int64(timestamp))
}

func generateLoggregatorMessage(messageString string, timestamp int64) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_OUT
	sourceName := "DEA"
	return &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceName:  &sourceName,
		Timestamp:   proto.Int64(timestamp),
	}
}
