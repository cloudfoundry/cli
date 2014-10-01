package configuration_test

import (
	"os"

	. "github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiskPersistor", func() {
	var (
		tmpDir        string
		diskPersistor DiskPersistor
	)

	BeforeEach(func() {
		tmpDir = os.TempDir()
		diskPersistor = NewDiskPersistor(tmpDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe(".Delete", func() {
		It("Deletes the correct file/filepath", func() {
			diskPersistor.Delete()

			dir, err := os.Stat(tmpDir)
			Expect(dir).To(BeNil())
			Expect(err).To(HaveOccured())
		})
	})

	Describe(".Save", func() {
		It("Writes the json file to the correct filepath", func() {
		})
	})

	Describe(".Load", func() {
		It("Will load an empty json file", func() {
		})

		It("Will load a json file with specific keys", func() {

		})
	})

})
