package api

import (
	"cf"
	"cf/net"
)

type FakeAppEventsRepo struct{
	AppGuid string
	Events []cf.EventFields
}


func (repo FakeAppEventsRepo)ListEvents(appGuid string) (events chan []cf.EventFields, statusChan chan net.ApiResponse) {
	repo.AppGuid = appGuid

	events = make(chan []cf.EventFields, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		for _, event := range repo.Events {
			events <- []cf.EventFields{event}
		}
		close(events)
		close(statusChan)
	}()

	return
}
