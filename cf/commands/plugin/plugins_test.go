package plugin_test

import (
	"path/filepath"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	testconfig "github.com/cloudfoundry/cli/cf/configuration/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugins", func() {
	var (
		ui                  *testterm.FakeUI
		config              *testconfig.FakeRepository
		requirementsFactory *testreq.FakeReqFactory

		plugin1, plugin2, emptyPlugin string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = &testconfig.FakeRepository{}

		plugin1 = filepath.Join("..", "..", "..", "fixtures", "plugins", "test_1")
		plugin2 = filepath.Join("..", "..", "..", "fixtures", "plugins", "test_2")
		emptyPlugin = filepath.Join("..", "..", "..", "fixtures", "plugins", "empty_plugin")

	})

	runCommand := func(args ...string) bool {
		cmd := NewPlugins(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails if the plugin cannot be started", func() {
		config.PluginsReturns(map[string]string{"test_245": "not/a/path/you/fool"})
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
		))
	})

	It("returns a list of available methods of a plugin", func() {
		config.PluginsReturns(map[string]string{"test_1": plugin1})
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Listing Installed Plugins..."},
			[]string{"OK"},
			[]string{"test_1", "test_1_cmd1"},
			[]string{"test_1", "test_1_cmd2"},
		))
	})

	It("does not list the plugin when it provides no available commands", func() {
		config.PluginsReturns(map[string]string{"empty_plugin": emptyPlugin})
		runCommand()
		Expect(ui.Outputs).NotTo(ContainSubstrings(
			[]string{"empty_plugin"},
		))
	})

	It("list multiple plugins and their associated commands", func() {
		config.PluginsReturns(map[string]string{
			"test_1": plugin1,
			"test_2": plugin2,
		})
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"test_1", "test_1_cmd1"},
			[]string{"test_1", "test_1_cmd2"},
			[]string{"test_1", "help"},
			[]string{"test_2", "test_2_cmd1"},
			[]string{"test_2", "test_2_cmd2"},
		))
	})
})
