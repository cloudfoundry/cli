package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLineSettings", func() {
	var (
		settings CommandLineSettings
	)

	BeforeEach(func() {
		settings = CommandLineSettings{}
	})

	Describe("ApplicationPath", func() {
		// more tests under command_line_settings_*OS*_test.go

		Context("when ProvidedAppPath is *not* set", func() {
			BeforeEach(func() {
				settings.CurrentDirectory = "current-dir"
			})

			It("returns the CurrentDirectory", func() {
				Expect(settings.ApplicationPath()).To(Equal("current-dir"))
			})
		})
	})
})
