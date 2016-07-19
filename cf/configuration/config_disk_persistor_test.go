package configuration_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiskPersistor", func() {
	var (
		tmpDir        string
		tmpFile       *os.File
		diskPersistor DiskPersistor
	)

	BeforeEach(func() {
		var err error

		tmpDir = os.TempDir()

		tmpFile, err = ioutil.TempFile(tmpDir, "tmp_file")
		Expect(err).ToNot(HaveOccurred())

		diskPersistor = NewDiskPersistor(tmpFile.Name())
	})

	AfterEach(func() {
		os.Remove(tmpFile.Name())
	})

	Describe(".Delete", func() {
		It("Deletes the correct file", func() {
			tmpFile.Close()
			diskPersistor.Delete()

			file, err := os.Stat(tmpFile.Name())
			Expect(file).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Save", func() {
		It("Writes the json file to the correct filepath", func() {
			d := &data{Info: "save test"}

			err := diskPersistor.Save(d)
			Expect(err).ToNot(HaveOccurred())

			dataBytes, err := ioutil.ReadFile(tmpFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(string(dataBytes)).To(ContainSubstring(d.Info))
		})
	})

	Describe(".Load", func() {
		It("Will load an empty json file", func() {
			d := &data{}

			err := diskPersistor.Load(d)
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Info).To(Equal(""))
		})

		It("Will load a json file with specific keys", func() {
			d := &data{}

			err := ioutil.WriteFile(tmpFile.Name(), []byte(`{"Info":"test string"}`), 0700)
			Expect(err).ToNot(HaveOccurred())

			err = diskPersistor.Load(d)
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Info).To(Equal("test string"))
		})
	})
})

type data struct {
	Info string
}

func (d *data) JSONMarshalV3() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

func (d *data) JSONUnmarshalV3(data []byte) error {
	return json.Unmarshal(data, d)
}
