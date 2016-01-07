package command_registry

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

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
	i18n.T = i18n.Init(core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), errorHandler))
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
	if strings.TrimSpace(name) == "" {
		return false
	}

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

func (r *registry) RemoveCommand(cmdName string) {
	delete(r.cmd, cmdName)
}

func (r *registry) TotalCommands() int {
	return len(r.cmd)
}

func (r *registry) MaxCommandNameLength() int {
	maxNameLen := 0
	for name := range r.cmd {
		if nameLen := utf8.RuneCountInString(name); nameLen > maxNameLen {
			maxNameLen = nameLen
		}
	}
	return maxNameLen
}

func (r *registry) Metadatas() []CommandMetadata {
	var m []CommandMetadata

	for _, c := range r.cmd {
		m = append(m, c.MetaData())
	}

	return m
}

func (r *registry) CommandUsage(cmdName string) string {
	output := ""
	cmd := r.FindCommand(cmdName)

	output = i18n.T("NAME") + ":" + "\n"
	output += "   " + cmd.MetaData().Name + " - " + cmd.MetaData().Description + "\n\n"

	output += i18n.T("USAGE") + ":" + "\n"
	output += "   " + strings.Replace(cmd.MetaData().Usage, "CF_NAME", cf.Name(), -1) + "\n"

	if cmd.MetaData().ShortName != "" {
		output += "\n" + i18n.T("ALIAS") + ":" + "\n"
		output += "   " + cmd.MetaData().ShortName + "\n"
	}

	if cmd.MetaData().Flags != nil {
		output += "\n" + i18n.T("OPTIONS") + ":" + "\n"
		output += flags.NewFlagContext(cmd.MetaData().Flags).ShowUsage(3)
	}

	return output
}
