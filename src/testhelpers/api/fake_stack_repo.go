package api

import (
	"cf/errors"
	"cf/models"
)

type FakeStackRepository struct {
	FindByNameStack models.Stack
	FindByNameName  string

	FindAllStacks []models.Stack
}

func (repo *FakeStackRepository) FindByName(name string) (stack models.Stack, apiErr errors.Error) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []models.Stack, apiErr errors.Error) {
	stacks = repo.FindAllStacks
	return
}
