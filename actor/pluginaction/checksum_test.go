package pluginaction_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checksums", func() {
	var (
		actor      *Actor
		fakeConfig *pluginactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		actor = NewActor(fakeConfig, nil)
	})

	Describe("ValidateFileChecksum", func() {
		var file *os.File
		BeforeEach(func() {
			var err error
			file, err = ioutil.TempFile("", "")
			defer file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(file.Name(), []byte("foo"), 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(file.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the checksums match", func() {
			It("returns true", func() {
				Expect(actor.ValidateFileChecksum(file.Name(), "0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33")).To(BeTrue())
			})
		})

		Context("when the checksums do not match", func() {
			It("returns false", func() {
				Expect(actor.ValidateFileChecksum(file.Name(), "blah")).To(BeFalse())
			})
		})
	})
})
