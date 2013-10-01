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

func (repo *FakeStackRepository) FindByName(name string) (stack cf.Stack, apiStatus net.ApiStatus) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []cf.Stack, apiStatus net.ApiStatus) {
	stacks = repo.FindAllStacks
	return
}

