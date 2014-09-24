package plugin_test

import (
	"io/ioutil"
	"os"
	"path"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/fileutils"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		config              configuration.Repository
		requirementsFactory *testreq.FakeReqFactory

		pluginFile *os.File
		homeDir    string
		curDir     string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		cmd := NewPluginInstall(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	setupTempExecutable := func() {
		var err error

		homeDir = path.Join(configuration.UserHomePath(), ".cf", "plugin")

		curDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		pluginFile, err = ioutil.TempFile("./", "test_plugin")
		Expect(err).ToNot(HaveOccurred())
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("failures", func() {
		It("if plugin name is already taken", func() {
			config.SetPlugin("fake_plugin", "/going/to/nowhere/fake_plugin")
			runCommand("/going/to/nowhere/fake_plugin")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin name", "fake_plugin", "is already taken"},
				[]string{"FAILED"},
			))
		})

		Context("io", func() {
			AfterEach(func() {
				os.Remove(path.Join(curDir, pluginFile.Name()))
				os.Remove(path.Join(homeDir, pluginFile.Name()))
			})

			It("if a file with the plugin name already exists under ~/.cf/plugin/", func() {
				setupTempExecutable()
				fileutils.CopyFile(path.Join(homeDir, pluginFile.Name()), path.Join(curDir, pluginFile.Name()))

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
			os.Remove(path.Join(homeDir, pluginFile.Name()))
		})

		It("copies the plugin into directory ~/.cf/plugin/PLUGIN_NAME", func() {
			_, err := os.Stat(path.Join(curDir, pluginFile.Name()))
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(path.Join(homeDir, pluginFile.Name()))
			Expect(err).ToNot(HaveOccurred())
		})

		It("populate the configuration map with the plugin name and location", func() {
			Expect(config.Plugins()[pluginFile.Name()]).To(Equal(path.Join(homeDir, pluginFile.Name())))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", pluginFile.Name()},
				[]string{"OK"},
				[]string{"Plugin", pluginFile.Name(), "successfully installed"},
			))
		})

	})
})
