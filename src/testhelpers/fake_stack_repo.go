package testhelpers

import (
	"cf"
	"cf/configuration"
)

type FakeStackRepository struct {
	FindByNameStack cf.Stack
	FindByNameName string
}

func (repo *FakeStackRepository) FindByName(config *configuration.Configuration, name string) (stack cf.Stack, err error) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

