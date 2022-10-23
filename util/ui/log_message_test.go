package ui_test

import (
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Log Message", func() {
	var (
		ui         *UI
		fakeConfig *uifakes.FakeConfig
		out        *Buffer
		errBuff    *Buffer
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		out = NewBuffer()
		ui.Out = out
		ui.OutForInteraction = out
		errBuff = NewBuffer()
		ui.Err = errBuff
	})

	It("sets the TimezoneLocation to the local timezone", func() {
		location := time.Now().Location()
		Expect(ui.TimezoneLocation).To(Equal(location))
	})

	Describe("DisplayLogMessage", func() {
		var message *uifakes.FakeLogMessage

		BeforeEach(func() {
			var err error
			ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
			Expect(err).NotTo(HaveOccurred())

			message = new(uifakes.FakeLogMessage)
			message.MessageReturns("This is a log message\r\n")
			message.TypeReturns("OUT")
			message.TimestampReturns(time.Unix(1468969692, 0)) // "2016-07-19T16:08:12-07:00"
			message.SourceTypeReturns("APP/PROC/WEB")
			message.SourceInstanceReturns("12")
		})

		Context("with header", func() {
			Context("single line log message", func() {
				It("prints out a single line to STDOUT", func() {
					ui.DisplayLogMessage(message, true)
					Expect(out).To(Say(`2016-07-19T16:08:12.00-0700 \[APP/PROC/WEB/12\] OUT This is a log message\n`))
				})
			})

			Context("multi-line log message", func() {
				BeforeEach(func() {
					var err error
					ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
					Expect(err).NotTo(HaveOccurred())

					message.MessageReturns("This is a log message\nThis is also a log message")
				})

				It("prints out mutliple lines to STDOUT", func() {
					ui.DisplayLogMessage(message, true)
					Expect(out).To(Say(`2016-07-19T16:08:12.00-0700 \[APP/PROC/WEB/12\] OUT This is a log message\n`))
					Expect(out).To(Say(`2016-07-19T16:08:12.00-0700 \[APP/PROC/WEB/12\] OUT This is also a log message\n`))
				})
			})
		})

		Context("without header", func() {
			Context("single line log message", func() {
				It("prints out a single line to STDOUT", func() {
					ui.DisplayLogMessage(message, false)
					Expect(out).To(Say("This is a log message\n"))
				})
			})

			Context("multi-line log message", func() {
				BeforeEach(func() {
					var err error
					ui.TimezoneLocation, err = time.LoadLocation("America/Los_Angeles")
					Expect(err).NotTo(HaveOccurred())

					message.MessageReturns("This is a log message\nThis is also a log message")
				})

				It("prints out mutliple lines to STDOUT", func() {
					ui.DisplayLogMessage(message, false)
					Expect(out).To(Say("This is a log message\n"))
					Expect(out).To(Say("This is also a log message\n"))
				})
			})
		})

		Context("error log lines", func() {
			BeforeEach(func() {
				message.TypeReturns("ERR")
			})
			It("colors the line red", func() {
				ui.DisplayLogMessage(message, false)
				Expect(out).To(Say("\x1b\\[31mThis is a log message\x1b\\[0m\n"))
			})
		})
	})
})
