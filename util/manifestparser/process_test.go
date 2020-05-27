package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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
	Describe("UnmarshalYAML", func() {
		var (
			yamlBytes []byte
			process   Process
			err       error
		)
		JustBeforeEach(func() {
			process = Process{}
			err = yaml.Unmarshal(yamlBytes, &process)
		})
		When("'disk_quota' (underscore, backwards-compatible) is specified", func() {
			BeforeEach(func() {
				yamlBytes = []byte("disk_quota: 5G")
			})
			It("unmarshals it properly", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(process).To(Equal(Process{DiskQuota: "5G", RemainingManifestFields: map[string]interface{}{}}))
			})
		})
		When("'disk-quota' (hyphen, new convention) is specified", func() {
			BeforeEach(func() {
				yamlBytes = []byte("disk-quota: 5G")
			})
			It("unmarshals it properly", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(process).To(Equal(Process{DiskQuota: "5G", RemainingManifestFields: map[string]interface{}{}}))
			})
		})

	})
})
