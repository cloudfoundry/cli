package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/codegangsta/cli"
)

var deps = setupDependencies()

type cliDependencies struct {
	termUI         terminal.UI
	configRepo     configuration.Repository
	manifestRepo   manifest.ManifestRepository
	apiRepoLocator api.RepositoryLocator
	gateways       map[string]net.Gateway
}

func setupDependencies() (deps *cliDependencies) {
	deps = new(cliDependencies)

	deps.termUI = terminal.NewUI(os.Stdin)

	deps.manifestRepo = manifest.NewManifestDiskRepository()

	deps.configRepo = configuration.NewRepositoryFromFilepath(configuration.DefaultFilePath(), func(err error) {
		if err != nil {
			deps.termUI.Failed(fmt.Sprintf("Config error: %s", err))
		}
	})

	i18n.T = i18n.Init(deps.configRepo)

	terminal.UserAskedForColors = deps.configRepo.ColorEnabled()
	terminal.InitColorSupport()

	if os.Getenv("CF_TRACE") != "" {
		trace.Logger = trace.NewLogger(os.Getenv("CF_TRACE"))
	} else {
		trace.Logger = trace.NewLogger(deps.configRepo.Trace())
	}

	deps.gateways = map[string]net.Gateway{
		"auth":             net.NewUAAGateway(deps.configRepo),
		"cloud-controller": net.NewCloudControllerGateway(deps.configRepo, time.Now),
		"uaa":              net.NewUAAGateway(deps.configRepo),
	}
	deps.apiRepoLocator = api.NewRepositoryLocator(deps.configRepo, deps.gateways)

	return
}

func main() {
	defer handlePanics()
	defer deps.configRepo.Close()

	cmdFactory := command_factory.NewFactory(deps.termUI, deps.configRepo, deps.manifestRepo, deps.apiRepoLocator)
	requirementsFactory := requirements.NewFactory(deps.termUI, deps.configRepo, deps.apiRepoLocator)
	cmdRunner := command_runner.NewRunner(cmdFactory, requirementsFactory)

	err := app.NewApp(cmdRunner, cmdFactory.CommandMetadatas()...).Run(os.Args)
	if err != nil {
		os.Exit(1)
	}

	gateways := gatewaySliceFromMap(deps.gateways)

	warningsCollector := net.NewWarningsCollector(deps.termUI, gateways...)
	warningsCollector.PrintWarnings()
}

func gatewaySliceFromMap(gateway_map map[string]net.Gateway) []net.WarningProducer {
	gateways := []net.WarningProducer{}
	for _, gateway := range gateway_map {
		gateways = append(gateways, gateway)
	}
	return gateways
}

func init() {
	cli.CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Description}}
{{with .ShortName}}
ALIAS:
   {{.}}
{{end}}
USAGE:
   {{.Usage}}{{with .Flags}}

OPTIONS:
{{range .}}   {{.}}
{{end}}{{else}}
{{end}}`

}

func handlePanics() {
	panic_printer.UI = terminal.NewUI(os.Stdin)

	commandArgs := strings.Join(os.Args, " ")
	stackTrace := generateBacktrace()

	err := recover()
	panic_printer.DisplayCrashDialog(err, commandArgs, stackTrace)

	if err != nil {
		os.Exit(1)
	}
}

func generateBacktrace() string {
	stackByteCount := 0
	STACK_SIZE_LIMIT := 1024 * 1024
	var bytes []byte
	for stackSize := 1024; (stackByteCount == 0 || stackByteCount == stackSize) && stackSize < STACK_SIZE_LIMIT; stackSize = 2 * stackSize {
		bytes = make([]byte, stackSize)
		stackByteCount = runtime.Stack(bytes, true)
	}
	stackTrace := "\t" + strings.Replace(string(bytes), "\n", "\n\t", -1)
	return stackTrace
}
