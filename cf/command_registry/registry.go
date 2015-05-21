package command_registry

import (
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
