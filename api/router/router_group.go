package router

import (
	"net/url"

	"code.cloudfoundry.org/cli/v8/api/router/internal"
	"code.cloudfoundry.org/cli/v8/api/router/routererror"
)

// RouterGroup represents a router group.
type RouterGroup struct {
	GUID            string `json:"guid"`
	Name            string `json:"name"`
	ReservablePorts string `json:"reservable_ports"`
	Type            string `json:"type"`
}

func (client *Client) GetRouterGroups() ([]RouterGroup, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouterGroups,
	})

	if err != nil {
		return nil, err
	}
	var routerGroups []RouterGroup

	var response = Response{
		Result: &routerGroups,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return nil, err
	}

	return routerGroups, nil
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
