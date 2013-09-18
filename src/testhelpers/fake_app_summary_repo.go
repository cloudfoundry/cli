package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAppSummaryRepo struct{
	GetSummaryApp cf.Application
	GetSummarySummary cf.AppSummary
}


func (repo *FakeAppSummaryRepo)GetSummary(app cf.Application) (summary cf.AppSummary, apiErr *net.ApiError) {
	repo.GetSummaryApp= app
	summary = repo.GetSummarySummary

	return
}
