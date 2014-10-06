package rpc_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	. "github.com/cloudfoundry/cli/plugin/rpc"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rpc", func() {

	Context("GetAllPluginCommands", func() {
		BeforeEach(func() {
			config_helpers.PluginRepoDir = func() string {
				return filepath.Join("..", "..", "fixtures", "config", "help-plugin-test-config")
			}
		})
		It("returns a list of installed plugin commands", func() {
			actual := GetAllPluginCommands()
			Expect(len(actual)).To(Equal(5))
			Expect(actual[0].Name).To(Equal("test_1_cmd1"))
			Expect(actual[0].HelpText).To(Equal("help text for test_1_cmd1"))
			Expect(actual[1].Name).To(Equal("test_1_cmd2"))
			Expect(actual[1].HelpText).To(Equal("help text for test_1_cmd2"))
			Expect(actual[2].Name).To(Equal("help"))
			Expect(actual[2].HelpText).To(Equal("help text for test_1_help"))
			Expect(actual[3].Name).To(Equal("test_2_cmd1"))
			Expect(actual[3].HelpText).To(Equal("help text for test_2_cmd1"))
			Expect(actual[4].Name).To(Equal("test_2_cmd2"))
			Expect(actual[4].HelpText).To(Equal("help text for test_2_cmd2"))
		})
	})
})
