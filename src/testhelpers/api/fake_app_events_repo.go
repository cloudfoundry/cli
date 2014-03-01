package api

import (
	"cf/errors"
	"cf/models"
)

type FakeAppEventsRepo struct {
	AppGuid string
	Events  []models.EventFields
	Error   errors.Error
}

func (repo FakeAppEventsRepo) ListEvents(appGuid string, cb func(models.EventFields) bool) errors.Error {
	repo.AppGuid = appGuid
	for _, e := range repo.Events {
		cb(e)
	}
	return repo.Error
}
