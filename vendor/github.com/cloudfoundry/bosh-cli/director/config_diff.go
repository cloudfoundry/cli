package director

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ConfigDiff struct {
	Diff [][]interface{}
}

type ConfigDiffResponse struct {
	Diff [][]interface{} `json:"diff"`
}

func NewConfigDiff(diff [][]interface{}) ConfigDiff {
	return ConfigDiff{
		Diff: diff,
	}
}

type DiffConfigError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (c Client) postConfigDiff(path string, manifest []byte, setHeaders func(*http.Request)) (ConfigDiffResponse, error) {
	var resp ConfigDiffResponse

	respBody, response, err := c.clientRequest.RawPost(path, manifest, setHeaders)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			if strings.Contains(err.Error(), "\"code\":440012") {
				// config couldn't be found => return only the director error description
				var descriptionExp = regexp.MustCompile(`description":"(.+?)"`)
				errorDescription := descriptionExp.FindStringSubmatch(err.Error())
				if len(errorDescription) > 0 {
					return resp, bosherr.Errorf(errorDescription[1])
				} else {
					return resp, bosherr.Errorf(err.Error())
				}
			} else {
				// endpoint couldn't be found => return empty diff, just for compatibility with directors which don't have the endpoint
				return resp, nil
			}
		} else {
			return resp, bosherr.WrapError(err, "Fetching diff result")
		}
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return resp, bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return resp, nil
}
