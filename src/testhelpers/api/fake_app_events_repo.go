package api

import "cf/models"

type FakeAppEventsRepo struct {
	AppGuid string
	Events  []models.EventFields
	Error   error
}

func (repo FakeAppEventsRepo) ListEvents(appGuid string, cb func(models.EventFields) bool) error {
	repo.AppGuid = appGuid
	for _, e := range repo.Events {
		cb(e)
	}
	return repo.Error
}
