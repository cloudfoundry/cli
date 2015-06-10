package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/command_runner"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/codegangsta/cli"
)

var deps = command_registry.NewDependency()

var cmdRegistry = command_registry.Commands

type cliDependencies struct {
	termUI         terminal.UI
	configRepo     core_config.Repository
	pluginConfig   plugin_config.PluginConfiguration
	manifestRepo   manifest.ManifestRepository
	apiRepoLocator api.RepositoryLocator
	gateways       map[string]net.Gateway
	teePrinter     *terminal.TeePrinter
	detector       detection.Detector
}

func main() {
	defer handlePanics(deps.TeePrinter)
	defer deps.Config.Close()

	//////////////// non-codegangsta path  ///////////////////////
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmdRegistry.CommandExists(cmd) {

			fc := flags.NewFlagContext(cmdRegistry.FindCommand(os.Args[1]).MetaData().Flags)

			if requestHelp(os.Args[2:]) {
				deps.Ui.Say(cmdRegistry.CommandUsage(cmd))
				os.Exit(0)
			} else {
				err := fc.Parse(os.Args[2:]...)
				if err != nil {
					deps.Ui.Failed("Incorrect Usage\n\n" + err.Error() + "\n\n" + cmdRegistry.CommandUsage(cmd))
				}
			}
			cmdRegistry.SetCommand(cmdRegistry.FindCommand(cmd).SetDependency(deps, false))

			cfCmd := cmdRegistry.FindCommand(cmd)
			reqs, err := cfCmd.Requirements(requirements.NewFactory(deps.Ui, deps.Config, deps.RepoLocator), fc)
			if err != nil {
				deps.Ui.Failed(err.Error())
			}

			for _, r := range reqs {
				if !r.Execute() {
					os.Exit(1)
				}
			}

			cfCmd.Execute(fc)
			os.Exit(0)
		}
	}
	//////////////////////////////////////////

	rpcService := newCliRpcServer(deps.TeePrinter, deps.TeePrinter)

	cmdFactory := command_factory.NewFactory(deps.Ui, deps.Config, deps.ManifestRepo, deps.RepoLocator, deps.PluginConfig, rpcService)
	requirementsFactory := requirements.NewFactory(deps.Ui, deps.Config, deps.RepoLocator)
	cmdRunner := command_runner.NewRunner(cmdFactory, requirementsFactory, deps.Ui)
	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { deps.Ui.Failed(fmt.Sprintf("Error read/writing plugin config: %s, ", err.Error())) })
	pluginList := pluginsConfig.Plugins()

	var badFlags string
	metaDatas := cmdFactory.CommandMetadatas()
	//return only metadata for current command
	metaDatas = mergePluginMetaData(metaDatas, pluginList)

	if len(os.Args) > 1 {
		flags := cmdFactory.GetCommandFlags(os.Args[1])
		totalArgs, _ := cmdFactory.GetCommandTotalArgs(os.Args[1])

		if args2skip := totalArgs + 2; len(os.Args) >= args2skip {
			badFlags = matchArgAndFlags(flags, os.Args[args2skip:])
		}

		if badFlags != "" {
			badFlags = badFlags + "\n\n"
		}
	}

	injectHelpTemplate(badFlags)

	theApp := app.NewApp(cmdRunner, metaDatas...)
	rpcService.SetTheApp(theApp)

	//command `cf` without argument
	if len(os.Args) == 1 || os.Args[1] == "help" || requestHelp(os.Args[2:]) {
		theApp.Run(os.Args)
	} else if cmdFactory.CheckIfCoreCmdExists(os.Args[1]) {
		callCoreCommand(os.Args[0:], theApp)
	} else {
		// run each plugin and find the method/
		// run method if exist
		ran := rpc.RunMethodIfExists(rpcService, os.Args[1:], pluginList)
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

func injectHelpTemplate(badFlags string) {
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
	gateways := gatewaySliceFromMap(deps.Gateways)

	warningsCollector := net.NewWarningsCollector(deps.Ui, gateways...)
	warningsCollector.PrintWarnings()
}

func matchArgAndFlags(flags []string, args []string) string {
	var badFlag string
	var lastPassed bool
	multipleFlagErr := false

Loop:
	for _, arg := range args {
		prefix := ""

		//only take flag name, ignore value after '='
		arg = strings.Split(arg, "=")[0]

		if arg == "--h" || arg == "-h" {
			continue Loop
		}

		if strings.HasPrefix(arg, "--") {
			prefix = "--"
		} else if strings.HasPrefix(arg, "-") {
			prefix = "-"
		}
		arg = strings.TrimLeft(arg, prefix)

		//skip verification for negative integers, e.g. -i -10
		if lastPassed {
			lastPassed = false
			if _, err := strconv.ParseInt(arg, 10, 32); err == nil {
				continue Loop
			}
		}

		if prefix != "" {
			for _, flag := range flags {
				if flag == arg {
					lastPassed = true
					continue Loop
				}
			}

			if badFlag == "" {
				badFlag = fmt.Sprintf("\"%s%s\"", prefix, arg)
			} else {
				multipleFlagErr = true
				badFlag = badFlag + fmt.Sprintf(", \"%s%s\"", prefix, arg)
			}
		}
	}

	if multipleFlagErr && badFlag != "" {
		badFlag = fmt.Sprintf("%s %s", T("Unknown flags:"), badFlag)
	} else if badFlag != "" {
		badFlag = fmt.Sprintf("%s %s", T("Unknown flag"), badFlag)
	}

	return badFlag
}

func mergePluginMetaData(coreMetas []command_metadata.CommandMetadata, pluginMetas map[string]plugin_config.PluginMetadata) []command_metadata.CommandMetadata {
	for _, meta := range pluginMetas {
		for _, cmd := range meta.Commands {
			tmpMeta := command_metadata.CommandMetadata{}
			if cmd.UsageDetails.Usage == "" {
				tmpMeta.Usage = "N/A"
			} else {
				tmpMeta.Usage = cmd.UsageDetails.Usage
			}
			tmpMeta.Name = cmd.Name
			tmpMeta.ShortName = cmd.Alias
			tmpMeta.Description = cmd.HelpText

			for k, v := range cmd.UsageDetails.Options {
				tmpMeta.Flags = append(tmpMeta.Flags, cli.BoolFlag{Name: k, Usage: v})
			}
			coreMetas = append(coreMetas, tmpMeta)
		}
	}

	return coreMetas
}

func requestHelp(args []string) bool {
	for _, v := range args {
		if v == "-h" || v == "--help" || v == "--h" {
			return true
		}
	}

	return false
}

func newCliRpcServer(outputCapture terminal.OutputCapture, terminalOutputSwitch terminal.TerminalOutputSwitch) *rpc.CliRpcService {
	cliServer, err := rpc.NewRpcService(nil, outputCapture, terminalOutputSwitch, deps.Config, deps.RepoLocator, rpc.NewNonCodegangstaRunner())
	if err != nil {
		fmt.Println("Error initializing RPC service: ", err)
		os.Exit(1)
	}

	return cliServer
}
