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
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/panichandler"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/jessevdk/go-flags"
)

type UI interface {
	DisplayError(err error)
}

var ErrFailed = errors.New("command failed")
var ParseErr = errors.New("incorrect type for arg")

func main() {
	defer panichandler.HandlePanic()
	parse(os.Args[1:])
}

func parse(args []string) {
	parser := flags.NewParser(&common.Commands, flags.HelpFlag)
	parser.CommandHandler = executionWrapper
	extraArgs, err := parser.ParseArgs(args)
	if err == nil {
		return
	}

	if flagErr, ok := err.(*flags.Error); ok {
		switch flagErr.Type {
		case flags.ErrHelp, flags.ErrUnknownFlag, flags.ErrExpectedArgument:
			_, found := reflect.TypeOf(common.Commands).FieldByNameFunc(
				func(fieldName string) bool {
					field, _ := reflect.TypeOf(common.Commands).FieldByName(fieldName)
					return parser.Active != nil && parser.Active.Name == field.Tag.Get("command")
				},
			)

			if found && flagErr.Type == flags.ErrUnknownFlag && parser.Active.Name == "set-env" {
				newArgs := []string{}
				for _, arg := range args {
					if arg[0] == '-' {
						newArgs = append(newArgs, fmt.Sprintf("%s%s", v2.WorkAroundPrefix, arg))
					} else {
						newArgs = append(newArgs, arg)
					}
				}
				parse(newArgs)
				return
			}

			if flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument {
				fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
			}

			if found {
				parse([]string{"help", parser.Active.Name})
			} else {
				switch len(extraArgs) {
				case 0:
					parse([]string{"help"})
				case 1:
					if !isOption(extraArgs[0]) || (len(args) > 1 && extraArgs[0] == "-a") {
						parse([]string{"help", extraArgs[0]})
					} else {
						parse([]string{"help"})
					}
				default:
					if isCommand(extraArgs[0]) {
						parse([]string{"help", extraArgs[0]})
					} else {
						parse(extraArgs[1:])
					}
				}
			}

			if flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument {
				os.Exit(1)
			}
		case flags.ErrRequired:
			fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
			parse([]string{"help", args[0]})
			os.Exit(1)
		case flags.ErrMarshal:
			errMessage := strings.Split(flagErr.Message, ":")
			fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", errMessage[0])
			parse([]string{"help", args[0]})
			os.Exit(1)
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
	} else if err == ErrFailed {
		os.Exit(1)
	} else if err == ParseErr {
		fmt.Println()
		parse([]string{"help", args[0]})
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}
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
	cfConfig, err := configv3.LoadConfig(configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	})
	if err != nil {
		return err
	}
	defer configv3.WriteConfig(cfConfig)

	if extendedCmd, ok := cmd.(command.ExtendedCommander); ok {
		commandUI, err := ui.NewUI(cfConfig)
		if err != nil {
			return err
		}

		err = extendedCmd.Setup(cfConfig, commandUI)
		if err != nil {
			return handleError(err, commandUI)
		}
		return handleError(extendedCmd.Execute(args), commandUI)
	}

	return fmt.Errorf("command does not conform to ExtendedCommander")
}

func handleError(err error, commandUI UI) error {
	if err == nil {
		return nil
	}

	commandUI.DisplayError(err)
	if _, isParseArgumentError := err.(command.ParseArgumentError); isParseArgumentError {
		return ParseErr
	}
	if _, isRequiredArgumentError := err.(command.RequiredArgumentError); isRequiredArgumentError {
		return ParseErr
	}

	return ErrFailed
}
