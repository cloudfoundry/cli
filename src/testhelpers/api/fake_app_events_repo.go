package api

import "cf/models"

type FakeAppEventsRepo struct {
	RecentEventsArgs struct {
		AppGuid string
		Limit   uint
	}

	RecentEventsReturns struct {
		Events []models.EventFields
		Error  error
	}
}

func (repo *FakeAppEventsRepo) RecentEvents(appGuid string, limit uint) ([]models.EventFields, error) {
	repo.RecentEventsArgs.AppGuid = appGuid
	repo.RecentEventsArgs.Limit = limit
	return repo.RecentEventsReturns.Events, repo.RecentEventsReturns.Error
}
