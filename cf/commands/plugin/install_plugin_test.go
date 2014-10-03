package plugin_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/fileutils"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory

		pluginFile *os.File
		homeDir    string
		pluginDir  string
		curDir     string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		config_helpers.PluginRepoDir = func() string {
			return filepath.Join("..", "..", "..", "fixtures", "config", "plugin-config")
		}
	})

	runCommand := func(args ...string) bool {
		cmd := NewPluginInstall(ui)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	setupTempExecutable := func() {
		var err error

		homeDir, err = ioutil.TempDir(os.TempDir(), "plugins")
		Expect(err).ToNot(HaveOccurred())

		config_helpers.PluginRepoDir = func() string {
			return homeDir
		}

		pluginDir = path.Join(homeDir, ".cf", "plugins")

		curDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		pluginFile, err = ioutil.TempFile("./", "test_plugin")
		Expect(err).ToNot(HaveOccurred())

		if runtime.GOOS != "windows" {
			err = os.Chmod(path.Join(curDir, pluginFile.Name()), 0700)
			Expect(err).ToNot(HaveOccurred())
		}
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("failures", func() {
		It("if plugin name is already taken", func() {
			runCommand(filepath.Join("..", "..", "..", "fixtures", "plugins", "test_1"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin name", "test_1", "is already taken"},
				[]string{"FAILED"},
			))
		})

		Context("io", func() {
			BeforeEach(func() {
				setupTempExecutable()
				err := os.MkdirAll(pluginDir, 0700)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove(path.Join(curDir, pluginFile.Name()))
				os.Remove(homeDir)
			})

			It("if a file with the plugin name already exists under ~/.cf/plugin/", func() {
				err := fileutils.CopyFile(path.Join(pluginDir, pluginFile.Name()), path.Join(curDir, pluginFile.Name()))
				Expect(err).NotTo(HaveOccurred())

				runCommand(path.Join(curDir, pluginFile.Name()))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Installing plugin"},
					[]string{"The file", pluginFile.Name(), "already exists"},
					[]string{"FAILED"},
				))
			})
		})
	})

	Describe("success", func() {
		BeforeEach(func() {
			setupTempExecutable()
			runCommand(path.Join(curDir, pluginFile.Name()))
		})

		AfterEach(func() {
			os.Remove(path.Join(curDir, pluginFile.Name()))
			os.Remove(homeDir)
		})

		It("copies the plugin into directory <FAKE_HOME_DIR>/.cf/plugins/PLUGIN_NAME", func() {
			_, err := os.Stat(path.Join(curDir, pluginFile.Name()))
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(path.Join(pluginDir, pluginFile.Name()))
			Expect(err).ToNot(HaveOccurred())
		})

		if runtime.GOOS != "windows" {
			It("Chmods the plugin so it is executable", func() {
				fileInfo, err := os.Stat(path.Join(pluginDir, pluginFile.Name()))
				Expect(err).ToNot(HaveOccurred())
				Expect(int(fileInfo.Mode())).To(Equal(0700))
			})
		}

		It("populate the configuration map with the plugin name and location", func() {
			pluginConfig := plugin_config.NewPluginConfig(func(err error) { Expect(err).ToNot(HaveOccurred()) })
			plugins := pluginConfig.Plugins()

			Expect(plugins[pluginFile.Name()]).To(Equal(path.Join(pluginDir, pluginFile.Name())))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", pluginFile.Name()},
				[]string{"OK"},
				[]string{"Plugin", pluginFile.Name(), "successfully installed"},
			))
		})

	})
})
