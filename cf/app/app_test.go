package app_test

import (
	"bytes"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_factory"
	testPluginConfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/plugin/rpc"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	testmanifest "github.com/cloudfoundry/cli/testhelpers/manifest"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/app"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// skipping test during the process of converting commands into non-codegangsta structure
var expectedCommandNames = []string{}

// var expectedCommandNames = []string{
// 	"bind-service", "buildpacks", "create-buildpack",
// 	"create-domain", "create-org", "create-route", "create-service", "create-service-auth-token",
// 	"create-service-broker", "create-space", "create-user", "create-user-provided-service", "curl",
// 	"delete", "delete-buildpack", "delete-domain", "delete-shared-domain", "delete-org", "delete-route",
// 	"delete-service", "delete-service-auth-token", "delete-service-broker", "delete-space", "delete-user",
// 	"domains", "env", "events", "files", "logs", "marketplace", "map-route",
// 	"passwd", "purge-service-offering", "push", "quotas", "rename", "rename-org",
// 	"rename-service", "rename-service-broker", "rename-space", "restage", "restart", "routes", "scale",
// 	"service-auth-tokens", "service-brokers", "set-env", "set-org-role",
// 	"set-space-role", "create-shared-domain", "stacks", "start", "stop",
// 	"unbind-service", "unmap-route", "unset-env", "unset-org-role", "unset-space-role",
// 	"update-buildpack", "update-service-broker", "update-service-auth-token", "update-user-provided-service",
// 	"quotas", "create-quota", "delete-quota", "quota", "set-quota", "install-plugin", "plugins", "uninstall-plugin",
// }

var _ = Describe("App", func() {
	var (
		app       *cli.App
		cmdRunner *FakeRunner
	)

	JustBeforeEach(func() {
		ui := &testterm.FakeUI{}
		config := testconfig.NewRepository()
		pluginConfig := &testPluginConfig.FakePluginConfiguration{}
		manifestRepo := &testmanifest.FakeManifestRepository{}

		repoLocator := api.NewRepositoryLocator(config, map[string]net.Gateway{
			"auth":             net.NewUAAGateway(config, ui),
			"cloud-controller": net.NewCloudControllerGateway(config, time.Now, &testterm.FakeUI{}),
			"uaa":              net.NewUAAGateway(config, ui),
		})

		rpcService, _ := rpc.NewRpcService(nil, nil, nil, nil, api.RepositoryLocator{}, nil)
		cmdFactory := command_factory.NewFactory(ui, config, manifestRepo, repoLocator, pluginConfig, rpcService)
		cmdRunner = &FakeRunner{cmdFactory: cmdFactory}
		app = NewApp(cmdRunner, cmdFactory.CommandMetadatas()...)
	})

	Describe("trace file integration", func() {
		var (
			output *bytes.Buffer
		)

		BeforeEach(func() {
			output = bytes.NewBuffer(make([]byte, 1024))
			trace.SetStdout(output)
			trace.EnableTrace()
		})

		It("prints its version number to the trace output when constructed", func() {
			Expect(strings.Split(output.String(), "\n")).To(ContainSubstrings(
				[]string{"VERSION:"},
				[]string{cf.Version},
			))
		})
	})

	Context("when given a command name to run", func() {
		It("runs the command with that name", func() {
			for _, cmdName := range expectedCommandNames {
				app.Run([]string{"", cmdName})
				Expect(cmdRunner.cmdName).To(Equal(cmdName))
			}
		})
	})

	Context("when the user provides an unknown command name", func() {
		It("should complain loudly and then panic", func() {
			Expect(func() {
				io_helpers.CaptureOutput(func() {
					app.Run([]string{"cf", "zoidberg"})
				})
			}).To(Panic())
		})
	})

	It("includes the built on date in its version string", func() {
		Expect(app.Version).To(Equal(cf.Version + "-" + cf.BuiltOnDate))
	})
})

type FakeRunner struct {
	cmdFactory command_factory.Factory
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
