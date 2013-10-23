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

func TestLogMessageOutput(t *testing.T) {
	cloud_controller := logmessage.LogMessage_CLOUD_CONTROLLER
	router := logmessage.LogMessage_ROUTER
	uaa := logmessage.LogMessage_UAA
	dea := logmessage.LogMessage_DEA
	wardenContainer := logmessage.LogMessage_WARDEN_CONTAINER

	stdout := logmessage.LogMessage_OUT
	stderr := logmessage.LogMessage_ERR

	zone, _ := time.Now().Zone()
	date := fmt.Sprintf("2013 Sep 20 09:33:30 %s", zone)
	logTime, err := time.Parse("2006 Jan 2 15:04:05 MST", date)

	assert.NoError(t, err)
	expectedTZ := logTime.Format("-0700")

	timestamp := logTime.UnixNano()

	sourceId := "0"

	protoMessage := &logmessage.LogMessage{
		Message:     []byte("Hello World!\n\r\n\r"),
		AppId:       proto.String("my-app-guid"),
		MessageType: &stdout,
		SourceId:    &sourceId,
		Timestamp:   &timestamp,
	}

	msg := createMessage(t, protoMessage, &cloud_controller, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [API]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))

	msg = createMessage(t, protoMessage, &cloud_controller, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [API]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "1"
	msg = createMessage(t, protoMessage, &router, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [RTR]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &router, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [RTR]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "2"
	msg = createMessage(t, protoMessage, &uaa, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [UAA]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &uaa, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [UAA]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "3"
	msg = createMessage(t, protoMessage, &dea, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [DEA]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &dea, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [DEA]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStderrColor("ERR Hello World!"))

	sourceId = "4"
	msg = createMessage(t, protoMessage, &wardenContainer, &stdout)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [App/4]", expectedTZ))
	assert.Contains(t, logMessageOutput(msg), terminal.LogStdoutColor("OUT Hello World!"))
	msg = createMessage(t, protoMessage, &wardenContainer, &stderr)
	assert.Contains(t, logMessageOutput(msg), fmt.Sprintf("2013-09-20T09:33:30.00%s [App/4]", expectedTZ))
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
