package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/commands_loader"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/rpc"
)

var cmdRegistry = command_registry.Commands

func main() {
	traceEnv := os.Getenv("CF_TRACE")
	traceLogger := trace.NewLogger(false, traceEnv, "")

	//handle `cf -v` for cf version
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		os.Args[1] = "version"
	}

	//handles `cf`
	if len(os.Args) == 1 {
		os.Args = []string{os.Args[0], "help"}
	}

	//handles `cf [COMMAND] -h ...`
	//rearrange args to `cf help COMMAND` and let `command help` to print out usage
	os.Args = append([]string{os.Args[0]}, handleHelp(os.Args[1:])...)

	newArgs, isVerbose := handleVerbose(os.Args)
	os.Args = newArgs
	traceLogger = trace.NewLogger(isVerbose, traceEnv, "")

	errFunc := func(err error) {
		if err != nil {
			ui := terminal.NewUI(os.Stdin, terminal.NewTeePrinter(), traceLogger)
			ui.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}

	// Only used to get Trace, so our errorHandler doesn't matter, since it's not used
	config := core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), errFunc)
	defer config.Close()

	traceConfigVal := config.Trace()

	traceLogger = trace.NewLogger(isVerbose, traceEnv, traceConfigVal)

	deps := command_registry.NewDependency(traceLogger)
	defer handlePanics(deps.TeePrinter, deps.Logger)
	defer deps.Config.Close()

	//handle `cf --build`
	if len(os.Args) == 2 && (os.Args[1] == "--build" || os.Args[1] == "-b") {
		deps.Ui.Say(T("{{.CFName}} was built with Go version: {{.GoVersion}}",
			map[string]interface{}{
				"CFName":    os.Args[0],
				"GoVersion": runtime.Version(),
			}))
		os.Exit(0)
	}

	warningProducers := []net.WarningProducer{}
	for _, warningProducer := range deps.Gateways {
		warningProducers = append(warningProducers, warningProducer)
	}

	warningsCollector := net.NewWarningsCollector(deps.Ui, warningProducers...)

	commands_loader.Load()

	//run core command
	cmdName := os.Args[1]
	cmd := cmdRegistry.FindCommand(cmdName)
	if cmd != nil {
		meta := cmd.MetaData()
		flagContext := flags.NewFlagContext(meta.Flags)
		flagContext.SkipFlagParsing(meta.SkipFlagParsing)

		cmdArgs := os.Args[2:]
		err := flagContext.Parse(cmdArgs...)
		if err != nil {
			usage := cmdRegistry.CommandUsage(cmdName)
			deps.Ui.Failed(T("Incorrect Usage") + "\n\n" + err.Error() + "\n\n" + usage)
		}

		cmd = cmd.SetDependency(deps, false)
		cmdRegistry.SetCommand(cmd)

		requirementsFactory := requirements.NewFactory(deps.Config, deps.RepoLocator)
		reqs := cmd.Requirements(requirementsFactory, flagContext)

		for _, req := range reqs {
			err = req.Execute()
			if err != nil {
				deps.Ui.Failed(err.Error())
			}
		}

		cmd.Execute(flagContext)

		warningsCollector.PrintWarnings()

		os.Exit(0)
	}

	//non core command, try plugin command
	rpcService, err := rpc.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpc.NewCommandRunner(), deps.Logger)
	if err != nil {
		deps.Ui.Say(T("Error initializing RPC service: ") + err.Error())
		os.Exit(1)
	}

	pluginConfig := plugin_config.NewPluginConfig(func(err error) {
		deps.Ui.Failed(fmt.Sprintf("Error read/writing plugin config: %s, ", err.Error()))
	})
	pluginList := pluginConfig.Plugins()

	ran := rpc.RunMethodIfExists(rpcService, os.Args[1:], pluginList)
	if !ran {
		deps.Ui.Say("'" + os.Args[1] + T("' is not a registered command. See 'cf help'"))
		os.Exit(1)
	}

}

func handlePanics(printer terminal.Printer, logger trace.Printer) {
	panic_printer.UI = terminal.NewUI(os.Stdin, printer, logger)

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

func handleHelp(args []string) []string {
	hIndex := -1

	for i, v := range args {
		if v == "-h" || v == "--help" || v == "--h" {
			hIndex = i
			break
		}
	}

	if hIndex == -1 {
		return args
	} else if len(args) > 1 {
		if hIndex == 0 {
			return []string{"help", args[1]}
		} else {
			return []string{"help", args[0]}
		}
	} else {
		return []string{"help"}
	}
}

func handleVerbose(args []string) ([]string, bool) {
	var verbose bool
	idx := -1

	for i, arg := range args {
		if arg == "-v" {
			idx = i
			break
		}
	}

	if idx != -1 && len(args) > 1 {
		verbose = true
		args = append(args[:idx], args[idx+1:]...)
	}

	return args, verbose
}
