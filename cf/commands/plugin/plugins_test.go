package plugin_test

import (
	"path/filepath"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
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
	)

	BeforeEach(func() {

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		config_helpers.UserHomeDir = func() string {
			return filepath.Join("..", "..", "..", "fixtures", "config", "plugin-config")
		}
	})

	runCommand := func(args ...string) bool {
		cmd := NewPlugins(ui)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails if the plugin cannot be started", func() {
		config_helpers.UserHomeDir = func() string {
			return filepath.Join("..", "..", "..", "fixtures", "config", "bad-plugin-config")
		}

		runCommand()
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"test_245"},
		))
	})

	It("returns a list of available methods of a plugin", func() {
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Listing Installed Plugins..."},
			[]string{"OK"},
			[]string{"test_1", "test_1_cmd1"},
			[]string{"test_1", "test_1_cmd2"},
		))
	})

	It("does not list the plugin when it provides no available commands", func() {
		runCommand()
		Expect(ui.Outputs).NotTo(ContainSubstrings(
			[]string{"empty_plugin"},
		))
	})

	It("list multiple plugins and their associated commands", func() {
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
