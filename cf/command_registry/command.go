package command_registry

import (
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

//go:generate counterfeiter -o fakes/fake_command.go . Command
type Command interface {
	MetaData() CommandMetadata
	SetDependency(deps Dependency, pluginCall bool) Command
	Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) (reqs []requirements.Requirement, err error)
	Execute(context flags.FlagContext)
}

type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           string
	Description     string
	Flags           map[string]flags.FlagSet
	SkipFlagParsing bool
	TotalArgs       int //Optional: number of required arguments to skip for flag verification
}
