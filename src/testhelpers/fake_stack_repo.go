package testhelpers

import (
	"cf"
)

type FakeStackRepository struct {
	FindByNameStack cf.Stack
	FindByNameName string
}

func (repo *FakeStackRepository) FindByName(name string) (stack cf.Stack, err error) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

