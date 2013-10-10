package api

import "cf/net"

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateUnauthorized bool
	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (score string, apiResponse net.ApiResponse){
	repo.ScoredPassword = password
	score = repo.Score
	return
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) (apiResponse net.ApiResponse) {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new

	if repo.UpdateUnauthorized {
		apiResponse = net.NewApiResponse("Authorization Failed", "unauthorized", 401)
	}

	return
}

