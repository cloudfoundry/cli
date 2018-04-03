package director

import (
	"net/http"
)

//go:generate counterfeiter . Adjustment

type Adjustment interface {
	Adjust(req *http.Request, retried bool) error
	NeedsReadjustment(*http.Response) bool
}

//go:generate counterfeiter . AdjustedClient

type AdjustedClient interface {
	Do(*http.Request) (*http.Response, error)
}

type AdjustableClient struct {
	client     AdjustedClient
	adjustment Adjustment
}

func NewAdjustableClient(client AdjustedClient, adjustment Adjustment) AdjustableClient {
	return AdjustableClient{client: client, adjustment: adjustment}
}

func (c AdjustableClient) Do(req *http.Request) (*http.Response, error) {
	retried := req.Body != nil

	err := c.adjustment.Adjust(req, retried)
	if err != nil {
		return nil, err
	}

	requestBodyBeforeAdjustment := req.Body
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if c.adjustment.NeedsReadjustment(resp) {
		resp.Body.Close()
		err := c.adjustment.Adjust(req, true)
		if err != nil {
			return nil, err
		}

		// Try one more time again after an adjustment
		req.Body = requestBodyBeforeAdjustment
		return c.client.Do(req)
	}

	return resp, nil
}
