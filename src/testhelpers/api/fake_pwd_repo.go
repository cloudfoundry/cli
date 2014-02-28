package api

import "cf/errors"

type FakePasswordRepo struct {
	Score          string
	ScoredPassword string

	UpdateUnauthorized bool
	UpdateNewPassword  string
	UpdateOldPassword  string
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) (apiResponse errors.Error) {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new

	if repo.UpdateUnauthorized {
		apiResponse = errors.NewError("Authorization Failed", "unauthorized", 401)
	}

	return
}
