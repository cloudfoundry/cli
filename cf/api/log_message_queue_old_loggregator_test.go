package api_test

import (
	"fmt"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("is a priority queue used to sort loggregator messages", func() {
	It("PriorityQueue returns a new queue", func() {
		pq := NewLoggregator_SortedMessageQueue(10*time.Millisecond, time.Now)

		msg3 := logLoggregatorMessageWithTime("message 3", 130)
		pq.PushMessage(msg3)
		msg2 := logLoggregatorMessageWithTime("message 2", 120)
		pq.PushMessage(msg2)
		msg4 := logLoggregatorMessageWithTime("message 4", 140)
		pq.PushMessage(msg4)
		msg1 := logLoggregatorMessageWithTime("message 1", 110)
		pq.PushMessage(msg1)

		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal(getLoggregatorMsgString(msg1)))
		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal(getLoggregatorMsgString(msg2)))
		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal(getLoggregatorMsgString(msg3)))
		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal(getLoggregatorMsgString(msg4)))
	})

	It("pops on empty queue", func() {
		pq := NewLoggregator_SortedMessageQueue(10*time.Millisecond, time.Now)
		Expect(pq.PopMessage()).To(BeNil())
	})

	It("NextTimeStamp returns the timestamp of the log message at the head of the queue", func() {
		currentTime := time.Unix(5, 0)
		clock := func() time.Time {
			return currentTime
		}

		pq := NewLoggregator_SortedMessageQueue(5*time.Second, clock)
		Expect(pq.NextTimestamp()).To(Equal(MAX_INT64))

		msg2 := logLoggregatorMessageWithTime("message 2", 130)
		pq.PushMessage(msg2)

		currentTime = time.Unix(6, 0)
		msg1 := logLoggregatorMessageWithTime("message 1", 100)
		pq.PushMessage(msg1)
		Expect(pq.NextTimestamp()).To(Equal(time.Unix(11, 0).UnixNano()))

		readMessage := pq.PopMessage()
		Expect(readMessage.GetTimestamp()).To(Equal(int64(100)))
		Expect(pq.NextTimestamp()).To(Equal(time.Unix(10, 0).UnixNano()))

		readMessage = pq.PopMessage()
		Expect(readMessage.GetTimestamp()).To(Equal(int64(130)))
		Expect(pq.NextTimestamp()).To(Equal(MAX_INT64))
	})

	It("sorts messages based on their timestamp", func() {
		pq := NewLoggregator_SortedMessageQueue(10*time.Millisecond, time.Now)

		msg1 := logLoggregatorMessageWithTime("message first", 109)
		pq.PushMessage(msg1)

		for i := 1; i < 1000; i++ {
			msg := logLoggregatorMessageWithTime(fmt.Sprintf("message %d", i), 110)
			pq.PushMessage(msg)
		}
		msg2 := logLoggregatorMessageWithTime("message last", 111)
		pq.PushMessage(msg2)

		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal("message first"))

		for i := 1; i < 1000; i++ {
			Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal(fmt.Sprintf("message %d", i)))
		}

		Expect(getLoggregatorMsgString(pq.PopMessage())).To(Equal("message last"))
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

func getLoggregatorMsgString(message *logmessage.LogMessage) string {
	return string(message.GetMessage())
}
