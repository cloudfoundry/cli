package application_test

import (
	. "cf/commands/application"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

func createMessage(sourceId string, sourceName string, msgType logmessage.LogMessage_MessageType, date time.Time) *logmessage.LogMessage {
	timestamp := date.UnixNano()
	return &logmessage.LogMessage{
		Message:     []byte("Hello World!\n\r\n\r"),
		AppId:       proto.String("my-app-guid"),
		MessageType: &msgType,
		SourceId:    &sourceId,
		Timestamp:   &timestamp,
		SourceName:  &sourceName,
	}
}

var _ = Describe("Helpers", func() {
	Context("when the message comes from an app", func() {
	    It("includes the instance index", func() {
			date := time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)
			msg := createMessage("4", "App", logmessage.LogMessage_OUT, date)
			Expect(LogMessageOutput(msg)).To(Equal("2014-04-04T04:39:20.00-0700 [App/4]   OUT Hello World!"))
		})
	})

	Context("when the message comes from a cloudfoundry component", func() {
	    It("doesn't include the instance index", func() {
			date := time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)
			msg := createMessage("4", "DEA", logmessage.LogMessage_OUT, date)
			Expect(LogMessageOutput(msg)).To(Equal("2014-04-04T04:39:20.00-0700 [DEA]     OUT Hello World!"))
	    })
	})

	Context("when the message was written to stderr", func() {
	    It("shows the log type as 'ERR'", func() {
			date := time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)
			msg := createMessage("4", "DEA", logmessage.LogMessage_ERR, date)
			Expect(LogMessageOutput(msg)).To(Equal("2014-04-04T04:39:20.00-0700 [DEA]     ERR Hello World!"))
		})
	})
})
