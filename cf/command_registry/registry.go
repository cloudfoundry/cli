package command_registry

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
	. "github.com/cloudfoundry/cli/cf/terminal"
)

var _ = initI18nFunc()
var Commands = NewRegistry()

func initI18nFunc() bool {
	errorHandler := func(err error) {
		if err != nil {
			fmt.Println(FailureColor("FAILED"))
			fmt.Println("Error read/writing config: ", err.Error())
			os.Exit(1)
		}
	}
	T = Init(core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), errorHandler), &detection.JibberJabberDetector{})
	return true
}

type registry struct {
	cmd   map[string]Command
	alias map[string]string
}

func NewRegistry() *registry {
	return &registry{
		cmd:   make(map[string]Command),
		alias: make(map[string]string),
	}
}

func Register(cmd Command) {
	m := cmd.MetaData()
	Commands.cmd[m.Name] = cmd

	Commands.alias[m.ShortName] = m.Name
}

func (r *registry) FindCommand(name string) Command {
	if _, ok := r.cmd[name]; ok {
		return r.cmd[name]
	}

	if alias, exists := r.alias[name]; exists {
		return r.cmd[alias]
	}

	return nil
}

func (r *registry) CommandExists(name string) bool {
	var ok bool

	if _, ok = r.cmd[name]; !ok {
		alias, exists := r.alias[name]

		if exists {
			_, ok = r.cmd[alias]
		}
	}

	return ok
}

func (r *registry) SetCommand(cmd Command) {
	r.cmd[cmd.MetaData().Name] = cmd
}

func (r *registry) CommandUsage(cmdName string) string {
	output := ""
	cmd := r.FindCommand(cmdName)

	output = T("NAME") + ":" + "\n"
	output += "   " + cmd.MetaData().Name + " - " + cmd.MetaData().Description + "\n\n"

	output += T("USAGE") + ":" + "\n"
	output += "   " + cmd.MetaData().Usage + "\n"

	if cmd.MetaData().ShortName != "" {
		output += "\n" + T("ALIAS") + ":" + "\n"
		output += "   " + cmd.MetaData().ShortName + "\n"
	}

	if cmd.MetaData().Flags != nil {
		output += "\n" + T("OPTIONS") + ":" + "\n"

		//find longest name length
		l := 0
		for n, _ := range cmd.MetaData().Flags {
			if len(n) > l {
				l = len(n)
			}
		}

		//print non-bool flags first
		for n, f := range cmd.MetaData().Flags {
			switch f.GetValue().(type) {
			case bool:
			default:
				output += "   -" + n + strings.Repeat(" ", 7+(l-len(n))) + f.String() + "\n"
			}
		}

		//then bool flags
		for n, f := range cmd.MetaData().Flags {
			switch f.GetValue().(type) {
			case bool:
				output += "   --" + n + strings.Repeat(" ", 6+(l-len(n))) + f.String() + "\n"
			}
		}
	}

	return output
}
