package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeStackRepository struct {
	FindByNameStack cf.Stack
	FindByNameName string

	FindAllStacks []cf.Stack
}

func (repo *FakeStackRepository) FindByName(name string) (stack cf.Stack, apiErr *api.ApiError) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []cf.Stack, apiErr *api.ApiError) {
	return repo.FindAllStacks, nil
}

