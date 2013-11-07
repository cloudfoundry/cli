package paginator

import (
	"cf/net"
)

type FakePaginator struct{
	TotalResults []string
}

func(p *FakePaginator)Next() (results []string, apiResponse net.ApiResponse){
	if !p.HasNext() {
		return
	}
	current := p.TotalResults[0]
	p.TotalResults = p.TotalResults[1:]
	results = []string{current}
	return
}

func(p *FakePaginator)HasNext() bool {
	return len(p.TotalResults) != 0
}
