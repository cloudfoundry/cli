// +build windows

package pluginaction_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("install actions", func() {
	var (
		actor         *Actor
		fakeConfig    *pluginactionfakes.FakeConfig
		tempPluginDir string
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		var err error
		tempPluginDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		actor = NewActor(fakeConfig, nil)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempPluginDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("CreateExecutableCopy", func() {
		Context("when the file exists", func() {
			var pluginPath string

			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = tempFile.WriteString("cthulhu")
				Expect(err).ToNot(HaveOccurred())
				err = tempFile.Close()
				Expect(err).ToNot(HaveOccurred())

				pluginPath = tempFile.Name()
			})

			AfterEach(func() {
				err := os.Remove(pluginPath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("adds .exe to the end of the filename", func() {
				copyPath, err := actor.CreateExecutableCopy(pluginPath, tempPluginDir)
				Expect(err).ToNot(HaveOccurred())

				Expect(copyPath).To(HaveSuffix(".exe"))
			})
		})
	})
})
