package paginator

import (
	"cf/net"
	"errors"
)

type Paginator interface {
	Next() (results []string, apiResponse net.ApiResponse)
	HasNext() bool
}

func ForEach(p Paginator, cb func([]string, error)) (noop bool) {
	noop = true
	for p.HasNext() {
		rows, apiResponse := p.Next()
		if apiResponse.IsNotSuccessful() {
			cb([]string{}, errors.New(apiResponse.Message))
			return
		}

		if noop == true && len(rows) > 0 {
			noop = false
		}

		cb(rows, nil)
	}
	return
}
