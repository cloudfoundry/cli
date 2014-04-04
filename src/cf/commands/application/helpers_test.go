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
	date := time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)

	Context("when the message comes from an app", func() {
	    It("includes the instance index", func() {
			msg := createMessage("4", "App", logmessage.LogMessage_OUT, date)
			Expect(LogMessageOutput(msg, time.UTC)).To(Equal("2014-04-04T11:39:20.00+0000 [App/4]   OUT Hello World!"))
		})
	})

	Context("when the message comes from a cloudfoundry component", func() {
	    It("doesn't include the instance index", func() {
			msg := createMessage("4", "DEA", logmessage.LogMessage_OUT, date)
			Expect(LogMessageOutput(msg, time.UTC)).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]     OUT Hello World!"))
	    })
	})

	Context("when the message was written to stderr", func() {
	    It("shows the log type as 'ERR'", func() {
			msg := createMessage("4", "DEA", logmessage.LogMessage_ERR, date)
			Expect(LogMessageOutput(msg, time.UTC)).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]     ERR Hello World!"))
		})
	})

	It("formats the time in the given time zone", func() {
		msg := createMessage("4", "DEA", logmessage.LogMessage_ERR, date)
		Expect(LogMessageOutput(msg, time.FixedZone("the-zone", 3 * 60 * 60))).To(Equal("2014-04-04T14:39:20.00+0300 [DEA]     ERR Hello World!"))
	})
})
