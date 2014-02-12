package api_test

import (
	. "cf/api"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"time"
)

func logMessageWithTime(t mr.TestingT, messageString string, timestamp int64) *logmessage.Message {
	data, err := proto.Marshal(generateMessage(messageString, timestamp))
	assert.NoError(t, err)
	message, err := logmessage.ParseMessage(data)
	assert.NoError(t, err)

	return message
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

func getMsgString(message *logmessage.Message) string {
	return string(message.GetLogMessage().GetMessage())
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestPriorityQueue", func() {
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)

		msg3 := logMessageWithTime(mr.T(), "message 3", int64(130))
		pq.PushMessage(msg3)
		msg2 := logMessageWithTime(mr.T(), "message 2", int64(120))
		pq.PushMessage(msg2)
		msg4 := logMessageWithTime(mr.T(), "message 4", int64(140))
		pq.PushMessage(msg4)
		msg1 := logMessageWithTime(mr.T(), "message 1", int64(110))
		pq.PushMessage(msg1)

		Expect(getMsgString(pq.PopMessage())).To(Equal(getMsgString(msg1)))
		assert.Equal(mr.T(), getMsgString(pq.PopMessage()), getMsgString(msg2))
		assert.Equal(mr.T(), getMsgString(pq.PopMessage()), getMsgString(msg3))
		assert.Equal(mr.T(), getMsgString(pq.PopMessage()), getMsgString(msg4))
	})
	It("TestPopOnEmptyQueue", func() {
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)

		var msg *logmessage.Message
		msg = nil
		assert.Equal(mr.T(), pq.PopMessage(), msg)
	})

	It("TestNextTimestamp", func() {
		currentTime := time.Unix(5, 0)
		clock := func() time.Time {
			return currentTime
		}

		pq := NewSortedMessageQueue(5*time.Second, clock)
		assert.Equal(mr.T(), pq.NextTimestamp(), MAX_INT64)

		msg2 := logMessageWithTime(mr.T(), "message 2", int64(130))
		pq.PushMessage(msg2)

		currentTime = time.Unix(6, 0)
		msg1 := logMessageWithTime(mr.T(), "message 1", int64(100))
		pq.PushMessage(msg1)
		Expect(pq.NextTimestamp()).To(Equal(time.Unix(11, 0).UnixNano()))

		readMessage := pq.PopMessage().GetLogMessage()
		Expect(readMessage.GetTimestamp()).To(Equal(int64(100)))
		Expect(pq.NextTimestamp()).To(Equal(time.Unix(10, 0).UnixNano()))

		readMessage = pq.PopMessage().GetLogMessage()
		Expect(readMessage.GetTimestamp()).To(Equal(int64(130)))
		assert.Equal(mr.T(), pq.NextTimestamp(), MAX_INT64)
	})

	It("TestStableSort", func() {
		pq := NewSortedMessageQueue(10*time.Millisecond, time.Now)

		msg1 := logMessageWithTime(mr.T(), "message first", int64(109))
		pq.PushMessage(msg1)

		for i := 1; i < 1000; i++ {
			msg := logMessageWithTime(mr.T(), fmt.Sprintf("message %d", i), int64(110))
			pq.PushMessage(msg)
		}
		msg2 := logMessageWithTime(mr.T(), "message last", int64(111))
		pq.PushMessage(msg2)

		assert.Equal(mr.T(), getMsgString(pq.PopMessage()), "message first")

		for i := 1; i < 1000; i++ {
			assert.Equal(mr.T(), getMsgString(pq.PopMessage()), fmt.Sprintf("message %d", i))
		}

		assert.Equal(mr.T(), getMsgString(pq.PopMessage()), "message last")
	})
})
