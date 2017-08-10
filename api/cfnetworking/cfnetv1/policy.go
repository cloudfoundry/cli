package cfnetv1

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cfnetworking"
	"code.cloudfoundry.org/cli/api/cfnetworking/cfnetv1/internal"
)

type PolicyProtocol string

const (
	PolicyProtocolTCP PolicyProtocol = "tcp"
	PolicyProtocolUDP PolicyProtocol = "udp"
)

type PolicyList struct {
	TotalPolicies int      `json:"total_policies,omitempty"`
	Policies      []Policy `json:"policies"`
}

type Policy struct {
	Source      PolicySource      `json:"source"`
	Destination PolicyDestination `json:"destination"`
}

type PolicySource struct {
	ID string `json:"id"`
}

type PolicyDestination struct {
	ID       string         `json:"id"`
	Protocol PolicyProtocol `json:"protocol"`
	Ports    Ports          `json:"ports"`
}

// CreatePolicies will create the network policy with the given parameters.
func (client Client) CreatePolicies(policies []Policy) error {
	rawJSON, err := json.Marshal(PolicyList{Policies: policies})
	if err != nil {
		return err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.CreatePolicies,
		Body:        bytes.NewReader(rawJSON),
	})
	if err != nil {
		return err
	}

	return client.connection.Make(request, &cfnetworking.Response{})
}
