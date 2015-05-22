package command_registry

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
)

var _ = initI18nFunc()
var Commands = NewRegistry()

func initI18nFunc() bool {
	T = Init(core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), nil), &detection.JibberJabberDetector{})
	return true
}

type registry struct {
	cmd map[string]Command
}

func NewRegistry() *registry {
	return &registry{
		cmd: make(map[string]Command),
	}
}

func Register(cmd Command) {
	m := cmd.MetaData()
	name := m.Name
	Commands.cmd[name] = cmd
}

func (r *registry) FindCommand(name string) Command {
	if _, ok := r.cmd[name]; ok {
		return r.cmd[name]
	}
	return nil
}

func (r *registry) CommandExists(name string) bool {
	_, ok := r.cmd[name]
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
	output += "   " + cmd.MetaData().Usage + "\n\n"

	if cmd.MetaData().ShortName != "" {
		output += T("ALIAS") + ":" + "\n"
		output += "   " + cmd.MetaData().ShortName + "\n\n"
	}

	if cmd.MetaData().Flags != nil {
		output += T("OPTIONS") + ":" + "\n"

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
