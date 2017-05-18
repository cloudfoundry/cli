// +build !windows

package pluginaction_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Checks file permissions for UNIX platforms
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

			It("gives the copy 0700 permission", func() {
				copyPath, err := actor.CreateExecutableCopy(pluginPath, tempPluginDir)
				Expect(err).ToNot(HaveOccurred())

				stat, err := os.Stat(copyPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Mode()).To(Equal(os.FileMode(0700)))
			})
		})
	})

	Describe("InstallPluginFromLocalPath", func() {
		var (
			plugin     configv3.Plugin
			installErr error

			pluginHomeDir string
			pluginPath    string
			tempDir       string
		)

		BeforeEach(func() {
			plugin = configv3.Plugin{
				Name: "some-plugin",
				Commands: []configv3.PluginCommand{
					{Name: "some-command"},
				},
			}

			pluginFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			err = pluginFile.Close()
			Expect(err).NotTo(HaveOccurred())

			pluginPath = pluginFile.Name()

			tempDir, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			pluginHomeDir = filepath.Join(tempDir, ".cf", "plugin")
		})

		AfterEach(func() {
			err := os.Remove(pluginPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			installErr = actor.InstallPluginFromPath(pluginPath, plugin)
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeConfig.PluginHomeReturns(pluginHomeDir)
			})

			It("gives the executable 0755 permission", func() {
				Expect(installErr).ToNot(HaveOccurred())

				installedPluginPath := filepath.Join(pluginHomeDir, "some-plugin")
				stat, err := os.Stat(installedPluginPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Mode()).To(Equal(os.FileMode(0755)))
			})
		})
	})
})
