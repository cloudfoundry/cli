package help_test

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/help"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/plugin/rpc"

	testPluginConfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

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

func createCommandFactory() command_factory.Factory {
	fakeUI := &testterm.FakeUI{}
	configRepo := testconfig.NewRepository()
	pluginConfig := &testPluginConfig.FakePluginConfiguration{}

	manifestRepo := manifest.NewManifestDiskRepository()
	apiRepoLocator := api.NewRepositoryLocator(configRepo, map[string]net.Gateway{
		"auth":             net.NewUAAGateway(configRepo, fakeUI),
		"cloud-controller": net.NewCloudControllerGateway(configRepo, time.Now, fakeUI),
		"uaa":              net.NewUAAGateway(configRepo, fakeUI),
	})

	rpcService, _ := rpc.NewRpcService(nil, nil, nil, nil, api.RepositoryLocator{}, nil)
	return command_factory.NewFactory(fakeUI, configRepo, manifestRepo, apiRepoLocator, pluginConfig, rpcService)
}

func commandInOutput(cmdName string, output []string) bool {
	for _, line := range output {
		if strings.TrimSpace(line) == strings.TrimSpace(cmdName) {
			return true
		}
	}
	return false
}
