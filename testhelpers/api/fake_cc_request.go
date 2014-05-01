package api

import (
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	"net/http"
)

func NewCloudControllerTestRequest(request testnet.TestRequest) testnet.TestRequest {
	request.Header = http.Header{
		"accept":        {"application/json"},
		"authorization": {"BEARER my_access_token"},
	}

	return request
}
