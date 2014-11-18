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
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

var deps = setupDependencies()

type cliDependencies struct {
	termUI         terminal.UI
	configRepo     core_config.Repository
	pluginConfig   plugin_config.PluginConfiguration
	manifestRepo   manifest.ManifestRepository
	apiRepoLocator api.RepositoryLocator
	gateways       map[string]net.Gateway
	teePrinter     *terminal.TeePrinter
}

func setupDependencies() (deps *cliDependencies) {
	deps = new(cliDependencies)

	deps.teePrinter = terminal.NewTeePrinter()

	deps.termUI = terminal.NewUI(os.Stdin, deps.teePrinter)

	deps.manifestRepo = manifest.NewManifestDiskRepository()

	errorHandler := func(err error) {
		if err != nil {
			deps.termUI.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}
	deps.configRepo = core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), errorHandler)
	deps.pluginConfig = plugin_config.NewPluginConfig(errorHandler)

	T = Init(deps.configRepo)

	terminal.UserAskedForColors = deps.configRepo.ColorEnabled()
	terminal.InitColorSupport()

	if os.Getenv("CF_TRACE") != "" {
		trace.Logger = trace.NewLogger(os.Getenv("CF_TRACE"))
	} else {
		trace.Logger = trace.NewLogger(deps.configRepo.Trace())
	}

	deps.gateways = map[string]net.Gateway{
		"auth":             net.NewUAAGateway(deps.configRepo, deps.termUI),
		"cloud-controller": net.NewCloudControllerGateway(deps.configRepo, time.Now, deps.termUI),
		"uaa":              net.NewUAAGateway(deps.configRepo, deps.termUI),
	}
	deps.apiRepoLocator = api.NewRepositoryLocator(deps.configRepo, deps.gateways)

	return
}

func main() {
	defer handlePanics(deps.teePrinter)
	defer deps.configRepo.Close()

	cmdFactory := command_factory.NewFactory(deps.termUI, deps.configRepo, deps.manifestRepo, deps.apiRepoLocator, deps.pluginConfig)
	requirementsFactory := requirements.NewFactory(deps.termUI, deps.configRepo, deps.apiRepoLocator)
	cmdRunner := command_runner.NewRunner(cmdFactory, requirementsFactory, deps.termUI)

	var badFlags string
	metaDatas := cmdFactory.CommandMetadatas()

	if len(os.Args) > 1 {
		var flags []string
		for _, cmd := range metaDatas {
			if os.Args[1] == cmd.Name || os.Args[1] == cmd.ShortName {
				for _, flag := range cmd.Flags {
					switch t := flag.(type) {
					default:
					case cli.IntFlag:
						flags = append(flags, t.Name)
					case cli.StringFlag:
						flags = append(flags, t.Name)
					case cli.BoolFlag:
						flags = append(flags, t.Name)
					}
				}
			}
		}

		badFlags = matchArgAndFlags(flags, os.Args[2:])
		if badFlags != "" {
			badFlags = badFlags + "\n\n"
		}
	}

	injectTemplate(badFlags)

	theApp := app.NewApp(cmdRunner, metaDatas...)
	//command `cf` without argument
	if len(os.Args) == 1 || os.Args[1] == "help" {
		theApp.Run(os.Args)
	} else if cmdFactory.CheckIfCoreCmdExists(os.Args[1]) {
		callCoreCommand(os.Args[0:], theApp)
	} else {
		// run each plugin and find the method/
		// run method if exist
		ran := rpc.RunMethodIfExists(theApp, os.Args[1:], deps.teePrinter, deps.teePrinter)
		if !ran {
			theApp.Run(os.Args)
		}
	}
}

func gatewaySliceFromMap(gateway_map map[string]net.Gateway) []net.WarningProducer {
	gateways := []net.WarningProducer{}
	for _, gateway := range gateway_map {
		gateways = append(gateways, gateway)
	}
	return gateways
}

func injectTemplate(badFlags string) {
	cli.CommandHelpTemplate = fmt.Sprintf(`%sNAME:
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
{{end}}`, badFlags)
}

func handlePanics(printer terminal.Printer) {
	panic_printer.UI = terminal.NewUI(os.Stdin, printer)

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

func callCoreCommand(args []string, theApp *cli.App) {
	err := theApp.Run(args)
	if err != nil {
		os.Exit(1)
	}
	gateways := gatewaySliceFromMap(deps.gateways)

	warningsCollector := net.NewWarningsCollector(deps.termUI, gateways...)
	warningsCollector.PrintWarnings()
}

func matchArgAndFlags(flags []string, args []string) string {
	var badFlag string

Loop:
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			arg = strings.TrimLeft(arg, "--")
			for _, flag := range flags {
				if flag == arg {
					continue Loop
				}
			}
			if badFlag == "" {
				badFlag = fmt.Sprintf("%s \"--%s\"", T("Unknown flag"), arg)
			} else {
				badFlag = strings.Replace(badFlag, T("Unknown flag"), T("Unknown flags:"), 1)
				badFlag = badFlag + fmt.Sprintf(", \"--%s\"", arg)
			}
		}
	}
	return badFlag
}
