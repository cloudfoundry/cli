package consumer_test

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConsumer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consumer Suite")
}

func createMessage(message string, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}

	logMessage := createLogMessage(message, timestamp)

	return &events.Envelope{
		LogMessage: logMessage,
		EventType:  events.Envelope_LogMessage.Enum(),
		Origin:     proto.String("fake-origin-1"),
		Timestamp:  proto.Int64(timestamp),
	}
}

func createContainerMetric(instanceIndex int32, timestamp int64) *events.Envelope {
	if timestamp == 0 {
		timestamp = time.Now().UnixNano()
	}

	cm := &events.ContainerMetric{
		ApplicationId: proto.String("appId"),
		InstanceIndex: proto.Int32(instanceIndex),
		CpuPercentage: proto.Float64(1),
		MemoryBytes:   proto.Uint64(2),
		DiskBytes:     proto.Uint64(3),
	}

	return &events.Envelope{
		ContainerMetric: cm,
		EventType:       events.Envelope_ContainerMetric.Enum(),
		Origin:          proto.String("fake-origin-1"),
		Timestamp:       proto.Int64(timestamp),
	}
}

func createLogMessage(message string, timestamp int64) *events.LogMessage {
	return &events.LogMessage{
		Message:     []byte(message),
		MessageType: events.LogMessage_OUT.Enum(),
		AppId:       proto.String("my-app-guid"),
		SourceType:  proto.String("DEA"),
		Timestamp:   proto.Int64(timestamp),
	}
}

func marshalMessage(message *events.Envelope) []byte {
	data, err := proto.Marshal(message)
	if err != nil {
		println(err.Error())
	}

	return data
}
