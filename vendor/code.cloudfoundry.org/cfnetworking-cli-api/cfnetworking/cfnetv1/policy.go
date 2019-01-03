package cfnetv1

import (
	"bytes"
	"encoding/json"

	"strings"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1/internal"
)

type PolicyProtocol string

const (
	PolicyProtocolTCP PolicyProtocol = "tcp"
	PolicyProtocolUDP PolicyProtocol = "udp"
)

type PolicyList struct {
	TotalPolicies  int            `json:"total_policies,omitempty"`
	Policies       []Policy       `json:"policies,omitempty"`
	EgressPolicies []EgressPolicy `json:"egress_policies,omitempty"`
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

type EgressPolicy struct {
	Source      EgressPolicySource      `json:"source"`
	Destination EgressPolicyDestination `json:"destination"`
}

type EgressPolicySource struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

type EgressPolicyDestination struct {
	IPs      []IP           `json:"ips"`
	Protocol PolicyProtocol `json:"protocol"`
	Ports    []Ports        `json:"ports"`
}

type IP struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CreatePolicies will create the network policy with the given parameters.
func (client Client) CreatePolicies(policies PolicyList) error {
	rawJSON, err := json.Marshal(policies)
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

// ListPolicies will list the policies with the app guids in either the source or destination.
func (client Client) ListPolicies(appGUIDs ...string) (PolicyList, error) {
	var request *cfnetworking.Request
	var err error
	if len(appGUIDs) == 0 {
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.ListPolicies,
		})
	} else {
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.ListPolicies,
			Query: map[string][]string{
				"id": {strings.Join(appGUIDs, ",")},
			},
		})
	}
	if err != nil {
		return PolicyList{}, err
	}

	policies := PolicyList{}
	response := &cfnetworking.Response{}

	err = client.connection.Make(request, response)
	if err != nil {
		return PolicyList{}, err
	}

	err = json.Unmarshal(response.RawResponse, &policies)
	if err != nil {
		return PolicyList{}, err
	}

	return policies, nil
}

// RemovePolicies will remove the network policy with the given parameters.
func (client Client) RemovePolicies(policies []Policy) error {
	rawJSON, err := json.Marshal(PolicyList{Policies: policies})
	if err != nil {
		return err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeletePolicies,
		Body:        bytes.NewReader(rawJSON),
	})
	return client.connection.Make(request, &cfnetworking.Response{})
}
