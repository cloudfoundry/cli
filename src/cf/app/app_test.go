package app_test

import (
	"bytes"
	"cf"
	"cf/api"
	. "cf/app"
	"cf/commands"
	"cf/net"
	"cf/trace"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testconfig "testhelpers/configuration"
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

var _ = Describe("App", func() {
	It("#NewApp", func() {
		ui := &testterm.FakeUI{}
		config := testconfig.NewRepository()
		manifestRepo := &testmanifest.FakeManifestRepository{}

		repoLocator := api.NewRepositoryLocator(config, map[string]net.Gateway{
			"auth":             net.NewUAAGateway(config),
			"cloud-controller": net.NewCloudControllerGateway(config),
			"uaa":              net.NewUAAGateway(config),
		})

		cmdFactory := commands.NewFactory(ui, config, manifestRepo, repoLocator)
		cmdRunner := &FakeRunner{cmdFactory: cmdFactory}

		for _, cmdName := range expectedCommandNames {
			output := bytes.NewBuffer(make([]byte, 1024))
			trace.SetStdout(output)
			trace.EnableTrace()

			app, err := NewApp(cmdRunner)
			Expect(err).NotTo(HaveOccurred())

			Expect(output.String()).To(ContainSubstring("VERSION:\n" + cf.Version))

			app.Run([]string{"", cmdName})
			Expect(cmdRunner.cmdName).To(Equal(cmdName))
		}
	})
})

type FakeRunner struct {
	cmdFactory commands.Factory
	cmdName    string
}

func (runner *FakeRunner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
	_, err = runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		GinkgoT().Fatal("Error instantiating command with name", cmdName)
		return
	}
	runner.cmdName = cmdName
	return
}
