package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"

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

	Describe("OverrideManifestSettings", func() {
		// more tests under command_line_settings_*OS*_test.go

		var input, output manifest.Application

		BeforeEach(func() {
			input.Name = "steve"
		})

		JustBeforeEach(func() {
			output = settings.OverrideManifestSettings(input)
		})

		Describe("name", func() {
			Context("when the command line settings provides a name", func() {
				BeforeEach(func() {
					settings.Name = "not-steve"
				})

				It("overrides the name", func() {
					Expect(output.Name).To(Equal("not-steve"))
				})
			})

			Context("when the command line settings name is blank", func() {
				It("passes the manifest name through", func() {
					Expect(output.Name).To(Equal("steve"))
				})
			})
		})
	})
})
