package api

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	pq := newSortedMessageQueue(10 * time.Millisecond)

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
	pq := newSortedMessageQueue(10 * time.Millisecond)

	var msg *logmessage.Message
	msg = nil
	assert.Equal(t, pq.PopMessage(), msg)
}

func TestNextTimestamp(t *testing.T) {
	pq := newSortedMessageQueue(5 * time.Second)

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
	pq := newSortedMessageQueue(10 * time.Millisecond)

	msg1 := logMessageWithTime(t, "message first", int64(109))
	pq.PushMessage(msg1)

	for i := 1; i < 1000; i++ {
		msg := logMessageWithTime(t, fmt.Sprintf("message %s", i), int64(110))
		pq.PushMessage(msg)
	}
	msg2 := logMessageWithTime(t, "message last", int64(111))
	pq.PushMessage(msg2)

	assert.Equal(t, getMsgString(pq.PopMessage()), "message first")

	for i := 1; i < 1000; i++ {
		assert.Equal(t, getMsgString(pq.PopMessage()), fmt.Sprintf("message %s", i))
	}

	assert.Equal(t, getMsgString(pq.PopMessage()), "message last")
}

func BenchmarkPushMessages(b *testing.B) {
	r := rand.New(rand.NewSource(99))
	pq := newSortedMessageQueue(10 * time.Millisecond)
	for i := 0; i < b.N; i++ {
		msg := logMessageForBenchmark(b, fmt.Sprintf("message %s", i), r.Int63())
		pq.PushMessage(msg)
	}
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

func newSortedMessageQueue(printTimeBuffer time.Duration) *SortedMessageQueue {
	return &SortedMessageQueue{printTimeBuffer: printTimeBuffer}
}
