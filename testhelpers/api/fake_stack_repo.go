package api

import "github.com/cloudfoundry/cli/cf/models"

type FakeStackRepository struct {
	FindByNameStack models.Stack
	FindByNameName  string

	FindAllStacks []models.Stack
}

func (repo *FakeStackRepository) FindByName(name string) (stack models.Stack, apiErr error) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []models.Stack, apiErr error) {
	stacks = repo.FindAllStacks
	return
}
