package logs_test

import (
	. "code.cloudfoundry.org/cli/cf/api/logs"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NoaaMessageQueue", func() {
	It("sorts messages based on their timestamp, clearing after it's enumerated over", func() {
		pq := NewNoaaMessageQueue()

		msg3 := noaaMessageWithTime("message 3", 130)
		msg2 := noaaMessageWithTime("message 2", 120)
		msg4 := noaaMessageWithTime("message 4", 140)
		msg1 := noaaMessageWithTime("message 1", 110)

		pq.PushMessage(msg3)
		pq.PushMessage(msg2)
		pq.PushMessage(msg4)
		pq.PushMessage(msg1)

		var messages []*events.LogMessage

		pq.EnumerateAndClear(func(m *events.LogMessage) {
			messages = append(messages, m)
		})

		Expect(messages).To(Equal([]*events.LogMessage{
			msg1,
			msg2,
			msg3,
			msg4,
		}))

		var messagesAfter []*events.LogMessage

		pq.EnumerateAndClear(func(m *events.LogMessage) {
			messagesAfter = append(messagesAfter, m)
		})

		Expect(messagesAfter).To(BeEmpty())
	})
})

func noaaMessageWithTime(messageString string, timestamp int) *events.LogMessage {
	return generateNoaaMessage(messageString, int64(timestamp))
}

func generateNoaaMessage(messageString string, timestamp int64) *events.LogMessage {
	messageType := events.LogMessage_OUT
	sourceType := "DEA"
	return &events.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(timestamp),
	}
}
