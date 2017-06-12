package ui_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var (
		ui         *UI
		fakeConfig *uifakes.FakeConfig
		out        *Buffer
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		out = NewBuffer()
		ui.Out = out
		ui.Err = NewBuffer()
	})

	Describe("DisplayKeyValueTableForV3App", func() {
		Context("when the app is running properly", func() {
			BeforeEach(func() {
				ui.DisplayKeyValueTableForV3App([][]string{
					{"name:", "dora"},
					{"requested state:", "started"},
					{"processes:", "web:1/1,worker:2/2"}},
					[]string{},
				)
			})

			It("displays a table with the no change in coloring", func() {
				Expect(ui.Out).To(Say("name:              dora\n"))
				Expect(ui.Out).To(Say("requested state:   started\n"))
				Expect(ui.Out).To(Say("processes:         web:1/1,worker:2/2\n"))
			})
		})

		Context("when the app is stopped and has 0 instances", func() {
			BeforeEach(func() {
				ui.DisplayKeyValueTableForV3App([][]string{
					{"name:", "dora"},
					{"requested state:", "stopped"},
					{"processes:", "web:1/1,worker:2/2"}},
					[]string{})
			})

			It("displays a table with the no change in coloring", func() {
				Expect(ui.Out).To(Say("name:              dora\n"))
				Expect(ui.Out).To(Say("requested state:   stopped\n"))
				Expect(ui.Out).To(Say("processes:         web:1/1,worker:2/2\n"))
			})
		})

		Context("when the app is started and has 1 crashed process", func() {
			BeforeEach(func() {
				ui.DisplayKeyValueTableForV3App([][]string{
					{"name:", "dora"},
					{"requested state:", "started"},
					{"processes:", "web:0/1,worker:2/2"}},
					[]string{"web"})
			})

			It("displays a table with requested state and crashed instance count in red", func() {
				Expect(ui.Out).To(Say("name:              dora\n"))
				Expect(ui.Out).To(Say("requested state:   \x1b\\[31;1mstarted\x1b\\[0m\n"))
				Expect(ui.Out).To(Say("processes:         \x1b\\[31;1mweb:0/1\x1b\\[0m,worker:2/2\n"))
			})
		})
	})

	Describe("DisplayInstancesTableForApp", func() {
		Context("in english", func() {
			It("displays a table with red coloring for down and crashed", func() {
				ui.DisplayInstancesTableForApp([][]string{
					{"", "header1", "header2", "header3"},
					{"#0", "starting", "val1", "val2"},
					{"#1", "down", "val1", "val2"},
					{"#2", "crashed", "val1", "val2"},
				})

				Expect(ui.Out).To(Say("\x1b\\[1mheader1\x1b\\[0m\\s+\x1b\\[1mheader2\x1b\\[0m\\s+\x1b\\[1mheader3\x1b\\[0m")) // Makes sure empty values are not bolded
				Expect(ui.Out).To(Say("#0\\s+starting\\s+val1\\s+val2"))
				Expect(ui.Out).To(Say("#1\\s+\x1b\\[31;1mdown\x1b\\[0m\\s+val1\\s+val2"))
				Expect(ui.Out).To(Say("#2\\s+\x1b\\[31;1mcrashed\x1b\\[0m\\s+val1\\s+val2"))
			})
		})

		Context("in a non-english language", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				ui.Out = NewBuffer()
				ui.Err = NewBuffer()
			})

			It("displays a table with red coloring for down and crashed", func() {
				ui.DisplayInstancesTableForApp([][]string{
					{"", "header1", "header2", "header3"},
					{"#0", ui.TranslateText("starting"), "val1", "val2"},
					{"#1", ui.TranslateText("down"), "val1", "val2"},
					{"#2", ui.TranslateText("crashed"), "val1", "val2"},
				})

				Expect(ui.Out).To(Say("\x1b\\[1mheader1\x1b\\[0m\\s+\x1b\\[1mheader2\x1b\\[0m\\s+\x1b\\[1mheader3\x1b\\[0m")) // Makes sure empty values are not bolded
				Expect(ui.Out).To(Say("#0\\s+%s\\s+val1\\s+val2", ui.TranslateText("starting")))
				Expect(ui.Out).To(Say("#1\\s+\x1b\\[31;1m%s\x1b\\[0m\\s+val1\\s+val2", ui.TranslateText("down")))
				Expect(ui.Out).To(Say("#2\\s+\x1b\\[31;1m%s\x1b\\[0m\\s+val1\\s+val2", ui.TranslateText("crashed")))
			})
		})
	})

	Describe("DisplayKeyValueTableForApp", func() {
		Context("when the app is running properly", func() {
			BeforeEach(func() {
				ui.DisplayKeyValueTableForApp([][]string{
					{"name:", "dora"},
					{"requested state:", "started"},
					{"instances:", "1/1"},
				})
			})

			It("displays a table with the no change in coloring", func() {
				Expect(ui.Out).To(Say("name:              dora\n"))
				Expect(ui.Out).To(Say("requested state:   started\n"))
				Expect(ui.Out).To(Say("instances:         1/1\n"))
			})
		})

		Context("when the app is stopped and has 0 instances", func() {
			Context("in english", func() {
				BeforeEach(func() {
					ui.DisplayKeyValueTableForApp([][]string{
						{"name:", "dora"},
						{"requested state:", "stopped"},
						{"instances:", "0/1"},
					})
				})

				It("displays a table with the no change in coloring", func() {
					Expect(ui.Out).To(Say("name:              dora\n"))
					Expect(ui.Out).To(Say("requested state:   stopped\n"))
					Expect(ui.Out).To(Say("instances:         0/1\n"))
				})
			})

			Context("in a non-english language", func() {
				BeforeEach(func() {
					fakeConfig.LocaleReturns("fr-FR")

					var err error
					ui, err = NewUI(fakeConfig)
					Expect(err).NotTo(HaveOccurred())
					ui.Out = NewBuffer()
					ui.Err = NewBuffer()

					ui.DisplayKeyValueTableForApp([][]string{
						{"name:", "dora"},
						{"requested state:", ui.TranslateText("stopped")},
						{"instances:", "0/1"},
					})
				})

				It("displays a table with the no change in coloring", func() {
					Expect(ui.Out).To(Say("name:              dora\n"))
					Expect(ui.Out).To(Say("requested state:   %s\n", ui.TranslateText("stopped")))
					Expect(ui.Out).To(Say("instances:         0/1\n"))
				})
			})
		})

		Context("when the app is not stopped and has 0 instances", func() {
			Context("in english", func() {
				BeforeEach(func() {
					ui.DisplayKeyValueTableForApp([][]string{
						{"name:", "dora"},
						{"requested state:", "running"},
						{"instances:", "0/1"},
					})
				})

				It("displays a table with requested state and instances in red", func() {
					Expect(ui.Out).To(Say("name:              dora\n"))
					Expect(ui.Out).To(Say("requested state:   \x1b\\[31;1mrunning\x1b\\[0m\n"))
					Expect(ui.Out).To(Say("instances:         \x1b\\[31;1m0/1\x1b\\[0m\n"))
				})
			})

			Context("in a non-english language", func() {
				BeforeEach(func() {
					fakeConfig.LocaleReturns("fr-FR")

					var err error
					ui, err = NewUI(fakeConfig)
					Expect(err).NotTo(HaveOccurred())
					ui.Out = NewBuffer()
					ui.Err = NewBuffer()

					ui.DisplayKeyValueTableForApp([][]string{
						{"name:", "dora"},
						{"requested state:", ui.TranslateText("running")},
						{"instances:", "0/1"},
					})
				})

				It("displays a table with requested state and instances in red", func() {
					Expect(ui.Out).To(Say("name:              dora\n"))
					Expect(ui.Out).To(Say("requested state:   \x1b\\[31;1m%s\x1b\\[0m\n", ui.TranslateText("running")))
					Expect(ui.Out).To(Say("instances:         \x1b\\[31;1m0/1\x1b\\[0m\n"))
				})
			})
		})
	})
})
