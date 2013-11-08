package paginator

import (
	"cf/net"
	"errors"
)

type Paginator interface {
	Next() (results []string, apiResponse net.ApiResponse)
	HasNext() bool
}

func ForEach(p Paginator, cb func([]string, error)) {
	for {
		rows, apiResponse := p.Next()
		if apiResponse.IsNotSuccessful() {
			cb([]string{}, errors.New(apiResponse.Message))
			return
		}

		cb(rows, nil)

		if !p.HasNext() {
			break
		}
	}
}
