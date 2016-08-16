package main

import (
	"fmt"
	"os"
	"reflect"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/utils/config"
	"code.cloudfoundry.org/cli/utils/panichandler"
	"github.com/jessevdk/go-flags"
)

func main() {
	defer panichandler.HandlePanic()
	parse(os.Args[1:])
}

func parse(args []string) {
	parser := flags.NewParser(&v2.Commands, flags.HelpFlag)
	parser.CommandHandler = myCommandHandler
	extraArgs, err := parser.ParseArgs(args)
	if err == nil {
		return
	}

	if flagErr, ok := err.(*flags.Error); ok {
		switch flagErr.Type {
		case flags.ErrHelp, flags.ErrUnknownFlag, flags.ErrExpectedArgument:
			_, found := reflect.TypeOf(v2.Commands).FieldByNameFunc(
				func(fieldName string) bool {
					field, _ := reflect.TypeOf(v2.Commands).FieldByName(fieldName)
					return parser.Active != nil && parser.Active.Name == field.Tag.Get("command")
				},
			)

			if found {
				parse([]string{"help", parser.Active.Name})
			} else {
				switch len(extraArgs) {
				case 0:
					parse([]string{"help"})
				case 1:
					if extraArgs[0] == "-v" || extraArgs[0] == "--version" {
						parse([]string{"version"})
					} else {
						parse([]string{"help", extraArgs[0]})
					}
				default:
					parse(extraArgs[1:])
				}
			}

			if flagErr.Type == flags.ErrUnknownFlag || flagErr.Type == flags.ErrExpectedArgument {
				os.Exit(1)
			}
		case flags.ErrRequired:
			fmt.Printf("%s\n\n", flagErr.Error())
			parse(append([]string{"help"}, args...))
			os.Exit(1)
		case flags.ErrUnknownCommand:
			cmd.Main(os.Getenv("CF_TRACE"), os.Args)
		case flags.ErrCommandRequired:
			parse([]string{"help"})
		default:
			fmt.Printf("unexpected flag error\ntype: %s\nmessage: %s\n", flagErr.Type, flagErr.Error())
		}
	} else {
		fmt.Println("unexpected error:", err.Error())
	}
}

func myCommandHandler(cmd flags.Commander, args []string) error {
	config, _ := config.LoadConfig()
	//defer write config
	if extendedCmd, ok := cmd.(commands.ExtendedCommander); ok {
		err := extendedCmd.Setup(config)
		if err != nil {
			return err
		}
		return extendedCmd.Execute(args)
	}

	return fmt.Errorf("unable to setup command")
}
