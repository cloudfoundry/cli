package director

import (
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

func (d DirectorImpl) MatchPackages(manifest interface{}, compiled bool) ([]string, error) {
	var matches []string

	if compiled {
		matches, _ = d.client.MatchCompiledPackages(manifest)
	} else {
		matches, _ = d.client.MatchPackages(manifest)
	}

	return matches, nil
}

func (c Client) MatchPackages(manifest interface{}) ([]string, error) {
	path := "/packages/matches"

	reqBody, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	var resp []string

	err = c.clientRequest.Post(path, reqBody, setHeaders, &resp)
	if err != nil {
		return nil, bosherr.WrapError(err, "Matching packages")
	}

	return resp, nil
}

func (c Client) MatchCompiledPackages(manifest interface{}) ([]string, error) {
	path := "/packages/matches_compiled"

	reqBody, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	var resp []string

	err = c.clientRequest.Post(path, reqBody, setHeaders, &resp)
	if err != nil {
		return nil, bosherr.WrapError(err, "Matching compiled packages")
	}

	return resp, nil
}
