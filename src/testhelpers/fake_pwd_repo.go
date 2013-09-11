package testhelpers

import "cf/api"

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateUnauthorized bool
	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (string, *api.ApiError){
	repo.ScoredPassword = password
	return repo.Score, nil
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) (apiErr *api.ApiError) {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new

	if repo.UpdateUnauthorized {
		apiErr = api.NewApiError("Authorization Failed", "unauthorized", 401)
	}

	return
}

