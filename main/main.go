package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"path/filepath"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/configuration/confighelpers"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/panicprinter"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/commandsloader"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/cloudfoundry/cli/spellcheck"

	netrpc "net/rpc"
)

var cmdRegistry = commandregistry.Commands

func main() {
	traceEnv := os.Getenv("CF_TRACE")

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

	errFunc := func(err error) {
		if err != nil {
			ui := terminal.NewUI(
				os.Stdin,
				Writer,
				terminal.NewTeePrinter(Writer),
				trace.NewLogger(Writer, isVerbose, traceEnv, ""),
			)
			ui.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}

	// Only used to get Trace, so our errorHandler doesn't matter, since it's not used
	configPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		errFunc(err)
	}
	config := coreconfig.NewRepositoryFromFilepath(configPath, errFunc)
	defer config.Close()

	traceConfigVal := config.Trace()

	// Writer is assigned in writer_unix.go/writer_windows.go
	traceLogger := trace.NewLogger(Writer, isVerbose, traceEnv, traceConfigVal)

	deps := commandregistry.NewDependency(Writer, traceLogger)
	defer handlePanics(deps.TeePrinter, deps.Logger)
	defer deps.Config.Close()

	//handle `cf --build`
	if len(os.Args) == 2 && (os.Args[1] == "--build" || os.Args[1] == "-b") {
		deps.UI.Say(T("{{.CFName}} was built with Go version: {{.GoVersion}}",
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

	warningsCollector := net.NewWarningsCollector(deps.UI, warningProducers...)

	commandsloader.Load()

	//run core command
	cmdName := os.Args[1]
	cmd := cmdRegistry.FindCommand(cmdName)
	if cmd != nil {
		meta := cmd.MetaData()
		flagContext := flags.NewFlagContext(meta.Flags)
		flagContext.SkipFlagParsing(meta.SkipFlagParsing)

		cmdArgs := os.Args[2:]
		err = flagContext.Parse(cmdArgs...)
		if err != nil {
			usage := cmdRegistry.CommandUsage(cmdName)
			deps.UI.Failed(T("Incorrect Usage") + "\n\n" + err.Error() + "\n\n" + usage)
		}

		cmd = cmd.SetDependency(deps, false)
		cmdRegistry.SetCommand(cmd)

		requirementsFactory := requirements.NewFactory(deps.Config, deps.RepoLocator)
		reqs := cmd.Requirements(requirementsFactory, flagContext)

		for _, req := range reqs {
			err = req.Execute()
			if err != nil {
				deps.UI.Failed(err.Error())
			}
		}

		err = cmd.Execute(flagContext)
		if err != nil {
			ui := terminal.NewUI(os.Stdin, Writer, terminal.NewTeePrinter(Writer), traceLogger)
			ui.Failed(err.Error())
		}

		warningsCollector.PrintWarnings()

		os.Exit(0)
	}

	//non core command, try plugin command
	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpc.NewCommandRunner(), deps.Logger, Writer, server)
	if err != nil {
		deps.UI.Say(T("Error initializing RPC service: ") + err.Error())
		os.Exit(1)
	}

	pluginPath := filepath.Join(confighelpers.PluginRepoDir(), ".cf", "plugins")
	pluginConfig := pluginconfig.NewPluginConfig(
		func(err error) {
			deps.UI.Failed(fmt.Sprintf("Error read/writing plugin config: %s, ", err.Error()))
		},
		configuration.NewDiskPersistor(filepath.Join(pluginPath, "config.json")),
		pluginPath,
	)
	pluginList := pluginConfig.Plugins()

	ran := rpc.RunMethodIfExists(rpcService, os.Args[1:], pluginList)
	if !ran {
		deps.UI.Say("'" + os.Args[1] + T("' is not a registered command. See 'cf help'"))
		suggestCommands(cmdName, deps.UI, append(cmdRegistry.ListCommands(), pluginConfig.ListCommands()...))
		os.Exit(1)
	}
}

func suggestCommands(cmdName string, ui terminal.UI, cmdsList []string) {
	cmdSuggester := spellcheck.NewCommandSuggester(cmdsList)
	recommendedCmds := cmdSuggester.Recommend(cmdName)
	if len(recommendedCmds) != 0 {
		ui.Say("\n" + T("Did you mean?"))
		for _, suggestion := range recommendedCmds {
			ui.Say("      " + suggestion)
		}
	}
}

func handlePanics(printer terminal.Printer, logger trace.Printer) {
	panicPrinter := panicprinter.NewPanicPrinter(terminal.NewUI(os.Stdin, Writer, printer, logger))

	commandArgs := strings.Join(os.Args, " ")
	stackTrace := generateBacktrace()

	err := recover()
	panicPrinter.DisplayCrashDialog(err, commandArgs, stackTrace)

	if err != nil {
		os.Exit(1)
	}
}

func generateBacktrace() string {
	stackByteCount := 0
	stackSizeLimit := 1024 * 1024
	var bytes []byte
	for stackSize := 1024; (stackByteCount == 0 || stackByteCount == stackSize) && stackSize < stackSizeLimit; stackSize = 2 * stackSize {
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
		}
		return []string{"help", args[0]}
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
