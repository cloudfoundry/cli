package api

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityMessageQueue(10 * time.Millisecond)

	msg3 := logMessageWithTime(t, "message 3", int64(130))
	pq.PushMessage(msg3)
	msg2 := logMessageWithTime(t, "message 2", int64(120))
	pq.PushMessage(msg2)
	msg4 := logMessageWithTime(t, "message 4", int64(140))
	pq.PushMessage(msg4)
	msg1 := logMessageWithTime(t, "message 1", int64(110))
	pq.PushMessage(msg1)

	assert.Equal(t, getMsgString(pq.PopMessage()), getMsgString(msg1))
	assert.Equal(t, getMsgString(pq.PopMessage()), getMsgString(msg2))
	assert.Equal(t, getMsgString(pq.PopMessage()), getMsgString(msg3))
	assert.Equal(t, getMsgString(pq.PopMessage()), getMsgString(msg4))
}

func TestPopOnEmptyQueue(t *testing.T) {
	pq := NewPriorityMessageQueue(10 * time.Millisecond)

	var msg *logmessage.Message
	msg = nil
	assert.Equal(t, pq.PopMessage(), msg)
}

func TestNextTimestamp(t *testing.T) {
	pq := NewPriorityMessageQueue(5 * time.Second)

	assert.Equal(t, pq.NextTimestamp(), MAX_INT64)

	msg2 := logMessageWithTime(t, "message 2", int64(130))
	pq.PushMessage(msg2)
	timeNowWhenInsertingMessage1 := time.Now()

	time.Sleep(50 * time.Millisecond)

	msg1 := logMessageWithTime(t, "message 1", int64(100))
	pq.PushMessage(msg1)

	allowedDelta := (20 * time.Microsecond).Nanoseconds()

	timeWhenOutputtable := time.Now().Add(5 * time.Second).UnixNano()
	assert.True(t, pq.NextTimestamp()-timeWhenOutputtable < allowedDelta)
	assert.True(t, pq.NextTimestamp()-timeWhenOutputtable > -allowedDelta)

	pq.PopMessage()

	timeWhenOutputtable = timeNowWhenInsertingMessage1.Add(5 * time.Second).UnixNano()
	assert.True(t, pq.NextTimestamp()-timeWhenOutputtable < allowedDelta)
	assert.True(t, pq.NextTimestamp()-timeWhenOutputtable > -allowedDelta)
}

func TestStableSort(t *testing.T) {
	pq := NewPriorityMessageQueue(10 * time.Millisecond)

	msg1 := logMessageWithTime(t, "message 1", int64(110))
	pq.PushMessage(msg1)
	msg4 := logMessageWithTime(t, "message 4", int64(130))
	pq.PushMessage(msg4)
	msg5 := logMessageWithTime(t, "message 5", int64(150))
	pq.PushMessage(msg5)
	msg2 := logMessageWithTime(t, "message 2", int64(110))
	pq.PushMessage(msg2)
	msg3 := logMessageWithTime(t, "message 3", int64(110))
	pq.PushMessage(msg3)

	assert.Equal(t, pq.PopMessage(), msg1)
	assert.Equal(t, pq.PopMessage(), msg2)
	assert.Equal(t, pq.PopMessage(), msg3)
	assert.Equal(t, pq.PopMessage(), msg4)
	assert.Equal(t, pq.PopMessage(), msg5)
}

func logMessageWithTime(t *testing.T, messageString string, timestamp int64) *logmessage.Message {
	data, err := proto.Marshal(generateMessage(messageString, timestamp))
	assert.NoError(t, err)
	message, err := logmessage.ParseMessage(data)
	assert.NoError(t, err)

	return message
}

func logMessageForBenchmark(b *testing.B, messageString string, timestamp int64) *logmessage.Message {
	data, _ := proto.Marshal(generateMessage(messageString, timestamp))
	message, _ := logmessage.ParseMessage(data)
	return message
}

func generateMessage(messageString string, timestamp int64) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	return &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String("my-app-guid"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(timestamp),
	}
}

func getMsgString(message *logmessage.Message) string {
	return string(message.GetLogMessage().GetMessage())
}
