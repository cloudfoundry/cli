package requirements

import (
	"cf/api"
	"cf/terminal"
)

type Factory struct {
	ui                terminal.UI
	repositoryLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, repositoryLocator api.RepositoryLocator) (factory Factory) {
	factory.ui = ui
	factory.repositoryLocator = repositoryLocator
	return
}
