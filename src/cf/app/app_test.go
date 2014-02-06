package app_test

import (
	"cf/api"
	. "cf/app"
	"cf/commands"
	"cf/configuration"
	"cf/net"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testmanifest "testhelpers/manifest"
	testterm "testhelpers/terminal"
)

var expectedCommandNames = []string{
	"api", "app", "apps", "auth", "bind-service", "buildpacks", "create-buildpack",
	"create-domain", "create-org", "create-route", "create-service", "create-service-auth-token",
	"create-service-broker", "create-space", "create-user", "create-user-provided-service", "curl",
	"delete", "delete-buildpack", "delete-domain", "delete-shared-domain", "delete-org", "delete-route",
	"delete-service", "delete-service-auth-token", "delete-service-broker", "delete-space", "delete-user",
	"domains", "env", "events", "files", "login", "logout", "logs", "marketplace", "map-route", "org",
	"org-users", "orgs", "passwd", "purge-service-offering", "push", "quotas", "rename", "rename-org",
	"rename-service", "rename-service-broker", "rename-space", "restart", "routes", "scale",
	"service", "service-auth-tokens", "service-brokers", "services", "set-env", "set-org-role", "set-quota",
	"set-space-role", "create-shared-domain", "space", "space-users", "spaces", "stacks", "start", "stop",
	"target", "unbind-service", "unmap-route", "unset-env", "unset-org-role", "unset-space-role",
	"update-buildpack", "update-service-broker", "update-service-auth-token", "update-user-provided-service",
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
			ui := &testterm.FakeUI{}
			config := &configuration.Configuration{}
			manifestRepo := &testmanifest.FakeManifestRepository{}

			repoLocator := api.NewRepositoryLocator(config, map[string]net.Gateway{
				"auth":             net.NewUAAGateway(),
				"cloud-controller": net.NewCloudControllerGateway(),
				"uaa":              net.NewUAAGateway(),
			})

			t := mr.T()
			cmdFactory := commands.NewFactory(ui, config, manifestRepo, repoLocator)
			cmdRunner := &FakeRunner{cmdFactory: cmdFactory, t: t}

			for _, cmdName := range expectedCommandNames {
				app, _ := NewApp(cmdRunner)
				app.Run([]string{"", cmdName})
				assert.Equal(t, cmdRunner.cmdName, cmdName)
			}
		})
	})
}
