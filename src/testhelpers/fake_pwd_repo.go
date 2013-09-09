package testhelpers

type FakePasswordRepo struct {
	Score string
	ScoredPassword string

	UpdateNewPassword string
	UpdateOldPassword string
}



func (repo *FakePasswordRepo) GetScore(password string) (string, error){
	repo.ScoredPassword = password
	return repo.Score, nil
}

func (repo *FakePasswordRepo) UpdatePassword(old string, new string) error {
	repo.UpdateOldPassword = old
	repo.UpdateNewPassword = new
	return nil
}

