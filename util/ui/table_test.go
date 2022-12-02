package ui_test

import (
	"strings"

	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Table", func() {
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

	Describe("DisplayKeyValueTable", func() {
		JustBeforeEach(func() {
			ui.DisplayKeyValueTable(" ",
				[][]string{
					{"wut0:", ""},
					{"wut1:", "hi hi"},
					nil,
					[]string{},
					{"wut2:", strings.Repeat("a", 9)},
					{"wut3:", "hi hi " + strings.Repeat("a", 9)},
					{"wut4:", strings.Repeat("a", 15) + " " + strings.Repeat("b", 15)},
				},
				2)
		})

		Context("in a TTY", func() {
			BeforeEach(func() {
				ui.IsTTY = true
				ui.TerminalWidth = 20
			})

			It("displays a table with the last column wrapping according to width", func() {
				Expect(out).To(Say(" wut0:  \n"))
				Expect(out).To(Say(" wut1:  hi hi\n"))
				Expect(out).To(Say(" wut2:  %s\n", strings.Repeat("a", 9)))
				Expect(out).To(Say(" wut3:  hi hi\n"))
				Expect(out).To(Say("        %s\n", strings.Repeat("a", 9)))
				Expect(out).To(Say(" wut4:  %s\n", strings.Repeat("a", 15)))
				Expect(out).To(Say("        %s\n", strings.Repeat("b", 15)))
			})
		})
	})

	Describe("DisplayTableWithHeader", func() {
		It("makes the first row bold", func() {
			ui.DisplayTableWithHeader(" ",
				[][]string{
					{"", "header1", "header2", "header3"},
					{"#0", "data1", "data2", "data3"},
				},
				2)
			Expect(out).To(Say("    \x1b\\[1mheader1\x1b\\[0m")) // Makes sure empty values are not bolded
			Expect(out).To(Say("\x1b\\[1mheader2\x1b\\[0m"))
			Expect(out).To(Say("\x1b\\[1mheader3\x1b\\[0m"))
			Expect(out).To(Say("#0  data1    data2    data3"))
		})
	})
})
