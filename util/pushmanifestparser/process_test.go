package pushmanifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process", func() {
	Describe("SetStartCommand", func() {
		var (
			process Process
			command string
		)

		BeforeEach(func() {
			process = Process{}
			command = "./start.sh"
		})

		JustBeforeEach(func() {
			process.SetStartCommand(command)
		})

		When("the remaining fields map exists", func() {
			BeforeEach(func() {
				process.RemainingManifestFields = map[string]interface{}{}
			})

			It("sets the start command in the map", func() {
				Expect(process.RemainingManifestFields["command"]).To(Equal("./start.sh"))
			})

			When("the command is nil", func() {
				BeforeEach(func() {
					command = ""
				})

				It("sets the start command to nil in the map", func() {
					Expect(process.RemainingManifestFields["command"]).To(BeNil())
				})
			})
		})

		When("the remaining fields map does not exist", func() {
			It("sets the start command in the map", func() {
				Expect(process.RemainingManifestFields["command"]).To(Equal("./start.sh"))
			})
		})
	})
})
