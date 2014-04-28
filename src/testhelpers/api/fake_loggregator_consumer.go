package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type FakeLoggregatorConsumer struct {
	RecentCalledWith struct {
		AppGuid   string
		AuthToken string
	}

	RecentReturns struct {
		Messages  []*logmessage.LogMessage
		Err       []error
		callIndex int
	}

	TailFunc func(appGuid, token string) (<-chan *logmessage.LogMessage, error)

	IsClosed bool

	OnConnectCallback func()

	closeChan chan bool
}

func NewFakeLoggregatorConsumer() *FakeLoggregatorConsumer {
	return &FakeLoggregatorConsumer{
		closeChan: make(chan bool, 1),
	}
}

func (c *FakeLoggregatorConsumer) Recent(appGuid string, authToken string) ([]*logmessage.LogMessage, error) {
	c.RecentCalledWith.AppGuid = appGuid
	c.RecentCalledWith.AuthToken = authToken

	var err error
	if c.RecentReturns.callIndex < len(c.RecentReturns.Err) {
		err = c.RecentReturns.Err[c.RecentReturns.callIndex]
		c.RecentReturns.callIndex++
	}

	return c.RecentReturns.Messages, err
}

func (c *FakeLoggregatorConsumer) Close() error {
	c.IsClosed = true
	c.closeChan <- true
	return nil
}

func (c *FakeLoggregatorConsumer) SetOnConnectCallback(cb func()) {
	c.OnConnectCallback = cb
}

func (c *FakeLoggregatorConsumer) Tail(appGuid string, authToken string) (<-chan *logmessage.LogMessage, error) {
	return c.TailFunc(appGuid, authToken)
}

func (c *FakeLoggregatorConsumer) WaitForClose() {
	<-c.closeChan
}
