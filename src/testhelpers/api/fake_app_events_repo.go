package api

import (
	"cf/models"
	"cf/net"
)

type FakeAppEventsRepo struct {
	AppGuid     string
	Events      []models.EventFields
	ApiResponse net.ApiResponse
}

func (repo FakeAppEventsRepo) ListEvents(appGuid string, cb func(models.EventFields) bool) net.ApiResponse {
	repo.AppGuid = appGuid
	for _, e := range repo.Events {
		cb(e)
	}
	return repo.ApiResponse
}
