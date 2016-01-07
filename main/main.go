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
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/rpc"
)

var deps = command_registry.NewDependency()

var cmdRegistry = command_registry.Commands

func main() {
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

	//handle `cf --build`
	if len(os.Args) == 2 && (os.Args[1] == "--build" || os.Args[1] == "-b") {
		deps.Ui.Say(T("{{.CFName}} was built with Go version: {{.GoVersion}}",
			map[string]interface{}{
				"CFName":    os.Args[0],
				"GoVersion": runtime.Version(),
			}))
		os.Exit(0)
	}

	//handles `cf [COMMAND] -h ...`
	//rearrage args to `cf help COMMAND` and let `command help` to print out usage
	if requestHelp(os.Args[2:]) {
		os.Args[2] = os.Args[1]
		os.Args[1] = "help"
	}

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
			deps.Ui.Failed("Incorrect Usage\n\n" + err.Error() + "\n\n" + usage)
		}

		cmd = cmd.SetDependency(deps, false)
		cmdRegistry.SetCommand(cmd)

		requirementsFactory := requirements.NewFactory(deps.Ui, deps.Config, deps.RepoLocator)
		reqs, err := cmd.Requirements(requirementsFactory, flagContext)
		if err != nil {
			deps.Ui.Failed(err.Error())
		}

		for _, req := range reqs {
			req.Execute()
		}

		cmd.Execute(flagContext)
		os.Exit(0)
	}

	//non core command, try plugin command
	rpcService := newCliRpcServer(deps.TeePrinter, deps.TeePrinter)

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
