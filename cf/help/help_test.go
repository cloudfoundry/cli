package help_test

import (
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/help"

	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Help", func() {
	It("shows help for all commands", func() {
		dummyTemplate := `
{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
{{.Name}}
{{end}}{{end}}{{end}}
`
		output := io_helpers.CaptureOutput(func() {
			help.ShowHelp(dummyTemplate)
		})

		for _, metadata := range command_registry.Commands.Metadatas() {
			Expect(commandInOutput(metadata.Name, output)).To(BeTrue(), metadata.Name+" not in help")
		}
	})

	It("shows help for all installed plugin's commands", func() {
		config_helpers.PluginRepoDir = func() string {
			return filepath.Join("..", "..", "fixtures", "config", "help-plugin-test-config")
		}

		dummyTemplate := `
{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
{{.Name}}
{{end}}{{end}}{{end}}
`
		output := io_helpers.CaptureOutput(func() {
			help.ShowHelp(dummyTemplate)
		})

		Expect(commandInOutput("test1_cmd2", output)).To(BeTrue(), "plugin command: test1_cmd2 not in help")
		Expect(commandInOutput("test2_cmd1", output)).To(BeTrue(), "plugin command: test2_cmd1 not in help")
		Expect(commandInOutput("test2_cmd2", output)).To(BeTrue(), "plugin command: test2_cmd2 not in help")

	})

	It("shows command's alias in help for installed plugin", func() {
		config_helpers.PluginRepoDir = func() string {
			return filepath.Join("..", "..", "fixtures", "config", "help-plugin-test-config")
		}

		dummyTemplate := `
{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
{{.Name}}
{{end}}{{end}}{{end}}
`
		output := io_helpers.CaptureOutput(func() {
			help.ShowHelp(dummyTemplate)
		})

		Expect(commandInOutput("test1_cmd1, test1_cmd1_alias", output)).To(BeTrue(), "plugin command alias: test1_cmd1_alias not in help")
	})

})

func commandInOutput(cmdName string, output []string) bool {
	for _, line := range output {
		if strings.TrimSpace(line) == strings.TrimSpace(cmdName) {
			return true
		}
	}
	return false
}
