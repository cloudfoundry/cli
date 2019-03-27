package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"encoding/json"
)

type Domain struct {
	GUID     string `json:"guid,omitempty"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
}

func (client Client) CreateDomain(domain Domain) (Domain, Warnings, error) {
	bodyBytes, err := json.Marshal(domain)
	if err != nil {
		return Domain{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Domain{}, nil, err
	}

	var ccDomain Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccDomain,
	}

	err = client.connection.Make(request, &response)

	return ccDomain, response.Warnings, err
}
