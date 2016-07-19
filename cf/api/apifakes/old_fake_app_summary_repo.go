package apifakes

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
)

type OldFakeAppSummaryRepo struct {
	GetSummariesInCurrentSpaceApps []models.Application

	GetSummaryErrorCode string
	GetSummaryAppGUID   string
	GetSummarySummary   models.Application
}

func (repo *OldFakeAppSummaryRepo) GetSummariesInCurrentSpace() (apps []models.Application, apiErr error) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *OldFakeAppSummaryRepo) GetSummary(appGUID string) (summary models.Application, apiErr error) {
	repo.GetSummaryAppGUID = appGUID
	summary = repo.GetSummarySummary

	if repo.GetSummaryErrorCode != "" {
		apiErr = errors.NewHTTPError(400, repo.GetSummaryErrorCode, "Error")
	}

	return
}
