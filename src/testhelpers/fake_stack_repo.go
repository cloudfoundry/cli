package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeStackRepository struct {
	FindByNameStack cf.Stack
	FindByNameName string

	FindAllStacks []cf.Stack
}

func (repo *FakeStackRepository) FindByName(name string) (stack cf.Stack, apiErr *net.ApiError) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []cf.Stack, apiErr *net.ApiError) {
	return repo.FindAllStacks, nil
}

