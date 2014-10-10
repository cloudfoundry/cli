package plugin_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	testconfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              *testconfig.FakePluginConfiguration

		pluginFile      *os.File
		old_PLUGINS_DIR string
		homeDir         string
		pluginDir       string
		curDir          string

		test_1 string
		test_2 string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = &testconfig.FakePluginConfiguration{}

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		test_1 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.exe")
		test_2 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.exe")
	})

	AfterEach(func() {
		err := os.Setenv("CF_PLUGINS_DIR", old_PLUGINS_DIR)
		Expect(err).NotTo(HaveOccurred())
	})

	runCommand := func(args ...string) bool {
		cmd := NewPluginInstall(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	setupTempExecutable := func() {
		var err error

		homeDir, err = ioutil.TempDir(os.TempDir(), "plugins")
		Expect(err).ToNot(HaveOccurred())

		old_PLUGINS_DIR = os.Getenv("CF_PLUGINS_DIR")
		err = os.Setenv("CF_PLUGINS_DIR", homeDir)
		Expect(err).NotTo(HaveOccurred())

		pluginDir = filepath.Join(homeDir, ".cf", "plugins")

		curDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		pluginFile, err = ioutil.TempFile("./", "test_plugin")
		Expect(err).ToNot(HaveOccurred())

		if runtime.GOOS != "windows" {
			err = os.Chmod(test_1, 0700)
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
			config.PluginsReturns(map[string]string{"CliPlugin": "do/not/care"})
			runCommand(test_1)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin name", "CliPlugin", "is already taken"},
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
				os.Remove(filepath.Join(curDir, pluginFile.Name()))
				os.Remove(homeDir)
			})

			It("if a file with the plugin name already exists under ~/.cf/plugin/", func() {
				config.PluginsReturns(map[string]string{"useless": "do/not/care"})
				config.GetPluginPathReturns(curDir)

				runCommand(filepath.Join(curDir, pluginFile.Name()))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Installing plugin"},
					[]string{"The file", pluginFile.Name(), "already exists"},
					[]string{"FAILED"},
				))
			})
		})
	})

	Describe("success", func() {
		var (
			sourceBinaryPath string
		)

		BeforeEach(func() {
			setupTempExecutable()

			err := os.MkdirAll(pluginDir, 0700)
			Expect(err).ToNot(HaveOccurred())

			sourceBinaryPath = filepath.Join("..", "..", "..", "fixtures", "plugins", "test_1.exe")
			config.GetPluginPathReturns(pluginDir)
			runCommand(sourceBinaryPath)
		})

		AfterEach(func() {
			os.Remove(filepath.Join(curDir, pluginFile.Name()))
			os.Remove(homeDir)
		})

		It("copies the plugin into directory <FAKE_HOME_DIR>/.cf/plugins/PLUGIN_FILE_NAME", func() {
			_, err := os.Stat(test_1)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())
		})

		if runtime.GOOS != "windows" {
			It("Chmods the plugin so it is executable", func() {
				fileInfo, err := os.Stat(filepath.Join(pluginDir, "test_1.exe"))
				Expect(err).ToNot(HaveOccurred())
				Expect(int(fileInfo.Mode())).To(Equal(0700))
			})
		}

		It("populate the configuration map with the plugin name and location", func() {
			pluginName, pluginExecutable := config.SetPluginArgsForCall(0)

			Expect(pluginName).To(Equal("CliPlugin"))
			Expect(pluginExecutable).To(Equal(filepath.Join(pluginDir, "test_1.exe")))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", sourceBinaryPath},
				[]string{"OK"},
				[]string{"Plugin", "CliPlugin", "successfully installed"},
			))
		})

	})
})
