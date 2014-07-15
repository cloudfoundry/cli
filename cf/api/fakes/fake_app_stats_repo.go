package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeAppStatsRepo struct {
	GetStatsAppGuid    string
	GetStatsResponses  [][]models.AppStatsFields
	GetStatsErrorCodes []string
}

func (repo *FakeAppStatsRepo) GetStats(appGuid string) (stats []models.AppStatsFields, apiErr error) {
	repo.GetStatsAppGuid = appGuid

	if len(repo.GetStatsResponses) > 0 {
		stats = repo.GetStatsResponses[0]

		if len(repo.GetStatsResponses) > 1 {
			repo.GetStatsResponses = repo.GetStatsResponses[1:]
		}
	}

	if len(repo.GetStatsErrorCodes) > 0 {
		errorCode := repo.GetStatsErrorCodes[0]

		// don't slice away the last one if this is all we have
		if len(repo.GetStatsErrorCodes) > 1 {
			repo.GetStatsErrorCodes = repo.GetStatsErrorCodes[1:]
		}

		if errorCode != "" {
			apiErr = errors.NewHttpError(400, errorCode, "Error staging app")
		}
	}

	return
}
