package logs_test

import (
	"time"

	testlogs "code.cloudfoundry.org/cli/util/testhelpers/logs"

	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("loggregatorMessage", func() {
	Describe("ToLog", func() {
		var date time.Time

		BeforeEach(func() {
			date = time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)
		})

		Context("when the message comes", func() {
			It("include the instance index", func() {
				msg := testlogs.NewLogMessage("Hello World!", "", "DEA", "4", logmessage.LogMessage_OUT, date)
				Expect(terminal.Decolorize(msg.ToLog(time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA/4]      OUT Hello World!"))
			})

			It("doesn't include the instance index if sourceID is empty", func() {
				msg := testlogs.NewLogMessage("Hello World!", "", "DEA", "", logmessage.LogMessage_OUT, date)
				Expect(terminal.Decolorize(msg.ToLog(time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]        OUT Hello World!"))
			})
		})

		Context("when the message was written to stderr", func() {
			It("shows the log type as 'ERR'", func() {
				msg := testlogs.NewLogMessage("Hello World!", "", "DEA", "4", logmessage.LogMessage_ERR, date)
				Expect(terminal.Decolorize(msg.ToLog(time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA/4]      ERR Hello World!"))
			})
		})

		It("formats the time in the given time zone", func() {
			msg := testlogs.NewLogMessage("Hello World!", "", "DEA", "4", logmessage.LogMessage_ERR, date)
			Expect(terminal.Decolorize(msg.ToLog(time.FixedZone("the-zone", 3*60*60)))).To(Equal("2014-04-04T14:39:20.00+0300 [DEA/4]      ERR Hello World!"))
		})
	})
})
