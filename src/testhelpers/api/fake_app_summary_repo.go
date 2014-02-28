package api

import (
	"cf/errors"
	"cf/models"
	"net/http"
)

type FakeAppSummaryRepo struct {
	GetSummariesInCurrentSpaceApps []models.AppSummary

	GetSummaryErrorCode string
	GetSummaryAppGuid   string
	GetSummarySummary   models.AppSummary
}

func (repo *FakeAppSummaryRepo) GetSummariesInCurrentSpace() (apps []models.AppSummary, apiResponse errors.Error) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *FakeAppSummaryRepo) GetSummary(appGuid string) (summary models.AppSummary, apiResponse errors.Error) {
	repo.GetSummaryAppGuid = appGuid
	summary = repo.GetSummarySummary

	if repo.GetSummaryErrorCode != "" {
		apiResponse = errors.NewError("Error", repo.GetSummaryErrorCode, http.StatusBadRequest)
	}

	return
}
