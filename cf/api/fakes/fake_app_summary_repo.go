package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeAppSummaryRepo struct {
	GetSummariesInCurrentSpaceApps []models.Application

	GetSummaryErrorCode   string
	GetSummaryAppGuid     string
	GetSpaceGuid          string
	GetSummarySummary     models.Application
	GetSpaceSummariesApps []models.Application
}

func (repo *FakeAppSummaryRepo) GetSummariesInCurrentSpace() (apps []models.Application, apiErr error) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *FakeAppSummaryRepo) GetSummary(appGuid string) (summary models.Application, apiErr error) {
	repo.GetSummaryAppGuid = appGuid
	summary = repo.GetSummarySummary

	if repo.GetSummaryErrorCode != "" {
		apiErr = errors.NewHttpError(400, repo.GetSummaryErrorCode, "Error")
	}

	return
}

func (repo *FakeAppSummaryRepo) GetSpaceSummaries(spaceGuid string) (apps []models.Application, apiErr error) {
	repo.GetSpaceGuid = spaceGuid
	apps = repo.GetSpaceSummariesApps
	return
}
