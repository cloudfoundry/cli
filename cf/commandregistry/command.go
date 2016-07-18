package commandregistry

import (
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/requirements"
)

//go:generate counterfeiter . Command

type Command interface {
	MetaData() CommandMetadata
	SetDependency(deps Dependency, pluginCall bool) Command
	Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) []requirements.Requirement
	Execute(context flags.FlagContext) error
}

type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           []string
	Description     string
	Flags           map[string]flags.FlagSet
	SkipFlagParsing bool
	TotalArgs       int //Optional: number of required arguments to skip for flag verification
	Hidden          bool
	Examples        []string
}
