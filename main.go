package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/panichandler"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type UI interface {
	DisplayError(err error)
	DisplayWarning(template string, templateValues ...map[string]interface{})
	DisplayText(template string, templateValues ...map[string]interface{})
}

type DisplayUsage interface {
	DisplayUsage()
}

type TriggerLegacyMain interface {
	LegacyMain()
	error
}

var ErrFailed = errors.New("command failed")
var ParseErr = errors.New("incorrect type for arg")

func main() {
	defer panichandler.HandlePanic()
	exitStatus := parse(os.Args[1:])
	if exitStatus != 0 {
		os.Exit(exitStatus)
	}
}

func parse(args []string) int {
	parser := flags.NewParser(&common.Commands, flags.HelpFlag)
	parser.CommandHandler = executionWrapper
	extraArgs, err := parser.ParseArgs(args)
	if err == nil {
		return 0
	}

	if flagErr, ok := err.(*flags.Error); ok {
		return handleFlagErrorAndCommandHelp(flagErr, parser, extraArgs, args)
	} else if err == ErrFailed {
		return 1
	} else if err == ParseErr {
		fmt.Println()
		parse([]string{"help", args[0]})
		return 1
	} else if exitError, ok := err.(*ssh.ExitError); ok {
		return exitError.ExitStatus()
	}

	fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
	return 1
}

func handleFlagErrorAndCommandHelp(flagErr *flags.Error, parser *flags.Parser, extraArgs []string, originalArgs []string) int {
	switch flagErr.Type {
	case flags.ErrHelp, flags.ErrUnknownFlag, flags.ErrExpectedArgument, flags.ErrInvalidChoice:
		_, found := reflect.TypeOf(common.Commands).FieldByNameFunc(
			func(fieldName string) bool {
				field, _ := reflect.TypeOf(common.Commands).FieldByName(fieldName)
				return parser.Active != nil && parser.Active.Name == field.Tag.Get("command")
			},
		)

		if found && flagErr.Type == flags.ErrUnknownFlag && (parser.Active.Name == "set-env" || parser.Active.Name == "v3-set-env") {
			newArgs := []string{}
			for _, arg := range originalArgs {
				if arg[0] == '-' {
					newArgs = append(newArgs, fmt.Sprintf("%s%s", flag.WorkAroundPrefix, arg))
				} else {
					newArgs = append(newArgs, arg)
				}
			}
			parse(newArgs)
			return 0
		}

		if flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument || flagErr.Type == flags.ErrInvalidChoice {
			fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
		}

		var helpErrored int
		if found {
			helpErrored = parse([]string{"help", parser.Active.Name})
		} else {
			switch len(extraArgs) {
			case 0:
				helpErrored = parse([]string{"help"})
			case 1:
				if !isOption(extraArgs[0]) || (len(originalArgs) > 1 && extraArgs[0] == "-a") {
					helpErrored = parse([]string{"help", extraArgs[0]})
				} else {
					helpErrored = parse([]string{"help"})
				}
			default:
				if isCommand(extraArgs[0]) {
					helpErrored = parse([]string{"help", extraArgs[0]})
				} else {
					helpErrored = parse(extraArgs[1:])
				}
			}
		}

		if helpErrored > 0 || flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument || flagErr.Type == flags.ErrInvalidChoice {
			return 1
		}
	case flags.ErrRequired:
		fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
		parse([]string{"help", originalArgs[0]})
		return 1
	case flags.ErrMarshal:
		errMessage := strings.Split(flagErr.Message, ":")
		fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", errMessage[0])
		parse([]string{"help", originalArgs[0]})
		return 1
	case flags.ErrUnknownCommand:
		cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	case flags.ErrCommandRequired:
		if common.Commands.VerboseOrVersion {
			parse([]string{"version"})
		} else {
			parse([]string{"help"})
		}
	default:
		fmt.Fprintf(os.Stderr, "Unexpected flag error\ntype: %s\nmessage: %s\n", flagErr.Type, flagErr.Error())
	}
	return 0
}

func isCommand(s string) bool {
	_, found := reflect.TypeOf(common.Commands).FieldByNameFunc(
		func(fieldName string) bool {
			field, _ := reflect.TypeOf(common.Commands).FieldByName(fieldName)
			return s == field.Tag.Get("command") || s == field.Tag.Get("alias")
		})

	return found
}

func isOption(s string) bool {
	return strings.HasPrefix(s, "-")
}

func executionWrapper(cmd flags.Commander, args []string) error {
	cfConfig, configErr := configv3.LoadConfig(configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	})
	if configErr != nil {
		if _, ok := configErr.(translatableerror.EmptyConfigError); !ok {
			return configErr
		}
	}

	commandUI, err := ui.NewUI(cfConfig)
	if err != nil {
		return err
	}

	// TODO: when the line in the old code under `cf` which calls
	// configv3.LoadConfig() is finally removed, then we should replace the code
	// path above with the following:
	//
	// var configErrTemplate string
	// if configErr != nil {
	// 	if ce, ok := configErr.(translatableerror.EmptyConfigError); ok {
	// 		configErrTemplate = ce.Error()
	// 	} else {
	// 		return configErr
	// 	}
	// }

	// commandUI, err := ui.NewUI(cfConfig)
	// if err != nil {
	// 	return err
	// }

	// if configErr != nil {
	//   commandUI.DisplayWarning(configErrTemplate, map[string]interface{}{
	//   	"FilePath": configv3.ConfigFilePath(),
	//   })
	// }

	defer func() {
		configWriteErr := configv3.WriteConfig(cfConfig)
		if configWriteErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %s", configWriteErr.Error())
		}
	}()

	if extendedCmd, ok := cmd.(command.ExtendedCommander); ok {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Level(cfConfig.LogLevel()))

		err = extendedCmd.Setup(cfConfig, commandUI)
		if err != nil {
			return handleError(err, commandUI)
		}
		return handleError(extendedCmd.Execute(args), commandUI)
	}

	return fmt.Errorf("command does not conform to ExtendedCommander")
}

func handleError(passedErr error, commandUI UI) error {
	if passedErr == nil {
		return nil
	}

	translatedErr := translatableerror.ConvertToTranslatableError(passedErr)

	switch typedErr := translatedErr.(type) {
	case TriggerLegacyMain:
		if typedErr.Error() != "" {
			commandUI.DisplayWarning("")
			commandUI.DisplayWarning(typedErr.Error())
		}

		// TODO: Replace this section with os.Args when v2-push has been replaced
		var args []string
		for _, arg := range os.Args {
			if arg == "v2-push" {
				args = append(args, "push")
			} else {
				args = append(args, arg)
			}
		}
		cmd.Main(os.Getenv("CF_TRACE"), args)
	case *ssh.ExitError:
		exitStatus := typedErr.ExitStatus()
		if sig := typedErr.Signal(); sig != "" {
			commandUI.DisplayText("Process terminated by signal: {{.Signal}}. Exited with {{.ExitCode}}", map[string]interface{}{
				"Signal":   sig,
				"ExitCode": exitStatus,
			})
		}
		return passedErr
	}

	commandUI.DisplayError(translatedErr)

	if _, ok := translatedErr.(DisplayUsage); ok {
		return ParseErr
	}

	return ErrFailed
}
