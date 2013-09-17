package testhelpers

import "cf/net"

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateUnauthorized bool
	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (string, *net.ApiError){
	repo.ScoredPassword = password
	return repo.Score, nil
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) (apiErr *net.ApiError) {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new

	if repo.UpdateUnauthorized {
		apiErr = net.NewApiError("Authorization Failed", "unauthorized", 401)
	}

	return
}

