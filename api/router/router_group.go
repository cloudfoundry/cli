package router

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/router/internal"
)

// RouterGroup represents router group
type RouterGroup struct {
	GUID            string `json:"guid"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	ReservablePorts string `json:"reservable_ports"`
}

// GetRouterGroupsByName returns a list of RouterGroups
func (client *Client) GetRouterGroupsByName(name string) ([]RouterGroup, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouterGroups,
		Query:       url.Values{"name": []string{name}},
	})

	if err != nil {
		return nil, err
	}
	var fullRouterGroupList []RouterGroup

	var response = Response{
		Result: &fullRouterGroupList,
	}
	err = client.connection.Make(request, &response)
	return fullRouterGroupList, err
}
