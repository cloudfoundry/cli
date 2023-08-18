package command_parser

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
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var ErrFailed = errors.New("command failed")
var ParseErr = errors.New("incorrect type for arg")

type UI interface {
	DisplayError(err error)
	DisplayWarning(template string, templateValues ...map[string]interface{})
	DisplayText(template string, templateValues ...map[string]interface{})
	FlushDeferred()
}

type DisplayUsage interface {
	DisplayUsage()
}

type TriggerLegacyMain interface {
	LegacyMain()
	error
}

type CommandParser struct {
	Config *configv3.Config
	UI     *ui.UI
}

func NewCommandParser(config *configv3.Config) (CommandParser, error) {
	return CommandParser{Config: config}, nil
}

func (p *CommandParser) ParseCommandFromArgs(ui *ui.UI, args []string) (int, error) {
	p.UI = ui
	return p.parse(args, &common.Commands)
}

func (p *CommandParser) executionWrapper(cmd flags.Commander, args []string) error {
	cfConfig := p.Config
	cfConfig.Flags = configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	}
	defer p.UI.FlushDeferred()

	err := preventExtraArgs(args)
	if err != nil {
		return p.handleError(err)
	}

	err = cfConfig.CreatePluginHome()
	if err != nil {
		return p.handleError(err)
	}

	defer func() {
		configWriteErr := cfConfig.WriteConfig()
		if configWriteErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %s", configWriteErr.Error())
		}
	}()

	if extendedCmd, ok := cmd.(command.ExtendedCommander); ok {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.Level(cfConfig.LogLevel()))

		err = extendedCmd.Setup(cfConfig, p.UI)
		if err != nil {
			return p.handleError(err)
		}

		err = extendedCmd.Execute(args)
		return p.handleError(err)
	}

	return fmt.Errorf("command does not conform to ExtendedCommander")
}

func (p *CommandParser) handleError(passedErr error) error {
	if passedErr == nil {
		return nil
	}

	translatedErr := translatableerror.ConvertToTranslatableError(passedErr)

	switch typedErr := translatedErr.(type) {
	case translatableerror.V3V2SwitchError:
		log.Info("Received a V3V2SwitchError - switch to the V2 version of the command")
		return passedErr
	case TriggerLegacyMain:
		if typedErr.Error() != "" {
			p.UI.DisplayWarning("")
			p.UI.DisplayWarning(typedErr.Error())
		}
		cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	case *ssh.ExitError:
		exitStatus := typedErr.ExitStatus()
		if sig := typedErr.Signal(); sig != "" {
			p.UI.DisplayText("Process terminated by signal: {{.Signal}}. Exited with {{.ExitCode}}", map[string]interface{}{
				"Signal":   sig,
				"ExitCode": exitStatus,
			})
		}
		return passedErr
	}

	p.UI.DisplayError(translatedErr)

	if _, ok := translatedErr.(DisplayUsage); ok {
		return ParseErr
	}

	return ErrFailed
}

func (p *CommandParser) handleFlagErrorAndCommandHelp(flagErr *flags.Error, flagsParser *flags.Parser, extraArgs []string, originalArgs []string, commandList interface{}) (int, error) {
	switch flagErr.Type {
	case flags.ErrHelp, flags.ErrUnknownFlag, flags.ErrExpectedArgument, flags.ErrInvalidChoice:
		_, commandExists := reflect.TypeOf(common.Commands).FieldByNameFunc(
			func(fieldName string) bool {
				field, _ := reflect.TypeOf(common.Commands).FieldByName(fieldName)
				return flagsParser.Active != nil && flagsParser.Active.Name == field.Tag.Get("command")
			},
		)

		var helpExitCode int
		var err error
		if commandExists && flagErr.Type == flags.ErrUnknownFlag && (flagsParser.Active.Name == "set-env" || flagsParser.Active.Name == "v3-set-env") {
			newArgs := []string{}
			for _, arg := range originalArgs {
				if arg[0] == '-' {
					newArgs = append(newArgs, fmt.Sprintf("%s%s", flag.WorkAroundPrefix, arg))
				} else {
					newArgs = append(newArgs, arg)
				}
			}
			return p.parse(newArgs, commandList)
		}

		if flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument || flagErr.Type == flags.ErrInvalidChoice {
			fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
		}

		if commandExists {
			helpExitCode, err = p.parse([]string{"help", flagsParser.Active.Name}, commandList)
		} else {
			switch len(extraArgs) {
			case 0:
				helpExitCode, err = p.parse([]string{"help"}, commandList)
			case 1:
				if !isOption(extraArgs[0]) || (len(originalArgs) > 1 && extraArgs[0] == "-a") {
					helpExitCode, err = p.parse([]string{"help", extraArgs[0]}, commandList)
				} else {
					helpExitCode, err = p.parse([]string{"help"}, commandList)
				}
			default:
				if isCommand(extraArgs[0]) {
					helpExitCode, err = p.parse([]string{"help", extraArgs[0]}, commandList)
				} else {
					helpExitCode, err = p.parse(extraArgs[1:], commandList)
				}
			}
		}

		if flagErr.Type == flags.ErrHelp && helpExitCode == 0 {
			return 0, nil
		} else {
			return 1, err
		}

	case flags.ErrRequired, flags.ErrMarshal:
		fmt.Fprintf(os.Stderr, "Incorrect Usage: %s\n\n", flagErr.Error())
		_, err := p.parse([]string{"help", originalArgs[0]}, commandList)
		return 1, err

	case flags.ErrUnknownCommand:
		if containsHelpFlag(originalArgs) {
			return p.parse([]string{"help", originalArgs[0]}, commandList)
		} else {
			return 0, UnknownCommandError{CommandName: originalArgs[0]}
		}

	case flags.ErrCommandRequired:
		if common.Commands.VerboseOrVersion {
			return p.parse([]string{"version"}, commandList)
		} else {
			return p.parse([]string{"help"}, commandList)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unexpected flag error\ntype: %s\nmessage: %s\n", flagErr.Type, flagErr.Error())
	}

	return 0, nil
}

func (p *CommandParser) parse(args []string, commandList interface{}) (int, error) {
	flagsParser := flags.NewParser(commandList, flags.HelpFlag)
	flagsParser.CommandHandler = p.executionWrapper
	extraArgs, err := flagsParser.ParseArgs(args)
	if err == nil {
		return 0, nil
	} else if _, ok := err.(translatableerror.V3V2SwitchError); ok {
		return 1, err
	} else if flagErr, ok := err.(*flags.Error); ok {
		return p.handleFlagErrorAndCommandHelp(flagErr, flagsParser, extraArgs, args, commandList)
	} else if err == ErrFailed {
		return 1, nil
	} else if err == ParseErr {
		fmt.Println()
		p.parse([]string{"help", args[0]}, commandList) //nolint: errcheck
		return 1, err
	} else if exitError, ok := err.(*ssh.ExitError); ok {
		return exitError.ExitStatus(), nil
	}

	fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
	return 1, nil
}

func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "--h" {
			return true
		}
	}
	return false
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
