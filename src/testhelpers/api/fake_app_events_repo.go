package api

import (
	"cf/models"
	"cf/net"
)

type FakeAppEventsRepo struct {
	AppGuid string
	Events  []models.EventFields
	ApiResponse net.ApiResponse
}

func (repo FakeAppEventsRepo) ListEvents(appGuid string, cb func([]models.EventFields) bool) net.ApiResponse {
	repo.AppGuid = appGuid
	if len(repo.Events) > 0 {
		cb(repo.Events)
	}
	return repo.ApiResponse
}
