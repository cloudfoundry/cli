package uaa

import (
	"fmt"
	"net/http"
)

func (client *Client) GetLoginPrompts() (map[string][]string, error) {
	type loginResponse struct {
		Prompts map[string][]string `json:"prompts"`
	}

	request, err := client.newRequest(requestOptions{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("%s/login", client.LoginLink()),
	})
	if err != nil {
		return nil, err
	}

	info := loginResponse{}
	response := Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return nil, err
	}

	return info.Prompts, nil
}
