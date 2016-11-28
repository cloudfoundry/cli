package commandregistry

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/util/configv3"

	. "code.cloudfoundry.org/cli/cf/terminal"
)

var _ = initI18nFunc()
var Commands = NewRegistry()

func initI18nFunc() bool {
	config, err := configv3.LoadConfig()
	if err != nil {
		fmt.Println(FailureColor("FAILED"))
		fmt.Println("Error read/writing config: ", err.Error())
		os.Exit(1)
	}
	T = Init(config)
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

func (r *registry) ListCommands() []string {
	keys := make([]string, len(r.cmd))

	i := 0
	for k := range r.cmd {
		keys[i] = k
		i++
	}

	return keys
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
	cmd := r.FindCommand(cmdName)

	return CLICommandUsagePresenter(cmd).Usage()
}

type usagePresenter struct {
	cmd Command
}

func CLICommandUsagePresenter(cmd Command) *usagePresenter {
	return &usagePresenter{
		cmd: cmd,
	}
}

func (u *usagePresenter) Usage() string {
	metadata := u.cmd.MetaData()

	output := T("NAME:") + "\n"
	output += "   " + metadata.Name + " - " + metadata.Description + "\n\n"

	output += T("USAGE:") + "\n"
	output += "   " + strings.Replace(strings.Join(metadata.Usage, ""), "CF_NAME", cf.Name, -1) + "\n"

	if len(metadata.Examples) > 0 {
		output += "\n"
		output += fmt.Sprintf("%s:\n", T("EXAMPLES"))
		for _, e := range metadata.Examples {
			output += fmt.Sprintf("   %s\n", strings.Replace(e, "CF_NAME", cf.Name, -1))
		}
	}

	if metadata.ShortName != "" {
		output += "\n" + T("ALIAS:") + "\n"
		output += "   " + metadata.ShortName + "\n"
	}

	if metadata.Flags != nil {
		output += "\n" + T("OPTIONS:") + "\n"
		output += flags.NewFlagContext(metadata.Flags).ShowUsage(3)
	}

	return output
}
