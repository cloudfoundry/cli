package api_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Testing with ginkgo", func() {
	It("PriorityQueue", func() {
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)

		msg3 := logMessageWithTime("message 3", 130)
		pq.PushMessage(msg3)
		msg2 := logMessageWithTime("message 2", 120)
		pq.PushMessage(msg2)
		msg4 := logMessageWithTime("message 4", 140)
		pq.PushMessage(msg4)
		msg1 := logMessageWithTime("message 1", 110)
		pq.PushMessage(msg1)

		Expect(getMsgString(pq.PopMessage())).To(Equal(getMsgString(msg1)))
		Expect(getMsgString(pq.PopMessage())).To(Equal(getMsgString(msg2)))
		Expect(getMsgString(pq.PopMessage())).To(Equal(getMsgString(msg3)))
		Expect(getMsgString(pq.PopMessage())).To(Equal(getMsgString(msg4)))
	})

	It("pops on empty queue", func() {
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)
		Expect(pq.PopMessage()).To(BeNil())
	})

	It("NextTimeStamp returns the timestamp of the log message at the head of the queue", func() {
		currentTime := time.Unix(5, 0)
		clock := func() time.Time {
			return currentTime
		}

		pq := NewSortedMessageQueue(5*time.Second, clock)
		Expect(pq.NextTimestamp()).To(Equal(MAX_INT64))

		msg2 := logMessageWithTime("message 2", 130)
		pq.PushMessage(msg2)

		currentTime = time.Unix(6, 0)
		msg1 := logMessageWithTime("message 1", 100)
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
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)

		msg1 := logMessageWithTime("message first", 109)
		pq.PushMessage(msg1)

		for i := 1; i < 1000; i++ {
			msg := logMessageWithTime(fmt.Sprintf("message %d", i), 110)
			pq.PushMessage(msg)
		}
		msg2 := logMessageWithTime("message last", 111)
		pq.PushMessage(msg2)

		Expect(getMsgString(pq.PopMessage())).To(Equal("message first"))

		for i := 1; i < 1000; i++ {
			Expect(getMsgString(pq.PopMessage())).To(Equal(fmt.Sprintf("message %d", i)))
		}

		Expect(getMsgString(pq.PopMessage())).To(Equal("message last"))
	})
})

func logMessageWithTime(messageString string, timestamp int) *logmessage.LogMessage {
	return generateMessage(messageString, int64(timestamp))
}

func generateMessage(messageString string, timestamp int64) *logmessage.LogMessage {
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

func getMsgString(message *logmessage.LogMessage) string {
	return string(message.GetMessage())
}
