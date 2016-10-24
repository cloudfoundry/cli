package apifakes

import (
	"net/http"

	testnet "code.cloudfoundry.org/cli/utils/testhelpers/net"
)

func NewCloudControllerTestRequest(request testnet.TestRequest) testnet.TestRequest {
	request.Header = http.Header{
		"accept":        {"application/json"},
		"authorization": {"BEARER my_access_token"},
	}

	return request
}
