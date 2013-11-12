package api

import (
	"cf"
	"cf/net"
)

type FakeAppEventsRepo struct{
	Application cf.Application
	Events []cf.Event
}


func (repo FakeAppEventsRepo)ListEvents(app cf.Application) (events chan []cf.Event, statusChan chan net.ApiResponse) {
	repo.Application = app

	events = make(chan []cf.Event, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		for _, event := range repo.Events {
			events <- []cf.Event{event}
		}
		close(events)
		close(statusChan)
	}()

	return
}
