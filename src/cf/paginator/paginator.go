package paginator

import (
	"cf/net"
)

type Paginator interface {
	Next() (results []string, apiResponse net.ApiResponse)
	HasNext() bool
}
