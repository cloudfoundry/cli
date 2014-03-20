package api

import (
	"cf/errors"
	"cf/models"
)

type FakeAppSummaryRepo struct {
	GetSummariesInCurrentSpaceApps []models.AppSummary

	GetSummaryErrorCode string
	GetSummaryAppGuid   string
	GetSummarySummary   models.AppSummary
}

func (repo *FakeAppSummaryRepo) GetSummariesInCurrentSpace() (apps []models.AppSummary, apiErr error) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *FakeAppSummaryRepo) GetSummary(appGuid string) (summary models.AppSummary, apiErr error) {
	repo.GetSummaryAppGuid = appGuid
	summary = repo.GetSummarySummary

	if repo.GetSummaryErrorCode != "" {
		apiErr = errors.NewHttpError(400, "", "", repo.GetSummaryErrorCode, "Error")
	}

	return
}
