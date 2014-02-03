package app

import (
	"cf/api"
	"cf/commands"
	"cf/configuration"
	"cf/net"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testconfig "testhelpers/configuration"
	testmanifest "testhelpers/manifest"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func availableCmdNames() (names []string) {
	reqFactory := &testreq.FakeReqFactory{}
	cmdRunner := commands.NewRunner(nil, reqFactory)
	app, _ := NewApp(cmdRunner)

	for _, cliCmd := range app.Commands {
		if cliCmd.Name != "help" {
			names = append(names, cliCmd.Name)
		}
	}
	return
}

type FakeRunner struct {
	cmdFactory commands.Factory
	t          mr.TestingT
	cmdName    string
}

func (runner *FakeRunner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
	_, err = runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		runner.t.Fatal("Error instantiating command with name", cmdName)
		return
	}
	runner.cmdName = cmdName
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCommands", func() {

			for _, cmdName := range availableCmdNames() {
				ui := &testterm.FakeUI{}
				config := &configuration.Configuration{}
				configRepo := testconfig.FakeConfigRepository{}
				manifestRepo := &testmanifest.FakeManifestRepository{}

				repoLocator := api.NewRepositoryLocator(config, configRepo, map[string]net.Gateway{
					"auth":             net.NewUAAGateway(),
					"cloud-controller": net.NewCloudControllerGateway(),
					"uaa":              net.NewUAAGateway(),
				})

				cmdFactory := commands.NewFactory(ui, config, configRepo, manifestRepo, repoLocator)
				cmdRunner := &FakeRunner{cmdFactory: cmdFactory, t: mr.T()}
				app, _ := NewApp(cmdRunner)
				app.Run([]string{"", cmdName})

				assert.Equal(mr.T(), cmdRunner.cmdName, cmdName)
			}
		})
	})
}
