package strategy

import (
	"net/url"
	"strconv"
)

type query struct {
	resultsPerPage uint64
	orderDirection string
	q map[string]string
}

func buildURL(path string, query query) string {
	values := url.Values{}
	if query.resultsPerPage != 0 {
		values.Set("results-per-page", strconv.FormatUint(query.resultsPerPage, 10))
	}
	if query.orderDirection != "" {
		values.Set("order-direction", query.orderDirection)
	}
	if query.q != nil {
		q := ""
		for key, value := range query.q {
			q += key + ":" + value
		}
		values.Set("q", q)
	}
	return path + "?" + values.Encode()
}
