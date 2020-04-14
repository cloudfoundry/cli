package ccv3

type RouterGroup struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (client *Client) GetRouterGroups() ([]RouterGroup, Warnings, error) {
	var responseBody []RouterGroup

	_, warnings, err := client.MakeRequest(RequestParams{
		URL:          client.Routing() + "/v1/router_groups",
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
