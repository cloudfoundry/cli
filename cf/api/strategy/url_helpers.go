package strategy

import (
	"net/url"
	"path"
	"strconv"
)

type params struct {
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

func buildURL(path string, params params) string {
	query := url.Values{}

	if params.inlineRelationsDepth != 0 {
		query.Set("inline-relations-depth", strconv.FormatUint(params.inlineRelationsDepth, 10))
	}

	if params.resultsPerPage != 0 {
		query.Set("results-per-page", strconv.FormatUint(params.resultsPerPage, 10))
	}

	if params.orderDirection != "" {
		query.Set("order-direction", params.orderDirection)
	}

	if params.q != nil {
		q := ""
		for key, value := range params.q {
			q += key + ":" + value
		}
		query.Set("q", q)
	}

	if params.recursive {
		query.Set("recursive", "true")
	}

	return path + "?" + query.Encode()
}
