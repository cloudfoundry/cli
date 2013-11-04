package application

import (
	"cf/terminal"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimestampFormat(t *testing.T) {
	assert.Equal(t,TIMESTAMP_FORMAT,"2006-01-02T15:04:05.00-0700")
}

func TestLogMessageOutput(t *testing.T) {
	cloud_controller := logmessage.LogMessage_CLOUD_CONTROLLER
	router := logmessage.LogMessage_ROUTER
	uaa := logmessage.LogMessage_UAA
	dea := logmessage.LogMessage_DEA
	wardenContainer := logmessage.LogMessage_WARDEN_CONTAINER

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

	msg := createMessage(t, protoMessage, &cloud_controller, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))

	msg = createMessage(t, protoMessage, &cloud_controller, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [API]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "1"
	msg = createMessage(t, protoMessage, &router, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &router, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [RTR]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "2"
	msg = createMessage(t, protoMessage, &uaa, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &uaa, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [UAA]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "3"
	msg = createMessage(t, protoMessage, &dea, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &dea, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [DEA]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "4"
	msg = createMessage(t, protoMessage, &wardenContainer, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &wardenContainer, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("%s [App/4]", date.Format(TIMESTAMP_FORMAT)))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))
}

func createMessage(t *testing.T, protoMsg *logmessage.LogMessage, sourceType *logmessage.LogMessage_SourceType, msgType *logmessage.LogMessage_MessageType) (msg *logmessage.Message) {
	protoMsg.SourceType = sourceType
	protoMsg.MessageType = msgType

	data, err := proto.Marshal(protoMsg)
	assert.NoError(t, err)

	msg, err = logmessage.ParseMessage(data)
	assert.NoError(t, err)

	return
}
