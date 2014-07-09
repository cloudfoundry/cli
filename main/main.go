package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/codegangsta/cli"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

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

	deps := setupDependencies()
	defer deps.configRepo.Close()

	cmdFactory := command_factory.NewFactory(deps.termUI, deps.configRepo, deps.manifestRepo, deps.apiRepoLocator)
	requirementsFactory := requirements.NewFactory(deps.termUI, deps.configRepo, deps.apiRepoLocator)
	cmdRunner := command_runner.NewRunner(cmdFactory, requirementsFactory)

	app.NewApp(cmdRunner, cmdFactory.CommandMetadatas()...).Run(os.Args)

	gateways := gatewaySliceFromMap(deps.gateways)
	net.NewWarningsCollector(deps.termUI, gateways...).PrintWarnings()
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
	err := recover()
	if err != nil && err != terminal.FailedWasCalled {
		switch err := err.(type) {
		case errors.Exception:
			if err.DisplayCrashDialog {
				displayCrashDialog(err.Message)
			} else {
				fmt.Println(err.Message)
			}
		case error:
			displayCrashDialog(err.Error())
		case string:
			displayCrashDialog(err)
		default:
			displayCrashDialog("An unexpected type of error")
		}
	}

	if err != nil {
		os.Exit(1)
	}
}

func displayCrashDialog(errorMessage string) {
	formattedString := `

Aww shucks.

Something completely unexpected happened. This is a bug in %s.
Please file this bug : https://github.com/cloudfoundry/cli/issues
Tell us that you ran this command:

	%s

this error occurred:

	%s

and this stack trace:

%s
	`
	stackByteCount := 0
	STACK_SIZE_LIMIT := 1024 * 1024
	var bytes []byte
	for stackSize := 1024; (stackByteCount == 0 || stackByteCount == stackSize) && stackSize < STACK_SIZE_LIMIT; stackSize = 2 * stackSize {
		bytes = make([]byte, stackSize)
		stackByteCount = runtime.Stack(bytes, true)
	}
	stackTrace := "\t" + strings.Replace(string(bytes), "\n", "\n\t", -1)
	println(fmt.Sprintf(formattedString, cf.Name(), strings.Join(os.Args, " "), errorMessage, stackTrace))
}
