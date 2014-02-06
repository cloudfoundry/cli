package api

import (
	"cf/models"
	"cf/net"
	"cf/api"
)

type FakeAppEventsRepo struct {
	AppGuid string
	Events  []models.EventFields
	ApiResponse net.ApiResponse
}

func (repo FakeAppEventsRepo) ListEvents(appGuid string, cb api.ListEventsCallback) net.ApiResponse {
	repo.AppGuid = appGuid
	if len(repo.Events) > 0 {
		cb(repo.Events)
	}
	return repo.ApiResponse
}
