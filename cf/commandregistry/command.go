package commandregistry

import (
	"code.cloudfoundry.org/cli/v9/cf/flags"
	"code.cloudfoundry.org/cli/v9/cf/requirements"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Command

type Command interface {
	MetaData() CommandMetadata
	SetDependency(deps Dependency, pluginCall bool) Command
	Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) ([]requirements.Requirement, error)
	Execute(context flags.FlagContext) error
}

type CommandMetadata struct {
	Name            string
	ShortName       string
	Usage           []string
	Description     string
	Flags           map[string]flags.FlagSet
	SkipFlagParsing bool
	TotalArgs       int // Optional: number of required arguments to skip for flag verification
	Hidden          bool
	Examples        []string
}
