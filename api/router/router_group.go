package router

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/router/internal"
	"code.cloudfoundry.org/cli/api/router/routererror"
)

// RouterGroup represents a router group.
type RouterGroup struct {
	GUID            string `json:"guid"`
	Name            string `json:"name"`
	ReservablePorts string `json:"reservable_ports"`
	Type            string `json:"type"`
}

// GetRouterGroupByName returns a list of RouterGroups.
func (client *Client) GetRouterGroupByName(name string) (RouterGroup, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouterGroups,
		Query:       url.Values{"name": []string{name}},
	})

	if err != nil {
		return RouterGroup{}, err
	}
	var routerGroups []RouterGroup

	var response = Response{
		Result: &routerGroups,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return RouterGroup{}, err
	}

	for _, routerGroup := range routerGroups {
		if routerGroup.Name == name {
			return routerGroup, nil
		}
	}

	return RouterGroup{}, routererror.ResourceNotFoundError{}
}
