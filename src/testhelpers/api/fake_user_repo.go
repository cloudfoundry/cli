package api

import (
	"cf/net"
	"cf"
)

type FakeUserRepository struct {
	CreateUserUser cf.User
	CreateUserExists bool
}

func (repo *FakeUserRepository) Create(user cf.User) (apiResponse net.ApiResponse) {
	repo.CreateUserUser = user

	if repo.CreateUserExists {
		apiResponse = net.NewApiResponse("User already exists", cf.USER_EXISTS, 400)
	}

	return
}
