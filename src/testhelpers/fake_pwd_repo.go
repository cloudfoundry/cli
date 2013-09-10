package testhelpers

import "cf/api"

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (string, *api.ApiError){
	repo.ScoredPassword = password
	return repo.Score, nil
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) *api.ApiError {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new
	return nil
}

