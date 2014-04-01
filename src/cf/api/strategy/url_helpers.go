package strategy

import (
	"net/url"
	"path"
	"strconv"
)

type query struct {
	resultsPerPage       uint64
	orderDirection       string
	q                    map[string]string
	recursive            bool
	inlineRelationsDepth uint64
}

func v2(segments ...string) string {
	segments = append([]string{"/v2"}, segments...)
	return path.Join(segments...)
}

func buildURL(path string, query query) string {
	values := url.Values{}

	if query.inlineRelationsDepth != 0 {
		values.Set("inline-relations-depth", strconv.FormatUint(query.inlineRelationsDepth, 10))
	}

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

	if query.recursive {
		values.Set("recursive", "true")
	}

	return path + "?" + values.Encode()
}
