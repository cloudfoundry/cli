package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAppEventsRepo struct{
	Application cf.Application
	Events []cf.Event
}


func (repo FakeAppEventsRepo)ListEvents(app cf.Application) (events []cf.Event, apiResponse net.ApiResponse) {
	repo.Application = app
	events = repo.Events

	return
}
