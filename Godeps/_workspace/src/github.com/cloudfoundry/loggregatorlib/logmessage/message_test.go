package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenerateMessageReturnsMessage(t *testing.T) {
	message, err := GenerateMessage(LogMessage_ERR, "myMessage", "123appID", "DEA")
	assert.NoError(t, err)

	assert.Equal(t, *message.GetLogMessage().AppId, "123appID")
	assert.Equal(t, *message.GetLogMessage().SourceName, "DEA")
	assert.Equal(t, *message.GetLogMessage().MessageType, LogMessage_ERR)
	assert.Equal(t, message.GetLogMessage().Message, []byte("myMessage"))
}

func TestEnvelopeExtractionForMessageFails(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessageWithSourceName(t, appMessageString, "App", "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)

	_, err := ParseEnvelope(marshalledMessage, "")
	assert.Error(t, err)
}

func TestExtractionFromEnvelopeWithValidSignature(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessageWithSourceName(t, appMessageString, "App", "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)
	marshalledEnvelope := MarshalledLogEnvelope(t, unmarshalledMessage, "some secret")

	message, err := ParseEnvelope(marshalledEnvelope, "some secret")
	assert.NoError(t, err)

	assert.Equal(t, 36, len(message.GetRawMessage()))
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
	assert.Equal(t, "App", message.logMessage.GetSourceName())
}

func TestExtractionFromEnvelopeWithInvalidSignature(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessageWithSourceName(t, appMessageString, "App", "myApp")
	marshalledEnvelope := MarshalledLogEnvelope(t, unmarshalledMessage, "some secret")

	_, err := ParseEnvelope(marshalledEnvelope, "invalid secret")
	assert.NotNil(t, err)

}

func TestExtractEnvelopeFromRawBytes(t *testing.T) {
	//This allows us to verify that the same extraction can be done on the Ruby side
	data := []uint8{10, 9, 109, 121, 95, 97, 112, 112, 95, 105, 100, 18, 64, 200, 50, 155, 229, 192, 81, 84, 207, 6, 73, 170, 77, 69, 0, 228, 210, 19, 158, 158, 196, 167, 164, 202, 189, 124, 54, 25, 26, 200, 250, 65, 64, 213, 183, 116, 76, 142, 82, 219, 61, 103, 39, 98, 171, 3, 123, 48, 162, 232, 216, 69, 38, 151, 75, 36, 40, 253, 162, 1, 9, 40, 219, 229, 55, 26, 43, 10, 12, 72, 101, 108, 108, 111, 32, 116, 104, 101, 114, 101, 33, 16, 1, 24, 224, 151, 169, 222, 161, 217, 246, 177, 38, 34, 9, 109, 121, 95, 97, 112, 112, 95, 105, 100, 40, 1, 50, 2, 52, 50}
	receivedEnvelope := &LogEnvelope{}
	err := proto.Unmarshal(data, receivedEnvelope)
	assert.NoError(t, err)
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetMessage(), []byte("Hello there!"))
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetAppId(), "my_app_id")
	assert.Equal(t, receivedEnvelope.GetRoutingKey(), "my_app_id")
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetSourceId(), "42")

	assert.True(t, receivedEnvelope.VerifySignature("secret"))
}

func TestThatSignatureValidatesWhenItMatches(t *testing.T) {
	secret := "super-secret"
	logMessage := NewLogMessageWithSourceName(t, "the logs", "App", "appid")

	envelope := &LogEnvelope{
		LogMessage: logMessage,
		RoutingKey: proto.String(*logMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	assert.True(t, envelope.VerifySignature(secret))
}

func TestThatSignatureDoesNotValidateWhenItDoesntMatch(t *testing.T) {
	envelope := &LogEnvelope{
		LogMessage: &LogMessage{},
		RoutingKey: proto.String("app_id"),
		Signature:  []byte{0, 1, 2}, //some bad signature
	}

	assert.False(t, envelope.VerifySignature("super-secret"))
}

func TestThatSignatureDoesNotValidateWhenSecretIsIncorrect(t *testing.T) {
	secret := "super-secret"
	logMessage := NewLogMessageWithSourceName(t, "the logs", "App", "appid")

	envelope := &LogEnvelope{
		LogMessage: logMessage,
		RoutingKey: proto.String(*logMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	assert.False(t, envelope.VerifySignature(secret+"not the right secret"))
}

func NewLogMessageWithSourceName(t *testing.T, messageString, sourceName, appId string) *LogMessage {
	currentTime := time.Now()

	messageType := LogMessage_OUT
	protoMessage := &LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String(appId),
		MessageType: &messageType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
		SourceName:  proto.String(sourceName),
	}
	return protoMessage
}

func MarshallLogMessage(t *testing.T, unmarshalledMessage *LogMessage) []byte {
	message, err := proto.Marshal(unmarshalledMessage)
	assert.NoError(t, err)

	return message
}

func MarshalledLogEnvelope(t *testing.T, unmarshalledMessage *LogMessage, secret string) []byte {
	envelope := &LogEnvelope{
		LogMessage: unmarshalledMessage,
		RoutingKey: proto.String(*unmarshalledMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	marshalledEnvelope, err := proto.Marshal(envelope)
	assert.NoError(t, err)

	return marshalledEnvelope
}

func UnmarshalLogEnvelope(t *testing.T, data []byte) *LogEnvelope {
	logEnvelope := new(LogEnvelope)
	err := proto.Unmarshal(data, logEnvelope)
	assert.NoError(t, err)
	return logEnvelope
}
