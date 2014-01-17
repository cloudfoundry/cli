package api

import (
	"cf"
	"cf/net"
	"net/http"
)

type FakeAppSummaryRepo struct {
	GetSummariesInCurrentSpaceApps []cf.AppSummary

	GetSummaryErrorCode string
	GetSummaryAppGuid   string
	GetSummarySummary   cf.AppSummary
}

func (repo *FakeAppSummaryRepo) GetSummariesInCurrentSpace() (apps []cf.AppSummary, apiResponse net.ApiResponse) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *FakeAppSummaryRepo) GetSummary(appGuid string) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	repo.GetSummaryAppGuid = appGuid
	summary = repo.GetSummarySummary

	if repo.GetSummaryErrorCode != "" {
		apiResponse = net.NewApiResponse("Error", repo.GetSummaryErrorCode, http.StatusBadRequest)
	}

	return
}
