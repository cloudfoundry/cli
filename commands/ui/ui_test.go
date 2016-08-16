package ui_test

import (
	. "code.cloudfoundry.org/cli/commands/ui"
	"code.cloudfoundry.org/cli/commands/ui/uifakes"
	"code.cloudfoundry.org/cli/utils/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("UI", func() {
	var (
		ui         UI
		fakeConfig *uifakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(config.ColorEnabled)
		ui = NewUI(fakeConfig)
		ui.Out = NewBuffer()
	})

	Describe("DisplayText", func() {
		Context("when only a string is passed in", func() {
			It("displays the string to Out with a newline", func() {
				ui.DisplayText("some-string")

				Expect(ui.Out).To(Say("some-string\n"))
			})
		})

		Context("when a map is passed in", func() {
			It("merges the map content with the string", func() {
				ui.DisplayText("some-string {{.SomeMapValue}}", map[string]interface{}{
					"SomeMapValue": "my-map-value",
				})

				Expect(ui.Out).To(Say("some-string my-map-value\n"))
			})
		})

		Context("when multiple maps are passed in", func() {
			It("merges all map value with the string", func() {
				ui.DisplayText("some-string {{.SomeMapValue}} {{.SomeOtherMapValue}}",
					map[string]interface{}{
						"SomeMapValue": "my-map-value",
					},
					map[string]interface{}{
						"SomeOtherMapValue": "my-other-map-value",
					},
				)

				Expect(ui.Out).To(Say("some-string my-map-value my-other-map-value\n"))
			})

			Context("when maps share the same key", func() {
				It("keeps the rightmost map value", func() {
					ui.DisplayText("some-string {{.SomeMapValue}}",
						map[string]interface{}{
							"SomeMapValue": "my-map-value",
						},
						map[string]interface{}{
							"SomeMapValue": "my-other-map-value",
						},
					)

					Expect(ui.Out).To(Say("some-string my-other-map-value\n"))
				})
			})
		})
	})

	Describe("DisplayNewline", func() {
		It("displays a new line", func() {
			ui.DisplayNewline()

			Expect(ui.Out).To(Say("\n"))
		})
	})

	Describe("DisplayHelpHeader", func() {
		It("bolds and colorizes the input string", func() {
			ui.DisplayHelpHeader("some-text")
			Expect(ui.Out).To(Say("\x1b\\[38;1msome-text\x1b\\[0m"))
		})
	})
})
