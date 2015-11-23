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

		Expect(strings.Count(strings.Join(output, ""), "login")).To(Equal(1))
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
		Expect(commandInOutput("test2_really_long_really_long_really_long_command_name", output)).To(BeTrue(), "plugin command: test2_really_long_command_name not in help")
	})

	It("adjusts the output format to the longest length of plugin command name", func() {
		config_helpers.PluginRepoDir = func() string {
			return filepath.Join("..", "..", "fixtures", "config", "help-plugin-test-config")
		}

		dummyTemplate := `
{{range .Commands}}{{range .CommandSubGroups}}{{range .}}
{{.Name}}%%%{{.Description}}
{{end}}{{end}}{{end}}
`
		output := io_helpers.CaptureOutput(func() {
			help.ShowHelp(dummyTemplate)
		})

		cmdNameLen := len(strings.Split(output[2], "%%%")[0])

		for _, line := range output {
			if strings.TrimSpace(line) == "" {
				continue
			}

			expectedLen := len(strings.Split(line, "%%%")[0])
			Ω(cmdNameLen).To(Equal(expectedLen))
		}

	})

	It("does not show command's alias in help for installed plugin", func() {
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

		Expect(commandInOutput("test1_cmd1_alias", output)).To(BeFalse(), "plugin command alias: test1_cmd1_alias should not be in help")
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
