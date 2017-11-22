package noaa_test

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNoaa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Noaa Suite")
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
