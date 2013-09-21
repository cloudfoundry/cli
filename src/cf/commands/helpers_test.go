package commands

import (
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

	timestamp := logTime.UnixNano()

	sourceId := "0"

	msg := logmessage.LogMessage{
		Message:     []byte("Hello World!"),
		MessageType: &stdout,
		SourceId:    &sourceId,
		Timestamp:   &timestamp,
	}

	msg.SourceType = &cloud_controller
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app API/0 Hello World!")
	msg.MessageType = &stderr
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app API/0 STDERR Hello World!")

	sourceId = "1"
	msg.SourceType = &router
	msg.MessageType = &stdout
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app Router/1 Hello World!")
	msg.MessageType = &stderr
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app Router/1 STDERR Hello World!")

	sourceId = "2"
	msg.SourceType = &uaa
	msg.MessageType = &stdout
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app UAA/2 Hello World!")
	msg.MessageType = &stderr
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app UAA/2 STDERR Hello World!")

	sourceId = "3"
	msg.SourceType = &dea
	msg.MessageType = &stdout
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app Executor/3 Hello World!")
	msg.MessageType = &stderr
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app Executor/3 STDERR Hello World!")

	sourceId = "4"
	msg.SourceType = &wardenContainer
	msg.MessageType = &stdout
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app App/4 Hello World!")
	msg.MessageType = &stderr
	assert.Contains(t, logMessageOutput("my-app", msg), "Sep 20 09:33:30 my-app App/4 STDERR Hello World!")
}
