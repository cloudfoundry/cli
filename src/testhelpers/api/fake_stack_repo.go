package api

import (
	"cf/models"
	"cf/net"
)

type FakeStackRepository struct {
	FindByNameStack models.Stack
	FindByNameName  string

	FindAllStacks []models.Stack
}

func (repo *FakeStackRepository) FindByName(name string) (stack models.Stack, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	stack = repo.FindByNameStack

	return
}

func (repo *FakeStackRepository) FindAll() (stacks []models.Stack, apiResponse net.ApiResponse) {
	stacks = repo.FindAllStacks
	return
}
