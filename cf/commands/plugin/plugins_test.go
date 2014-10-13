package plugin_test

import (
	"net/rpc"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	testconfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
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
		requirementsFactory *testreq.FakeReqFactory
		config              *testconfig.FakePluginConfiguration
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = &testconfig.FakePluginConfiguration{}

		rpc.DefaultServer = rpc.NewServer()
	})

	runCommand := func(args ...string) bool {
		cmd := NewPlugins(ui, config)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails if the plugin cannot be started", func() {
		config.PluginsReturns(map[string]string{
			"test_245":     "../no/executable/found",
			"anotherThing": "something",
		})

		runCommand()
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"test_245"},
			[]string{"anotherThing"},
		))
	})

	It("returns a list of available methods of a plugin", func() {
		config.PluginsReturns(map[string]string{
			"Test1": "../../../fixtures/plugins/test_1.exe",
		})

		runCommand()

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Listing Installed Plugins..."},
			[]string{"OK"},
			[]string{"Test1", "test_1_cmd1"},
			[]string{"Test1", "test_1_cmd2"},
		))
	})

	It("does not list the plugin when it provides no available commands", func() {
		config.PluginsReturns(map[string]string{
			"EmptyPlugin": "../../../fixtures/plugins/empty_plugin.exe",
		})

		runCommand()
		Expect(ui.Outputs).NotTo(ContainSubstrings(
			[]string{"EmptyPlugin"},
		))
	})

	It("list multiple plugins and their associated commands", func() {
		config.PluginsReturns(map[string]string{
			"Test1": "../../../fixtures/plugins/test_1.exe",
			"Test2": "../../../fixtures/plugins/test_2.exe",
		})

		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Test1", "test_1_cmd1"},
			[]string{"Test1", "test_1_cmd2"},
			[]string{"Test1", "help"},
			[]string{"Test2", "test_2_cmd1"},
			[]string{"Test2", "test_2_cmd2"},
		))
	})
})
