package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MergeAndValidateSettingsAndManifest", func() {
	var (
		settings CommandLineSettings
	)

	BeforeEach(func() {
		settings = CommandLineSettings{}
	})

	Describe("ApplicationPath", func() {
		Context("when ProvidedAppPath is set", func() {
			BeforeEach(func() {
				settings.ProvidedAppPath = "some-path"
			})

			It("returns the ProvidedAppPath", func() {
				Expect(settings.ApplicationPath()).To(Equal("some-path"))
			})
		})

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
