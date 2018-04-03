package director

import (
	"time"

	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Client struct {
	clientRequest     ClientRequest
	taskClientRequest TaskClientRequest
}

func NewClient(
	endpoint string,
	httpClient *httpclient.HTTPClient,
	taskReporter TaskReporter,
	fileReporter FileReporter,
	logger boshlog.Logger,
) Client {
	clientRequest := NewClientRequest(endpoint, httpClient, fileReporter, logger)
	taskClientRequest := NewTaskClientRequest(clientRequest, taskReporter, 500*time.Millisecond)
	return Client{clientRequest, taskClientRequest}
}

func (c Client) WithContext(contextId string) Client {
	clientRequest := c.clientRequest.WithContext(contextId)

	taskClientRequest := c.taskClientRequest
	taskClientRequest.clientRequest = clientRequest

	return Client{clientRequest, taskClientRequest}
}
