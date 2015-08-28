package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/help"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/panic_printer"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/commands_loader"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/simonleung8/flags"
)

var deps = command_registry.NewDependency()

var cmdRegistry = command_registry.Commands

func main() {
	commands_loader.Load()

	defer handlePanics(deps.TeePrinter)
	defer deps.Config.Close()

	//handles `cf` | `cf -h` || `cf -help`
	if len(os.Args) == 1 || os.Args[1] == "--help" || os.Args[1] == "-help" ||
		os.Args[1] == "--h" || os.Args[1] == "-h" {
		help.ShowHelp(help.GetHelpTemplate())
		os.Exit(0)
	}

	//handle `cf -v` for cf version
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		deps.Ui.Say(os.Args[0] + " version " + cf.Version + "-" + cf.BuiltOnDate)
		os.Exit(0)
	}

	//handles `cf [COMMAND] -h ...`
	//rearrage args to `cf help COMMAND` and let `command help` to print out usage
	if requestHelp(os.Args[2:]) {
		os.Args[2] = os.Args[1]
		os.Args[1] = "help"
	}

	//run core command
	cmd := os.Args[1]

	if cmdRegistry.CommandExists(cmd) {
		meta := cmdRegistry.FindCommand(os.Args[1]).MetaData()
		fc := flags.NewFlagContext(meta.Flags)
		fc.SkipFlagParsing(meta.SkipFlagParsing)

		err := fc.Parse(os.Args[2:]...)
		if err != nil {
			deps.Ui.Failed("Incorrect Usage\n\n" + err.Error() + "\n\n" + cmdRegistry.CommandUsage(cmd))
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

	//non core command, try plugin command
	rpcService := newCliRpcServer(deps.TeePrinter, deps.TeePrinter)

	pluginsConfig := plugin_config.NewPluginConfig(func(err error) { deps.Ui.Failed(fmt.Sprintf("Error read/writing plugin config: %s, ", err.Error())) })
	pluginList := pluginsConfig.Plugins()

	ran := rpc.RunMethodIfExists(rpcService, os.Args[1:], pluginList)

	if !ran {
		deps.Ui.Say("'" + os.Args[1] + T("' is not a registered command. See 'cf help'"))
		os.Exit(1)
	}
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

func requestHelp(args []string) bool {
	for _, v := range args {
		if v == "-h" || v == "--help" || v == "--h" {
			return true
		}
	}

	return false
}

func newCliRpcServer(outputCapture terminal.OutputCapture, terminalOutputSwitch terminal.TerminalOutputSwitch) *rpc.CliRpcService {
	cliServer, err := rpc.NewRpcService(outputCapture, terminalOutputSwitch, deps.Config, deps.RepoLocator, rpc.NewNonCodegangstaRunner())
	if err != nil {
		deps.Ui.Say(T("Error initializing RPC service: ") + err.Error())
		os.Exit(1)
	}

	return cliServer
}
