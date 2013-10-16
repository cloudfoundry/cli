package api

import (
	"cf"
	"cf/net"
)

type FakeAppSummaryRepo struct{
	GetSummariesInCurrentSpaceApps []cf.Application

	GetSummaryApp cf.Application
	GetSummarySummary cf.AppSummary
}

func (repo *FakeAppSummaryRepo)GetSummariesInCurrentSpace() (apps []cf.Application, apiResponse net.ApiResponse) {
	apps = repo.GetSummariesInCurrentSpaceApps
	return
}

func (repo *FakeAppSummaryRepo)GetSummary(app cf.Application) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	repo.GetSummaryApp= app
	summary = repo.GetSummarySummary

	return
}
