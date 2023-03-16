package uaa

import (
	"fmt"
	"net/http"
)

func (client *Client) GetAPIVersion() (string, error) {
	type loginResponse struct {
		App struct {
			Version string `json:"version"`
		} `json:"app"`
	}

	request, err := client.newRequest(requestOptions{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("%s/login", client.LoginLink()),
	})

	if err != nil {
		return "", err
	}

	info := loginResponse{}
	response := Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return "", err
	}

	return info.App.Version, nil
}
