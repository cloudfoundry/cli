package api

import (
	"cf"
	"cf/net"
)

type FakeAppSummaryRepo struct{
	GetSummaryApp cf.Application
	GetSummarySummary cf.AppSummary
}


func (repo *FakeAppSummaryRepo)GetSummary(app cf.Application) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	repo.GetSummaryApp= app
	summary = repo.GetSummarySummary

	return
}
